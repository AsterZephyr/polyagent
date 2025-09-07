package recommendation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// NewRecommendationOrchestrator creates a new recommendation business orchestrator
func NewRecommendationOrchestrator(config *OrchestratorConfig, logger *logrus.Logger) *RecommendationOrchestrator {
	if config == nil {
		config = &OrchestratorConfig{
			MaxConcurrentTasks: 100,
			TaskTimeout:        5 * time.Minute,
			HealthCheckInterval: 30 * time.Second,
			MetricsInterval:    1 * time.Minute,
			RetryPolicy: &RetryPolicy{
				MaxRetries:        3,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 2.0,
				MaxDelay:          30 * time.Second,
			},
		}
	}

	orchestrator := &RecommendationOrchestrator{
		agents:      make(map[string]RecommendationAgent),
		taskQueue:   make(chan *RecommendationTask, config.MaxConcurrentTasks*2),
		resultQueue: make(chan *RecommendationResult, config.MaxConcurrentTasks*2),
		logger:      logger,
		config:      config,
	}

	return orchestrator
}

// RegisterAgent registers a recommendation agent
func (ro *RecommendationOrchestrator) RegisterAgent(agent RecommendationAgent) error {
	if agent == nil {
		return fmt.Errorf("agent cannot be nil")
	}

	agentID := agent.GetID()
	if agentID == "" {
		return fmt.Errorf("agent ID cannot be empty")
	}

	ro.agents[agentID] = agent

	ro.logger.WithFields(logrus.Fields{
		"agent_id":   agentID,
		"agent_type": agent.GetType(),
		"capabilities": agent.GetCapabilities(),
	}).Info("Recommendation agent registered")

	return nil
}

// SubmitTask submits a task for processing
func (ro *RecommendationOrchestrator) SubmitTask(task *RecommendationTask) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}

	if task.ID == "" {
		task.ID = ro.generateTaskID()
	}

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}

	if task.MaxRetries == 0 {
		task.MaxRetries = ro.config.RetryPolicy.MaxRetries
	}

	select {
	case ro.taskQueue <- task:
		ro.logger.WithFields(logrus.Fields{
			"task_id":   task.ID,
			"task_type": task.Type,
			"priority":  task.Priority,
		}).Info("Task submitted to queue")
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// ProcessTask processes a single task
func (ro *RecommendationOrchestrator) ProcessTask(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	// Find suitable agent for the task
	agent, err := ro.findAgentForTask(task)
	if err != nil {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("no suitable agent found: %s", err.Error()),
			CreatedAt: time.Now(),
		}, err
	}

	// Create timeout context
	taskCtx := ctx
	if task.Deadline != nil {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithDeadline(ctx, *task.Deadline)
		defer cancel()
	} else {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx, ro.config.TaskTimeout)
		defer cancel()
	}

	// Execute task with retry logic
	var result *RecommendationResult
	var lastErr error

	for attempt := 0; attempt <= task.MaxRetries; attempt++ {
		if attempt > 0 {
			// Apply backoff delay
			delay := ro.calculateRetryDelay(attempt)
			ro.logger.WithFields(logrus.Fields{
				"task_id":  task.ID,
				"attempt":  attempt + 1,
				"delay":    delay,
			}).Info("Retrying task after delay")

			select {
			case <-time.After(delay):
			case <-taskCtx.Done():
				return &RecommendationResult{
					TaskID:    task.ID,
					Success:   false,
					Error:     "task cancelled during retry delay",
					CreatedAt: time.Now(),
				}, taskCtx.Err()
			}
		}

		// Execute the task
		start := time.Now()
		result, lastErr = agent.Process(taskCtx, task)
		duration := time.Since(start)

		if lastErr == nil && result.Success {
			ro.logger.WithFields(logrus.Fields{
				"task_id":     task.ID,
				"agent_id":    agent.GetID(),
				"duration":    duration,
				"attempt":     attempt + 1,
			}).Info("Task completed successfully")
			return result, nil
		}

		ro.logger.WithFields(logrus.Fields{
			"task_id":  task.ID,
			"agent_id": agent.GetID(),
			"attempt":  attempt + 1,
			"error":    lastErr,
		}).Warn("Task execution failed")

		task.RetryCount++
	}

	// All retries exhausted
	return &RecommendationResult{
		TaskID:    task.ID,
		Success:   false,
		Error:     fmt.Sprintf("task failed after %d retries: %s", task.MaxRetries, lastErr.Error()),
		CreatedAt: time.Now(),
	}, lastErr
}

