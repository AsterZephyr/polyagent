package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ToolIntegrationLayer handles tool execution and LLM integration
type ToolIntegrationLayer struct {
	registry *ToolRegistry
	logger   *logrus.Logger
	metrics  *ToolMetrics
}

// ToolMetrics tracks tool usage and performance
type ToolMetrics struct {
	TotalExecutions   int64                    `json:"total_executions"`
	SuccessfulCalls   int64                    `json:"successful_calls"`
	FailedCalls       int64                    `json:"failed_calls"`
	AverageLatency    time.Duration            `json:"average_latency"`
	ToolUsageStats    map[string]*ToolStats    `json:"tool_usage_stats"`
	LastUpdated       time.Time                `json:"last_updated"`
}

// ToolStats tracks individual tool performance
type ToolStats struct {
	CallCount       int64         `json:"call_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	TotalLatency    time.Duration `json:"total_latency"`
	AverageLatency  time.Duration `json:"average_latency"`
	LastUsed        time.Time     `json:"last_used"`
	LastError       string        `json:"last_error"`
}

// ToolExecutionResult represents the result of tool execution
type ToolExecutionResult struct {
	ToolName     string                 `json:"tool_name"`
	Success      bool                   `json:"success"`
	Result       interface{}            `json:"result"`
	Error        string                 `json:"error,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	Timestamp    time.Time              `json:"timestamp"`
	Parameters   map[string]interface{} `json:"parameters"`
}

// NewToolIntegrationLayer creates a new tool integration layer
func NewToolIntegrationLayer(logger *logrus.Logger) *ToolIntegrationLayer {
	return &ToolIntegrationLayer{
		registry: NewToolRegistry(),
		logger:   logger,
		metrics: &ToolMetrics{
			ToolUsageStats: make(map[string]*ToolStats),
			LastUpdated:    time.Now(),
		},
	}
}

// ProcessToolCalls processes tool calls from LLM responses
func (til *ToolIntegrationLayer) ProcessToolCalls(ctx context.Context, toolCalls []ToolCall) ([]*ToolExecutionResult, error) {
	if len(toolCalls) == 0 {
		return nil, nil
	}

	til.logger.WithField("tool_count", len(toolCalls)).Info("Processing tool calls")

	results := make([]*ToolExecutionResult, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		result := til.executeToolCall(ctx, toolCall)
		results = append(results, result)
		
		til.updateMetrics(result)
		
		til.logger.WithFields(logrus.Fields{
			"tool_name":      result.ToolName,
			"success":        result.Success,
			"execution_time": result.ExecutionTime,
		}).Info("Tool call executed")
	}

	return results, nil
}

// executeToolCall executes a single tool call
func (til *ToolIntegrationLayer) executeToolCall(ctx context.Context, toolCall ToolCall) *ToolExecutionResult {
	startTime := time.Now()
	
	result := &ToolExecutionResult{
		ToolName:  toolCall.Function.Name,
		Timestamp: startTime,
	}

	// Parse parameters
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to parse tool arguments: %v", err)
		result.ExecutionTime = time.Since(startTime)
		return result
	}
	
	result.Parameters = params

	// Execute tool
	toolResult, err := til.registry.ExecuteTool(ctx, toolCall.Function.Name, params)
	result.ExecutionTime = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		til.logger.WithError(err).WithField("tool_name", toolCall.Function.Name).Error("Tool execution failed")
	} else {
		result.Success = true
		result.Result = toolResult
		til.logger.WithField("tool_name", toolCall.Function.Name).Info("Tool executed successfully")
	}

	return result
}

// GetAvailableTools returns all available tools for LLM context
func (til *ToolIntegrationLayer) GetAvailableTools() []Tool {
	return til.registry.GetAllTools()
}

// GetToolMetrics returns tool usage metrics
func (til *ToolIntegrationLayer) GetToolMetrics() *ToolMetrics {
	return til.metrics
}

