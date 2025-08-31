package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
	"github.com/sirupsen/logrus"
)

// BenchmarkAgentCreation measures agent creation performance
func BenchmarkAgentCreation(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	agentConfig := &orchestration.AgentConfig{
		Name:         "BenchAgent",
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		agentConfig.ID = "" // Reset ID to force generation
		_, err := orchestrator.CreateAgent(agentConfig)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMessageProcessing measures message processing performance
func BenchmarkMessageProcessing(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	agentConfig := &orchestration.AgentConfig{
		ID:           "bench-agent",
		Name:         "BenchAgent",
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
	}

	_, err := orchestrator.CreateAgent(agentConfig)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	message := "Hello, how are you?"
	sessionID := "test-session"
	userID := "test-user"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := orchestrator.ProcessMessage(ctx, "bench-agent", sessionID, message, userID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentProcessing measures concurrent processing performance
func BenchmarkConcurrentProcessing(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	agentConfig := &orchestration.AgentConfig{
		ID:           "concurrent-agent",
		Name:         "ConcurrentAgent", 
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
	}

	_, err := orchestrator.CreateAgent(agentConfig)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	message := "Process this message"

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sessionID := time.Now().Format(time.RFC3339Nano)
			_, err := orchestrator.ProcessMessage(ctx, "concurrent-agent", sessionID, message, "test-user")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkContextManagement measures context operations performance
func BenchmarkContextManagement(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	contextMgr := orchestration.NewContextManager(logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, err := contextMgr.CreateEnhancedContext("session-1", "workflow-1")
		if err != nil {
			b.Fatal(err)
		}
		
		// Add some context operations
		ctx.SetState("key", "value")
		_, _ = ctx.GetState("key")
		
		// Build context message
		_ = ctx.BuildContextMessage()
	}
}

// BenchmarkMemoryUsage measures memory consumption
func BenchmarkMemoryUsage(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)
		
		// Create multiple agents to test memory usage
		for j := 0; j < 10; j++ {
			agentConfig := &orchestration.AgentConfig{
				Name:         "MemoryAgent",
				Type:         orchestration.AgentTypeConversational,
				SystemPrompt: "You are a helpful assistant",
				Model:        "gpt-4",
			}
			_, _ = orchestrator.CreateAgent(agentConfig)
		}
	}
}

// BenchmarkLatency measures end-to-end latency
func BenchmarkLatency(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	
	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	agentConfig := &orchestration.AgentConfig{
		ID:           "latency-agent",
		Name:         "LatencyAgent",
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a helpful assistant",
		Model:        "gpt-4",
	}

	_, err := orchestrator.CreateAgent(agentConfig)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	
	// Measure different message sizes
	messages := []string{
		"Hi",                                              // Short
		"Please help me understand this concept in detail", // Medium
		"Can you provide a comprehensive analysis of the following complex scenario with multiple factors and considerations that need to be evaluated thoroughly?", // Long
	}

	for _, message := range messages {
		b.Run("msg_len_"+string(rune(len(message))), func(b *testing.B) {
			b.ResetTimer()
			
			for i := 0; i < b.N; i++ {
				start := time.Now()
				_, err := orchestrator.ProcessMessage(ctx, "latency-agent", "session", message, "user")
				if err != nil {
					b.Fatal(err)
				}
				duration := time.Since(start)
				b.ReportMetric(float64(duration.Nanoseconds()), "ns/op")
			}
		})
	}
}