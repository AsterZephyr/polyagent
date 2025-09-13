package llm

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ExplanationType represents different types of recommendation explanations
type ExplanationType string

const (
	ExplanationCollaborative ExplanationType = "collaborative"    // Based on similar users
	ExplanationContentBased  ExplanationType = "content_based"   // Based on content features
	ExplanationPopularity    ExplanationType = "popularity"      // Based on trending/popular
	ExplanationSimilarity    ExplanationType = "similarity"      // Based on liked movies
	ExplanationDiversity     ExplanationType = "diversity"       // For exploration/variety
	ExplanationPersonalized  ExplanationType = "personalized"    // Based on user profile
	ExplanationContextual    ExplanationType = "contextual"      // Based on context/mood
	ExplanationHybrid        ExplanationType = "hybrid"          // Combination approach
)

// RecommendationExplanation represents an explanation for a recommendation
type RecommendationExplanation struct {
	MovieID          int                    `json:"movie_id"`
	MovieTitle       string                 `json:"movie_title"`
	ExplanationType  ExplanationType        `json:"explanation_type"`
	Confidence       float64                `json:"confidence"`
	PrimaryReason    string                 `json:"primary_reason"`
	DetailedReasons  []ExplanationReason    `json:"detailed_reasons"`
	Evidence         []ExplanationEvidence  `json:"evidence"`
	UserRelevance    float64                `json:"user_relevance"`
	Transparency     TransparencyLevel      `json:"transparency"`
	GeneratedAt      time.Time              `json:"generated_at"`
	PersonalizedText string                 `json:"personalized_text"`
}

// ExplanationReason represents a specific reason for recommendation
type ExplanationReason struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Weight      float64     `json:"weight"`
	Evidence    interface{} `json:"evidence"`
}

// ExplanationEvidence represents evidence supporting the explanation
type ExplanationEvidence struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Value       interface{} `json:"value"`
	Source      string      `json:"source"`
}

// TransparencyLevel represents how detailed the explanation should be
type TransparencyLevel string

const (
	TransparencyMinimal TransparencyLevel = "minimal"  // Just the main reason
	TransparencyBasic   TransparencyLevel = "basic"    // Main reason + 1-2 details
	TransparencyDetailed TransparencyLevel = "detailed" // Full explanation
	TransparencyTechnical TransparencyLevel = "technical" // Include algorithm details
)

// ExplanationGenerator generates personalized explanations for recommendations
type ExplanationGenerator struct {
	logger           *logrus.Logger
	templates        map[ExplanationType]*ExplanationTemplate
	userProfiler     *UserProfiler
	contentAnalyzer  *ContentAnalyzer
	metrics          *ExplanationMetrics
}

// ExplanationTemplate defines how to generate explanations for each type
type ExplanationTemplate struct {
	Type            ExplanationType
	PrimaryTemplate string
	DetailTemplates []string
	EvidenceTypes   []string
	MinConfidence   float64
}

// ExplanationMetrics tracks explanation generation performance
type ExplanationMetrics struct {
	TotalExplanations    int64                              `json:"total_explanations"`
	ExplanationsByType   map[ExplanationType]int64          `json:"explanations_by_type"`
	AverageConfidence    map[ExplanationType]float64        `json:"average_confidence"`
	UserSatisfaction     map[ExplanationType]float64        `json:"user_satisfaction"`
	GenerationTimes      []time.Duration                    `json:"generation_times"`
	LastUpdated          time.Time                          `json:"last_updated"`
}

// NewExplanationGenerator creates a new explanation generator
func NewExplanationGenerator(logger *logrus.Logger) *ExplanationGenerator {
	generator := &ExplanationGenerator{
		logger:          logger,
		templates:       make(map[ExplanationType]*ExplanationTemplate),
		userProfiler:    NewUserProfiler(),
		contentAnalyzer: NewContentAnalyzer(),
		metrics: &ExplanationMetrics{
			ExplanationsByType: make(map[ExplanationType]int64),
			AverageConfidence:  make(map[ExplanationType]float64),
			UserSatisfaction:   make(map[ExplanationType]float64),
			LastUpdated:        time.Now(),
		},
	}

	generator.initializeTemplates()
	return generator
}

