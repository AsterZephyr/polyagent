package orchestration

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ParameterSchema defines parameter validation rules
type ParameterSchema struct {
	Type        string      `json:"type"`        // "string", "number", "boolean", "array", "object"
	Required    bool        `json:"required"`    // Whether parameter is required
	Default     interface{} `json:"default"`     // Default value if not provided
	Description string      `json:"description"` // Parameter description
	Pattern     string      `json:"pattern"`     // Regex pattern for validation (strings only)
	Min         *float64    `json:"min"`         // Minimum value (numbers only)
	Max         *float64    `json:"max"`         // Maximum value (numbers only)
	Enum        []string    `json:"enum"`        // Allowed values
}

// ToolDefinition defines tool schema and metadata
type ToolDefinition struct {
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Parameters  map[string]ParameterSchema `json:"parameters"`
	Fallback    string                     `json:"fallback"` // Fallback strategy: "retry", "mock", "error"
	Timeout     time.Duration              `json:"timeout"`
	MaxRetries  int                        `json:"max_retries"`
}

// ToolResult encapsulates tool execution result with metadata
type ToolResult struct {
	Success   bool          `json:"success"`
	Result    interface{}   `json:"result"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration"`
	Retries   int           `json:"retries"`
	Timestamp time.Time     `json:"timestamp"`
}

// Tool represents enhanced tool interface with validation and error handling
type Tool interface {
	Name() string
	Description() string
	GetDefinition() *ToolDefinition
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
	Validate(args map[string]interface{}) error
	GetFallback(args map[string]interface{}) (interface{}, error)
}

// ToolRegistry manages tool registration and execution with enhanced error handling
type ToolRegistry struct {
	tools  map[string]Tool
	mutex  sync.RWMutex
	logger *logrus.Logger
	stats  map[string]*ToolStats
}

// ToolStats tracks tool performance and reliability metrics
type ToolStats struct {
	TotalCalls    int64         `json:"total_calls"`
	SuccessCount  int64         `json:"success_count"`
	ErrorCount    int64         `json:"error_count"`
	AvgDuration   time.Duration `json:"avg_duration"`
	LastUsed      time.Time     `json:"last_used"`
	LastError     string        `json:"last_error,omitempty"`
	SuccessRate   float64       `json:"success_rate"`
	mutex         sync.RWMutex
}

// NewToolRegistry creates a new tool registry with enhanced capabilities
func NewToolRegistry(logger *logrus.Logger) *ToolRegistry {
	return &ToolRegistry{
		tools:  make(map[string]Tool),
		logger: logger,
		stats:  make(map[string]*ToolStats),
	}
}

// RegisterTool registers a tool with validation and error handling
func (tr *ToolRegistry) RegisterTool(tool Tool) error {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	
	name := tool.Name()
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	
	// Initialize stats for the tool
	tr.stats[name] = &ToolStats{
		LastUsed: time.Now(),
	}
	
	tr.tools[name] = tool
	tr.logger.WithFields(logrus.Fields{
		"tool_name":   name,
		"description": tool.Description(),
	}).Info("Tool registered successfully")
	
	return nil
}

