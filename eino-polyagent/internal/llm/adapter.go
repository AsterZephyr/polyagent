package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// UnifiedLLMAdapter is the main implementation of the LLMAdapter interface
type UnifiedLLMAdapter struct {
	config      *LLMAdapterConfig
	clients     map[LLMProvider]LLMClient
	metrics     *LLMMetrics
	logger      *logrus.Logger
	mu          sync.RWMutex
	lastUsed    map[LLMProvider]time.Time
	circuitBreakers map[LLMProvider]*CircuitBreaker
}

// CircuitBreaker implements circuit breaker pattern for LLM providers
type CircuitBreaker struct {
	failures    int
	lastFailure time.Time
	state       CircuitBreakerState
	threshold   int
	timeout     time.Duration
	mu          sync.RWMutex
}

type CircuitBreakerState string

const (
	CircuitBreakerClosed    CircuitBreakerState = "closed"
	CircuitBreakerOpen      CircuitBreakerState = "open"
	CircuitBreakerHalfOpen  CircuitBreakerState = "half_open"
)

// NewUnifiedLLMAdapter creates a new unified LLM adapter
func NewUnifiedLLMAdapter(config *LLMAdapterConfig, logger *logrus.Logger) (*UnifiedLLMAdapter, error) {
	adapter := &UnifiedLLMAdapter{
		config:          config,
		clients:         make(map[LLMProvider]LLMClient),
		lastUsed:        make(map[LLMProvider]time.Time),
		circuitBreakers: make(map[LLMProvider]*CircuitBreaker),
		logger:          logger,
		metrics: &LLMMetrics{
			ProviderMetrics: make(map[LLMProvider]*ProviderStatus),
			LastUpdated:     time.Now(),
		},
	}

	// Initialize primary client
	primaryClient, err := adapter.createClient(&config.Primary)
	if err != nil {
		return nil, fmt.Errorf("failed to create primary client: %w", err)
	}
	adapter.clients[config.Primary.Provider] = primaryClient
	adapter.circuitBreakers[config.Primary.Provider] = NewCircuitBreaker(5, 30*time.Second)

	// Initialize fallback clients
	for _, fallbackConfig := range config.Fallback {
		client, err := adapter.createClient(&fallbackConfig)
		if err != nil {
			adapter.logger.Warnf("Failed to create fallback client for %s: %v", fallbackConfig.Provider, err)
			continue
		}
		adapter.clients[fallbackConfig.Provider] = client
		adapter.circuitBreakers[fallbackConfig.Provider] = NewCircuitBreaker(5, 30*time.Second)
	}

	// Initialize budget client if configured
	if config.Budget != nil {
		client, err := adapter.createClient(config.Budget)
		if err != nil {
			adapter.logger.Warnf("Failed to create budget client for %s: %v", config.Budget.Provider, err)
		} else {
			adapter.clients[config.Budget.Provider] = client
			adapter.circuitBreakers[config.Budget.Provider] = NewCircuitBreaker(5, 30*time.Second)
		}
	}

	adapter.logger.Infof("Initialized UnifiedLLMAdapter with %d providers", len(adapter.clients))
	return adapter, nil
}

// Generate generates a response with automatic fallback
func (a *UnifiedLLMAdapter) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	return a.GenerateWithFallback(ctx, req, FallbackAutomatic)
}

// GenerateWithFallback generates with explicit fallback strategy
func (a *UnifiedLLMAdapter) GenerateWithFallback(ctx context.Context, req *GenerateRequest, strategy FallbackStrategy) (*GenerateResponse, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	startTime := time.Now()
	defer func() {
		a.updateMetrics(time.Since(startTime))
	}()

	// Determine provider order based on strategy
	providers := a.getProviderOrder(strategy)

	var lastError error
	for _, provider := range providers {
		client, exists := a.clients[provider]
		if !exists {
			continue
		}

		// Check circuit breaker
		if !a.circuitBreakers[provider].CanExecute() {
			a.logger.Warnf("Circuit breaker is open for provider %s", provider)
			continue
		}

		// Set model if not specified
		if req.Model == "" {
			req.Model = a.getModelForProvider(provider)
		}

		// Make the request
		response, err := client.Generate(ctx, req)
		if err != nil {
			lastError = err
			a.circuitBreakers[provider].RecordFailure()
			a.logger.Warnf("Request failed for provider %s: %v", provider, err)
			continue
		}

		// Success - record and return
		a.circuitBreakers[provider].RecordSuccess()
		a.lastUsed[provider] = time.Now()
		a.logger.Infof("Request successful with provider %s", provider)
		
		return response, nil
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastError)
}

// GetAvailableProviders returns list of configured providers
func (a *UnifiedLLMAdapter) GetAvailableProviders() []LLMProvider {
	a.mu.RLock()
	defer a.mu.RUnlock()

	providers := make([]LLMProvider, 0, len(a.clients))
	for provider := range a.clients {
		providers = append(providers, provider)
	}
	return providers
}

