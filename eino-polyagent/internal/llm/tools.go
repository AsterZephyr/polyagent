package llm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ToolRegistry manages recommendation system-specific tools
type ToolRegistry struct {
	tools map[string]RecommendationTool
}

// RecommendationTool interface for recommendation-specific tools
type RecommendationTool interface {
	Execute(ctx context.Context, params map[string]interface{}) (interface{}, error)
	GetDefinition() Tool
	GetName() string
	GetDescription() string
}

// NewToolRegistry creates a new tool registry with recommendation tools
func NewToolRegistry() *ToolRegistry {
	registry := &ToolRegistry{
		tools: make(map[string]RecommendationTool),
	}

	// Register recommendation system tools
	registry.RegisterTool(&MovieSearchTool{})
	registry.RegisterTool(&UserPreferenceTool{})
	registry.RegisterTool(&ContentFilterTool{})
	registry.RegisterTool(&RecommendationGeneratorTool{})
	registry.RegisterTool(&UserInteractionTool{})
	registry.RegisterTool(&PopularityAnalyzerTool{})

	return registry
}

// RegisterTool registers a tool in the registry
func (r *ToolRegistry) RegisterTool(tool RecommendationTool) {
	r.tools[tool.GetName()] = tool
}

// GetTool retrieves a tool by name
func (r *ToolRegistry) GetTool(name string) (RecommendationTool, bool) {
	tool, exists := r.tools[name]
	return tool, exists
}

// GetAllTools returns all available tool definitions
func (r *ToolRegistry) GetAllTools() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool.GetDefinition())
	}
	return tools
}

// ExecuteTool executes a tool with given parameters
func (r *ToolRegistry) ExecuteTool(ctx context.Context, name string, params map[string]interface{}) (interface{}, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Execute(ctx, params)
}

// MovieSearchTool - Search for movies based on criteria
type MovieSearchTool struct{}

func (t *MovieSearchTool) GetName() string {
	return "search_movies"
}

func (t *MovieSearchTool) GetDescription() string {
	return "Search for movies based on genre, year range, rating criteria, and keywords"
}

func (t *MovieSearchTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"genre": map[string]interface{}{
						"type":        "string",
						"description": "Movie genre (Action, Comedy, Drama, Sci-Fi, etc.)",
					},
					"year_range": map[string]interface{}{
						"type":        "array",
						"description": "Array with start and end year [start_year, end_year]",
						"items": map[string]interface{}{
							"type": "integer",
						},
					},
					"min_rating": map[string]interface{}{
						"type":        "number",
						"description": "Minimum rating score (1-5)",
					},
					"keywords": map[string]interface{}{
						"type":        "string",
						"description": "Keywords to search in movie titles and descriptions",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return (default: 10)",
					},
				},
				"required": []string{},
			},
		},
	}
}

func (t *MovieSearchTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract parameters with defaults
	genre, _ := params["genre"].(string)
	keywords, _ := params["keywords"].(string)
	minRating := float64(0)
	if rating, ok := params["min_rating"].(float64); ok {
		minRating = rating
	}
	
	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	var yearRange []int
	if yr, ok := params["year_range"].([]interface{}); ok && len(yr) == 2 {
		if start, ok := yr[0].(float64); ok {
			if end, ok := yr[1].(float64); ok {
				yearRange = []int{int(start), int(end)}
			}
		}
	}

	// Mock movie search results - in production this would query the database
	movies := []map[string]interface{}{
		{
			"id":          1,
			"title":       "The Matrix",
			"year":        1999,
			"genre":       []string{"Action", "Sci-Fi"},
			"rating":      4.5,
			"description": "A computer hacker learns about the true nature of reality",
		},
		{
			"id":          2,
			"title":       "Inception",
			"year":        2010,
			"genre":       []string{"Action", "Sci-Fi", "Thriller"},
			"rating":      4.7,
			"description": "A thief enters people's dreams to steal their secrets",
		},
		{
			"id":          3,
			"title":       "The Godfather",
			"year":        1972,
			"genre":       []string{"Crime", "Drama"},
			"rating":      4.9,
			"description": "The patriarch of an organized crime dynasty",
		},
	}

	// Apply filters
	var filteredMovies []map[string]interface{}
	for _, movie := range movies {
		// Genre filter
		if genre != "" {
			genres, _ := movie["genre"].([]string)
			genreMatch := false
			for _, g := range genres {
				if strings.EqualFold(g, genre) {
					genreMatch = true
					break
				}
			}
			if !genreMatch {
				continue
			}
		}

		// Rating filter
		if movieRating, ok := movie["rating"].(float64); ok && movieRating < minRating {
			continue
		}

		// Year range filter
		if len(yearRange) == 2 {
			if movieYear, ok := movie["year"].(int); ok {
				if movieYear < yearRange[0] || movieYear > yearRange[1] {
					continue
				}
			}
		}

		// Keywords filter
		if keywords != "" {
			title, _ := movie["title"].(string)
			description, _ := movie["description"].(string)
			keywordsLower := strings.ToLower(keywords)
			if !strings.Contains(strings.ToLower(title), keywordsLower) &&
				!strings.Contains(strings.ToLower(description), keywordsLower) {
				continue
			}
		}

		filteredMovies = append(filteredMovies, movie)
	}

	// Apply limit
	if len(filteredMovies) > limit {
		filteredMovies = filteredMovies[:limit]
	}

	return map[string]interface{}{
		"movies": filteredMovies,
		"count":  len(filteredMovies),
		"query": map[string]interface{}{
			"genre":      genre,
			"year_range": yearRange,
			"min_rating": minRating,
			"keywords":   keywords,
		},
	}, nil
}

