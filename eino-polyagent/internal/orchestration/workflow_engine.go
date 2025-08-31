package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// WorkflowStep represents a single step in serial execution
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	AgentID     string                 `json:"agent_id"`
	Type        WorkflowStepType       `json:"type"`
	Condition   *WorkflowCondition     `json:"condition,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Status      StepStatus             `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	ErrorMsg    string                 `json:"error_msg,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type WorkflowStepType string

const (
	StepTypeProcess     WorkflowStepType = "process"
	StepTypeAnalyze     WorkflowStepType = "analyze"
	StepTypeGenerate    WorkflowStepType = "generate"
	StepTypeValidate    WorkflowStepType = "validate"
	StepTypeSummarize   WorkflowStepType = "summarize"
	StepTypeDecision    WorkflowStepType = "decision"
	StepTypeCompress    WorkflowStepType = "compress"
)

type StepStatus string

const (
	StatusPending    StepStatus = "pending"
	StatusRunning    StepStatus = "running"
	StatusCompleted  StepStatus = "completed"
	StatusFailed     StepStatus = "failed"
	StatusSkipped    StepStatus = "skipped"
)

type WorkflowCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// WorkflowContext represents the shared context throughout workflow execution
type WorkflowContext struct {
	ID              string                     `json:"id"`
	SessionID       string                     `json:"session_id"`
	UserID          string                     `json:"user_id"`
	WorkflowID      string                     `json:"workflow_id"`
	ConversationLog []ConversationEntry        `json:"conversation_log"`
	KeyDecisions    []KeyDecision              `json:"key_decisions"`
	SharedState     map[string]interface{}     `json:"shared_state"`
	StepResults     map[string]*WorkflowStep   `json:"step_results"`
	CompressedInfo  *CompressedContext         `json:"compressed_info,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
	TokensUsed      int                        `json:"tokens_used"`
	MaxContextSize  int                        `json:"max_context_size"`
}

// ConversationEntry represents each interaction in the workflow
type ConversationEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	StepID    string                 `json:"step_id"`
	AgentID   string                 `json:"agent_id"`
	Type      string                 `json:"type"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// KeyDecision represents important decisions made during workflow
type KeyDecision struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	StepID      string                 `json:"step_id"`
	AgentID     string                 `json:"agent_id"`
	Decision    string                 `json:"decision"`
	Reasoning   string                 `json:"reasoning"`
	Impact      DecisionImpact         `json:"impact"`
	Alternatives []string              `json:"alternatives,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type DecisionImpact string

const (
	ImpactLow      DecisionImpact = "low"
	ImpactMedium   DecisionImpact = "medium" 
	ImpactHigh     DecisionImpact = "high"
	ImpactCritical DecisionImpact = "critical"
)

// CompressedContext for long-running workflows
type CompressedContext struct {
	Summary          string            `json:"summary"`
	KeyPoints        []string          `json:"key_points"`
	CriticalDecisions []KeyDecision    `json:"critical_decisions"`
	ProjectState     map[string]interface{} `json:"project_state"`
	CompressionRatio float64          `json:"compression_ratio"`
	CompressedAt     time.Time        `json:"compressed_at"`
}

