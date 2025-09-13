package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// BaseClient provides common functionality for all LLM clients
type BaseClient struct {
	config   *LLMConfig
	logger   *logrus.Logger
	provider LLMProvider
	model    string
}

// NewBaseClient creates a new base client
func NewBaseClient(config *LLMConfig, logger *logrus.Logger) *BaseClient {
	return &BaseClient{
		config:   config,
		logger:   logger,
		provider: config.Provider,
		model:    config.Model,
	}
}

// GetProvider returns the provider name
func (c *BaseClient) GetProvider() LLMProvider {
	return c.provider
}

// GetModel returns the model name
func (c *BaseClient) GetModel() string {
	return c.model
}

// OpenAIClient implements LLMClient for OpenAI
type OpenAIClient struct {
	*BaseClient
	httpClient *http.Client
	baseURL    string
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(config *LLMConfig, logger *logrus.Logger) (*OpenAIClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	client := &OpenAIClient{
		BaseClient: NewBaseClient(config, logger),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: baseURL,
	}

	logger.Infof("Created OpenAI client for model: %s", config.Model)
	return client, nil
}

// Generate generates a response using OpenAI API
func (c *OpenAIClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	c.logger.Debugf("Generating response with OpenAI model: %s", c.model)
	
	// Prepare OpenAI API request
	openaiReq := map[string]interface{}{
		"model":       c.model,
		"messages":    c.convertMessages(req.Messages),
		"temperature": req.Temperature,
		"max_tokens":  req.MaxTokens,
	}

	if len(req.Tools) > 0 {
		openaiReq["tools"] = c.convertTools(req.Tools)
		openaiReq["tool_choice"] = "auto"
	}

	// Marshal request
	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// Make API call
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var openaiResp map[string]interface{}
	if err := json.Unmarshal(body, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to unified format
	return c.convertResponse(openaiResp), nil
}

// GenerateStream generates a streaming response
func (c *OpenAIClient) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *GenerateResponse, error) {
	ch := make(chan *GenerateResponse, 1)
	go func() {
		defer close(ch)
		// TODO: Implement streaming
		response, _ := c.Generate(ctx, req)
		ch <- response
	}()
	return ch, nil
}

// HealthCheck checks if the OpenAI client is healthy
func (c *OpenAIClient) HealthCheck(ctx context.Context) error {
	// TODO: Implement actual health check
	c.logger.Debug("OpenAI health check passed")
	return nil
}

// Close closes the OpenAI client
func (c *OpenAIClient) Close() error {
	c.logger.Info("Closing OpenAI client")
	return nil
}

// Helper methods for OpenAI client

func (c *OpenAIClient) convertMessages(messages []Message) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		converted[i] = map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.Name != "" {
			converted[i]["name"] = msg.Name
		}
	}
	return converted
}

func (c *OpenAIClient) convertTools(tools []Tool) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		converted[i] = map[string]interface{}{
			"type": tool.Type,
			"function": map[string]interface{}{
				"name":        tool.Function.Name,
				"description": tool.Function.Description,
				"parameters":  tool.Function.Parameters,
			},
		}
	}
	return converted
}

