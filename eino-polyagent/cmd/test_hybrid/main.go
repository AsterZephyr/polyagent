package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("=== 本地工具与远程LLM混合架构测试 ===")

	ctx := context.Background()

	// Configure LLM adapter
	llmConfig := &llm.LLMAdapterConfig{
		Primary: llm.LLMConfig{
			Provider:    llm.ProviderOpenAI,
			Model:       "gpt-3.5-turbo",
			APIKey:      "test-key", // Use actual key for real testing
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		LoadBalancing:    false,
		CostOptimization: true,
	}

	// Configure hybrid system
	hybridConfig := &llm.HybridConfig{
		LocalExecutionThreshold:   0.6,
		RemoteExecutionThreshold:  0.7,
		HybridExecutionThreshold:  0.5,
		MaxLocalExecutionTime:     2 * time.Second,
		MaxRemoteExecutionTime:    10 * time.Second,
		LocalToolTimeout:          1 * time.Second,
		MinQualityScore:          0.5,
		QualityVsSpeedTradeoff:   0.7,
		LocalExecutionCost:       0.001,
		RemoteExecutionCost:      0.01,
		CostOptimizationEnabled:  true,
		EnableFallback:           true,
		FallbackStrategy:         "adaptive",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create hybrid recommendation system
	hybridSystem, err := llm.NewHybridRecommendationSystem(hybridConfig, llmConfig, logger)
	if err != nil {
		log.Fatalf("创建混合推荐系统失败: %v", err)
	}

	// Test scenarios
	testScenarios := []struct {
		name        string
		taskType    llm.TaskType
		data        map[string]interface{}
		preferences *llm.ExecutionPreferences
		context     *llm.ExecutionContext
	}{
		{
			name:     "本地协同过滤推荐",
			taskType: llm.TaskMovieRecommendation,
			data: map[string]interface{}{
				"user_id": "user123",
				"top_k":   5,
			},
			preferences: &llm.ExecutionPreferences{
				PreferredMode:      llm.ExecutionLocal,
				MaxLatency:         2 * time.Second,
				RequireExplanation: false,
				CostSensitive:      true,
				QualityThreshold:   0.7,
			},
			context: &llm.ExecutionContext{
				CurrentLoad:     0.3,
				AvailableBudget: 10.0,
				NetworkLatency:  50,
				IsOffline:       false,
				Timestamp:       time.Now(),
				DeviceType:      "mobile",
			},
		},
		{
			name:     "远程LLM意图分析",
			taskType: llm.TaskIntentAnalysis,
			data: map[string]interface{}{
				"message": "我想看一些科幻电影，最好是最近几年的",
				"conversation_context": map[string]interface{}{
					"previous_messages": []string{"你好", "我想要电影推荐"},
					"user_profile":      map[string]interface{}{"preferred_genres": []string{"Action"}},
				},
			},
			preferences: &llm.ExecutionPreferences{
				PreferredMode:      llm.ExecutionRemote,
				MaxLatency:         10 * time.Second,
				RequireExplanation: true,
				CostSensitive:      false,
				QualityThreshold:   0.8,
			},
			context: &llm.ExecutionContext{
				CurrentLoad:     0.7,
				AvailableBudget: 5.0,
				NetworkLatency:  100,
				IsOffline:       false,
				Timestamp:       time.Now(),
				DeviceType:      "desktop",
			},
		},
		{
			name:     "混合模式内容过滤",
			taskType: llm.TaskContentFiltering,
			data: map[string]interface{}{
				"criteria": map[string]interface{}{
					"genres":     []interface{}{"Action", "Sci-Fi"},
					"year_range": map[string]interface{}{"min": 2015.0, "max": 2023.0},
					"min_rating": 4.0,
				},
			},
			preferences: &llm.ExecutionPreferences{
				PreferredMode:      llm.ExecutionHybrid,
				MaxLatency:         5 * time.Second,
				RequireExplanation: true,
				CostSensitive:      true,
				QualityThreshold:   0.8,
			},
			context: &llm.ExecutionContext{
				CurrentLoad:     0.5,
				AvailableBudget: 8.0,
				NetworkLatency:  200,
				IsOffline:       false,
				Timestamp:       time.Now(),
				DeviceType:      "tablet",
			},
		},
		{
			name:     "自动选择相似度计算",
			taskType: llm.TaskSimilarityCalc,
			data: map[string]interface{}{
				"source_item": map[string]interface{}{
					"genres": []interface{}{"Action", "Adventure"},
					"rating": 4.5,
					"year":   2020.0,
				},
				"target_items": []interface{}{
					map[string]interface{}{"genres": []interface{}{"Action", "Thriller"}, "rating": 4.2, "year": 2019.0},
					map[string]interface{}{"genres": []interface{}{"Comedy", "Romance"}, "rating": 3.8, "year": 2021.0},
					map[string]interface{}{"genres": []interface{}{"Action", "Adventure"}, "rating": 4.6, "year": 2018.0},
				},
			},
			preferences: &llm.ExecutionPreferences{
				PreferredMode:      llm.ExecutionAuto,
				MaxLatency:         3 * time.Second,
				RequireExplanation: false,
				CostSensitive:      true,
				QualityThreshold:   0.7,
			},
			context: &llm.ExecutionContext{
				CurrentLoad:     0.4,
				AvailableBudget: 12.0,
				NetworkLatency:  80,
				IsOffline:       false,
				Timestamp:       time.Now(),
				DeviceType:      "mobile",
			},
		},
		{
			name:     "离线模式推荐",
			taskType: llm.TaskMovieRecommendation,
			data: map[string]interface{}{
				"user_id": "user456",
				"top_k":   3,
			},
			preferences: &llm.ExecutionPreferences{
				PreferredMode:      llm.ExecutionAuto,
				MaxLatency:         1 * time.Second,
				RequireExplanation: false,
				CostSensitive:      true,
				QualityThreshold:   0.6,
			},
			context: &llm.ExecutionContext{
				CurrentLoad:     0.2,
				AvailableBudget: 15.0,
				NetworkLatency:  2000, // High latency
				IsOffline:       true,  // Offline mode
				Timestamp:       time.Now(),
				DeviceType:      "mobile",
			},
		},
	}

	// Execute test scenarios
	for i, scenario := range testScenarios {
		fmt.Printf("\n--- 测试场景 %d: %s ---\n", i+1, scenario.name)

		// Create execution request
		request := &llm.HybridExecutionRequest{
			TaskID:      fmt.Sprintf("task-%d", i+1),
			TaskType:    scenario.taskType,
			UserID:      "test-user",
			SessionID:   "test-session",
			Data:        scenario.data,
			Preferences: scenario.preferences,
			Context:     scenario.context,
		}

		// Execute task
		startTime := time.Now()
		response, err := hybridSystem.ExecuteTask(ctx, request)
		executionTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("执行失败: %v\n", err)
			continue
		}

		// Display results
		fmt.Printf("执行模式: %s\n", response.ExecutionMode)
		fmt.Printf("处理时间: %v\n", response.ProcessingTime)
		fmt.Printf("质量评分: %.2f\n", response.QualityScore)
		fmt.Printf("实际执行时间: %v\n", executionTime)

		if len(response.LocalToolsUsed) > 0 {
			fmt.Printf("使用的本地工具: %v\n", response.LocalToolsUsed)
		}

		if response.RemoteCallsMade > 0 {
			fmt.Printf("远程调用次数: %d\n", response.RemoteCallsMade)
		}

		// Display result summary
		if result, ok := response.Result.(map[string]interface{}); ok {
			if combined, exists := result["combined_approach"]; exists && combined.(bool) {
				fmt.Println("结果类型: 混合结果")
				if primary, exists := result["primary_result"]; exists {
					fmt.Printf("主要结果: %s\n", summarizeResult(primary))
				}
			} else {
				fmt.Printf("结果: %s\n", summarizeResult(response.Result))
			}
		} else {
			fmt.Printf("结果: %s\n", summarizeResult(response.Result))
		}

		if response.Explanation != "" {
			fmt.Printf("解释: %s\n", response.Explanation)
		}

		// Performance analysis
		fmt.Printf("性能分析:\n")
		if response.ProcessingTime < scenario.preferences.MaxLatency {
			fmt.Printf("  ✓ 延迟要求满足 (目标: %v, 实际: %v)\n", scenario.preferences.MaxLatency, response.ProcessingTime)
		} else {
			fmt.Printf("  ✗ 延迟要求未满足 (目标: %v, 实际: %v)\n", scenario.preferences.MaxLatency, response.ProcessingTime)
		}

		if response.QualityScore >= scenario.preferences.QualityThreshold {
			fmt.Printf("  ✓ 质量要求满足 (目标: %.2f, 实际: %.2f)\n", scenario.preferences.QualityThreshold, response.QualityScore)
		} else {
			fmt.Printf("  ✗ 质量要求未满足 (目标: %.2f, 实际: %.2f)\n", scenario.preferences.QualityThreshold, response.QualityScore)
		}

		time.Sleep(500 * time.Millisecond) // Brief pause between tests
	}

	// Display system metrics
	fmt.Println("\n=== 系统性能指标 ===")
	metrics := hybridSystem.GetMetrics()
	metricsJSON, _ := json.MarshalIndent(metrics, "", "  ")
	fmt.Printf("混合系统指标:\n%s\n", string(metricsJSON))

	// Performance comparison
	fmt.Println("\n=== 执行模式统计 ===")
	fmt.Printf("本地执行: %d 次\n", metrics.LocalExecutions)
	fmt.Printf("远程执行: %d 次\n", metrics.RemoteExecutions)
	fmt.Printf("混合执行: %d 次\n", metrics.HybridExecutions)
	fmt.Printf("总执行次数: %d 次\n", metrics.TotalExecutions)
	fmt.Printf("平均延迟: %v\n", metrics.AverageLatency)
	fmt.Printf("平均质量评分: %.2f\n", metrics.QualityScore)

	if metrics.TotalExecutions > 0 {
		localRate := float64(metrics.LocalExecutions) / float64(metrics.TotalExecutions) * 100
		remoteRate := float64(metrics.RemoteExecutions) / float64(metrics.TotalExecutions) * 100
		hybridRate := float64(metrics.HybridExecutions) / float64(metrics.TotalExecutions) * 100

		fmt.Printf("\n执行模式分布:\n")
		fmt.Printf("  本地执行: %.1f%%\n", localRate)
		fmt.Printf("  远程执行: %.1f%%\n", remoteRate)
		fmt.Printf("  混合执行: %.1f%%\n", hybridRate)
	}

	fmt.Println("\n=== 混合架构测试完成 ===")
}

// summarizeResult provides a brief summary of the result
func summarizeResult(result interface{}) string {
	switch r := result.(type) {
	case map[string]interface{}:
		if recommendations, exists := r["recommendations"]; exists {
			if recs, ok := recommendations.([]interface{}); ok {
				return fmt.Sprintf("推荐了 %d 部电影", len(recs))
			}
		}
		if intentType, exists := r["intent_type"]; exists {
			return fmt.Sprintf("意图类型: %v", intentType)
		}
		if similarities, exists := r["similarities"]; exists {
			if sims, ok := similarities.([]interface{}); ok {
				return fmt.Sprintf("计算了 %d 个相似度", len(sims))
			}
		}
		if rawResponse, exists := r["raw_response"]; exists {
			return fmt.Sprintf("文本响应: %.50s...", rawResponse)
		}
		return fmt.Sprintf("结构化数据 (%d 个字段)", len(r))
	case []interface{}:
		return fmt.Sprintf("列表数据 (%d 项)", len(r))
	case string:
		if len(r) > 50 {
			return fmt.Sprintf("文本: %.50s...", r)
		}
		return fmt.Sprintf("文本: %s", r)
	default:
		return fmt.Sprintf("数据类型: %T", result)
	}
}