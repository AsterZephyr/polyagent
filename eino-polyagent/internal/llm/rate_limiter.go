package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(key string) bool
	
	// Reserve reserves capacity for a request
	Reserve(key string) *Reservation
	
	// Wait waits until the request is allowed
	Wait(ctx context.Context, key string) error
	
	// GetStats returns rate limiter statistics
	GetStats(key string) *RateLimitStats
	
	// Reset resets the rate limiter for a key
	Reset(key string)
}

// Reservation represents a reservation for future execution
type Reservation struct {
	OK        bool
	Delay     time.Duration
	TimeToAct time.Time
}

// RateLimitStats provides statistics about rate limiting
type RateLimitStats struct {
	Key            string    `json:"key"`
	RequestCount   int64     `json:"request_count"`
	AllowedCount   int64     `json:"allowed_count"`
	DroppedCount   int64     `json:"dropped_count"`
	CurrentRate    float64   `json:"current_rate"`
	BurstCapacity  int       `json:"burst_capacity"`
	LastRequest    time.Time `json:"last_request"`
	WindowStart    time.Time `json:"window_start"`
}

// TokenBucketLimiter implements rate limiting using token bucket algorithm
type TokenBucketLimiter struct {
	buckets map[string]*TokenBucket
	config  *RateLimitConfig
	logger  *logrus.Logger
	mu      sync.RWMutex
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	stats      *RateLimitStats
	mu         sync.Mutex
}

// RateLimitConfig configures rate limiting behavior
type RateLimitConfig struct {
	DefaultRate     float64       `json:"default_rate"`     // requests per second
	DefaultBurst    int           `json:"default_burst"`    // burst capacity
	CleanupInterval time.Duration `json:"cleanup_interval"` // cleanup interval for unused buckets
	KeyRules        map[string]*RateLimitRule `json:"key_rules"` // per-key rate limit rules
}

// RateLimitRule defines rate limiting rules for specific keys
type RateLimitRule struct {
	Rate  float64 `json:"rate"`  // requests per second
	Burst int     `json:"burst"` // burst capacity
}

