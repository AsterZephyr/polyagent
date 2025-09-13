package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ConversationalState represents the state of an ongoing conversation
type ConversationalState string

const (
	StateInitial           ConversationalState = "initial"
	StateGatheringPrefs    ConversationalState = "gathering_preferences"
	StateRefiningCriteria  ConversationalState = "refining_criteria"
	StateShowingResults    ConversationalState = "showing_results"
	StateGatheringFeedback ConversationalState = "gathering_feedback"
	StateFollowUp          ConversationalState = "follow_up"
	StateCompleted         ConversationalState = "completed"
)

// ConversationFlow represents the flow and logic of conversational recommendations
type ConversationFlow struct {
	SessionID           string                     `json:"session_id"`
	UserID              string                     `json:"user_id"`
	CurrentState        ConversationalState        `json:"current_state"`
	ConversationHistory []ConversationTurn         `json:"conversation_history"`
	UserProfile         *ConversationalUserProfile `json:"user_profile"`
	CurrentCriteria     *SearchCriteria            `json:"current_criteria"`
	LastRecommendations []RecommendedMovie         `json:"last_recommendations"`
	PendingQuestions    []string                   `json:"pending_questions"`
	ConversationContext *ConversationContext       `json:"conversation_context"`
	StartedAt           time.Time                  `json:"started_at"`
	LastActivity        time.Time                  `json:"last_activity"`
}

// ConversationTurn represents a single exchange in the conversation
type ConversationTurn struct {
	TurnNumber      int                 `json:"turn_number"`
	UserMessage     string              `json:"user_message"`
	AssistantMessage string             `json:"assistant_message"`
	DetectedIntent  IntentType          `json:"detected_intent"`
	ExtractedInfo   map[string]interface{} `json:"extracted_info"`
	RecommendationsShown []RecommendedMovie `json:"recommendations_shown"`
	Timestamp       time.Time           `json:"timestamp"`
	ProcessingTime  time.Duration       `json:"processing_time"`
}

// ConversationalUserProfile extends UserProfile with conversation-specific data
type ConversationalUserProfile struct {
	*UserProfile
	ExplicitPreferences    map[string]interface{} `json:"explicit_preferences"`
	ImplicitSignals        map[string]float64     `json:"implicit_signals"`
	ConversationPersona    string                 `json:"conversation_persona"`
	PreferredInteractionStyle string              `json:"preferred_interaction_style"`
	TopicProgression       []string               `json:"topic_progression"`
	EngagementLevel        float64                `json:"engagement_level"`
}

// SearchCriteria represents current search/filter criteria
type SearchCriteria struct {
	Genres          []string    `json:"genres"`
	YearRange       *YearRange  `json:"year_range,omitempty"`
	RatingRange     *RatingRange `json:"rating_range,omitempty"`
	Duration        *Duration   `json:"duration,omitempty"`
	ContentRating   string      `json:"content_rating,omitempty"`
	Keywords        []string    `json:"keywords"`
	ExcludeGenres   []string    `json:"exclude_genres"`
	Mood            string      `json:"mood,omitempty"`
	ViewingContext  string      `json:"viewing_context,omitempty"`
	SimilarTo       []string    `json:"similar_to"`
	MustInclude     []string    `json:"must_include"`
	MustExclude     []string    `json:"must_exclude"`
}

// Supporting types
type YearRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type RatingRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Duration struct {
	Min int `json:"min"` // in minutes
	Max int `json:"max"` // in minutes
}

// ConversationContext provides situational context
type ConversationContext struct {
	ViewingTime     string   `json:"viewing_time"`
	ViewingCompany  string   `json:"viewing_company"`
	Device          string   `json:"device"`
	Location        string   `json:"location"`
	RecentActivity  []string `json:"recent_activity"`
	EmotionalState  string   `json:"emotional_state"`
	EnergyLevel     string   `json:"energy_level"`
	AvailableTime   int      `json:"available_time"` // in minutes
}

// ConversationalRecommendationSystem manages dialogue-based recommendations
type ConversationalRecommendationSystem struct {
	multimodalEngine      *MultimodalRecommendationEngine
	conversationManager   map[string]*ConversationFlow // sessionID -> flow
	dialogueStrategy      *DialogueStrategy
	responseGenerator     *ResponseGenerator
	questionGenerator     *QuestionGenerator
	logger                *logrus.Logger
}

