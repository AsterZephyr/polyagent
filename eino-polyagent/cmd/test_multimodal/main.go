package main

import (
	"context"
	"fmt"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🎭 Testing Multimodal Content Analysis System")
	fmt.Println("============================================")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test 1: Multimodal Content Creation
	fmt.Println("\n📁 Test 1: Multimodal Content Creation")
	testMultimodalContentCreation(logger)

	// Test 2: Individual Modality Analysis
	fmt.Println("\n🔍 Test 2: Individual Modality Analysis")
	testModalityAnalysis(logger)

	// Test 3: Multimodal Analysis Integration
	fmt.Println("\n🧠 Test 3: Multimodal Analysis Integration")
	testMultimodalAnalysis(logger)

	// Test 4: Multimodal Recommendation Engine (Mock Mode)
	fmt.Println("\n🤖 Test 4: Multimodal Recommendation Engine")
	testMultimodalRecommendationEngine(logger)

	// Test 5: Content Database Operations
	fmt.Println("\n💾 Test 5: Content Database Operations")
	testContentDatabase(logger)

	fmt.Println("\n🎉 Multimodal Content Analysis Testing Completed!")
}

func testMultimodalContentCreation(logger *logrus.Logger) {
	// Create test multimodal content
	content := &llm.MultimodalContent{
		ID:         "test_content_1",
		MovieID:    1001,
		MovieTitle: "Blade Runner 2049",
		Modalities: make(map[llm.ModalityType]*llm.ContentData),
		ProcessedAt: time.Now(),
	}

	// Add image content (movie poster)
	content.Modalities[llm.ModalityImage] = &llm.ContentData{
		Type: llm.ModalityImage,
		URL:  "https://example.com/poster_blade_runner_2049.jpg",
		Metadata: map[string]interface{}{
			"width":       1920,
			"height":      1080,
			"format":      "jpeg",
			"content_type": "movie_poster",
		},
		ProcessedAt: time.Now(),
	}

	// Add text content (synopsis)
	synopsis := "Officer K, a new blade runner for the LAPD, unearths a long-buried secret that has the potential to plunge what's left of society into chaos. His discovery leads him on a quest to find Rick Deckard, a former blade runner who's been missing for 30 years."
	content.Modalities[llm.ModalityText] = &llm.ContentData{
		Type: llm.ModalityText,
		Data: []byte(synopsis),
		Metadata: map[string]interface{}{
			"language":     "en",
			"length":       len(synopsis),
			"content_type": "synopsis",
		},
		ProcessedAt: time.Now(),
	}

	// Add audio content (trailer soundtrack)
	content.Modalities[llm.ModalityAudio] = &llm.ContentData{
		Type: llm.ModalityAudio,
		URL:  "https://example.com/trailer_blade_runner_2049.mp3",
		Metadata: map[string]interface{}{
			"duration":     180,
			"format":       "mp3",
			"bitrate":      320,
			"content_type": "trailer_soundtrack",
		},
		ProcessedAt: time.Now(),
	}

	fmt.Printf("✅ Created multimodal content for %s:\n", content.MovieTitle)
	fmt.Printf("   Content ID: %s\n", content.ID)
	fmt.Printf("   Movie ID: %d\n", content.MovieID)
	fmt.Printf("   Modalities: %d\n", len(content.Modalities))

	for modalityType, modalityData := range content.Modalities {
		fmt.Printf("   📊 %s:\n", modalityType)
		if modalityData.URL != "" {
			fmt.Printf("      URL: %s\n", modalityData.URL)
		}
		if len(modalityData.Data) > 0 {
			fmt.Printf("      Data Length: %d bytes\n", len(modalityData.Data))
		}
		fmt.Printf("      Metadata: %d fields\n", len(modalityData.Metadata))
		fmt.Printf("      Processed: %s\n", modalityData.ProcessedAt.Format("15:04:05"))
	}
}

