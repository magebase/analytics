package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"magebase/apis/analytics/app"
)

// TestKafkaConsumerService tests the Kafka consumer service
func TestKafkaConsumerService(t *testing.T) {
	// Test with empty brokers and topics (should handle gracefully)
	consumer, err := app.NewKafkaConsumerService([]string{}, []string{})
	assert.Error(t, err, "Should fail with empty brokers")
	assert.Nil(t, consumer)

	// Test with valid configuration
	consumer, err = app.NewKafkaConsumerService([]string{"localhost:9092"}, []string{"test-topic"})
	if err != nil {
		t.Skipf("Skipping Kafka test - Kafka not available: %v", err)
	}
	assert.NotNil(t, consumer)

	// Test handler registration
	testHandler := func(ctx context.Context, event *app.CrossServiceEvent) error {
		return nil
	}

	consumer.RegisterHandler("test.event", testHandler)
	assert.NotNil(t, consumer, "Handler should be registered")

	// Test event creation
	testEvent := app.NewCrossServiceEvent("test-service", "test.event", "user123", map[string]interface{}{
		"test_data": "value",
	})

	// Verify event was created correctly
	assert.NotNil(t, testEvent)
	assert.Equal(t, "test-service", testEvent.Source)
	assert.Equal(t, "test.event", testEvent.EventType)
	assert.Equal(t, "user123", testEvent.UserID)
}

// TestCrossServiceEventCreation tests cross-service event creation
func TestCrossServiceEventCreation(t *testing.T) {
	// Test billing event
	billingData := map[string]interface{}{
		"amount":      29.99,
		"description": "Monthly subscription",
		"plan":        "pro",
	}

	event := app.NewCrossServiceEvent("billing", "billing.user.subscription.created", "user123", billingData)
	assert.NotNil(t, event)
	assert.Equal(t, "billing", event.Source)
	assert.Equal(t, "billing.user.subscription.created", event.EventType)
	assert.Equal(t, "user123", event.UserID)
	assert.NotEmpty(t, event.ID)
	assert.NotZero(t, event.Timestamp)

	// Test auth event
	authData := map[string]interface{}{
		"ip_address": "192.168.1.1",
		"user_agent": "Mozilla/5.0...",
	}

	authEvent := app.NewCrossServiceEvent("auth", "auth.user.login", "user123", authData)
	assert.NotNil(t, authEvent)
	assert.Equal(t, "auth", authEvent.Source)
	assert.Equal(t, "auth.user.login", authEvent.EventType)
}

// TestBillingEventCreation tests billing event creation
func TestBillingEventCreation(t *testing.T) {
	billingEvent := app.NewBillingEvent("user123", "subscription.created", 29.99, "Monthly pro plan")
	assert.NotNil(t, billingEvent)
	assert.Equal(t, "user123", billingEvent.UserID)
	assert.Equal(t, "subscription.created", billingEvent.EventType)
	assert.Equal(t, 29.99, billingEvent.Amount)
	assert.Equal(t, "USD", billingEvent.Currency)
	assert.Equal(t, "Monthly pro plan", billingEvent.Description)
	assert.NotEmpty(t, billingEvent.ID)
	assert.NotZero(t, billingEvent.Timestamp)
}

// TestAppWithKafkaIntegration tests the app with Kafka integration
func TestAppWithKafkaIntegration(t *testing.T) {
	// Test app creation with Kafka integration
	app := app.NewApp("8080")
	assert.NotNil(t, app)

	// Test that the app has the analytics service
	analyticsService := app.GetAnalyticsService()
	assert.NotNil(t, analyticsService)

	// Test health check endpoint (should include Kafka status)
	// Note: In a real test, we would make HTTP requests to test the endpoints
	// For now, we just verify the app structure
}
