package app

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"
)

// HeatmapService generates heatmaps from analytics events
type HeatmapService struct {
	analyticsService *AnalyticsService
}

// Heatmap represents a heatmap visualization
type Heatmap struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "click", "scroll", "movement"
	Page        string    `json:"page"`
	Data        [][]int   `json:"data"` // 2D grid representing heatmap intensity
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// HeatmapPoint represents a single point in a heatmap
type HeatmapPoint struct {
	X         int     `json:"x"`
	Y         int     `json:"y"`
	Intensity int     `json:"intensity"`
	Weight    float64 `json:"weight"` // Additional weight factor
}

// HeatmapQuery represents a query for heatmap generation
type HeatmapQuery struct {
	Page      string    `json:"page"`
	Type      string    `json:"type"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	UserID    string    `json:"user_id,omitempty"`
	Threshold int       `json:"threshold"` // Minimum intensity to include
}

// HeatmapResult represents the computed heatmap results
type HeatmapResult struct {
	HeatmapID   string         `json:"heatmap_id"`
	HeatmapName string         `json:"heatmap_name"`
	Page        string         `json:"page"`
	Type        string         `json:"type"`
	TimeRange   TimeRange      `json:"time_range"`
	Data        [][]int        `json:"data"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Points      []HeatmapPoint `json:"points"`
	Stats       HeatmapStats   `json:"stats"`
	ComputedAt  time.Time      `json:"computed_at"`
}

// HeatmapStats represents statistics about the heatmap
type HeatmapStats struct {
	TotalPoints  int     `json:"total_points"`
	MaxIntensity int     `json:"max_intensity"`
	MinIntensity int     `json:"min_intensity"`
	AvgIntensity float64 `json:"avg_intensity"`
	HotspotCount int     `json:"hotspot_count"`
	CoverageArea float64 `json:"coverage_area"` // Percentage of area with activity
}

// NewHeatmapService creates a new heatmap service instance
func NewHeatmapService(analyticsService *AnalyticsService) *HeatmapService {
	return &HeatmapService{
		analyticsService: analyticsService,
	}
}

// CreateHeatmap creates a new heatmap
func (s *HeatmapService) CreateHeatmap(ctx context.Context, name, description, heatmapType, page string, width, height int) (*Heatmap, error) {
	if name == "" {
		return nil, fmt.Errorf("heatmap name is required")
	}

	if page == "" {
		return nil, fmt.Errorf("page is required")
	}

	if heatmapType == "" {
		return nil, fmt.Errorf("heatmap type is required")
	}

	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("width and height must be positive")
	}

	// Validate heatmap type
	validTypes := map[string]bool{"click": true, "scroll": true, "movement": true}
	if !validTypes[heatmapType] {
		return nil, fmt.Errorf("invalid heatmap type: %s. Valid types are: click, scroll, movement", heatmapType)
	}

	heatmap := &Heatmap{
		ID:          generateHeatmapID(),
		Name:        name,
		Description: description,
		Type:        heatmapType,
		Page:        page,
		Width:       width,
		Height:      height,
		Data:        make([][]int, height),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Initialize the 2D grid
	for i := range heatmap.Data {
		heatmap.Data[i] = make([]int, width)
	}

	log.Printf("Created heatmap: %s for page: %s, type: %s, dimensions: %dx%d",
		heatmap.ID, page, heatmapType, width, height)

	return heatmap, nil
}

// GenerateHeatmap generates a heatmap from analytics events
func (s *HeatmapService) GenerateHeatmap(ctx context.Context, query HeatmapQuery) (*HeatmapResult, error) {
	if query.Page == "" {
		return nil, fmt.Errorf("page is required")
	}

	if query.Type == "" {
		return nil, fmt.Errorf("heatmap type is required")
	}

	if query.Width <= 0 || query.Height <= 0 {
		return nil, fmt.Errorf("width and height must be positive")
	}

	// Set default dimensions if not provided
	if query.Width == 0 {
		query.Width = 1920 // Default desktop width
	}
	if query.Height == 0 {
		query.Height = 1080 // Default desktop height
	}

	// Generate mock heatmap data for demonstration
	// In production, this would query the analytics database for real event data
	heatmapData, points := s.generateMockHeatmapData(query)

	result := &HeatmapResult{
		HeatmapID:   generateHeatmapID(),
		HeatmapName: fmt.Sprintf("%s Heatmap - %s", query.Type, query.Page),
		Page:        query.Page,
		Type:        query.Type,
		TimeRange:   TimeRange{Start: query.Start, End: query.End},
		Data:        heatmapData,
		Width:       query.Width,
		Height:      query.Height,
		Points:      points,
		Stats:       s.calculateHeatmapStats(heatmapData, points),
		ComputedAt:  time.Now(),
	}

	return result, nil
}

// generateMockHeatmapData generates mock heatmap data for demonstration
func (s *HeatmapService) generateMockHeatmapData(query HeatmapQuery) ([][]int, []HeatmapPoint) {
	width := query.Width
	height := query.Height

	// Create 2D grid
	data := make([][]int, height)
	for i := range data {
		data[i] = make([]int, width)
	}

	var points []HeatmapPoint

	// Generate mock click points for demonstration
	// In production, this would be based on real event data
	switch query.Type {
	case "click":
		points = s.generateMockClickPoints(width, height)
	case "scroll":
		points = s.generateMockScrollPoints(width, height)
	case "movement":
		points = s.generateMockMovementPoints(width, height)
	}

	// Apply points to the grid
	for _, point := range points {
		// Ensure coordinates are within bounds
		if point.X >= 0 && point.X < width && point.Y >= 0 && point.Y < height {
			data[point.Y][point.X] += point.Intensity
		}
	}

	// Apply Gaussian blur for more realistic heatmap appearance
	data = s.applyGaussianBlur(data, 3)

	return data, points
}