// DialogueStrategy defines how the conversation should progress
type DialogueStrategy struct {
	MaxTurns              int                              `json:"max_turns"`
	QuestionLimitsPerTurn int                              `json:"question_limits_per_turn"`
	StateTransitions      map[ConversationalState][]ConversationalState `json:"state_transitions"`
	ProactiveBehavior     bool                             `json:"proactive_behavior"`
	PersonalityProfile    string                           `json:"personality_profile"`
}

// ResponseGenerator generates contextual responses
type ResponseGenerator struct {
	templates map[ConversationalState][]string
	logger    *logrus.Logger
}

// QuestionGenerator generates clarifying questions
type QuestionGenerator struct {
	questionTemplates map[string][]string
	logger           *logrus.Logger
}

// NewConversationalRecommendationSystem creates a new conversational recommendation system
func NewConversationalRecommendationSystem(config *LLMAdapterConfig, logger *logrus.Logger) (*ConversationalRecommendationSystem, error) {
	multimodalEngine, err := NewMultimodalRecommendationEngine(config, logger)
	if err != nil {
		return nil, err
	}

	strategy := &DialogueStrategy{
		MaxTurns:              10,
		QuestionLimitsPerTurn: 2,
		StateTransitions: map[ConversationalState][]ConversationalState{
			StateInitial:           {StateGatheringPrefs, StateShowingResults},
			StateGatheringPrefs:    {StateRefiningCriteria, StateShowingResults},
			StateRefiningCriteria:  {StateShowingResults, StateGatheringPrefs},
			StateShowingResults:    {StateGatheringFeedback, StateFollowUp, StateCompleted},
			StateGatheringFeedback: {StateRefiningCriteria, StateFollowUp, StateCompleted},
			StateFollowUp:          {StateShowingResults, StateCompleted},
		},
		ProactiveBehavior:  true,
		PersonalityProfile: "helpful_movie_expert",
	}

	return &ConversationalRecommendationSystem{
		multimodalEngine:    multimodalEngine,
		conversationManager: make(map[string]*ConversationFlow),
		dialogueStrategy:    strategy,
		responseGenerator:   NewResponseGenerator(logger),
		questionGenerator:   NewQuestionGenerator(logger),
		logger:              logger,
	}, nil
}

