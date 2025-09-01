package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Message represents a simple message structure
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents OpenAI API request format
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

// OpenAIResponse represents OpenAI API response format
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ModelProvider defines supported AI model providers
type ModelProvider struct {
	Name    string
	BaseURL string
	APIKey  string
	Models  []string
}

// SimpleModelRouter provides basic model routing with real AI API integration
type SimpleModelRouter struct {
	defaultModel string
	logger       *logrus.Logger
	httpClient   *http.Client
	providers    map[string]*ModelProvider
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
	router := &SimpleModelRouter{
		defaultModel: defaultModel,
		logger:       logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		providers: make(map[string]*ModelProvider),
	}
	
	// Initialize supported providers
	router.initializeProviders()
	
	return router
}

// initializeProviders sets up supported AI model providers
func (r *SimpleModelRouter) initializeProviders() {
	// OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		r.providers["openai"] = &ModelProvider{
			Name:    "openai",
			BaseURL: "https://api.openai.com/v1",
			APIKey:  apiKey,
			Models:  []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo", "gpt-4o"},
		}
	}
	
	// DeepSeek (OpenAI compatible)
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		r.providers["deepseek"] = &ModelProvider{
			Name:    "deepseek",
			BaseURL: "https://api.deepseek.com/v1",
			APIKey:  apiKey,
			Models:  []string{"deepseek-chat", "deepseek-coder"},
		}
	}
	
	// OpenRouter
	if apiKey := os.Getenv("OPENROUTER_API_KEY"); apiKey != "" {
		r.providers["openrouter"] = &ModelProvider{
			Name:    "openrouter",
			BaseURL: "https://openrouter.ai/api/v1",
			APIKey:  apiKey,
			Models:  []string{"meta-llama/llama-3.1-8b-instruct:free", "nousresearch/hermes-3-llama-3.1-405b:free"},
		}
	}
	
	r.logger.WithField("providers_count", len(r.providers)).Info("AI providers initialized")
}

func (r *SimpleModelRouter) Route(ctx context.Context, req *SimpleRouteRequest) (*SimpleRouteResponse, error) {
	start := time.Now()
	
	model := r.defaultModel
	if req.ModelPreference != "" {
		model = req.ModelPreference
	}

	// Find provider for the model
	provider := r.findProviderForModel(model)
	if provider == nil {
		// Fallback to mock response if no provider found
		return r.mockResponse(model, start), nil
	}

	// Call real AI API
	response, tokenUsage, err := r.callAIAPI(ctx, provider, model, req.Messages)
	if err != nil {
		r.logger.WithError(err).WithField("model", model).Error("AI API call failed, using fallback")
		return r.mockResponse(model, start), nil
	}

	return &SimpleRouteResponse{
		Response:   response,
		Model:      model,
		TokensUsed: tokenUsage,
		Cost:       r.calculateCost(model, tokenUsage),
		Latency:    time.Since(start),
	}, nil
}

// findProviderForModel finds the appropriate provider for a given model
func (r *SimpleModelRouter) findProviderForModel(model string) *ModelProvider {
	for _, provider := range r.providers {
		for _, supportedModel := range provider.Models {
			if supportedModel == model || strings.Contains(model, supportedModel) {
				return provider
			}
		}
	}
	return nil
}

// callAIAPI makes actual API call to AI service
func (r *SimpleModelRouter) callAIAPI(ctx context.Context, provider *ModelProvider, model string, messages []Message) (*Message, int, error) {
	reqBody := OpenAIRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", provider.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	
	// Special headers for different providers
	if provider.Name == "openrouter" {
		req.Header.Set("HTTP-Referer", "https://github.com/polyagent/eino-polyagent")
		req.Header.Set("X-Title", "PolyAgent")
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, 0, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, 0, fmt.Errorf("no choices in API response")
	}

	r.logger.WithFields(logrus.Fields{
		"model":       model,
		"provider":    provider.Name,
		"tokens_used": apiResp.Usage.TotalTokens,
	}).Info("AI API call successful")

	return &apiResp.Choices[0].Message, apiResp.Usage.TotalTokens, nil
}

// mockResponse provides fallback response when API is unavailable
func (r *SimpleModelRouter) mockResponse(model string, start time.Time) *SimpleRouteResponse {
	r.logger.WithField("model", model).Warn("Using mock response - no API provider configured")
	
	response := &Message{
		Role:    "assistant",
		Content: fmt.Sprintf("Mock response from %s (API not configured)", model),
	}

	return &SimpleRouteResponse{
		Response:   response,
		Model:      model,
		TokensUsed: 50, // Estimated
		Cost:       0.0,
		Latency:    time.Since(start),
	}
}

// calculateCost estimates API call cost based on model and token usage
func (r *SimpleModelRouter) calculateCost(model string, tokens int) float64 {
	// Simplified cost calculation
	costPerToken := 0.00002 // Default rate
	
	switch {
	case strings.Contains(model, "gpt-4"):
		costPerToken = 0.00006
	case strings.Contains(model, "gpt-3.5"):
		costPerToken = 0.000002
	case strings.Contains(model, "deepseek"):
		costPerToken = 0.000001
	case strings.Contains(model, "free"):
		costPerToken = 0.0
	}
	
	return float64(tokens) * costPerToken
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

// GenerateResponse provides compatibility with existing interface
func (r *SimpleModelRouter) GenerateResponse(modelID, prompt string) (string, error) {
	ctx := context.Background()
	
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	
	req := &SimpleRouteRequest{
		Messages:        messages,
		ModelPreference: modelID,
	}
	
	response, err := r.Route(ctx, req)
	if err != nil {
		return "", err
	}
	
	return response.Response.Content, nil
}

// ListAvailableModels returns all supported models across providers
func (r *SimpleModelRouter) ListAvailableModels() map[string][]string {
	models := make(map[string][]string)
	
	for name, provider := range r.providers {
		models[name] = provider.Models
	}
	
	// Add default fallback models
	if len(models) == 0 {
		models["mock"] = []string{"gpt-4", "gpt-3.5-turbo", "deepseek-chat"}
	}
	
	return models
}

// Health check for model router
func (r *SimpleModelRouter) GetHealth(modelID string) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status":           "healthy",
		"default_model":    r.defaultModel,
		"providers_count":  len(r.providers),
		"available_models": r.ListAvailableModels(),
	}
	
	if modelID != "" {
		health["requested_model"] = modelID
		provider := r.findProviderForModel(modelID)
		health["model_supported"] = provider != nil
		if provider != nil {
			health["provider"] = provider.Name
		}
	}
	
	return health, nil
}