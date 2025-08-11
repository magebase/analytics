package app

import (
	"sync"
	"time"
)

// RateLimiter implements basic rate limiting per user and endpoint
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int           // Maximum requests per window
	window   time.Duration // Time window for rate limiting
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    100,                    // 100 requests per minute by default
		window:   time.Minute,            // 1 minute window
	}
}

// AllowRequest checks if a request should be allowed based on rate limiting
func (r *RateLimiter) AllowRequest(userID, endpoint string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := userID + ":" + endpoint
	now := time.Now()

	// Clean up old requests outside the window
	r.cleanupOldRequests(key, now)

	// Check if user has exceeded the limit
	if len(r.requests[key]) >= r.limit {
		return false
	}

	// Add current request
	r.requests[key] = append(r.requests[key], now)
	return true
}

// cleanupOldRequests removes requests that are outside the current window
func (r *RateLimiter) cleanupOldRequests(key string, now time.Time) {
	if requests, exists := r.requests[key]; exists {
		var validRequests []time.Time
		cutoff := now.Add(-r.window)

		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		r.requests[key] = validRequests
	}
}

// SetLimit sets the rate limit for requests
func (r *RateLimiter) SetLimit(limit int) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.limit = limit
}

// SetWindow sets the time window for rate limiting
func (r *RateLimiter) SetWindow(window time.Duration) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.window = window
}

// GetRemainingRequests returns the number of remaining requests for a user/endpoint
func (r *RateLimiter) GetRemainingRequests(userID, endpoint string) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := userID + ":" + endpoint
	now := time.Now()

	// Clean up old requests
	r.cleanupOldRequests(key, now)

	remaining := r.limit - len(r.requests[key])
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// Reset clears all rate limiting data
func (r *RateLimiter) Reset() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.requests = make(map[string][]time.Time)
}
