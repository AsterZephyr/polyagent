package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// IntentAwareLLMAdapter enhances the LLM adapter with intent understanding
type IntentAwareLLMAdapter struct {
	*EnhancedLLMAdapter
	intentAnalyzer *IntentAnalyzer
	logger         *logrus.Logger
}

// NewIntentAwareLLMAdapter creates a new intent-aware LLM adapter
func NewIntentAwareLLMAdapter(config *LLMAdapterConfig, logger *logrus.Logger) (*IntentAwareLLMAdapter, error) {
	enhancedAdapter, err := NewEnhancedLLMAdapter(config, logger)
	if err != nil {
		return nil, err
	}

	return &IntentAwareLLMAdapter{
		EnhancedLLMAdapter: enhancedAdapter,
		intentAnalyzer:     NewIntentAnalyzer(logger),
		logger:             logger,
	}, nil
}

// ProcessRecommendationWithIntent processes a recommendation query with intent analysis
func (ia *IntentAwareLLMAdapter) ProcessRecommendationWithIntent(ctx context.Context, userQuery string, userID string, sessionID string) (*IntentAwareRecommendationResult, error) {
	startTime := time.Now()

	// Create intent context
	intentContext := &IntentContext{
		UserID:           userID,
		SessionID:        sessionID,
		ConversationTurn: 1, // TODO: Track conversation history
		TimeOfDay:        getCurrentTimeOfDay(),
		DayOfWeek:        time.Now().Weekday().String(),
		// TODO: Load previous intents and user preferences
	}

	// Analyze user intent
	intent, err := ia.intentAnalyzer.AnalyzeIntent(ctx, userQuery, intentContext)
	if err != nil {
		ia.logger.WithError(err).Error("Failed to analyze user intent")
		return nil, fmt.Errorf("intent analysis failed: %w", err)
	}

	// Create intent-aware system prompt
	systemPrompt := ia.createIntentAwareSystemPrompt(intent)

	// Create enhanced user message with intent information
	enhancedUserMessage := ia.createEnhancedUserMessage(userQuery, intent)

	// Build messages for LLM
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",
			Content: enhancedUserMessage,
		},
	}

	// Create LLM request
	req := &GenerateRequest{
		Messages:    messages,
		Temperature: ia.getTemperatureForIntent(intent.Type),
		MaxTokens:   ia.getMaxTokensForIntent(intent.Type),
	}

	// Generate response with tools
	response, toolResults, err := ia.GenerateWithTools(ctx, req)
	if err != nil {
		ia.logger.WithError(err).Error("Failed to generate intent-aware response")
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Create result
	result := &IntentAwareRecommendationResult{
		UserID:         userID,
		SessionID:      sessionID,
		Query:          userQuery,
		Intent:         intent,
		Response:       response.Choices[0].Message.Content,
		ToolResults:    toolResults,
		ProcessingTime: time.Since(startTime),
		Timestamp:      time.Now(),
		Model:          response.Model,
		TokensUsed:     response.Usage.TotalTokens,
		Confidence:     intent.Confidence,
	}

	ia.logger.WithFields(logrus.Fields{
		"user_id":         userID,
		"intent_type":     intent.Type,
		"confidence":      intent.Confidence,
		"processing_time": result.ProcessingTime,
		"tools_used":      len(toolResults),
	}).Info("Intent-aware recommendation processed successfully")

	return result, nil
}

