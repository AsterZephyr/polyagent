package llm

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// IntentType represents different types of user intentions
type IntentType string

const (
	IntentRecommendation     IntentType = "recommendation"
	IntentSearch            IntentType = "search"
	IntentExploration       IntentType = "exploration"
	IntentComparison        IntentType = "comparison"
	IntentInformation       IntentType = "information"
	IntentPersonalization   IntentType = "personalization"
	IntentFeedback          IntentType = "feedback"
	IntentUndefined         IntentType = "undefined"
)

// Intent represents a parsed user intention
type Intent struct {
	Type           IntentType             `json:"type"`
	Confidence     float64                `json:"confidence"`
	Entities       map[string]interface{} `json:"entities"`
	Context        *IntentContext         `json:"context"`
	RawQuery       string                 `json:"raw_query"`
	ParsedAt       time.Time              `json:"parsed_at"`
	Suggestions    []string               `json:"suggestions,omitempty"`
}

// IntentContext provides contextual information about the intent
type IntentContext struct {
	UserID           string                 `json:"user_id,omitempty"`
	SessionID        string                 `json:"session_id,omitempty"`
	ConversationTurn int                    `json:"conversation_turn"`
	TimeOfDay        string                 `json:"time_of_day"`
	DayOfWeek        string                 `json:"day_of_week"`
	PreviousIntents  []IntentType           `json:"previous_intents,omitempty"`
	UserPreferences  map[string]interface{} `json:"user_preferences,omitempty"`
}

// Entity represents extracted entities from user queries
type Entity struct {
	Type       string      `json:"type"`
	Value      interface{} `json:"value"`
	Confidence float64     `json:"confidence"`
	StartPos   int         `json:"start_pos"`
	EndPos     int         `json:"end_pos"`
}

// IntentAnalyzer analyzes user queries to extract intentions
type IntentAnalyzer struct {
	logger         *logrus.Logger
	patterns       map[IntentType][]*IntentPattern
	entityExtractor *EntityExtractor
	metrics        *IntentMetrics
}

// IntentPattern represents a pattern for matching intentions
type IntentPattern struct {
	Pattern    *regexp.Regexp
	Keywords   []string
	Weight     float64
	Examples   []string
}

// IntentMetrics tracks intent analysis performance
type IntentMetrics struct {
	TotalQueries      int64                    `json:"total_queries"`
	IntentCounts      map[IntentType]int64     `json:"intent_counts"`
	ConfidenceScores  map[IntentType][]float64 `json:"confidence_scores"`
	ProcessingTimes   []time.Duration          `json:"processing_times"`
	LastUpdated       time.Time                `json:"last_updated"`
}

// NewIntentAnalyzer creates a new intent analyzer
func NewIntentAnalyzer(logger *logrus.Logger) *IntentAnalyzer {
	analyzer := &IntentAnalyzer{
		logger:          logger,
		patterns:        make(map[IntentType][]*IntentPattern),
		entityExtractor: NewEntityExtractor(),
		metrics: &IntentMetrics{
			IntentCounts:     make(map[IntentType]int64),
			ConfidenceScores: make(map[IntentType][]float64),
			LastUpdated:      time.Now(),
		},
	}

	analyzer.initializePatterns()
	return analyzer
}

// AnalyzeIntent analyzes a user query to extract intent
func (ia *IntentAnalyzer) AnalyzeIntent(ctx context.Context, query string, context *IntentContext) (*Intent, error) {
	startTime := time.Now()
	defer func() {
		ia.updateMetrics(time.Since(startTime))
	}()

	// Normalize query
	normalizedQuery := ia.normalizeQuery(query)

	// Extract entities
	entities, err := ia.entityExtractor.ExtractEntities(normalizedQuery)
	if err != nil {
		ia.logger.WithError(err).Warn("Failed to extract entities")
		entities = make(map[string]interface{})
	}

	// Determine intent type and confidence
	intentType, confidence := ia.classifyIntent(normalizedQuery, entities)

	// Generate suggestions based on intent
	suggestions := ia.generateSuggestions(intentType, entities)

	intent := &Intent{
		Type:        intentType,
		Confidence:  confidence,
		Entities:    entities,
		Context:     context,
		RawQuery:    query,
		ParsedAt:    time.Now(),
		Suggestions: suggestions,
	}

	// Update metrics
	ia.metrics.IntentCounts[intentType]++
	ia.metrics.ConfidenceScores[intentType] = append(ia.metrics.ConfidenceScores[intentType], confidence)
	ia.metrics.TotalQueries++

	ia.logger.WithFields(logrus.Fields{
		"intent_type": intentType,
		"confidence":  confidence,
		"entities":    len(entities),
		"query":       query,
	}).Info("Intent analyzed successfully")

	return intent, nil
}