// findAgentForTask finds the most suitable agent for a given task
func (ro *RecommendationOrchestrator) findAgentForTask(task *RecommendationTask) (RecommendationAgent, error) {
	var suitableAgents []RecommendationAgent

	for _, agent := range ro.agents {
		if ro.isAgentSuitableForTask(agent, task) && agent.GetStatus() == StatusIdle {
			suitableAgents = append(suitableAgents, agent)
		}
	}

	if len(suitableAgents) == 0 {
		return nil, fmt.Errorf("no suitable agent found for task type: %s", task.Type)
	}

	// Select the best agent based on performance metrics
	return ro.selectBestAgent(suitableAgents), nil
}

// isAgentSuitableForTask checks if an agent can handle a specific task type
func (ro *RecommendationOrchestrator) isAgentSuitableForTask(agent RecommendationAgent, task *RecommendationTask) bool {
	agentType := agent.GetType()
	taskType := task.Type

	switch taskType {
	case TaskDataCollection, TaskFeatureEngineering, TaskDataCleaning, TaskDataValidation:
		return agentType == AgentTypeData
	case TaskModelTraining, TaskModelEvaluation, TaskHyperParamTuning, TaskModelDeployment:
		return agentType == AgentTypeModel
	case TaskRealTimeInference, TaskCacheManagement, TaskLoadBalancing, TaskServiceMonitoring:
		return agentType == AgentTypeService
	case TaskABTesting, TaskMetricsAnalysis, TaskEffectEvaluation, TaskReportGeneration:
		return agentType == AgentTypeEval
	}

	return false
}

// selectBestAgent selects the best performing agent from candidates
func (ro *RecommendationOrchestrator) selectBestAgent(agents []RecommendationAgent) RecommendationAgent {
	if len(agents) == 1 {
		return agents[0]
	}

	bestAgent := agents[0]
	bestScore := ro.calculateAgentScore(bestAgent)

	for i := 1; i < len(agents); i++ {
		score := ro.calculateAgentScore(agents[i])
		if score > bestScore {
			bestAgent = agents[i]
			bestScore = score
		}
	}

	return bestAgent
}

// calculateAgentScore calculates agent performance score for selection
func (ro *RecommendationOrchestrator) calculateAgentScore(agent RecommendationAgent) float64 {
	metrics := agent.GetMetrics()
	if metrics == nil {
		return 0.0
	}

	// Weighted scoring: success rate (60%) + latency (25%) + resource efficiency (15%)
	successRateScore := metrics.SuccessRate * 0.6
	
	// Lower latency = higher score (inverse relationship)
	latencyScore := 0.0
	if metrics.AverageLatency > 0 {
		latencyScore = (1.0 - float64(metrics.AverageLatency.Seconds())/60.0) * 0.25
		if latencyScore < 0 {
			latencyScore = 0
		}
	}

	// Lower resource usage = higher score
	resourceScore := 0.0
	if metrics.ResourceUsage != nil {
		resourceEfficiency := 1.0 - (metrics.ResourceUsage.CPUPercent/100.0 + 
			float64(metrics.ResourceUsage.MemoryMB)/8192.0) / 2.0
		if resourceEfficiency > 0 {
			resourceScore = resourceEfficiency * 0.15
		}
	}

	return successRateScore + latencyScore + resourceScore
}

// calculateRetryDelay calculates delay for retry attempts with exponential backoff
func (ro *RecommendationOrchestrator) calculateRetryDelay(attempt int) time.Duration {
	policy := ro.config.RetryPolicy
	delay := float64(policy.InitialDelay)
	
	for i := 1; i < attempt; i++ {
		delay *= policy.BackoffMultiplier
	}

	result := time.Duration(delay)
	if result > policy.MaxDelay {
		result = policy.MaxDelay
	}

	return result
}