// generateMockClickPoints generates mock click points
func (s *HeatmapService) generateMockClickPoints(width, height int) []HeatmapPoint {
	var points []HeatmapPoint

	// Generate some realistic click patterns
	clickAreas := []struct {
		x, y, radius int
		intensity    int
	}{
		{width / 4, height / 4, 50, 10},        // Top-left area
		{width / 2, height / 2, 80, 15},        // Center area
		{3 * width / 4, 3 * height / 4, 60, 8}, // Bottom-right area
		{width / 2, height / 4, 40, 12},        // Top-center area
	}

	for _, area := range clickAreas {
		// Generate multiple clicks in each area
		for i := 0; i < 20; i++ {
			x := area.x + (i%area.radius - area.radius/2)
			y := area.y + (i%area.radius - area.radius/2)

			// Add some randomness
			x += (i*7)%20 - 10
			y += (i*11)%20 - 10

			points = append(points, HeatmapPoint{
				X:         x,
				Y:         y,
				Intensity: area.intensity,
				Weight:    1.0,
			})
		}
	}

	return points
}

// generateMockScrollPoints generates mock scroll points
func (s *HeatmapService) generateMockScrollPoints(width, height int) []HeatmapPoint {
	var points []HeatmapPoint

	// Scroll patterns are typically more vertical
	for y := 0; y < height; y += 50 {
		for x := 0; x < width; x += 100 {
			intensity := 5 + (y % 20) // Higher intensity in middle areas
			points = append(points, HeatmapPoint{
				X:         x,
				Y:         y,
				Intensity: intensity,
				Weight:    0.8,
			})
		}
	}

	return points
}

// generateMockMovementPoints generates mock mouse movement points
func (s *HeatmapService) generateMockMovementPoints(width, height int) []HeatmapPoint {
	var points []HeatmapPoint

	// Generate a path-like pattern
	path := []struct {
		x, y int
	}{
		{0, height / 2},
		{width / 4, height / 3},
		{width / 2, height / 2},
		{3 * width / 4, 2 * height / 3},
		{width, height / 2},
	}

	for i := 0; i < len(path)-1; i++ {
		start := path[i]
		end := path[i+1]

		// Generate points along the path
		steps := 50
		for j := 0; j <= steps; j++ {
			t := float64(j) / float64(steps)
			x := int(float64(start.x)*(1-t) + float64(end.x)*t)
			y := int(float64(start.y)*(1-t) + float64(end.y)*t)

			// Add some randomness
			x += (j*13)%30 - 15
			y += (j*17)%30 - 15

			points = append(points, HeatmapPoint{
				X:         x,
				Y:         y,
				Intensity: 3,
				Weight:    0.6,
			})
		}
	}

	return points
}

// applyGaussianBlur applies a simple Gaussian blur to the heatmap data
func (s *HeatmapService) applyGaussianBlur(data [][]int, radius int) [][]int {
	height := len(data)
	width := len(data[0])

	// Create a copy for the blurred result
	blurred := make([][]int, height)
	for i := range blurred {
		blurred[i] = make([]int, width)
	}

	// Simple blur implementation
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sum := 0
			count := 0

			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					ny, nx := y+dy, x+dx
					if ny >= 0 && ny < height && nx >= 0 && nx < width {
						sum += data[ny][nx]
						count++
					}
				}
			}

			if count > 0 {
				blurred[y][x] = sum / count
			}
		}
	}

	return blurred
}

// calculateHeatmapStats calculates statistics for the heatmap
func (s *HeatmapService) calculateHeatmapStats(data [][]int, points []HeatmapPoint) HeatmapStats {
	height := len(data)
	width := len(data[0])

	var totalIntensity int
	maxIntensity := 0
	minIntensity := math.MaxInt32
	activeCells := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			intensity := data[y][x]
			totalIntensity += intensity

			if intensity > maxIntensity {
				maxIntensity = intensity
			}
			if intensity < minIntensity {
				minIntensity = intensity
			}
			if intensity > 0 {
				activeCells++
			}
		}
	}

	totalCells := width * height
	coverageArea := 0.0
	if totalCells > 0 {
		coverageArea = float64(activeCells) / float64(totalCells) * 100
	}

	avgIntensity := 0.0
	if totalCells > 0 {
		avgIntensity = float64(totalIntensity) / float64(totalCells)
	}

	// Count hotspots (areas with high intensity)
	hotspotCount := 0
	hotspotThreshold := maxIntensity / 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if data[y][x] > hotspotThreshold {
				hotspotCount++
			}
		}
	}

	return HeatmapStats{
		TotalPoints:  len(points),
		MaxIntensity: maxIntensity,
		MinIntensity: minIntensity,
		AvgIntensity: avgIntensity,
		HotspotCount: hotspotCount,
		CoverageArea: coverageArea,
	}
}

// GetHeatmap retrieves a heatmap by ID
func (s *HeatmapService) GetHeatmap(ctx context.Context, heatmapID string) (*Heatmap, error) {
	if heatmapID == "" {
		return nil, fmt.Errorf("heatmap ID is required")
	}

	// In a real implementation, this would fetch from a database
	// For now, return a mock heatmap
	return &Heatmap{
		ID:          heatmapID,
		Name:        "Sample Heatmap",
		Description: "A sample heatmap for demonstration",
		Type:        "click",
		Page:        "/home",
		Width:       1920,
		Height:      1080,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// generateHeatmapID generates a unique heatmap ID
func generateHeatmapID() string {
	return fmt.Sprintf("heatmap_%d", time.Now().UnixNano())
}