func testModalityAnalysis(logger *logrus.Logger) {
	// Create mock LLM adapter for image analysis
	config := &llm.LLMAdapterConfig{
		Primary: llm.LLMConfig{
			Provider: llm.ProviderOpenAI,
			Model:    "gpt-4o-mini",
			APIKey:   "mock-key",
		},
	}

	adapter, err := llm.NewIntentAwareLLMAdapter(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create LLM adapter: %v\n", err)
		return
	}

	// Test individual analyzers
	fmt.Printf("Testing individual modality analyzers:\n\n")

	// Test Image Analysis
	fmt.Printf("🖼️  Image Analysis:\n")
	imageAnalyzer := llm.NewImageAnalyzer(adapter, logger)
	imageContent := &llm.ContentData{
		Type: llm.ModalityImage,
		URL:  "https://example.com/action_movie_poster.jpg",
		Metadata: map[string]interface{}{
			"genre_hint": "action",
			"style":      "modern",
		},
	}

	ctx := context.Background()
	visualAnalysis, imageConfidence, err := imageAnalyzer.AnalyzeImage(ctx, imageContent)
	if err != nil {
		fmt.Printf("   ❌ Image analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Image Analysis Results:\n")
		fmt.Printf("      Confidence: %.2f\n", imageConfidence)
		fmt.Printf("      Cinematic Style: %s\n", visualAnalysis.CinematicStyle)
		fmt.Printf("      Color Palette: %v\n", visualAnalysis.ColorPalette)
		fmt.Printf("      Objects Detected: %v\n", visualAnalysis.ObjectsDetected)
		fmt.Printf("      Scene Type: %s\n", visualAnalysis.SceneType)
		fmt.Printf("      Face Count: %d\n", visualAnalysis.FaceCount)
		fmt.Printf("      Brightness: %.2f\n", visualAnalysis.Brightness)
		fmt.Printf("      Contrast: %.2f\n", visualAnalysis.Contrast)
	}

	// Test Audio Analysis
	fmt.Printf("\n🎵 Audio Analysis:\n")
	audioAnalyzer := llm.NewAudioAnalyzer(logger)
	audioContent := &llm.ContentData{
		Type: llm.ModalityAudio,
		URL:  "https://example.com/thriller_soundtrack.mp3",
		Metadata: map[string]interface{}{
			"duration": 120,
			"genre":    "thriller",
		},
	}

	audioAnalysis, audioConfidence, err := audioAnalyzer.AnalyzeAudio(ctx, audioContent)
	if err != nil {
		fmt.Printf("   ❌ Audio analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Audio Analysis Results:\n")
		fmt.Printf("      Confidence: %.2f\n", audioConfidence)
		fmt.Printf("      Music Genre: %s\n", audioAnalysis.MusicGenre)
		fmt.Printf("      Energy: %.2f\n", audioAnalysis.Energy)
		fmt.Printf("      Valence: %.2f\n", audioAnalysis.Valence)
		fmt.Printf("      Tempo: %.1f BPM\n", audioAnalysis.Tempo)
		fmt.Printf("      Sound Effects: %v\n", audioAnalysis.SoundEffects)
		fmt.Printf("      Voice Count: %d\n", audioAnalysis.VoiceCount)
	}

	// Test Text Analysis
	fmt.Printf("\n📝 Text Analysis:\n")
	textAnalyzer := llm.NewTextAnalyzer(logger)
	textContent := &llm.ContentData{
		Type: llm.ModalityText,
		Data: []byte("An intense action-thriller featuring spectacular explosions and heart-pounding chase sequences. A hero must save the world from imminent destruction."),
		Metadata: map[string]interface{}{
			"content_type": "description",
			"language":     "en",
		},
	}

	textAnalysis, textConfidence, err := textAnalyzer.AnalyzeText(ctx, textContent)
	if err != nil {
		fmt.Printf("   ❌ Text analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Text Analysis Results:\n")
		fmt.Printf("      Confidence: %.2f\n", textConfidence)
		fmt.Printf("      Sentiment: %.2f\n", textAnalysis.Sentiment)
		fmt.Printf("      Emotions: %v\n", textAnalysis.Emotions)
		fmt.Printf("      Key Phrases: %v\n", textAnalysis.KeyPhrases)
		fmt.Printf("      Topics: %v\n", textAnalysis.Topics)
		fmt.Printf("      Complexity: %.2f\n", textAnalysis.Complexity)
		fmt.Printf("      Reading Level: %s\n", textAnalysis.ReadingLevel)
	}

	// Test Video Analysis
	fmt.Printf("\n🎬 Video Analysis:\n")
	videoAnalyzer := llm.NewVideoAnalyzer(logger)
	videoContent := &llm.ContentData{
		Type: llm.ModalityVideo,
		URL:  "https://example.com/movie_trailer.mp4",
		Metadata: map[string]interface{}{
			"duration":   180,
			"resolution": "1920x1080",
			"framerate":  24,
		},
	}

	videoAnalysis, videoConfidence, err := videoAnalyzer.AnalyzeVideo(ctx, videoContent)
	if err != nil {
		fmt.Printf("   ❌ Video analysis failed: %v\n", err)
	} else {
		fmt.Printf("   ✅ Video Analysis Results:\n")
		fmt.Printf("      Confidence: %.2f\n", videoConfidence)
		fmt.Printf("      Duration: %.1f seconds\n", videoAnalysis.Duration)
		fmt.Printf("      Frame Rate: %.1f fps\n", videoAnalysis.FrameRate)
		fmt.Printf("      Resolution: %s\n", videoAnalysis.Resolution)
		fmt.Printf("      Scene Changes: %d\n", videoAnalysis.SceneChanges)
		fmt.Printf("      Action Level: %.2f\n", videoAnalysis.ActionLevel)
		fmt.Printf("      Camera Movement: %s\n", videoAnalysis.CameraMovement)
		fmt.Printf("      Color Grading: %s\n", videoAnalysis.ColorGrading)
	}
}

