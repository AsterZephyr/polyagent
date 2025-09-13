package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RemoteLLMManager manages remote LLM processing
type RemoteLLMManager struct {
	llmAdapter LLMAdapter
	logger     *logrus.Logger
	metrics    *RemoteLLMMetrics
}

// RemoteLLMMetrics tracks remote LLM processing metrics
type RemoteLLMMetrics struct {
	TotalRequests    int64         `json:"total_requests"`
	SuccessfulCalls  int64         `json:"successful_calls"`
	FailedCalls      int64         `json:"failed_calls"`
	AverageLatency   time.Duration `json:"average_latency"`
	TotalTokensUsed  int64         `json:"total_tokens_used"`
	EstimatedCost    float64       `json:"estimated_cost"`
	LastExecution    time.Time     `json:"last_execution"`
}

// RemoteLLMResult represents the result of remote LLM processing
type RemoteLLMResult struct {
	Result         interface{}   `json:"result"`
	QualityScore   float64       `json:"quality_score"`
	ProcessingTime time.Duration `json:"processing_time"`
	TokensUsed     int           `json:"tokens_used"`
	Provider       LLMProvider   `json:"provider"`
	Model          string        `json:"model"`
	Explanation    string        `json:"explanation,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewRemoteLLMManager creates a new remote LLM manager
func NewRemoteLLMManager(llmAdapter LLMAdapter, logger *logrus.Logger) *RemoteLLMManager {
	return &RemoteLLMManager{
		llmAdapter: llmAdapter,
		logger:     logger,
		metrics: &RemoteLLMMetrics{
			LastExecution: time.Now(),
		},
	}
}

// ExecuteTask executes a task using remote LLM
func (rlm *RemoteLLMManager) ExecuteTask(ctx context.Context, req *HybridExecutionRequest) (*RemoteLLMResult, error) {
	startTime := time.Now()
	
	rlm.logger.WithFields(logrus.Fields{
		"task_type": req.TaskType,
		"task_id":   req.TaskID,
	}).Info("Executing task with remote LLM")

	// Generate task-specific prompt
	prompt, err := rlm.generateTaskPrompt(req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Prepare LLM request
	llmReq := &GenerateRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: rlm.getSystemPrompt(req.TaskType),
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
		Metadata: map[string]interface{}{
			"task_type": req.TaskType,
			"task_id":   req.TaskID,
			"user_id":   req.UserID,
		},
	}

	// Execute LLM request
	response, err := rlm.llmAdapter.Generate(ctx, llmReq)
	if err != nil {
		rlm.updateMetrics(false, time.Since(startTime), 0)
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Process LLM response
	result, qualityScore, err := rlm.processLLMResponse(req.TaskType, response)
	if err != nil {
		rlm.updateMetrics(false, time.Since(startTime), response.Usage.TotalTokens)
		return nil, fmt.Errorf("failed to process LLM response: %w", err)
	}

	processingTime := time.Since(startTime)
	rlm.updateMetrics(true, processingTime, response.Usage.TotalTokens)

	remoteLLMResult := &RemoteLLMResult{
		Result:         result,
		QualityScore:   qualityScore,
		ProcessingTime: processingTime,
		TokensUsed:     response.Usage.TotalTokens,
		Provider:       ProviderOpenAI, // This should be dynamically determined
		Model:          response.Model,
		Metadata: map[string]interface{}{
			"prompt_tokens":      response.Usage.PromptTokens,
			"completion_tokens":  response.Usage.CompletionTokens,
			"finish_reason":      response.Choices[0].FinishReason,
			"response_id":        response.ID,
		},
	}

	// Add explanation if available
	if explanation := rlm.extractExplanation(response); explanation != "" {
		remoteLLMResult.Explanation = explanation
	}

	rlm.logger.WithFields(logrus.Fields{
		"processing_time": processingTime,
		"tokens_used":     response.Usage.TotalTokens,
		"quality_score":   qualityScore,
		"provider":        remoteLLMResult.Provider,
	}).Info("Remote LLM execution completed")

	return remoteLLMResult, nil
}

// generateTaskPrompt generates a task-specific prompt
func (rlm *RemoteLLMManager) generateTaskPrompt(req *HybridExecutionRequest) (string, error) {
	switch req.TaskType {
	case TaskMovieRecommendation:
		return rlm.generateMovieRecommendationPrompt(req)
	case TaskIntentAnalysis:
		return rlm.generateIntentAnalysisPrompt(req)
	case TaskExplanationGen:
		return rlm.generateExplanationPrompt(req)
	case TaskMultimodalAnalysis:
		return rlm.generateMultimodalAnalysisPrompt(req)
	case TaskUserProfiling:
		return rlm.generateUserProfilingPrompt(req)
	default:
		return "", fmt.Errorf("unsupported task type: %s", req.TaskType)
	}
}

// generateMovieRecommendationPrompt generates prompts for movie recommendations
func (rlm *RemoteLLMManager) generateMovieRecommendationPrompt(req *HybridExecutionRequest) (string, error) {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Generate movie recommendations based on the following information:\n\n")
	
	// Add user preferences
	if userPrefs, exists := req.Data["user_preferences"]; exists {
		prefsJSON, _ := json.Marshal(userPrefs)
		promptBuilder.WriteString(fmt.Sprintf("User Preferences: %s\n\n", string(prefsJSON)))
	}
	
	// Add viewing history
	if history, exists := req.Data["viewing_history"]; exists {
		historyJSON, _ := json.Marshal(history)
		promptBuilder.WriteString(fmt.Sprintf("Viewing History: %s\n\n", string(historyJSON)))
	}
	
	// Add contextual information
	if req.Context != nil {
		promptBuilder.WriteString("Context:\n")
		if req.Context.DeviceType != "" {
			promptBuilder.WriteString(fmt.Sprintf("- Device: %s\n", req.Context.DeviceType))
		}
		if !req.Context.Timestamp.IsZero() {
			promptBuilder.WriteString(fmt.Sprintf("- Time: %s\n", req.Context.Timestamp.Format("15:04")))
		}
		promptBuilder.WriteString("\n")
	}
	
	// Add specific requirements
	topK := 5
	if k, exists := req.Data["top_k"]; exists {
		if kInt, ok := k.(int); ok {
			topK = kInt
		}
	}
	
	promptBuilder.WriteString(fmt.Sprintf("Please provide exactly %d movie recommendations in JSON format with the following structure:\n", topK))
	promptBuilder.WriteString(`{
  "recommendations": [
    {
      "movie_id": "string",
      "title": "string",
      "rating": number,
      "genres": ["string"],
      "year": number,
      "explanation": "string"
    }
  ]
}`)
	
	return promptBuilder.String(), nil
}

// generateIntentAnalysisPrompt generates prompts for intent analysis
func (rlm *RemoteLLMManager) generateIntentAnalysisPrompt(req *HybridExecutionRequest) (string, error) {
	userMessage, ok := req.Data["message"].(string)
	if !ok {
		return "", fmt.Errorf("message not found in request data")
	}
	
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Analyze the user's intent from the following message:\n\n")
	promptBuilder.WriteString(fmt.Sprintf("User Message: \"%s\"\n\n", userMessage))
	
	// Add conversation context if available
	if context, exists := req.Data["conversation_context"]; exists {
		contextJSON, _ := json.Marshal(context)
		promptBuilder.WriteString(fmt.Sprintf("Conversation Context: %s\n\n", string(contextJSON)))
	}
	
	promptBuilder.WriteString("Classify the intent and extract relevant entities. Provide the response in JSON format:\n")
	promptBuilder.WriteString(`{
  "intent_type": "string (search_movies|get_recommendations|express_preference|ask_details|provide_feedback|general_chat|undefined)",
  "confidence": number (0-1),
  "entities": {
    "genres": ["string"],
    "actors": ["string"],
    "directors": ["string"],
    "movies": ["string"],
    "year_range": {"start": number, "end": number},
    "rating_preference": {"min": number, "max": number}
  },
  "user_preferences": {
    "explicit": ["string"],
    "implicit": ["string"]
  },
  "explanation": "string"
}`)
	
	return promptBuilder.String(), nil
}

// generateExplanationPrompt generates prompts for explanation generation
func (rlm *RemoteLLMManager) generateExplanationPrompt(req *HybridExecutionRequest) (string, error) {
	recommendations, ok := req.Data["recommendations"].([]interface{})
	if !ok {
		return "", fmt.Errorf("recommendations not found in request data")
	}
	
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Generate explanations for the following movie recommendations:\n\n")
	
	recsJSON, _ := json.Marshal(recommendations)
	promptBuilder.WriteString(fmt.Sprintf("Recommendations: %s\n\n", string(recsJSON)))
	
	// Add user profile if available
	if profile, exists := req.Data["user_profile"]; exists {
		profileJSON, _ := json.Marshal(profile)
		promptBuilder.WriteString(fmt.Sprintf("User Profile: %s\n\n", string(profileJSON)))
	}
	
	// Add explanation preferences
	explanationType := "detailed"
	if eType, exists := req.Data["explanation_type"]; exists {
		if typeStr, ok := eType.(string); ok {
			explanationType = typeStr
		}
	}
	
	promptBuilder.WriteString(fmt.Sprintf("Generate %s explanations for each recommendation. ", explanationType))
	promptBuilder.WriteString("Provide the response in JSON format:\n")
	promptBuilder.WriteString(`{
  "explanations": [
    {
      "movie_id": "string",
      "explanation": "string",
      "reasoning": "string",
      "confidence": number,
      "explanation_type": "string"
    }
  ],
  "overall_rationale": "string"
}`)
	
	return promptBuilder.String(), nil
}

// generateMultimodalAnalysisPrompt generates prompts for multimodal analysis
func (rlm *RemoteLLMManager) generateMultimodalAnalysisPrompt(req *HybridExecutionRequest) (string, error) {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Perform multimodal analysis of the provided content:\n\n")
	
	// Add text content
	if text, exists := req.Data["text"]; exists {
		promptBuilder.WriteString(fmt.Sprintf("Text: %s\n\n", text))
	}
	
	// Add image description
	if imageDesc, exists := req.Data["image_description"]; exists {
		promptBuilder.WriteString(fmt.Sprintf("Image Description: %s\n\n", imageDesc))
	}
	
	// Add audio description
	if audioDesc, exists := req.Data["audio_description"]; exists {
		promptBuilder.WriteString(fmt.Sprintf("Audio Description: %s\n\n", audioDesc))
	}
	
	promptBuilder.WriteString("Analyze the content and provide insights for movie recommendations. ")
	promptBuilder.WriteString("Provide the response in JSON format:\n")
	promptBuilder.WriteString(`{
  "content_analysis": {
    "mood": "string",
    "themes": ["string"],
    "style": "string",
    "emotional_tone": "string"
  },
  "recommendation_signals": {
    "preferred_genres": ["string"],
    "visual_preferences": ["string"],
    "audio_preferences": ["string"],
    "narrative_preferences": ["string"]
  },
  "confidence": number,
  "cross_modal_synthesis": "string"
}`)
	
	return promptBuilder.String(), nil
}

// generateUserProfilingPrompt generates prompts for user profiling
func (rlm *RemoteLLMManager) generateUserProfilingPrompt(req *HybridExecutionRequest) (string, error) {
	var promptBuilder strings.Builder
	
	promptBuilder.WriteString("Update the user profile based on the following interaction data:\n\n")
	
	// Add current profile
	if currentProfile, exists := req.Data["current_profile"]; exists {
		profileJSON, _ := json.Marshal(currentProfile)
		promptBuilder.WriteString(fmt.Sprintf("Current Profile: %s\n\n", string(profileJSON)))
	}
	
	// Add new interaction data
	if interactions, exists := req.Data["interactions"]; exists {
		interactionsJSON, _ := json.Marshal(interactions)
		promptBuilder.WriteString(fmt.Sprintf("New Interactions: %s\n\n", string(interactionsJSON)))
	}
	
	promptBuilder.WriteString("Update the user profile and provide the response in JSON format:\n")
	promptBuilder.WriteString(`{
  "updated_profile": {
    "preferred_genres": ["string"],
    "disliked_genres": ["string"],
    "preferred_actors": ["string"],
    "preferred_directors": ["string"],
    "viewing_patterns": {
      "preferred_time": "string",
      "typical_session_length": number,
      "device_preferences": ["string"]
    },
    "personality_traits": {
      "openness": number,
      "exploration_tendency": number,
      "quality_sensitivity": number
    }
  },
  "confidence_updates": {
    "genre_confidence": number,
    "actor_confidence": number,
    "overall_confidence": number
  },
  "profile_changes": ["string"]
}`)
	
	return promptBuilder.String(), nil
}

// getSystemPrompt returns task-specific system prompts
func (rlm *RemoteLLMManager) getSystemPrompt(taskType TaskType) string {
	switch taskType {
	case TaskMovieRecommendation:
		return "You are an expert movie recommendation system. Provide personalized, high-quality movie recommendations with clear explanations. Always respond in valid JSON format."
		
	case TaskIntentAnalysis:
		return "You are an expert at understanding user intent in conversational movie recommendation contexts. Analyze user messages to extract intent, entities, and preferences. Always respond in valid JSON format."
		
	case TaskExplanationGen:
		return "You are an expert at generating clear, personalized explanations for movie recommendations. Create explanations that help users understand why specific movies were recommended. Always respond in valid JSON format."
		
	case TaskMultimodalAnalysis:
		return "You are an expert at analyzing multimodal content (text, images, audio) to derive insights for movie recommendations. Extract themes, moods, and preferences from diverse content types. Always respond in valid JSON format."
		
	case TaskUserProfiling:
		return "You are an expert at building and updating user profiles for personalized recommendations. Analyze user interactions to infer preferences, patterns, and traits. Always respond in valid JSON format."
		
	default:
		return "You are a helpful AI assistant specialized in movie recommendations. Always respond in valid JSON format when requested."
	}
}

// processLLMResponse processes the LLM response based on task type
func (rlm *RemoteLLMManager) processLLMResponse(taskType TaskType, response *GenerateResponse) (interface{}, float64, error) {
	if len(response.Choices) == 0 {
		return nil, 0.0, fmt.Errorf("no choices in LLM response")
	}
	
	content := response.Choices[0].Message.Content
	
	// Try to parse JSON response
	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonResult); err != nil {
		// Fallback to text processing
		rlm.logger.Warn("Failed to parse JSON response, using text fallback")
		return rlm.processTextResponse(taskType, content)
	}
	
	// Calculate quality score based on response completeness
	qualityScore := rlm.calculateQualityScore(taskType, jsonResult, response)
	
	return jsonResult, qualityScore, nil
}

// processTextResponse processes non-JSON text responses
func (rlm *RemoteLLMManager) processTextResponse(taskType TaskType, content string) (interface{}, float64, error) {
	// Basic text processing fallback
	result := map[string]interface{}{
		"raw_response": content,
		"task_type":    taskType,
		"processed":    false,
		"fallback":     true,
	}
	
	// Basic quality assessment for text responses
	qualityScore := 0.5
	if len(content) > 100 {
		qualityScore = 0.6
	}
	if strings.Contains(strings.ToLower(content), "recommend") {
		qualityScore += 0.1
	}
	
	return result, qualityScore, nil
}

// calculateQualityScore calculates a quality score for the response
func (rlm *RemoteLLMManager) calculateQualityScore(taskType TaskType, result map[string]interface{}, response *GenerateResponse) float64 {
	baseScore := 0.7 // Base score for successful JSON parsing
	
	// Task-specific quality checks
	switch taskType {
	case TaskMovieRecommendation:
		if recs, exists := result["recommendations"]; exists {
			if recsList, ok := recs.([]interface{}); ok && len(recsList) > 0 {
				baseScore += 0.2
				// Check if recommendations have required fields
				if rec, ok := recsList[0].(map[string]interface{}); ok {
					fieldCount := 0
					for _, field := range []string{"title", "rating", "genres", "explanation"} {
						if _, exists := rec[field]; exists {
							fieldCount++
						}
					}
					baseScore += float64(fieldCount) * 0.025 // Up to 0.1 for all fields
				}
			}
		}
		
	case TaskIntentAnalysis:
		if intent, exists := result["intent_type"]; exists && intent != "" {
			baseScore += 0.1
		}
		if confidence, exists := result["confidence"]; exists {
			if confFloat, ok := confidence.(float64); ok && confFloat > 0.5 {
				baseScore += 0.1
			}
		}
		if entities, exists := result["entities"]; exists {
			if entMap, ok := entities.(map[string]interface{}); ok && len(entMap) > 0 {
				baseScore += 0.1
			}
		}
		
	case TaskExplanationGen:
		if explanations, exists := result["explanations"]; exists {
			if expList, ok := explanations.([]interface{}); ok && len(expList) > 0 {
				baseScore += 0.2
			}
		}
		if rationale, exists := result["overall_rationale"]; exists && rationale != "" {
			baseScore += 0.1
		}
	}
	
	// Response length and coherence bonus
	if len(response.Choices[0].Message.Content) > 200 {
		baseScore += 0.05
	}
	
	// Finish reason bonus
	if response.Choices[0].FinishReason == "stop" {
		baseScore += 0.05
	}
	
	// Cap at 1.0
	if baseScore > 1.0 {
		baseScore = 1.0
	}
	
	return baseScore
}

// extractExplanation extracts explanation from LLM response
func (rlm *RemoteLLMManager) extractExplanation(response *GenerateResponse) string {
	content := response.Choices[0].Message.Content
	
	// Try to extract explanation from JSON
	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(content), &jsonResult); err == nil {
		// Look for explanation fields
		for _, field := range []string{"explanation", "reasoning", "rationale"} {
			if exp, exists := jsonResult[field]; exists {
				if expStr, ok := exp.(string); ok && expStr != "" {
					return expStr
				}
			}
		}
	}
	
	// Fallback to first sentence of content
	sentences := strings.Split(content, ".")
	if len(sentences) > 0 && len(sentences[0]) > 20 {
		return sentences[0] + "."
	}
	
	return ""
}

// updateMetrics updates remote LLM metrics
func (rlm *RemoteLLMManager) updateMetrics(success bool, duration time.Duration, tokensUsed int) {
	rlm.metrics.TotalRequests++
	
	if success {
		rlm.metrics.SuccessfulCalls++
	} else {
		rlm.metrics.FailedCalls++
	}
	
	rlm.metrics.TotalTokensUsed += int64(tokensUsed)
	
	// Update average latency
	if rlm.metrics.TotalRequests == 1 {
		rlm.metrics.AverageLatency = duration
	} else {
		rlm.metrics.AverageLatency = (rlm.metrics.AverageLatency + duration) / 2
	}
	
	// Estimate cost (rough estimate: $0.002 per 1K tokens)
	rlm.metrics.EstimatedCost += float64(tokensUsed) * 0.000002
	
	rlm.metrics.LastExecution = time.Now()
}

// GetMetrics returns current metrics
func (rlm *RemoteLLMManager) GetMetrics() *RemoteLLMMetrics {
	// Return a copy to avoid race conditions
	metricsCopy := *rlm.metrics
	return &metricsCopy
}