// RetryManager manages retry logic with exponential backoff
type RetryManager struct {
	config *RetryConfig
	logger *logrus.Logger
}

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffMultiple float64       `json:"backoff_multiple"`
	Jitter          bool          `json:"jitter"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// FailoverManager manages failover between providers
type FailoverManager struct {
	providers []LLMProvider
	config    *FailoverConfig
	health    map[LLMProvider]*HealthStatus
	logger    *logrus.Logger
	mu        sync.RWMutex
}

// FailoverConfig configures failover behavior
type FailoverConfig struct {
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	FailureThreshold    int           `json:"failure_threshold"`
	RecoveryThreshold   int           `json:"recovery_threshold"`
	CircuitBreakerEnabled bool        `json:"circuit_breaker_enabled"`
	FallbackProvider    LLMProvider   `json:"fallback_provider"`
}

// HealthStatus tracks provider health
type HealthStatus struct {
	Provider        LLMProvider   `json:"provider"`
	IsHealthy       bool          `json:"is_healthy"`
	ConsecutiveFailures int       `json:"consecutive_failures"`
	ConsecutiveSuccess  int       `json:"consecutive_success"`
	LastCheck       time.Time     `json:"last_check"`
	LastError       string        `json:"last_error,omitempty"`
	ResponseTime    time.Duration `json:"response_time"`
	CircuitOpen     bool          `json:"circuit_open"`
}

// NewTokenBucketLimiter creates a new token bucket rate limiter
func NewTokenBucketLimiter(config *RateLimitConfig, logger *logrus.Logger) *TokenBucketLimiter {
	if config == nil {
		config = createDefaultRateLimitConfig()
	}

	limiter := &TokenBucketLimiter{
		buckets: make(map[string]*TokenBucket),
		config:  config,
		logger:  logger,
	}

	// Start cleanup routine
	go limiter.cleanup()

	logger.Info("Token bucket rate limiter initialized")
	return limiter
}

// Allow checks if a request is allowed
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	reservation := tbl.Reserve(key)
	return reservation.OK && reservation.Delay == 0
}

// Reserve reserves capacity for a request
func (tbl *TokenBucketLimiter) Reserve(key string) *Reservation {
	bucket := tbl.getBucket(key)
	
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Update statistics
	bucket.stats.RequestCount++
	bucket.stats.LastRequest = time.Now()

	// Refill tokens
	bucket.refill()

	if bucket.tokens >= 1.0 {
		bucket.tokens--
		bucket.stats.AllowedCount++
		return &Reservation{
			OK:        true,
			Delay:     0,
			TimeToAct: time.Now(),
		}
	}

	// Calculate wait time
	tokensNeeded := 1.0 - bucket.tokens
	delay := time.Duration(tokensNeeded / bucket.refillRate * float64(time.Second))

	bucket.stats.DroppedCount++
	
	return &Reservation{
		OK:        false,
		Delay:     delay,
		TimeToAct: time.Now().Add(delay),
	}
}

// Wait waits until the request is allowed
func (tbl *TokenBucketLimiter) Wait(ctx context.Context, key string) error {
	reservation := tbl.Reserve(key)
	
	if reservation.OK && reservation.Delay == 0 {
		return nil
	}

	if !reservation.OK {
		// Wait for the calculated delay
		select {
		case <-time.After(reservation.Delay):
			// Try again after waiting
			if tbl.Allow(key) {
				return nil
			}
			return fmt.Errorf("rate limit exceeded for key: %s", key)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// GetStats returns rate limiter statistics
func (tbl *TokenBucketLimiter) GetStats(key string) *RateLimitStats {
	bucket := tbl.getBucket(key)
	
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// Calculate current rate
	elapsed := time.Since(bucket.stats.WindowStart).Seconds()
	if elapsed > 0 {
		bucket.stats.CurrentRate = float64(bucket.stats.AllowedCount) / elapsed
	}

	// Return a copy
	statsCopy := *bucket.stats
	return &statsCopy
}

// Reset resets the rate limiter for a key
func (tbl *TokenBucketLimiter) Reset(key string) {
	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	if bucket, exists := tbl.buckets[key]; exists {
		bucket.mu.Lock()
		bucket.tokens = bucket.capacity
		bucket.lastRefill = time.Now()
		bucket.stats = &RateLimitStats{
			Key:           key,
			BurstCapacity: int(bucket.capacity),
			WindowStart:   time.Now(),
		}
		bucket.mu.Unlock()
	}

	tbl.logger.WithField("key", key).Debug("Rate limiter reset")
}

// getBucket gets or creates a token bucket for a key
func (tbl *TokenBucketLimiter) getBucket(key string) *TokenBucket {
	tbl.mu.RLock()
	if bucket, exists := tbl.buckets[key]; exists {
		tbl.mu.RUnlock()
		return bucket
	}
	tbl.mu.RUnlock()

	tbl.mu.Lock()
	defer tbl.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := tbl.buckets[key]; exists {
		return bucket
	}

	// Get rate and burst for this key
	rate := tbl.config.DefaultRate
	burst := tbl.config.DefaultBurst

	if rule, exists := tbl.config.KeyRules[key]; exists {
		rate = rule.Rate
		burst = rule.Burst
	}

	bucket := &TokenBucket{
		tokens:     float64(burst),
		capacity:   float64(burst),
		refillRate: rate,
		lastRefill: time.Now(),
		stats: &RateLimitStats{
			Key:           key,
			BurstCapacity: burst,
			WindowStart:   time.Now(),
		},
	}

	tbl.buckets[key] = bucket
	return bucket
}

// refill refills tokens in the bucket
func (bucket *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	
	tokensToAdd := elapsed * bucket.refillRate
	bucket.tokens = min(bucket.capacity, bucket.tokens + tokensToAdd)
	bucket.lastRefill = now
}

// cleanup removes unused buckets periodically
func (tbl *TokenBucketLimiter) cleanup() {
	ticker := time.NewTicker(tbl.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tbl.mu.Lock()
		now := time.Now()
		
		for key, bucket := range tbl.buckets {
			bucket.mu.Lock()
			// Remove buckets that haven't been used for twice the cleanup interval
			if now.Sub(bucket.stats.LastRequest) > 2*tbl.config.CleanupInterval {
				delete(tbl.buckets, key)
				tbl.logger.WithField("key", key).Debug("Removed unused rate limiter bucket")
			}
			bucket.mu.Unlock()
		}
		
		tbl.mu.Unlock()
	}
}

// NewRetryManager creates a new retry manager
func NewRetryManager(config *RetryConfig, logger *logrus.Logger) *RetryManager {
	if config == nil {
		config = createDefaultRetryConfig()
	}

	return &RetryManager{
		config: config,
		logger: logger,
	}
}

// ExecuteWithRetry executes a function with retry logic
func (rm *RetryManager) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= rm.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := rm.calculateDelay(attempt)
			
			rm.logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"delay":   delay,
			}).Debug("Retrying after delay")

			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				rm.logger.WithField("attempts", attempt+1).Info("Operation succeeded after retries")
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.logger.WithError(err).Debug("Non-retryable error, stopping retries")
			return err
		}

		rm.logger.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"error":   err,
		}).Debug("Operation failed, will retry")
	}

	rm.logger.WithFields(logrus.Fields{
		"max_retries": rm.config.MaxRetries,
		"last_error":  lastErr,
	}).Error("All retry attempts failed")

	return fmt.Errorf("operation failed after %d attempts: %w", rm.config.MaxRetries+1, lastErr)
}

// calculateDelay calculates delay with exponential backoff
func (rm *RetryManager) calculateDelay(attempt int) time.Duration {
	delay := time.Duration(float64(rm.config.InitialDelay) * pow(rm.config.BackoffMultiple, float64(attempt-1)))
	
	if delay > rm.config.MaxDelay {
		delay = rm.config.MaxDelay
	}

	// Add jitter if enabled
	if rm.config.Jitter {
		jitterAmount := time.Duration(float64(delay) * 0.1) // 10% jitter
		delay += time.Duration(float64(jitterAmount) * (2*float64(time.Now().UnixNano()%1000)/1000 - 1))
	}

	return delay
}

// isRetryableError checks if an error is retryable
func (rm *RetryManager) isRetryableError(err error) bool {
	if len(rm.config.RetryableErrors) == 0 {
		// Default retryable conditions
		errStr := err.Error()
		return contains(errStr, []string{
			"timeout",
			"connection reset",
			"connection refused",
			"temporary failure",
			"rate limit",
			"429", // Too Many Requests
			"502", // Bad Gateway
			"503", // Service Unavailable
			"504", // Gateway Timeout
		})
	}

	errStr := err.Error()
	for _, retryableErr := range rm.config.RetryableErrors {
		if contains(errStr, []string{retryableErr}) {
			return true
		}
	}

	return false
}

// NewFailoverManager creates a new failover manager
func NewFailoverManager(providers []LLMProvider, config *FailoverConfig, logger *logrus.Logger) *FailoverManager {
	if config == nil {
		config = createDefaultFailoverConfig()
	}

	fm := &FailoverManager{
		providers: providers,
		config:    config,
		health:    make(map[LLMProvider]*HealthStatus),
		logger:    logger,
	}

	// Initialize health status for all providers
	for _, provider := range providers {
		fm.health[provider] = &HealthStatus{
			Provider:  provider,
			IsHealthy: true,
			LastCheck: time.Now(),
		}
	}

	// Start health check routine
	if config.HealthCheckInterval > 0 {
		go fm.healthCheck()
	}

	logger.WithField("providers", len(providers)).Info("Failover manager initialized")
	return fm
}

// GetHealthyProvider returns the first healthy provider
func (fm *FailoverManager) GetHealthyProvider() (LLMProvider, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, provider := range fm.providers {
		if status := fm.health[provider]; status.IsHealthy && !status.CircuitOpen {
			return provider, nil
		}
	}

	// If no healthy provider, try fallback
	if fm.config.FallbackProvider != "" {
		return fm.config.FallbackProvider, nil
	}

	return "", fmt.Errorf("no healthy providers available")
}

// ReportSuccess reports successful request for a provider
func (fm *FailoverManager) ReportSuccess(provider LLMProvider, responseTime time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if status, exists := fm.health[provider]; exists {
		status.ConsecutiveFailures = 0
		status.ConsecutiveSuccess++
		status.ResponseTime = responseTime
		status.LastCheck = time.Now()
		status.LastError = ""

		// Recover if threshold met
		if status.ConsecutiveSuccess >= fm.config.RecoveryThreshold {
			if !status.IsHealthy || status.CircuitOpen {
				fm.logger.WithField("provider", provider).Info("Provider recovered")
				status.IsHealthy = true
				status.CircuitOpen = false
			}
		}
	}
}

// ReportFailure reports failed request for a provider
func (fm *FailoverManager) ReportFailure(provider LLMProvider, err error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if status, exists := fm.health[provider]; exists {
		status.ConsecutiveSuccess = 0
		status.ConsecutiveFailures++
		status.LastCheck = time.Now()
		status.LastError = err.Error()

		// Mark unhealthy if threshold exceeded
		if status.ConsecutiveFailures >= fm.config.FailureThreshold {
			if status.IsHealthy {
				fm.logger.WithFields(logrus.Fields{
					"provider": provider,
					"failures": status.ConsecutiveFailures,
				}).Warn("Provider marked as unhealthy")
				status.IsHealthy = false
			}

			// Open circuit breaker if enabled
			if fm.config.CircuitBreakerEnabled && !status.CircuitOpen {
				fm.logger.WithField("provider", provider).Warn("Circuit breaker opened")
				status.CircuitOpen = true
			}
		}
	}
}

// GetHealthStatus returns health status for all providers
func (fm *FailoverManager) GetHealthStatus() map[LLMProvider]*HealthStatus {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	// Return a copy
	healthCopy := make(map[LLMProvider]*HealthStatus)
	for provider, status := range fm.health {
		statusCopy := *status
		healthCopy[provider] = &statusCopy
	}

	return healthCopy
}

// healthCheck performs periodic health checks
func (fm *FailoverManager) healthCheck() {
	ticker := time.NewTicker(fm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		fm.mu.Lock()
		for provider, status := range fm.health {
			// Simple health check logic
			// In a real implementation, this would ping the provider
			if !status.IsHealthy && time.Since(status.LastCheck) > 5*time.Minute {
				// Try to recover after 5 minutes
				status.ConsecutiveFailures = max(0, status.ConsecutiveFailures-1)
				
				if status.ConsecutiveFailures < fm.config.FailureThreshold {
					status.IsHealthy = true
					status.CircuitOpen = false
					fm.logger.WithField("provider", provider).Info("Provider health check recovery")
				}
			}
		}
		fm.mu.Unlock()
	}
}

// Utility functions
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

func contains(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(substr) > 0 && len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// Default configurations
func createDefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		DefaultRate:     10.0,                // 10 requests per second
		DefaultBurst:    20,                  // burst of 20 requests
		CleanupInterval: 5 * time.Minute,     // cleanup every 5 minutes
		KeyRules: map[string]*RateLimitRule{
			"openai":    {Rate: 50.0, Burst: 100},
			"claude":    {Rate: 30.0, Burst: 60},
			"fallback":  {Rate: 5.0, Burst: 10},
		},
	}
}

func createDefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:      3,
		InitialDelay:    time.Second,
		MaxDelay:        30 * time.Second,
		BackoffMultiple: 2.0,
		Jitter:          true,
		RetryableErrors: []string{
			"timeout",
			"connection reset",
			"429",
			"502",
			"503",
			"504",
		},
	}
}

func createDefaultFailoverConfig() *FailoverConfig {
	return &FailoverConfig{
		HealthCheckInterval:   time.Minute,
		FailureThreshold:      3,
		RecoveryThreshold:     2,
		CircuitBreakerEnabled: true,
		FallbackProvider:      "",
	}
}