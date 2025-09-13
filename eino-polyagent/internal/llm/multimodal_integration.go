package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// MultimodalRecommendationEngine combines multimodal analysis with recommendations
type MultimodalRecommendationEngine struct {
	*ExplainableRecommendationEngine
	multimodalAnalyzer *MultimodalAnalyzer
	contentDB          *MultimodalContentDB
	logger             *logrus.Logger
}

// MultimodalContentDB stores and manages multimodal content
type MultimodalContentDB struct {
	content map[int]*MultimodalContent // movieID -> content
	logger  *logrus.Logger
}

// NewMultimodalRecommendationEngine creates a new multimodal recommendation engine
func NewMultimodalRecommendationEngine(config *LLMAdapterConfig, logger *logrus.Logger) (*MultimodalRecommendationEngine, error) {
	explainableEngine, err := NewExplainableRecommendationEngine(config, logger)
	if err != nil {
		return nil, err
	}

	multimodalAnalyzer := NewMultimodalAnalyzer(explainableEngine.llmAdapter, logger)

	return &MultimodalRecommendationEngine{
		ExplainableRecommendationEngine: explainableEngine,
		multimodalAnalyzer:              multimodalAnalyzer,
		contentDB:                       NewMultimodalContentDB(logger),
		logger:                          logger,
	}, nil
}

// NewMultimodalContentDB creates a new multimodal content database
func NewMultimodalContentDB(logger *logrus.Logger) *MultimodalContentDB {
	return &MultimodalContentDB{
		content: make(map[int]*MultimodalContent),
		logger:  logger,
	}
}

// ProcessMultimodalRecommendation processes recommendations with multimodal content analysis
func (mre *MultimodalRecommendationEngine) ProcessMultimodalRecommendation(ctx context.Context, req *MultimodalRecommendationRequest) (*MultimodalRecommendationResult, error) {
	startTime := time.Now()

	// First, get regular explainable recommendations
	explainableReq := &ExplainableRecommendationRequest{
		UserID:            req.UserID,
		SessionID:         req.SessionID,
		Query:             req.Query,
		UserProfile:       req.UserProfile,
		SimilarMovies:     req.SimilarMovies,
		TransparencyLevel: req.TransparencyLevel,
	}

	explainableResult, err := mre.ProcessExplainableRecommendation(ctx, explainableReq)
	if err != nil {
		mre.logger.WithError(err).Error("Failed to process explainable recommendation")
		return nil, fmt.Errorf("explainable recommendation failed: %w", err)
	}

	// Enhance recommendations with multimodal analysis
	enhancedMovies := []EnhancedRecommendedMovie{}
	
	for _, movie := range explainableResult.RecommendedMovies {
		enhanced, err := mre.enhanceMovieWithMultimodal(ctx, movie, req)
		if err != nil {
			mre.logger.WithError(err).WithField("movie_id", movie.ID).Warn("Failed to enhance movie with multimodal analysis")
			// Continue with basic movie data
			enhanced = &EnhancedRecommendedMovie{
				RecommendedMovie: movie,
				MultimodalScore:  0.5, // Default score
			}
		}
		enhancedMovies = append(enhancedMovies, *enhanced)
	}

	// Generate multimodal-aware explanations
	enhancedExplanations := mre.enhanceExplanationsWithMultimodal(explainableResult.Explanations, enhancedMovies)

	// Create multimodal recommendation result
	result := &MultimodalRecommendationResult{
		ExplainableRecommendationResult: *explainableResult,
		EnhancedMovies:                  enhancedMovies,
		EnhancedExplanations:            enhancedExplanations,
		MultimodalProcessingTime:        time.Since(startTime),
		ContentAnalysisCount:            len(enhancedMovies),
		AverageMultimodalScore:          mre.calculateAverageMultimodalScore(enhancedMovies),
	}

	mre.logger.WithFields(logrus.Fields{
		"user_id":                    req.UserID,
		"movies_enhanced":            len(enhancedMovies),
		"multimodal_processing_time": result.MultimodalProcessingTime,
		"average_multimodal_score":   result.AverageMultimodalScore,
	}).Info("Processed multimodal recommendation successfully")

	return result, nil
}

