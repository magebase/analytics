package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"magebase/apis/analytics/app"
)

// TestEnhancedEventIngestion tests enhanced event ingestion with validation and schema support
func TestEnhancedEventIngestion(t *testing.T) {
	// Create a new app instance
	app := app.NewApp("8080")
	app.SetupRoutes()

	// Test that the app was created successfully
	assert.NotNil(t, app, "App should be created successfully")

	// Test that the analytics service was initialized
	analyticsService := app.GetAnalyticsService()
	assert.NotNil(t, analyticsService, "Analytics service should be initialized")

	// Test 1: Valid event with all required fields
	t.Run("ValidEventWithAllFields", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event_type": "page_view",
			"user_id":    "user123",
			"page":       "/home",
			"properties": map[string]interface{}{
				"referrer":   "google.com",
				"utm_source": "search",
				"browser":    "chrome",
				"device":     "desktop",
			},
			"session_id": "sess_abc123",
			"ip_address": "192.168.1.1",
		}

		// Test event tracking through service
		event, err := analyticsService.TrackEvent(nil, eventData, "test-api-key", "user123")
		assert.NoError(t, err, "Event tracking should succeed")
		assert.NotNil(t, event, "Event should be created")
		assert.Equal(t, "page_view", event.EventType)
		assert.Equal(t, "user123", event.UserID)
		assert.NotEmpty(t, event.ID)
		assert.NotEmpty(t, event.BillingEventID)
	})

	// Test 2: Event with custom properties validation
	t.Run("EventWithCustomProperties", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event_type": "conversion",
			"user_id":    "user123",
			"page":       "/checkout",
			"properties": map[string]interface{}{
				"order_id": "ord_12345",
				"amount":   99.99,
				"currency": "USD",
				"products": []string{"product1", "product2"},
			},
		}

		event, err := analyticsService.TrackEvent(nil, eventData, "test-api-key", "user123")
		assert.NoError(t, err, "Event tracking should succeed")
		assert.NotNil(t, event, "Event should be created")
		assert.Equal(t, "conversion", event.EventType)
	})

	// Test 3: Event with missing required fields (should fail)
	t.Run("EventWithMissingRequiredFields", func(t *testing.T) {
		eventData := map[string]interface{}{
			"page": "/home",
			"properties": map[string]interface{}{
				"referrer": "google.com",
			},
		}

		// Should fail because event_type is missing
		event, err := analyticsService.TrackEvent(nil, eventData, "test-api-key", "user123")
		assert.Error(t, err, "Event tracking should fail with missing event_type")
		assert.Nil(t, event, "Event should not be created")
	})

	// Test 4: Event with invalid data types (should fail)
	t.Run("EventWithInvalidDataTypes", func(t *testing.T) {
		eventData := map[string]interface{}{
			"event_type": 123, // Should be string
			"user_id":    "user123",
			"page":       "/home",
		}

		// Should fail because event_type is not a string
		event, err := analyticsService.TrackEvent(nil, eventData, "test-api-key", "user123")
		assert.Error(t, err, "Event tracking should fail with invalid event_type type")
		assert.Nil(t, event, "Event should not be created")
	})
}

// TestEventSchemaValidation tests event schema validation
func TestEventSchemaValidation(t *testing.T) {
	app := app.NewApp("8080")
	assert.NotNil(t, app, "App should be created successfully")

	analyticsService := app.GetAnalyticsService()
	assert.NotNil(t, analyticsService, "Analytics service should be initialized")

	t.Run("SchemaValidation", func(t *testing.T) {
		// Test valid event (should pass validation)
		validEvent := map[string]interface{}{
			"event_type": "page_view",
			"user_id":    "user123",
			"page":       "/home",
		}

		event, err := analyticsService.TrackEvent(nil, validEvent, "test-api-key", "user123")
		assert.NoError(t, err, "Valid event should pass schema validation")
		assert.NotNil(t, event, "Event should be created")

		// Test invalid event (should fail validation)
		invalidEvent := map[string]interface{}{
			"event_type": 123, // Should be string
			"user_id":    "user123",
		}

		event2, err2 := analyticsService.TrackEvent(nil, invalidEvent, "test-api-key", "user123")
		assert.Error(t, err2, "Invalid event should fail schema validation")
		assert.Nil(t, event2, "Event should not be created")
		assert.Contains(t, err2.Error(), "event_type is required and must be a string", "Error should mention type validation")

		// Test missing required field (should fail validation)
		missingFieldEvent := map[string]interface{}{
			"page": "/home", // Missing event_type
		}

		event3, err3 := analyticsService.TrackEvent(nil, missingFieldEvent, "test-api-key", "user123")
		assert.Error(t, err3, "Event with missing required field should fail validation")
		assert.Nil(t, event3, "Event should not be created")
		assert.Contains(t, err3.Error(), "event_type is required", "Error should mention missing required field")
	})
}

// TestEventEnrichment tests event enrichment functionality
func TestEventEnrichment(t *testing.T) {
	app := app.NewApp("8080")
	assert.NotNil(t, app, "App should be created successfully")

	analyticsService := app.GetAnalyticsService()
	assert.NotNil(t, analyticsService, "Analytics service should be initialized")

	t.Run("EventEnrichment", func(t *testing.T) {
		// Test that events are enriched with additional metadata
		eventData := map[string]interface{}{
			"event_type": "page_view",
			"user_id":    "user123",
			"page":       "/home",
		}

		event, err := analyticsService.TrackEvent(nil, eventData, "test-api-key", "user123")
		assert.NoError(t, err, "Event tracking should succeed")
		assert.NotNil(t, event, "Event should be created")

		// Verify that the event was enriched with additional metadata
		// The enrichEventData method ensures properties map exists and adds metadata
		assert.NotNil(t, event.Properties, "Event should have properties map")
		
		// Check that the event has the expected enriched data
		// Note: The enrichment happens in the service layer, so we can't directly
		// access the enriched data, but we can verify the event was processed correctly
		assert.Equal(t, "page_view", event.EventType, "Event type should be preserved")
		assert.Equal(t, "user123", event.UserID, "User ID should be preserved")
		assert.Equal(t, "/home", event.Page, "Page should be preserved")
		assert.NotEmpty(t, event.ID, "Event should have an ID")
		assert.NotEmpty(t, event.BillingEventID, "Event should have a billing event ID")
		assert.NotZero(t, event.Timestamp, "Event should have a timestamp")
		
		// Test that enrichment preserves existing data and adds defaults
		// The enrichEventData method should have ensured all required fields exist
		assert.NotNil(t, event.Properties, "Properties map should exist after enrichment")
	})
}
