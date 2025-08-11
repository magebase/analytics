package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// AnalyticsService handles analytics event processing and billing integration
type AnalyticsService struct {
	events map[string]*AnalyticsEvent // In-memory storage for now
}

// NewAnalyticsService creates a new analytics service instance
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		events: make(map[string]*AnalyticsEvent),
	}
}

// TrackEvent processes and stores an analytics event
func (s *AnalyticsService) TrackEvent(ctx context.Context, eventData map[string]interface{}, apiKey, userID string) (*AnalyticsEvent, error) {
	// Validate required fields
	if err := s.validateEventData(eventData); err != nil {
		return nil, fmt.Errorf("invalid event data: %w", err)
	}

	// Create analytics event
	event := &AnalyticsEvent{
		ID:         uuid.New().String(),
		EventType:  eventData["event_type"].(string),
		UserID:     userID,
		Page:       s.getStringValue(eventData, "page"),
		Timestamp:  time.Now(),
		Properties: s.getMapValue(eventData, "properties"),
		APIKey:     apiKey,
	}

	// Generate billing event ID (simulating billing integration)
	event.BillingEventID = uuid.New().String()

	// Store event (in-memory for now)
	s.events[event.ID] = event

	// Log the event for debugging
	log.Printf("Tracked event: %s for user: %s, billing_event_id: %s", event.EventType, event.UserID, event.BillingEventID)

	return event, nil
}

// GetUsage retrieves usage statistics for a user
func (s *AnalyticsService) GetUsage(ctx context.Context, userID, startDateStr, endDateStr string) (*UsageSummary, error) {
	// Parse dates - handle both date formats
	var startDate, endDate time.Time
	var err error
	
	// Try parsing as "2006-01-02" first
	startDate, err = time.Parse("2006-01-02", startDateStr)
	if err != nil {
		// Try parsing as RFC3339 format
		startDate, err = time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
	}
	
	endDate, err = time.Parse("2006-01-02", endDateStr)
	if err != nil {
		// Try parsing as RFC3339 format
		endDate, err = time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
	}

	// Calculate usage from stored events
	eventsByType := make(map[string]int64)
	var totalEvents int64

	for _, event := range s.events {
		if event.UserID == userID && 
		   event.Timestamp.After(startDate) && 
		   event.Timestamp.Before(endDate.Add(24*time.Hour)) {
			totalEvents++
			eventsByType[event.EventType]++
		}
	}

	// Calculate billing summary (simulating billing integration)
	billingSummary := s.calculateBillingSummary(eventsByType)

	usage := &UsageSummary{
		UserID:        userID,
		TotalEvents:   totalEvents,
		EventsByType:  eventsByType,
		BillingSummary: billingSummary,
		Period: UsagePeriod{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	return usage, nil
}

// validateEventData validates the incoming event data
func (s *AnalyticsService) validateEventData(eventData map[string]interface{}) error {
	// Check required fields
	if eventData["event_type"] == nil {
		return fmt.Errorf("event_type is required")
	}

	if _, ok := eventData["event_type"].(string); !ok {
		return fmt.Errorf("event_type must be a string")
	}

	return nil
}

// getStringValue safely extracts a string value from the event data
func (s *AnalyticsService) getStringValue(eventData map[string]interface{}, key string) string {
	if val, ok := eventData[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// getMapValue safely extracts a map value from the event data
func (s *AnalyticsService) getMapValue(eventData map[string]interface{}, key string) map[string]interface{} {
	if val, ok := eventData[key]; ok {
		if m, ok := val.(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}

// calculateBillingSummary calculates billing information based on event types
func (s *AnalyticsService) calculateBillingSummary(eventsByType map[string]int64) BillingSummary {
	costBreakdown := make(map[string]float64)
	var totalCost float64

	// Simple pricing model (events per type)
	for eventType, count := range eventsByType {
		var cost float64
		switch eventType {
		case "page_view":
			cost = float64(count) * 0.001 // $0.001 per page view
		case "click":
			cost = float64(count) * 0.002 // $0.002 per click
		case "conversion":
			cost = float64(count) * 0.01  // $0.01 per conversion
		default:
			cost = float64(count) * 0.0005 // $0.0005 per other event
		}
		costBreakdown[eventType] = cost
		totalCost += cost
	}

	return BillingSummary{
		TotalCost:     totalCost,
		CostBreakdown: costBreakdown,
		Currency:      "USD",
	}
}