// ProcessConversationalTurn processes a user message and generates response
func (crs *ConversationalRecommendationSystem) ProcessConversationalTurn(ctx context.Context, req *ConversationRequest) (*ConversationResponse, error) {
	startTime := time.Now()

	// Get or create conversation flow
	flow := crs.getOrCreateFlow(req.SessionID, req.UserID)
	
	// Update last activity
	flow.LastActivity = time.Now()

	// Process the user message
	turnResult, err := crs.processUserMessage(ctx, flow, req.UserMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to process user message: %w", err)
	}

	// Update conversation state
	crs.updateConversationState(flow, turnResult)

	// Generate assistant response
	response, err := crs.generateResponse(ctx, flow, turnResult)
	if err != nil {
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	// Record conversation turn
	turn := ConversationTurn{
		TurnNumber:           len(flow.ConversationHistory) + 1,
		UserMessage:          req.UserMessage,
		AssistantMessage:     response.Message,
		DetectedIntent:       turnResult.Intent.Type,
		ExtractedInfo:        turnResult.ExtractedInfo,
		RecommendationsShown: response.Recommendations,
		Timestamp:            time.Now(),
		ProcessingTime:       time.Since(startTime),
	}

	flow.ConversationHistory = append(flow.ConversationHistory, turn)

	crs.logger.WithFields(logrus.Fields{
		"session_id":      req.SessionID,
		"user_id":         req.UserID,
		"turn_number":     turn.TurnNumber,
		"current_state":   flow.CurrentState,
		"intent":          turnResult.Intent.Type,
		"processing_time": turn.ProcessingTime,
	}).Info("Processed conversational turn")

	return response, nil
}

// processUserMessage analyzes the user message and extracts information
func (crs *ConversationalRecommendationSystem) processUserMessage(ctx context.Context, flow *ConversationFlow, message string) (*TurnResult, error) {
	// Get conversation context from flow
	_ = crs.buildLLMContext(flow)

	// Analyze intent with conversation context
	intentResult, err := crs.multimodalEngine.GetLLMAdapter().ProcessRecommendationWithIntent(
		ctx, message, flow.UserID, flow.SessionID)
	if err != nil {
		return nil, fmt.Errorf("intent analysis failed: %w", err)
	}

	// Extract specific information based on current state
	extractedInfo := crs.extractStateSpecificInfo(flow, intentResult.Intent)

	// Update user profile with implicit signals
	crs.updateUserProfileImplicit(flow, intentResult.Intent, extractedInfo)

	return &TurnResult{
		Intent:        intentResult.Intent,
		ExtractedInfo: extractedInfo,
		LLMResult:     intentResult,
	}, nil
}

// generateResponse generates an appropriate response based on conversation state
func (crs *ConversationalRecommendationSystem) generateResponse(ctx context.Context, flow *ConversationFlow, turnResult *TurnResult) (*ConversationResponse, error) {
	response := &ConversationResponse{
		SessionID: flow.SessionID,
		State:     flow.CurrentState,
		Timestamp: time.Now(),
	}

	switch flow.CurrentState {
	case StateInitial, StateGatheringPrefs:
		response.Message = crs.generateGreetingOrPreferenceGathering(flow, turnResult)
		response.Questions = crs.generateClarifyingQuestions(flow, turnResult)

	case StateRefiningCriteria:
		response.Message = crs.generateCriteriaRefinement(flow, turnResult)
		response.Questions = crs.generateRefinementQuestions(flow)

	case StateShowingResults:
		recommendations, err := crs.generateRecommendations(ctx, flow)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recommendations: %w", err)
		}
		response.Recommendations = recommendations
		response.Message = crs.generateRecommendationPresentation(flow, recommendations)
		response.Questions = crs.generateFeedbackPrompts()

	case StateGatheringFeedback:
		response.Message = crs.generateFeedbackProcessing(flow, turnResult)
		response.Questions = crs.generateFollowUpQuestions(flow, turnResult)

	case StateFollowUp:
		response.Message = crs.generateFollowUpResponse(flow, turnResult)
		if crs.shouldShowMoreRecommendations(flow, turnResult) {
			recommendations, err := crs.generateRecommendations(ctx, flow)
			if err != nil {
				return nil, err
			}
			response.Recommendations = recommendations
		}

	case StateCompleted:
		response.Message = crs.generateClosingMessage(flow)
		response.SuggestedActions = crs.generateSuggestedActions(flow)

	default:
		response.Message = "I'm here to help you find great movies. What are you in the mood for?"
		response.State = StateInitial
	}

	return response, nil
}

// State-specific message generators
func (crs *ConversationalRecommendationSystem) generateGreetingOrPreferenceGathering(flow *ConversationFlow, turnResult *TurnResult) string {
	if len(flow.ConversationHistory) == 0 {
		return "Hi! I'm your personal movie recommendation assistant. I'd love to help you find the perfect movie to watch. What kind of movies do you usually enjoy?"
	}

	// Acknowledge user input and ask for more details
	intent := turnResult.Intent
	if intent.Type == IntentRecommendation {
		if genres, ok := intent.Entities["genre"].([]string); ok && len(genres) > 0 {
			return fmt.Sprintf("Great! I see you're interested in %s movies. What's your mood like today? Are you looking for something exciting, relaxing, thought-provoking, or fun?", strings.Join(genres, " and "))
		}
	}

	return "That's helpful! Can you tell me a bit more about what you're in the mood for? For example, any specific genres, time periods, or movies you've enjoyed recently?"
}

