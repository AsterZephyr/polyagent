package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ExplainableRecommendationEngine combines recommendation generation with explanations
type ExplainableRecommendationEngine struct {
	llmAdapter            *IntentAwareLLMAdapter
	explanationGenerator  *ExplanationGenerator
	conversationManager   *ConversationManager
	logger                *logrus.Logger
}

// NewExplainableRecommendationEngine creates a new explainable recommendation engine
func NewExplainableRecommendationEngine(config *LLMAdapterConfig, logger *logrus.Logger) (*ExplainableRecommendationEngine, error) {
	adapter, err := NewIntentAwareLLMAdapter(config, logger)
	if err != nil {
		return nil, err
	}

	return &ExplainableRecommendationEngine{
		llmAdapter:           adapter,
		explanationGenerator: NewExplanationGenerator(logger),
		conversationManager:  NewConversationManager(logger),
		logger:               logger,
	}, nil
}

// ProcessExplainableRecommendation processes a recommendation request with explanations
func (ere *ExplainableRecommendationEngine) ProcessExplainableRecommendation(ctx context.Context, req *ExplainableRecommendationRequest) (*ExplainableRecommendationResult, error) {
	startTime := time.Now()

	// Get conversation context
	_ = ere.conversationManager.GetConversationContext(req.UserID, req.SessionID)

	// Process the recommendation using intent-aware LLM
	intentResult, err := ere.llmAdapter.ProcessRecommendationWithIntent(ctx, req.Query, req.UserID, req.SessionID)
	if err != nil {
		ere.logger.WithError(err).Error("Failed to process intent-aware recommendation")
		return nil, fmt.Errorf("recommendation processing failed: %w", err)
	}

	// Extract recommended movies from tool results
	recommendedMovies := ere.extractRecommendedMovies(intentResult.ToolResults)

	// Generate explanations for each recommended movie
	explanations := []RecommendationExplanation{}
	for _, movie := range recommendedMovies {
		explanationReq := ere.buildExplanationRequest(movie, req, intentResult.Intent)
		
		explanation, err := ere.explanationGenerator.GenerateExplanation(ctx, explanationReq)
		if err != nil {
			ere.logger.WithError(err).WithField("movie_id", movie.ID).Warn("Failed to generate explanation")
			continue
		}
		
		explanations = append(explanations, *explanation)
	}

	// Create enhanced response with explanations
	result := &ExplainableRecommendationResult{
		UserID:           req.UserID,
		SessionID:        req.SessionID,
		Query:            req.Query,
		Intent:           intentResult.Intent,
		Response:         intentResult.Response,
		RecommendedMovies: recommendedMovies,
		Explanations:     explanations,
		ToolResults:      intentResult.ToolResults,
		ProcessingTime:   time.Since(startTime),
		Timestamp:        time.Now(),
		Model:            intentResult.Model,
		TokensUsed:       intentResult.TokensUsed,
		OverallConfidence: ere.calculateOverallConfidence(explanations),
	}

	// Update conversation with explanations
	ere.updateConversationWithExplanations(req.UserID, req.SessionID, intentResult.Intent, req.Query, result)

	ere.logger.WithFields(logrus.Fields{
		"user_id":            req.UserID,
		"intent_type":        intentResult.Intent.Type,
		"movies_recommended": len(recommendedMovies),
		"explanations_generated": len(explanations),
		"processing_time":    result.ProcessingTime,
	}).Info("Processed explainable recommendation successfully")

	return result, nil
}

// extractRecommendedMovies extracts movie recommendations from tool results
func (ere *ExplainableRecommendationEngine) extractRecommendedMovies(toolResults []*ToolExecutionResult) []RecommendedMovie {
	movies := []RecommendedMovie{}

	for _, result := range toolResults {
		switch result.ToolName {
		case "search_movies":
			if data, ok := result.Result.(map[string]interface{}); ok {
				if moviesList, ok := data["movies"].([]interface{}); ok {
					for _, movieData := range moviesList {
						if movie, ok := movieData.(map[string]interface{}); ok {
							recommendedMovie := ere.parseMovieFromToolResult(movie)
							movies = append(movies, recommendedMovie)
						}
					}
				}
			}

		case "generate_recommendations":
			if data, ok := result.Result.(map[string]interface{}); ok {
				if recsList, ok := data["recommendations"].([]interface{}); ok {
					for _, recData := range recsList {
						if rec, ok := recData.(map[string]interface{}); ok {
							recommendedMovie := ere.parseRecommendationFromToolResult(rec)
							movies = append(movies, recommendedMovie)
						}
					}
				}
			}
		}
	}

	return movies
}

// parseMovieFromToolResult parses a movie from search tool results
func (ere *ExplainableRecommendationEngine) parseMovieFromToolResult(movieData map[string]interface{}) RecommendedMovie {
	movie := RecommendedMovie{}

	if id, ok := movieData["id"].(float64); ok {
		movie.ID = int(id)
	}
	if title, ok := movieData["title"].(string); ok {
		movie.Title = title
	}
	if rating, ok := movieData["rating"].(float64); ok {
		movie.Rating = rating
	}
	if year, ok := movieData["year"].(float64); ok {
		movie.Year = int(year)
	}
	if genres, ok := movieData["genre"].([]interface{}); ok {
		for _, genre := range genres {
			if g, ok := genre.(string); ok {
				movie.Genres = append(movie.Genres, g)
			}
		}
	}
	if description, ok := movieData["description"].(string); ok {
		movie.Description = description
	}

	return movie
}

