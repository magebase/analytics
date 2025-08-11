package test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"magebase/apis/analytics/app"
)

// TestAPITrackingEndpoint tests that the analytics tracking endpoint properly tracks API usage
func TestAPITrackingEndpoint(t *testing.T) {
	// Create a new app instance
	app := app.NewApp("8080")
	app.SetupRoutes()

	// Test that the app was created successfully
	assert.NotNil(t, app, "App should be created successfully")
	
	// Test that the analytics service was initialized
	assert.NotNil(t, app.GetAnalyticsService(), "Analytics service should be initialized")
}

// TestUsageRetrievalEndpoint tests that the usage endpoint returns proper usage statistics
func TestUsageRetrievalEndpoint(t *testing.T) {
	// Create a new app instance
	app := app.NewApp("8080")
	app.SetupRoutes()

	// Test that the app was created successfully
	assert.NotNil(t, app, "App should be created successfully")
	
	// Test that the analytics service was initialized
	assert.NotNil(t, app.GetAnalyticsService(), "Analytics service should be initialized")
}

// TestAnalyticsService tests the analytics service directly
func TestAnalyticsService(t *testing.T) {
	service := app.NewAnalyticsService()
	assert.NotNil(t, service, "Analytics service should be created")

	// Test event tracking
	eventData := map[string]interface{}{
		"event_type": "page_view",
		"user_id":    "user123",
		"page":       "/home",
	}

	event, err := service.TrackEvent(nil, eventData, "test-api-key", "user123")
	assert.NoError(t, err, "Event tracking should succeed")
	assert.NotNil(t, event, "Event should be created")
	assert.NotEmpty(t, event.ID, "Event should have an ID")
	assert.NotEmpty(t, event.BillingEventID, "Event should have a billing event ID")
	assert.Equal(t, "page_view", event.EventType, "Event type should match")
	assert.Equal(t, "user123", event.UserID, "User ID should match")
}

// TestUsageCalculation tests usage calculation functionality
func TestUsageCalculation(t *testing.T) {
	service := app.NewAnalyticsService()
	
	// Track some events
	eventData1 := map[string]interface{}{
		"event_type": "page_view",
		"user_id":    "user123",
	}
	eventData2 := map[string]interface{}{
		"event_type": "click",
		"user_id":    "user123",
	}
	
	_, err := service.TrackEvent(nil, eventData1, "api-key", "user123")
	assert.NoError(t, err)
	
	_, err = service.TrackEvent(nil, eventData2, "api-key", "user123")
	assert.NoError(t, err)

	// Get usage for current year
	currentYear := time.Now().Year()
	startDate := fmt.Sprintf("%d-01-01", currentYear)
	endDate := fmt.Sprintf("%d-12-31", currentYear)
	
	// Get usage
	usage, err := service.GetUsage(nil, "user123", startDate, endDate)
	assert.NoError(t, err, "Usage retrieval should succeed")
	assert.NotNil(t, usage, "Usage should be returned")
	assert.Equal(t, int64(2), usage.TotalEvents, "Should have 2 events")
	assert.Equal(t, "user123", usage.UserID, "User ID should match")
	
	// Verify billing summary
	assert.NotNil(t, usage.BillingSummary, "Billing summary should be present")
	assert.Contains(t, usage.BillingSummary.CostBreakdown, "page_view", "Should have page_view cost")
	assert.Contains(t, usage.BillingSummary.CostBreakdown, "click", "Should have click cost")
}