// ExecuteTool executes a tool with full error handling, retries, and fallback
func (tr *ToolRegistry) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) *ToolResult {
	start := time.Now()
	
	tr.mutex.RLock()
	tool, exists := tr.tools[toolName]
	tr.mutex.RUnlock()
	
	if !exists {
		return &ToolResult{
			Success:   false,
			Error:     fmt.Sprintf("tool '%s' not found", toolName),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}
	
	// Update stats
	tr.updateStats(toolName, func(stats *ToolStats) {
		stats.TotalCalls++
		stats.LastUsed = time.Now()
	})
	
	// Validate parameters
	if err := tool.Validate(args); err != nil {
		tr.updateStats(toolName, func(stats *ToolStats) {
			stats.ErrorCount++
			stats.LastError = fmt.Sprintf("validation error: %s", err.Error())
		})
		
		return &ToolResult{
			Success:   false,
			Error:     fmt.Sprintf("parameter validation failed: %s", err.Error()),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}
	}
	
	// Execute with retry logic
	definition := tool.GetDefinition()
	maxRetries := definition.MaxRetries
	if maxRetries == 0 {
		maxRetries = 1 // At least one attempt
	}
	
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create timeout context
		execCtx := ctx
		if definition.Timeout > 0 {
			var cancel context.CancelFunc
			execCtx, cancel = context.WithTimeout(ctx, definition.Timeout)
			defer cancel()
		}
		
		result, err := tool.Execute(execCtx, args)
		if err == nil {
			// Success
			tr.updateStats(toolName, func(stats *ToolStats) {
				stats.SuccessCount++
				duration := time.Since(start)
				stats.AvgDuration = (stats.AvgDuration + duration) / 2
				stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalCalls)
			})
			
			return &ToolResult{
				Success:   true,
				Result:    result,
				Duration:  time.Since(start),
				Retries:   attempt,
				Timestamp: time.Now(),
			}
		}
		
		lastErr = err
		tr.logger.WithFields(logrus.Fields{
			"tool":    toolName,
			"attempt": attempt + 1,
			"error":   err.Error(),
		}).Warn("Tool execution attempt failed")
		
		// Wait before retry (except for last attempt)
		if attempt < maxRetries-1 {
			time.Sleep(time.Millisecond * 100 * time.Duration(attempt+1)) // Exponential backoff
		}
	}
	
	// All retries failed, try fallback
	fallbackResult := tr.handleFallback(tool, args, lastErr)
	
	tr.updateStats(toolName, func(stats *ToolStats) {
		stats.ErrorCount++
		stats.LastError = lastErr.Error()
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalCalls)
	})
	
	return &ToolResult{
		Success:   fallbackResult != nil,
		Result:    fallbackResult,
		Error:     lastErr.Error(),
		Duration:  time.Since(start),
		Retries:   maxRetries - 1,
		Timestamp: time.Now(),
	}
}

// handleFallback applies fallback strategy when tool execution fails
func (tr *ToolRegistry) handleFallback(tool Tool, args map[string]interface{}, originalErr error) interface{} {
	definition := tool.GetDefinition()
	
	switch definition.Fallback {
	case "mock":
		result, err := tool.GetFallback(args)
		if err == nil {
			tr.logger.WithField("tool", tool.Name()).Info("Using fallback mock response")
			return result
		}
		tr.logger.WithError(err).WithField("tool", tool.Name()).Error("Fallback mock failed")
		
	case "retry":
		// Retry logic is already handled in ExecuteTool
		
	case "error":
	default:
		// Return error - no fallback
	}
	
	return nil
}

// updateStats safely updates tool statistics
func (tr *ToolRegistry) updateStats(toolName string, updateFn func(*ToolStats)) {
	tr.mutex.Lock()
	defer tr.mutex.Unlock()
	
	if stats, exists := tr.stats[toolName]; exists {
		stats.mutex.Lock()
		updateFn(stats)
		stats.mutex.Unlock()
	}
}

// GetToolStats returns performance statistics for a tool
func (tr *ToolRegistry) GetToolStats(toolName string) (*ToolStats, bool) {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	
	if stats, exists := tr.stats[toolName]; exists {
		stats.mutex.RLock()
		defer stats.mutex.RUnlock()
		
		// Return copy to avoid race conditions
		statsCopy := *stats
		return &statsCopy, true
	}
	
	return nil, false
}

// ListTools returns all registered tools with their definitions
func (tr *ToolRegistry) ListTools() map[string]*ToolDefinition {
	tr.mutex.RLock()
	defer tr.mutex.RUnlock()
	
	result := make(map[string]*ToolDefinition)
	for name, tool := range tr.tools {
		result[name] = tool.GetDefinition()
	}
	
	return result
}