func testMultimodalAnalysis(logger *logrus.Logger) {
	// Create multimodal analyzer
	config := &llm.LLMAdapterConfig{
		Primary: llm.LLMConfig{
			Provider: llm.ProviderOpenAI,
			Model:    "gpt-4o-mini",
			APIKey:   "mock-key",
		},
	}

	adapter, err := llm.NewIntentAwareLLMAdapter(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create LLM adapter: %v\n", err)
		return
	}

	analyzer := llm.NewMultimodalAnalyzer(adapter, logger)

	// Create comprehensive multimodal content
	content := &llm.MultimodalContent{
		ID:         "test_multimodal_1",
		MovieID:    2001,
		MovieTitle: "Mad Max: Fury Road",
		Modalities: map[llm.ModalityType]*llm.ContentData{
			llm.ModalityImage: {
				Type: llm.ModalityImage,
				URL:  "https://example.com/mad_max_poster.jpg",
				Metadata: map[string]interface{}{
					"style": "post_apocalyptic",
					"color": "desert_tones",
				},
			},
			llm.ModalityText: {
				Type: llm.ModalityText,
				Data: []byte("In a post-apocalyptic wasteland, Max teams up with Furiosa to flee from cult leader Immortan Joe and his army in an armored tanker truck, leading to a road war."),
				Metadata: map[string]interface{}{
					"genre_hints": []string{"action", "adventure", "thriller"},
				},
			},
			llm.ModalityAudio: {
				Type: llm.ModalityAudio,
				URL:  "https://example.com/mad_max_trailer.mp3",
				Metadata: map[string]interface{}{
					"energy_level": "high",
					"instruments":  []string{"drums", "guitar", "orchestral"},
				},
			},
		},
		ProcessedAt: time.Now(),
	}

	fmt.Printf("Analyzing multimodal content for %s:\n", content.MovieTitle)

	ctx := context.Background()
	analysis, err := analyzer.AnalyzeMultimodalContent(ctx, content)
	if err != nil {
		fmt.Printf("❌ Multimodal analysis failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Multimodal Analysis Results:\n")
	fmt.Printf("   Analysis Time: %v\n", analysis.AnalysisTime)
	fmt.Printf("   Overall Sentiment: %.2f\n", analysis.OverallSentiment)
	fmt.Printf("   Quality Score: %.2f\n", analysis.QualityScore)

	fmt.Printf("\n   🎭 Detected Genres:\n")
	for _, genre := range analysis.Genres {
		fmt.Printf("      %s (%.0f%% confidence): %v\n", genre.Genre, genre.Confidence*100, genre.Evidence)
	}

	fmt.Printf("\n   😊 Mood Analysis:\n")
	for _, mood := range analysis.Mood {
		fmt.Printf("      %s: %.0f%% intensity, %.0f%% confidence\n", mood.Mood, mood.Intensity*100, mood.Confidence*100)
	}

	if analysis.VisualStyle != nil {
		fmt.Printf("\n   🎨 Visual Style:\n")
		fmt.Printf("      Cinematic Style: %s\n", analysis.VisualStyle.CinematicStyle)
		fmt.Printf("      Color Palette: %v\n", analysis.VisualStyle.ColorPalette)
		fmt.Printf("      Scene Type: %s\n", analysis.VisualStyle.SceneType)
	}

	if analysis.AudioFeatures != nil {
		fmt.Printf("\n   🎵 Audio Features:\n")
		fmt.Printf("      Music Genre: %s\n", analysis.AudioFeatures.MusicGenre)
		fmt.Printf("      Energy: %.2f\n", analysis.AudioFeatures.Energy)
		fmt.Printf("      Tempo: %.1f BPM\n", analysis.AudioFeatures.Tempo)
	}

	fmt.Printf("\n   📊 Content Themes: %v\n", analysis.ContentThemes)

	if analysis.AgeRating != nil {
		fmt.Printf("\n   🔞 Age Rating Analysis:\n")
		fmt.Printf("      Suggested Rating: %s\n", analysis.AgeRating.SuggestedRating)
		fmt.Printf("      Content Warnings: %v\n", analysis.AgeRating.ContentWarnings)
		fmt.Printf("      Violence Level: %.2f\n", analysis.AgeRating.Violence)
	}

	fmt.Printf("\n   🎯 Confidence by Modality:\n")
	for modalityType, confidence := range analysis.Confidence {
		fmt.Printf("      %s: %.2f\n", modalityType, confidence)
	}

	// Test metrics
	metrics := analyzer.GetMetrics()
	fmt.Printf("\n📈 Analyzer Metrics:\n")
	fmt.Printf("   Total Analyses: %d\n", metrics.TotalAnalyses)
	fmt.Printf("   Analyses by Type:\n")
	for modalityType, count := range metrics.AnalysesByType {
		fmt.Printf("      %s: %d\n", modalityType, count)
	}
}

func testMultimodalRecommendationEngine(logger *logrus.Logger) {
	// Create multimodal recommendation engine
	config := &llm.LLMAdapterConfig{
		Primary: llm.LLMConfig{
			Provider:    llm.ProviderOpenAI,
			Model:       "gpt-4o-mini",
			APIKey:      "mock-openai-key",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		},
	}

	engine, err := llm.NewMultimodalRecommendationEngine(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create multimodal recommendation engine: %v\n", err)
		return
	}

	fmt.Printf("✅ Multimodal Recommendation Engine created successfully\n")

	ctx := context.Background()

	// Test multimodal recommendation request
	req := &llm.MultimodalRecommendationRequest{
		UserID:    "multimodal_user",
		SessionID: "multimodal_session",
		Query:     "I want visually stunning sci-fi movies with great soundtracks",
		UserProfile: &llm.UserProfile{
			UserID:      "multimodal_user",
			Preferences: map[string]float64{"Sci-Fi": 0.9, "Action": 0.7},
		},
		TransparencyLevel: llm.TransparencyDetailed,
		ContentSources: map[llm.ModalityType]llm.ContentSource{
			llm.ModalityImage: {
				Type: "url",
				URL:  "https://example.com/scifi_poster.jpg",
			},
			llm.ModalityText: {
				Type: "text",
				Data: "Epic space adventure with stunning visual effects",
			},
		},
		AnalysisPreferences: &llm.MultimodalAnalysisPreferences{
			IncludeVisualAnalysis: true,
			IncludeAudioAnalysis:  true,
			IncludeTextAnalysis:   true,
			DetailLevel:           "comprehensive",
		},
	}

	fmt.Printf("\n🎬 Processing Multimodal Recommendation:\n")
	fmt.Printf("   User: %s\n", req.UserID)
	fmt.Printf("   Query: \"%s\"\n", req.Query)
	fmt.Printf("   Content Sources: %d modalities\n", len(req.ContentSources))

	// This will fail with authentication error in mock mode
	result, err := engine.ProcessMultimodalRecommendation(ctx, req)
	if err != nil {
		fmt.Printf("   🔄 Expected error in mock mode: %s\n", getShortError(err.Error()))
		
		// Test the multimodal analyzer directly
		fmt.Printf("\n📊 Testing multimodal analyzer directly:\n")
		analyzer := engine.GetMultimodalAnalyzer()
		metrics := analyzer.GetMetrics()
		fmt.Printf("   Analyzer initialized: ✅\n")
		fmt.Printf("   Total analyses performed: %d\n", metrics.TotalAnalyses)
	} else {
		fmt.Printf("   ✅ Multimodal recommendation processed successfully\n")
		fmt.Printf("   Enhanced Movies: %d\n", len(result.EnhancedMovies))
		fmt.Printf("   Enhanced Explanations: %d\n", len(result.EnhancedExplanations))
		fmt.Printf("   Multimodal Processing Time: %v\n", result.MultimodalProcessingTime)
		fmt.Printf("   Average Multimodal Score: %.2f\n", result.AverageMultimodalScore)
		fmt.Printf("   Content Analysis Count: %d\n", result.ContentAnalysisCount)
	}

	// Test content database operations
	fmt.Printf("\n💾 Testing Content Database:\n")
	contentDB := engine.GetContentDB()
	
	// Create and store test content
	testContent := &llm.MultimodalContent{
		ID:         "db_test_content",
		MovieID:    9999,
		MovieTitle: "Database Test Movie",
		Modalities: map[llm.ModalityType]*llm.ContentData{
			llm.ModalityText: {
				Type: llm.ModalityText,
				Data: []byte("Test content for database"),
			},
		},
		ProcessedAt: time.Now(),
	}
	
	contentDB.StoreContent(testContent)
	fmt.Printf("   ✅ Stored content for movie ID %d\n", testContent.MovieID)
	
	// Retrieve content
	retrieved, exists := contentDB.GetContent(testContent.MovieID)
	if exists {
		fmt.Printf("   ✅ Retrieved content: %s\n", retrieved.MovieTitle)
	} else {
		fmt.Printf("   ❌ Failed to retrieve stored content\n")
	}
	
	// List all content
	allContent := contentDB.ListContent()
	fmt.Printf("   📊 Total content entries: %d\n", len(allContent))
}

func testContentDatabase(logger *logrus.Logger) {
	contentDB := llm.NewMultimodalContentDB(logger)

	// Test storing multiple content entries
	testMovies := []struct {
		id    int
		title string
	}{
		{1001, "The Matrix"},
		{1002, "Blade Runner 2049"},
		{1003, "Interstellar"},
		{1004, "Dune"},
	}

	fmt.Printf("Testing content database operations:\n\n")

	// Store content for multiple movies
	for i, movie := range testMovies {
		content := &llm.MultimodalContent{
			ID:         fmt.Sprintf("content_%d", movie.id),
			MovieID:    movie.id,
			MovieTitle: movie.title,
			Modalities: map[llm.ModalityType]*llm.ContentData{
				llm.ModalityImage: {
					Type: llm.ModalityImage,
					URL:  fmt.Sprintf("https://example.com/poster_%d.jpg", movie.id),
				},
				llm.ModalityText: {
					Type: llm.ModalityText,
					Data: []byte(fmt.Sprintf("Synopsis for %s", movie.title)),
				},
			},
			ProcessedAt: time.Now(),
		}

		contentDB.StoreContent(content)
		fmt.Printf("✅ Stored content %d: %s (%d modalities)\n", i+1, movie.title, len(content.Modalities))
	}

	// Test retrieval
	fmt.Printf("\n🔍 Testing content retrieval:\n")
	for _, movie := range testMovies {
		content, exists := contentDB.GetContent(movie.id)
		if exists {
			fmt.Printf("   ✅ Retrieved %s: %d modalities\n", content.MovieTitle, len(content.Modalities))
		} else {
			fmt.Printf("   ❌ Failed to retrieve content for movie ID %d\n", movie.id)
		}
	}

	// Test listing all content
	fmt.Printf("\n📋 Listing all stored content:\n")
	allContent := contentDB.ListContent()
	for i, content := range allContent {
		fmt.Printf("   %d. %s (ID: %d) - %d modalities\n", 
			i+1, content.MovieTitle, content.MovieID, len(content.Modalities))
	}

	fmt.Printf("\n📊 Database Statistics:\n")
	fmt.Printf("   Total entries: %d\n", len(allContent))
	
	// Count modalities
	totalModalities := 0
	modalityCount := make(map[llm.ModalityType]int)
	for _, content := range allContent {
		totalModalities += len(content.Modalities)
		for modalityType := range content.Modalities {
			modalityCount[modalityType]++
		}
	}
	
	fmt.Printf("   Total modalities: %d\n", totalModalities)
	fmt.Printf("   Modality distribution:\n")
	for modalityType, count := range modalityCount {
		fmt.Printf("      %s: %d\n", modalityType, count)
	}
}

func getShortError(fullError string) string {
	if len(fullError) > 120 {
		return fullError[:120] + "..."
	}
	return fullError
}

// Demonstration of multimodal capabilities
func demonstrateMultimodalCapabilities() {
	fmt.Println("\n🎭 Multimodal Analysis Capabilities")
	fmt.Println("===================================")

	capabilities := map[string][]string{
		"🖼️ Image Analysis": {
			"Movie poster visual style analysis",
			"Color palette and mood detection",
			"Object and scene recognition",
			"Cinematic style classification",
			"Composition and aesthetic evaluation",
			"Face detection and character analysis",
		},
		"🎵 Audio Analysis": {
			"Soundtrack genre classification",
			"Energy and valence detection",
			"Tempo and rhythm analysis",
			"Sound effect identification",
			"Voice and music separation",
			"Emotional tone assessment",
		},
		"📝 Text Analysis": {
			"Synopsis sentiment analysis",
			"Key phrase and topic extraction",
			"Emotional content detection",
			"Complexity and readability scoring",
			"Genre hint identification",
			"Thematic element recognition",
		},
		"🎬 Video Analysis": {
			"Scene change detection",
			"Action level assessment",
			"Camera movement analysis",
			"Color grading evaluation",
			"Cut frequency measurement",
			"Visual pacing analysis",
		},
		"🔄 Cross-Modal Synthesis": {
			"Genre consensus across modalities",
			"Mood aggregation and weighting",
			"Quality score calculation",
			"Content warning generation",
			"Age rating prediction",
			"Overall sentiment synthesis",
		},
		"🎯 Recommendation Enhancement": {
			"Multimodal similarity scoring",
			"Visual style matching",
			"Audio preference alignment",
			"Content quality assessment",
			"Explanation enrichment",
			"User experience personalization",
		},
	}

	for category, features := range capabilities {
		fmt.Printf("\n%s:\n", category)
		for _, feature := range features {
			fmt.Printf("   • %s\n", feature)
		}
	}
}

func init() {
	// Run capability demonstration
	go func() {
		time.Sleep(100 * time.Millisecond)
		demonstrateMultimodalCapabilities()
	}()
}