// createIntentAwareSystemPrompt creates a system prompt based on detected intent
func (ia *IntentAwareLLMAdapter) createIntentAwareSystemPrompt(intent *Intent) string {
	basePrompt := `You are an expert movie recommendation assistant with advanced intent understanding capabilities.`

	switch intent.Type {
	case IntentRecommendation:
		return basePrompt + `

DETECTED INTENT: User wants movie recommendations.

Your primary goal is to provide personalized movie suggestions. Focus on:
- Using the analyze_user_preferences tool to understand their taste
- Applying appropriate content filters based on context
- Generating diverse, personalized recommendations
- Explaining why each recommendation matches their preferences

Available recommendation strategies:
- Collaborative filtering for users with similar tastes
- Content-based filtering for genre/style preferences  
- Hybrid approaches for comprehensive suggestions

Be conversational and ask clarifying questions if needed.`

	case IntentSearch:
		return basePrompt + `

DETECTED INTENT: User is searching for specific movie information.

Your primary goal is to help find specific movies or information. Focus on:
- Using the search_movies tool to find exact matches
- Providing detailed information about requested movies
- Offering related suggestions after answering their search query
- Being precise and informative in your responses

If the search yields multiple results, present them clearly and ask for clarification.`

	case IntentExploration:
		return basePrompt + `

DETECTED INTENT: User wants to explore and discover movies.

Your primary goal is to help them discover new content. Focus on:
- Using popularity analysis to show trending content
- Suggesting diverse genres and styles to broaden horizons
- Highlighting hidden gems and underrated movies
- Encouraging exploration with follow-up suggestions

Be enthusiastic about discovery and offer varied options.`

	case IntentComparison:
		return basePrompt + `

DETECTED INTENT: User wants to compare movies.

Your primary goal is to help them compare different movies. Focus on:
- Providing detailed comparisons of ratings, genres, themes
- Highlighting similarities and differences
- Using analytical tools to support comparisons
- Helping them make informed viewing decisions

Present comparisons in a clear, structured format.`

	case IntentPersonalization:
		return basePrompt + `

DETECTED INTENT: User wants to personalize their experience.

Your primary goal is to learn about and adapt to their preferences. Focus on:
- Updating their preference profile based on feedback
- Asking about their viewing habits and tastes
- Customizing future recommendations based on their input
- Building a comprehensive understanding of their preferences

Be patient and thorough in gathering preference information.`

	case IntentFeedback:
		return basePrompt + `

DETECTED INTENT: User is providing feedback on movies or recommendations.

Your primary goal is to learn from their feedback. Focus on:
- Recording their ratings and opinions using tracking tools
- Understanding what they liked or disliked about specific movies
- Using feedback to improve future recommendations
- Acknowledging their input and showing how it helps

Be grateful for feedback and show how you'll use it to improve.`

	case IntentInformation:
		return basePrompt + `

DETECTED INTENT: User wants detailed information about movies or cinema.

Your primary goal is to provide comprehensive information. Focus on:
- Delivering accurate, detailed information about requested topics
- Using search tools to verify facts and details
- Providing context and background information
- Being educational and informative in your responses

Prioritize accuracy and depth in your information delivery.`

	default:
		return basePrompt + `

INTENT UNCLEAR: Unable to clearly determine user's specific intention.

Your approach should be:
- Ask clarifying questions to better understand their needs
- Offer multiple types of assistance (search, recommendations, information)
- Be helpful and guide them toward their desired outcome
- Use available tools conservatively until intent is clearer

Focus on understanding what they're looking for before providing specific assistance.`
	}
}

// createEnhancedUserMessage creates an enhanced user message with intent information
func (ia *IntentAwareLLMAdapter) createEnhancedUserMessage(originalQuery string, intent *Intent) string {
	enhanced := fmt.Sprintf("User Query: %s\n\n", originalQuery)

	// Add detected entities if any
	if len(intent.Entities) > 0 {
		enhanced += "Detected Information:\n"
		for entityType, value := range intent.Entities {
			enhanced += fmt.Sprintf("- %s: %v\n", entityType, value)
		}
		enhanced += "\n"
	}

	// Add context information
	if intent.Context != nil {
		enhanced += "Context:\n"
		if intent.Context.TimeOfDay != "" {
			enhanced += fmt.Sprintf("- Time: %s\n", intent.Context.TimeOfDay)
		}
		if intent.Context.DayOfWeek != "" {
			enhanced += fmt.Sprintf("- Day: %s\n", intent.Context.DayOfWeek)
		}
		enhanced += "\n"
	}

	enhanced += fmt.Sprintf("Intent Confidence: %.2f\n\n", intent.Confidence)
	enhanced += "Please provide a helpful response using appropriate tools as needed."

	return enhanced
}

// getTemperatureForIntent returns appropriate temperature for different intent types
func (ia *IntentAwareLLMAdapter) getTemperatureForIntent(intentType IntentType) float64 {
	switch intentType {
	case IntentSearch, IntentInformation:
		return 0.3 // More focused and factual
	case IntentRecommendation, IntentExploration:
		return 0.7 // More creative and diverse
	case IntentComparison:
		return 0.4 // Balanced analytical approach
	case IntentPersonalization, IntentFeedback:
		return 0.5 // Moderate creativity with structure
	default:
		return 0.6 // Default balanced approach
	}
}

// getMaxTokensForIntent returns appropriate max tokens for different intent types
func (ia *IntentAwareLLMAdapter) getMaxTokensForIntent(intentType IntentType) int {
	switch intentType {
	case IntentInformation, IntentComparison:
		return 2500 // Longer responses for detailed information
	case IntentRecommendation, IntentExploration:
		return 2000 // Moderate length for recommendations
	case IntentSearch:
		return 1500 // Shorter, focused responses
	case IntentPersonalization, IntentFeedback:
		return 1800 // Moderate length for interaction
	default:
		return 1500 // Default length
	}
}

