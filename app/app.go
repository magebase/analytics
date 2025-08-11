package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// App represents the analytics application
type App struct {
	app              *fiber.App
	tracer           trace.Tracer
	port             string
	analyticsService *AnalyticsService
	kafkaConsumer    *KafkaConsumerService
	dashboardService *DashboardService
	funnelService    *FunnelService
	heatmapService   *HeatmapService
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

	// Initialize dashboard service
	dashboardService := NewDashboardService()

	// Initialize funnel service
	funnelService := NewFunnelService(analyticsService)

	// Initialize heatmap service
	heatmapService := NewHeatmapService(analyticsService)

	// Create app instance first
	appInstance := &App{
		app:              app,
		tracer:           tracer,
		port:             port,
		analyticsService: analyticsService,
		kafkaConsumer:    nil, // Will be initialized after creation
		dashboardService: dashboardService,
		funnelService:    funnelService,
		heatmapService:   heatmapService,
	}

	// Start dashboard service
	dashboardService.Start()

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
	// Initialize middleware
	apiTrackingMiddleware := NewAPITrackingMiddleware(s.analyticsService)
	rateLimitMiddleware := NewRateLimitMiddleware(s.analyticsService)
	samplingMiddleware := NewSamplingMiddleware(s.analyticsService)

	// Apply global middleware for all routes
	s.app.Use(apiTrackingMiddleware.TrackAPIUsage())
	s.app.Use(rateLimitMiddleware.RateLimit())
	s.app.Use(samplingMiddleware.Sample())

	// Health check endpoint
	s.app.Get("/health", s.healthCheck)

	// Analytics endpoints
	analytics := s.app.Group("/api/v1/analytics")
	analytics.Post("/events", s.trackEvent)
	analytics.Get("/usage", s.getUsage)

	// Real-time dashboard WebSocket endpoint
	s.app.Get("/api/v1/dashboard/feed", websocket.New(s.dashboardService.HandleWebSocket))

	// Funnel analysis endpoints
	funnels := s.app.Group("/api/v1/funnels")
	funnels.Post("/", s.createFunnel)
	funnels.Get("/:id/compute", s.computeFunnel)
	funnels.Get("/:id/steps", s.getFunnelSteps)

	// Heatmap endpoints
	heatmaps := s.app.Group("/api/v1/heatmaps")
	heatmaps.Post("/", s.createHeatmap)
	heatmaps.Post("/generate", s.generateHeatmap)
	heatmaps.Get("/:id", s.getHeatmap)

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
		"status":  status,
		"topics":  s.getKafkaTopics(),
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

// GetDashboardService returns the dashboard service for testing purposes
func (s *App) GetDashboardService() *DashboardService {
	return s.dashboardService
}

// GetFunnelService returns the funnel service for testing purposes
func (s *App) GetFunnelService() *FunnelService {
	return s.funnelService
}

// GetHeatmapService returns the heatmap service for testing purposes
func (s *App) GetHeatmapService() *HeatmapService {
	return s.heatmapService
}

// createFunnel handles funnel creation requests
func (s *App) createFunnel(c *fiber.Ctx) error {
	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Steps       []struct {
			Name        string                 `json:"name"`
			EventType   string                 `json:"event_type"`
			Filters     map[string]interface{} `json:"filters,omitempty"`
			Order       int                    `json:"order"`
			Description string                 `json:"description,omitempty"`
		} `json:"steps"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Convert request steps to Step structs
	var steps []Step
	for _, reqStep := range request.Steps {
		step := Step{
			ID:          fmt.Sprintf("step_%d", reqStep.Order),
			Name:        reqStep.Name,
			EventType:   reqStep.EventType,
			Filters:     reqStep.Filters,
			Order:       reqStep.Order,
			Description: reqStep.Description,
		}
		steps = append(steps, step)
	}

	funnel, err := s.funnelService.CreateFunnel(c.Context(), request.Name, request.Description, steps)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"funnel":  funnel,
		"message": "Funnel created successfully",
	})
}

// computeFunnel handles funnel computation requests
func (s *App) computeFunnel(c *fiber.Ctx) error {
	funnelID := c.Params("id")
	if funnelID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Funnel ID is required",
		})
	}

	// Parse query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	userID := c.Query("user_id")

	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
		})
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
		})
	}

	query := FunnelQuery{
		FunnelID: funnelID,
		UserID:   userID,
		Start:    start,
		End:      end,
	}

	result, err := s.funnelService.ComputeFunnel(c.Context(), query)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"result": result,
	})
}

// getFunnelSteps retrieves the steps for a specific funnel
func (s *App) getFunnelSteps(c *fiber.Ctx) error {
	funnelID := c.Params("id")
	if funnelID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Funnel ID is required",
		})
	}

	steps, err := s.funnelService.GetFunnelSteps(c.Context(), funnelID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"steps":  steps,
	})
}

// createHeatmap handles heatmap creation requests
func (s *App) createHeatmap(c *fiber.Ctx) error {
	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Page        string `json:"page"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	heatmap, err := s.heatmapService.CreateHeatmap(c.Context(), request.Name, request.Description, request.Type, request.Page, request.Width, request.Height)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"heatmap": heatmap,
		"message": "Heatmap created successfully",
	})
}

// generateHeatmap handles heatmap generation requests
func (s *App) generateHeatmap(c *fiber.Ctx) error {
	var request HeatmapQuery

	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Set default time range if not provided
	if request.Start.IsZero() {
		request.Start = time.Now().AddDate(0, 0, -30)
	}
	if request.End.IsZero() {
		request.End = time.Now()
	}

	result, err := s.heatmapService.GenerateHeatmap(c.Context(), request)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"result": result,
	})
}

// getHeatmap retrieves a specific heatmap
func (s *App) getHeatmap(c *fiber.Ctx) error {
	heatmapID := c.Params("id")
	if heatmapID == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Heatmap ID is required",
		})
	}

	heatmap, err := s.heatmapService.GetHeatmap(c.Context(), heatmapID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"heatmap": heatmap,
	})
}
