package llm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ExecutionMode defines where a task should be executed
type ExecutionMode string

const (
	ExecutionLocal  ExecutionMode = "local"
	ExecutionRemote ExecutionMode = "remote"
	ExecutionHybrid ExecutionMode = "hybrid"
	ExecutionAuto   ExecutionMode = "auto"
)

// TaskType defines different types of tasks
type TaskType string

const (
	TaskMovieRecommendation TaskType = "movie_recommendation"
	TaskIntentAnalysis     TaskType = "intent_analysis"
	TaskExplanationGen     TaskType = "explanation_generation"
	TaskMultimodalAnalysis TaskType = "multimodal_analysis"
	TaskUserProfiling      TaskType = "user_profiling"
	TaskContentFiltering   TaskType = "content_filtering"
	TaskSimilarityCalc     TaskType = "similarity_calculation"
)

// LocalTool defines the interface for local computation tools
type LocalTool interface {
	// GetName returns the tool name
	GetName() string
	
	// GetCapabilities returns supported task types
	GetCapabilities() []TaskType
	
	// Execute performs local computation
	Execute(ctx context.Context, req *LocalExecutionRequest) (*LocalExecutionResult, error)
	
	// GetPerformanceMetrics returns execution metrics
	GetPerformanceMetrics() *LocalToolMetrics
	
	// CanHandle checks if tool can handle the task
	CanHandle(taskType TaskType, complexity int) bool
}

// RemoteLLMTask defines tasks that require LLM processing
type RemoteLLMTask interface {
	// GetTaskType returns the task type
	GetTaskType() TaskType
	
	// RequiresLLM checks if task needs LLM processing
	RequiresLLM() bool
	
	// EstimateComplexity estimates task complexity (1-10)
	EstimateComplexity() int
	
	// GetPrompt generates LLM prompt for the task
	GetPrompt() string
	
	// ProcessLLMResponse processes LLM response
	ProcessLLMResponse(response *GenerateResponse) (interface{}, error)
}

// HybridExecutionRequest represents a request for hybrid processing
type HybridExecutionRequest struct {
	TaskID      string                 `json:"task_id"`
	TaskType    TaskType               `json:"task_type"`
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id"`
	Data        map[string]interface{} `json:"data"`
	Preferences *ExecutionPreferences  `json:"preferences,omitempty"`
	Context     *ExecutionContext      `json:"context"`
}

// ExecutionPreferences defines user preferences for execution
type ExecutionPreferences struct {
	PreferredMode      ExecutionMode `json:"preferred_mode"`
	MaxLatency         time.Duration `json:"max_latency"`
	RequireExplanation bool          `json:"require_explanation"`
	CostSensitive      bool          `json:"cost_sensitive"`
	QualityThreshold   float64       `json:"quality_threshold"`
}

// ExecutionContext provides context for decision making
type ExecutionContext struct {
	CurrentLoad     float64   `json:"current_load"`
	AvailableBudget float64   `json:"available_budget"`
	NetworkLatency  int64     `json:"network_latency_ms"`
	IsOffline       bool      `json:"is_offline"`
	Timestamp       time.Time `json:"timestamp"`
	DeviceType      string    `json:"device_type"`
}

