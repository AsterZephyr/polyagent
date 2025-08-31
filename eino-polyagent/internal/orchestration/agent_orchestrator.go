package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/sirupsen/logrus"
)

type AgentType string

const (
	AgentTypeConversational AgentType = "conversational"
	AgentTypeTaskOriented   AgentType = "task_oriented"
	AgentTypeWorkflowBased  AgentType = "workflow_based"
)

type AgentConfig struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            AgentType              `json:"type"`
	SystemPrompt    string                 `json:"system_prompt"`
	Model           string                 `json:"model"`
	Temperature     float32                `json:"temperature"`
	MaxTokens       int                    `json:"max_tokens"`
	ToolsEnabled    bool                   `json:"tools_enabled"`
	MemoryEnabled   bool                   `json:"memory_enabled"`
	SafetyFilters   []string               `json:"safety_filters"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type ProcessResult struct {
	AgentID      string                 `json:"agent_id"`
	SessionID    string                 `json:"session_id"`
	Content      string                 `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	Latency      time.Duration          `json:"latency"`
	Cost         float64                `json:"cost"`
	TokensUsed   int                    `json:"tokens_used"`
}

type SessionContext struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	AgentID     string            `json:"agent_id"`
	Messages    []MessageEntry    `json:"messages"`
	State       map[string]interface{} `json:"state"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type MessageEntry struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Simple Agent interface
type Agent interface {
	GetID() string
	Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error)
}

type AgentOrchestrator struct {
	agents         map[string]Agent
	sessions       map[string]*SessionContext
	tools          map[string]Tool
	modelRouter    *ai.ModelRouter
	workflowEngine *WorkflowEngine
	contextManager *ContextManager
	config         *config.Config
	logger         *logrus.Logger
	mu             sync.RWMutex
}

type ConversationalAgent struct {
	id          string
	config      *AgentConfig
	modelRouter *ai.ModelRouter
	logger      *logrus.Logger
}

func NewAgentOrchestrator(cfg *config.Config, modelRouter *ai.ModelRouter, logger *logrus.Logger) *AgentOrchestrator {
	orchestrator := &AgentOrchestrator{
		agents:      make(map[string]Agent),
		sessions:    make(map[string]*SessionContext),
		tools:       make(map[string]Tool),
		modelRouter: modelRouter,
		config:      cfg,
		logger:      logger,
	}

	orchestrator.contextManager = NewContextManager(logger)
	orchestrator.workflowEngine = NewWorkflowEngine(orchestrator, logger)
	orchestrator.initializeTools()
	
	return orchestrator
}

func (o *AgentOrchestrator) CreateAgent(config *AgentConfig) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if config.ID == "" {
		config.ID = o.generateAgentID()
	}

	agent := &ConversationalAgent{
		id:          config.ID,
		config:      config,
		modelRouter: o.modelRouter,
		logger:      o.logger,
	}

	o.agents[config.ID] = agent
	
	o.logger.WithFields(logrus.Fields{
		"agent_id": config.ID,
		"type":     config.Type,
		"name":     config.Name,
	}).Info("Agent created successfully")

	return config.ID, nil
}

func (ca *ConversationalAgent) GetID() string {
	return ca.id
}

func (ca *ConversationalAgent) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	start := time.Now()

	// Build request for model router
	routeReq := &ai.RouteRequest{
		Messages: []ai.Message{
			{
				Role:    "user",
				Content: req.Message,
			},
		},
		ModelPreference: ca.config.Model,
		UserID:          req.UserID,
		SessionID:       req.SessionID,
	}

	// Use simple router instead of complex AI routing
	response, err := ca.modelRouter.Route(ctx, routeReq)
	if err != nil {
		return nil, fmt.Errorf("model routing failed: %w", err)
	}

	return &ProcessResponse{
		Content:    response.Response.Content,
		StepID:     req.StepID,
		TokensUsed: response.TokensUsed,
		Cost:       response.Cost,
		Latency:    time.Since(start),
		Metadata:   make(map[string]interface{}),
	}, nil
}

func (o *AgentOrchestrator) ProcessMessage(ctx context.Context, agentID, sessionID, message, userID string) (*ProcessResult, error) {
	agent, err := o.getAgent(agentID)
	if err != nil {
		return nil, err
	}

	req := &ProcessRequest{
		SessionID: sessionID,
		Message:   message,
		UserID:    userID,
	}

	resp, err := agent.Process(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ProcessResult{
		AgentID:     agentID,
		SessionID:   sessionID,
		Content:     resp.Content,
		Metadata:    resp.Metadata,
		Latency:     resp.Latency,
		Cost:        resp.Cost,
		TokensUsed:  resp.TokensUsed,
	}, nil
}

func (o *AgentOrchestrator) StreamResponse(ctx context.Context, agentID, sessionID, message, userID string) (<-chan string, error) {
	// Simple implementation - just return the processed message as a single chunk
	ch := make(chan string, 1)
	
	result, err := o.ProcessMessage(ctx, agentID, sessionID, message, userID)
	if err != nil {
		close(ch)
		return ch, err
	}
	
	ch <- result.Content
	close(ch)
	return ch, nil
}

func (o *AgentOrchestrator) getAgent(agentID string) (Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

func (o *AgentOrchestrator) GetAgents() map[string]*AgentConfig {
	o.mu.RLock()
	defer o.mu.RUnlock()

	configs := make(map[string]*AgentConfig)
	for id, agent := range o.agents {
		if convAgent, ok := agent.(*ConversationalAgent); ok {
			configs[id] = convAgent.config
		}
	}

	return configs
}

func (o *AgentOrchestrator) DeleteAgent(agentID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	_, exists := o.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	delete(o.agents, agentID)
	return nil
}

func (o *AgentOrchestrator) initializeTools() {
	calculatorTool := NewCalculatorTool()
	o.tools["calculator"] = calculatorTool
	
	o.logger.WithField("tools_count", len(o.tools)).Info("Tools initialized")
}

func (o *AgentOrchestrator) generateAgentID() string {
	return fmt.Sprintf("agent_%d", time.Now().UnixNano())
}