// classifyIntent determines the intent type and confidence score
func (ia *IntentAnalyzer) classifyIntent(query string, entities map[string]interface{}) (IntentType, float64) {
	bestIntent := IntentUndefined
	bestScore := 0.0

	queryLower := strings.ToLower(query)

	for intentType, patterns := range ia.patterns {
		score := 0.0

		for _, pattern := range patterns {
			// Pattern matching score
			if pattern.Pattern.MatchString(queryLower) {
				score += pattern.Weight * 0.4
			}

			// Keyword matching score
			keywordMatches := 0
			for _, keyword := range pattern.Keywords {
				if strings.Contains(queryLower, strings.ToLower(keyword)) {
					keywordMatches++
				}
			}
			if len(pattern.Keywords) > 0 {
				score += (float64(keywordMatches) / float64(len(pattern.Keywords))) * pattern.Weight * 0.3
			}
		}

		// Entity-based scoring
		entityScore := ia.calculateEntityScore(intentType, entities)
		score += entityScore * 0.3

		if score > bestScore {
			bestScore = score
			bestIntent = intentType
		}
	}

	// Normalize confidence score
	confidence := bestScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return bestIntent, confidence
}

// calculateEntityScore calculates score based on extracted entities
func (ia *IntentAnalyzer) calculateEntityScore(intentType IntentType, entities map[string]interface{}) float64 {
	score := 0.0

	switch intentType {
	case IntentRecommendation:
		if _, hasGenre := entities["genre"]; hasGenre {
			score += 0.3
		}
		if _, hasPreference := entities["preference"]; hasPreference {
			score += 0.3
		}
		if _, hasMood := entities["mood"]; hasMood {
			score += 0.2
		}

	case IntentSearch:
		if _, hasTitle := entities["movie_title"]; hasTitle {
			score += 0.4
		}
		if _, hasActor := entities["actor"]; hasActor {
			score += 0.3
		}
		if _, hasYear := entities["year"]; hasYear {
			score += 0.2
		}

	case IntentComparison:
		if titles, hasMovies := entities["movie_titles"]; hasMovies {
			if movieList, ok := titles.([]string); ok && len(movieList) >= 2 {
				score += 0.5
			}
		}

	case IntentFeedback:
		if _, hasRating := entities["rating"]; hasRating {
			score += 0.3
		}
		if _, hasSentiment := entities["sentiment"]; hasSentiment {
			score += 0.3
		}
	}

	return score
}

// generateSuggestions generates follow-up suggestions based on intent
func (ia *IntentAnalyzer) generateSuggestions(intentType IntentType, entities map[string]interface{}) []string {
	suggestions := []string{}

	switch intentType {
	case IntentRecommendation:
		suggestions = append(suggestions, 
			"Would you like me to consider your viewing history?",
			"Any specific genre or mood you're in for?",
			"Are you looking for recent releases or classics?")

	case IntentSearch:
		suggestions = append(suggestions,
			"I can help you find detailed information about movies",
			"Would you like similar movie recommendations?",
			"Interested in cast information or plot details?")

	case IntentExploration:
		suggestions = append(suggestions,
			"I can show you trending movies in different genres",
			"Would you like to explore by decade or director?",
			"Interested in discovering hidden gems?")

	case IntentComparison:
		suggestions = append(suggestions,
			"I can compare ratings, genres, and user reviews",
			"Would you like to see similar movies to both?",
			"Interested in which one matches your preferences better?")

	case IntentPersonalization:
		suggestions = append(suggestions,
			"I can learn from your ratings and preferences",
			"Would you like to update your genre preferences?",
			"Interested in creating a personalized watchlist?")

	case IntentFeedback:
		suggestions = append(suggestions,
			"Your feedback helps improve recommendations",
			"Would you like similar or different suggestions?",
			"Any specific aspects you liked or disliked?")

	default:
		suggestions = append(suggestions,
			"I can help you find movies, get recommendations, or explore new genres",
			"Try asking for movie suggestions or searching for specific titles",
			"I can also help you discover trending or popular movies")
	}

	return suggestions
}

// normalizeQuery cleans and normalizes the input query
func (ia *IntentAnalyzer) normalizeQuery(query string) string {
	// Remove extra whitespace
	normalized := regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(query), " ")
	
	// Remove special characters (keep basic punctuation)
	normalized = regexp.MustCompile(`[^\w\s\-'.,!?]`).ReplaceAllString(normalized, "")
	
	return normalized
}

