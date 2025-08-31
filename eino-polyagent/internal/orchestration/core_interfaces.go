package orchestration

import (
	"context"
	"fmt"
	"time"
	
	"github.com/sirupsen/logrus"
)

// ProcessRequest represents a unified agent processing request
type ProcessRequest struct {
	SessionID   string                 `json:"session_id"`
	Message     string                 `json:"message"`
	UserID      string                 `json:"user_id"`
	StepID      string                 `json:"step_id,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessResponse represents agent processing response
type ProcessResponse struct {
	Content     string                 `json:"content"`
	StepID      string                 `json:"step_id"`
	TokensUsed  int                    `json:"tokens_used"`
	Cost        float64                `json:"cost"`
	Latency     time.Duration          `json:"latency"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error,omitempty"`
}

// CoreAgent defines the essential agent interface
type CoreAgent interface {
	Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error)
}

// ConfigurableAgent adds configuration capabilities
type ConfigurableAgent interface {
	CoreAgent
	GetConfig() *AgentConfig
	UpdateConfig(config *AgentConfig) error
}

// StreamingAgent adds streaming response capabilities
type StreamingAgent interface {
	CoreAgent
	StreamProcess(ctx context.Context, req *ProcessRequest) (<-chan string, error)
}

// ManagedAgent adds lifecycle management
type ManagedAgent interface {
	CoreAgent
	GetID() string
	Start() error
	Stop() error
	Health() error
}

// SimpleWorkflow represents a streamlined workflow
type SimpleWorkflow struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Steps       []*WorkflowStep  `json:"steps"`
	Context     *SimpleContext   `json:"context"`
	Status      WorkflowStatus   `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	StartedAt   *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time       `json:"completed_at,omitempty"`
}

// WorkflowExecutor handles workflow execution
type WorkflowExecutor interface {
	Execute(ctx context.Context, workflow *SimpleWorkflow) error
	GetStatus(workflowID string) (WorkflowStatus, error)
}

// BasicWorkflowExecutor provides simple workflow execution
type BasicWorkflowExecutor struct {
	orchestrator CoreAgent
	contextMgr   *ContextManager
}

func NewBasicWorkflowExecutor(orchestrator CoreAgent, logger *logrus.Logger) *BasicWorkflowExecutor {
	return &BasicWorkflowExecutor{
		orchestrator: orchestrator,
		contextMgr:   NewContextManager(logger),
	}
}

func (bwe *BasicWorkflowExecutor) Execute(ctx context.Context, workflow *SimpleWorkflow) error {
	startTime := time.Now()
	workflow.StartedAt = &startTime
	workflow.Status = WorkflowStatusRunning

	defer func() {
		endTime := time.Now()
		workflow.CompletedAt = &endTime
		if workflow.Status == WorkflowStatusRunning {
			workflow.Status = WorkflowStatusCompleted
		}
	}()

	// Execute steps sequentially
	for _, step := range workflow.Steps {
		if err := bwe.executeStep(ctx, step, workflow); err != nil {
			workflow.Status = WorkflowStatusFailed
			return err
		}
	}

	return nil
}

func (bwe *BasicWorkflowExecutor) executeStep(ctx context.Context, step *WorkflowStep, workflow *SimpleWorkflow) error {
	step.Status = StatusRunning
	step.StartTime = time.Now()

	defer func() {
		step.EndTime = time.Now()
		step.Duration = step.EndTime.Sub(step.StartTime)
	}()

	// Build context message
	contextMsg := workflow.Context.BuildContextMessage()
	
	// Add step-specific input
	if input, exists := step.Input["message"]; exists {
		contextMsg += fmt.Sprintf("\nCURRENT TASK: %s\n", input)
	}

	req := &ProcessRequest{
		SessionID: workflow.Context.SessionID,
		Message:   contextMsg,
		UserID:    "system",
		StepID:    step.ID,
		Context:   step.Input,
		Metadata:  step.Metadata,
	}

	response, err := bwe.orchestrator.Process(ctx, req)
	if err != nil {
		step.Status = StatusFailed
		step.ErrorMsg = err.Error()
		return err
	}

	// Update step output
	step.Output = map[string]interface{}{
		"content":     response.Content,
		"tokens_used": response.TokensUsed,
		"cost":        response.Cost,
		"latency":     response.Latency,
	}

	// Update workflow context
	workflow.Context.AddMessage(ConversationEntry{
		Timestamp: time.Now(),
		StepID:    step.ID,
		Type:      "step_result",
		Content:   response.Content,
	})

	step.Status = StatusCompleted
	return nil
}

func (bwe *BasicWorkflowExecutor) GetStatus(workflowID string) (WorkflowStatus, error) {
	// In a real implementation, this would check stored workflow status
	return WorkflowStatusCompleted, nil
}