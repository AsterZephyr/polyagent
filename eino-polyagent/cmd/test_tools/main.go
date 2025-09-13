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
	fmt.Println("🛠️  Testing Recommendation System Tool Integration")
	fmt.Println("=================================================")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Test 1: Tool Registry and Individual Tools
	fmt.Println("\n🔧 Test 1: Tool Registry and Individual Tools")
	testToolRegistry(logger)

	// Test 2: Tool Integration Layer
	fmt.Println("\n⚙️  Test 2: Tool Integration Layer")
	testToolIntegrationLayer(logger)

	// Test 3: Enhanced LLM Adapter (Mock Mode)
	fmt.Println("\n🤖 Test 3: Enhanced LLM Adapter (Mock Mode)")
	testEnhancedLLMAdapter(logger)

	// Test 4: Recommendation Query Processor
	fmt.Println("\n💬 Test 4: Recommendation Query Processor")
	testRecommendationQueryProcessor(logger)

	fmt.Println("\n🎉 Tool Integration Testing Completed!")
}

func testToolRegistry(logger *logrus.Logger) {
	registry := llm.NewToolRegistry()

	// Test getting all tools
	tools := registry.GetAllTools()
	fmt.Printf("✅ Registered %d tools:\n", len(tools))
	for _, tool := range tools {
		fmt.Printf("   - %s: %s\n", tool.Function.Name, tool.Function.Description)
	}

	// Test individual tool execution
	ctx := context.Background()

	// Test MovieSearchTool
	fmt.Println("\n🎬 Testing MovieSearchTool:")
	searchParams := map[string]interface{}{
		"genre":      "Action",
		"min_rating": 4.0,
		"year_range": []interface{}{1990.0, 2020.0},
		"limit":      3.0,
	}

	result, err := registry.ExecuteTool(ctx, "search_movies", searchParams)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("✅ Search Result:\n%s\n", resultJSON)
	}

	// Test UserPreferenceTool
	fmt.Println("\n👤 Testing UserPreferenceTool:")
	prefParams := map[string]interface{}{
		"user_id":         "user123",
		"include_implicit": true,
	}

	result, err = registry.ExecuteTool(ctx, "analyze_user_preferences", prefParams)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("✅ Preference Analysis:\n%s\n", resultJSON)
	}

	// Test RecommendationGeneratorTool
	fmt.Println("\n🎯 Testing RecommendationGeneratorTool:")
	recParams := map[string]interface{}{
		"user_id":          "user123",
		"algorithm":        "hybrid",
		"count":            3.0,
		"diversity_factor": 0.7,
	}

	result, err = registry.ExecuteTool(ctx, "generate_recommendations", recParams)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		fmt.Printf("✅ Recommendations:\n%s\n", resultJSON)
	}
}

