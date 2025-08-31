package orchestration

import (
	"fmt"
	"time"
)

// SimpleWorkflowBuilder provides simplified workflow creation
type SimpleWorkflowBuilder struct {
	workflow *SimpleWorkflow
	steps    []*WorkflowStep
}

func NewSimpleWorkflowBuilder() *SimpleWorkflowBuilder {
	return &SimpleWorkflowBuilder{
		workflow: &SimpleWorkflow{
			ID:        fmt.Sprintf("workflow_%d", time.Now().UnixNano()),
			Steps:     []*WorkflowStep{},
			Status:    WorkflowStatusDraft,
			CreatedAt: time.Now(),
		},
		steps: []*WorkflowStep{},
	}
}

func (b *SimpleWorkflowBuilder) Named(name, description string) *SimpleWorkflowBuilder {
	b.workflow.Name = name
	return b
}

func (b *SimpleWorkflowBuilder) AddStep(name, stepType, agentID string) *SimpleWorkflowBuilder {
	step := &WorkflowStep{
		ID:        fmt.Sprintf("step_%d", time.Now().UnixNano()),
		Name:      name,
		AgentID:   agentID,
		Type:      WorkflowStepType(stepType),
		Status:    StatusPending,
		Input:     make(map[string]interface{}),
		Metadata:  make(map[string]interface{}),
	}
	
	b.steps = append(b.steps, step)
	return b
}

func (b *SimpleWorkflowBuilder) Build() *SimpleWorkflow {
	b.workflow.Steps = b.steps
	return b.workflow
}

// Simplified template functions
func NewCodingWorkflow(agentID string) *SimpleWorkflow {
	return NewSimpleWorkflowBuilder().
		Named("Coding Workflow", "Complete coding task with analysis and implementation").
		AddStep("Analyze Requirements", "analyze", agentID).
		AddStep("Generate Code", "generate", agentID).
		AddStep("Validate Code", "validate", agentID).
		Build()
}

func NewResearchWorkflow(agentID string) *SimpleWorkflow {
	return NewSimpleWorkflowBuilder().
		Named("Research Workflow", "Comprehensive research with analysis and synthesis").
		AddStep("Analyze Topic", "analyze", agentID).
		AddStep("Gather Information", "process", agentID).
		AddStep("Synthesize Findings", "analyze", agentID).
		Build()
}

func NewProblemSolvingWorkflow(agentID string) *SimpleWorkflow {
	return NewSimpleWorkflowBuilder().
		Named("Problem Solving Workflow", "Systematic problem-solving approach").
		AddStep("Define Problem", "analyze", agentID).
		AddStep("Generate Solutions", "generate", agentID).
		AddStep("Evaluate Options", "validate", agentID).
		Build()
}

func NewContentCreationWorkflow(agentID string) *SimpleWorkflow {
	return NewSimpleWorkflowBuilder().
		Named("Content Creation Workflow", "Structured content creation process").
		AddStep("Research Topic", "analyze", agentID).
		AddStep("Create Draft", "generate", agentID).
		AddStep("Review Content", "validate", agentID).
		Build()
}

func NewDataAnalysisWorkflow(agentID string) *SimpleWorkflow {
	return NewSimpleWorkflowBuilder().
		Named("Data Analysis Workflow", "Comprehensive data analysis pipeline").
		AddStep("Explore Data", "analyze", agentID).
		AddStep("Process Data", "process", agentID).
		AddStep("Generate Insights", "generate", agentID).
		Build()
}