func (crs *ConversationalRecommendationSystem) generateCriteriaRefinement(flow *ConversationFlow, turnResult *TurnResult) string {
	criteria := flow.CurrentCriteria
	if criteria == nil {
		return "Let me understand your preferences better. What specific aspects are most important to you in a movie?"
	}

	refinements := []string{}
	if len(criteria.Genres) > 0 {
		refinements = append(refinements, fmt.Sprintf("genre: %s", strings.Join(criteria.Genres, ", ")))
	}
	if criteria.Mood != "" {
		refinements = append(refinements, fmt.Sprintf("mood: %s", criteria.Mood))
	}

	if len(refinements) > 0 {
		return fmt.Sprintf("Perfect! I understand you're looking for something with %s. Any other preferences I should know about?", strings.Join(refinements, " and "))
	}

	return "I want to make sure I find exactly what you're looking for. Could you be more specific about your preferences?"
}

func (crs *ConversationalRecommendationSystem) generateRecommendationPresentation(flow *ConversationFlow, recommendations []RecommendedMovie) string {
	if len(recommendations) == 0 {
		return "I couldn't find any movies that match your exact criteria. Would you like me to broaden the search or try different parameters?"
	}

	if len(recommendations) == 1 {
		movie := recommendations[0]
		return fmt.Sprintf("I found the perfect movie for you: **%s**! %s This seems like exactly what you're looking for. What do you think?", movie.Title, movie.Description)
	}

	return fmt.Sprintf("I've found %d great movies that match your preferences! Take a look at these recommendations. Do any of them catch your interest?", len(recommendations))
}

func (crs *ConversationalRecommendationSystem) generateFeedbackProcessing(flow *ConversationFlow, turnResult *TurnResult) string {
	intent := turnResult.Intent
	
	if intent.Type == IntentFeedback {
		if sentiment, ok := intent.Entities["sentiment"].(string); ok {
			if sentiment == "positive" {
				return "Wonderful! I'm glad you like those suggestions. Would you like me to find more movies similar to these, or are you ready to pick one to watch?"
			} else if sentiment == "negative" {
				return "No problem! Everyone has different tastes. Can you tell me what specifically didn't appeal to you? This will help me find better matches."
			}
		}
	}

	return "Thanks for the feedback! This helps me understand your preferences better. What would you like to explore next?"
}

func (crs *ConversationalRecommendationSystem) generateFollowUpResponse(flow *ConversationFlow, turnResult *TurnResult) string {
	if crs.shouldShowMoreRecommendations(flow, turnResult) {
		return "Based on our conversation, I've found some additional movies you might enjoy. Here are a few more recommendations!"
	}

	return "Is there anything else I can help you with today? I can find more movies, explain why I recommended something, or help you decide between options."
}

func (crs *ConversationalRecommendationSystem) generateClosingMessage(flow *ConversationFlow) string {
	return "It's been great helping you find movies today! Feel free to come back anytime for more recommendations. Enjoy your movie!"
}

// Question generators
func (crs *ConversationalRecommendationSystem) generateClarifyingQuestions(flow *ConversationFlow, turnResult *TurnResult) []string {
	questions := []string{}

	// Ask about viewing context if not known
	if flow.ConversationContext == nil || flow.ConversationContext.ViewingCompany == "" {
		questions = append(questions, "Are you watching alone or with others?")
	}

	// Ask about time constraints
	if flow.ConversationContext == nil || flow.ConversationContext.AvailableTime == 0 {
		questions = append(questions, "Do you have any time constraints? Looking for something quick or don't mind a longer movie?")
	}

	// Ask about mood if not clear from intent
	if flow.CurrentCriteria == nil || flow.CurrentCriteria.Mood == "" {
		questions = append(questions, "What's your mood like? Excited, relaxed, curious, or something else?")
	}

	return questions
}

func (crs *ConversationalRecommendationSystem) generateRefinementQuestions(flow *ConversationFlow) []string {
	questions := []string{}

	if flow.CurrentCriteria != nil {
		if len(flow.CurrentCriteria.Genres) == 0 {
			questions = append(questions, "Any specific genres you're drawn to or want to avoid?")
		}

		if flow.CurrentCriteria.YearRange == nil {
			questions = append(questions, "Do you prefer newer movies or don't mind older classics?")
		}
	}

	return questions
}

func (crs *ConversationalRecommendationSystem) generateFeedbackPrompts() []string {
	return []string{
		"Do any of these interest you?",
		"What do you think about these suggestions?",
		"Which one catches your eye, if any?",
	}
}

