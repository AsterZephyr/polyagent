package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🧠 Testing User Intent Understanding System")
	fmt.Println("==========================================")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test 1: Basic Intent Analysis
	fmt.Println("\n🎯 Test 1: Basic Intent Analysis")
	testIntentAnalysis(logger)

	// Test 2: Entity Extraction
	fmt.Println("\n🔍 Test 2: Entity Extraction")
	testEntityExtraction(logger)

	// Test 3: Intent-Aware LLM Adapter (Mock Mode)
	fmt.Println("\n🤖 Test 3: Intent-Aware LLM Adapter")
	testIntentAwareLLMAdapter(logger)

	// Test 4: Conversation Management
	fmt.Println("\n💬 Test 4: Conversation Management")
	testConversationManager(logger)

	// Test 5: Intent Pattern Matching
	fmt.Println("\n🎲 Test 5: Intent Pattern Matching")
	testIntentPatterns(logger)

	fmt.Println("\n🎉 Intent Understanding System Testing Completed!")
}

func testIntentAnalysis(logger *logrus.Logger) {
	analyzer := llm.NewIntentAnalyzer(logger)
	ctx := context.Background()

	// Test queries with different intents
	testQueries := []struct {
		query       string
		expectedIntent llm.IntentType
		description string
	}{
		{
			query:       "Can you recommend some good action movies?",
			expectedIntent: llm.IntentRecommendation,
			description: "Basic recommendation request",
		},
		{
			query:       "I'm looking for movies similar to The Matrix",
			expectedIntent: llm.IntentRecommendation,
			description: "Similarity-based recommendation",
		},
		{
			query:       "Find me information about Inception",
			expectedIntent: llm.IntentSearch,
			description: "Movie information search",
		},
		{
			query:       "What's trending in sci-fi movies right now?",
			expectedIntent: llm.IntentExploration,
			description: "Trending exploration query",
		},
		{
			query:       "Compare Avatar vs Titanic",
			expectedIntent: llm.IntentComparison,
			description: "Movie comparison request",
		},
		{
			query:       "I rate The Godfather 5 stars, it was amazing",
			expectedIntent: llm.IntentFeedback,
			description: "Movie rating feedback",
		},
		{
			query:       "Update my preferences - I don't like horror movies",
			expectedIntent: llm.IntentPersonalization,
			description: "Preference update",
		},
		{
			query:       "Why is The Godfather considered a masterpiece?",
			expectedIntent: llm.IntentInformation,
			description: "Analytical information request",
		},
		{
			query:       "Something good to watch tonight",
			expectedIntent: llm.IntentRecommendation,
			description: "Casual recommendation",
		},
	}

	fmt.Printf("Testing %d different query types:\n\n", len(testQueries))

	for i, test := range testQueries {
		// Create mock context
		context := &llm.IntentContext{
			UserID:           "test_user",
			SessionID:        "test_session",
			ConversationTurn: i + 1,
			TimeOfDay:        "evening",
			DayOfWeek:        "Friday",
		}

		intent, err := analyzer.AnalyzeIntent(ctx, test.query, context)
		if err != nil {
			fmt.Printf("❌ Query %d failed: %v\n", i+1, err)
			continue
		}

		// Check if intent matches expectation
		matchIcon := "✅"
		if intent.Type != test.expectedIntent {
			matchIcon = "⚠️"
		}

		fmt.Printf("%s Query %d: %s\n", matchIcon, i+1, test.description)
		fmt.Printf("   Query: \"%s\"\n", test.query)
		fmt.Printf("   Expected: %s | Detected: %s | Confidence: %.2f\n", 
			test.expectedIntent, intent.Type, intent.Confidence)
		
		if len(intent.Entities) > 0 {
			fmt.Printf("   Entities: %v\n", intent.Entities)
		}
		
		if len(intent.Suggestions) > 0 {
			fmt.Printf("   Suggestions: %s\n", intent.Suggestions[0])
		}
		fmt.Println()
	}

	// Display metrics
	metrics := analyzer.GetMetrics()
	fmt.Printf("📊 Intent Analysis Metrics:\n")
	fmt.Printf("   Total Queries: %d\n", metrics.TotalQueries)
	fmt.Printf("   Intent Distribution:\n")
	for intentType, count := range metrics.IntentCounts {
		fmt.Printf("     %s: %d\n", intentType, count)
	}
}

