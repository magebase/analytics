package test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"magebase/apis/analytics/app"
)

// TestFunnelService tests the funnel service functionality
func TestFunnelService(t *testing.T) {
	t.Run("FunnelServiceCreation", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)
		assert.NotNil(t, service, "Funnel service should be created")
	})

	t.Run("CreateFunnel", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		steps := []app.Step{
			{ID: "step1", Name: "Page View", EventType: "page_view", Order: 1},
			{ID: "step2", Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
			{ID: "step3", Name: "Checkout", EventType: "checkout", Order: 3},
		}

		funnel, err := service.CreateFunnel(context.Background(), "E-commerce Funnel", "User journey from page view to checkout", steps)
		assert.NoError(t, err, "Funnel creation should succeed")
		assert.NotNil(t, funnel, "Funnel should be created")
		assert.Equal(t, "E-commerce Funnel", funnel.Name)
		assert.Equal(t, 3, len(funnel.Steps))
		assert.NotEmpty(t, funnel.ID)
	})

	t.Run("CreateFunnelValidation", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		// Test empty name
		_, err := service.CreateFunnel(context.Background(), "", "Description", []app.Step{})
		assert.Error(t, err, "Should fail with empty name")
		assert.Contains(t, err.Error(), "funnel name is required")

		// Test insufficient steps
		steps := []app.Step{
			{ID: "step1", Name: "Page View", EventType: "page_view", Order: 1},
		}
		_, err = service.CreateFunnel(context.Background(), "Test Funnel", "Description", steps)
		assert.Error(t, err, "Should fail with insufficient steps")
		assert.Contains(t, err.Error(), "funnel must have at least 2 steps")

		// Test invalid step order
		invalidSteps := []app.Step{
			{ID: "step1", Name: "Page View", EventType: "page_view", Order: 2},
			{ID: "step2", Name: "Add to Cart", EventType: "add_to_cart", Order: 1},
		}
		_, err = service.CreateFunnel(context.Background(), "Test Funnel", "Description", invalidSteps)
		assert.Error(t, err, "Should fail with invalid step order")
		assert.Contains(t, err.Error(), "step order must be sequential")

		// Test missing event type
		invalidSteps2 := []app.Step{
			{ID: "step1", Name: "Page View", EventType: "", Order: 1},
			{ID: "step2", Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
		}
		_, err = service.CreateFunnel(context.Background(), "Test Funnel", "Description", invalidSteps2)
		assert.Error(t, err, "Should fail with missing event type")
		assert.Contains(t, err.Error(), "event type is required")
	})

	t.Run("ComputeFunnel", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		query := app.FunnelQuery{
			FunnelID: "test_funnel",
			Start:    time.Now().AddDate(0, 0, -30),
			End:      time.Now(),
		}

		result, err := service.ComputeFunnel(context.Background(), query)
		assert.NoError(t, err, "Funnel computation should succeed")
		assert.NotNil(t, result, "Result should be returned")
		assert.Equal(t, "test_funnel", result.FunnelID)
		assert.Equal(t, "Sample Funnel", result.FunnelName)
		assert.Equal(t, 4, len(result.Steps))
		assert.Greater(t, result.TotalUsers, int64(0))
		assert.Greater(t, result.ConversionRate, 0.0)
	})

	t.Run("ComputeFunnelValidation", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		// Test empty funnel ID
		query := app.FunnelQuery{
			FunnelID: "",
			Start:    time.Now().AddDate(0, 0, -30),
			End:      time.Now(),
		}

		_, err := service.ComputeFunnel(context.Background(), query)
		assert.Error(t, err, "Should fail with empty funnel ID")
		assert.Contains(t, err.Error(), "funnel ID is required")
	})

	t.Run("GetFunnelSteps", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		steps, err := service.GetFunnelSteps(context.Background(), "test_funnel")
		assert.NoError(t, err, "Should retrieve funnel steps")
		assert.NotNil(t, steps, "Steps should be returned")
		assert.Equal(t, 4, len(steps))
		assert.Equal(t, "Page View", steps[0].Name)
		assert.Equal(t, "page_view", steps[0].EventType)
		assert.Equal(t, 1, steps[0].Order)
	})

	t.Run("GetFunnelStepsValidation", func(t *testing.T) {
		analyticsService := app.NewAnalyticsService()
		service := app.NewFunnelService(analyticsService)

		// Test empty funnel ID
		_, err := service.GetFunnelSteps(context.Background(), "")
		assert.Error(t, err, "Should fail with empty funnel ID")
		assert.Contains(t, err.Error(), "funnel ID is required")
	})

	t.Run("FunnelStepStructure", func(t *testing.T) {
		step := app.Step{
			ID:          "step1",
			Name:        "Page View",
			EventType:   "page_view",
			Order:       1,
			Description: "User views a page",
			Filters: map[string]interface{}{
				"page_type": "product",
			},
		}

		assert.Equal(t, "step1", step.ID)
		assert.Equal(t, "Page View", step.Name)
		assert.Equal(t, "page_view", step.EventType)
		assert.Equal(t, 1, step.Order)
		assert.Equal(t, "User views a page", step.Description)
		assert.Contains(t, step.Filters, "page_type")
		assert.Equal(t, "product", step.Filters["page_type"])
	})

	t.Run("FunnelResultStructure", func(t *testing.T) {
		result := app.FunnelResult{
			FunnelID:       "funnel_123",
			FunnelName:     "Test Funnel",
			ConversionRate: 15.5,
			TotalUsers:     1000,
			ComputedAt:     time.Now(),
		}

		assert.Equal(t, "funnel_123", result.FunnelID)
		assert.Equal(t, "Test Funnel", result.FunnelName)
		assert.Equal(t, 15.5, result.ConversionRate)
		assert.Equal(t, int64(1000), result.TotalUsers)
		assert.NotNil(t, result.ComputedAt)
	})

	t.Run("StepResultStructure", func(t *testing.T) {
		stepResult := app.StepResult{
			StepID:         "step1",
			StepName:       "Page View",
			EventCount:     1500,
			UniqueUsers:    1000,
			DropOffRate:    0.0,
			ConversionRate: 100.0,
		}

		assert.Equal(t, "step1", stepResult.StepID)
		assert.Equal(t, "Page View", stepResult.StepName)
		assert.Equal(t, int64(1500), stepResult.EventCount)
		assert.Equal(t, int64(1000), stepResult.UniqueUsers)
		assert.Equal(t, 0.0, stepResult.DropOffRate)
		assert.Equal(t, 100.0, stepResult.ConversionRate)
	})
}

// TestFunnelIntegration tests funnel service integration with the main app
func TestFunnelIntegration(t *testing.T) {
	t.Run("AppWithFunnelService", func(t *testing.T) {
		app := app.NewApp("8080")
		assert.NotNil(t, app, "App should be created successfully")

		funnelService := app.GetFunnelService()
		assert.NotNil(t, funnelService, "Funnel service should be initialized")
	})

	t.Run("FunnelEndpointsAvailable", func(t *testing.T) {
		app := app.NewApp("8080")
		app.SetupRoutes()

		// Test that the funnel service is accessible
		funnelService := app.GetFunnelService()
		assert.NotNil(t, funnelService, "Funnel service should be available")
	})
}