// SerialWorkflow represents a complete workflow definition
type SerialWorkflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []*WorkflowStep        `json:"steps"`
	Context     *WorkflowContext       `json:"context"`
	Config      *WorkflowConfig        `json:"config"`
	Status      WorkflowStatus         `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt time.Time              `json:"completed_at"`
	TotalDuration time.Duration        `json:"total_duration"`
}

type WorkflowStatus string

const (
	WorkflowStatusDraft      WorkflowStatus = "draft"
	WorkflowStatusRunning    WorkflowStatus = "running"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusFailed     WorkflowStatus = "failed"
	WorkflowStatusPaused     WorkflowStatus = "paused"
)

type WorkflowConfig struct {
	MaxContextSize        int  `json:"max_context_size"`
	EnableCompression     bool `json:"enable_compression"`
	CompressionThreshold  int  `json:"compression_threshold"`
	EnableDecisionTracking bool `json:"enable_decision_tracking"`
	FailFast             bool `json:"fail_fast"`
	RetryAttempts        int  `json:"retry_attempts"`
}

// WorkflowEngine manages serial workflow execution with advanced context engineering
type WorkflowEngine struct {
	orchestrator *AgentOrchestrator
	workflows    map[string]*SerialWorkflow
	contexts     map[string]*WorkflowContext
	logger       *logrus.Logger
	mu           sync.RWMutex
}

func NewWorkflowEngine(orchestrator *AgentOrchestrator, logger *logrus.Logger) *WorkflowEngine {
	return &WorkflowEngine{
		orchestrator: orchestrator,
		workflows:    make(map[string]*SerialWorkflow),
		contexts:     make(map[string]*WorkflowContext),
		logger:       logger,
	}
}

// CreateWorkflow creates a new serial workflow
func (we *WorkflowEngine) CreateWorkflow(name, description string, steps []*WorkflowStep, config *WorkflowConfig) (*SerialWorkflow, error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflowID := fmt.Sprintf("workflow_%d", time.Now().UnixNano())
	
	if config == nil {
		config = &WorkflowConfig{
			MaxContextSize:        50000,
			EnableCompression:     true,
			CompressionThreshold:  40000,
			EnableDecisionTracking: true,
			FailFast:             false,
			RetryAttempts:        3,
		}
	}

	workflow := &SerialWorkflow{
		ID:          workflowID,
		Name:        name,
		Description: description,
		Steps:       steps,
		Context:     we.createWorkflowContext(workflowID),
		Config:      config,
		Status:      WorkflowStatusDraft,
		CreatedAt:   time.Now(),
	}

	we.workflows[workflowID] = workflow
	we.contexts[workflowID] = workflow.Context

	we.logger.WithFields(logrus.Fields{
		"workflow_id": workflowID,
		"steps_count": len(steps),
		"name":        name,
	}).Info("Workflow created")

	return workflow, nil
}

func (we *WorkflowEngine) createWorkflowContext(workflowID string) *WorkflowContext {
	return &WorkflowContext{
		ID:              fmt.Sprintf("ctx_%d", time.Now().UnixNano()),
		WorkflowID:      workflowID,
		ConversationLog: []ConversationEntry{},
		KeyDecisions:    []KeyDecision{},
		SharedState:     make(map[string]interface{}),
		StepResults:     make(map[string]*WorkflowStep),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		MaxContextSize:  50000,
	}
}

// ExecuteWorkflow executes a workflow in strict serial order
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string, userID string) error {
	workflow, err := we.getWorkflow(workflowID)
	if err != nil {
		return err
	}

	workflow.Status = WorkflowStatusRunning
	workflow.StartedAt = time.Now()

	defer func() {
		workflow.CompletedAt = time.Now()
		workflow.TotalDuration = workflow.CompletedAt.Sub(workflow.StartedAt)
	}()

	we.logger.WithField("workflow_id", workflowID).Info("Starting workflow execution")

	// Execute steps in strict serial order
	for i, step := range workflow.Steps {
		// Check context size and compress if needed
		if err := we.manageContextSize(workflow); err != nil {
			we.logger.WithError(err).Warn("Context management warning")
		}

		// Check step condition
		if step.Condition != nil && !we.evaluateCondition(step.Condition, workflow.Context) {
			step.Status = StatusSkipped
			we.logger.WithFields(logrus.Fields{
				"step_id": step.ID,
				"step_name": step.Name,
			}).Info("Step skipped due to condition")
			continue
		}

		// Execute step with full context
		if err := we.executeStep(ctx, step, workflow, userID); err != nil {
			step.Status = StatusFailed
			step.ErrorMsg = err.Error()
			
			if workflow.Config.FailFast {
				workflow.Status = WorkflowStatusFailed
				return fmt.Errorf("workflow failed at step %d (%s): %w", i, step.Name, err)
			}
			
			we.logger.WithError(err).WithFields(logrus.Fields{
				"step_id": step.ID,
				"step_name": step.Name,
			}).Error("Step failed, continuing workflow")
			continue
		}

		// Record step completion in context
		workflow.Context.StepResults[step.ID] = step
		workflow.Context.UpdatedAt = time.Now()

		we.logger.WithFields(logrus.Fields{
			"step_id": step.ID,
			"step_name": step.Name,
			"duration": step.Duration,
		}).Info("Step completed successfully")
	}

	workflow.Status = WorkflowStatusCompleted
	we.logger.WithField("workflow_id", workflowID).Info("Workflow completed successfully")
	
	return nil
}

// executeStep executes a single workflow step with full context access
func (we *WorkflowEngine) executeStep(ctx context.Context, step *WorkflowStep, workflow *SerialWorkflow, userID string) error {
	step.Status = StatusRunning
	step.StartTime = time.Now()

	defer func() {
		step.EndTime = time.Now()
		step.Duration = step.EndTime.Sub(step.StartTime)
	}()

	// Build enhanced context for this step
	contextualMessage := we.buildContextualMessage(step, workflow.Context)

	// Add conversation entry
	we.addConversationEntry(workflow.Context, ConversationEntry{
		Timestamp: time.Now(),
		StepID:    step.ID,
		AgentID:   step.AgentID,
		Type:      "step_start",
		Content:   contextualMessage,
		Metadata:  step.Metadata,
	})

	// Execute step through agent orchestrator
	result, err := we.orchestrator.ProcessMessage(ctx, step.AgentID, workflow.Context.SessionID, contextualMessage, userID)
	if err != nil {
		return fmt.Errorf("step execution failed: %w", err)
	}

	// Process result and extract key decisions
	step.Output = map[string]interface{}{
		"response":     result.Content,
		"tokens_used":  result.TokensUsed,
		"cost":         result.Cost,
		"latency":      result.Latency,
		"metadata":     result.Metadata,
	}

	// Extract and record key decisions
	if workflow.Config.EnableDecisionTracking {
		decisions := we.extractKeyDecisions(step, result)
		workflow.Context.KeyDecisions = append(workflow.Context.KeyDecisions, decisions...)
	}

	// Update context with result
	workflow.Context.TokensUsed += result.TokensUsed
	
	// Add conversation entry for completion
	we.addConversationEntry(workflow.Context, ConversationEntry{
		Timestamp: time.Now(),
		StepID:    step.ID,
		AgentID:   step.AgentID,
		Type:      "step_complete",
		Content:   result.Content,
		Metadata:  map[string]interface{}{
			"tokens_used": result.TokensUsed,
			"cost":        result.Cost,
		},
	})

	step.Status = StatusCompleted
	return nil
}

// buildContextualMessage creates rich context for each step
func (we *WorkflowEngine) buildContextualMessage(step *WorkflowStep, context *WorkflowContext) string {
	contextMsg := fmt.Sprintf(`WORKFLOW STEP EXECUTION