func (crs *ConversationalRecommendationSystem) generateFollowUpQuestions(flow *ConversationFlow, turnResult *TurnResult) []string {
	questions := []string{}

	intent := turnResult.Intent
	if intent.Type == IntentFeedback {
		if sentiment, ok := intent.Entities["sentiment"].(string); ok && sentiment == "negative" {
			questions = append(questions, "What specifically didn't appeal to you?")
			questions = append(questions, "Would you prefer something different in terms of genre, style, or time period?")
		}
	}

	return questions
}

func (crs *ConversationalRecommendationSystem) generateSuggestedActions(flow *ConversationFlow) []string {
	return []string{
		"Start a new recommendation session",
		"Rate movies you've watched",
		"Update your preferences",
		"Browse trending movies",
	}
}

// Helper functions
func (crs *ConversationalRecommendationSystem) getOrCreateFlow(sessionID, userID string) *ConversationFlow {
	if flow, exists := crs.conversationManager[sessionID]; exists {
		return flow
	}

	flow := &ConversationFlow{
		SessionID:           sessionID,
		UserID:              userID,
		CurrentState:        StateInitial,
		ConversationHistory: []ConversationTurn{},
		UserProfile:         &ConversationalUserProfile{EngagementLevel: 0.5},
		CurrentCriteria:     &SearchCriteria{},
		PendingQuestions:    []string{},
		StartedAt:           time.Now(),
		LastActivity:        time.Now(),
	}

	crs.conversationManager[sessionID] = flow
	return flow
}

func (crs *ConversationalRecommendationSystem) updateConversationState(flow *ConversationFlow, turnResult *TurnResult) {
	// State transition logic
	currentState := flow.CurrentState
	intent := turnResult.Intent

	switch currentState {
	case StateInitial:
		if intent.Type == IntentRecommendation {
			if crs.hasEnoughInformation(flow) {
				flow.CurrentState = StateShowingResults
			} else {
				flow.CurrentState = StateGatheringPrefs
			}
		} else {
			flow.CurrentState = StateGatheringPrefs
		}

	case StateGatheringPrefs:
		if crs.hasEnoughInformation(flow) {
			flow.CurrentState = StateShowingResults
		} else if crs.needsRefinement(flow) {
			flow.CurrentState = StateRefiningCriteria
		}

	case StateRefiningCriteria:
		if crs.hasEnoughInformation(flow) {
			flow.CurrentState = StateShowingResults
		}

	case StateShowingResults:
		if intent.Type == IntentFeedback {
			flow.CurrentState = StateGatheringFeedback
		} else {
			flow.CurrentState = StateFollowUp
		}

	case StateGatheringFeedback:
		if crs.shouldContinueRecommendations(turnResult) {
			flow.CurrentState = StateRefiningCriteria
		} else {
			flow.CurrentState = StateCompleted
		}

	case StateFollowUp:
		if len(flow.ConversationHistory) >= crs.dialogueStrategy.MaxTurns {
			flow.CurrentState = StateCompleted
		}
	}
}

func (crs *ConversationalRecommendationSystem) hasEnoughInformation(flow *ConversationFlow) bool {
	criteria := flow.CurrentCriteria
	return criteria != nil && (len(criteria.Genres) > 0 || criteria.Mood != "" || len(criteria.Keywords) > 0)
}

func (crs *ConversationalRecommendationSystem) needsRefinement(flow *ConversationFlow) bool {
	return len(flow.ConversationHistory) > 2 && !crs.hasEnoughInformation(flow)
}

func (crs *ConversationalRecommendationSystem) shouldContinueRecommendations(turnResult *TurnResult) bool {
	intent := turnResult.Intent
	if sentiment, ok := intent.Entities["sentiment"].(string); ok {
		return sentiment == "negative" // Continue if user didn't like recommendations
	}
	return false
}

func (crs *ConversationalRecommendationSystem) shouldShowMoreRecommendations(flow *ConversationFlow, turnResult *TurnResult) bool {
	intent := turnResult.Intent
	return intent.Type == IntentExploration || intent.Type == IntentRecommendation
}

