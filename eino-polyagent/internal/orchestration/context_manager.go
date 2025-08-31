package orchestration

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ContextManager implements practical context management
type ContextManager struct {
	contexts map[string]*SimpleContext
	logger   *logrus.Logger
	mu       sync.RWMutex
}

// SimpleContext provides practical context management with enhanced intelligence
type SimpleContext struct {
	ID              string                 `json:"id"`
	SessionID       string                 `json:"session_id"`
	WorkflowID      string                 `json:"workflow_id"`
	Messages        []ConversationEntry    `json:"messages"`
	Decisions       []KeyDecision          `json:"decisions"`
	SharedState     map[string]interface{} `json:"shared_state"`
	StepHistory     map[string]interface{} `json:"step_history"`
	TokensUsed      int                    `json:"tokens_used"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	
	// Enhanced intelligence features
	ContextSummary  string                 `json:"context_summary"`
	KeyInsights     []string               `json:"key_insights"`
	NextSteps       []string               `json:"next_steps"`
	Complexity      float64                `json:"complexity"`
	Priority        string                 `json:"priority"`
	
	mu              sync.RWMutex
}

func NewContextManager(logger *logrus.Logger) *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*SimpleContext),
		logger:   logger,
	}
}

func (cm *ContextManager) CreateEnhancedContext(sessionID, workflowID string) (*SimpleContext, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	contextID := fmt.Sprintf("ctx_%d", time.Now().UnixNano())
	
	ctx := &SimpleContext{
		ID:              contextID,
		SessionID:       sessionID,
		WorkflowID:      workflowID,
		Messages:        []ConversationEntry{},
		Decisions:       []KeyDecision{},
		SharedState:     make(map[string]interface{}),
		StepHistory:     make(map[string]interface{}),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		ContextSummary:  "",
		KeyInsights:     []string{},
		NextSteps:       []string{},
		Complexity:      0.0,
		Priority:        "normal",
	}

	cm.contexts[contextID] = ctx
	
	cm.logger.WithFields(logrus.Fields{
		"context_id":  contextID,
		"session_id":  sessionID,
		"workflow_id": workflowID,
	}).Info("Context created")

	return ctx, nil
}

func (cm *ContextManager) GetContext(contextID string) (*SimpleContext, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[contextID]
	if !exists {
		return nil, fmt.Errorf("context not found: %s", contextID)
	}

	return ctx, nil
}

func (cm *ContextManager) AddContextAtom(contextID string, atomType string, content string, metadata map[string]interface{}) error {
	ctx, err := cm.GetContext(contextID)
	if err != nil {
		return err
	}

	entry := ConversationEntry{
		Timestamp: time.Now(),
		StepID:    "atom",
		Type:      string(atomType),
		Content:   content,
		Metadata:  metadata,
	}

	ctx.AddMessage(entry)
	return nil
}

func (cm *ContextManager) UpdateContextState(contextID string, updates map[string]interface{}) error {
	ctx, err := cm.GetContext(contextID)
	if err != nil {
		return err
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	for key, value := range updates {
		ctx.SharedState[key] = value
	}
	ctx.UpdatedAt = time.Now()

	return nil
}

func (ctx *SimpleContext) AddMessage(entry ConversationEntry) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	
	ctx.Messages = append(ctx.Messages, entry)
	ctx.UpdatedAt = time.Now()
	
	// Keep only recent messages to prevent memory bloat
	if len(ctx.Messages) > 100 {
		ctx.Messages = ctx.Messages[len(ctx.Messages)-50:]
	}
	
	// Auto-update intelligence after adding message
	ctx.mu.Unlock() // Unlock temporarily for UpdateIntelligence
	ctx.UpdateIntelligence()
	ctx.mu.Lock() // Re-lock for defer unlock
}

func (ctx *SimpleContext) AddDecision(decision KeyDecision) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	
	ctx.Decisions = append(ctx.Decisions, decision)
	ctx.UpdatedAt = time.Now()
}

func (ctx *SimpleContext) SetState(key string, value interface{}) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	
	ctx.SharedState[key] = value
	ctx.UpdatedAt = time.Now()
}

func (ctx *SimpleContext) GetState(key string) (interface{}, bool) {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	value, exists := ctx.SharedState[key]
	return value, exists
}

func (ctx *SimpleContext) BuildContextMessage() string {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	msg := "WORKFLOW CONTEXT:\n\n"
	
	// Add recent messages
	if len(ctx.Messages) > 0 {
		msg += "CONVERSATION HISTORY:\n"
		start := max(0, len(ctx.Messages)-10)
		for i := start; i < len(ctx.Messages); i++ {
			entry := ctx.Messages[i]
			msg += fmt.Sprintf("- %s (%s): %s\n", entry.StepID, entry.Type, entry.Content)
		}
		msg += "\n"
	}
	
	// Add key decisions
	if len(ctx.Decisions) > 0 {
		msg += "KEY DECISIONS:\n"
		for _, decision := range ctx.Decisions {
			msg += fmt.Sprintf("- %s: %s\n", decision.Decision, decision.Reasoning)
		}
		msg += "\n"
	}
	
	// Add shared state
	if len(ctx.SharedState) > 0 {
		msg += "SHARED STATE:\n"
		for key, value := range ctx.SharedState {
			msg += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}
	
	return msg
}

// Enhanced intelligence methods
func (ctx *SimpleContext) UpdateSummary() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	
	if len(ctx.Messages) == 0 {
		ctx.ContextSummary = "No messages yet"
		return
	}
	
	// Simple AI-like summarization
	recentMessages := ctx.Messages
	if len(recentMessages) > 5 {
		recentMessages = recentMessages[len(recentMessages)-5:]
	}
	
	ctx.ContextSummary = fmt.Sprintf("Recent activity involves %d messages, %d decisions. Latest focus: %s",
		len(ctx.Messages), len(ctx.Decisions), 
		recentMessages[len(recentMessages)-1].Type)
	
	ctx.UpdatedAt = time.Now()
}

func (ctx *SimpleContext) ExtractInsights() []string {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	insights := []string{}
	
	// Analyze decision patterns
	if len(ctx.Decisions) > 0 {
		insights = append(insights, fmt.Sprintf("Made %d key decisions", len(ctx.Decisions)))
	}
	
	// Analyze message patterns
	if len(ctx.Messages) > 10 {
		insights = append(insights, "High activity session with extensive conversation")
	}
	
	// Analyze complexity
	if ctx.Complexity > 0.7 {
		insights = append(insights, "High complexity task requiring careful attention")
	}
	
	return insights
}

func (ctx *SimpleContext) SuggestNextSteps() []string {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	suggestions := []string{}
	
	// Based on recent activity
	if len(ctx.Messages) > 0 {
		lastMessage := ctx.Messages[len(ctx.Messages)-1]
		if lastMessage.Type == "question" {
			suggestions = append(suggestions, "Provide comprehensive answer to user question")
		} else if lastMessage.Type == "task" {
			suggestions = append(suggestions, "Break down task into actionable steps")
		}
	}
	
	// Based on decisions
	if len(ctx.Decisions) == 0 {
		suggestions = append(suggestions, "Identify key decision points")
	}
	
	// Default suggestions
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Continue conversation", "Analyze requirements", "Propose solutions")
	}
	
	return suggestions
}

func (ctx *SimpleContext) CalculateComplexity() float64 {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	
	complexity := 0.0
	
	// Message complexity
	complexity += float64(len(ctx.Messages)) * 0.1
	
	// Decision complexity
	complexity += float64(len(ctx.Decisions)) * 0.2
	
	// State complexity
	complexity += float64(len(ctx.SharedState)) * 0.15
	
	// Normalize to 0-1 range
	if complexity > 1.0 {
		complexity = 1.0
	}
	
	return complexity
}

func (ctx *SimpleContext) UpdateIntelligence() {
	ctx.UpdateSummary()
	ctx.KeyInsights = ctx.ExtractInsights()
	ctx.NextSteps = ctx.SuggestNextSteps()
	ctx.Complexity = ctx.CalculateComplexity()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}