// CalculatorTool provides basic arithmetic calculations with enhanced validation
type CalculatorTool struct{}

func NewCalculatorTool() *CalculatorTool {
	return &CalculatorTool{}
}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Performs basic arithmetic calculations (add, subtract, multiply, divide)"
}

func (t *CalculatorTool) GetDefinition() *ToolDefinition {
	return &ToolDefinition{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: map[string]ParameterSchema{
			"operation": {
				Type:        "string",
				Required:    true,
				Description: "Arithmetic operation to perform",
				Enum:        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": {
				Type:        "number",
				Required:    true,
				Description: "First operand",
			},
			"b": {
				Type:        "number",
				Required:    true,
				Description: "Second operand",
			},
		},
		Fallback:   "mock",
		Timeout:    time.Second * 5,
		MaxRetries: 2,
	}
}

func (t *CalculatorTool) Validate(args map[string]interface{}) error {
	definition := t.GetDefinition()
	
	for paramName, schema := range definition.Parameters {
		value, exists := args[paramName]
		
		// Check required parameters
		if schema.Required && !exists {
			return fmt.Errorf("required parameter '%s' is missing", paramName)
		}
		
		if !exists {
			// Use default value if provided
			if schema.Default != nil {
				args[paramName] = schema.Default
			}
			continue
		}
		
		// Validate type
		if err := t.validateParameterType(paramName, value, schema); err != nil {
			return err
		}
		
		// Validate enum values
		if len(schema.Enum) > 0 {
			valueStr := fmt.Sprintf("%v", value)
			valid := false
			for _, enumValue := range schema.Enum {
				if enumValue == valueStr {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("parameter '%s' must be one of: %s", paramName, strings.Join(schema.Enum, ", "))
			}
		}
		
		// Validate number ranges
		if schema.Type == "number" {
			if num, ok := value.(float64); ok {
				if schema.Min != nil && num < *schema.Min {
					return fmt.Errorf("parameter '%s' must be >= %f", paramName, *schema.Min)
				}
				if schema.Max != nil && num > *schema.Max {
					return fmt.Errorf("parameter '%s' must be <= %f", paramName, *schema.Max)
				}
			}
		}
	}
	
	return nil
}

func (t *CalculatorTool) validateParameterType(paramName string, value interface{}, schema ParameterSchema) error {
	switch schema.Type {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("parameter '%s' must be a string", paramName)
		}
	case "number":
		switch value.(type) {
		case int, int32, int64, float32, float64:
			// Valid number types
		default:
			return fmt.Errorf("parameter '%s' must be a number", paramName)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("parameter '%s' must be a boolean", paramName)
		}
	case "array":
		if reflect.TypeOf(value).Kind() != reflect.Slice {
			return fmt.Errorf("parameter '%s' must be an array", paramName)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("parameter '%s' must be an object", paramName)
		}
	}
	
	return nil
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	operation, _ := args["operation"].(string)
	
	// Convert numbers to float64
	var a, b float64
	switch v := args["a"].(type) {
	case int:
		a = float64(v)
	case int64:
		a = float64(v)
	case float32:
		a = float64(v)
	case float64:
		a = v
	default:
		return nil, fmt.Errorf("invalid type for parameter 'a'")
	}
	
	switch v := args["b"].(type) {
	case int:
		b = float64(v)
	case int64:
		b = float64(v)
	case float32:
		b = float64(v)
	case float64:
		b = v
	default:
		return nil, fmt.Errorf("invalid type for parameter 'b'")
	}
	
	switch operation {
	case "add":
		return a + b, nil
	case "subtract":
		return a - b, nil
	case "multiply":
		return a * b, nil
	case "divide":
		if b == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return a / b, nil
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (t *CalculatorTool) GetFallback(args map[string]interface{}) (interface{}, error) {
	// Simple mock calculation result
	return map[string]interface{}{
		"result": 42.0,
		"note":   "fallback response - actual calculation failed",
	}, nil
}