// updateMetrics updates tool usage metrics
func (til *ToolIntegrationLayer) updateMetrics(result *ToolExecutionResult) {
	til.metrics.TotalExecutions++
	
	if result.Success {
		til.metrics.SuccessfulCalls++
	} else {
		til.metrics.FailedCalls++
	}

	// Update average latency
	if til.metrics.TotalExecutions == 1 {
		til.metrics.AverageLatency = result.ExecutionTime
	} else {
		til.metrics.AverageLatency = (til.metrics.AverageLatency + result.ExecutionTime) / 2
	}

	// Update tool-specific stats
	toolName := result.ToolName
	if _, exists := til.metrics.ToolUsageStats[toolName]; !exists {
		til.metrics.ToolUsageStats[toolName] = &ToolStats{}
	}

	stats := til.metrics.ToolUsageStats[toolName]
	stats.CallCount++
	stats.TotalLatency += result.ExecutionTime
	stats.AverageLatency = stats.TotalLatency / time.Duration(stats.CallCount)
	stats.LastUsed = result.Timestamp

	if result.Success {
		stats.SuccessCount++
	} else {
		stats.ErrorCount++
		stats.LastError = result.Error
	}

	til.metrics.LastUpdated = time.Now()
}

// Enhanced LLM Adapter with Tool Integration
type EnhancedLLMAdapter struct {
	*UnifiedLLMAdapter
	toolLayer *ToolIntegrationLayer
}

// NewEnhancedLLMAdapter creates an LLM adapter with tool integration
func NewEnhancedLLMAdapter(config *LLMAdapterConfig, logger *logrus.Logger) (*EnhancedLLMAdapter, error) {
	baseAdapter, err := NewUnifiedLLMAdapter(config, logger)
	if err != nil {
		return nil, err
	}

	return &EnhancedLLMAdapter{
		UnifiedLLMAdapter: baseAdapter,
		toolLayer:         NewToolIntegrationLayer(logger),
	}, nil
}

