package app

import (
	"crypto/md5"
	"math/rand"
	"sync"
	"time"
)

// RequestSampler implements request sampling for cost control
type RequestSampler struct {
	sampleRates map[string]float64 // Sample rate per endpoint (0.0 to 1.0)
	mutex       sync.RWMutex
	rand        *rand.Rand
}

// NewRequestSampler creates a new request sampler instance
func NewRequestSampler() *RequestSampler {
	return &RequestSampler{
		sampleRates: make(map[string]float64),
		rand:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ShouldSample determines if a request should be sampled based on user ID and endpoint
func (s *RequestSampler) ShouldSample(userID, endpoint string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Get sample rate for this endpoint (default to 1.0 = 100% sampling)
	sampleRate := s.sampleRates[endpoint]
	if sampleRate == 0 {
		sampleRate = 1.0 // Default to 100% sampling
	}

	// If 100% sampling, always return true
	if sampleRate >= 1.0 {
		return true
	}

	// If 0% sampling, never return true
	if sampleRate <= 0.0 {
		return false
	}

	// Generate a deterministic hash for the user ID to ensure consistent sampling
	hash := md5.Sum([]byte(userID + ":" + endpoint))
	hashInt := int(hash[0]) + int(hash[1])*256 + int(hash[2])*65536 + int(hash[3])*16777216

	// Use the hash to determine if this request should be sampled
	// This ensures the same user always gets the same sampling decision for the same endpoint
	randomValue := float64(hashInt%10000) / 10000.0

	return randomValue < sampleRate
}

// SetSampleRate sets the sampling rate for a specific endpoint
func (s *RequestSampler) SetSampleRate(endpoint string, rate float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clamp rate between 0.0 and 1.0
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}

	s.sampleRates[endpoint] = rate
}

// GetSampleRate gets the current sampling rate for an endpoint
func (s *RequestSampler) GetSampleRate(endpoint string) float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	rate, exists := s.sampleRates[endpoint]
	if !exists {
		return 1.0 // Default to 100% sampling
	}
	return rate
}

// SetGlobalSampleRate sets the sampling rate for all endpoints
func (s *RequestSampler) SetGlobalSampleRate(rate float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Clamp rate between 0.0 and 1.0
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}

	// Set for all existing endpoints
	for endpoint := range s.sampleRates {
		s.sampleRates[endpoint] = rate
	}
}

// GetSamplingStats returns sampling statistics for monitoring
func (s *RequestSampler) GetSamplingStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	stats := make(map[string]interface{})
	for endpoint, rate := range s.sampleRates {
		stats[endpoint] = map[string]interface{}{
			"sample_rate": rate,
			"percentage":  rate * 100,
		}
	}

	return stats
}

// Reset clears all sampling configuration
func (s *RequestSampler) Reset() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sampleRates = make(map[string]float64)
}