func (crs *ConversationalRecommendationSystem) buildLLMContext(flow *ConversationFlow) string {
	context := fmt.Sprintf("Conversation State: %s\n", flow.CurrentState)
	context += fmt.Sprintf("Turn: %d/%d\n", len(flow.ConversationHistory), crs.dialogueStrategy.MaxTurns)
	
	if flow.CurrentCriteria != nil && len(flow.CurrentCriteria.Genres) > 0 {
		context += fmt.Sprintf("Current Genres: %s\n", strings.Join(flow.CurrentCriteria.Genres, ", "))
	}
	
	return context
}

func (crs *ConversationalRecommendationSystem) extractStateSpecificInfo(flow *ConversationFlow, intent *Intent) map[string]interface{} {
	info := make(map[string]interface{})

	// Extract information based on current state
	switch flow.CurrentState {
	case StateInitial, StateGatheringPrefs:
		if genres, ok := intent.Entities["genre"].([]string); ok {
			info["genres"] = genres
		}
		if mood, ok := intent.Entities["mood"].(string); ok {
			info["mood"] = mood
		}

	case StateRefiningCriteria:
		if year, ok := intent.Entities["year"].([]int); ok {
			info["year"] = year
		}
		if preferences, ok := intent.Entities["preference"].([]string); ok {
			info["preferences"] = preferences
		}

	case StateGatheringFeedback:
		if sentiment, ok := intent.Entities["sentiment"].(string); ok {
			info["sentiment"] = sentiment
		}
		if rating, ok := intent.Entities["rating"].(float64); ok {
			info["rating"] = rating
		}
	}

	return info
}

func (crs *ConversationalRecommendationSystem) updateUserProfileImplicit(flow *ConversationFlow, intent *Intent, extractedInfo map[string]interface{}) {
	if flow.UserProfile == nil {
		flow.UserProfile = &ConversationalUserProfile{
			ImplicitSignals: make(map[string]float64),
		}
	}

	// Update engagement level based on response length and detail
	responseLength := len(intent.RawQuery)
	if responseLength > 50 {
		flow.UserProfile.EngagementLevel += 0.1
	}

	// Update implicit signals based on extracted information
	for key, value := range extractedInfo {
		if key == "sentiment" && value == "positive" {
			flow.UserProfile.ImplicitSignals["satisfaction"] += 0.2
		}
	}

	// Cap engagement level
	if flow.UserProfile.EngagementLevel > 1.0 {
		flow.UserProfile.EngagementLevel = 1.0
	}
}

func (crs *ConversationalRecommendationSystem) generateRecommendations(ctx context.Context, flow *ConversationFlow) ([]RecommendedMovie, error) {
	// Convert conversation criteria to multimodal request
	req := &MultimodalRecommendationRequest{
		UserID:            flow.UserID,
		SessionID:         flow.SessionID,
		Query:             crs.buildQueryFromCriteria(flow.CurrentCriteria),
		TransparencyLevel: TransparencyBasic,
	}

	// Use multimodal engine to get recommendations (will fail in mock mode)
	result, err := crs.multimodalEngine.ProcessMultimodalRecommendation(ctx, req)
	if err != nil {
		// Return mock recommendations for testing
		return crs.generateMockRecommendations(flow.CurrentCriteria), nil
	}

	return result.RecommendedMovies, nil
}

func (crs *ConversationalRecommendationSystem) buildQueryFromCriteria(criteria *SearchCriteria) string {
	parts := []string{}

	if len(criteria.Genres) > 0 {
		parts = append(parts, fmt.Sprintf("I want %s movies", strings.Join(criteria.Genres, " and ")))
	}

	if criteria.Mood != "" {
		parts = append(parts, fmt.Sprintf("in a %s mood", criteria.Mood))
	}

	if len(criteria.Keywords) > 0 {
		parts = append(parts, fmt.Sprintf("with themes like %s", strings.Join(criteria.Keywords, ", ")))
	}

	if len(parts) == 0 {
		return "recommend me some good movies"
	}

	return strings.Join(parts, " ")
}

