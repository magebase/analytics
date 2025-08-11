package app

import (
	"time"

	"github.com/google/uuid"
)

// AnalyticsEvent represents an analytics event that needs to be tracked
type AnalyticsEvent struct {
	ID          string                 `json:"id"`
	EventType   string                 `json:"event_type"`
	UserID      string                 `json:"user_id"`
	Page        string                 `json:"page,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	APIKey      string                 `json:"api_key"`
	BillingEventID string              `json:"billing_event_id,omitempty"`
}

// UsageSummary represents usage statistics for a user
type UsageSummary struct {
	UserID        string                 `json:"user_id"`
	TotalEvents   int64                  `json:"total_events"`
	EventsByType  map[string]int64       `json:"events_by_type"`
	BillingSummary BillingSummary        `json:"billing_summary"`
	Period        UsagePeriod            `json:"period"`
}

// BillingSummary represents billing information for usage
type BillingSummary struct {
	TotalCost     float64                `json:"total_cost"`
	CostBreakdown map[string]float64     `json:"cost_breakdown"`
	Currency      string                 `json:"currency"`
}

// UsagePeriod represents the time period for usage queries
type UsagePeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
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
	}
}