Step: %s (%s)
Type: %s

COMPLETE WORKFLOW CONTEXT:
`, step.Name, step.ID, step.Type)

	// Add conversation history
	if len(context.ConversationLog) > 0 {
		contextMsg += "\n=== CONVERSATION HISTORY ===\n"
		for _, entry := range context.ConversationLog {
			contextMsg += fmt.Sprintf("[%s] %s (%s): %s\n", 
				entry.Timestamp.Format("15:04:05"), entry.StepID, entry.Type, entry.Content[:min(200, len(entry.Content))])
		}
	}

	// Add key decisions
	if len(context.KeyDecisions) > 0 {
		contextMsg += "\n=== KEY DECISIONS MADE ===\n"
		for _, decision := range context.KeyDecisions {
			contextMsg += fmt.Sprintf("- [%s] %s: %s (Impact: %s)\n", 
				decision.StepID, decision.Decision, decision.Reasoning, decision.Impact)
		}
	}

	// Add previous step results
	if len(context.StepResults) > 0 {
		contextMsg += "\n=== PREVIOUS STEP OUTPUTS ===\n"
		for stepID, stepResult := range context.StepResults {
			if stepResult.Output != nil {
				if response, ok := stepResult.Output["response"].(string); ok {
					contextMsg += fmt.Sprintf("Step %s: %s\n", stepID, response[:min(300, len(response))])
				}
			}
		}
	}

	// Add current step input
	if step.Input != nil {
		contextMsg += "\n=== CURRENT STEP INPUT ===\n"
		if inputJSON, err := json.MarshalIndent(step.Input, "", "  "); err == nil {
			contextMsg += string(inputJSON)
		}
	}

	// Add compressed context if available
	if context.CompressedInfo != nil {
		contextMsg += fmt.Sprintf("\n=== COMPRESSED CONTEXT ===\n%s\n", context.CompressedInfo.Summary)
		for _, point := range context.CompressedInfo.KeyPoints {
			contextMsg += fmt.Sprintf("• %s\n", point)
		}
	}

	contextMsg += "\nNow execute this step with full awareness of all previous context and decisions.\n"
	return contextMsg
}

// manageContextSize implements context compression when needed
func (we *WorkflowEngine) manageContextSize(workflow *SerialWorkflow) error {
	if !workflow.Config.EnableCompression {
		return nil
	}

	contextSize := we.estimateContextSize(workflow.Context)
	
	if contextSize > workflow.Config.CompressionThreshold {
		we.logger.WithFields(logrus.Fields{
			"context_size": contextSize,
			"threshold":    workflow.Config.CompressionThreshold,
		}).Info("Context size threshold exceeded, compressing")
		
		return we.compressContext(workflow.Context)
	}
	
	return nil
}

func (we *WorkflowEngine) estimateContextSize(context *WorkflowContext) int {
	// Simple estimation based on JSON length
	if data, err := json.Marshal(context); err == nil {
		return len(data)
	}
	return 0
}

// compressContext implements intelligent context compression
func (we *WorkflowEngine) compressContext(context *WorkflowContext) error {
	originalSize := we.estimateContextSize(context)
	
	// Keep only critical decisions and recent conversation entries
	if len(context.ConversationLog) > 20 {
		// Keep first 5, last 10, and compress middle
		recentEntries := append(context.ConversationLog[:5], context.ConversationLog[len(context.ConversationLog)-10:]...)
		context.ConversationLog = recentEntries
	}

	// Keep only high and critical impact decisions
	criticalDecisions := []KeyDecision{}
	for _, decision := range context.KeyDecisions {
		if decision.Impact == ImpactHigh || decision.Impact == ImpactCritical {
			criticalDecisions = append(criticalDecisions, decision)
		}
	}

	// Create compressed summary
	context.CompressedInfo = &CompressedContext{
		Summary:           we.generateContextSummary(context),
		KeyPoints:         we.extractKeyPoints(context),
		CriticalDecisions: criticalDecisions,
		ProjectState:      context.SharedState,
		CompressionRatio:  float64(we.estimateContextSize(context)) / float64(originalSize),
		CompressedAt:      time.Now(),
	}

	context.UpdatedAt = time.Now()
	
	we.logger.WithFields(logrus.Fields{
		"original_size":     originalSize,
		"compressed_size":   we.estimateContextSize(context),
		"compression_ratio": context.CompressedInfo.CompressionRatio,
	}).Info("Context compressed successfully")

	return nil
}

func (we *WorkflowEngine) generateContextSummary(context *WorkflowContext) string {
	summary := "WORKFLOW PROGRESS SUMMARY:\n"
	
	completedSteps := 0
	for _, step := range context.StepResults {
		if step.Status == StatusCompleted {
			completedSteps++
		}
	}
	
	summary += fmt.Sprintf("- Completed %d workflow steps\n", completedSteps)
	summary += fmt.Sprintf("- Made %d key decisions\n", len(context.KeyDecisions))
	summary += fmt.Sprintf("- Total tokens used: %d\n", context.TokensUsed)
	
	if len(context.KeyDecisions) > 0 {
		summary += "\nMost important decisions:\n"
		for _, decision := range context.KeyDecisions[:min(3, len(context.KeyDecisions))] {
			summary += fmt.Sprintf("• %s\n", decision.Decision)
		}
	}
	
	return summary
}

func (we *WorkflowEngine) extractKeyPoints(context *WorkflowContext) []string {
	keyPoints := []string{}
	
	// Extract key points from step results
	for stepID, step := range context.StepResults {
		if step.Status == StatusCompleted && step.Output != nil {
			if response, ok := step.Output["response"].(string); ok && len(response) > 100 {
				keyPoints = append(keyPoints, fmt.Sprintf("%s: %s", stepID, response[:100]+"..."))
			}
		}
	}
	
	return keyPoints[:min(10, len(keyPoints))]
}

// Helper functions
func (we *WorkflowEngine) addConversationEntry(context *WorkflowContext, entry ConversationEntry) {
	context.ConversationLog = append(context.ConversationLog, entry)
	context.UpdatedAt = time.Now()
}

func (we *WorkflowEngine) extractKeyDecisions(step *WorkflowStep, result *ProcessResult) []KeyDecision {
	// Simple decision extraction - in practice, this would use more sophisticated NLP
	decisions := []KeyDecision{}
	
	if result.Content != "" && len(result.Content) > 50 {
		decision := KeyDecision{
			ID:        fmt.Sprintf("decision_%d", time.Now().UnixNano()),
			Timestamp: time.Now(),
			StepID:    step.ID,
			AgentID:   step.AgentID,
			Decision:  fmt.Sprintf("Step %s output generated", step.Name),
			Reasoning: result.Content[:min(200, len(result.Content))],
			Impact:    ImpactMedium,
			Metadata:  result.Metadata,
		}
		decisions = append(decisions, decision)
	}
	
	return decisions
}

func (we *WorkflowEngine) evaluateCondition(condition *WorkflowCondition, context *WorkflowContext) bool {
	// Simple condition evaluation - extend as needed
	switch condition.Operator {
	case "exists":
		_, exists := context.SharedState[condition.Field]
		return exists
	case "equals":
		if value, exists := context.SharedState[condition.Field]; exists {
			return value == condition.Value
		}
		return false
	default:
		return true
	}
}

func (we *WorkflowEngine) getWorkflow(workflowID string) (*SerialWorkflow, error) {
	we.mu.RLock()
	defer we.mu.RUnlock()
	
	workflow, exists := we.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}
	
	return workflow, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}