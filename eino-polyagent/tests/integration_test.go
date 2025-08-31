package tests

import (
	"context"
	"testing"
	"time"

	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
	"github.com/sirupsen/logrus"
)

func TestFullSystemIntegration(t *testing.T) {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Setup configuration
	cfg := &config.Config{}

	// Setup model router
	modelConfig := &ai.ModelConfig{
		DefaultModel: "gpt-4",
	}
	modelRouter := ai.NewModelRouter(modelConfig, logger)

	// Setup orchestrator
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	t.Run("Agent Creation", func(t *testing.T) {
		agentConfig := &orchestration.AgentConfig{
			Name:         "TestAgent",
			Type:         orchestration.AgentTypeConversational,
			SystemPrompt: "You are a helpful test assistant",
			Model:        "gpt-4",
			Temperature:  0.7,
			MaxTokens:    1000,
		}

		agentID, err := orchestrator.CreateAgent(agentConfig)
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}

		if agentID == "" {
			t.Fatal("Agent ID should not be empty")
		}

		t.Logf("Created agent with ID: %s", agentID)
	})

	t.Run("Message Processing", func(t *testing.T) {
		// First create an agent
		agentConfig := &orchestration.AgentConfig{
			ID:           "test-agent-msg",
			Name:         "MessageTestAgent",
			Type:         orchestration.AgentTypeConversational,
			SystemPrompt: "You are a helpful test assistant",
			Model:        "gpt-4",
		}

		_, err := orchestrator.CreateAgent(agentConfig)
		if err != nil {
			t.Fatalf("Failed to create agent: %v", err)
		}

		// Test message processing
		ctx := context.Background()
		result, err := orchestrator.ProcessMessage(
			ctx,
			"test-agent-msg",
			"test-session-123",
			"Hello, how are you?",
			"test-user",
		)

		if err != nil {
			t.Fatalf("Failed to process message: %v", err)
		}

		if result.Content == "" {
			t.Fatal("Response content should not be empty")
		}

		if result.AgentID != "test-agent-msg" {
			t.Fatalf("Expected agent ID 'test-agent-msg', got '%s'", result.AgentID)
		}

		if result.SessionID != "test-session-123" {
			t.Fatalf("Expected session ID 'test-session-123', got '%s'", result.SessionID)
		}

		t.Logf("Response: %s", result.Content)
		t.Logf("Latency: %v", result.Latency)
	})

	t.Run("Context Management", func(t *testing.T) {
		contextMgr := orchestration.NewContextManager(logger)

		// Create context
		ctx, err := contextMgr.CreateEnhancedContext("session-ctx-test", "workflow-ctx-test")
		if err != nil {
			t.Fatalf("Failed to create context: %v", err)
		}

		if ctx.SessionID != "session-ctx-test" {
			t.Fatalf("Expected session ID 'session-ctx-test', got '%s'", ctx.SessionID)
		}

		// Add messages
		ctx.AddMessage(orchestration.ConversationEntry{
			Timestamp: time.Now(),
			StepID:    "step1",
			Type:      "question",
			Content:   "What is the weather like?",
		})

		ctx.AddMessage(orchestration.ConversationEntry{
			Timestamp: time.Now(),
			StepID:    "step2",
			Type:      "answer",
			Content:   "I can help you with weather information.",
		})

		// Test intelligent features
		if ctx.ContextSummary == "" {
			t.Fatal("Context summary should be generated")
		}

		if len(ctx.Messages) != 2 {
			t.Fatalf("Expected 2 messages, got %d", len(ctx.Messages))
		}

		// Test context message building
		contextMsg := ctx.BuildContextMessage()
		if contextMsg == "" {
			t.Fatal("Context message should not be empty")
		}

		t.Logf("Context Summary: %s", ctx.ContextSummary)
		t.Logf("Key Insights: %v", ctx.KeyInsights)
		t.Logf("Complexity: %.2f", ctx.Complexity)
	})

	t.Run("Workflow Creation", func(t *testing.T) {
		// Create a coding workflow
		workflow := orchestration.NewCodingWorkflow("test-agent")

		if workflow.Name != "Coding Workflow" {
			t.Fatalf("Expected workflow name 'Coding Workflow', got '%s'", workflow.Name)
		}

		if len(workflow.Steps) == 0 {
			t.Fatal("Workflow should have steps")
		}

		for i, step := range workflow.Steps {
			if step.Name == "" {
				t.Fatalf("Step %d should have a name", i)
			}
			if step.AgentID != "test-agent" {
				t.Fatalf("Step %d should have agent ID 'test-agent', got '%s'", i, step.AgentID)
			}
			t.Logf("Step %d: %s (%s)", i+1, step.Name, step.Type)
		}
	})

	t.Run("Agent Configuration", func(t *testing.T) {
		// Test getting agents
		agents := orchestrator.GetAgents()

		if len(agents) == 0 {
			t.Log("No agents configured (expected if none were created)")
		} else {
			for id, config := range agents {
				if config.Name == "" {
					t.Errorf("Agent %s should have a name", id)
				}
				t.Logf("Agent %s: %s (%s)", id, config.Name, config.Type)
			}
		}
	})

	t.Run("Streaming Response", func(t *testing.T) {
		// Create agent for streaming test
		agentConfig := &orchestration.AgentConfig{
			ID:           "stream-test-agent",
			Name:         "StreamTestAgent",
			Type:         orchestration.AgentTypeConversational,
			SystemPrompt: "You are a streaming test assistant",
			Model:        "gpt-4",
		}

		_, err := orchestrator.CreateAgent(agentConfig)
		if err != nil {
			t.Fatalf("Failed to create streaming agent: %v", err)
		}

		ctx := context.Background()
		stream, err := orchestrator.StreamResponse(
			ctx,
			"stream-test-agent",
			"stream-session",
			"Tell me a short story",
			"test-user",
		)

		if err != nil {
			t.Fatalf("Failed to start streaming: %v", err)
		}

		// Read from stream
		select {
		case content := <-stream:
			if content == "" {
				t.Fatal("Streamed content should not be empty")
			}
			t.Logf("Streamed content: %s", content)
		case <-time.After(5 * time.Second):
			t.Fatal("Streaming timeout")
		}
	})
}

func TestPerformanceBasics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise

	cfg := &config.Config{}
	modelRouter := ai.NewModelRouter(&ai.ModelConfig{}, logger)
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	// Create agent
	agentConfig := &orchestration.AgentConfig{
		ID:           "perf-test-agent",
		Name:         "PerformanceTestAgent",
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a performance test assistant",
		Model:        "gpt-4",
	}

	_, err := orchestrator.CreateAgent(agentConfig)
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Measure processing time
	start := time.Now()
	ctx := context.Background()

	_, err = orchestrator.ProcessMessage(
		ctx,
		"perf-test-agent",
		"perf-session",
		"Quick test message",
		"perf-user",
	)

	if err != nil {
		t.Fatalf("Failed to process message: %v", err)
	}

	duration := time.Since(start)
	t.Logf("Message processing took: %v", duration)

	// Basic performance expectation
	if duration > 1*time.Second {
		t.Logf("Warning: Message processing took longer than expected: %v", duration)
	}
}