func (c *OpenAIClient) convertResponse(openaiResp map[string]interface{}) *GenerateResponse {
	response := &GenerateResponse{
		ID:      getStringFromMap(openaiResp, "id"),
		Object:  getStringFromMap(openaiResp, "object"),
		Model:   getStringFromMap(openaiResp, "model"),
		Created: getInt64FromMap(openaiResp, "created"),
	}

	// Parse choices
	if choicesData, ok := openaiResp["choices"].([]interface{}); ok && len(choicesData) > 0 {
		choices := make([]Choice, len(choicesData))
		for i, choiceData := range choicesData {
			if choiceMap, ok := choiceData.(map[string]interface{}); ok {
				choice := Choice{
					Index:        getIntFromMap(choiceMap, "index"),
					FinishReason: getStringFromMap(choiceMap, "finish_reason"),
				}

				// Parse message
				if msgData, ok := choiceMap["message"].(map[string]interface{}); ok {
					choice.Message = Message{
						Role:    getStringFromMap(msgData, "role"),
						Content: getStringFromMap(msgData, "content"),
					}

					// Parse tool calls if present
					if toolCallsData, ok := msgData["tool_calls"].([]interface{}); ok {
						toolCalls := make([]ToolCall, len(toolCallsData))
						for j, tcData := range toolCallsData {
							if tcMap, ok := tcData.(map[string]interface{}); ok {
								toolCall := ToolCall{
									ID:   getStringFromMap(tcMap, "id"),
									Type: getStringFromMap(tcMap, "type"),
								}

								if funcData, ok := tcMap["function"].(map[string]interface{}); ok {
									toolCall.Function = FunctionCall{
										Name:      getStringFromMap(funcData, "name"),
										Arguments: getStringFromMap(funcData, "arguments"),
									}
								}

								toolCalls[j] = toolCall
							}
						}
						choice.Message.ToolCalls = toolCalls
					}
				}

				choices[i] = choice
			}
		}
		response.Choices = choices
	}

	// Parse usage
	if usageData, ok := openaiResp["usage"].(map[string]interface{}); ok {
		response.Usage = Usage{
			PromptTokens:     getIntFromMap(usageData, "prompt_tokens"),
			CompletionTokens: getIntFromMap(usageData, "completion_tokens"),
			TotalTokens:      getIntFromMap(usageData, "total_tokens"),
		}
	}

	return response
}

// Helper functions for map parsing
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	return 0
}

// ClaudeClient implements LLMClient for Claude/Anthropic
type ClaudeClient struct {
	*BaseClient
	httpClient *http.Client
	baseURL    string
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(config *LLMConfig, logger *logrus.Logger) (*ClaudeClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Claude API key is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	client := &ClaudeClient{
		BaseClient: NewBaseClient(config, logger),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL: baseURL,
	}

	logger.Infof("Created Claude client for model: %s", config.Model)
	return client, nil
}

// Generate generates a response using Claude API
func (c *ClaudeClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	c.logger.Debugf("Generating response with Claude model: %s", c.model)
	
	// Prepare Claude API request
	claudeReq := map[string]interface{}{
		"model":     c.model,
		"messages":  c.convertMessagesForClaude(req.Messages),
		"max_tokens": req.MaxTokens,
	}

	if req.Temperature > 0 {
		claudeReq["temperature"] = req.Temperature
	}

	if len(req.Tools) > 0 {
		claudeReq["tools"] = c.convertToolsForClaude(req.Tools)
	}

	// Marshal request
	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	// Make API call
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var claudeResp map[string]interface{}
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert to unified format
	return c.convertClaudeResponse(claudeResp), nil
}

// GenerateStream generates a streaming response
func (c *ClaudeClient) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *GenerateResponse, error) {
	ch := make(chan *GenerateResponse, 1)
	go func() {
		defer close(ch)
		response, _ := c.Generate(ctx, req)
		ch <- response
	}()
	return ch, nil
}

// HealthCheck checks if the Claude client is healthy
func (c *ClaudeClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Claude health check passed")
	return nil
}

// Close closes the Claude client
func (c *ClaudeClient) Close() error {
	c.logger.Info("Closing Claude client")
	return nil
}

// Helper methods for Claude client