// enhanceMovieWithMultimodal enhances a movie recommendation with multimodal analysis
func (mre *MultimodalRecommendationEngine) enhanceMovieWithMultimodal(ctx context.Context, movie RecommendedMovie, req *MultimodalRecommendationRequest) (*EnhancedRecommendedMovie, error) {
	// Try to get existing multimodal content
	content, exists := mre.contentDB.GetContent(movie.ID)
	if !exists {
		// Create and analyze new content
		var err error
		content, err = mre.createMultimodalContent(ctx, movie, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create multimodal content: %w", err)
		}
		
		// Store for future use
		mre.contentDB.StoreContent(content)
	}

	// Perform multimodal analysis if not already done
	if content.Analysis == nil {
		analysis, err := mre.multimodalAnalyzer.AnalyzeMultimodalContent(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze multimodal content: %w", err)
		}
		content.Analysis = analysis
	}

	// Calculate multimodal score
	multimodalScore := mre.calculateMultimodalScore(content.Analysis, req)

	// Extract visual and audio insights
	visualInsights := mre.extractVisualInsights(content.Analysis.VisualStyle)
	audioInsights := mre.extractAudioInsights(content.Analysis.AudioFeatures)

	enhanced := &EnhancedRecommendedMovie{
		RecommendedMovie:    movie,
		MultimodalContent:   content,
		MultimodalScore:     multimodalScore,
		VisualInsights:      visualInsights,
		AudioInsights:       audioInsights,
		DetectedGenres:      content.Analysis.Genres,
		MoodAnalysis:        content.Analysis.Mood,
		QualityScore:        content.Analysis.QualityScore,
		ContentWarnings:     content.Analysis.AgeRating.ContentWarnings,
		EnhancementTime:     time.Now(),
	}

	return enhanced, nil
}

// createMultimodalContent creates multimodal content for a movie
func (mre *MultimodalRecommendationEngine) createMultimodalContent(ctx context.Context, movie RecommendedMovie, req *MultimodalRecommendationRequest) (*MultimodalContent, error) {
	content := &MultimodalContent{
		ID:          fmt.Sprintf("movie_%d_%d", movie.ID, time.Now().Unix()),
		MovieID:     movie.ID,
		MovieTitle:  movie.Title,
		Modalities:  make(map[ModalityType]*ContentData),
		ProcessedAt: time.Now(),
	}

	// Load content from various sources (URLs, data, etc.)
	if req.ContentSources != nil {
		for modalityType, source := range req.ContentSources {
			contentData, err := mre.loadContentFromSource(modalityType, source)
			if err != nil {
				mre.logger.WithError(err).WithField("modality", modalityType).Warn("Failed to load content from source")
				continue
			}
			content.Modalities[modalityType] = contentData
		}
	}

	// If no external sources, create mock content for testing
	if len(content.Modalities) == 0 {
		content.Modalities = mre.createMockContent(movie)
	}

	return content, nil
}