// parseRecommendationFromToolResult parses a recommendation from recommendation tool results
func (ere *ExplainableRecommendationEngine) parseRecommendationFromToolResult(recData map[string]interface{}) RecommendedMovie {
	movie := RecommendedMovie{}

	if id, ok := recData["movie_id"].(float64); ok {
		movie.ID = int(id)
	}
	if title, ok := recData["title"].(string); ok {
		movie.Title = title
	}
	if confidence, ok := recData["confidence"].(float64); ok {
		movie.Confidence = confidence
	}
	if predictedRating, ok := recData["predicted_rating"].(float64); ok {
		movie.PredictedRating = predictedRating
	}
	if reason, ok := recData["reason"].(string); ok {
		movie.Reason = reason
	}

	return movie
}

// buildExplanationRequest builds an explanation request from recommendation data
func (ere *ExplainableRecommendationEngine) buildExplanationRequest(movie RecommendedMovie, req *ExplainableRecommendationRequest, intent *Intent) *ExplanationRequest {
	explanationReq := &ExplanationRequest{
		MovieID:           movie.ID,
		MovieTitle:        movie.Title,
		TransparencyLevel: req.TransparencyLevel,
		MovieFeatures:     map[string]interface{}{
			"genres":      movie.Genres,
			"rating":      movie.Rating,
			"year":        movie.Year,
			"description": movie.Description,
		},
		Context: &RecommendationContext{},
	}

	// Add context from intent
	if intent.Context != nil {
		explanationReq.Context.TimeOfDay = intent.Context.TimeOfDay
		explanationReq.Context.DayOfWeek = intent.Context.DayOfWeek
	}

	// Extract mood from intent entities
	if mood, ok := intent.Entities["mood"].(string); ok {
		explanationReq.Context.Mood = mood
	}

	// Build user profile from conversation history and intent
	if req.UserProfile != nil {
		explanationReq.UserProfile = req.UserProfile
	} else {
		// Create basic profile from intent entities
		explanationReq.UserProfile = &UserProfile{
			UserID:      req.UserID,
			Preferences: make(map[string]float64),
		}

		// Extract preferences from intent entities
		if genres, ok := intent.Entities["genre"].([]string); ok {
			for _, genre := range genres {
				explanationReq.UserProfile.Preferences[genre] = 0.8
			}
		}
	}

	// Add similar movies if available from intent or context
	if req.SimilarMovies != nil {
		explanationReq.SimilarMovies = req.SimilarMovies
	}

	return explanationReq
}

// calculateOverallConfidence calculates overall confidence across all explanations
func (ere *ExplainableRecommendationEngine) calculateOverallConfidence(explanations []RecommendationExplanation) float64 {
	if len(explanations) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, explanation := range explanations {
		totalConfidence += explanation.Confidence
	}

	return totalConfidence / float64(len(explanations))
}

// updateConversationWithExplanations updates conversation context with explanation data
func (ere *ExplainableRecommendationEngine) updateConversationWithExplanations(userID, sessionID string, intent *Intent, query string, result *ExplainableRecommendationResult) {
	// Create enhanced assistant message with explanations
	assistantMessage := result.Response

	if len(result.Explanations) > 0 {
		assistantMessage += "\n\nHere's why I recommended these movies:\n"
		for i, explanation := range result.Explanations {
			assistantMessage += fmt.Sprintf("%d. %s: %s\n", i+1, explanation.MovieTitle, explanation.PersonalizedText)
		}
	}

	ere.conversationManager.UpdateConversation(userID, sessionID, intent, query, assistantMessage)
}

// GenerateDetailedExplanation generates a detailed explanation for a specific movie
func (ere *ExplainableRecommendationEngine) GenerateDetailedExplanation(ctx context.Context, req *DetailedExplanationRequest) (*DetailedExplanationResult, error) {
	// Create comprehensive explanation request
	explanationReq := &ExplanationRequest{
		MovieID:           req.MovieID,
		MovieTitle:        req.MovieTitle,
		MovieFeatures:     req.MovieFeatures,
		UserProfile:       req.UserProfile,
		SimilarMovies:     req.SimilarMovies,
		Context:           req.Context,
		TransparencyLevel: TransparencyTechnical, // Always use highest transparency for detailed explanations
	}

	explanation, err := ere.explanationGenerator.GenerateExplanation(ctx, explanationReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate detailed explanation: %w", err)
	}

	// Generate additional insights using LLM
	insights, err := ere.generateLLMInsights(ctx, explanation, req)
	if err != nil {
		ere.logger.WithError(err).Warn("Failed to generate LLM insights")
		insights = ""
	}

	result := &DetailedExplanationResult{
		Explanation:    *explanation,
		LLMInsights:    insights,
		AnalysisDepth:  "comprehensive",
		GeneratedAt:    time.Now(),
	}

	return result, nil
}