func (c *ClaudeClient) convertMessagesForClaude(messages []Message) []map[string]interface{} {
	converted := make([]map[string]interface{}, 0, len(messages))
	
	// Claude expects messages to start with user role
	for _, msg := range messages {
		if msg.Role == "system" {
			// System messages are handled differently in Claude
			continue
		}
		
		converted = append(converted, map[string]interface{}{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	
	return converted
}

func (c *ClaudeClient) convertToolsForClaude(tools []Tool) []map[string]interface{} {
	converted := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		converted[i] = map[string]interface{}{
			"name":         tool.Function.Name,
			"description":  tool.Function.Description,
			"input_schema": tool.Function.Parameters,
		}
	}
	return converted
}

func (c *ClaudeClient) convertClaudeResponse(claudeResp map[string]interface{}) *GenerateResponse {
	response := &GenerateResponse{
		ID:      getStringFromMap(claudeResp, "id"),
		Object:  "chat.completion",
		Model:   getStringFromMap(claudeResp, "model"),
		Created: time.Now().Unix(),
	}

	// Claude has different response format
	if contentArray, ok := claudeResp["content"].([]interface{}); ok && len(contentArray) > 0 {
		choice := Choice{
			Index:        0,
			FinishReason: getStringFromMap(claudeResp, "stop_reason"),
		}

		// Extract text content
		if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
			if contentType := getStringFromMap(contentItem, "type"); contentType == "text" {
				choice.Message = Message{
					Role:    "assistant",
					Content: getStringFromMap(contentItem, "text"),
				}
			}
		}

		response.Choices = []Choice{choice}
	}

	// Claude usage format
	if usageData, ok := claudeResp["usage"].(map[string]interface{}); ok {
		response.Usage = Usage{
			PromptTokens:     getIntFromMap(usageData, "input_tokens"),
			CompletionTokens: getIntFromMap(usageData, "output_tokens"),
			TotalTokens:      getIntFromMap(usageData, "input_tokens") + getIntFromMap(usageData, "output_tokens"),
		}
	}

	return response
}

// QwenClient implements LLMClient for Qwen
type QwenClient struct {
	*BaseClient
}

// NewQwenClient creates a new Qwen client
func NewQwenClient(config *LLMConfig, logger *logrus.Logger) (*QwenClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Qwen API key is required")
	}

	client := &QwenClient{
		BaseClient: NewBaseClient(config, logger),
	}

	logger.Infof("Created Qwen client for model: %s", config.Model)
	return client, nil
}

// Generate generates a response using Qwen API
func (c *QwenClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	c.logger.Debugf("Generating response with Qwen model: %s", c.model)
	
	// TODO: Implement actual Qwen API call
	return &GenerateResponse{
		ID:      "mock-qwen-response",
		Object:  "chat.completion",
		Model:   c.model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Mock response from Qwen",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     90,
			CompletionTokens: 60,
			TotalTokens:      150,
		},
	}, nil
}

// GenerateStream generates a streaming response
func (c *QwenClient) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *GenerateResponse, error) {
	ch := make(chan *GenerateResponse, 1)
	go func() {
		defer close(ch)
		response, _ := c.Generate(ctx, req)
		ch <- response
	}()
	return ch, nil
}

// HealthCheck checks if the Qwen client is healthy
func (c *QwenClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("Qwen health check passed")
	return nil
}

// Close closes the Qwen client
func (c *QwenClient) Close() error {
	c.logger.Info("Closing Qwen client")
	return nil
}

// OpenRouterClient implements LLMClient for OpenRouter/K2
type OpenRouterClient struct {
	*BaseClient
}

// NewOpenRouterClient creates a new OpenRouter client
func NewOpenRouterClient(config *LLMConfig, logger *logrus.Logger) (*OpenRouterClient, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenRouter API key is required")
	}

	client := &OpenRouterClient{
		BaseClient: NewBaseClient(config, logger),
	}

	logger.Infof("Created OpenRouter client for model: %s", config.Model)
	return client, nil
}

// Generate generates a response using OpenRouter API
func (c *OpenRouterClient) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	c.logger.Debugf("Generating response with OpenRouter model: %s", c.model)
	
	// TODO: Implement actual OpenRouter API call
	return &GenerateResponse{
		ID:      "mock-openrouter-response",
		Object:  "chat.completion",
		Model:   c.model,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    "assistant",
					Content: "Mock response from OpenRouter",
				},
				FinishReason: "stop",
			},
		},
		Usage: Usage{
			PromptTokens:     85,
			CompletionTokens: 65,
			TotalTokens:      150,
		},
	}, nil
}

// GenerateStream generates a streaming response
func (c *OpenRouterClient) GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *GenerateResponse, error) {
	ch := make(chan *GenerateResponse, 1)
	go func() {
		defer close(ch)
		response, _ := c.Generate(ctx, req)
		ch <- response
	}()
	return ch, nil
}

// HealthCheck checks if the OpenRouter client is healthy
func (c *OpenRouterClient) HealthCheck(ctx context.Context) error {
	c.logger.Debug("OpenRouter health check passed")
	return nil
}

// Close closes the OpenRouter client
func (c *OpenRouterClient) Close() error {
	c.logger.Info("Closing OpenRouter client")
	return nil
}