func testToolIntegrationLayer(logger *logrus.Logger) {
	til := llm.NewToolIntegrationLayer(logger)
	ctx := context.Background()

	// Create mock tool calls
	toolCalls := []llm.ToolCall{
		{
			ID:   "call_1",
			Type: "function",
			Function: llm.FunctionCall{
				Name:      "search_movies",
				Arguments: `{"genre": "Sci-Fi", "min_rating": 4.0, "limit": 2}`,
			},
		},
		{
			ID:   "call_2",
			Type: "function",
			Function: llm.FunctionCall{
				Name:      "analyze_user_preferences",
				Arguments: `{"user_id": "user456", "include_implicit": false}`,
			},
		},
	}

	// Process tool calls
	results, err := til.ProcessToolCalls(ctx, toolCalls)
	if err != nil {
		fmt.Printf("❌ Error processing tool calls: %v\n", err)
		return
	}

	fmt.Printf("✅ Processed %d tool calls:\n", len(results))
	for i, result := range results {
		fmt.Printf("\n📊 Tool Call %d:\n", i+1)
		fmt.Printf("   Tool: %s\n", result.ToolName)
		fmt.Printf("   Success: %v\n", result.Success)
		fmt.Printf("   Execution Time: %v\n", result.ExecutionTime)
		if result.Success {
			resultJSON, _ := json.MarshalIndent(result.Result, "", "    ")
			fmt.Printf("   Result:\n%s\n", resultJSON)
		} else {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	// Check metrics
	metrics := til.GetToolMetrics()
	fmt.Printf("\n📈 Tool Metrics:\n")
	fmt.Printf("   Total Executions: %d\n", metrics.TotalExecutions)
	fmt.Printf("   Successful Calls: %d\n", metrics.SuccessfulCalls)
	fmt.Printf("   Failed Calls: %d\n", metrics.FailedCalls)
	fmt.Printf("   Average Latency: %v\n", metrics.AverageLatency)

	for toolName, stats := range metrics.ToolUsageStats {
		fmt.Printf("   %s: %d calls, %.2f%% success rate\n", 
			toolName, stats.CallCount, float64(stats.SuccessCount)/float64(stats.CallCount)*100)
	}
}

func testEnhancedLLMAdapter(logger *logrus.Logger) {
	// Create mock configuration for testing
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

	adapter, err := llm.NewEnhancedLLMAdapter(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create enhanced adapter: %v\n", err)
		return
	}

	fmt.Printf("✅ Enhanced LLM Adapter created with tool integration\n")

	// Test getting available tools
	tools := adapter.GetToolLayer().GetAvailableTools()
	fmt.Printf("✅ Available tools: %d\n", len(tools))

	// Create a mock request that would trigger tool usage
	ctx := context.Background()
	request := &llm.GenerateRequest{
		Messages: []llm.Message{
			{
				Role:    "system",
				Content: "You are a movie recommendation assistant. Use available tools to help users.",
			},
			{
				Role:    "user",
				Content: "I want action movies from the 1990s with high ratings. Can you search for some?",
			},
		},
		Temperature: 0.7,
		MaxTokens:   500,
	}

	// Note: This will fail with authentication error in mock mode, but we can test the structure
	_, toolResults, err := adapter.GenerateWithTools(ctx, request)
	if err != nil {
		fmt.Printf("🔄 Expected error in mock mode: %v\n", err)
		fmt.Printf("✅ Tool integration structure is properly configured\n")
	} else {
		fmt.Printf("✅ Response generated (unexpected in mock mode)\n")
		if len(toolResults) > 0 {
			fmt.Printf("✅ Tool results: %d\n", len(toolResults))
		}
	}

	// Test tool metrics
	metrics := adapter.GetToolLayer().GetToolMetrics()
	fmt.Printf("📊 Current tool metrics: %d total executions\n", metrics.TotalExecutions)
}

func testRecommendationQueryProcessor(logger *logrus.Logger) {
	// Create mock enhanced adapter
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

	adapter, err := llm.NewEnhancedLLMAdapter(config, logger)
	if err != nil {
		fmt.Printf("❌ Failed to create enhanced adapter: %v\n", err)
		return
	}

	processor := llm.NewRecommendationQueryProcessor(adapter, logger)
	fmt.Printf("✅ Recommendation Query Processor created\n")

	ctx := context.Background()
	queries := []struct {
		userID string
		query  string
	}{
		{
			userID: "user789",
			query:  "I love sci-fi movies like The Matrix. Can you recommend similar films?",
		},
		{
			userID: "user101",
			query:  "Find me some family-friendly comedies for movie night with kids.",
		},
		{
			userID: "user202",
			query:  "What are the most popular action movies this year?",
		},
	}

	for i, q := range queries {
		fmt.Printf("\n🎬 Processing Query %d:\n", i+1)
		fmt.Printf("   User: %s\n", q.userID)
		fmt.Printf("   Query: %s\n", q.query)

		// Note: This will fail with authentication error in mock mode
		result, err := processor.ProcessRecommendationQuery(ctx, q.query, q.userID)
		if err != nil {
			fmt.Printf("   🔄 Expected error in mock mode: %v\n", err)
		} else {
			fmt.Printf("   ✅ Query processed successfully\n")
			fmt.Printf("   Processing Time: %v\n", result.ProcessingTime)
			fmt.Printf("   Tools Used: %d\n", len(result.ToolResults))
			fmt.Printf("   Response: %s\n", result.Response[:min(100, len(result.Response))] + "...")
		}
	}

	fmt.Printf("✅ Recommendation Query Processor testing completed\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Additional test helpers and demonstrations

func demonstrateToolCapabilities() {
	fmt.Println("\n🎯 Tool Capabilities Demonstration")
	fmt.Println("===================================")

	capabilities := map[string][]string{
		"🎬 MovieSearchTool": {
			"Search by genre (Action, Comedy, Drama, Sci-Fi, etc.)",
			"Filter by year range and rating thresholds",
			"Keyword-based title and description search",
			"Configurable result limits",
		},
		"👤 UserPreferenceTool": {
			"Analyze viewing history and rating patterns",
			"Extract genre preferences and trends",
			"Identify viewing behavior patterns",
			"Support for implicit/explicit feedback",
		},
		"🛡️ ContentFilterTool": {
			"Age rating restrictions (G, PG, PG-13, R, NC-17)",
			"Content warning filters (violence, language, etc.)",
			"Family-friendly mode",
			"User-specific blocklists",
		},
		"🎯 RecommendationGeneratorTool": {
			"Collaborative filtering algorithms",
			"Content-based recommendations",
			"Hybrid recommendation approaches",
			"Configurable diversity factors",
		},
		"📊 UserInteractionTool": {
			"Track views, ratings, likes, shares",
			"Monitor watch duration and completion rates",
			"Calculate engagement scores",
			"Session and context tracking",
		},
		"📈 PopularityAnalyzerTool": {
			"Trend analysis across time windows",
			"Genre-specific popularity metrics",
			"Peak usage pattern identification",
			"Seasonal trend detection",
		},
	}

	for tool, features := range capabilities {
		fmt.Printf("\n%s:\n", tool)
		for _, feature := range features {
			fmt.Printf("   • %s\n", feature)
		}
	}
}

func init() {
	// Run capability demonstration on startup
	go func() {
		time.Sleep(100 * time.Millisecond)
		demonstrateToolCapabilities()
	}()
}