package app

import (
	"context"
	"log"
	"net/http"
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

	return &App{
		app:             app,
		tracer:          tracer,
		port:            port,
		analyticsService: analyticsService,
	}
}

// SetupRoutes configures all the application routes
func (a *App) SetupRoutes() {
	// Health check endpoint
	a.app.Get("/health", a.healthCheck)
	
	// Analytics endpoints
	analytics := a.app.Group("/api/v1/analytics")
	analytics.Post("/events", a.trackEvent)
	analytics.Get("/usage", a.getUsage)
}

// Start begins the application server
func (a *App) Start(ctx context.Context) error {
	log.Printf("Starting analytics service on port %s", a.port)
	return a.app.Listen(":" + a.port)
}

// healthCheck handles health check requests
func (a *App) healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "analytics",
	})
}

// trackEvent handles analytics event tracking
func (a *App) trackEvent(c *fiber.Ctx) error {
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
	event, err := a.analyticsService.TrackEvent(c.Context(), eventData, apiKey, userID)
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
func (a *App) getUsage(c *fiber.Ctx) error {
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
	usage, err := a.analyticsService.GetUsage(c.Context(), userID, startDate, endDate)
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

// GetFiberApp returns the underlying Fiber app for testing purposes
func (a *App) GetFiberApp() *fiber.App {
	return a.app
}

// GetAnalyticsService returns the analytics service for testing purposes
func (a *App) GetAnalyticsService() *AnalyticsService {
	return a.analyticsService
}