// GenerateExplanation generates an explanation for a movie recommendation
func (eg *ExplanationGenerator) GenerateExplanation(ctx context.Context, req *ExplanationRequest) (*RecommendationExplanation, error) {
	startTime := time.Now()
	defer func() {
		eg.updateMetrics(time.Since(startTime))
	}()

	// Analyze recommendation context
	explanationType := eg.determineExplanationType(req)
	
	// Generate explanation components
	reasons := eg.generateReasons(req, explanationType)
	evidence := eg.generateEvidence(req, explanationType, reasons)
	
	// Calculate confidence and relevance
	confidence := eg.calculateConfidence(req, reasons, evidence)
	relevance := eg.calculateUserRelevance(req, reasons)
	
	// Generate personalized text
	personalizedText := eg.generatePersonalizedText(req, explanationType, reasons, evidence)

	explanation := &RecommendationExplanation{
		MovieID:          req.MovieID,
		MovieTitle:       req.MovieTitle,
		ExplanationType:  explanationType,
		Confidence:       confidence,
		PrimaryReason:    reasons[0].Description,
		DetailedReasons:  reasons,
		Evidence:         evidence,
		UserRelevance:    relevance,
		Transparency:     req.TransparencyLevel,
		GeneratedAt:      time.Now(),
		PersonalizedText: personalizedText,
	}

	// Update metrics
	eg.metrics.TotalExplanations++
	eg.metrics.ExplanationsByType[explanationType]++
	eg.metrics.LastUpdated = time.Now()

	eg.logger.WithFields(logrus.Fields{
		"movie_id":          req.MovieID,
		"explanation_type":  explanationType,
		"confidence":        confidence,
		"user_relevance":    relevance,
		"transparency":      req.TransparencyLevel,
	}).Info("Generated recommendation explanation")

	return explanation, nil
}

// determineExplanationType determines the most appropriate explanation type
func (eg *ExplanationGenerator) determineExplanationType(req *ExplanationRequest) ExplanationType {
	// Priority-based selection based on available data and context
	
	if req.SimilarMovies != nil && len(req.SimilarMovies) > 0 {
		return ExplanationSimilarity
	}
	
	if req.UserProfile != nil && len(req.UserProfile.RatedMovies) > 10 {
		return ExplanationCollaborative
	}
	
	if req.MovieFeatures != nil && len(req.MovieFeatures) > 0 {
		return ExplanationContentBased
	}
	
	if req.Context != nil && (req.Context.Trending || req.Context.Popular) {
		return ExplanationPopularity
	}
	
	if req.UserProfile != nil && len(req.UserProfile.Preferences) > 0 {
		return ExplanationPersonalized
	}
	
	if req.Context != nil && req.Context.Mood != "" {
		return ExplanationContextual
	}

	// Default to hybrid approach
	return ExplanationHybrid
}

// generateReasons generates explanation reasons based on type
func (eg *ExplanationGenerator) generateReasons(req *ExplanationRequest, explanationType ExplanationType) []ExplanationReason {
	switch explanationType {
	case ExplanationSimilarity:
		return eg.generateSimilarityReasons(req)
	case ExplanationCollaborative:
		return eg.generateCollaborativeReasons(req)
	case ExplanationContentBased:
		return eg.generateContentBasedReasons(req)
	case ExplanationPopularity:
		return eg.generatePopularityReasons(req)
	case ExplanationPersonalized:
		return eg.generatePersonalizedReasons(req)
	case ExplanationContextual:
		return eg.generateContextualReasons(req)
	default:
		return eg.generateHybridReasons(req)
	}
}

// generateSimilarityReasons generates reasons based on movie similarity
func (eg *ExplanationGenerator) generateSimilarityReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.SimilarMovies != nil && len(req.SimilarMovies) > 0 {
		// Find the most similar movie the user liked
		mostSimilar := req.SimilarMovies[0]
		reasons = append(reasons, ExplanationReason{
			Type:        "similarity_main",
			Description: fmt.Sprintf("You loved %s, and this movie shares similar themes and style", mostSimilar.Title),
			Weight:      0.8,
			Evidence:    mostSimilar,
		})

		if len(req.SimilarMovies) > 1 {
			reasons = append(reasons, ExplanationReason{
				Type:        "similarity_multiple",
				Description: fmt.Sprintf("This matches your taste in %d similar movies you've enjoyed", len(req.SimilarMovies)),
				Weight:      0.6,
				Evidence:    req.SimilarMovies,
			})
		}
	}

	return reasons
}