// UserPreferenceTool - Analyze user preferences
type UserPreferenceTool struct{}

func (t *UserPreferenceTool) GetName() string {
	return "analyze_user_preferences"
}

func (t *UserPreferenceTool) GetDescription() string {
	return "Analyze user preferences based on viewing history and ratings"
}

func (t *UserPreferenceTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "User ID to analyze preferences for",
					},
					"include_implicit": map[string]interface{}{
						"type":        "boolean",
						"description": "Include implicit feedback (views, time spent) in analysis",
					},
				},
				"required": []string{"user_id"},
			},
		},
	}
}

func (t *UserPreferenceTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, _ := params["user_id"].(string)
	includeImplicit, _ := params["include_implicit"].(bool)

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Mock user preference analysis
	preferences := map[string]interface{}{
		"user_id": userID,
		"genres": map[string]interface{}{
			"Action":  0.8,
			"Sci-Fi":  0.9,
			"Comedy":  0.3,
			"Drama":   0.6,
			"Horror":  0.1,
		},
		"preferred_decades": []string{"1990s", "2000s", "2010s"},
		"rating_pattern": map[string]interface{}{
			"average_rating":     4.2,
			"rating_variance":    0.7,
			"harsh_critic":       false,
			"generous_reviewer":  true,
		},
		"viewing_behavior": map[string]interface{}{
			"active_hours":       []string{"19:00-23:00"},
			"weekend_preference": true,
			"binge_watcher":      true,
		},
		"implicit_signals": map[string]interface{}{
			"completion_rate":    0.85,
			"rewatch_frequency":  0.2,
			"social_shares":      15,
		},
	}

	if !includeImplicit {
		delete(preferences, "implicit_signals")
	}

	return preferences, nil
}

// ContentFilterTool - Filter content based on user constraints
type ContentFilterTool struct{}

func (t *ContentFilterTool) GetName() string {
	return "filter_content"
}

func (t *ContentFilterTool) GetDescription() string {
	return "Filter movie content based on age ratings, content warnings, and user constraints"
}

func (t *ContentFilterTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"age_rating": map[string]interface{}{
						"type":        "string",
						"description": "Maximum age rating (G, PG, PG-13, R, NC-17)",
					},
					"content_warnings": map[string]interface{}{
						"type":        "array",
						"description": "Content types to avoid (violence, language, sexual_content, substance_abuse)",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"family_friendly": map[string]interface{}{
						"type":        "boolean",
						"description": "Only show family-friendly content",
					},
					"user_blocklist": map[string]interface{}{
						"type":        "array",
						"description": "List of movie IDs or genres to block",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		},
	}
}

func (t *ContentFilterTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	ageRating, _ := params["age_rating"].(string)
	familyFriendly, _ := params["family_friendly"].(bool)

	var contentWarnings []string
	if cw, ok := params["content_warnings"].([]interface{}); ok {
		for _, warning := range cw {
			if w, ok := warning.(string); ok {
				contentWarnings = append(contentWarnings, w)
			}
		}
	}

	var userBlocklist []string
	if ub, ok := params["user_blocklist"].([]interface{}); ok {
		for _, item := range ub {
			if b, ok := item.(string); ok {
				userBlocklist = append(userBlocklist, b)
			}
		}
	}

	// Mock content filtering rules
	filterRules := map[string]interface{}{
		"age_rating_filter": map[string]interface{}{
			"enabled":     ageRating != "",
			"max_rating":  ageRating,
		},
		"content_filters": contentWarnings,
		"family_mode":     familyFriendly,
		"blocked_items":   userBlocklist,
		"additional_rules": map[string]interface{}{
			"exclude_adult_content":    familyFriendly,
			"filter_graphic_violence":  contains(contentWarnings, "violence"),
			"filter_strong_language":   contains(contentWarnings, "language"),
			"filter_sexual_content":    contains(contentWarnings, "sexual_content"),
			"filter_substance_abuse":   contains(contentWarnings, "substance_abuse"),
		},
	}

	return map[string]interface{}{
		"filter_rules": filterRules,
		"applied_at":   time.Now().UTC(),
		"rule_count":   len(contentWarnings) + len(userBlocklist) + 1,
	}, nil
}

