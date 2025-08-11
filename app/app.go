package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// App represents the analytics application
type App struct {
	app             *fiber.App
	tracer          trace.Tracer
	port            string
	analyticsService *AnalyticsService
	kafkaConsumer   *KafkaConsumerService
}

// NewApp creates a new analytics application instance
func NewApp(port string) *App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := http.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Initialize tracer
	tracer := otel.Tracer("analytics")

	// Initialize analytics service
	analyticsService := NewAnalyticsService()

	// Create app instance first
	appInstance := &App{
		app:             app,
		tracer:          tracer,
		port:            port,
		analyticsService: analyticsService,
		kafkaConsumer:   nil, // Will be initialized after creation
	}

	// Initialize Kafka consumer service
	kafkaConsumer := appInstance.initializeKafkaConsumer()
	appInstance.kafkaConsumer = kafkaConsumer

	return appInstance
}

// initializeKafkaConsumer initializes the Kafka consumer service
func (s *App) initializeKafkaConsumer() *KafkaConsumerService {
	// Get Kafka configuration from environment
	brokers := s.getKafkaBrokers()
	topics := s.getKafkaTopics()

	if len(brokers) == 0 || len(topics) == 0 {
		log.Println("Warning: Kafka configuration not found, consumer service will not start")
		return nil
	}

	consumer, err := NewKafkaConsumerService(brokers, topics)
	if err != nil {
		log.Printf("Warning: Failed to create Kafka consumer: %v", err)
		return nil
	}

	// Start the consumer service
	if err := consumer.Start(); err != nil {
		log.Printf("Warning: Failed to start Kafka consumer: %v", err)
		return nil
	}

	log.Printf("Kafka consumer service started for topics: %v", topics)
	return consumer
}

// getKafkaBrokers gets Kafka broker addresses from environment
func (s *App) getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"} // Default
	}
	return strings.Split(brokers, ",")
}

// getKafkaTopics gets Kafka topics from environment
func (s *App) getKafkaTopics() []string {
	topics := os.Getenv("KAFKA_TOPICS")
	if topics == "" {
		return []string{"billing", "auth", "payments", "analytics"} // Default topics
	}
	return strings.Split(topics, ",")
}

// SetupRoutes configures all the application routes
func (s *App) SetupRoutes() {
	// Health check endpoint
	s.app.Get("/health", s.healthCheck)
	
	// Analytics endpoints
	analytics := s.app.Group("/api/v1/analytics")
	analytics.Post("/events", s.trackEvent)
	analytics.Get("/usage", s.getUsage)
	
	// Kafka consumer status endpoint
	s.app.Get("/api/v1/kafka/status", s.getKafkaStatus)
}

// Start begins the application server
func (s *App) Start(ctx context.Context) error {
	log.Printf("Starting analytics service on port %s", s.port)
	return s.app.Listen(":" + s.port)
}

// Stop gracefully shuts down the application
func (s *App) Stop() {
	if s.kafkaConsumer != nil {
		s.kafkaConsumer.Stop()
	}
	log.Println("Analytics service stopped")
}

// healthCheck handles health check requests
func (s *App) healthCheck(c *fiber.Ctx) error {
	kafkaStatus := "disabled"
	if s.kafkaConsumer != nil {
		kafkaStatus = "running"
	}

	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "analytics",
		"kafka":   kafkaStatus,
	})
}

// trackEvent handles analytics event tracking
func (s *App) trackEvent(c *fiber.Ctx) error {
	// Parse request body
	var eventData map[string]interface{}
	if err := c.BodyParser(&eventData); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Extract headers
	apiKey := c.Get("X-API-Key")
	userID := c.Get("X-User-ID")

	if apiKey == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "API key is required",
		})
	}

	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	// Track the event
	event, err := s.analyticsService.TrackEvent(c.Context(), eventData, apiKey, userID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return success response
	return c.JSON(fiber.Map{
		"status":           "success",
		"event_id":         event.ID,
		"tracked_at":       event.Timestamp,
		"billing_event_id": event.BillingEventID,
	})
}

// getUsage retrieves usage statistics
func (s *App) getUsage(c *fiber.Ctx) error {
	// Extract query parameters
	userID := c.Query("user_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if userID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02") // Default to 30 days ago
	}

	if endDate == "" {
		endDate = time.Now().Format("2006-01-02") // Default to today
	}

	// Get usage statistics
	usage, err := s.analyticsService.GetUsage(c.Context(), userID, startDate, endDate)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return usage data
	return c.JSON(fiber.Map{
		"total_events":    usage.TotalEvents,
		"events_by_type":  usage.EventsByType,
		"billing_summary": usage.BillingSummary,
		"total_cost":      usage.BillingSummary.TotalCost,
		"cost_breakdown":  usage.BillingSummary.CostBreakdown,
	})
}

// getKafkaStatus returns the status of the Kafka consumer service
func (s *App) getKafkaStatus(c *fiber.Ctx) error {
	status := "disabled"
	if s.kafkaConsumer != nil {
		status = "running"
	}

	return c.JSON(fiber.Map{
		"status": status,
		"topics": s.getKafkaTopics(),
		"brokers": s.getKafkaBrokers(),
	})
}

// GetFiberApp returns the underlying Fiber app for testing purposes
func (s *App) GetFiberApp() *fiber.App {
	return s.app
}

// GetAnalyticsService returns the analytics service for testing purposes
func (s *App) GetAnalyticsService() *AnalyticsService {
	return s.analyticsService
}
