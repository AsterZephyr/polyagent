package main

import (
	"context"
	"fmt"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🧠 Testing Intelligent Recommendation Explanation System")
	fmt.Println("=======================================================")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test 1: Explanation Generator
	fmt.Println("\n💡 Test 1: Explanation Generator")
	testExplanationGenerator(logger)

	// Test 2: Different Explanation Types
	fmt.Println("\n🎯 Test 2: Different Explanation Types")
	testExplanationTypes(logger)

	// Test 3: Transparency Levels
	fmt.Println("\n🔍 Test 3: Transparency Levels")
	testTransparencyLevels(logger)

	// Test 4: Explainable Recommendation Engine (Mock Mode)
	fmt.Println("\n🤖 Test 4: Explainable Recommendation Engine")
	testExplainableRecommendationEngine(logger)

	// Test 5: Evidence and Confidence Calculation
	fmt.Println("\n📊 Test 5: Evidence and Confidence Calculation")
	testEvidenceAndConfidence(logger)

	fmt.Println("\n🎉 Explanation System Testing Completed!")
}

func testExplanationGenerator(logger *logrus.Logger) {
	generator := llm.NewExplanationGenerator(logger)
	ctx := context.Background()

	// Create test explanation request
	req := &llm.ExplanationRequest{
		MovieID:    1,
		MovieTitle: "Blade Runner 2049",
		MovieFeatures: map[string]interface{}{
			"genres":    []string{"Sci-Fi", "Drama", "Thriller"},
			"director":  "Denis Villeneuve",
			"rating":    4.5,
			"year":      2017,
		},
		UserProfile: &llm.UserProfile{
			UserID: "test_user",
			RatedMovies: []llm.RatedMovie{
				{MovieID: 100, Title: "Blade Runner", Rating: 5.0, Genres: []string{"Sci-Fi"}},
				{MovieID: 101, Title: "Arrival", Rating: 4.5, Genres: []string{"Sci-Fi", "Drama"}},
			},
			Preferences: map[string]float64{
				"Sci-Fi": 0.9,
				"Drama":  0.7,
			},
		},
		SimilarMovies: []llm.SimilarMovie{
			{MovieID: 100, Title: "Blade Runner", SimilarityScore: 0.95, UserRating: 5.0},
			{MovieID: 102, Title: "Ex Machina", SimilarityScore: 0.82, UserRating: 4.5},
		},
		Context: &llm.RecommendationContext{
			TimeOfDay: "evening",
			Mood:      "contemplative",
		},
		TransparencyLevel: llm.TransparencyDetailed,
	}

	explanation, err := generator.GenerateExplanation(ctx, req)
	if err != nil {
		fmt.Printf("❌ Failed to generate explanation: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated explanation for %s:\n", explanation.MovieTitle)
	fmt.Printf("   Type: %s\n", explanation.ExplanationType)
	fmt.Printf("   Confidence: %.2f\n", explanation.Confidence)
	fmt.Printf("   User Relevance: %.2f\n", explanation.UserRelevance)
	fmt.Printf("   Primary Reason: %s\n", explanation.PrimaryReason)
	fmt.Printf("   Detailed Reasons: %d\n", len(explanation.DetailedReasons))
	fmt.Printf("   Evidence: %d pieces\n", len(explanation.Evidence))
	fmt.Printf("   Personalized Text: %s\n", explanation.PersonalizedText)

	// Display detailed reasons
	if len(explanation.DetailedReasons) > 0 {
		fmt.Printf("\n   📋 Detailed Reasons:\n")
		for i, reason := range explanation.DetailedReasons {
			fmt.Printf("      %d. %s (weight: %.2f)\n", i+1, reason.Description, reason.Weight)
		}
	}

	// Display evidence
	if len(explanation.Evidence) > 0 {
		fmt.Printf("\n   📊 Supporting Evidence:\n")
		for i, evidence := range explanation.Evidence {
			fmt.Printf("      %d. %s: %v (source: %s)\n", i+1, evidence.Description, evidence.Value, evidence.Source)
		}
	}
}

func testExplanationTypes(logger *logrus.Logger) {
	generator := llm.NewExplanationGenerator(logger)
	ctx := context.Background()

	// Test different explanation scenarios
	scenarios := []struct {
		name        string
		request     *llm.ExplanationRequest
		expectedType llm.ExplanationType
	}{
		{
			name: "Similarity-based explanation",
			request: &llm.ExplanationRequest{
				MovieID:    2,
				MovieTitle: "Inception",
				SimilarMovies: []llm.SimilarMovie{
					{MovieID: 200, Title: "The Matrix", SimilarityScore: 0.88, UserRating: 5.0},
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			expectedType: llm.ExplanationSimilarity,
		},
		{
			name: "Collaborative filtering explanation",
			request: &llm.ExplanationRequest{
				MovieID:    3,
				MovieTitle: "The Godfather",
				UserProfile: &llm.UserProfile{
					UserID: "user123",
					RatedMovies: make([]llm.RatedMovie, 15), // Many rated movies
					Preferences: map[string]float64{"Drama": 0.8, "Crime": 0.9},
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			expectedType: llm.ExplanationCollaborative,
		},
		{
			name: "Content-based explanation",
			request: &llm.ExplanationRequest{
				MovieID:    4,
				MovieTitle: "Interstellar",
				MovieFeatures: map[string]interface{}{
					"genres":   []string{"Sci-Fi", "Drama"},
					"director": "Christopher Nolan",
					"rating":   4.6,
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			expectedType: llm.ExplanationContentBased,
		},
		{
			name: "Popularity-based explanation",
			request: &llm.ExplanationRequest{
				MovieID:    5,
				MovieTitle: "Top Gun: Maverick",
				Context: &llm.RecommendationContext{
					Trending: true,
					Popular:  true,
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			expectedType: llm.ExplanationPopularity,
		},
		{
			name: "Contextual explanation",
			request: &llm.ExplanationRequest{
				MovieID:    6,
				MovieTitle: "The Princess Bride",
				Context: &llm.RecommendationContext{
					TimeOfDay: "weekend",
					Mood:      "family",
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			expectedType: llm.ExplanationContextual,
		},
	}

	fmt.Printf("Testing %d explanation type scenarios:\n\n", len(scenarios))

	for i, scenario := range scenarios {
		explanation, err := generator.GenerateExplanation(ctx, scenario.request)
		if err != nil {
			fmt.Printf("❌ Scenario %d (%s) failed: %v\n", i+1, scenario.name, err)
			continue
		}

		typeMatch := "✅"
		if explanation.ExplanationType != scenario.expectedType {
			typeMatch = "⚠️"
		}

		fmt.Printf("%s Scenario %d: %s\n", typeMatch, i+1, scenario.name)
		fmt.Printf("   Expected: %s | Got: %s\n", scenario.expectedType, explanation.ExplanationType)
		fmt.Printf("   Confidence: %.2f | Reasons: %d\n", explanation.Confidence, len(explanation.DetailedReasons))
		fmt.Printf("   Primary: %s\n", explanation.PrimaryReason)
		fmt.Println()
	}
}

func testTransparencyLevels(logger *logrus.Logger) {
	generator := llm.NewExplanationGenerator(logger)
	ctx := context.Background()

	// Base request
	baseReq := &llm.ExplanationRequest{
		MovieID:    10,
		MovieTitle: "Dune",
		MovieFeatures: map[string]interface{}{
			"genres":   []string{"Sci-Fi", "Adventure"},
			"director": "Denis Villeneuve",
			"rating":   4.4,
		},
		UserProfile: &llm.UserProfile{
			UserID:      "transparency_test",
			Preferences: map[string]float64{"Sci-Fi": 0.9},
		},
		SimilarMovies: []llm.SimilarMovie{
			{MovieID: 300, Title: "Blade Runner 2049", SimilarityScore: 0.85},
		},
	}

	transparencyLevels := []llm.TransparencyLevel{
		llm.TransparencyMinimal,
		llm.TransparencyBasic,
		llm.TransparencyDetailed,
		llm.TransparencyTechnical,
	}

	fmt.Printf("Testing transparency levels for movie: %s\n\n", baseReq.MovieTitle)

	for i, level := range transparencyLevels {
		req := *baseReq // Copy the request
		req.TransparencyLevel = level

		explanation, err := generator.GenerateExplanation(ctx, &req)
		if err != nil {
			fmt.Printf("❌ Transparency level %s failed: %v\n", level, err)
			continue
		}

		fmt.Printf("🔍 %d. Transparency Level: %s\n", i+1, level)
		fmt.Printf("   Text Length: %d characters\n", len(explanation.PersonalizedText))
		fmt.Printf("   Reasons Included: %d\n", len(explanation.DetailedReasons))
		fmt.Printf("   Generated Text: %s\n", explanation.PersonalizedText)
		fmt.Println()
	}
}

func testExplainableRecommendationEngine(logger *logrus.Logger) {
	// Create mock configuration
	config := &llm.LLMAdapterConfig{
		LoadBalancing:    true,
		CostOptimization: true,
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

	engine, err := llm.NewExplainableRecommendationEngine(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create explainable recommendation engine: %v\n", err)
		return
	}

	fmt.Printf("✅ Explainable Recommendation Engine created successfully\n")

	ctx := context.Background()

	// Test explainable recommendation requests
	testRequests := []struct {
		name    string
		request *llm.ExplainableRecommendationRequest
	}{
		{
			name: "Sci-Fi recommendation with user profile",
			request: &llm.ExplainableRecommendationRequest{
				UserID:    "user_sci_fi",
				SessionID: "session_001",
				Query:     "I love sci-fi movies like Interstellar, recommend something similar",
				UserProfile: &llm.UserProfile{
					UserID:      "user_sci_fi",
					Preferences: map[string]float64{"Sci-Fi": 0.95, "Drama": 0.7},
				},
				SimilarMovies: []llm.SimilarMovie{
					{MovieID: 400, Title: "Interstellar", SimilarityScore: 1.0, UserRating: 5.0},
				},
				TransparencyLevel: llm.TransparencyDetailed,
			},
		},
		{
			name: "Family movie recommendation",
			request: &llm.ExplainableRecommendationRequest{
				UserID:            "family_user",
				SessionID:         "session_002",
				Query:             "Find me good family movies for tonight",
				TransparencyLevel: llm.TransparencyBasic,
			},
		},
		{
			name: "Action movie exploration",
			request: &llm.ExplainableRecommendationRequest{
				UserID:            "action_fan",
				SessionID:         "session_003",
				Query:             "What are some great action movies I might have missed?",
				TransparencyLevel: llm.TransparencyTechnical,
			},
		},
	}

	for i, test := range testRequests {
		fmt.Printf("\n🎬 Test %d: %s\n", i+1, test.name)
		fmt.Printf("   Query: \"%s\"\n", test.request.Query)

		// This will fail with authentication error in mock mode
		result, err := engine.ProcessExplainableRecommendation(ctx, test.request)
		if err != nil {
			fmt.Printf("   🔄 Expected error in mock mode: %s\n", getShortError(err.Error()))
		} else {
			fmt.Printf("   ✅ Successfully processed explainable recommendation\n")
			fmt.Printf("   Intent: %s (confidence: %.2f)\n", result.Intent.Type, result.Intent.Confidence)
			fmt.Printf("   Movies Recommended: %d\n", len(result.RecommendedMovies))
			fmt.Printf("   Explanations Generated: %d\n", len(result.Explanations))
			fmt.Printf("   Overall Confidence: %.2f\n", result.OverallConfidence)
		}
	}

	// Test detailed explanation generation (this should work even in mock mode)
	fmt.Printf("\n📖 Testing Detailed Explanation Generation:\n")
	
	detailedReq := &llm.DetailedExplanationRequest{
		MovieID:    500,
		MovieTitle: "The Matrix",
		MovieFeatures: map[string]interface{}{
			"genres":    []string{"Action", "Sci-Fi"},
			"director":  "The Wachowskis",
			"rating":    4.7,
			"year":      1999,
		},
		UserProfile: &llm.UserProfile{
			UserID:      "matrix_fan",
			Preferences: map[string]float64{"Sci-Fi": 0.9, "Action": 0.8},
		},
		SimilarMovies: []llm.SimilarMovie{
			{MovieID: 501, Title: "Blade Runner", SimilarityScore: 0.78},
		},
	}

	// The detailed explanation generation part should work
	generator := engine.GetExplanationGenerator()
	explanationReq := &llm.ExplanationRequest{
		MovieID:           detailedReq.MovieID,
		MovieTitle:        detailedReq.MovieTitle,
		MovieFeatures:     detailedReq.MovieFeatures,
		UserProfile:       detailedReq.UserProfile,
		SimilarMovies:     detailedReq.SimilarMovies,
		TransparencyLevel: llm.TransparencyTechnical,
	}

	explanation, err := generator.GenerateExplanation(ctx, explanationReq)
	if err != nil {
		fmt.Printf("❌ Detailed explanation failed: %v\n", err)
	} else {
		fmt.Printf("✅ Generated detailed explanation for %s\n", explanation.MovieTitle)
		fmt.Printf("   Type: %s | Confidence: %.2f\n", explanation.ExplanationType, explanation.Confidence)
		fmt.Printf("   Technical Details: %s\n", explanation.PersonalizedText)
	}
}

func testEvidenceAndConfidence(logger *logrus.Logger) {
	generator := llm.NewExplanationGenerator(logger)
	ctx := context.Background()

	// Test confidence calculation with different scenarios
	scenarios := []struct {
		name        string
		request     *llm.ExplanationRequest
		description string
	}{
		{
			name: "High confidence scenario",
			request: &llm.ExplanationRequest{
				MovieID:    601,
				MovieTitle: "High Confidence Movie",
				MovieFeatures: map[string]interface{}{
					"genres":   []string{"Action", "Thriller"},
					"director": "Christopher Nolan",
					"rating":   4.8,
				},
				UserProfile: &llm.UserProfile{
					Preferences: map[string]float64{"Action": 0.95, "Thriller": 0.9},
				},
				SimilarMovies: []llm.SimilarMovie{
					{MovieID: 602, Title: "Favorite Movie", SimilarityScore: 0.95, UserRating: 5.0},
				},
				TransparencyLevel: llm.TransparencyDetailed,
			},
			description: "Strong preferences, high similarity, quality movie",
		},
		{
			name: "Medium confidence scenario",
			request: &llm.ExplanationRequest{
				MovieID:    603,
				MovieTitle: "Medium Confidence Movie",
				MovieFeatures: map[string]interface{}{
					"genres": []string{"Comedy"},
					"rating": 3.8,
				},
				UserProfile: &llm.UserProfile{
					Preferences: map[string]float64{"Comedy": 0.6},
				},
				TransparencyLevel: llm.TransparencyBasic,
			},
			description: "Moderate preferences, average rating",
		},
		{
			name: "Low confidence scenario",
			request: &llm.ExplanationRequest{
				MovieID:           604,
				MovieTitle:        "Low Confidence Movie",
				MovieFeatures:     map[string]interface{}{"rating": 3.2},
				TransparencyLevel: llm.TransparencyMinimal,
			},
			description: "Limited information, no strong preferences",
		},
	}

	fmt.Printf("Testing confidence calculation across different scenarios:\n\n")

	for i, scenario := range scenarios {
		explanation, err := generator.GenerateExplanation(ctx, scenario.request)
		if err != nil {
			fmt.Printf("❌ Scenario %d failed: %v\n", i+1, err)
			continue
		}

		confidenceLevel := "Low"
		if explanation.Confidence >= 0.7 {
			confidenceLevel = "High"
		} else if explanation.Confidence >= 0.4 {
			confidenceLevel = "Medium"
		}

		fmt.Printf("📊 Scenario %d: %s\n", i+1, scenario.name)
		fmt.Printf("   Description: %s\n", scenario.description)
		fmt.Printf("   Confidence: %.3f (%s)\n", explanation.Confidence, confidenceLevel)
		fmt.Printf("   User Relevance: %.3f\n", explanation.UserRelevance)
		fmt.Printf("   Evidence Count: %d\n", len(explanation.Evidence))
		fmt.Printf("   Reason Count: %d\n", len(explanation.DetailedReasons))
		
		if len(explanation.DetailedReasons) > 0 {
			avgWeight := 0.0
			for _, reason := range explanation.DetailedReasons {
				avgWeight += reason.Weight
			}
			avgWeight /= float64(len(explanation.DetailedReasons))
			fmt.Printf("   Average Reason Weight: %.3f\n", avgWeight)
		}
		fmt.Println()
	}

	// Display metrics
	metrics := generator.GetMetrics()
	fmt.Printf("📈 Explanation Generator Metrics:\n")
	fmt.Printf("   Total Explanations: %d\n", metrics.TotalExplanations)
	fmt.Printf("   Explanations by Type:\n")
	for expType, count := range metrics.ExplanationsByType {
		fmt.Printf("     %s: %d\n", expType, count)
	}
	
	if len(metrics.GenerationTimes) > 0 {
		avgTime := time.Duration(0)
		for _, t := range metrics.GenerationTimes {
			avgTime += t
		}
		avgTime = avgTime / time.Duration(len(metrics.GenerationTimes))
		fmt.Printf("   Average Generation Time: %v\n", avgTime)
	}
}

func getShortError(fullError string) string {
	if len(fullError) > 100 {
		return fullError[:100] + "..."
	}
	return fullError
}

// Demonstration of explanation capabilities
func demonstrateExplanationCapabilities() {
	fmt.Println("\n💡 Explanation System Capabilities")
	fmt.Println("==================================")

	capabilities := map[string][]string{
		"🎯 Explanation Types": {
			"Similarity-based (\"Because you liked The Matrix\")",
			"Collaborative filtering (\"Users like you also enjoyed\")",
			"Content-based (\"Matches your genre preferences\")",
			"Popularity-based (\"Currently trending and highly rated\")",
			"Personalized (\"Fits your viewing patterns\")",
			"Contextual (\"Perfect for your current mood\")",
			"Hybrid (\"Combines multiple recommendation factors\")",
		},
		"🔍 Transparency Levels": {
			"Minimal: Just the main reason",
			"Basic: Main reason + 1-2 supporting details",
			"Detailed: Full explanation with multiple reasons",
			"Technical: Algorithm details and confidence scores",
		},
		"📊 Evidence Types": {
			"Similarity scores to liked movies",
			"User ratings from similar profiles",
			"Professional critic ratings",
			"Content feature matching",
			"Trending and popularity metrics",
			"Viewing pattern analysis",
		},
		"🧠 Personalization Features": {
			"User preference integration",
			"Viewing history analysis",
			"Contextual mood consideration",
			"Time-based recommendations",
			"Confidence-weighted explanations",
			"Multi-factor reasoning",
		},
		"🔬 Analysis Capabilities": {
			"Confidence score calculation",
			"User relevance assessment",
			"Evidence strength evaluation",
			"Reason weight optimization",
			"Cross-factor correlation",
			"Explanation quality metrics",
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
		demonstrateExplanationCapabilities()
	}()
}