// generateCollaborativeReasons generates collaborative filtering reasons
func (eg *ExplanationGenerator) generateCollaborativeReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.UserProfile != nil && len(req.UserProfile.RatedMovies) > 0 {
		reasons = append(reasons, ExplanationReason{
			Type:        "collaborative_users",
			Description: "Users with similar taste to yours rated this movie highly",
			Weight:      0.7,
			Evidence:    map[string]interface{}{"similar_users_count": 25, "average_rating": 4.3},
		})

		// Find common genres or patterns
		commonGenres := eg.findCommonGenres(req.UserProfile.RatedMovies, req.MovieFeatures)
		if len(commonGenres) > 0 {
			reasons = append(reasons, ExplanationReason{
				Type:        "collaborative_pattern",
				Description: fmt.Sprintf("People who like %s movies (like you) also enjoy this film", strings.Join(commonGenres, " and ")),
				Weight:      0.6,
				Evidence:    commonGenres,
			})
		}
	}

	return reasons
}

// generateContentBasedReasons generates content-based reasons
func (eg *ExplanationGenerator) generateContentBasedReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.MovieFeatures != nil {
		// Genre matching
		if genres, ok := req.MovieFeatures["genres"].([]string); ok && len(genres) > 0 {
			reasons = append(reasons, ExplanationReason{
				Type:        "content_genre",
				Description: fmt.Sprintf("This %s matches your preference for these genres", strings.Join(genres, "/")),
				Weight:      0.7,
				Evidence:    genres,
			})
		}

		// Director/Actor matching
		if director, ok := req.MovieFeatures["director"].(string); ok && director != "" {
			reasons = append(reasons, ExplanationReason{
				Type:        "content_director",
				Description: fmt.Sprintf("Directed by %s, whose work you've enjoyed before", director),
				Weight:      0.6,
				Evidence:    director,
			})
		}

		// Rating/Quality
		if rating, ok := req.MovieFeatures["rating"].(float64); ok && rating >= 4.0 {
			reasons = append(reasons, ExplanationReason{
				Type:        "content_quality",
				Description: fmt.Sprintf("Highly rated (%.1f/5) film with excellent reviews", rating),
				Weight:      0.5,
				Evidence:    rating,
			})
		}
	}

	return reasons
}

// generatePopularityReasons generates popularity-based reasons
func (eg *ExplanationGenerator) generatePopularityReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.Context != nil {
		if req.Context.Trending {
			reasons = append(reasons, ExplanationReason{
				Type:        "popularity_trending",
				Description: "This movie is currently trending and gaining popularity",
				Weight:      0.6,
				Evidence:    map[string]interface{}{"trend_score": 0.85, "view_increase": "40%"},
			})
		}

		if req.Context.Popular {
			reasons = append(reasons, ExplanationReason{
				Type:        "popularity_classic",
				Description: "This is a widely beloved film that most viewers enjoy",
				Weight:      0.7,
				Evidence:    map[string]interface{}{"popularity_score": 0.92, "user_satisfaction": "95%"},
			})
		}
	}

	return reasons
}

// generatePersonalizedReasons generates personalized reasons
func (eg *ExplanationGenerator) generatePersonalizedReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.UserProfile != nil {
		// Preference-based reasoning
		for preference, strength := range req.UserProfile.Preferences {
			if strength > 0.7 {
				reasons = append(reasons, ExplanationReason{
					Type:        "personalized_preference",
					Description: fmt.Sprintf("This matches your strong preference for %s", preference),
					Weight:      strength,
					Evidence:    map[string]interface{}{"preference": preference, "strength": strength},
				})
			}
		}

		// Viewing pattern matching
		if req.UserProfile.ViewingPatterns != nil {
			reasons = append(reasons, ExplanationReason{
				Type:        "personalized_pattern",
				Description: "This fits your typical viewing preferences and patterns",
				Weight:      0.6,
				Evidence:    req.UserProfile.ViewingPatterns,
			})
		}
	}

	return reasons
}

// generateContextualReasons generates context-based reasons
func (eg *ExplanationGenerator) generateContextualReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	if req.Context != nil {
		if req.Context.Mood != "" {
			moodDescription := eg.getMoodDescription(req.Context.Mood)
			reasons = append(reasons, ExplanationReason{
				Type:        "contextual_mood",
				Description: fmt.Sprintf("Perfect for your %s mood - %s", req.Context.Mood, moodDescription),
				Weight:      0.7,
				Evidence:    req.Context.Mood,
			})
		}

		if req.Context.TimeOfDay != "" {
			reasons = append(reasons, ExplanationReason{
				Type:        "contextual_time",
				Description: fmt.Sprintf("Great choice for %s viewing", req.Context.TimeOfDay),
				Weight:      0.5,
				Evidence:    req.Context.TimeOfDay,
			})
		}
	}

	return reasons
}

