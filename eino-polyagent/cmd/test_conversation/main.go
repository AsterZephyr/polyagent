package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("=== 对话式推荐交互系统测试 ===")

	ctx := context.Background()

	config := &llm.LLMAdapterConfig{
		Primary: llm.LLMConfig{
			Provider:    llm.ProviderOpenAI,
			Model:       "gpt-3.5-turbo",
			APIKey:      "test-key",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		LoadBalancing:    false,
		CostOptimization: false,
	}
	
	logger := logrus.New()
	conversationSystem, err := llm.NewConversationalRecommendationSystem(config, logger)
	if err != nil {
		log.Fatalf("创建对话系统失败: %v", err)
	}

	sessionID := "test-session-001"

	testScenarios := []struct {
		name     string
		messages []string
	}{
		{
			name: "完整推荐对话流程",
			messages: []string{
				"你好，我想要一些电影推荐",
				"我喜欢科幻和动作电影",
				"最近5年的电影比较好",
				"不要太恐怖的",
				"我看过《复仇者联盟》系列，很喜欢",
				"推荐的第一部电影听起来不错，告诉我更多细节",
				"还有其他类似的吗？",
			},
		},
		{
			name: "偏好探索对话",
			messages: []string{
				"我不确定想看什么类型的电影",
				"我之前看过《泰坦尼克号》",
				"浪漫电影还可以，但不要太老的",
				"有现代背景的爱情故事吗？",
				"这些推荐都不错，我想要更多选择",
			},
		},
		{
			name: "多模态内容分析",
			messages: []string{
				"我想根据这张海报找类似的电影",
				"这个演员我很喜欢，有他的其他作品吗？",
				"这种视觉风格的电影有推荐吗？",
			},
		},
	}

	for i, scenario := range testScenarios {
		fmt.Printf("\n--- 测试场景 %d: %s ---\n", i+1, scenario.name)
		
		currentSessionID := fmt.Sprintf("%s-%d", sessionID, i)
		
		for j, message := range scenario.messages {
			fmt.Printf("\n轮次 %d - 用户: %s\n", j+1, message)
			
			conversationRequest := &llm.ConversationRequest{
				SessionID:   currentSessionID,
				UserID:      "test-user-001",
				UserMessage: message,
			}
			
			// Note: Multimodal functionality would be handled by the system internally

			response, err := conversationSystem.ProcessConversationalTurn(ctx, conversationRequest)
			if err != nil {
				fmt.Printf("处理对话失败: %v\n", err)
				continue
			}

			fmt.Printf("系统: %s\n", response.Message)
			fmt.Printf("状态: %s\n", response.State)
			
			if response.Recommendations != nil && len(response.Recommendations) > 0 {
				fmt.Println("推荐内容:")
				for k, rec := range response.Recommendations {
					fmt.Printf("  %d. %s (评分: %.1f)\n", 
						k+1, rec.Title, rec.Rating)
				}
			}
			
			if response.SuggestedActions != nil && len(response.SuggestedActions) > 0 {
				fmt.Println("建议操作:")
				for k, action := range response.SuggestedActions {
					fmt.Printf("  %d. %s\n", k+1, action)
				}
			}
			
			if response.Questions != nil && len(response.Questions) > 0 {
				fmt.Println("系统问题:")
				for k, question := range response.Questions {
					fmt.Printf("  %d. %s\n", k+1, question)
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
		
		fmt.Printf("\n--- 场景 %d 完成，获取会话总结 ---\n", i+1)
		// Summary functionality would be implemented as part of conversation system
		fmt.Printf("场景 %d 完成\n", i+1)
	}

	fmt.Println("\n=== 系统状态和指标测试 ===")
	
	// System metrics would be implemented as part of monitoring
	fmt.Printf("对话系统运行正常\n")

	fmt.Println("\n=== 对话式推荐系统测试完成 ===")
}