// HybridExecutionResponse represents the response from hybrid processing
type HybridExecutionResponse struct {
	TaskID        string                 `json:"task_id"`
	Result        interface{}            `json:"result"`
	ExecutionMode ExecutionMode          `json:"execution_mode"`
	ProcessingTime time.Duration         `json:"processing_time"`
	LocalToolsUsed []string              `json:"local_tools_used"`
	RemoteCallsMade int                  `json:"remote_calls_made"`
	QualityScore   float64               `json:"quality_score"`
	Explanation    string                `json:"explanation,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// LocalExecutionRequest represents a local tool execution request
type LocalExecutionRequest struct {
	TaskType TaskType               `json:"task_type"`
	Data     map[string]interface{} `json:"data"`
	Context  *ExecutionContext      `json:"context"`
}

// LocalExecutionResult represents the result of local execution
type LocalExecutionResult struct {
	Result       interface{}   `json:"result"`
	Confidence   float64       `json:"confidence"`
	ProcessingTime time.Duration `json:"processing_time"`
	ToolsUsed    []string      `json:"tools_used"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// LocalToolMetrics represents performance metrics for local tools
type LocalToolMetrics struct {
	TotalExecutions   int64         `json:"total_executions"`
	AverageLatency    time.Duration `json:"average_latency"`
	SuccessRate       float64       `json:"success_rate"`
	CPUUsage          float64       `json:"cpu_usage"`
	MemoryUsage       float64       `json:"memory_usage"`
	LastUpdated       time.Time     `json:"last_updated"`
}

// DecisionEngine makes decisions about execution strategy
type DecisionEngine struct {
	localTools     map[TaskType][]LocalTool
	llmAdapter     LLMAdapter
	logger         *logrus.Logger
	config         *HybridConfig
	metrics        *HybridMetrics
	mu             sync.RWMutex
}

// HybridConfig configures the hybrid execution system
type HybridConfig struct {
	// Thresholds for decision making
	LocalExecutionThreshold   float64       `json:"local_execution_threshold"`
	RemoteExecutionThreshold  float64       `json:"remote_execution_threshold"`
	HybridExecutionThreshold  float64       `json:"hybrid_execution_threshold"`
	
	// Performance settings
	MaxLocalExecutionTime     time.Duration `json:"max_local_execution_time"`
	MaxRemoteExecutionTime    time.Duration `json:"max_remote_execution_time"`
	LocalToolTimeout          time.Duration `json:"local_tool_timeout"`
	
	// Quality settings
	MinQualityScore          float64 `json:"min_quality_score"`
	QualityVsSpeedTradeoff   float64 `json:"quality_vs_speed_tradeoff"`
	
	// Cost settings
	LocalExecutionCost       float64 `json:"local_execution_cost"`
	RemoteExecutionCost      float64 `json:"remote_execution_cost"`
	CostOptimizationEnabled  bool    `json:"cost_optimization_enabled"`
	
	// Fallback settings
	EnableFallback           bool `json:"enable_fallback"`
	FallbackStrategy         string `json:"fallback_strategy"`
}

// HybridMetrics tracks hybrid execution metrics
type HybridMetrics struct {
	LocalExecutions    int64         `json:"local_executions"`
	RemoteExecutions   int64         `json:"remote_executions"`
	HybridExecutions   int64         `json:"hybrid_executions"`
	TotalExecutions    int64         `json:"total_executions"`
	AverageLatency     time.Duration `json:"average_latency"`
	CostSavings        float64       `json:"cost_savings"`
	QualityScore       float64       `json:"quality_score"`
	FallbackRate       float64       `json:"fallback_rate"`
	LastUpdated        time.Time     `json:"last_updated"`
	mu                 sync.RWMutex
}

// HybridRecommendationSystem combines local tools with remote LLM capabilities
type HybridRecommendationSystem struct {
	decisionEngine    *DecisionEngine
	localToolManager *LocalToolManager
	remoteLLMManager *RemoteLLMManager
	logger           *logrus.Logger
	config           *HybridConfig
	metrics          *HybridMetrics
}

// NewHybridRecommendationSystem creates a new hybrid recommendation system
func NewHybridRecommendationSystem(config *HybridConfig, llmConfig *LLMAdapterConfig, logger *logrus.Logger) (*HybridRecommendationSystem, error) {
	if config == nil {
		config = createDefaultHybridConfig()
	}

	// Create LLM adapter
	llmAdapter, err := NewUnifiedLLMAdapter(llmConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM adapter: %w", err)
	}

	// Initialize components
	decisionEngine := &DecisionEngine{
		localTools: make(map[TaskType][]LocalTool),
		llmAdapter: llmAdapter,
		logger:     logger,
		config:     config,
		metrics:    &HybridMetrics{},
	}

	localToolManager := NewLocalToolManager(logger)
	remoteLLMManager := NewRemoteLLMManager(llmAdapter, logger)

	system := &HybridRecommendationSystem{
		decisionEngine:    decisionEngine,
		localToolManager: localToolManager,
		remoteLLMManager: remoteLLMManager,
		logger:           logger,
		config:           config,
		metrics:          &HybridMetrics{LastUpdated: time.Now()},
	}

	// Register default local tools
	if err := system.registerDefaultLocalTools(); err != nil {
		return nil, fmt.Errorf("failed to register default tools: %w", err)
	}

	logger.Info("Hybrid recommendation system initialized successfully")
	return system, nil
}

// ExecuteTask executes a task using the optimal execution strategy
func (hrs *HybridRecommendationSystem) ExecuteTask(ctx context.Context, req *HybridExecutionRequest) (*HybridExecutionResponse, error) {
	startTime := time.Now()
	
	hrs.logger.WithFields(logrus.Fields{
		"task_id":   req.TaskID,
		"task_type": req.TaskType,
		"user_id":   req.UserID,
	}).Info("Starting hybrid task execution")

	// Determine optimal execution strategy
	strategy, err := hrs.decisionEngine.DetermineExecutionStrategy(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to determine execution strategy: %w", err)
	}

	var result interface{}
	var executionMode ExecutionMode
	var localToolsUsed []string
	var remoteCallsMade int
	var qualityScore float64

	// Execute based on strategy
	switch strategy {
	case ExecutionLocal:
		localResult, err := hrs.localToolManager.ExecuteTask(ctx, req)
		if err != nil {
			// Fallback to remote if enabled
			if hrs.config.EnableFallback {
				hrs.logger.Warn("Local execution failed, falling back to remote")
				remoteResult, fallbackErr := hrs.remoteLLMManager.ExecuteTask(ctx, req)
				if fallbackErr != nil {
					return nil, fmt.Errorf("both local and remote execution failed: %w", err)
				}
				result = remoteResult.Result
				executionMode = ExecutionRemote
				remoteCallsMade = 1
				qualityScore = remoteResult.QualityScore
			} else {
				return nil, fmt.Errorf("local execution failed: %w", err)
			}
		} else {
			result = localResult.Result
			executionMode = ExecutionLocal
			localToolsUsed = localResult.ToolsUsed
			qualityScore = localResult.Confidence
		}

	case ExecutionRemote:
		remoteResult, err := hrs.remoteLLMManager.ExecuteTask(ctx, req)
		if err != nil {
			// Fallback to local if enabled
			if hrs.config.EnableFallback {
				hrs.logger.Warn("Remote execution failed, falling back to local")
				localResult, fallbackErr := hrs.localToolManager.ExecuteTask(ctx, req)
				if fallbackErr != nil {
					return nil, fmt.Errorf("both remote and local execution failed: %w", err)
				}
				result = localResult.Result
				executionMode = ExecutionLocal
				localToolsUsed = localResult.ToolsUsed
				qualityScore = localResult.Confidence
			} else {
				return nil, fmt.Errorf("remote execution failed: %w", err)
			}
		} else {
			result = remoteResult.Result
			executionMode = ExecutionRemote
			remoteCallsMade = 1
			qualityScore = remoteResult.QualityScore
		}

	case ExecutionHybrid:
		hybridResult, err := hrs.executeHybridTask(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("hybrid execution failed: %w", err)
		}
		result = hybridResult.Result
		executionMode = ExecutionHybrid
		localToolsUsed = hybridResult.LocalToolsUsed
		remoteCallsMade = hybridResult.RemoteCallsMade
		qualityScore = hybridResult.QualityScore
	}

	processingTime := time.Since(startTime)
	
	// Update metrics
	hrs.updateMetrics(executionMode, processingTime, qualityScore)

	response := &HybridExecutionResponse{
		TaskID:          req.TaskID,
		Result:          result,
		ExecutionMode:   executionMode,
		ProcessingTime:  processingTime,
		LocalToolsUsed:  localToolsUsed,
		RemoteCallsMade: remoteCallsMade,
		QualityScore:    qualityScore,
		Metadata: map[string]interface{}{
			"strategy":     strategy,
			"start_time":   startTime,
			"end_time":     time.Now(),
		},
	}

	hrs.logger.WithFields(logrus.Fields{
		"task_id":         req.TaskID,
		"execution_mode":  executionMode,
		"processing_time": processingTime,
		"quality_score":   qualityScore,
	}).Info("Hybrid task execution completed")

	return response, nil
}

// DetermineExecutionStrategy determines the optimal execution strategy
func (de *DecisionEngine) DetermineExecutionStrategy(ctx context.Context, req *HybridExecutionRequest) (ExecutionMode, error) {
	de.mu.RLock()
	defer de.mu.RUnlock()

	// Check user preferences first
	if req.Preferences != nil && req.Preferences.PreferredMode != ExecutionAuto {
		return req.Preferences.PreferredMode, nil
	}

	// Calculate execution scores
	localScore := de.calculateLocalScore(req)
	remoteScore := de.calculateRemoteScore(req)
	hybridScore := de.calculateHybridScore(req)

	de.logger.WithFields(logrus.Fields{
		"task_type":    req.TaskType,
		"local_score":  localScore,
		"remote_score": remoteScore,
		"hybrid_score": hybridScore,
	}).Debug("Execution strategy scores calculated")

	// Determine strategy based on scores
	if localScore >= de.config.LocalExecutionThreshold && localScore > remoteScore && localScore > hybridScore {
		return ExecutionLocal, nil
	}
	
	if remoteScore >= de.config.RemoteExecutionThreshold && remoteScore > localScore && remoteScore > hybridScore {
		return ExecutionRemote, nil
	}
	
	if hybridScore >= de.config.HybridExecutionThreshold {
		return ExecutionHybrid, nil
	}

	// Default to remote if no clear winner
	return ExecutionRemote, nil
}

// calculateLocalScore calculates the score for local execution
func (de *DecisionEngine) calculateLocalScore(req *HybridExecutionRequest) float64 {
	score := 0.0
	
	// Check if local tools can handle the task
	if tools, exists := de.localTools[req.TaskType]; exists && len(tools) > 0 {
		for _, tool := range tools {
			complexity := 5 // Default complexity
			if tool.CanHandle(req.TaskType, complexity) {
				metrics := tool.GetPerformanceMetrics()
				score += metrics.SuccessRate * 0.4 // Weight by success rate
				
				// Add latency bonus
				if metrics.AverageLatency < de.config.MaxLocalExecutionTime {
					score += 0.3
				}
				
				// Add resource usage penalty
				resourcePenalty := (metrics.CPUUsage + metrics.MemoryUsage) / 200.0
				score -= resourcePenalty
				
				break // Use the best tool
			}
		}
	}
	
	// Context bonuses
	if req.Context != nil {
		// Offline bonus
		if req.Context.IsOffline {
			score += 0.5
		}
		
		// Low latency bonus
		if req.Context.NetworkLatency > 200 {
			score += 0.2
		}
		
		// Low load bonus
		if req.Context.CurrentLoad < 0.5 {
			score += 0.1
		}
	}
	
	// Cost optimization bonus
	if req.Preferences != nil && req.Preferences.CostSensitive {
		score += 0.2
	}
	
	return score
}

// calculateRemoteScore calculates the score for remote execution
func (de *DecisionEngine) calculateRemoteScore(req *HybridExecutionRequest) float64 {
	score := 0.5 // Base score for LLM capability
	
	// Task complexity bonus
	complexTasks := map[TaskType]float64{
		TaskIntentAnalysis:     0.3,
		TaskExplanationGen:     0.4,
		TaskMultimodalAnalysis: 0.3,
		TaskUserProfiling:      0.2,
	}
	
	if bonus, exists := complexTasks[req.TaskType]; exists {
		score += bonus
	}
	
	// Quality requirement bonus
	if req.Preferences != nil && req.Preferences.RequireExplanation {
		score += 0.2
	}
	
	// Context penalties
	if req.Context != nil {
		// High latency penalty
		if req.Context.NetworkLatency > 500 {
			score -= 0.3
		}
		
		// Offline penalty
		if req.Context.IsOffline {
			score -= 1.0 // Cannot execute remotely
		}
		
		// Budget constraint penalty
		if req.Context.AvailableBudget < de.config.RemoteExecutionCost {
			score -= 0.4
		}
	}
	
	return score
}

// calculateHybridScore calculates the score for hybrid execution
func (de *DecisionEngine) calculateHybridScore(req *HybridExecutionRequest) float64 {
	localScore := de.calculateLocalScore(req)
	remoteScore := de.calculateRemoteScore(req)
	
	// Hybrid is beneficial when both local and remote have moderate scores
	if localScore > 0.3 && remoteScore > 0.3 && localScore < 0.8 && remoteScore < 0.8 {
		return (localScore + remoteScore) / 2 + 0.1 // Small hybrid bonus
	}
	
	return 0.0
}

// executeHybridTask executes a task using both local and remote resources
func (hrs *HybridRecommendationSystem) executeHybridTask(ctx context.Context, req *HybridExecutionRequest) (*HybridExecutionResponse, error) {
	// Execute local and remote in parallel
	localChan := make(chan *LocalExecutionResult, 1)
	remoteChan := make(chan *RemoteLLMResult, 1)
	errorChan := make(chan error, 2)

	// Launch local execution
	go func() {
		result, err := hrs.localToolManager.ExecuteTask(ctx, req)
		if err != nil {
			errorChan <- err
		} else {
			localChan <- result
		}
	}()

	// Launch remote execution
	go func() {
		result, err := hrs.remoteLLMManager.ExecuteTask(ctx, req)
		if err != nil {
			errorChan <- err
		} else {
			remoteChan <- result
		}
	}()

	// Collect results
	var localResult *LocalExecutionResult
	var remoteResult *RemoteLLMResult
	errors := make([]error, 0, 2)

	for i := 0; i < 2; i++ {
		select {
		case lr := <-localChan:
			localResult = lr
		case rr := <-remoteChan:
			remoteResult = rr
		case err := <-errorChan:
			errors = append(errors, err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Combine results intelligently
	var finalResult interface{}
	var localToolsUsed []string
	var remoteCallsMade int
	var qualityScore float64

	if localResult != nil && remoteResult != nil {
		// Both succeeded - combine results
		finalResult = hrs.combineResults(localResult, remoteResult)
		localToolsUsed = localResult.ToolsUsed
		remoteCallsMade = 1
		qualityScore = (localResult.Confidence + remoteResult.QualityScore) / 2
	} else if localResult != nil {
		// Only local succeeded
		finalResult = localResult.Result
		localToolsUsed = localResult.ToolsUsed
		qualityScore = localResult.Confidence
	} else if remoteResult != nil {
		// Only remote succeeded
		finalResult = remoteResult.Result
		remoteCallsMade = 1
		qualityScore = remoteResult.QualityScore
	} else {
		// Both failed
		return nil, fmt.Errorf("hybrid execution failed: local and remote errors: %v", errors)
	}

	return &HybridExecutionResponse{
		TaskID:          req.TaskID,
		Result:          finalResult,
		ExecutionMode:   ExecutionHybrid,
		LocalToolsUsed:  localToolsUsed,
		RemoteCallsMade: remoteCallsMade,
		QualityScore:    qualityScore,
	}, nil
}

// combineResults combines local and remote results intelligently
func (hrs *HybridRecommendationSystem) combineResults(local *LocalExecutionResult, remote *RemoteLLMResult) interface{} {
	// Task-specific result combination logic
	combined := map[string]interface{}{
		"local_result":      local.Result,
		"remote_result":     remote.Result,
		"local_confidence":  local.Confidence,
		"remote_quality":    remote.QualityScore,
		"combined_approach": true,
	}

	// If local confidence is high, prioritize local result
	if local.Confidence > 0.8 {
		combined["primary_result"] = local.Result
		combined["explanation"] = remote.Result // Use remote for explanation
	} else {
		// Otherwise, prioritize remote result
		combined["primary_result"] = remote.Result
		combined["local_validation"] = local.Result
	}

	return combined
}

// registerDefaultLocalTools registers default local tools
func (hrs *HybridRecommendationSystem) registerDefaultLocalTools() error {
	// Register collaborative filtering tool
	cfTool := NewCollaborativeFilteringTool(hrs.logger)
	hrs.decisionEngine.RegisterLocalTool(TaskMovieRecommendation, cfTool)

	// Register content filtering tool  
	contentTool := NewContentFilteringTool(hrs.logger)
	hrs.decisionEngine.RegisterLocalTool(TaskContentFiltering, contentTool)

	// Register similarity calculation tool
	simTool := NewSimilarityCalculationTool(hrs.logger)
	hrs.decisionEngine.RegisterLocalTool(TaskSimilarityCalc, simTool)

	return nil
}

// RegisterLocalTool registers a local tool for a specific task type
func (de *DecisionEngine) RegisterLocalTool(taskType TaskType, tool LocalTool) {
	de.mu.Lock()
	defer de.mu.Unlock()

	if de.localTools[taskType] == nil {
		de.localTools[taskType] = make([]LocalTool, 0)
	}
	de.localTools[taskType] = append(de.localTools[taskType], tool)

	de.logger.WithFields(logrus.Fields{
		"task_type": taskType,
		"tool_name": tool.GetName(),
	}).Info("Local tool registered")
}

// updateMetrics updates system metrics
func (hrs *HybridRecommendationSystem) updateMetrics(mode ExecutionMode, duration time.Duration, quality float64) {
	hrs.metrics.mu.Lock()
	defer hrs.metrics.mu.Unlock()

	hrs.metrics.TotalExecutions++
	
	switch mode {
	case ExecutionLocal:
		hrs.metrics.LocalExecutions++
	case ExecutionRemote:
		hrs.metrics.RemoteExecutions++
	case ExecutionHybrid:
		hrs.metrics.HybridExecutions++
	}

	// Update average latency
	if hrs.metrics.TotalExecutions == 1 {
		hrs.metrics.AverageLatency = duration
	} else {
		hrs.metrics.AverageLatency = (hrs.metrics.AverageLatency + duration) / 2
	}

	// Update quality score
	if hrs.metrics.TotalExecutions == 1 {
		hrs.metrics.QualityScore = quality
	} else {
		hrs.metrics.QualityScore = (hrs.metrics.QualityScore + quality) / 2
	}

	hrs.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current system metrics
func (hrs *HybridRecommendationSystem) GetMetrics() *HybridMetrics {
	hrs.metrics.mu.RLock()
	defer hrs.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *hrs.metrics
	return &metricsCopy
}

// createDefaultHybridConfig creates default hybrid configuration
func createDefaultHybridConfig() *HybridConfig {
	return &HybridConfig{
		LocalExecutionThreshold:   0.6,
		RemoteExecutionThreshold:  0.7,
		HybridExecutionThreshold:  0.5,
		MaxLocalExecutionTime:     2 * time.Second,
		MaxRemoteExecutionTime:    10 * time.Second,
		LocalToolTimeout:          1 * time.Second,
		MinQualityScore:          0.5,
		QualityVsSpeedTradeoff:   0.7,
		LocalExecutionCost:       0.001,
		RemoteExecutionCost:      0.01,
		CostOptimizationEnabled:  true,
		EnableFallback:           true,
		FallbackStrategy:         "adaptive",
	}
}