// initializePatterns sets up intent classification patterns
func (ia *IntentAnalyzer) initializePatterns() {
	// Recommendation patterns
	ia.patterns[IntentRecommendation] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(recommend|suggest|find me|what should i watch|good movies?)`),
			Keywords: []string{"recommend", "suggest", "what to watch", "good movies", "similar"},
			Weight:   1.0,
			Examples: []string{"recommend me a movie", "what should I watch tonight", "suggest something good"},
		},
		{
			Pattern:  regexp.MustCompile(`(like|love|enjoyed|similar to)`),
			Keywords: []string{"like", "love", "similar", "enjoyed", "based on"},
			Weight:   0.8,
			Examples: []string{"I love action movies", "something similar to The Matrix"},
		},
	}

	// Search patterns
	ia.patterns[IntentSearch] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(find|search|look for|about|tell me about)`),
			Keywords: []string{"find", "search", "about", "information", "details"},
			Weight:   1.0,
			Examples: []string{"find movies with Tom Hanks", "tell me about Inception"},
		},
		{
			Pattern:  regexp.MustCompile(`(who|what|when|where|cast|director|plot)`),
			Keywords: []string{"who", "what", "when", "cast", "director", "plot", "story"},
			Weight:   0.9,
			Examples: []string{"who directed The Godfather", "what is Inception about"},
		},
	}

	// Exploration patterns
	ia.patterns[IntentExploration] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(explore|discover|browse|trending|popular|new)`),
			Keywords: []string{"explore", "discover", "trending", "popular", "new releases"},
			Weight:   1.0,
			Examples: []string{"explore sci-fi movies", "what's trending now"},
		},
		{
			Pattern:  regexp.MustCompile(`(random|surprise|anything|don't know)`),
			Keywords: []string{"random", "surprise", "anything", "don't know", "open to"},
			Weight:   0.7,
			Examples: []string{"surprise me", "anything good", "I don't know what to watch"},
		},
	}

	// Comparison patterns
	ia.patterns[IntentComparison] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(compare|vs|versus|between|which|better)`),
			Keywords: []string{"compare", "vs", "versus", "between", "which", "better", "difference"},
			Weight:   1.0,
			Examples: []string{"compare Avatar vs Titanic", "which is better"},
		},
	}

	// Information patterns
	ia.patterns[IntentInformation] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(how|why|explain|analysis|review)`),
			Keywords: []string{"how", "why", "explain", "analysis", "review", "critique"},
			Weight:   0.9,
			Examples: []string{"why is The Godfather so popular", "explain the plot"},
		},
	}

	// Personalization patterns
	ia.patterns[IntentPersonalization] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(my|preferences|profile|taste|favorite|hate)`),
			Keywords: []string{"my", "preferences", "profile", "taste", "favorite", "hate", "dislike"},
			Weight:   0.9,
			Examples: []string{"update my preferences", "I hate horror movies"},
		},
	}

	// Feedback patterns
	ia.patterns[IntentFeedback] = []*IntentPattern{
		{
			Pattern:  regexp.MustCompile(`(rate|rating|review|liked|disliked|good|bad|awful|amazing)`),
			Keywords: []string{"rate", "rating", "review", "liked", "disliked", "loved", "hated"},
			Weight:   0.8,
			Examples: []string{"I rate this 5 stars", "I didn't like that movie"},
		},
	}
}

// GetMetrics returns intent analysis metrics
func (ia *IntentAnalyzer) GetMetrics() *IntentMetrics {
	return ia.metrics
}

// updateMetrics updates processing time metrics
func (ia *IntentAnalyzer) updateMetrics(duration time.Duration) {
	ia.metrics.ProcessingTimes = append(ia.metrics.ProcessingTimes, duration)
	ia.metrics.LastUpdated = time.Now()
	
	// Keep only last 1000 processing times
	if len(ia.metrics.ProcessingTimes) > 1000 {
		ia.metrics.ProcessingTimes = ia.metrics.ProcessingTimes[len(ia.metrics.ProcessingTimes)-1000:]
	}
}

// EntityExtractor extracts named entities from user queries
type EntityExtractor struct {
	patterns map[string]*regexp.Regexp
}

// NewEntityExtractor creates a new entity extractor
func NewEntityExtractor() *EntityExtractor {
	extractor := &EntityExtractor{
		patterns: make(map[string]*regexp.Regexp),
	}
	extractor.initializePatterns()
	return extractor
}

// ExtractEntities extracts entities from a normalized query
func (ee *EntityExtractor) ExtractEntities(query string) (map[string]interface{}, error) {
	entities := make(map[string]interface{})
	queryLower := strings.ToLower(query)

	// Extract genres
	if genres := ee.extractGenres(queryLower); len(genres) > 0 {
		entities["genre"] = genres
	}

	// Extract years
	if years := ee.extractYears(queryLower); len(years) > 0 {
		entities["year"] = years
	}

	// Extract ratings
	if rating := ee.extractRating(queryLower); rating > 0 {
		entities["rating"] = rating
	}

	// Extract sentiment
	if sentiment := ee.extractSentiment(queryLower); sentiment != "" {
		entities["sentiment"] = sentiment
	}

	// Extract movie titles (basic pattern matching)
	if titles := ee.extractMovieTitles(queryLower); len(titles) > 0 {
		if len(titles) == 1 {
			entities["movie_title"] = titles[0]
		} else {
			entities["movie_titles"] = titles
		}
	}

	// Extract preferences
	if preferences := ee.extractPreferences(queryLower); len(preferences) > 0 {
		entities["preference"] = preferences
	}

	// Extract mood
	if mood := ee.extractMood(queryLower); mood != "" {
		entities["mood"] = mood
	}

	return entities, nil
}

// initializePatterns sets up entity extraction patterns
func (ee *EntityExtractor) initializePatterns() {
	ee.patterns["year"] = regexp.MustCompile(`\b(19|20)\d{2}\b`)
	ee.patterns["rating"] = regexp.MustCompile(`\b([1-5])\s*(star|out of 5|/5)\b`)
	ee.patterns["movie_title"] = regexp.MustCompile(`"([^"]+)"|'([^']+)'`)
}

