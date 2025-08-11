package app

import (
	"context"
	"fmt"
	"log"
	"time"
)

// FunnelService computes conversion funnels from analytics events
type FunnelService struct {
	analyticsService *AnalyticsService
}

// Funnel represents a conversion funnel with multiple steps
type Funnel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Step represents a step in a conversion funnel
type Step struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	EventType   string                 `json:"event_type"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Order       int                    `json:"order"`
	Description string                 `json:"description,omitempty"`
}

// FunnelResult represents the computed results of a funnel
type FunnelResult struct {
	FunnelID       string       `json:"funnel_id"`
	FunnelName     string       `json:"funnel_name"`
	TimeRange      TimeRange    `json:"time_range"`
	Steps          []StepResult `json:"steps"`
	ConversionRate float64      `json:"conversion_rate"`
	TotalUsers     int64        `json:"total_users"`
	ComputedAt     time.Time    `json:"computed_at"`
}

// StepResult represents the results for a specific funnel step
type StepResult struct {
	StepID         string  `json:"step_id"`
	StepName       string  `json:"step_name"`
	EventCount     int64   `json:"event_count"`
	UniqueUsers    int64   `json:"unique_users"`
	DropOffRate    float64 `json:"drop_off_rate"`
	ConversionRate float64 `json:"conversion_rate"`
}

// TimeRange represents a time period for funnel analysis
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// FunnelQuery represents a query for funnel computation
type FunnelQuery struct {
	FunnelID string    `json:"funnel_id"`
	UserID   string    `json:"user_id,omitempty"`
	Start    time.Time `json:"start"`
	End      time.Time `json:"end"`
}

// NewFunnelService creates a new funnel service instance
func NewFunnelService(analyticsService *AnalyticsService) *FunnelService {
	return &FunnelService{
		analyticsService: analyticsService,
	}
}

// CreateFunnel creates a new conversion funnel
func (s *FunnelService) CreateFunnel(ctx context.Context, name, description string, steps []Step) (*Funnel, error) {
	if name == "" {
		return nil, fmt.Errorf("funnel name is required")
	}

	if len(steps) < 2 {
		return nil, fmt.Errorf("funnel must have at least 2 steps")
	}

	// Validate step order
	for i, step := range steps {
		if step.Order != i+1 {
			return nil, fmt.Errorf("step order must be sequential starting from 1")
		}
		if step.EventType == "" {
			return nil, fmt.Errorf("event type is required for step %d", i+1)
		}
	}

	funnel := &Funnel{
		ID:          generateFunnelID(),
		Name:        name,
		Description: description,
		Steps:       steps,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// In a real implementation, this would be stored in a database
	log.Printf("Created funnel: %s with %d steps", funnel.ID, len(funnel.Steps))

	return funnel, nil
}

// ComputeFunnel computes the conversion funnel results for a given time range
func (s *FunnelService) ComputeFunnel(ctx context.Context, query FunnelQuery) (*FunnelResult, error) {
	if query.FunnelID == "" {
		return nil, fmt.Errorf("funnel ID is required")
	}

	// In a real implementation, this would fetch the funnel definition from a database
	// For now, we'll create a mock funnel
	funnel := &Funnel{
		ID:   query.FunnelID,
		Name: "Sample Funnel",
		Steps: []Step{
			{ID: "step1", Name: "Page View", EventType: "page_view", Order: 1},
			{ID: "step2", Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
			{ID: "step3", Name: "Checkout", EventType: "checkout", Order: 3},
			{ID: "step4", Name: "Purchase", EventType: "purchase", Order: 4},
		},
	}

	// Compute funnel results
	result := &FunnelResult{
		FunnelID:   funnel.ID,
		FunnelName: funnel.Name,
		TimeRange:  TimeRange{Start: query.Start, End: query.End},
		ComputedAt: time.Now(),
	}

	// In a real implementation, this would query the analytics database
	// For now, we'll generate mock data
	result.Steps = s.generateMockStepResults(funnel.Steps)
	result.TotalUsers = s.calculateTotalUsers(result.Steps)
	result.ConversionRate = s.calculateOverallConversionRate(result.Steps)

	return result, nil
}

// generateMockStepResults generates mock step results for demonstration
func (s *FunnelService) generateMockStepResults(steps []Step) []StepResult {
	var results []StepResult
	var previousUsers int64 = 1000 // Start with 1000 users

	for i, step := range steps {
		// Simulate drop-off rates
		var currentUsers int64
		if i == 0 {
			currentUsers = previousUsers
		} else {
			// Simulate realistic drop-off (e.g., 70% retention between steps)
			currentUsers = int64(float64(previousUsers) * 0.7)
		}

		dropOffRate := 0.0
		if i > 0 {
			dropOffRate = float64(previousUsers-currentUsers) / float64(previousUsers) * 100
		}

		conversionRate := 0.0
		if i > 0 {
			conversionRate = float64(currentUsers) / float64(1000) * 100
		}

		stepResult := StepResult{
			StepID:         step.ID,
			StepName:       step.Name,
			EventCount:     currentUsers * 2, // Assume 2 events per user on average
			UniqueUsers:    currentUsers,
			DropOffRate:    dropOffRate,
			ConversionRate: conversionRate,
		}

		results = append(results, stepResult)
		previousUsers = currentUsers
	}

	return results
}

// calculateTotalUsers calculates the total unique users across all steps
func (s *FunnelService) calculateTotalUsers(steps []StepResult) int64 {
	if len(steps) == 0 {
		return 0
	}
	return steps[0].UniqueUsers
}

// calculateOverallConversionRate calculates the overall conversion rate from first to last step
func (s *FunnelService) calculateOverallConversionRate(steps []StepResult) float64 {
	if len(steps) < 2 {
		return 0.0
	}

	firstStepUsers := steps[0].UniqueUsers
	lastStepUsers := steps[len(steps)-1].UniqueUsers

	if firstStepUsers == 0 {
		return 0.0
	}

	return float64(lastStepUsers) / float64(firstStepUsers) * 100
}

// GetFunnelSteps returns the steps for a given funnel
func (s *FunnelService) GetFunnelSteps(ctx context.Context, funnelID string) ([]Step, error) {
	if funnelID == "" {
		return nil, fmt.Errorf("funnel ID is required")
	}

	// In a real implementation, this would fetch from a database
	// For now, return a mock funnel
	return []Step{
		{ID: "step1", Name: "Page View", EventType: "page_view", Order: 1},
		{ID: "step2", Name: "Add to Cart", EventType: "add_to_cart", Order: 2},
		{ID: "step3", Name: "Checkout", EventType: "checkout", Order: 3},
		{ID: "step4", Name: "Purchase", EventType: "purchase", Order: 4},
	}, nil
}

// generateFunnelID generates a unique funnel ID
func generateFunnelID() string {
	return fmt.Sprintf("funnel_%d", time.Now().UnixNano())
}