// getCurrentTimeOfDay returns current time period
func getCurrentTimeOfDay() string {
	hour := time.Now().Hour()
	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "afternoon"
	case hour >= 17 && hour < 21:
		return "evening"
	default:
		return "night"
	}
}

// IntentAwareRecommendationResult represents the result with intent analysis
type IntentAwareRecommendationResult struct {
	UserID         string                   `json:"user_id"`
	SessionID      string                   `json:"session_id"`
	Query          string                   `json:"query"`
	Intent         *Intent                  `json:"intent"`
	Response       string                   `json:"response"`
	ToolResults    []*ToolExecutionResult   `json:"tool_results"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Timestamp      time.Time                `json:"timestamp"`
	Model          string                   `json:"model"`
	TokensUsed     int                      `json:"tokens_used"`
	Confidence     float64                  `json:"confidence"`
}

// ConversationManager manages multi-turn conversations with intent history
type ConversationManager struct {
	conversations map[string]*Conversation
	logger        *logrus.Logger
}

// Conversation represents a user conversation session
type Conversation struct {
	UserID           string                            `json:"user_id"`
	SessionID        string                            `json:"session_id"`
	IntentHistory    []*Intent                         `json:"intent_history"`
	MessageHistory   []Message                         `json:"message_history"`
	UserPreferences  map[string]interface{}            `json:"user_preferences"`
	LastInteraction  time.Time                         `json:"last_interaction"`
	TotalInteractions int                              `json:"total_interactions"`
}

// NewConversationManager creates a new conversation manager
func NewConversationManager(logger *logrus.Logger) *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*Conversation),
		logger:        logger,
	}
}

// GetOrCreateConversation gets or creates a conversation session
func (cm *ConversationManager) GetOrCreateConversation(userID, sessionID string) *Conversation {
	key := fmt.Sprintf("%s:%s", userID, sessionID)
	
	if conv, exists := cm.conversations[key]; exists {
		return conv
	}

	conv := &Conversation{
		UserID:           userID,
		SessionID:        sessionID,
		IntentHistory:    make([]*Intent, 0),
		MessageHistory:   make([]Message, 0),
		UserPreferences:  make(map[string]interface{}),
		LastInteraction:  time.Now(),
		TotalInteractions: 0,
	}

	cm.conversations[key] = conv
	cm.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"session_id": sessionID,
	}).Info("Created new conversation session")

	return conv
}

// UpdateConversation updates conversation with new intent and messages
func (cm *ConversationManager) UpdateConversation(userID, sessionID string, intent *Intent, userMessage, assistantMessage string) {
	conv := cm.GetOrCreateConversation(userID, sessionID)
	
	conv.IntentHistory = append(conv.IntentHistory, intent)
	conv.MessageHistory = append(conv.MessageHistory, Message{
		Role:    "user",
		Content: userMessage,
	})
	conv.MessageHistory = append(conv.MessageHistory, Message{
		Role:    "assistant",
		Content: assistantMessage,
	})
	
	conv.LastInteraction = time.Now()
	conv.TotalInteractions++

	// Keep only last 20 interactions to manage memory
	if len(conv.IntentHistory) > 20 {
		conv.IntentHistory = conv.IntentHistory[len(conv.IntentHistory)-20:]
	}
	if len(conv.MessageHistory) > 40 { // 20 pairs of user/assistant messages
		conv.MessageHistory = conv.MessageHistory[len(conv.MessageHistory)-40:]
	}

	cm.logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"session_id":  sessionID,
		"intent_type": intent.Type,
		"turn_count":  conv.TotalInteractions,
	}).Info("Updated conversation session")
}

// GetConversationContext creates intent context from conversation history
func (cm *ConversationManager) GetConversationContext(userID, sessionID string) *IntentContext {
	conv := cm.GetOrCreateConversation(userID, sessionID)
	
	context := &IntentContext{
		UserID:           userID,
		SessionID:        sessionID,
		ConversationTurn: conv.TotalInteractions + 1,
		TimeOfDay:        getCurrentTimeOfDay(),
		DayOfWeek:        time.Now().Weekday().String(),
		UserPreferences:  conv.UserPreferences,
	}

	// Add previous intents (last 5)
	if len(conv.IntentHistory) > 0 {
		context.PreviousIntents = make([]IntentType, 0)
		start := len(conv.IntentHistory) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(conv.IntentHistory); i++ {
			context.PreviousIntents = append(context.PreviousIntents, conv.IntentHistory[i].Type)
		}
	}

	return context
}

// GetIntentAnalyzer returns the intent analyzer for direct access
func (ia *IntentAwareLLMAdapter) GetIntentAnalyzer() *IntentAnalyzer {
	return ia.intentAnalyzer
}