// loadContentFromSource loads content from a source based on modality type
func (mre *MultimodalRecommendationEngine) loadContentFromSource(modalityType ModalityType, source ContentSource) (*ContentData, error) {
	switch source.Type {
	case "url":
		return LoadContentFromURL(source.URL, modalityType)
	case "base64":
		return &ContentData{
			Type:        modalityType,
			Base64:      source.Data,
			Metadata:    source.Metadata,
			ProcessedAt: time.Now(),
		}, nil
	case "text":
		return &ContentData{
			Type:        modalityType,
			Data:        []byte(source.Data),
			Metadata:    source.Metadata,
			ProcessedAt: time.Now(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}
}

// createMockContent creates mock multimodal content for testing
func (mre *MultimodalRecommendationEngine) createMockContent(movie RecommendedMovie) map[ModalityType]*ContentData {
	modalities := make(map[ModalityType]*ContentData)

	// Mock image content (movie poster)
	modalities[ModalityImage] = &ContentData{
		Type: ModalityImage,
		URL:  fmt.Sprintf("https://example.com/posters/movie_%d.jpg", movie.ID),
		Metadata: map[string]interface{}{
			"width":  1920,
			"height": 1080,
			"format": "jpeg",
		},
		ProcessedAt: time.Now(),
	}

	// Mock text content (synopsis)
	synopsis := fmt.Sprintf("An exciting %s film featuring %s. %s", 
		movie.Genres[0], movie.Title, movie.Description)
	modalities[ModalityText] = &ContentData{
		Type: ModalityText,
		Data: []byte(synopsis),
		Metadata: map[string]interface{}{
			"language": "en",
			"length":   len(synopsis),
		},
		ProcessedAt: time.Now(),
	}

	// Mock audio content (trailer soundtrack)
	if len(movie.Genres) > 0 && (movie.Genres[0] == "Action" || movie.Genres[0] == "Thriller") {
		modalities[ModalityAudio] = &ContentData{
			Type: ModalityAudio,
			URL:  fmt.Sprintf("https://example.com/trailers/movie_%d.mp3", movie.ID),
			Metadata: map[string]interface{}{
				"duration": 120,
				"format":   "mp3",
				"bitrate":  320,
			},
			ProcessedAt: time.Now(),
		}
	}

	return modalities
}

// calculateMultimodalScore calculates a score based on multimodal analysis
func (mre *MultimodalRecommendationEngine) calculateMultimodalScore(analysis *MultimodalAnalysis, req *MultimodalRecommendationRequest) float64 {
	score := 0.0
	factors := 0

	// Factor in overall sentiment
	if analysis.OverallSentiment > 0 {
		score += analysis.OverallSentiment * 0.3
		factors++
	}

	// Factor in quality score
	if analysis.QualityScore > 0 {
		score += analysis.QualityScore * 0.4
		factors++
	}

	// Factor in genre alignment
	if req.UserProfile != nil && len(analysis.Genres) > 0 {
		genreAlignment := mre.calculateGenreAlignment(analysis.Genres, req.UserProfile.Preferences)
		score += genreAlignment * 0.3
		factors++
	}

	if factors == 0 {
		return 0.5 // Default neutral score
	}

	return score / float64(factors)
}

// calculateGenreAlignment calculates how well detected genres align with user preferences
func (mre *MultimodalRecommendationEngine) calculateGenreAlignment(detectedGenres []GenreDetection, userPreferences map[string]float64) float64 {
	if len(detectedGenres) == 0 || len(userPreferences) == 0 {
		return 0.5
	}

	totalAlignment := 0.0
	totalWeight := 0.0

	for _, genre := range detectedGenres {
		if preference, exists := userPreferences[genre.Genre]; exists {
			weight := genre.Confidence
			totalAlignment += preference * weight
			totalWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0.5
	}

	return totalAlignment / totalWeight
}

// extractVisualInsights extracts key insights from visual analysis
func (mre *MultimodalRecommendationEngine) extractVisualInsights(visual *VisualStyleAnalysis) []string {
	if visual == nil {
		return []string{}
	}

	insights := []string{}

	if visual.CinematicStyle != "" {
		insights = append(insights, fmt.Sprintf("Cinematic style: %s", visual.CinematicStyle))
	}

	if len(visual.ColorPalette) > 0 {
		insights = append(insights, fmt.Sprintf("Color palette: %v", visual.ColorPalette))
	}

	if visual.SceneType != "" {
		insights = append(insights, fmt.Sprintf("Scene type: %s", visual.SceneType))
	}

	if len(visual.ObjectsDetected) > 0 {
		insights = append(insights, fmt.Sprintf("Key elements: %v", visual.ObjectsDetected))
	}

	return insights
}

// extractAudioInsights extracts key insights from audio analysis
func (mre *MultimodalRecommendationEngine) extractAudioInsights(audio *AudioFeatureAnalysis) []string {
	if audio == nil {
		return []string{}
	}

	insights := []string{}

	if audio.MusicGenre != "" {
		insights = append(insights, fmt.Sprintf("Music style: %s", audio.MusicGenre))
	}

	if audio.Energy > 0.7 {
		insights = append(insights, "High-energy soundtrack")
	} else if audio.Energy < 0.3 {
		insights = append(insights, "Calm, atmospheric audio")
	}

	if audio.Tempo > 120 {
		insights = append(insights, "Fast-paced audio")
	} else if audio.Tempo < 80 {
		insights = append(insights, "Slow, deliberate pacing")
	}

	if len(audio.SoundEffects) > 0 {
		insights = append(insights, fmt.Sprintf("Sound effects: %v", audio.SoundEffects))
	}

	return insights
}

// enhanceExplanationsWithMultimodal enhances explanations with multimodal insights
func (mre *MultimodalRecommendationEngine) enhanceExplanationsWithMultimodal(explanations []RecommendationExplanation, enhancedMovies []EnhancedRecommendedMovie) []EnhancedRecommendationExplanation {
	enhanced := []EnhancedRecommendationExplanation{}

	for i, explanation := range explanations {
		if i < len(enhancedMovies) {
			movie := enhancedMovies[i]
			
			enhancedExplanation := EnhancedRecommendationExplanation{
				RecommendationExplanation: explanation,
				MultimodalInsights:        mre.generateMultimodalInsights(movie),
				VisualExplanation:         mre.generateVisualExplanation(movie.VisualInsights),
				AudioExplanation:          mre.generateAudioExplanation(movie.AudioInsights),
				ContentQualityReason:      mre.generateQualityReason(movie.QualityScore),
				EnhancementTime:           time.Now(),
			}

			enhanced = append(enhanced, enhancedExplanation)
		}
	}

	return enhanced
}

// generateMultimodalInsights generates insights from multimodal analysis
func (mre *MultimodalRecommendationEngine) generateMultimodalInsights(movie EnhancedRecommendedMovie) []string {
	insights := []string{}

	// Add visual insights
	insights = append(insights, movie.VisualInsights...)

	// Add audio insights
	insights = append(insights, movie.AudioInsights...)

	// Add genre insights
	for _, genre := range movie.DetectedGenres {
		if genre.Confidence > 0.7 {
			insights = append(insights, fmt.Sprintf("Strong %s genre indicators (%.0f%% confidence)", 
				genre.Genre, genre.Confidence*100))
		}
	}

	// Add mood insights
	for _, mood := range movie.MoodAnalysis {
		if mood.Confidence > 0.6 {
			insights = append(insights, fmt.Sprintf("Conveys %s mood with %.0f%% intensity", 
				mood.Mood, mood.Intensity*100))
		}
	}

	return insights
}

// generateVisualExplanation generates explanation based on visual analysis
func (mre *MultimodalRecommendationEngine) generateVisualExplanation(visualInsights []string) string {
	if len(visualInsights) == 0 {
		return ""
	}

	if len(visualInsights) == 1 {
		return fmt.Sprintf("Visual analysis shows: %s", visualInsights[0])
	}

	return fmt.Sprintf("Visual analysis reveals %s and %s", 
		visualInsights[0], visualInsights[1])
}

// generateAudioExplanation generates explanation based on audio analysis
func (mre *MultimodalRecommendationEngine) generateAudioExplanation(audioInsights []string) string {
	if len(audioInsights) == 0 {
		return ""
	}

	if len(audioInsights) == 1 {
		return fmt.Sprintf("Audio features: %s", audioInsights[0])
	}

	return fmt.Sprintf("Audio analysis indicates %s with %s", 
		audioInsights[0], audioInsights[1])
}

// generateQualityReason generates explanation based on quality score
func (mre *MultimodalRecommendationEngine) generateQualityReason(qualityScore float64) string {
	if qualityScore >= 0.8 {
		return "High production quality evident from multimodal analysis"
	} else if qualityScore >= 0.6 {
		return "Good overall production quality"
	} else if qualityScore >= 0.4 {
		return "Adequate production quality"
	} else {
		return "Mixed quality indicators from content analysis"
	}
}

// calculateAverageMultimodalScore calculates average multimodal score
func (mre *MultimodalRecommendationEngine) calculateAverageMultimodalScore(movies []EnhancedRecommendedMovie) float64 {
	if len(movies) == 0 {
		return 0.0
	}

	total := 0.0
	for _, movie := range movies {
		total += movie.MultimodalScore
	}

	return total / float64(len(movies))
}

// Supporting types for multimodal recommendation system

// ContentSource represents a source of multimodal content
type ContentSource struct {
	Type     string                 `json:"type"`     // "url", "base64", "text"
	URL      string                 `json:"url,omitempty"`
	Data     string                 `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MultimodalRecommendationRequest represents a request with multimodal content
type MultimodalRecommendationRequest struct {
	UserID            string                            `json:"user_id"`
	SessionID         string                            `json:"session_id"`
	Query             string                            `json:"query"`
	UserProfile       *UserProfile                      `json:"user_profile,omitempty"`
	SimilarMovies     []SimilarMovie                    `json:"similar_movies,omitempty"`
	TransparencyLevel TransparencyLevel                 `json:"transparency_level"`
	ContentSources    map[ModalityType]ContentSource    `json:"content_sources,omitempty"`
	AnalysisPreferences *MultimodalAnalysisPreferences `json:"analysis_preferences,omitempty"`
}

// MultimodalAnalysisPreferences specifies analysis preferences
type MultimodalAnalysisPreferences struct {
	IncludeVisualAnalysis bool `json:"include_visual_analysis"`
	IncludeAudioAnalysis  bool `json:"include_audio_analysis"`
	IncludeTextAnalysis   bool `json:"include_text_analysis"`
	DetailLevel           string `json:"detail_level"` // "basic", "detailed", "comprehensive"
}

// MultimodalRecommendationResult represents result with multimodal enhancements
type MultimodalRecommendationResult struct {
	ExplainableRecommendationResult
	EnhancedMovies             []EnhancedRecommendedMovie      `json:"enhanced_movies"`
	EnhancedExplanations       []EnhancedRecommendationExplanation `json:"enhanced_explanations"`
	MultimodalProcessingTime   time.Duration                   `json:"multimodal_processing_time"`
	ContentAnalysisCount       int                             `json:"content_analysis_count"`
	AverageMultimodalScore     float64                         `json:"average_multimodal_score"`
}

// EnhancedRecommendedMovie represents a movie enhanced with multimodal analysis
type EnhancedRecommendedMovie struct {
	RecommendedMovie  
	MultimodalContent   *MultimodalContent `json:"multimodal_content"`
	MultimodalScore     float64            `json:"multimodal_score"`
	VisualInsights      []string           `json:"visual_insights"`
	AudioInsights       []string           `json:"audio_insights"`
	DetectedGenres      []GenreDetection   `json:"detected_genres"`
	MoodAnalysis        []MoodAnalysis     `json:"mood_analysis"`
	QualityScore        float64            `json:"quality_score"`
	ContentWarnings     []string           `json:"content_warnings"`
	EnhancementTime     time.Time          `json:"enhancement_time"`
}

// EnhancedRecommendationExplanation represents explanation enhanced with multimodal insights
type EnhancedRecommendationExplanation struct {
	RecommendationExplanation
	MultimodalInsights   []string  `json:"multimodal_insights"`
	VisualExplanation    string    `json:"visual_explanation"`
	AudioExplanation     string    `json:"audio_explanation"`
	ContentQualityReason string    `json:"content_quality_reason"`
	EnhancementTime      time.Time `json:"enhancement_time"`
}

// Database operations

// StoreContent stores multimodal content
func (db *MultimodalContentDB) StoreContent(content *MultimodalContent) {
	db.content[content.MovieID] = content
	db.logger.WithFields(logrus.Fields{
		"movie_id":     content.MovieID,
		"movie_title":  content.MovieTitle,
		"modalities":   len(content.Modalities),
	}).Info("Stored multimodal content")
}

// GetContent retrieves multimodal content by movie ID
func (db *MultimodalContentDB) GetContent(movieID int) (*MultimodalContent, bool) {
	content, exists := db.content[movieID]
	return content, exists
}

// ListContent lists all stored content
func (db *MultimodalContentDB) ListContent() []*MultimodalContent {
	var result []*MultimodalContent
	for _, content := range db.content {
		result = append(result, content)
	}
	return result
}

// GetMultimodalAnalyzer returns the multimodal analyzer for direct access
func (mre *MultimodalRecommendationEngine) GetMultimodalAnalyzer() *MultimodalAnalyzer {
	return mre.multimodalAnalyzer
}

// GetContentDB returns the content database for direct access
func (mre *MultimodalRecommendationEngine) GetContentDB() *MultimodalContentDB {
	return mre.contentDB
}