// RecommendationGeneratorTool - Generate personalized recommendations
type RecommendationGeneratorTool struct{}

func (t *RecommendationGeneratorTool) GetName() string {
	return "generate_recommendations"
}

func (t *RecommendationGeneratorTool) GetDescription() string {
	return "Generate personalized movie recommendations using collaborative filtering and content-based algorithms"
}

func (t *RecommendationGeneratorTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "User ID to generate recommendations for",
					},
					"algorithm": map[string]interface{}{
						"type":        "string",
						"description": "Recommendation algorithm (collaborative_filtering, content_based, hybrid)",
						"enum":        []string{"collaborative_filtering", "content_based", "hybrid"},
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Number of recommendations to generate (default: 5)",
					},
					"diversity_factor": map[string]interface{}{
						"type":        "number",
						"description": "Diversity factor to avoid similar recommendations (0.0-1.0)",
					},
				},
				"required": []string{"user_id"},
			},
		},
	}
}

func (t *RecommendationGeneratorTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, _ := params["user_id"].(string)
	algorithm, _ := params["algorithm"].(string)
	
	count := 5
	if c, ok := params["count"].(float64); ok {
		count = int(c)
	}

	diversityFactor := 0.5
	if df, ok := params["diversity_factor"].(float64); ok {
		diversityFactor = df
	}

	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if algorithm == "" {
		algorithm = "hybrid"
	}

	// Mock recommendation generation
	recommendations := []map[string]interface{}{
		{
			"movie_id":       101,
			"title":          "Blade Runner 2049",
			"confidence":     0.92,
			"reason":         "Based on your love for sci-fi and high ratings for similar movies",
			"similarity_score": 0.87,
			"predicted_rating": 4.6,
		},
		{
			"movie_id":       102,
			"title":          "Mad Max: Fury Road",
			"confidence":     0.88,
			"reason":         "Action-packed film similar to movies you've rated highly",
			"similarity_score": 0.84,
			"predicted_rating": 4.4,
		},
		{
			"movie_id":       103,
			"title":          "Ex Machina",
			"confidence":     0.85,
			"reason":         "Intelligent sci-fi thriller matching your preferences",
			"similarity_score": 0.81,
			"predicted_rating": 4.5,
		},
	}

	// Apply count limit
	if len(recommendations) > count {
		recommendations = recommendations[:count]
	}

	return map[string]interface{}{
		"user_id":           userID,
		"algorithm":         algorithm,
		"recommendations":   recommendations,
		"diversity_factor":  diversityFactor,
		"generated_at":      time.Now().UTC(),
		"total_count":       len(recommendations),
	}, nil
}

// UserInteractionTool - Track and analyze user interactions
type UserInteractionTool struct{}

func (t *UserInteractionTool) GetName() string {
	return "track_user_interaction"
}

func (t *UserInteractionTool) GetDescription() string {
	return "Track user interactions with recommendations and content"
}

func (t *UserInteractionTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"user_id": map[string]interface{}{
						"type":        "string",
						"description": "User ID",
					},
					"interaction_type": map[string]interface{}{
						"type":        "string",
						"description": "Type of interaction (view, rate, like, share, watch, skip)",
						"enum":        []string{"view", "rate", "like", "share", "watch", "skip"},
					},
					"movie_id": map[string]interface{}{
						"type":        "string",
						"description": "Movie ID that was interacted with",
					},
					"rating": map[string]interface{}{
						"type":        "number",
						"description": "Rating given by user (1-5, optional)",
					},
					"duration_watched": map[string]interface{}{
						"type":        "number",
						"description": "Duration watched in minutes (for watch interactions)",
					},
				},
				"required": []string{"user_id", "interaction_type", "movie_id"},
			},
		},
	}
}

