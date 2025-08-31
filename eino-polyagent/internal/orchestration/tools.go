package orchestration

import (
	"context"
)

// Tool represents a simple tool interface
type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// CalculatorTool provides basic calculations
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Performs basic arithmetic calculations"
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Simple mock implementation
	return "calculation result", nil
}