package gateway

import (
	"context"
	"time"
)

// GatewayService defines the main gateway service interface
type GatewayService interface {
	// HTTP handling
	HandleChatRequest(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	HandleHealthCheck(ctx context.Context) (*HealthResponse, error)
	
	// Authentication & Authorization
	Authenticate(ctx context.Context, token string) (*UserContext, error)
	Authorize(ctx context.Context, user *UserContext, resource string) error
	
	// Rate limiting
	CheckRateLimit(ctx context.Context, userID string, endpoint string) error
	
	// Circuit breaker
	IsServiceAvailable(serviceName string) bool
	RecordServiceCall(serviceName string, success bool, duration time.Duration)
}

// Request/Response types
type ChatRequest struct {
	Message     string            `json:"message"`
	SessionID   string            `json:"session_id,omitempty"`
	AgentID     string            `json:"agent_id,omitempty"`
	Context     map[string]string `json:"context,omitempty"`
	UseTools    bool              `json:"use_tools"`
	StreamMode  bool              `json:"stream_mode"`
}

type ChatResponse struct {
	Response     string            `json:"response"`
	SessionID    string            `json:"session_id"`
	ModelUsed    string            `json:"model_used"`
	Usage        *TokenUsage       `json:"usage"`
	Cost         float64           `json:"cost"`
	ToolsCalled  []string          `json:"tools_called"`
	ProcessTime  time.Duration     `json:"process_time_ms"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type UserContext struct {
	UserID      string   `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	RateLimit   int      `json:"rate_limit"`
	ExpiresAt   int64    `json:"expires_at"`
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    time.Duration     `json:"uptime"`
	Services  map[string]string `json:"services"`
	Timestamp int64             `json:"timestamp"`
}

// GatewayConfig holds gateway configuration
type GatewayConfig struct {
	Port                int           `yaml:"port"`
	ReadTimeout         time.Duration `yaml:"read_timeout"`
	WriteTimeout        time.Duration `yaml:"write_timeout"`
	MaxRequestSize      int64         `yaml:"max_request_size"`
	EnableCORS          bool          `yaml:"enable_cors"`
	EnableRateLimit     bool          `yaml:"enable_rate_limit"`
	EnableCircuitBreaker bool          `yaml:"enable_circuit_breaker"`
	
	// Service endpoints
	AgentServiceURL    string `yaml:"agent_service_url"`
	WorkflowEngineURL  string `yaml:"workflow_engine_url"`
	ModelRouterURL     string `yaml:"model_router_url"`
	
	// Security
	JWTSecret          string        `yaml:"jwt_secret"`
	TokenExpiry        time.Duration `yaml:"token_expiry"`
	
	// Rate limiting
	DefaultRateLimit   int           `yaml:"default_rate_limit"`
	RateLimitWindow    time.Duration `yaml:"rate_limit_window"`
}

// Implementation
type gatewayService struct {
	config      *GatewayConfig
	rateLimiter RateLimiter
	circuitBreaker CircuitBreaker
	authenticator Authenticator
	
	// Service clients
	agentClient    AgentServiceClient
	workflowClient WorkflowEngineClient
	routerClient   ModelRouterClient
}

// NewGatewayService creates a new gateway service instance
func NewGatewayService(config *GatewayConfig) GatewayService {
	return &gatewayService{
		config: config,
		rateLimiter: NewRateLimiter(config.DefaultRateLimit, config.RateLimitWindow),
		circuitBreaker: NewCircuitBreaker(),
		authenticator: NewJWTAuthenticator(config.JWTSecret),
		
		// Initialize service clients
		agentClient: NewAgentServiceClient(config.AgentServiceURL),
		workflowClient: NewWorkflowEngineClient(config.WorkflowEngineURL), 
		routerClient: NewModelRouterClient(config.ModelRouterURL),
	}
}

// Service client interfaces for dependency injection and testing
type AgentServiceClient interface {
	ProcessMessage(ctx context.Context, req *ProcessMessageRequest) (*ProcessMessageResponse, error)
}

type WorkflowEngineClient interface {
	ExecuteWorkflow(ctx context.Context, req *ExecuteWorkflowRequest) (*ExecuteWorkflowResponse, error)
}

type ModelRouterClient interface {
	RouteRequest(ctx context.Context, req *RouteRequest) (*RouteResponse, error)
}

// Supporting interfaces
type RateLimiter interface {
	Allow(userID string, endpoint string) bool
	Reset(userID string) error
}

type CircuitBreaker interface {
	Call(serviceName string, fn func() error) error
	IsOpen(serviceName string) bool
}

type Authenticator interface {
	ValidateToken(token string) (*UserContext, error)
	GenerateToken(userID string, roles []string) (string, error)
}