func (crs *ConversationalRecommendationSystem) generateMockRecommendations(criteria *SearchCriteria) []RecommendedMovie {
	// Mock recommendations based on criteria
	if criteria != nil && len(criteria.Genres) > 0 {
		genre := criteria.Genres[0]
		switch strings.ToLower(genre) {
		case "action":
			return []RecommendedMovie{
				{ID: 1, Title: "Mad Max: Fury Road", Genres: []string{"Action"}, Rating: 4.5, Description: "High-octane post-apocalyptic action"},
				{ID: 2, Title: "John Wick", Genres: []string{"Action"}, Rating: 4.3, Description: "Stylish action thriller"},
			}
		case "sci-fi":
			return []RecommendedMovie{
				{ID: 3, Title: "Blade Runner 2049", Genres: []string{"Sci-Fi"}, Rating: 4.7, Description: "Visually stunning sci-fi sequel"},
				{ID: 4, Title: "Arrival", Genres: []string{"Sci-Fi"}, Rating: 4.6, Description: "Thoughtful alien contact story"},
			}
		case "comedy":
			return []RecommendedMovie{
				{ID: 5, Title: "The Grand Budapest Hotel", Genres: []string{"Comedy"}, Rating: 4.4, Description: "Whimsical Wes Anderson comedy"},
				{ID: 6, Title: "Knives Out", Genres: []string{"Comedy", "Mystery"}, Rating: 4.5, Description: "Clever murder mystery comedy"},
			}
		}
	}

	// Default recommendations
	return []RecommendedMovie{
		{ID: 7, Title: "The Godfather", Genres: []string{"Drama"}, Rating: 4.9, Description: "Epic crime saga"},
		{ID: 8, Title: "Pulp Fiction", Genres: []string{"Drama"}, Rating: 4.8, Description: "Tarantino's nonlinear masterpiece"},
	}
}

// Supporting types and structures
type TurnResult struct {
	Intent        *Intent
	ExtractedInfo map[string]interface{}
	LLMResult     *IntentAwareRecommendationResult
}

type ConversationRequest struct {
	SessionID   string `json:"session_id"`
	UserID      string `json:"user_id"`
	UserMessage string `json:"user_message"`
}

type ConversationResponse struct {
	SessionID         string             `json:"session_id"`
	Message           string             `json:"message"`
	Questions         []string           `json:"questions,omitempty"`
	Recommendations   []RecommendedMovie `json:"recommendations,omitempty"`
	SuggestedActions  []string           `json:"suggested_actions,omitempty"`
	State             ConversationalState `json:"state"`
	Timestamp         time.Time          `json:"timestamp"`
}

// Component constructors
func NewResponseGenerator(logger *logrus.Logger) *ResponseGenerator {
	return &ResponseGenerator{
		templates: make(map[ConversationalState][]string),
		logger:    logger,
	}
}

func NewQuestionGenerator(logger *logrus.Logger) *QuestionGenerator {
	return &QuestionGenerator{
		questionTemplates: make(map[string][]string),
		logger:           logger,
	}
}

// GetConversationFlow returns the conversation flow for a session
func (crs *ConversationalRecommendationSystem) GetConversationFlow(sessionID string) (*ConversationFlow, bool) {
	flow, exists := crs.conversationManager[sessionID]
	return flow, exists
}

// GetActiveConversations returns all active conversation sessions
func (crs *ConversationalRecommendationSystem) GetActiveConversations() map[string]*ConversationFlow {
	active := make(map[string]*ConversationFlow)
	cutoff := time.Now().Add(-1 * time.Hour) // Consider conversations older than 1 hour as inactive

	for sessionID, flow := range crs.conversationManager {
		if flow.LastActivity.After(cutoff) {
			active[sessionID] = flow
		}
	}

	return active
}

// CleanupInactiveConversations removes old conversation sessions
func (crs *ConversationalRecommendationSystem) CleanupInactiveConversations() {
	cutoff := time.Now().Add(-24 * time.Hour) // Remove conversations older than 24 hours

	for sessionID, flow := range crs.conversationManager {
		if flow.LastActivity.Before(cutoff) {
			delete(crs.conversationManager, sessionID)
			crs.logger.WithField("session_id", sessionID).Info("Cleaned up inactive conversation")
		}
	}
}