func (t *UserInteractionTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	userID, _ := params["user_id"].(string)
	interactionType, _ := params["interaction_type"].(string)
	movieID, _ := params["movie_id"].(string)
	
	var rating *float64
	if r, ok := params["rating"].(float64); ok {
		rating = &r
	}

	var durationWatched *float64
	if d, ok := params["duration_watched"].(float64); ok {
		durationWatched = &d
	}

	if userID == "" || interactionType == "" || movieID == "" {
		return nil, fmt.Errorf("user_id, interaction_type, and movie_id are required")
	}

	// Mock interaction tracking
	interaction := map[string]interface{}{
		"user_id":          userID,
		"movie_id":         movieID,
		"interaction_type": interactionType,
		"timestamp":        time.Now().UTC(),
		"session_id":       fmt.Sprintf("session_%d", time.Now().Unix()),
	}

	if rating != nil {
		interaction["rating"] = *rating
	}

	if durationWatched != nil {
		interaction["duration_watched"] = *durationWatched
	}

	// Additional analytics
	analytics := map[string]interface{}{
		"engagement_score": calculateEngagementScore(interactionType, rating, durationWatched),
		"feedback_type":    getFeedbackType(interactionType, rating),
		"recommendation_success": interactionType == "watch" || interactionType == "like" || (rating != nil && *rating >= 4.0),
	}

	return map[string]interface{}{
		"interaction": interaction,
		"analytics":   analytics,
		"recorded_at": time.Now().UTC(),
	}, nil
}

// PopularityAnalyzerTool - Analyze content popularity and trends
type PopularityAnalyzerTool struct{}

func (t *PopularityAnalyzerTool) GetName() string {
	return "analyze_popularity"
}

func (t *PopularityAnalyzerTool) GetDescription() string {
	return "Analyze movie popularity trends and patterns"
}

func (t *PopularityAnalyzerTool) GetDefinition() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        t.GetName(),
			Description: t.GetDescription(),
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"time_window": map[string]interface{}{
						"type":        "string",
						"description": "Time window for analysis (day, week, month, year)",
						"enum":        []string{"day", "week", "month", "year"},
					},
					"genre_filter": map[string]interface{}{
						"type":        "string",
						"description": "Filter by specific genre (optional)",
					},
					"metric": map[string]interface{}{
						"type":        "string",
						"description": "Popularity metric (views, ratings, shares, engagement)",
						"enum":        []string{"views", "ratings", "shares", "engagement"},
					},
				},
			},
		},
	}
}

func (t *PopularityAnalyzerTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	timeWindow, _ := params["time_window"].(string)
	genreFilter, _ := params["genre_filter"].(string)
	metric, _ := params["metric"].(string)

	if timeWindow == "" {
		timeWindow = "week"
	}
	if metric == "" {
		metric = "engagement"
	}

	// Mock popularity analysis
	popularMovies := []map[string]interface{}{
		{
			"movie_id":       1,
			"title":          "Top Gun: Maverick",
			"popularity_score": 0.95,
			"trend":          "rising",
			"view_count":     15420,
			"rating_count":   3240,
			"avg_rating":     4.6,
		},
		{
			"movie_id":       2,
			"title":          "Everything Everywhere All at Once",
			"popularity_score": 0.92,
			"trend":          "stable",
			"view_count":     12340,
			"rating_count":   2890,
			"avg_rating":     4.7,
		},
	}

	trends := map[string]interface{}{
		"genre_trends": map[string]float64{
			"Action":  0.8,
			"Comedy":  0.6,
			"Drama":   0.7,
			"Sci-Fi":  0.85,
			"Horror":  0.4,
		},
		"time_patterns": map[string]interface{}{
			"peak_hours":    []string{"19:00-22:00"},
			"peak_days":     []string{"Friday", "Saturday", "Sunday"},
			"seasonal_trend": "summer_blockbusters",
		},
	}

	return map[string]interface{}{
		"time_window":     timeWindow,
		"metric":          metric,
		"genre_filter":    genreFilter,
		"popular_movies":  popularMovies,
		"trends":          trends,
		"analyzed_at":     time.Now().UTC(),
	}, nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func calculateEngagementScore(interactionType string, rating *float64, durationWatched *float64) float64 {
	baseScore := map[string]float64{
		"view":  0.1,
		"rate":  0.5,
		"like":  0.7,
		"share": 0.8,
		"watch": 0.9,
		"skip":  -0.2,
	}

	score := baseScore[interactionType]

	if rating != nil {
		score += (*rating - 3.0) * 0.2 // Boost for high ratings
	}

	if durationWatched != nil && *durationWatched > 60 {
		score += 0.3 // Boost for longer viewing time
	}

	if score > 1.0 {
		score = 1.0
	}
	if score < -1.0 {
		score = -1.0
	}

	return score
}

func getFeedbackType(interactionType string, rating *float64) string {
	switch interactionType {
	case "rate":
		if rating != nil {
			if *rating >= 4.0 {
				return "positive_explicit"
			} else if *rating <= 2.0 {
				return "negative_explicit"
			}
			return "neutral_explicit"
		}
		return "explicit"
	case "like", "share", "watch":
		return "positive_implicit"
	case "skip":
		return "negative_implicit"
	default:
		return "neutral_implicit"
	}
}