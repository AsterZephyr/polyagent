package orchestration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
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
	Response     *schema.Message        `json:"response"`
	ToolCalls    []schema.ToolCall      `json:"tool_calls,omitempty"`
	Metadata     map[string]interface{} `json:"metadata"`
	Latency      time.Duration          `json:"latency"`
	Cost         float64                `json:"cost"`
	TokensUsed   int                    `json:"tokens_used"`
}

type SessionContext struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	AgentID     string            `json:"agent_id"`
	Messages    []schema.Message  `json:"messages"`
	State       map[string]interface{} `json:"state"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Agent interface {
	GetID() string
	GetConfig() *AgentConfig
	Process(ctx context.Context, sessionID string, message string, userID string) (*ProcessResult, error)
	StreamResponse(ctx context.Context, sessionID string, message string, userID string) (<-chan string, error)
	GetSessionHistory(sessionID string) ([]schema.Message, error)
	UpdateConfig(config *AgentConfig) error
	Stop() error
}

type AgentOrchestrator struct {
	agents      map[string]Agent
	sessions    map[string]*SessionContext
	tools       map[string]tool.Tool
	modelRouter *ai.ModelRouter
	config      *config.Config
	logger      *logrus.Logger
	mu          sync.RWMutex
}

type ConversationalAgent struct {
	id          string
	config      *AgentConfig
	chain       compose.Chain
	modelRouter *ai.ModelRouter
	tools       map[string]tool.Tool
	sessions    map[string]*SessionContext
	logger      *logrus.Logger
	mu          sync.RWMutex
}

func NewAgentOrchestrator(cfg *config.Config, modelRouter *ai.ModelRouter, logger *logrus.Logger) *AgentOrchestrator {
	orchestrator := &AgentOrchestrator{
		agents:      make(map[string]Agent),
		sessions:    make(map[string]*SessionContext),
		tools:       make(map[string]tool.Tool),
		modelRouter: modelRouter,
		config:      cfg,
		logger:      logger,
	}

	orchestrator.initializeTools()
	return orchestrator
}

func (o *AgentOrchestrator) CreateAgent(config *AgentConfig) (string, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if config.ID == "" {
		config.ID = o.generateAgentID()
	}

	var agent Agent
	var err error

	switch config.Type {
	case AgentTypeConversational:
		agent, err = o.createConversationalAgent(config)
	case AgentTypeTaskOriented:
		agent, err = o.createTaskOrientedAgent(config)
	case AgentTypeWorkflowBased:
		agent, err = o.createWorkflowBasedAgent(config)
	default:
		return "", fmt.Errorf("unsupported agent type: %s", config.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to create agent: %w", err)
	}

	o.agents[config.ID] = agent
	
	o.logger.WithFields(logrus.Fields{
		"agent_id": config.ID,
		"type":     config.Type,
		"name":     config.Name,
	}).Info("Agent created successfully")

	return config.ID, nil
}

func (o *AgentOrchestrator) createConversationalAgent(config *AgentConfig) (Agent, error) {
	agent := &ConversationalAgent{
		id:          config.ID,
		config:      config,
		modelRouter: o.modelRouter,
		tools:       make(map[string]tool.Tool),
		sessions:    make(map[string]*SessionContext),
		logger:      o.logger,
	}

	if config.ToolsEnabled {
		for name, tool := range o.tools {
			agent.tools[name] = tool
		}
	}

	chain, err := o.buildConversationalChain(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build chain: %w", err)
	}

	agent.chain = chain
	return agent, nil
}

func (o *AgentOrchestrator) buildConversationalChain(config *AgentConfig) (compose.Chain, error) {
	builder := compose.NewChainBuilder()

	systemMessage := schema.Message{
		Role:    "system",
		Content: config.SystemPrompt,
	}

	if config.ToolsEnabled {
		var tools []tool.Tool
		for _, t := range o.tools {
			tools = append(tools, t)
		}
		
		builder = builder.
			AddNode("system", compose.NewLambdaNode(func(ctx context.Context, input interface{}) (interface{}, error) {
				messages := input.([]schema.Message)
				messages = append([]schema.Message{systemMessage}, messages...)
				return messages, nil
			})).
			AddNode("model_with_tools", compose.NewToolCallingNode(
				o.createModelNode(config.Model),
				tools...,
			))
	} else {
		builder = builder.
			AddNode("system", compose.NewLambdaNode(func(ctx context.Context, input interface{}) (interface{}, error) {
				messages := input.([]schema.Message)
				messages = append([]schema.Message{systemMessage}, messages...)
				return messages, nil
			})).
			AddNode("model", o.createModelNode(config.Model))
	}

	return builder.Build()
}

func (o *AgentOrchestrator) createModelNode(modelName string) compose.Node {
	return compose.NewLambdaNode(func(ctx context.Context, input interface{}) (interface{}, error) {
		messages := input.([]schema.Message)
		
		routeReq := &ai.RouteRequest{
			Messages:        messages,
			ModelPreference: modelName,
			Strategy:        ai.StrategyBalanced,
			UserID:          "system",
			SessionID:       "system",
		}

		response, err := o.modelRouter.Route(ctx, routeReq)
		if err != nil {
			return nil, err
		}

		return response.Response, nil
	})
}

func (o *AgentOrchestrator) createTaskOrientedAgent(config *AgentConfig) (Agent, error) {
	return o.createConversationalAgent(config)
}

func (o *AgentOrchestrator) createWorkflowBasedAgent(config *AgentConfig) (Agent, error) {
	return o.createConversationalAgent(config)
}

func (o *AgentOrchestrator) ProcessMessage(ctx context.Context, agentID, sessionID, message, userID string) (*ProcessResult, error) {
	agent, err := o.getAgent(agentID)
	if err != nil {
		return nil, err
	}

	return agent.Process(ctx, sessionID, message, userID)
}

func (o *AgentOrchestrator) StreamResponse(ctx context.Context, agentID, sessionID, message, userID string) (<-chan string, error) {
	agent, err := o.getAgent(agentID)
	if err != nil {
		return nil, err
	}

	return agent.StreamResponse(ctx, sessionID, message, userID)
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
		configs[id] = agent.GetConfig()
	}

	return configs
}

func (o *AgentOrchestrator) DeleteAgent(agentID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	if err := agent.Stop(); err != nil {
		o.logger.WithField("agent_id", agentID).WithError(err).Error("Failed to stop agent")
	}

	delete(o.agents, agentID)
	return nil
}

func (o *AgentOrchestrator) initializeTools() {
	calculatorTool := tool.NewCalculatorTool()
	o.tools["calculator"] = calculatorTool
	
	o.logger.WithField("tools_count", len(o.tools)).Info("Tools initialized")
}

func (o *AgentOrchestrator) generateAgentID() string {
	return fmt.Sprintf("agent_%d", time.Now().UnixNano())
}

func (ca *ConversationalAgent) GetID() string {
	return ca.id
}

func (ca *ConversationalAgent) GetConfig() *AgentConfig {
	return ca.config
}

func (ca *ConversationalAgent) Process(ctx context.Context, sessionID, message, userID string) (*ProcessResult, error) {
	start := time.Now()

	session, err := ca.getOrCreateSession(sessionID, userID)
	if err != nil {
		return nil, err
	}

	userMessage := schema.Message{
		Role:    "user",
		Content: message,
	}

	session.Messages = append(session.Messages, userMessage)

	result, err := ca.chain.Invoke(ctx, session.Messages)
	if err != nil {
		return nil, fmt.Errorf("chain invocation failed: %w", err)
	}

	response, ok := result.(*schema.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	session.Messages = append(session.Messages, *response)
	session.UpdatedAt = time.Now()

	return &ProcessResult{
		AgentID:    ca.id,
		SessionID:  sessionID,
		Response:   response,
		Metadata:   map[string]interface{}{},
		Latency:    time.Since(start),
		Cost:       0,
		TokensUsed: 0,
	}, nil
}

func (ca *ConversationalAgent) StreamResponse(ctx context.Context, sessionID, message, userID string) (<-chan string, error) {
	responseChan := make(chan string, 100)
	
	go func() {
		defer close(responseChan)
		
		result, err := ca.Process(ctx, sessionID, message, userID)
		if err != nil {
			responseChan <- fmt.Sprintf("Error: %s", err.Error())
			return
		}

		content := result.Response.Content
		for i, char := range content {
			select {
			case <-ctx.Done():
				return
			case responseChan <- string(char):
				if i < len(content)-1 {
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	}()

	return responseChan, nil
}

func (ca *ConversationalAgent) GetSessionHistory(sessionID string) ([]schema.Message, error) {
	ca.mu.RLock()
	defer ca.mu.RUnlock()

	session, exists := ca.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Messages, nil
}

func (ca *ConversationalAgent) UpdateConfig(config *AgentConfig) error {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	ca.config = config
	
	chain, err := ca.buildChain(config)
	if err != nil {
		return fmt.Errorf("failed to rebuild chain: %w", err)
	}
	
	ca.chain = chain
	return nil
}

func (ca *ConversationalAgent) buildChain(config *AgentConfig) (compose.Chain, error) {
	return compose.NewChainBuilder().Build()
}

func (ca *ConversationalAgent) Stop() error {
	return nil
}

func (ca *ConversationalAgent) getOrCreateSession(sessionID, userID string) (*SessionContext, error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	session, exists := ca.sessions[sessionID]
	if !exists {
		session = &SessionContext{
			ID:        sessionID,
			UserID:    userID,
			AgentID:   ca.id,
			Messages:  []schema.Message{},
			State:     make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		ca.sessions[sessionID] = session
	}

	return session, nil
}