// extractGenres extracts movie genres from query
func (ee *EntityExtractor) extractGenres(query string) []string {
	genres := []string{
		"action", "adventure", "animation", "comedy", "crime", "documentary",
		"drama", "family", "fantasy", "history", "horror", "music", "mystery",
		"romance", "science fiction", "sci-fi", "thriller", "war", "western",
	}

	var found []string
	for _, genre := range genres {
		if strings.Contains(query, genre) {
			found = append(found, genre)
		}
	}
	return found
}

// extractYears extracts years from query
func (ee *EntityExtractor) extractYears(query string) []int {
	matches := ee.patterns["year"].FindAllString(query, -1)
	var years []int
	
	for _, match := range matches {
		var year int
		if _, err := fmt.Sscanf(match, "%d", &year); err == nil {
			years = append(years, year)
		}
	}
	return years
}

// extractRating extracts rating from query
func (ee *EntityExtractor) extractRating(query string) float64 {
	// Look for numeric ratings
	patterns := []string{
		`\b([1-5])\s*star`,
		`\b([1-5])\s*out\s*of\s*5`,
		`\b([1-5])/5\b`,
		`\brate.*?([1-5])\b`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(query); len(matches) > 1 {
			var rating float64
			if _, err := fmt.Sscanf(matches[1], "%f", &rating); err == nil {
				return rating
			}
		}
	}

	return 0
}

// extractSentiment extracts sentiment from query
func (ee *EntityExtractor) extractSentiment(query string) string {
	positive := []string{"love", "like", "enjoy", "great", "amazing", "awesome", "good", "excellent"}
	negative := []string{"hate", "dislike", "awful", "terrible", "bad", "boring", "worst"}

	for _, word := range positive {
		if strings.Contains(query, word) {
			return "positive"
		}
	}

	for _, word := range negative {
		if strings.Contains(query, word) {
			return "negative"
		}
	}

	return ""
}

// extractMovieTitles extracts quoted movie titles
func (ee *EntityExtractor) extractMovieTitles(query string) []string {
	matches := ee.patterns["movie_title"].FindAllStringSubmatch(query, -1)
	var titles []string

	for _, match := range matches {
		if len(match) > 1 && match[1] != "" {
			titles = append(titles, match[1])
		} else if len(match) > 2 && match[2] != "" {
			titles = append(titles, match[2])
		}
	}

	return titles
}

// extractPreferences extracts user preferences
func (ee *EntityExtractor) extractPreferences(query string) []string {
	preferences := []string{}

	if strings.Contains(query, "family") || strings.Contains(query, "kids") {
		preferences = append(preferences, "family-friendly")
	}
	if strings.Contains(query, "recent") || strings.Contains(query, "new") {
		preferences = append(preferences, "recent")
	}
	if strings.Contains(query, "classic") || strings.Contains(query, "old") {
		preferences = append(preferences, "classic")
	}
	if strings.Contains(query, "popular") || strings.Contains(query, "trending") {
		preferences = append(preferences, "popular")
	}

	return preferences
}

// extractMood extracts mood/context from query
func (ee *EntityExtractor) extractMood(query string) string {
	moods := map[string]string{
		"tonight":     "evening",
		"weekend":     "leisure",
		"date":        "romantic",
		"alone":       "solo",
		"friends":     "social",
		"relax":       "relaxing",
		"exciting":    "thrilling",
		"funny":       "humorous",
		"emotional":   "dramatic",
	}

	for keyword, mood := range moods {
		if strings.Contains(query, keyword) {
			return mood
		}
	}

	return ""
}