// generateTaskID generates a unique task ID
func (ro *RecommendationOrchestrator) generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// GetAgents returns all registered agents
func (ro *RecommendationOrchestrator) GetAgents() map[string]RecommendationAgent {
	result := make(map[string]RecommendationAgent)
	for id, agent := range ro.agents {
		result[id] = agent
	}
	return result
}

// GetSystemMetrics returns overall system metrics
func (ro *RecommendationOrchestrator) GetSystemMetrics() *SystemMetrics {
	metrics := &SystemMetrics{
		TotalAgents:      len(ro.agents),
		ActiveAgents:     0,
		QueuedTasks:      len(ro.taskQueue),
		ProcessingTasks:  0,
		TotalTasksToday:  0,
		SuccessRateToday: 0.0,
		AverageLatency:   0,
		Timestamp:        time.Now(),
	}

	for _, agent := range ro.agents {
		if agent.GetStatus() == StatusProcessing {
			metrics.ProcessingTasks++
		}
		if agent.GetStatus() != StatusError && agent.GetStatus() != StatusMaintenance {
			metrics.ActiveAgents++
		}
	}

	return metrics
}

// SystemMetrics represents overall system performance metrics
type SystemMetrics struct {
	TotalAgents      int           `json:"total_agents"`
	ActiveAgents     int           `json:"active_agents"`
	QueuedTasks      int           `json:"queued_tasks"`
	ProcessingTasks  int           `json:"processing_tasks"`
	TotalTasksToday  int64         `json:"total_tasks_today"`
	SuccessRateToday float64       `json:"success_rate_today"`
	AverageLatency   time.Duration `json:"average_latency"`
	Timestamp        time.Time     `json:"timestamp"`
}

// StartTaskProcessor starts the background task processor
func (ro *RecommendationOrchestrator) StartTaskProcessor(ctx context.Context) {
	ro.logger.Info("Starting recommendation orchestrator task processor")

	var wg sync.WaitGroup

	// Start task processing workers
	for i := 0; i < ro.config.MaxConcurrentTasks; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			ro.taskWorker(ctx, workerID)
		}(i)
	}

	// Start health check routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ro.healthCheckRoutine(ctx)
	}()

	// Start metrics collection routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		ro.metricsRoutine(ctx)
	}()

	wg.Wait()
	ro.logger.Info("Recommendation orchestrator stopped")
}

// taskWorker processes tasks from the queue
func (ro *RecommendationOrchestrator) taskWorker(ctx context.Context, workerID int) {
	ro.logger.WithField("worker_id", workerID).Info("Task worker started")

	for {
		select {
		case task := <-ro.taskQueue:
			result, err := ro.ProcessTask(ctx, task)
			if err != nil {
				ro.logger.WithError(err).WithField("task_id", task.ID).Error("Task processing failed")
			}

			// Send result to result queue
			select {
			case ro.resultQueue <- result:
			default:
				ro.logger.WithField("task_id", task.ID).Warn("Result queue full, dropping result")
			}

		case <-ctx.Done():
			ro.logger.WithField("worker_id", workerID).Info("Task worker shutting down")
			return
		}
	}
}

// healthCheckRoutine periodically checks agent health
func (ro *RecommendationOrchestrator) healthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(ro.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for agentID, agent := range ro.agents {
				if err := agent.HealthCheck(); err != nil {
					ro.logger.WithError(err).WithField("agent_id", agentID).Error("Agent health check failed")
				}
			}

		case <-ctx.Done():
			ro.logger.Info("Health check routine shutting down")
			return
		}
	}
}

// metricsRoutine periodically collects and logs system metrics
func (ro *RecommendationOrchestrator) metricsRoutine(ctx context.Context) {
	ticker := time.NewTicker(ro.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := ro.GetSystemMetrics()
			ro.logger.WithFields(logrus.Fields{
				"total_agents":       metrics.TotalAgents,
				"active_agents":      metrics.ActiveAgents,
				"queued_tasks":       metrics.QueuedTasks,
				"processing_tasks":   metrics.ProcessingTasks,
			}).Info("System metrics collected")

		case <-ctx.Done():
			ro.logger.Info("Metrics routine shutting down")
			return
		}
	}
}