func testEntityExtraction(logger *logrus.Logger) {
	extractor := llm.NewEntityExtractor()

	testQueries := []struct {
		query    string
		expected map[string]interface{}
	}{
		{
			query: "I love action and sci-fi movies from the 1990s",
			expected: map[string]interface{}{
				"genre": []string{"action", "sci-fi"},
				"year":  []int{1990},
				"sentiment": "positive",
			},
		},
		{
			query: "Find me family-friendly comedies for tonight",
			expected: map[string]interface{}{
				"genre": []string{"comedy"},
				"preference": []string{"family-friendly"},
				"mood": "evening",
			},
		},
		{
			query: "I rate \"The Matrix\" 5 stars out of 5",
			expected: map[string]interface{}{
				"movie_title": "The Matrix",
				"rating": 5.0,
				"sentiment": "positive",
			},
		},
		{
			query: "Compare \"Avatar\" vs \"Titanic\" - which is better?",
			expected: map[string]interface{}{
				"movie_titles": []string{"Avatar", "Titanic"},
			},
		},
	}

	fmt.Printf("Testing entity extraction on %d queries:\n\n", len(testQueries))

	for i, test := range testQueries {
		entities, err := extractor.ExtractEntities(test.query)
		if err != nil {
			fmt.Printf("❌ Query %d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("🔍 Query %d: \"%s\"\n", i+1, test.query)
		
		if len(entities) > 0 {
			entitiesJSON, _ := json.MarshalIndent(entities, "   ", "  ")
			fmt.Printf("   Extracted Entities:\n%s\n", entitiesJSON)
		} else {
			fmt.Printf("   No entities extracted\n")
		}
		fmt.Println()
	}
}

func testIntentAwareLLMAdapter(logger *logrus.Logger) {
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

	adapter, err := llm.NewIntentAwareLLMAdapter(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create intent-aware adapter: %v\n", err)
		return
	}

	fmt.Printf("✅ Intent-Aware LLM Adapter created successfully\n")

	ctx := context.Background()
	
	// Test different types of queries
	testQueries := []struct {
		query   string
		userID  string
		description string
	}{
		{
			query:   "I want sci-fi movies similar to Blade Runner",
			userID:  "user123",
			description: "Recommendation with specific preferences",
		},
		{
			query:   "Tell me about the cast of Inception",
			userID:  "user456",
			description: "Information search query",
		},
		{
			query:   "What's popular in comedy movies this year?",
			userID:  "user789",
			description: "Exploration and trend analysis",
		},
	}

	for i, test := range testQueries {
		fmt.Printf("\n🎬 Processing Query %d: %s\n", i+1, test.description)
		fmt.Printf("   User: %s\n", test.userID)
		fmt.Printf("   Query: \"%s\"\n", test.query)

		// This will fail with authentication error in mock mode, but shows structure
		result, err := adapter.ProcessRecommendationWithIntent(ctx, test.query, test.userID, fmt.Sprintf("session_%d", i+1))
		if err != nil {
			fmt.Printf("   🔄 Expected error in mock mode: %s\n", getShortError(err.Error()))
		} else {
			fmt.Printf("   ✅ Query processed successfully\n")
			fmt.Printf("   Intent: %s (confidence: %.2f)\n", result.Intent.Type, result.Intent.Confidence)
			fmt.Printf("   Processing Time: %v\n", result.ProcessingTime)
		}
	}

	// Test intent analyzer directly
	intentAnalyzer := adapter.GetIntentAnalyzer()
	metrics := intentAnalyzer.GetMetrics()
	fmt.Printf("\n📈 Intent Analyzer Metrics:\n")
	fmt.Printf("   Queries Processed: %d\n", metrics.TotalQueries)
	if len(metrics.ProcessingTimes) > 0 {
		avgTime := time.Duration(0)
		for _, t := range metrics.ProcessingTimes {
			avgTime += t
		}
		avgTime = avgTime / time.Duration(len(metrics.ProcessingTimes))
		fmt.Printf("   Average Processing Time: %v\n", avgTime)
	}
}

func testConversationManager(logger *logrus.Logger) {
	manager := llm.NewConversationManager(logger)

	userID := "test_user"
	sessionID := "test_session"

	fmt.Printf("Testing conversation management for User: %s, Session: %s\n\n", userID, sessionID)

	// Create some mock intents and conversations
	mockIntents := []*llm.Intent{
		{
			Type:       llm.IntentRecommendation,
			Confidence: 0.9,
			Entities:   map[string]interface{}{"genre": []string{"action"}},
			RawQuery:   "I want action movies",
			ParsedAt:   time.Now(),
		},
		{
			Type:       llm.IntentFeedback,
			Confidence: 0.8,
			Entities:   map[string]interface{}{"rating": 4.0, "sentiment": "positive"},
			RawQuery:   "I loved that movie, 4 stars",
			ParsedAt:   time.Now(),
		},
		{
			Type:       llm.IntentRecommendation,
			Confidence: 0.85,
			Entities:   map[string]interface{}{"genre": []string{"sci-fi"}, "preference": []string{"recent"}},
			RawQuery:   "Now recommend some recent sci-fi films",
			ParsedAt:   time.Now(),
		},
	}

	// Simulate conversation updates
	for i, intent := range mockIntents {
		userMessage := intent.RawQuery
		assistantMessage := fmt.Sprintf("Sure! Based on your request for %s, I have some great suggestions...", intent.Type)

		manager.UpdateConversation(userID, sessionID, intent, userMessage, assistantMessage)
		fmt.Printf("✅ Turn %d: Updated conversation with %s intent\n", i+1, intent.Type)
	}

	// Get conversation context
	context := manager.GetConversationContext(userID, sessionID)
	fmt.Printf("\n📋 Conversation Context:\n")
	fmt.Printf("   User ID: %s\n", context.UserID)
	fmt.Printf("   Session ID: %s\n", context.SessionID)
	fmt.Printf("   Turn: %d\n", context.ConversationTurn)
	fmt.Printf("   Time of Day: %s\n", context.TimeOfDay)
	fmt.Printf("   Day of Week: %s\n", context.DayOfWeek)
	
	if len(context.PreviousIntents) > 0 {
		fmt.Printf("   Previous Intents: %v\n", context.PreviousIntents)
	}

	// Get conversation details
	conv := manager.GetOrCreateConversation(userID, sessionID)
	fmt.Printf("\n💬 Conversation Details:\n")
	fmt.Printf("   Total Interactions: %d\n", conv.TotalInteractions)
	fmt.Printf("   Intent History Length: %d\n", len(conv.IntentHistory))
	fmt.Printf("   Message History Length: %d\n", len(conv.MessageHistory))
	fmt.Printf("   Last Interaction: %s\n", conv.LastInteraction.Format("15:04:05"))
}

func testIntentPatterns(logger *logrus.Logger) {
	analyzer := llm.NewIntentAnalyzer(logger)
	ctx := context.Background()

	// Test edge cases and pattern matching
	edgeCases := []struct {
		query       string
		description string
	}{
		{
			query:       "What should I watch?",
			description: "Ambiguous recommendation request",
		},
		{
			query:       "I hate horror movies but love comedies",
			description: "Mixed sentiment with preferences",
		},
		{
			query:       "Find movies with Tom Hanks from the 2000s",
			description: "Actor and time period search",
		},
		{
			query:       "Something random, surprise me!",
			description: "Open exploration request",
		},
		{
			query:       "Is Inception better than The Matrix?",
			description: "Implicit comparison question",
		},
		{
			query:       "My kids want to watch something funny tonight",
			description: "Family context with mood",
		},
		{
			query:       "I just finished watching Dune, what next?",
			description: "Continuation recommendation",
		},
		{
			query:       "",
			description: "Empty query",
		},
		{
			query:       "adsflkjasf random nonsense text 12345",
			description: "Nonsensical input",
		},
	}

	fmt.Printf("Testing %d edge cases and pattern variations:\n\n", len(edgeCases))

	for i, test := range edgeCases {
		context := &llm.IntentContext{
			UserID:           "edge_test_user",
			ConversationTurn: i + 1,
		}

		intent, err := analyzer.AnalyzeIntent(ctx, test.query, context)
		if err != nil {
			fmt.Printf("❌ Edge Case %d failed: %v\n", i+1, err)
			continue
		}

		fmt.Printf("🎲 Edge Case %d: %s\n", i+1, test.description)
		fmt.Printf("   Query: \"%s\"\n", test.query)
		fmt.Printf("   Detected Intent: %s (confidence: %.2f)\n", intent.Type, intent.Confidence)
		
		if len(intent.Entities) > 0 {
			fmt.Printf("   Entities: %d extracted\n", len(intent.Entities))
		}
		
		if len(intent.Suggestions) > 0 {
			fmt.Printf("   Has %d suggestions\n", len(intent.Suggestions))
		}
		fmt.Println()
	}
}

func getShortError(fullError string) string {
	if len(fullError) > 80 {
		return fullError[:80] + "..."
	}
	return fullError
}

// Demonstration of intent capabilities
func demonstrateIntentCapabilities() {
	fmt.Println("\n🎯 Intent Understanding Capabilities")
	fmt.Println("====================================")

	capabilities := map[string][]string{
		"🎬 Recommendation Intent": {
			"Detect requests for movie suggestions",
			"Identify similarity-based preferences (\"like The Matrix\")",
			"Recognize mood and context (\"tonight\", \"with family\")",
			"Extract genre and style preferences",
		},
		"🔍 Search Intent": {
			"Identify specific movie or actor searches",
			"Detect information requests (\"tell me about\")",
			"Recognize cast and crew queries",
			"Extract search criteria and filters",
		},
		"🚀 Exploration Intent": {
			"Detect browsing and discovery requests",
			"Identify trending and popularity queries",
			"Recognize open-ended exploration (\"surprise me\")",
			"Extract exploration parameters (genre, decade)",
		},
		"⚖️ Comparison Intent": {
			"Detect movie vs movie comparisons",
			"Identify \"which is better\" questions",
			"Extract comparison criteria",
			"Recognize implicit comparisons",
		},
		"👤 Personalization Intent": {
			"Detect preference updates (\"I hate horror\")",
			"Identify profile management requests",
			"Recognize taste refinement needs",
			"Extract user constraint changes",
		},
		"💬 Feedback Intent": {
			"Detect ratings and reviews",
			"Identify sentiment (positive/negative)",
			"Recognize explicit feedback (\"5 stars\")",
			"Extract opinion and critique information",
		},
		"📚 Information Intent": {
			"Detect analytical questions (\"why is it popular?\")",
			"Identify explanation requests",
			"Recognize educational queries",
			"Extract information focus areas",
		},
		"🧠 Entity Extraction": {
			"Movie titles (quoted and unquoted)",
			"Genres and categories",
			"Years and time periods",
			"Ratings and sentiments",
			"Actor names and crew",
			"Mood and context indicators",
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
		demonstrateIntentCapabilities()
	}()
}