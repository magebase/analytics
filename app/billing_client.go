package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BillingClient handles communication with the billing service
type BillingClient struct {
	baseURL    string
	httpClient *http.Client
}

// UsageRecord represents a usage record sent to the billing service
type UsageRecord struct {
	UserID    string                 `json:"user_id"`
	Service   string                 `json:"service"`
	Metric    string                 `json:"metric"`
	Amount    int64                  `json:"amount"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// BillingServiceEvent represents a billing event sent to the billing service
type BillingServiceEvent struct {
	UserID    string                 `json:"user_id"`
	Service   string                 `json:"service"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// BillingResponse represents the response from the billing service
type BillingResponse struct {
	ID string `json:"id"`
}

// NewBillingClient creates a new billing client instance
func NewBillingClient(baseURL string) *BillingClient {
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default billing service URL
	}

	return &BillingClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TrackUsage sends a usage record to the billing service
func (c *BillingClient) TrackUsage(ctx context.Context, record *UsageRecord) error {
	jsonData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal usage record: %w", err)
	}

	url := fmt.Sprintf("%s/usage", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send usage record: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	return nil
}

// TrackEvent sends a billing event to the billing service
func (c *BillingClient) TrackEvent(ctx context.Context, event *BillingServiceEvent) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal billing event: %w", err)
	}

	url := fmt.Sprintf("%s/event", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send billing event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("billing service returned status %d", resp.StatusCode)
	}

	return nil
}

// TrackAPICall tracks a single API call for billing purposes
func (c *BillingClient) TrackAPICall(ctx context.Context, userID, endpoint string, metadata map[string]interface{}) error {
	// Create usage record for the API call
	usageRecord := &UsageRecord{
		UserID:    userID,
		Service:   "analytics",
		Metric:    "api_call",
		Amount:    1,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"endpoint": endpoint,
			"metadata": metadata,
		},
	}

	// Send usage record to billing service
	if err := c.TrackUsage(ctx, usageRecord); err != nil {
		return fmt.Errorf("failed to track API call usage: %w", err)
	}

	// Create billing event for the API call
	billingEvent := &BillingServiceEvent{
		UserID:    userID,
		Service:   "analytics",
		EventType: "api_call",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"endpoint": endpoint,
			"metadata": metadata,
		},
	}

	// Send billing event to billing service
	if err := c.TrackEvent(ctx, billingEvent); err != nil {
		return fmt.Errorf("failed to track API call event: %w", err)
	}

	return nil
}