// generateLLMInsights uses the LLM to generate additional insights about the recommendation
func (ere *ExplainableRecommendationEngine) generateLLMInsights(ctx context.Context, explanation *RecommendationExplanation, req *DetailedExplanationRequest) (string, error) {
	// Create prompt for LLM to analyze the explanation
	prompt := fmt.Sprintf(`Analyze this movie recommendation explanation and provide additional insights:

Movie: %s
Explanation Type: %s
Confidence: %.2f
Primary Reason: %s

Detailed Reasons:
%s

Please provide:
1. What this recommendation reveals about the user's taste
2. How well this movie fits their profile (1-10 scale)
3. Potential concerns or reasons they might not like it
4. Similar movies they should explore next

Keep the analysis concise but insightful.`,
		explanation.MovieTitle,
		explanation.ExplanationType,
		explanation.Confidence,
		explanation.PrimaryReason,
		ere.formatReasonsForLLM(explanation.DetailedReasons))

	// Generate insights using LLM
	request := &GenerateRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are an expert film analyst providing insights about movie recommendations and user preferences.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   800,
	}

	response, err := ere.llmAdapter.Generate(ctx, request)
	if err != nil {
		return "", err
	}

	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response generated")
}

// formatReasonsForLLM formats explanation reasons for LLM analysis
func (ere *ExplainableRecommendationEngine) formatReasonsForLLM(reasons []ExplanationReason) string {
	var formatted []string
	for i, reason := range reasons {
		formatted = append(formatted, fmt.Sprintf("%d. %s (weight: %.2f)", i+1, reason.Description, reason.Weight))
	}
	return fmt.Sprintf("%v", formatted)
}

// Supporting types for the explainable recommendation system

// ExplainableRecommendationRequest represents a request for explainable recommendations
type ExplainableRecommendationRequest struct {
	UserID            string            `json:"user_id"`
	SessionID         string            `json:"session_id"`
	Query             string            `json:"query"`
	UserProfile       *UserProfile      `json:"user_profile,omitempty"`
	SimilarMovies     []SimilarMovie    `json:"similar_movies,omitempty"`
	TransparencyLevel TransparencyLevel `json:"transparency_level"`
}

// ExplainableRecommendationResult represents the result with explanations
type ExplainableRecommendationResult struct {
	UserID            string                     `json:"user_id"`
	SessionID         string                     `json:"session_id"`
	Query             string                     `json:"query"`
	Intent            *Intent                    `json:"intent"`
	Response          string                     `json:"response"`
	RecommendedMovies []RecommendedMovie         `json:"recommended_movies"`
	Explanations      []RecommendationExplanation `json:"explanations"`
	ToolResults       []*ToolExecutionResult     `json:"tool_results"`
	ProcessingTime    time.Duration              `json:"processing_time"`
	Timestamp         time.Time                  `json:"timestamp"`
	Model             string                     `json:"model"`
	TokensUsed        int                        `json:"tokens_used"`
	OverallConfidence float64                    `json:"overall_confidence"`
}

// RecommendedMovie represents a recommended movie with metadata
type RecommendedMovie struct {
	ID              int      `json:"id"`
	Title           string   `json:"title"`
	Genres          []string `json:"genres"`
	Year            int      `json:"year"`
	Rating          float64  `json:"rating"`
	Description     string   `json:"description"`
	Confidence      float64  `json:"confidence"`
	PredictedRating float64  `json:"predicted_rating"`
	Reason          string   `json:"reason"`
}

// DetailedExplanationRequest represents a request for detailed explanation
type DetailedExplanationRequest struct {
	MovieID       int                    `json:"movie_id"`
	MovieTitle    string                 `json:"movie_title"`
	MovieFeatures map[string]interface{} `json:"movie_features"`
	UserProfile   *UserProfile           `json:"user_profile"`
	SimilarMovies []SimilarMovie         `json:"similar_movies"`
	Context       *RecommendationContext `json:"context"`
}

// DetailedExplanationResult represents a detailed explanation result
type DetailedExplanationResult struct {
	Explanation   RecommendationExplanation `json:"explanation"`
	LLMInsights   string                    `json:"llm_insights"`
	AnalysisDepth string                    `json:"analysis_depth"`
	GeneratedAt   time.Time                 `json:"generated_at"`
}

// GetExplanationGenerator returns the explanation generator for direct access
func (ere *ExplainableRecommendationEngine) GetExplanationGenerator() *ExplanationGenerator {
	return ere.explanationGenerator
}

// GetConversationManager returns the conversation manager for direct access  
func (ere *ExplainableRecommendationEngine) GetConversationManager() *ConversationManager {
	return ere.conversationManager
}

// GetLLMAdapter returns the LLM adapter for direct access
func (ere *ExplainableRecommendationEngine) GetLLMAdapter() *IntentAwareLLMAdapter {
	return ere.llmAdapter
}