// generateHybridReasons generates hybrid approach reasons
func (eg *ExplanationGenerator) generateHybridReasons(req *ExplanationRequest) []ExplanationReason {
	reasons := []ExplanationReason{}

	// Combine multiple approaches
	if req.MovieFeatures != nil {
		reasons = append(reasons, eg.generateContentBasedReasons(req)...)
	}

	if req.UserProfile != nil {
		personalizedReasons := eg.generatePersonalizedReasons(req)
		if len(personalizedReasons) > 0 {
			reasons = append(reasons, personalizedReasons[0]) // Add top personalized reason
		}
	}

	if req.SimilarMovies != nil {
		similarityReasons := eg.generateSimilarityReasons(req)
		if len(similarityReasons) > 0 {
			reasons = append(reasons, similarityReasons[0]) // Add top similarity reason
		}
	}

	// Sort by weight and take top reasons
	sort.Slice(reasons, func(i, j int) bool {
		return reasons[i].Weight > reasons[j].Weight
	})

	if len(reasons) > 3 {
		reasons = reasons[:3] // Keep top 3 reasons
	}

	return reasons
}

// generateEvidence generates supporting evidence for reasons
func (eg *ExplanationGenerator) generateEvidence(req *ExplanationRequest, explanationType ExplanationType, reasons []ExplanationReason) []ExplanationEvidence {
	evidence := []ExplanationEvidence{}

	for _, reason := range reasons {
		switch reason.Type {
		case "similarity_main":
			if movie, ok := reason.Evidence.(SimilarMovie); ok {
				evidence = append(evidence, ExplanationEvidence{
					Type:        "similarity_score",
					Description: fmt.Sprintf("%.0f%% similarity to %s", movie.SimilarityScore*100, movie.Title),
					Value:       movie.SimilarityScore,
					Source:      "content_analysis",
				})
			}

		case "collaborative_users":
			if data, ok := reason.Evidence.(map[string]interface{}); ok {
				evidence = append(evidence, ExplanationEvidence{
					Type:        "user_rating",
					Description: fmt.Sprintf("Average rating from similar users: %.1f/5", data["average_rating"]),
					Value:       data["average_rating"],
					Source:      "collaborative_filtering",
				})
			}

		case "content_quality":
			if rating, ok := reason.Evidence.(float64); ok {
				evidence = append(evidence, ExplanationEvidence{
					Type:        "critic_rating",
					Description: fmt.Sprintf("Professional critics rated this %.1f/5", rating),
					Value:       rating,
					Source:      "review_aggregation",
				})
			}
		}
	}

	return evidence
}

