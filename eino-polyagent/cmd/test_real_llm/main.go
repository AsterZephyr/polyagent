package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🌐 Testing Real LLM API Calls")
	fmt.Println("===============================")

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Check for environment variables
	openaiKey := os.Getenv("OPENAI_API_KEY")
	claudeKey := os.Getenv("CLAUDE_API_KEY")
	qwenKey := os.Getenv("QWEN_API_KEY")
	k2Key := os.Getenv("K2_API_KEY")

	if openaiKey == "" && claudeKey == "" && qwenKey == "" && k2Key == "" {
		fmt.Println("⚠️  No API keys found in environment variables.")
		fmt.Println("📋 Available test modes:")
		fmt.Println("  1. Set OPENAI_API_KEY for OpenAI testing")
		fmt.Println("  2. Set CLAUDE_API_KEY for Claude testing") 
		fmt.Println("  3. Set QWEN_API_KEY for Qwen testing")
		fmt.Println("  4. Set K2_API_KEY for OpenRouter testing")
		fmt.Println("\n🔄 Running in mock mode...")
	}

	// Create configuration based on available keys
	config := &llm.LLMAdapterConfig{
		LoadBalancing:    true,
		CostOptimization: true,
	}

	// Configure primary provider
	if openaiKey != "" {
		config.Primary = llm.LLMConfig{
			Provider:    llm.ProviderOpenAI,
			Model:       "gpt-4o-mini", // Use cheaper model for testing
			APIKey:      openaiKey,
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		}
		fmt.Printf("✅ OpenAI configured as primary (model: %s)\n", config.Primary.Model)
	} else {
		// Mock OpenAI config
		config.Primary = llm.LLMConfig{
			Provider:    llm.ProviderOpenAI,
			Model:       "gpt-4o-mini",
			APIKey:      "mock-openai-key",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		}
		fmt.Printf("🔄 OpenAI configured as primary (mock mode)\n")
	}

	// Configure fallback providers
	if claudeKey != "" {
		config.Fallback = append(config.Fallback, llm.LLMConfig{
			Provider:    llm.ProviderClaude,
			Model:       "claude-3-5-haiku-20241022", // Use cheaper Claude model
			APIKey:      claudeKey,
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		})
		fmt.Printf("✅ Claude configured as fallback\n")
	} else if config.Primary.Provider != llm.ProviderClaude {
		config.Fallback = append(config.Fallback, llm.LLMConfig{
			Provider:    llm.ProviderClaude,
			Model:       "claude-3-5-haiku-20241022",
			APIKey:      "mock-claude-key",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		})
		fmt.Printf("🔄 Claude configured as fallback (mock mode)\n")
	}

	// Configure budget provider
	if qwenKey != "" {
		config.Budget = &llm.LLMConfig{
			Provider:    llm.ProviderQwen,
			Model:       "qwen2.5-7b-instruct", // Free tier model
			APIKey:      qwenKey,
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		}
		fmt.Printf("✅ Qwen configured as budget option\n")
	}

	// Create unified adapter
	adapter, err := llm.NewUnifiedLLMAdapter(config, logger)
	if err != nil {
		log.Fatalf("Failed to create LLM adapter: %v", err)
	}

	fmt.Printf("\n🚀 LLM Adapter initialized with %d providers\n", len(adapter.GetAvailableProviders()))

	// Test recommendation-specific queries
	testQueries := []struct {
		name    string
		message string
	}{
		{
			name:    "Movie Recommendation",
			message: "I love action movies and sci-fi films. Can you recommend 3 movies I might enjoy?",
		},
		{
			name:    "User Intent Analysis",
			message: "I'm looking for something to watch tonight with my family. We have kids aged 8 and 12.",
		},
		{
			name:    "Preference Learning",
			message: "I just watched The Matrix and absolutely loved it. What similar movies would you suggest?",
		},
		{
			name:    "Content Analysis",
			message: "Analyze the key themes and elements that make The Godfather a classic film.",
		},
	}

	ctx := context.Background()

	for i, query := range testQueries {
		fmt.Printf("\n🎬 Test %d: %s\n", i+1, query.name)
		fmt.Printf("📝 Query: %s\n", query.message)

		request := &llm.GenerateRequest{
			Messages: []llm.Message{
				{
					Role:    "system",
					Content: "You are an expert movie recommendation assistant. Provide helpful, personalized movie suggestions based on user preferences.",
				},
				{
					Role:    "user",
					Content: query.message,
				},
			},
			Temperature: 0.7,
			MaxTokens:   200,
		}

		startTime := time.Now()
		response, err := adapter.Generate(ctx, request)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			continue
		}

		fmt.Printf("✅ Success (took %v):\n", duration)
		fmt.Printf("📤 Model: %s\n", response.Model)
		fmt.Printf("💬 Response: %s\n", response.Choices[0].Message.Content)
		fmt.Printf("📊 Tokens: %d total (%d prompt + %d completion)\n",
			response.Usage.TotalTokens,
			response.Usage.PromptTokens,
			response.Usage.CompletionTokens)

		// Brief pause between requests
		time.Sleep(1 * time.Second)
	}

	// Test tool calling for recommendation system
	fmt.Printf("\n🛠️  Testing Tool Calling for Recommendations\n")

	toolRequest := &llm.GenerateRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: "Find me some good action movies from the last 5 years with high ratings.",
			},
		},
		Tools: []llm.Tool{
			{
				Type: "function",
				Function: llm.ToolFunction{
					Name:        "search_movies",
					Description: "Search for movies based on genre, year range, and rating criteria",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"genre": map[string]interface{}{
								"type":        "string",
								"description": "Movie genre (action, comedy, drama, etc.)",
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
								"description": "Minimum rating score (0-10)",
							},
						},
						"required": []string{"genre"},
					},
				},
			},
		},
	}

	toolResponse, err := adapter.Generate(ctx, toolRequest)
	if err != nil {
		fmt.Printf("❌ Tool calling failed: %v\n", err)
	} else {
		fmt.Printf("✅ Tool calling successful:\n")
		fmt.Printf("📤 Model: %s\n", toolResponse.Model)
		fmt.Printf("💬 Response: %s\n", toolResponse.Choices[0].Message.Content)
		
		if len(toolResponse.Choices[0].Message.ToolCalls) > 0 {
			toolCall := toolResponse.Choices[0].Message.ToolCalls[0]
			fmt.Printf("🔧 Tool Called: %s\n", toolCall.Function.Name)
			fmt.Printf("📋 Arguments: %s\n", toolCall.Function.Arguments)
		}
	}

	// Check provider health and performance
	fmt.Printf("\n📊 Provider Health Check\n")
	healthStatus := adapter.GetProviderStatus(ctx)
	for provider, status := range healthStatus {
		statusIcon := "✅"
		if !status.Available {
			statusIcon = "❌"
		}

		fmt.Printf("   %s %s:\n", statusIcon, provider)
		fmt.Printf("     Available: %v\n", status.Available)
		fmt.Printf("     Latency: %v\n", status.Latency)
		fmt.Printf("     Error Rate: %.2f%%\n", status.ErrorRate*100)
		if status.LastError != "" {
			fmt.Printf("     Last Error: %s\n", status.LastError)
		}
	}

	// Show adapter metrics
	fmt.Printf("\n📈 Adapter Metrics:\n")
	metrics := adapter.GetMetrics()
	fmt.Printf("   Total Requests: %d\n", metrics.TotalRequests)
	fmt.Printf("   Average Latency: %v\n", metrics.AverageLatency)
	fmt.Printf("   Success Rate: %.2f%%\n", metrics.SuccessRate*100)
	fmt.Printf("   Cost Estimate: $%.4f\n", metrics.CostEstimate)

	fmt.Printf("\n🎉 Real LLM API Testing Completed!\n")
	fmt.Printf("\n💡 Next Steps:\n")
	fmt.Printf("   • Set API keys to test real endpoints\n")
	fmt.Printf("   • Integrate with recommendation system\n")
	fmt.Printf("   • Add Eino memory framework\n")
	fmt.Printf("   • Implement tool calling handlers\n")
}