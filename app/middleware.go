package app

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// APITrackingMiddleware tracks all API requests for billing purposes
type APITrackingMiddleware struct {
	analyticsService *AnalyticsService
}

// NewAPITrackingMiddleware creates a new API tracking middleware
func NewAPITrackingMiddleware(analyticsService *AnalyticsService) *APITrackingMiddleware {
	return &APITrackingMiddleware{
		analyticsService: analyticsService,
	}
}

// TrackAPIUsage is the middleware function that tracks API usage
func (m *APITrackingMiddleware) TrackAPIUsage() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Extract user information from headers or context
		userID := c.Get("X-User-ID")
		apiKey := c.Get("X-API-Key")

		// If no user ID in header, try to get from query params or generate anonymous ID
		if userID == "" {
			userID = c.Query("user_id")
			if userID == "" {
				userID = "anonymous" // Default for unauthenticated requests
			}
		}

		// If no API key, try to get from query params
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		// Create metadata for billing
		metadata := map[string]interface{}{
			"method":      c.Method(),
			"path":        c.Path(),
			"user_agent":  c.Get("User-Agent"),
			"ip_address":  c.IP(),
			"timestamp":   start,
			"api_key":     apiKey,
			"status_code": 0, // Will be updated after response
		}

		// Track the API usage asynchronously to avoid blocking the request
		go func() {
			ctx := context.Background()
			if err := m.analyticsService.TrackAPIUsage(ctx, userID, c.Path(), c.Method(), metadata); err != nil {
				log.Printf("Warning: Failed to track API usage: %v", err)
			}
		}()

		// Process the request
		err := c.Next()

		// Update metadata with response information
		metadata["status_code"] = c.Response().StatusCode()
		metadata["response_time_ms"] = time.Since(start).Milliseconds()
		metadata["response_size"] = len(c.Response().Body())

		// Track the completed request with response data
		go func() {
			ctx := context.Background()
			if err := m.analyticsService.TrackAPIUsage(ctx, userID, c.Path(), c.Method(), metadata); err != nil {
				log.Printf("Warning: Failed to track completed API usage: %v", err)
			}
		}()

		return err
	}
}

// RateLimitMiddleware implements basic rate limiting
type RateLimitMiddleware struct {
	analyticsService *AnalyticsService
	rateLimiter      *RateLimiter
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(analyticsService *AnalyticsService) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		analyticsService: analyticsService,
		rateLimiter:      NewRateLimiter(),
	}
}

// RateLimit is the middleware function that implements rate limiting
func (m *RateLimitMiddleware) RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		if userID == "" {
			userID = c.Query("user_id")
			if userID == "" {
				userID = "anonymous"
			}
		}

		// Check if user has exceeded rate limit
		if !m.rateLimiter.AllowRequest(userID, c.Path()) {
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"retry_after": 60, // Retry after 1 minute
			})
		}

		return c.Next()
	}
}

// SamplingMiddleware implements request sampling for cost control
type SamplingMiddleware struct {
	analyticsService *AnalyticsService
	sampler          *RequestSampler
}

// NewSamplingMiddleware creates a new sampling middleware
func NewSamplingMiddleware(analyticsService *AnalyticsService) *SamplingMiddleware {
	return &SamplingMiddleware{
		analyticsService: analyticsService,
		sampler:          NewRequestSampler(),
	}
}

// Sample is the middleware function that implements request sampling
func (m *SamplingMiddleware) Sample() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		if userID == "" {
			userID = c.Query("user_id")
			if userID == "" {
				userID = "anonymous"
			}
		}

		// Check if this request should be sampled
		if !m.sampler.ShouldSample(userID, c.Path()) {
			// Add sampling header to response
			c.Set("X-Sampled", "true")
		}

		return c.Next()
	}
}