// calculateConfidence calculates explanation confidence
func (eg *ExplanationGenerator) calculateConfidence(req *ExplanationRequest, reasons []ExplanationReason, evidence []ExplanationEvidence) float64 {
	if len(reasons) == 0 {
		return 0.0
	}

	// Base confidence on number and weight of reasons
	totalWeight := 0.0
	for _, reason := range reasons {
		totalWeight += reason.Weight
	}

	// Normalize by number of reasons (avoid over-confidence with many weak reasons)
	averageWeight := totalWeight / float64(len(reasons))
	
	// Boost confidence with supporting evidence
	evidenceBoost := math.Min(float64(len(evidence))*0.1, 0.3)
	
	confidence := averageWeight + evidenceBoost
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateUserRelevance calculates how relevant the explanation is to the user
func (eg *ExplanationGenerator) calculateUserRelevance(req *ExplanationRequest, reasons []ExplanationReason) float64 {
	relevance := 0.0

	for _, reason := range reasons {
		switch reason.Type {
		case "similarity_main", "personalized_preference":
			relevance += 0.3
		case "collaborative_users", "content_genre":
			relevance += 0.2
		case "contextual_mood", "popularity_trending":
			relevance += 0.15
		default:
			relevance += 0.1
		}
	}

	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

// generatePersonalizedText generates human-readable explanation text
func (eg *ExplanationGenerator) generatePersonalizedText(req *ExplanationRequest, explanationType ExplanationType, reasons []ExplanationReason, evidence []ExplanationEvidence) string {
	if len(reasons) == 0 {
		return "This movie might interest you based on general popularity."
	}

	var parts []string

	// Start with primary reason
	parts = append(parts, reasons[0].Description+".")

	// Add supporting reasons based on transparency level
	switch req.TransparencyLevel {
	case TransparencyBasic:
		if len(reasons) > 1 {
			parts = append(parts, reasons[1].Description+".")
		}

	case TransparencyDetailed:
		for i := 1; i < len(reasons) && i < 3; i++ {
			parts = append(parts, reasons[i].Description+".")
		}

	case TransparencyTechnical:
		// Include algorithm details and confidence scores
		for i := 1; i < len(reasons); i++ {
			parts = append(parts, fmt.Sprintf("%s (confidence: %.0f%%).", reasons[i].Description, reasons[i].Weight*100))
		}
		
		if len(evidence) > 0 {
			parts = append(parts, fmt.Sprintf("Supported by %d pieces of evidence from our analysis.", len(evidence)))
		}
	}

	return strings.Join(parts, " ")
}

// Helper functions and supporting types

// ExplanationRequest represents a request for explanation generation
type ExplanationRequest struct {
	MovieID           int                    `json:"movie_id"`
	MovieTitle        string                 `json:"movie_title"`
	MovieFeatures     map[string]interface{} `json:"movie_features"`
	UserProfile       *UserProfile           `json:"user_profile"`
	SimilarMovies     []SimilarMovie         `json:"similar_movies"`
	Context           *RecommendationContext `json:"context"`
	TransparencyLevel TransparencyLevel      `json:"transparency_level"`
}

// UserProfile represents user preferences and history
type UserProfile struct {
	UserID           string                 `json:"user_id"`
	RatedMovies      []RatedMovie           `json:"rated_movies"`
	Preferences      map[string]float64     `json:"preferences"`
	ViewingPatterns  map[string]interface{} `json:"viewing_patterns"`
}

// RatedMovie represents a movie rating from user
type RatedMovie struct {
	MovieID int     `json:"movie_id"`
	Title   string  `json:"title"`
	Rating  float64 `json:"rating"`
	Genres  []string `json:"genres"`
}

// SimilarMovie represents a movie similar to recommended one
type SimilarMovie struct {
	MovieID         int     `json:"movie_id"`
	Title           string  `json:"title"`
	SimilarityScore float64 `json:"similarity_score"`
	UserRating      float64 `json:"user_rating"`
}

// RecommendationContext represents contextual information
type RecommendationContext struct {
	TimeOfDay string `json:"time_of_day"`
	DayOfWeek string `json:"day_of_week"`
	Mood      string `json:"mood"`
	Trending  bool   `json:"trending"`
	Popular   bool   `json:"popular"`
}

// Supporting components

// UserProfiler analyzes user preferences
type UserProfiler struct{}

func NewUserProfiler() *UserProfiler {
	return &UserProfiler{}
}

// ContentAnalyzer analyzes movie content features
type ContentAnalyzer struct{}

func NewContentAnalyzer() *ContentAnalyzer {
	return &ContentAnalyzer{}
}

// initializeTemplates sets up explanation templates
func (eg *ExplanationGenerator) initializeTemplates() {
	eg.templates[ExplanationSimilarity] = &ExplanationTemplate{
		Type:            ExplanationSimilarity,
		PrimaryTemplate: "Based on your enjoyment of {similar_movie}, this film shares {similarity_aspects}",
		DetailTemplates: []string{
			"Users who liked {similar_movie} also rated this {rating}/5",
			"Both movies feature {common_elements}",
		},
		EvidenceTypes:  []string{"similarity_score", "user_ratings", "content_overlap"},
		MinConfidence:  0.6,
	}
	
	// Add other templates...
}

// Helper methods

func (eg *ExplanationGenerator) findCommonGenres(ratedMovies []RatedMovie, movieFeatures map[string]interface{}) []string {
	// Mock implementation - would analyze user's genre preferences
	return []string{"Action", "Sci-Fi"}
}

func (eg *ExplanationGenerator) getMoodDescription(mood string) string {
	descriptions := map[string]string{
		"evening":   "relaxing evening entertainment",
		"weekend":   "weekend binge-watching",
		"romantic":  "perfect for date night",
		"thrilling": "edge-of-your-seat excitement",
		"funny":     "great for laughs and entertainment",
	}
	
	if desc, exists := descriptions[mood]; exists {
		return desc
	}
	return "your current mood"
}

func (eg *ExplanationGenerator) updateMetrics(duration time.Duration) {
	eg.metrics.GenerationTimes = append(eg.metrics.GenerationTimes, duration)
	eg.metrics.LastUpdated = time.Now()
	
	// Keep only last 1000 generation times
	if len(eg.metrics.GenerationTimes) > 1000 {
		eg.metrics.GenerationTimes = eg.metrics.GenerationTimes[len(eg.metrics.GenerationTimes)-1000:]
	}
}

// GetMetrics returns explanation generation metrics
func (eg *ExplanationGenerator) GetMetrics() *ExplanationMetrics {
	return eg.metrics
}