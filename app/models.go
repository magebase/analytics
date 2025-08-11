package app

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsEvent represents an analytics event that needs to be tracked
type AnalyticsEvent struct {
	ID             string                 `json:"id"`
	EventType      string                 `json:"event_type"`
	UserID         string                 `json:"user_id"`
	Page           string                 `json:"page,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
	APIKey         string                 `json:"api_key"`
	BillingEventID string                 `json:"billing_event_id,omitempty"`
	Source         string                 `json:"source,omitempty"`
}

// CrossServiceEvent represents events from other services (billing, auth, payments, etc.)
type CrossServiceEvent struct {
	ID            string                 `json:"id"`
	Source        string                 `json:"source"`     // e.g., "billing", "auth", "payments"
	EventType     string                 `json:"event_type"` // e.g., "user.login", "payment.completed"
	UserID        string                 `json:"user_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data"` // Service-specific data
	CorrelationID string                 `json:"correlation_id,omitempty"`
}

// UsageSummary represents usage statistics for a user
type UsageSummary struct {
	UserID         string           `json:"user_id"`
	TotalEvents    int64            `json:"total_events"`
	EventsByType   map[string]int64 `json:"events_by_type"`
	BillingSummary BillingSummary   `json:"billing_summary"`
	Period         UsagePeriod      `json:"period"`
}

// BillingSummary represents billing information for usage
type BillingSummary struct {
	TotalCost     float64            `json:"total_cost"`
	CostBreakdown map[string]float64 `json:"cost_breakdown"`
	Currency      string             `json:"currency"`
}

// UsagePeriod represents the time period for usage queries
type UsagePeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// BillingEvent represents a billing event for cost tracking
type BillingEvent struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	EventType   string    `json:"event_type"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
}

// NewAnalyticsEvent creates a new analytics event with a unique ID
func NewAnalyticsEvent(eventType, userID, page, apiKey string, properties map[string]interface{}) *AnalyticsEvent {
	return &AnalyticsEvent{
		ID:         uuid.New().String(),
		EventType:  eventType,
		UserID:     userID,
		Page:       page,
		Timestamp:  time.Now(),
		Properties: properties,
		APIKey:     apiKey,
		Source:     "analytics",
	}
}

// NewCrossServiceEvent creates a new cross-service event
func NewCrossServiceEvent(source, eventType, userID string, data map[string]interface{}) *CrossServiceEvent {
	return &CrossServiceEvent{
		ID:        uuid.New().String(),
		Source:    source,
		EventType: eventType,
		UserID:    userID,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// NewBillingEvent creates a new billing event
func NewBillingEvent(userID, eventType string, amount float64, description string) *BillingEvent {
	return &BillingEvent{
		ID:          uuid.New().String(),
		UserID:      userID,
		EventType:   eventType,
		Amount:      amount,
		Currency:    "USD",
		Timestamp:   time.Now(),
		Description: description,
	}
}