// GenerateWithTools generates a response with automatic tool execution
func (ea *EnhancedLLMAdapter) GenerateWithTools(ctx context.Context, req *GenerateRequest) (*GenerateResponse, []*ToolExecutionResult, error) {
	// Add available tools to request if not provided
	if len(req.Tools) == 0 {
		req.Tools = ea.toolLayer.GetAvailableTools()
	}

	// Generate initial response
	response, err := ea.Generate(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	// Process tool calls if any
	var toolResults []*ToolExecutionResult
	if len(response.Choices) > 0 && len(response.Choices[0].Message.ToolCalls) > 0 {
		toolResults, err = ea.toolLayer.ProcessToolCalls(ctx, response.Choices[0].Message.ToolCalls)
		if err != nil {
			ea.logger.WithError(err).Error("Failed to process tool calls")
			return response, nil, err
		}

		// If tools were executed successfully, generate a follow-up response with tool results
		if len(toolResults) > 0 {
			followUpResponse, err := ea.generateFollowUpWithToolResults(ctx, req, response, toolResults)
			if err != nil {
				ea.logger.WithError(err).Warn("Failed to generate follow-up response with tool results")
				// Return original response even if follow-up fails
				return response, toolResults, nil
			}
			return followUpResponse, toolResults, nil
		}
	}

	return response, toolResults, nil
}

// generateFollowUpWithToolResults generates a follow-up response incorporating tool results
func (ea *EnhancedLLMAdapter) generateFollowUpWithToolResults(ctx context.Context, originalReq *GenerateRequest, llmResponse *GenerateResponse, toolResults []*ToolExecutionResult) (*GenerateResponse, error) {
	// Create new messages including tool results
	newMessages := make([]Message, len(originalReq.Messages))
	copy(newMessages, originalReq.Messages)

	// Add the LLM's response with tool calls
	newMessages = append(newMessages, llmResponse.Choices[0].Message)

	// Add tool results as system messages
	for _, result := range toolResults {
		var content string
		if result.Success {
			resultJSON, _ := json.MarshalIndent(result.Result, "", "  ")
			content = fmt.Sprintf("Tool '%s' executed successfully:\n%s", result.ToolName, string(resultJSON))
		} else {
			content = fmt.Sprintf("Tool '%s' failed with error: %s", result.ToolName, result.Error)
		}

		newMessages = append(newMessages, Message{
			Role:    "system",
			Content: content,
		})
	}

	// Add instruction for the LLM to incorporate tool results
	newMessages = append(newMessages, Message{
		Role:    "user",
		Content: "Please provide a comprehensive response incorporating the results from the executed tools.",
	})

	// Generate follow-up response
	followUpReq := &GenerateRequest{
		Messages:    newMessages,
		Model:       originalReq.Model,
		Temperature: originalReq.Temperature,
		MaxTokens:   originalReq.MaxTokens,
		// Don't include tools in follow-up to avoid recursive tool calling
	}

	return ea.Generate(ctx, followUpReq)
}

// GetToolLayer returns the tool integration layer
func (ea *EnhancedLLMAdapter) GetToolLayer() *ToolIntegrationLayer {
	return ea.toolLayer
}

// RecommendationQueryProcessor handles specific recommendation queries
type RecommendationQueryProcessor struct {
	adapter   *EnhancedLLMAdapter
	logger    *logrus.Logger
}

// NewRecommendationQueryProcessor creates a new recommendation query processor
func NewRecommendationQueryProcessor(adapter *EnhancedLLMAdapter, logger *logrus.Logger) *RecommendationQueryProcessor {
	return &RecommendationQueryProcessor{
		adapter: adapter,
		logger:  logger,
	}
}

// ProcessRecommendationQuery processes a user's recommendation query
func (rqp *RecommendationQueryProcessor) ProcessRecommendationQuery(ctx context.Context, userQuery string, userID string) (*RecommendationQueryResult, error) {
	// Create system prompt for recommendation context
	systemPrompt := `You are an expert movie recommendation assistant. Your goal is to provide personalized, helpful movie recommendations based on user preferences and queries.

Available tools:
- search_movies: Find movies by genre, year, rating, and keywords
- analyze_user_preferences: Understand user's viewing history and preferences
- filter_content: Apply content filters for age ratings and content warnings
- generate_recommendations: Create personalized recommendations using ML algorithms
- track_user_interaction: Record user interactions with recommendations
- analyze_popularity: Get trending movies and popularity insights

When a user asks for recommendations:
1. First analyze their preferences if user_id is provided
2. Use appropriate search and filtering tools based on their requirements
3. Generate personalized recommendations
4. Explain your reasoning clearly
5. Track the interaction for future improvements

Be conversational, helpful, and always explain why you're recommending specific movies.`

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("User ID: %s\nQuery: %s", userID, userQuery),
		},
	}

	req := &GenerateRequest{
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
	}

	startTime := time.Now()
	response, toolResults, err := rqp.adapter.GenerateWithTools(ctx, req)
	duration := time.Since(startTime)

	if err != nil {
		rqp.logger.WithError(err).Error("Failed to process recommendation query")
		return nil, err
	}

	result := &RecommendationQueryResult{
		UserID:           userID,
		Query:            userQuery,
		Response:         response.Choices[0].Message.Content,
		ToolResults:      toolResults,
		ProcessingTime:   duration,
		Timestamp:        time.Now(),
		Model:            response.Model,
		TokensUsed:       response.Usage.TotalTokens,
	}

	rqp.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"processing_time": duration,
		"tools_used":      len(toolResults),
		"tokens_used":     response.Usage.TotalTokens,
	}).Info("Recommendation query processed successfully")

	return result, nil
}

// RecommendationQueryResult represents the result of processing a recommendation query
type RecommendationQueryResult struct {
	UserID         string                   `json:"user_id"`
	Query          string                   `json:"query"`
	Response       string                   `json:"response"`
	ToolResults    []*ToolExecutionResult   `json:"tool_results"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Timestamp      time.Time                `json:"timestamp"`
	Model          string                   `json:"model"`
	TokensUsed     int                      `json:"tokens_used"`
}