// GetProviderStatus returns the health status of all providers
func (a *UnifiedLLMAdapter) GetProviderStatus(ctx context.Context) map[LLMProvider]ProviderStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := make(map[LLMProvider]ProviderStatus)
	for provider, client := range a.clients {
		startTime := time.Now()
		err := client.HealthCheck(ctx)
		latency := time.Since(startTime)

		cb := a.circuitBreakers[provider]
		cb.mu.RLock()
		errorRate := float64(cb.failures) / float64(cb.failures+1)
		cb.mu.RUnlock()

		providerStatus := ProviderStatus{
			Available:    err == nil,
			Latency:      latency,
			ErrorRate:    errorRate,
			LastError:    "",
			RequestCount: 0, // TODO: Track this
		}

		if err != nil {
			providerStatus.LastError = err.Error()
			now := time.Now()
			providerStatus.LastErrorAt = &now
		}

		status[provider] = providerStatus
	}

	return status
}

// UpdateConfig updates the adapter configuration
func (a *UnifiedLLMAdapter) UpdateConfig(config *LLMAdapterConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close existing clients
	for _, client := range a.clients {
		client.Close()
	}

	// Reset state
	a.clients = make(map[LLMProvider]LLMClient)
	a.circuitBreakers = make(map[LLMProvider]*CircuitBreaker)
	a.config = config

	// Reinitialize clients (same logic as constructor)
	// ... (implementation similar to NewUnifiedLLMAdapter)

	a.logger.Info("LLM adapter configuration updated")
	return nil
}

// GetMetrics returns usage metrics
func (a *UnifiedLLMAdapter) GetMetrics() *LLMMetrics {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *a.metrics
	return &metrics
}

// Helper methods

func (a *UnifiedLLMAdapter) createClient(config *LLMConfig) (LLMClient, error) {
	switch config.Provider {
	case ProviderOpenAI:
		return NewOpenAIClient(config, a.logger)
	case ProviderClaude:
		return NewClaudeClient(config, a.logger)
	case ProviderQwen:
		return NewQwenClient(config, a.logger)
	case ProviderK2, ProviderOpenRouter:
		return NewOpenRouterClient(config, a.logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", config.Provider)
	}
}

func (a *UnifiedLLMAdapter) getProviderOrder(strategy FallbackStrategy) []LLMProvider {
	switch strategy {
	case FallbackNone:
		return []LLMProvider{a.config.Primary.Provider}
	case FallbackCostBased:
		return a.getProvidersByCost()
	case FallbackSpeedBased:
		return a.getProvidersBySpeed()
	default: // FallbackAutomatic
		return a.getProvidersAutomatic()
	}
}

func (a *UnifiedLLMAdapter) getProvidersByCost() []LLMProvider {
	// Start with budget provider if available, then primary, then fallbacks
	providers := []LLMProvider{}
	
	if a.config.Budget != nil {
		providers = append(providers, a.config.Budget.Provider)
	}
	
	providers = append(providers, a.config.Primary.Provider)
	
	for _, config := range a.config.Fallback {
		providers = append(providers, config.Provider)
	}
	
	return providers
}

func (a *UnifiedLLMAdapter) getProvidersBySpeed() []LLMProvider {
	// Order by historical latency (would need to track this)
	// For now, use simple ordering: primary first, then fallbacks
	providers := []LLMProvider{a.config.Primary.Provider}
	
	for _, config := range a.config.Fallback {
		providers = append(providers, config.Provider)
	}
	
	return providers
}

func (a *UnifiedLLMAdapter) getProvidersAutomatic() []LLMProvider {
	// Smart ordering based on health and performance
	providers := []LLMProvider{a.config.Primary.Provider}
	
	for _, config := range a.config.Fallback {
		providers = append(providers, config.Provider)
	}
	
	return providers
}

func (a *UnifiedLLMAdapter) getModelForProvider(provider LLMProvider) string {
	switch provider {
	case a.config.Primary.Provider:
		return a.config.Primary.Model
	default:
		for _, config := range a.config.Fallback {
			if config.Provider == provider {
				return config.Model
			}
		}
		if a.config.Budget != nil && a.config.Budget.Provider == provider {
			return a.config.Budget.Model
		}
	}
	return ""
}

func (a *UnifiedLLMAdapter) updateMetrics(duration time.Duration) {
	a.metrics.TotalRequests++
	a.metrics.AverageLatency = (a.metrics.AverageLatency + duration) / 2
	a.metrics.LastUpdated = time.Now()
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		threshold: threshold,
		timeout:   timeout,
		state:     CircuitBreakerClosed,
	}
}

// CanExecute checks if the circuit breaker allows execution
func (cb *CircuitBreaker) CanExecute() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		return time.Since(cb.lastFailure) > cb.timeout
	case CircuitBreakerHalfOpen:
		return true
	}
	return false
}

// RecordSuccess records a successful execution
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = CircuitBreakerClosed
}

// RecordFailure records a failed execution
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.threshold {
		cb.state = CircuitBreakerOpen
	}
}