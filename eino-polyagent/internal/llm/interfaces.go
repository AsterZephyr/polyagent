package llm

import (
	"context"
	"time"
)

// LLMProvider represents different LLM providers
type LLMProvider string

const (
	ProviderOpenAI     LLMProvider = "openai"
	ProviderClaude     LLMProvider = "claude"
	ProviderQwen       LLMProvider = "qwen"
	ProviderK2         LLMProvider = "k2"
	ProviderOpenRouter LLMProvider = "openrouter"
)

// Message represents a chat message
type Message struct {
	Role    string                 `json:"role"`    // "system", "user", "assistant", "tool"
	Content string                 `json:"content"`
	Name    string                 `json:"name,omitempty"`
	ToolCalls []ToolCall           `json:"tool_calls,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCall represents a tool/function call request
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool represents a tool/function definition
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a tool function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// GenerateRequest represents a generation request
type GenerateRequest struct {
	Messages    []Message              `json:"messages"`
	Tools       []Tool                 `json:"tools,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// GenerateResponse represents a generation response
type GenerateResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   Usage     `json:"usage"`
	Error   *LLMError `json:"error,omitempty"`
}

// Choice represents a generation choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMError represents an LLM error
type LLMError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// LLMClient defines the interface for LLM clients
type LLMClient interface {
	// Generate generates a response for the given messages
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	
	// GenerateStream generates a streaming response
	GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *GenerateResponse, error)
	
	// GetProvider returns the provider name
	GetProvider() LLMProvider
	
	// GetModel returns the model name
	GetModel() string
	
	// HealthCheck checks if the client is healthy
	HealthCheck(ctx context.Context) error
	
	// Close closes the client
	Close() error
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Provider    LLMProvider            `json:"provider"`
	Model       string                 `json:"model"`
	APIKey      string                 `json:"api_key"`
	BaseURL     string                 `json:"base_url,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	MaxRetries  int                    `json:"max_retries"`
	Temperature float64                `json:"temperature"`
	MaxTokens   int                    `json:"max_tokens"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LLMAdapterConfig represents the adapter configuration
type LLMAdapterConfig struct {
	Primary   LLMConfig   `json:"primary"`
	Fallback  []LLMConfig `json:"fallback,omitempty"`
	Budget    *LLMConfig  `json:"budget,omitempty"`
	LoadBalancing bool     `json:"load_balancing"`
	CostOptimization bool `json:"cost_optimization"`
}

// LLMAdapter defines the unified interface for all LLM providers
type LLMAdapter interface {
	// Generate generates a response with automatic fallback
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	
	// GenerateWithFallback generates with explicit fallback strategy
	GenerateWithFallback(ctx context.Context, req *GenerateRequest, strategy FallbackStrategy) (*GenerateResponse, error)
	
	// GetAvailableProviders returns list of configured providers
	GetAvailableProviders() []LLMProvider
	
	// GetProviderStatus returns the health status of all providers
	GetProviderStatus(ctx context.Context) map[LLMProvider]ProviderStatus
	
	// UpdateConfig updates the adapter configuration
	UpdateConfig(config *LLMAdapterConfig) error
	
	// GetMetrics returns usage metrics
	GetMetrics() *LLMMetrics
}

// FallbackStrategy defines fallback strategies
type FallbackStrategy string

const (
	FallbackNone       FallbackStrategy = "none"
	FallbackAutomatic  FallbackStrategy = "automatic"
	FallbackCostBased  FallbackStrategy = "cost_based"
	FallbackSpeedBased FallbackStrategy = "speed_based"
)

// ProviderStatus represents the status of a provider
type ProviderStatus struct {
	Available    bool          `json:"available"`
	Latency      time.Duration `json:"latency"`
	ErrorRate    float64       `json:"error_rate"`
	LastError    string        `json:"last_error,omitempty"`
	LastErrorAt  *time.Time    `json:"last_error_at,omitempty"`
	RequestCount int64         `json:"request_count"`
}

// LLMMetrics represents LLM usage metrics
type LLMMetrics struct {
	TotalRequests    int64                         `json:"total_requests"`
	TotalTokens      int64                         `json:"total_tokens"`
	AverageLatency   time.Duration                 `json:"average_latency"`
	SuccessRate      float64                       `json:"success_rate"`
	CostEstimate     float64                       `json:"cost_estimate"`
	ProviderMetrics  map[LLMProvider]*ProviderStatus `json:"provider_metrics"`
	LastUpdated      time.Time                     `json:"last_updated"`
}