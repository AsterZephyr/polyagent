package ai

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Message represents a simple message structure
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SimpleModelRouter provides basic model routing without over-engineering
type SimpleModelRouter struct {
	defaultModel string
	logger       *logrus.Logger
}

// SimpleRouteRequest simplified routing request
type SimpleRouteRequest struct {
	Messages        []Message `json:"messages"`
	ModelPreference string    `json:"model_preference,omitempty"`
	UserID          string    `json:"user_id"`
	SessionID       string    `json:"session_id"`
}

// SimpleRouteResponse simplified routing response
type SimpleRouteResponse struct {
	Response    *Message      `json:"response"`
	Model       string        `json:"model"`
	TokensUsed  int           `json:"tokens_used"`
	Cost        float64       `json:"cost"`
	Latency     time.Duration `json:"latency"`
}

func NewSimpleModelRouter(defaultModel string, logger *logrus.Logger) *SimpleModelRouter {
	return &SimpleModelRouter{
		defaultModel: defaultModel,
		logger:       logger,
	}
}

func (r *SimpleModelRouter) Route(ctx context.Context, req *SimpleRouteRequest) (*SimpleRouteResponse, error) {
	start := time.Now()
	
	model := r.defaultModel
	if req.ModelPreference != "" {
		model = req.ModelPreference
	}

	// Simple mock response - in production this would call actual AI models
	response := &Message{
		Role:    "assistant",
		Content: "This is a simplified response from " + model,
	}

	return &SimpleRouteResponse{
		Response:   response,
		Model:      model,
		TokensUsed: len(req.Messages) * 10, // Simple estimation
		Cost:       0.001,
		Latency:    time.Since(start),
	}, nil
}

// Compatibility with old interface
type RouteRequest = SimpleRouteRequest
type RouteResponse = SimpleRouteResponse
type ModelRouter = SimpleModelRouter

func NewModelRouter(cfg *ModelConfig, logger *logrus.Logger) *ModelRouter {
	return NewSimpleModelRouter("gpt-4", logger)
}

// Simplified config
type ModelConfig struct {
	DefaultModel string `json:"default_model"`
}

// Health check for model router
func (r *SimpleModelRouter) GetHealth(modelID string) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status": "healthy",
		"model":  r.defaultModel,
	}
	
	if modelID != "" {
		health["requested_model"] = modelID
	}
	
	return health, nil
}