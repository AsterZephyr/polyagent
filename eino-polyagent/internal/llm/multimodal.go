package llm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ModalityType represents different types of content modalities
type ModalityType string

const (
	ModalityText  ModalityType = "text"
	ModalityImage ModalityType = "image"
	ModalityAudio ModalityType = "audio"
	ModalityVideo ModalityType = "video"
)

// MultimodalContent represents content with multiple modalities
type MultimodalContent struct {
	ID          string                 `json:"id"`
	MovieID     int                    `json:"movie_id"`
	MovieTitle  string                 `json:"movie_title"`
	Modalities  map[ModalityType]*ContentData `json:"modalities"`
	Analysis    *MultimodalAnalysis    `json:"analysis"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// ContentData represents data for a specific modality
type ContentData struct {
	Type        ModalityType           `json:"type"`
	URL         string                 `json:"url,omitempty"`
	Data        []byte                 `json:"data,omitempty"`
	Base64      string                 `json:"base64,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// MultimodalAnalysis represents comprehensive analysis across modalities
type MultimodalAnalysis struct {
	OverallSentiment float64                       `json:"overall_sentiment"`
	Genres           []GenreDetection              `json:"genres"`
	Mood             []MoodAnalysis                `json:"mood"`
	VisualStyle      *VisualStyleAnalysis          `json:"visual_style"`
	AudioFeatures    *AudioFeatureAnalysis         `json:"audio_features"`
	ContentThemes    []string                      `json:"content_themes"`
	AgeRating        *AgeRatingAnalysis            `json:"age_rating"`
	QualityScore     float64                       `json:"quality_score"`
	AnalysisTime     time.Duration                 `json:"analysis_time"`
	Confidence       map[ModalityType]float64      `json:"confidence"`
}

// GenreDetection represents detected genre with confidence
type GenreDetection struct {
	Genre      string  `json:"genre"`
	Confidence float64 `json:"confidence"`
	Evidence   []string `json:"evidence"`
}

// MoodAnalysis represents detected mood/emotion
type MoodAnalysis struct {
	Mood       string  `json:"mood"`
	Intensity  float64 `json:"intensity"`
	Confidence float64 `json:"confidence"`
}

// VisualStyleAnalysis represents visual content analysis
type VisualStyleAnalysis struct {
	ColorPalette    []string `json:"color_palette"`
	Brightness      float64  `json:"brightness"`
	Contrast        float64  `json:"contrast"`
	Composition     string   `json:"composition"`
	CinematicStyle  string   `json:"cinematic_style"`
	ObjectsDetected []string `json:"objects_detected"`
	FaceCount       int      `json:"face_count"`
	SceneType       string   `json:"scene_type"`
}

// AudioFeatureAnalysis represents audio content analysis
type AudioFeatureAnalysis struct {
	Tempo           float64  `json:"tempo"`
	Energy          float64  `json:"energy"`
	Valence         float64  `json:"valence"`
	Instrumentality float64  `json:"instrumentality"`
	SpeechRatio     float64  `json:"speech_ratio"`
	MusicGenre      string   `json:"music_genre"`
	SoundEffects    []string `json:"sound_effects"`
	VoiceCount      int      `json:"voice_count"`
}

// AgeRatingAnalysis represents content appropriateness analysis
type AgeRatingAnalysis struct {
	SuggestedRating string   `json:"suggested_rating"`
	ContentWarnings []string `json:"content_warnings"`
	Violence        float64  `json:"violence"`
	Language        float64  `json:"language"`
	SexualContent   float64  `json:"sexual_content"`
	SubstanceUse    float64  `json:"substance_use"`
}

// MultimodalAnalyzer analyzes multimedia content for movies
type MultimodalAnalyzer struct {
	logger         *logrus.Logger
	imageAnalyzer  *ImageAnalyzer
	audioAnalyzer  *AudioAnalyzer
	textAnalyzer   *TextAnalyzer
	videoAnalyzer  *VideoAnalyzer
	llmAdapter     *IntentAwareLLMAdapter
	metrics        *MultimodalMetrics
}

// MultimodalMetrics tracks analysis performance
type MultimodalMetrics struct {
	TotalAnalyses     int64                    `json:"total_analyses"`
	AnalysesByType    map[ModalityType]int64   `json:"analyses_by_type"`
	AverageProcessTime map[ModalityType]time.Duration `json:"average_process_time"`
	SuccessRate       map[ModalityType]float64 `json:"success_rate"`
	LastUpdated       time.Time                `json:"last_updated"`
}

// NewMultimodalAnalyzer creates a new multimodal content analyzer
func NewMultimodalAnalyzer(llmAdapter *IntentAwareLLMAdapter, logger *logrus.Logger) *MultimodalAnalyzer {
	return &MultimodalAnalyzer{
		logger:        logger,
		imageAnalyzer: NewImageAnalyzer(llmAdapter, logger),
		audioAnalyzer: NewAudioAnalyzer(logger),
		textAnalyzer:  NewTextAnalyzer(logger),
		videoAnalyzer: NewVideoAnalyzer(logger),
		llmAdapter:    llmAdapter,
		metrics: &MultimodalMetrics{
			AnalysesByType:     make(map[ModalityType]int64),
			AverageProcessTime: make(map[ModalityType]time.Duration),
			SuccessRate:        make(map[ModalityType]float64),
			LastUpdated:        time.Now(),
		},
	}
}

// AnalyzeMultimodalContent analyzes content across multiple modalities
func (ma *MultimodalAnalyzer) AnalyzeMultimodalContent(ctx context.Context, content *MultimodalContent) (*MultimodalAnalysis, error) {
	startTime := time.Now()
	
	analysis := &MultimodalAnalysis{
		Confidence: make(map[ModalityType]float64),
	}

	// Analyze each modality in parallel
	analysisResults := make(chan modalityResult, len(content.Modalities))
	
	for modalityType, contentData := range content.Modalities {
		go func(mt ModalityType, cd *ContentData) {
			result := ma.analyzeModality(ctx, mt, cd)
			analysisResults <- result
		}(modalityType, contentData)
	}

	// Collect results
	modalityAnalyses := make(map[ModalityType]interface{})
	for i := 0; i < len(content.Modalities); i++ {
		result := <-analysisResults
		if result.err != nil {
			ma.logger.WithError(result.err).WithField("modality", result.modalityType).Warn("Modality analysis failed")
			continue
		}
		modalityAnalyses[result.modalityType] = result.analysis
		analysis.Confidence[result.modalityType] = result.confidence
	}

	// Synthesize cross-modal analysis
	ma.synthesizeAnalysis(analysis, modalityAnalyses)
	
	analysis.AnalysisTime = time.Since(startTime)
	
	// Update metrics
	ma.updateMetrics(content.Modalities, time.Since(startTime))

	ma.logger.WithFields(logrus.Fields{
		"movie_id":       content.MovieID,
		"modalities":     len(content.Modalities),
		"analysis_time":  analysis.AnalysisTime,
		"quality_score":  analysis.QualityScore,
	}).Info("Completed multimodal content analysis")

	return analysis, nil
}

// modalityResult represents the result of analyzing a single modality
type modalityResult struct {
	modalityType ModalityType
	analysis     interface{}
	confidence   float64
	err          error
}

// analyzeModality analyzes content for a specific modality
func (ma *MultimodalAnalyzer) analyzeModality(ctx context.Context, modalityType ModalityType, content *ContentData) modalityResult {
	startTime := time.Now()
	defer func() {
		ma.metrics.AverageProcessTime[modalityType] = time.Since(startTime)
	}()

	switch modalityType {
	case ModalityImage:
		analysis, confidence, err := ma.imageAnalyzer.AnalyzeImage(ctx, content)
		return modalityResult{modalityType, analysis, confidence, err}

	case ModalityAudio:
		analysis, confidence, err := ma.audioAnalyzer.AnalyzeAudio(ctx, content)
		return modalityResult{modalityType, analysis, confidence, err}

	case ModalityText:
		analysis, confidence, err := ma.textAnalyzer.AnalyzeText(ctx, content)
		return modalityResult{modalityType, analysis, confidence, err}

	case ModalityVideo:
		analysis, confidence, err := ma.videoAnalyzer.AnalyzeVideo(ctx, content)
		return modalityResult{modalityType, analysis, confidence, err}

	default:
		return modalityResult{modalityType, nil, 0, fmt.Errorf("unsupported modality type: %s", modalityType)}
	}
}

// synthesizeAnalysis combines analyses from different modalities
func (ma *MultimodalAnalyzer) synthesizeAnalysis(analysis *MultimodalAnalysis, modalityAnalyses map[ModalityType]interface{}) {
	// Combine genre detections
	analysis.Genres = ma.combineGenreDetections(modalityAnalyses)
	
	// Combine mood analyses
	analysis.Mood = ma.combineMoodAnalyses(modalityAnalyses)
	
	// Extract visual style if available
	if visualAnalysis, ok := modalityAnalyses[ModalityImage].(*VisualStyleAnalysis); ok {
		analysis.VisualStyle = visualAnalysis
	}
	
	// Extract audio features if available
	if audioAnalysis, ok := modalityAnalyses[ModalityAudio].(*AudioFeatureAnalysis); ok {
		analysis.AudioFeatures = audioAnalysis
	}
	
	// Calculate overall sentiment
	analysis.OverallSentiment = ma.calculateOverallSentiment(modalityAnalyses)
	
	// Extract content themes
	analysis.ContentThemes = ma.extractContentThemes(modalityAnalyses)
	
	// Analyze age rating
	analysis.AgeRating = ma.analyzeAgeRating(modalityAnalyses)
	
	// Calculate quality score
	analysis.QualityScore = ma.calculateQualityScore(modalityAnalyses, analysis.Confidence)
}

// ImageAnalyzer handles image content analysis
type ImageAnalyzer struct {
	llmAdapter *IntentAwareLLMAdapter
	logger     *logrus.Logger
}

// NewImageAnalyzer creates a new image analyzer
func NewImageAnalyzer(llmAdapter *IntentAwareLLMAdapter, logger *logrus.Logger) *ImageAnalyzer {
	return &ImageAnalyzer{
		llmAdapter: llmAdapter,
		logger:     logger,
	}
}

// AnalyzeImage analyzes image content (movie posters, stills, etc.)
func (ia *ImageAnalyzer) AnalyzeImage(ctx context.Context, content *ContentData) (*VisualStyleAnalysis, float64, error) {
	// For this implementation, we'll use LLM vision capabilities if available
	// In production, this could integrate with specialized vision APIs
	
	analysis := &VisualStyleAnalysis{
		ColorPalette:    []string{},
		ObjectsDetected: []string{},
	}
	
	// Mock analysis - in production this would use computer vision APIs
	if content.URL != "" {
		// Analyze image from URL
		analysis = ia.analyzeImageFromURL(content.URL)
	} else if len(content.Data) > 0 || content.Base64 != "" {
		// Analyze image from data
		analysis = ia.analyzeImageFromData(content)
	}
	
	// Use LLM for high-level interpretation if we have vision capabilities
	confidence := ia.interpretImageWithLLM(ctx, analysis, content)
	
	return analysis, confidence, nil
}

// analyzeImageFromURL analyzes image from URL
func (ia *ImageAnalyzer) analyzeImageFromURL(url string) *VisualStyleAnalysis {
	// Mock implementation - would use computer vision service
	return &VisualStyleAnalysis{
		ColorPalette:    []string{"dark blue", "orange", "black"},
		Brightness:      0.4,
		Contrast:        0.7,
		Composition:     "rule of thirds",
		CinematicStyle:  "dramatic",
		ObjectsDetected: []string{"person", "building", "vehicle"},
		FaceCount:       2,
		SceneType:       "action scene",
	}
}

// analyzeImageFromData analyzes image from raw data
func (ia *ImageAnalyzer) analyzeImageFromData(content *ContentData) *VisualStyleAnalysis {
	// Mock implementation - would process image data
	return &VisualStyleAnalysis{
		ColorPalette:    []string{"red", "black", "white"},
		Brightness:      0.6,
		Contrast:        0.8,
		Composition:     "centered",
		CinematicStyle:  "thriller",
		ObjectsDetected: []string{"person", "weapon", "car"},
		FaceCount:       1,
		SceneType:       "character portrait",
	}
}

// interpretImageWithLLM uses LLM for high-level image interpretation
func (ia *ImageAnalyzer) interpretImageWithLLM(ctx context.Context, analysis *VisualStyleAnalysis, content *ContentData) float64 {
	// This would use GPT-4V or Claude-3 vision capabilities in production
	// For now, return mock confidence based on detected features
	confidence := 0.7
	
	if len(analysis.ObjectsDetected) > 3 {
		confidence += 0.1
	}
	if analysis.FaceCount > 0 {
		confidence += 0.1
	}
	if analysis.CinematicStyle != "" {
		confidence += 0.1
	}
	
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// AudioAnalyzer handles audio content analysis
type AudioAnalyzer struct {
	logger *logrus.Logger
}

// NewAudioAnalyzer creates a new audio analyzer
func NewAudioAnalyzer(logger *logrus.Logger) *AudioAnalyzer {
	return &AudioAnalyzer{
		logger: logger,
	}
}

// AnalyzeAudio analyzes audio content (trailers, soundtracks, etc.)
func (aa *AudioAnalyzer) AnalyzeAudio(ctx context.Context, content *ContentData) (*AudioFeatureAnalysis, float64, error) {
	// Mock audio analysis - in production would use audio processing libraries
	analysis := &AudioFeatureAnalysis{
		Tempo:           120.0,
		Energy:          0.8,
		Valence:         0.6,
		Instrumentality: 0.7,
		SpeechRatio:     0.3,
		MusicGenre:      "orchestral",
		SoundEffects:    []string{"explosion", "gunshot", "engine"},
		VoiceCount:      2,
	}
	
	confidence := 0.75 // Mock confidence
	
	return analysis, confidence, nil
}

// TextAnalyzer handles text content analysis
type TextAnalyzer struct {
	logger *logrus.Logger
}

// NewTextAnalyzer creates a new text analyzer
func NewTextAnalyzer(logger *logrus.Logger) *TextAnalyzer {
	return &TextAnalyzer{
		logger: logger,
	}
}

// AnalyzeText analyzes text content (descriptions, reviews, subtitles)
func (ta *TextAnalyzer) AnalyzeText(ctx context.Context, content *ContentData) (*TextAnalysis, float64, error) {
	// Mock text analysis
	analysis := &TextAnalysis{
		Sentiment:     0.6,
		Emotions:      []string{"excitement", "tension", "hope"},
		KeyPhrases:    []string{"action-packed", "thrilling", "spectacular"},
		Topics:        []string{"heroism", "conflict", "technology"},
		Complexity:    0.7,
		ReadingLevel:  "high school",
	}
	
	confidence := 0.8
	
	return analysis, confidence, nil
}

// TextAnalysis represents text content analysis
type TextAnalysis struct {
	Sentiment     float64  `json:"sentiment"`
	Emotions      []string `json:"emotions"`
	KeyPhrases    []string `json:"key_phrases"`
	Topics        []string `json:"topics"`
	Complexity    float64  `json:"complexity"`
	ReadingLevel  string   `json:"reading_level"`
}

// VideoAnalyzer handles video content analysis
type VideoAnalyzer struct {
	logger *logrus.Logger
}

// NewVideoAnalyzer creates a new video analyzer
func NewVideoAnalyzer(logger *logrus.Logger) *VideoAnalyzer {
	return &VideoAnalyzer{
		logger: logger,
	}
}

// AnalyzeVideo analyzes video content (trailers, clips)
func (va *VideoAnalyzer) AnalyzeVideo(ctx context.Context, content *ContentData) (*VideoAnalysis, float64, error) {
	// Mock video analysis
	analysis := &VideoAnalysis{
		Duration:      120.0,
		FrameRate:     24.0,
		Resolution:    "1920x1080",
		SceneChanges:  15,
		ActionLevel:   0.8,
		CutFrequency:  0.5,
		CameraMovement: "dynamic",
		ColorGrading:  "high contrast",
	}
	
	confidence := 0.7
	
	return analysis, confidence, nil
}

// VideoAnalysis represents video content analysis
type VideoAnalysis struct {
	Duration       float64 `json:"duration"`
	FrameRate      float64 `json:"frame_rate"`
	Resolution     string  `json:"resolution"`
	SceneChanges   int     `json:"scene_changes"`
	ActionLevel    float64 `json:"action_level"`
	CutFrequency   float64 `json:"cut_frequency"`
	CameraMovement string  `json:"camera_movement"`
	ColorGrading   string  `json:"color_grading"`
}

// Helper methods for synthesis

func (ma *MultimodalAnalyzer) combineGenreDetections(analyses map[ModalityType]interface{}) []GenreDetection {
	genreMap := make(map[string]*GenreDetection)
	
	// Collect genre evidence from all modalities
	for modalityType, analysis := range analyses {
		genres := ma.extractGenresFromAnalysis(modalityType, analysis)
		for _, genre := range genres {
			if existing, exists := genreMap[genre.Genre]; exists {
				// Combine confidence scores
				existing.Confidence = (existing.Confidence + genre.Confidence) / 2
				existing.Evidence = append(existing.Evidence, genre.Evidence...)
			} else {
				genreMap[genre.Genre] = &genre
			}
		}
	}
	
	// Convert map to slice
	var result []GenreDetection
	for _, genre := range genreMap {
		result = append(result, *genre)
	}
	
	return result
}

func (ma *MultimodalAnalyzer) extractGenresFromAnalysis(modalityType ModalityType, analysis interface{}) []GenreDetection {
	switch modalityType {
	case ModalityImage:
		if visual, ok := analysis.(*VisualStyleAnalysis); ok {
			return ma.inferGenresFromVisual(visual)
		}
	case ModalityAudio:
		if audio, ok := analysis.(*AudioFeatureAnalysis); ok {
			return ma.inferGenresFromAudio(audio)
		}
	case ModalityText:
		if text, ok := analysis.(*TextAnalysis); ok {
			return ma.inferGenresFromText(text)
		}
	}
	return []GenreDetection{}
}

func (ma *MultimodalAnalyzer) inferGenresFromVisual(visual *VisualStyleAnalysis) []GenreDetection {
	genres := []GenreDetection{}
	
	// Infer genres from visual style
	if visual.CinematicStyle == "dramatic" {
		genres = append(genres, GenreDetection{
			Genre: "Drama", Confidence: 0.7, Evidence: []string{"dramatic cinematography"},
		})
	}
	if visual.CinematicStyle == "thriller" {
		genres = append(genres, GenreDetection{
			Genre: "Thriller", Confidence: 0.8, Evidence: []string{"thriller visual style"},
		})
	}
	if strings.Contains(strings.Join(visual.ObjectsDetected, " "), "weapon") {
		genres = append(genres, GenreDetection{
			Genre: "Action", Confidence: 0.6, Evidence: []string{"weapons detected"},
		})
	}
	
	return genres
}

func (ma *MultimodalAnalyzer) inferGenresFromAudio(audio *AudioFeatureAnalysis) []GenreDetection {
	genres := []GenreDetection{}
	
	if audio.Energy > 0.7 {
		genres = append(genres, GenreDetection{
			Genre: "Action", Confidence: 0.7, Evidence: []string{"high energy audio"},
		})
	}
	if audio.MusicGenre == "orchestral" {
		genres = append(genres, GenreDetection{
			Genre: "Drama", Confidence: 0.6, Evidence: []string{"orchestral soundtrack"},
		})
	}
	
	return genres
}

func (ma *MultimodalAnalyzer) inferGenresFromText(text *TextAnalysis) []GenreDetection {
	genres := []GenreDetection{}
	
	for _, phrase := range text.KeyPhrases {
		if strings.Contains(phrase, "action") {
			genres = append(genres, GenreDetection{
				Genre: "Action", Confidence: 0.8, Evidence: []string{"action keywords"},
			})
		}
		if strings.Contains(phrase, "thriller") || strings.Contains(phrase, "suspense") {
			genres = append(genres, GenreDetection{
				Genre: "Thriller", Confidence: 0.7, Evidence: []string{"thriller keywords"},
			})
		}
	}
	
	return genres
}

func (ma *MultimodalAnalyzer) combineMoodAnalyses(analyses map[ModalityType]interface{}) []MoodAnalysis {
	// Mock mood combination
	return []MoodAnalysis{
		{Mood: "intense", Intensity: 0.8, Confidence: 0.7},
		{Mood: "exciting", Intensity: 0.9, Confidence: 0.8},
	}
}

func (ma *MultimodalAnalyzer) calculateOverallSentiment(analyses map[ModalityType]interface{}) float64 {
	totalSentiment := 0.0
	count := 0
	
	for modalityType, analysis := range analyses {
		sentiment := ma.extractSentimentFromAnalysis(modalityType, analysis)
		if sentiment != 0 {
			totalSentiment += sentiment
			count++
		}
	}
	
	if count == 0 {
		return 0.5 // Neutral
	}
	
	return totalSentiment / float64(count)
}

func (ma *MultimodalAnalyzer) extractSentimentFromAnalysis(modalityType ModalityType, analysis interface{}) float64 {
	switch modalityType {
	case ModalityText:
		if text, ok := analysis.(*TextAnalysis); ok {
			return text.Sentiment
		}
	case ModalityAudio:
		if audio, ok := analysis.(*AudioFeatureAnalysis); ok {
			return audio.Valence
		}
	}
	return 0
}

func (ma *MultimodalAnalyzer) extractContentThemes(analyses map[ModalityType]interface{}) []string {
	themes := []string{}
	
	for modalityType, analysis := range analyses {
		modalityThemes := ma.extractThemesFromAnalysis(modalityType, analysis)
		themes = append(themes, modalityThemes...)
	}
	
	// Remove duplicates
	uniqueThemes := make(map[string]bool)
	var result []string
	for _, theme := range themes {
		if !uniqueThemes[theme] {
			uniqueThemes[theme] = true
			result = append(result, theme)
		}
	}
	
	return result
}

func (ma *MultimodalAnalyzer) extractThemesFromAnalysis(modalityType ModalityType, analysis interface{}) []string {
	switch modalityType {
	case ModalityText:
		if text, ok := analysis.(*TextAnalysis); ok {
			return text.Topics
		}
	}
	return []string{}
}

func (ma *MultimodalAnalyzer) analyzeAgeRating(analyses map[ModalityType]interface{}) *AgeRatingAnalysis {
	// Mock age rating analysis
	return &AgeRatingAnalysis{
		SuggestedRating: "PG-13",
		ContentWarnings: []string{"violence", "intense scenes"},
		Violence:        0.6,
		Language:        0.2,
		SexualContent:   0.1,
		SubstanceUse:    0.1,
	}
}

func (ma *MultimodalAnalyzer) calculateQualityScore(analyses map[ModalityType]interface{}, confidence map[ModalityType]float64) float64 {
	totalScore := 0.0
	totalWeight := 0.0
	
	for modalityType, conf := range confidence {
		weight := ma.getModalityWeight(modalityType)
		totalScore += conf * weight
		totalWeight += weight
	}
	
	if totalWeight == 0 {
		return 0.5
	}
	
	return totalScore / totalWeight
}

func (ma *MultimodalAnalyzer) getModalityWeight(modalityType ModalityType) float64 {
	weights := map[ModalityType]float64{
		ModalityImage: 0.3,
		ModalityAudio: 0.25,
		ModalityText:  0.25,
		ModalityVideo: 0.2,
	}
	
	if weight, exists := weights[modalityType]; exists {
		return weight
	}
	return 0.1
}

func (ma *MultimodalAnalyzer) updateMetrics(modalities map[ModalityType]*ContentData, processingTime time.Duration) {
	ma.metrics.TotalAnalyses++
	
	for modalityType := range modalities {
		ma.metrics.AnalysesByType[modalityType]++
	}
	
	ma.metrics.LastUpdated = time.Now()
}

// GetMetrics returns multimodal analysis metrics
func (ma *MultimodalAnalyzer) GetMetrics() *MultimodalMetrics {
	return ma.metrics
}

// LoadContentFromURL loads content from a URL
func LoadContentFromURL(url string, modalityType ModalityType) (*ContentData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch content: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	content := &ContentData{
		Type:        modalityType,
		URL:         url,
		Data:        data,
		Metadata:    make(map[string]interface{}),
		ProcessedAt: time.Now(),
	}

	// Convert to base64 for image/audio content
	if modalityType == ModalityImage || modalityType == ModalityAudio {
		content.Base64 = base64.StdEncoding.EncodeToString(data)
	}

	// Add basic metadata
	content.Metadata["content_length"] = len(data)
	content.Metadata["content_type"] = resp.Header.Get("Content-Type")

	return content, nil
}