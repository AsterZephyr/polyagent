package recommendation

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ModelAgent handles model training, evaluation, and deployment for recommendation systems
type ModelAgent struct {
	id         string
	status     AgentStatus
	logger     *logrus.Logger
	config     *ModelAgentConfig
	algorithms map[string]RecommendationAlgorithm
	models     map[string]*TrainedModel
	trainer    *ModelTrainer
	evaluator  *ModelEvaluator
	optimizer  *HyperParameterOptimizer
	deployer   *ModelDeployer
	registry   *ModelRegistry
	metrics    *AgentMetrics
	mutex      sync.RWMutex
	startTime  time.Time
}

// ModelAgentConfig defines configuration for ModelAgent
type ModelAgentConfig struct {
	MaxConcurrentTraining int           `json:"max_concurrent_training"`
	ModelRetentionDays    int           `json:"model_retention_days"`
	AutoDeployThreshold   float64       `json:"auto_deploy_threshold"`
	EvaluationInterval    time.Duration `json:"evaluation_interval"`
	TrainingTimeout       time.Duration `json:"training_timeout"`
	SupportedAlgorithms   []string      `json:"supported_algorithms"`
}

// RecommendationAlgorithm interface for different recommendation algorithms
type RecommendationAlgorithm interface {
	Name() string
	Train(ctx context.Context, trainingData *TrainingData, params *TrainingParams) (*TrainedModel, error)
	Predict(ctx context.Context, model *TrainedModel, input *PredictionInput) (*PredictionOutput, error)
	GetHyperParameters() map[string]HyperParameter
	GetMetrics() *AlgorithmMetrics
}

// TrainingData represents data used for model training
type TrainingData struct {
	ID           string                   `json:"id"`
	UserFeatures []map[string]interface{} `json:"user_features"`
	ItemFeatures []map[string]interface{} `json:"item_features"`
	Interactions []Interaction            `json:"interactions"`
	Metadata     map[string]interface{}   `json:"metadata"`
	SplitRatio   *DataSplit               `json:"split_ratio"`
	CreatedAt    time.Time                `json:"created_at"`
}

// Interaction represents user-item interaction
type Interaction struct {
	UserID    string                 `json:"user_id"`
	ItemID    string                 `json:"item_id"`
	Rating    float64                `json:"rating"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
}

// DataSplit defines train/validation/test split ratios
type DataSplit struct {
	TrainRatio      float64 `json:"train_ratio"`
	ValidationRatio float64 `json:"validation_ratio"`
	TestRatio       float64 `json:"test_ratio"`
}

// TrainingParams contains parameters for model training
type TrainingParams struct {
	Algorithm        string                 `json:"algorithm"`
	HyperParameters  map[string]interface{} `json:"hyperparameters"`
	EarlyStopping    *EarlyStoppingConfig   `json:"early_stopping"`
	ValidationMetric string                 `json:"validation_metric"`
	MaxEpochs        int                    `json:"max_epochs"`
	BatchSize        int                    `json:"batch_size"`
	LearningRate     float64                `json:"learning_rate"`
}

// EarlyStoppingConfig defines early stopping criteria
type EarlyStoppingConfig struct {
	Enabled  bool    `json:"enabled"`
	Patience int     `json:"patience"`
	MinDelta float64 `json:"min_delta"`
	Metric   string  `json:"metric"`
	Mode     string  `json:"mode"` // "min" or "max"
}

// TrainedModel represents a trained recommendation model
type TrainedModel struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Algorithm       string                 `json:"algorithm"`
	Version         string                 `json:"version"`
	Parameters      map[string]interface{} `json:"parameters"`
	TrainingMetrics *TrainingMetrics       `json:"training_metrics"`
	EvalMetrics     *EvaluationMetrics     `json:"eval_metrics"`
	ModelData       []byte                 `json:"model_data"` // Serialized model
	Metadata        map[string]interface{} `json:"metadata"`
	Status          ModelStatus            `json:"status"`
	CreatedAt       time.Time              `json:"created_at"`
	TrainingTime    time.Duration          `json:"training_time"`
}

// Note: ModelStatus, PredictionInput, PredictionOutput moved to agent_types.go

// TrainingMetrics contains metrics from model training
type TrainingMetrics struct {
	TrainLoss      []float64     `json:"train_loss"`
	ValidLoss      []float64     `json:"valid_loss"`
	Epochs         int           `json:"epochs"`
	TrainingTime   time.Duration `json:"training_time"`
	ConvergedEpoch int           `json:"converged_epoch"`
	FinalTrainLoss float64       `json:"final_train_loss"`
	FinalValidLoss float64       `json:"final_valid_loss"`
}

// EvaluationMetrics contains model evaluation results
type EvaluationMetrics struct {
	RMSE          float64            `json:"rmse"`
	MAE           float64            `json:"mae"`
	Precision     map[int]float64    `json:"precision"` // Precision@K
	Recall        map[int]float64    `json:"recall"`    // Recall@K
	NDCG          map[int]float64    `json:"ndcg"`      // NDCG@K
	HitRate       map[int]float64    `json:"hit_rate"`  // HitRate@K
	AUC           float64            `json:"auc"`
	Coverage      float64            `json:"coverage"`
	Diversity     float64            `json:"diversity"`
	Novelty       float64            `json:"novelty"`
	CustomMetrics map[string]float64 `json:"custom_metrics"`
}

// ModelTrainer handles model training orchestration
type ModelTrainer struct {
	maxConcurrent int
	activeJobs    map[string]*TrainingJob
	mutex         sync.RWMutex
}

// TrainingJob represents an active training job
type TrainingJob struct {
	ID           string    `json:"id"`
	Algorithm    string    `json:"algorithm"`
	Status       string    `json:"status"`
	Progress     float64   `json:"progress"`
	StartTime    time.Time `json:"start_time"`
	EstimatedETA time.Time `json:"estimated_eta"`
	Logs         []string  `json:"logs"`
}

// ModelEvaluator handles model evaluation
type ModelEvaluator struct {
	testSuites map[string]*EvaluationSuite
}

// EvaluationSuite defines a set of evaluation tests
type EvaluationSuite struct {
	Name       string                 `json:"name"`
	TestData   *TrainingData          `json:"test_data"`
	Metrics    []string               `json:"metrics"`
	Parameters map[string]interface{} `json:"parameters"`
}

// HyperParameterOptimizer handles automatic hyperparameter tuning
type HyperParameterOptimizer struct {
	method   string // "grid_search", "random_search", "bayesian"
	budget   int    // Number of trials
	parallel bool   // Whether to run trials in parallel
}

// ModelDeployer handles model deployment
type ModelDeployer struct {
	strategies map[string]DeploymentStrategy
}

// DeploymentStrategy defines how models are deployed
type DeploymentStrategy interface {
	Deploy(ctx context.Context, model *TrainedModel, config map[string]interface{}) error
	Rollback(ctx context.Context, modelID string) error
	HealthCheck(ctx context.Context, modelID string) error
}

// ModelRegistry manages model versions and metadata
type ModelRegistry struct {
	models   map[string]*TrainedModel
	versions map[string][]string // modelName -> versions
	active   map[string]string   // modelName -> activeVersion
	mutex    sync.RWMutex
}

// NewModelAgent creates a new ModelAgent instance
func NewModelAgent(id string, config *ModelAgentConfig, logger *logrus.Logger) (*ModelAgent, error) {
	if config == nil {
		config = &ModelAgentConfig{
			MaxConcurrentTraining: 3,
			ModelRetentionDays:    90,
			AutoDeployThreshold:   0.85,
			EvaluationInterval:    6 * time.Hour,
			TrainingTimeout:       2 * time.Hour,
			SupportedAlgorithms:   []string{"collaborative_filtering", "content_based", "matrix_factorization", "deep_learning"},
		}
	}

	agent := &ModelAgent{
		id:         id,
		status:     StatusIdle,
		logger:     logger,
		config:     config,
		algorithms: make(map[string]RecommendationAlgorithm),
		models:     make(map[string]*TrainedModel),
		trainer:    &ModelTrainer{maxConcurrent: config.MaxConcurrentTraining, activeJobs: make(map[string]*TrainingJob)},
		evaluator:  &ModelEvaluator{testSuites: make(map[string]*EvaluationSuite)},
		optimizer:  &HyperParameterOptimizer{method: "random_search", budget: 50, parallel: true},
		deployer:   &ModelDeployer{strategies: make(map[string]DeploymentStrategy)},
		registry:   &ModelRegistry{models: make(map[string]*TrainedModel), versions: make(map[string][]string), active: make(map[string]string)},
		metrics: &AgentMetrics{
			TasksProcessed: 0,
			SuccessRate:    1.0,
			ErrorCount:     0,
			ResourceUsage:  &ResourceUsage{},
		},
		startTime: time.Now(),
	}

	// Initialize built-in algorithms
	if err := agent.initializeAlgorithms(); err != nil {
		return nil, fmt.Errorf("failed to initialize ModelAgent algorithms: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"agent_id":   id,
		"agent_type": AgentTypeModel,
		"algorithms": len(agent.algorithms),
	}).Info("ModelAgent created successfully")

	return agent, nil
}

// GetID returns the agent's unique identifier
func (ma *ModelAgent) GetID() string {
	return ma.id
}

// GetType returns the agent type
func (ma *ModelAgent) GetType() RecommendationAgentType {
	return AgentTypeModel
}

// GetStatus returns current agent status
func (ma *ModelAgent) GetStatus() AgentStatus {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()
	return ma.status
}

// GetMetrics returns agent performance metrics
func (ma *ModelAgent) GetMetrics() *AgentMetrics {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	// Create a copy to avoid race conditions
	metricsCopy := *ma.metrics
	return &metricsCopy
}

// GetCapabilities returns list of agent capabilities
func (ma *ModelAgent) GetCapabilities() []string {
	return []string{
		"model_training",
		"model_evaluation",
		"hyperparameter_tuning",
		"model_deployment",
		"model_versioning",
		"performance_monitoring",
		"algorithm_comparison",
		"automatic_optimization",
	}
}

// Process handles incoming recommendation tasks
func (ma *ModelAgent) Process(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	ma.mutex.Lock()
	ma.status = StatusProcessing
	ma.mutex.Unlock()

	defer func() {
		ma.mutex.Lock()
		ma.status = StatusIdle
		ma.metrics.TasksProcessed++
		ma.mutex.Unlock()
	}()

	start := time.Now()

	result, err := ma.processTask(ctx, task)

	duration := time.Since(start)

	// Update metrics
	ma.mutex.Lock()
	if err == nil {
		ma.metrics.SuccessRate = float64(ma.metrics.TasksProcessed) / float64(ma.metrics.TasksProcessed+ma.metrics.ErrorCount+1)
	} else {
		ma.metrics.ErrorCount++
		ma.metrics.SuccessRate = float64(ma.metrics.TasksProcessed+1) / float64(ma.metrics.TasksProcessed+ma.metrics.ErrorCount+1)
	}

	// Update average latency
	if ma.metrics.AverageLatency == 0 {
		ma.metrics.AverageLatency = duration
	} else {
		ma.metrics.AverageLatency = (ma.metrics.AverageLatency + duration) / 2
	}
	ma.metrics.LastActiveTime = time.Now()
	ma.mutex.Unlock()

	if result != nil {
		result.Metrics = &TaskMetrics{
			ExecutionTime: duration,
			QualityScore:  0.92, // TODO: Calculate actual quality score
		}
	}

	return result, err
}

// processTask processes specific model-related tasks
func (ma *ModelAgent) processTask(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	switch task.Type {
	case TaskModelTraining:
		return ma.handleModelTraining(ctx, task)
	case TaskModelEvaluation:
		return ma.handleModelEvaluation(ctx, task)
	case TaskHyperParamTuning:
		return ma.handleHyperParameterTuning(ctx, task)
	case TaskModelDeployment:
		return ma.handleModelDeployment(ctx, task)
	default:
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("unsupported task type: %s", task.Type),
			CreatedAt: time.Now(),
		}, fmt.Errorf("unsupported task type: %s", task.Type)
	}
}

// handleModelTraining processes model training tasks
func (ma *ModelAgent) handleModelTraining(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	algorithm, ok := task.Parameters["algorithm"].(string)
	if !ok {
		algorithm = "collaborative_filtering" // Default algorithm
	}

	algo, exists := ma.algorithms[algorithm]
	if !exists {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("algorithm '%s' not supported", algorithm),
			CreatedAt: time.Now(),
		}, fmt.Errorf("algorithm '%s' not supported", algorithm)
	}

	// Generate mock training data
	trainingData := ma.generateMockTrainingData()

	// Create training parameters
	trainingParams := &TrainingParams{
		Algorithm:        algorithm,
		HyperParameters:  ma.extractHyperParameters(task.Parameters),
		MaxEpochs:        100,
		BatchSize:        64,
		LearningRate:     0.001,
		ValidationMetric: "rmse",
		EarlyStopping: &EarlyStoppingConfig{
			Enabled:  true,
			Patience: 10,
			MinDelta: 0.001,
			Metric:   "rmse",
			Mode:     "min",
		},
	}

	// Train the model
	trainedModel, err := algo.Train(ctx, trainingData, trainingParams)
	if err != nil {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("model training failed: %s", err.Error()),
			CreatedAt: time.Now(),
		}, err
	}

	// Register the model
	ma.registry.mutex.Lock()
	ma.models[trainedModel.ID] = trainedModel
	ma.registry.models[trainedModel.ID] = trainedModel
	ma.registry.mutex.Unlock()

	data := map[string]interface{}{
		"model_id":      trainedModel.ID,
		"algorithm":     trainedModel.Algorithm,
		"training_time": trainedModel.TrainingTime,
	}

	if trainedModel.TrainingMetrics != nil {
		data["final_loss"] = trainedModel.TrainingMetrics.FinalValidLoss
	}

	return &RecommendationResult{
		TaskID:    task.ID,
		Success:   true,
		Data:      data,
		CreatedAt: time.Now(),
	}, nil
}

// handleModelEvaluation processes model evaluation tasks
func (ma *ModelAgent) handleModelEvaluation(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	modelID, ok := task.Parameters["model_id"].(string)
	if !ok {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     "model_id parameter is required",
			CreatedAt: time.Now(),
		}, fmt.Errorf("model_id parameter is required")
	}

	ma.registry.mutex.RLock()
	model, exists := ma.models[modelID]
	ma.registry.mutex.RUnlock()

	if !exists {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("model '%s' not found", modelID),
			CreatedAt: time.Now(),
		}, fmt.Errorf("model '%s' not found", modelID)
	}

	// Generate evaluation results
	evalMetrics := ma.evaluateModel(ctx, model)

	// Update model with evaluation results
	ma.registry.mutex.Lock()
	model.EvalMetrics = evalMetrics
	ma.registry.mutex.Unlock()

	return &RecommendationResult{
		TaskID:  task.ID,
		Success: true,
		Data: map[string]interface{}{
			"model_id":           modelID,
			"evaluation_metrics": evalMetrics,
		},
		CreatedAt: time.Now(),
	}, nil
}

// handleHyperParameterTuning processes hyperparameter optimization tasks
func (ma *ModelAgent) handleHyperParameterTuning(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	algorithm, ok := task.Parameters["algorithm"].(string)
	if !ok {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     "algorithm parameter is required",
			CreatedAt: time.Now(),
		}, fmt.Errorf("algorithm parameter is required")
	}

	algo, exists := ma.algorithms[algorithm]
	if !exists {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("algorithm '%s' not supported", algorithm),
			CreatedAt: time.Now(),
		}, fmt.Errorf("algorithm '%s' not supported", algorithm)
	}

	// Perform hyperparameter optimization
	bestParams, bestScore := ma.optimizeHyperParameters(ctx, algo)

	return &RecommendationResult{
		TaskID:  task.ID,
		Success: true,
		Data: map[string]interface{}{
			"algorithm":           algorithm,
			"best_parameters":     bestParams,
			"best_score":          bestScore,
			"optimization_method": ma.optimizer.method,
		},
		CreatedAt: time.Now(),
	}, nil
}

// handleModelDeployment processes model deployment tasks
func (ma *ModelAgent) handleModelDeployment(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error) {
	modelID, ok := task.Parameters["model_id"].(string)
	if !ok {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     "model_id parameter is required",
			CreatedAt: time.Now(),
		}, fmt.Errorf("model_id parameter is required")
	}

	ma.registry.mutex.Lock()
	model, exists := ma.models[modelID]
	if exists {
		model.Status = ModelStatusDeployed
	}
	ma.registry.mutex.Unlock()

	if !exists {
		return &RecommendationResult{
			TaskID:    task.ID,
			Success:   false,
			Error:     fmt.Sprintf("model '%s' not found", modelID),
			CreatedAt: time.Now(),
		}, fmt.Errorf("model '%s' not found", modelID)
	}

	return &RecommendationResult{
		TaskID:  task.ID,
		Success: true,
		Data: map[string]interface{}{
			"model_id":          modelID,
			"deployment_status": "deployed",
			"deployment_time":   time.Now(),
		},
		CreatedAt: time.Now(),
	}, nil
}

// generateMockTrainingData generates sample training data for testing
func (ma *ModelAgent) generateMockTrainingData() *TrainingData {
	rand.Seed(time.Now().UnixNano())

	numUsers := 1000
	numItems := 500
	numInteractions := 5000

	userFeatures := make([]map[string]interface{}, numUsers)
	for i := 0; i < numUsers; i++ {
		userFeatures[i] = map[string]interface{}{
			"user_id":  fmt.Sprintf("user_%d", i),
			"age":      rand.Intn(60) + 18,
			"gender":   []string{"male", "female"}[rand.Intn(2)],
			"location": fmt.Sprintf("city_%d", rand.Intn(50)),
		}
	}

	itemFeatures := make([]map[string]interface{}, numItems)
	for i := 0; i < numItems; i++ {
		itemFeatures[i] = map[string]interface{}{
			"item_id":  fmt.Sprintf("item_%d", i),
			"category": fmt.Sprintf("category_%d", rand.Intn(20)),
			"price":    rand.Float64() * 100,
			"rating":   4.0 + rand.Float64(),
		}
	}

	interactions := make([]Interaction, numInteractions)
	for i := 0; i < numInteractions; i++ {
		interactions[i] = Interaction{
			UserID:    fmt.Sprintf("user_%d", rand.Intn(numUsers)),
			ItemID:    fmt.Sprintf("item_%d", rand.Intn(numItems)),
			Rating:    1.0 + rand.Float64()*4.0, // 1-5 rating
			Timestamp: time.Now().Add(-time.Duration(rand.Intn(365*24)) * time.Hour),
		}
	}

	return &TrainingData{
		ID:           fmt.Sprintf("training_data_%d", time.Now().UnixNano()),
		UserFeatures: userFeatures,
		ItemFeatures: itemFeatures,
		Interactions: interactions,
		SplitRatio: &DataSplit{
			TrainRatio:      0.7,
			ValidationRatio: 0.15,
			TestRatio:       0.15,
		},
		CreatedAt: time.Now(),
	}
}

// extractHyperParameters extracts hyperparameters from task parameters
func (ma *ModelAgent) extractHyperParameters(params map[string]interface{}) map[string]interface{} {
	hyperParams := make(map[string]interface{})

	if hp, ok := params["hyperparameters"]; ok {
		if hpMap, ok := hp.(map[string]interface{}); ok {
			return hpMap
		}
	}

	// Default hyperparameters
	hyperParams["learning_rate"] = 0.001
	hyperParams["regularization"] = 0.01
	hyperParams["embedding_dim"] = 64
	hyperParams["num_epochs"] = 100

	return hyperParams
}

// evaluateModel evaluates a trained model and returns metrics
func (ma *ModelAgent) evaluateModel(ctx context.Context, model *TrainedModel) *EvaluationMetrics {
	// Mock evaluation - in real implementation this would run actual evaluation
	rand.Seed(time.Now().UnixNano())

	return &EvaluationMetrics{
		RMSE: 0.8 + rand.Float64()*0.4, // 0.8-1.2
		MAE:  0.6 + rand.Float64()*0.3, // 0.6-0.9
		Precision: map[int]float64{
			5:  0.15 + rand.Float64()*0.1,  // 0.15-0.25
			10: 0.12 + rand.Float64()*0.08, // 0.12-0.20
			20: 0.08 + rand.Float64()*0.06, // 0.08-0.14
		},
		Recall: map[int]float64{
			5:  0.08 + rand.Float64()*0.05, // 0.08-0.13
			10: 0.15 + rand.Float64()*0.08, // 0.15-0.23
			20: 0.28 + rand.Float64()*0.12, // 0.28-0.40
		},
		NDCG: map[int]float64{
			5:  0.25 + rand.Float64()*0.1, // 0.25-0.35
			10: 0.30 + rand.Float64()*0.1, // 0.30-0.40
			20: 0.35 + rand.Float64()*0.1, // 0.35-0.45
		},
		AUC:       0.75 + rand.Float64()*0.2,  // 0.75-0.95
		Coverage:  0.60 + rand.Float64()*0.3,  // 0.60-0.90
		Diversity: 0.70 + rand.Float64()*0.2,  // 0.70-0.90
		Novelty:   0.65 + rand.Float64()*0.25, // 0.65-0.90
	}
}

// optimizeHyperParameters performs hyperparameter optimization
func (ma *ModelAgent) optimizeHyperParameters(ctx context.Context, algo RecommendationAlgorithm) (map[string]interface{}, float64) {
	// Mock hyperparameter optimization
	rand.Seed(time.Now().UnixNano())

	bestParams := map[string]interface{}{
		"learning_rate":  0.001 + rand.Float64()*0.009, // 0.001-0.01
		"regularization": 0.001 + rand.Float64()*0.049, // 0.001-0.05
		"embedding_dim":  32 + rand.Intn(97),           // 32-128
		"batch_size":     16 + rand.Intn(113),          // 16-128
	}

	bestScore := 0.70 + rand.Float64()*0.25 // 0.70-0.95

	ma.logger.WithFields(logrus.Fields{
		"algorithm":   algo.Name(),
		"best_params": bestParams,
		"best_score":  bestScore,
	}).Info("Hyperparameter optimization completed")

	return bestParams, bestScore
}

// UpdateConfiguration updates agent configuration
func (ma *ModelAgent) UpdateConfiguration(config map[string]interface{}) error {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()

	if maxTraining, ok := config["max_concurrent_training"].(int); ok {
		ma.config.MaxConcurrentTraining = maxTraining
		ma.trainer.maxConcurrent = maxTraining
	}

	if threshold, ok := config["auto_deploy_threshold"].(float64); ok {
		ma.config.AutoDeployThreshold = threshold
	}

	ma.logger.WithField("agent_id", ma.id).Info("ModelAgent configuration updated")
	return nil
}

// HealthCheck performs agent health verification
func (ma *ModelAgent) HealthCheck() error {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	if ma.status == StatusError {
		return fmt.Errorf("agent is in error state")
	}

	if len(ma.algorithms) == 0 {
		return fmt.Errorf("no algorithms registered")
	}

	return nil
}

// GetPerformanceStats returns detailed performance statistics
func (ma *ModelAgent) GetPerformanceStats() *PerformanceStats {
	ma.mutex.RLock()
	defer ma.mutex.RUnlock()

	uptime := time.Since(ma.startTime)
	successfulTasks := int64(float64(ma.metrics.TasksProcessed) * ma.metrics.SuccessRate)
	failedTasks := ma.metrics.TasksProcessed - successfulTasks

	return &PerformanceStats{
		Uptime:          uptime,
		TotalTasks:      ma.metrics.TasksProcessed,
		SuccessfulTasks: successfulTasks,
		FailedTasks:     failedTasks,
		AverageLatency:  ma.metrics.AverageLatency,
		P95Latency:      ma.metrics.AverageLatency * 135 / 100, // Estimated
		P99Latency:      ma.metrics.AverageLatency * 180 / 100, // Estimated
		ThroughputQPS:   float64(ma.metrics.TasksProcessed) / uptime.Seconds(),
		ErrorRate:       float64(failedTasks) / float64(ma.metrics.TasksProcessed),
	}
}

// initializeAlgorithms initializes built-in recommendation algorithms
func (ma *ModelAgent) initializeAlgorithms() error {
	// Initialize mock algorithms
	ma.algorithms["collaborative_filtering"] = &MockCollaborativeFiltering{}
	ma.algorithms["content_based"] = &MockContentBased{}
	ma.algorithms["matrix_factorization"] = &MockMatrixFactorization{}
	ma.algorithms["deep_learning"] = &MockDeepLearning{}

	ma.logger.WithField("algorithms", len(ma.algorithms)).Info("Recommendation algorithms initialized")
	return nil
}

// GetModels returns all registered models
func (ma *ModelAgent) GetModels() map[string]*TrainedModel {
	ma.registry.mutex.RLock()
	defer ma.registry.mutex.RUnlock()

	result := make(map[string]*TrainedModel)
	for id, model := range ma.models {
		modelCopy := *model
		result[id] = &modelCopy
	}

	return result
}

// Mock Algorithm Implementations

// MockCollaborativeFiltering implements collaborative filtering algorithm
type MockCollaborativeFiltering struct{}

func (mcf *MockCollaborativeFiltering) Name() string {
	return "collaborative_filtering"
}

func (mcf *MockCollaborativeFiltering) Train(ctx context.Context, trainingData *TrainingData, params *TrainingParams) (*TrainedModel, error) {
	// Mock training simulation
	start := time.Now()

	// Simulate training time
	time.Sleep(time.Millisecond * 500)

	trainingTime := time.Since(start)

	model := &TrainedModel{
		ID:         fmt.Sprintf("model_cf_%d", time.Now().UnixNano()),
		Name:       "Collaborative Filtering Model",
		Algorithm:  "collaborative_filtering",
		Version:    "1.0.0",
		Parameters: params.HyperParameters,
		TrainingMetrics: &TrainingMetrics{
			TrainLoss:      []float64{1.2, 1.0, 0.9, 0.85, 0.82},
			ValidLoss:      []float64{1.3, 1.1, 0.95, 0.88, 0.86},
			Epochs:         5,
			TrainingTime:   trainingTime,
			ConvergedEpoch: 5,
			FinalTrainLoss: 0.82,
			FinalValidLoss: 0.86,
		},
		Status:       ModelStatusTrained,
		CreatedAt:    time.Now(),
		TrainingTime: trainingTime,
		Metadata: map[string]interface{}{
			"num_users":        len(trainingData.UserFeatures),
			"num_items":        len(trainingData.ItemFeatures),
			"num_interactions": len(trainingData.Interactions),
		},
	}

	return model, nil
}

func (mcf *MockCollaborativeFiltering) Predict(ctx context.Context, model *TrainedModel, input *PredictionInput) (*PredictionOutput, error) {
	// Mock prediction
	rand.Seed(time.Now().UnixNano())

	recommendations := make([]RecommendationItem, input.TopK)
	for i := 0; i < input.TopK; i++ {
		recommendations[i] = RecommendationItem{
			ItemID:     fmt.Sprintf("item_%d", rand.Intn(1000)),
			Score:      rand.Float64(),
			Rank:       i + 1,
			Confidence: 0.7 + rand.Float64()*0.3,
		}
	}

	return &PredictionOutput{
		UserID:          input.UserID,
		Recommendations: recommendations,
		ModelID:         model.ID,
		Timestamp:       time.Now(),
	}, nil
}

func (mcf *MockCollaborativeFiltering) GetHyperParameters() map[string]HyperParameter {
	return map[string]HyperParameter{
		"num_factors": {
			Name:        "num_factors",
			Type:        "int",
			Min:         10,
			Max:         200,
			Default:     64,
			Description: "Number of latent factors",
		},
		"learning_rate": {
			Name:        "learning_rate",
			Type:        "float",
			Min:         0.0001,
			Max:         0.1,
			Default:     0.01,
			Description: "Learning rate for optimization",
		},
	}
}

func (mcf *MockCollaborativeFiltering) GetMetrics() *AlgorithmMetrics {
	return &AlgorithmMetrics{
		TrainingTime:   2 * time.Minute,
		PredictionTime: 100 * time.Millisecond,
		MemoryUsage:    1024 * 1024, // 1MB
		ModelSize:      512 * 1024,  // 512KB
		Accuracy:       0.85,
		LastUpdated:    time.Now().Add(-1 * time.Hour),
	}
}

// Additional mock algorithms would follow similar pattern...
type MockContentBased struct{}
type MockMatrixFactorization struct{}
type MockDeepLearning struct{}

// Implement the interface methods for other mock algorithms (shortened for brevity)
func (mcb *MockContentBased) Name() string { return "content_based" }
func (mcb *MockContentBased) Train(ctx context.Context, trainingData *TrainingData, params *TrainingParams) (*TrainedModel, error) {
	start := time.Now()
	time.Sleep(time.Millisecond * 100)
	trainingTime := time.Since(start)

	return &TrainedModel{
		ID:           "content_model",
		Algorithm:    "content_based",
		Status:       ModelStatusTrained,
		CreatedAt:    time.Now(),
		TrainingTime: trainingTime,
		TrainingMetrics: &TrainingMetrics{
			TrainLoss:      []float64{1.5, 1.2, 1.0, 0.9, 0.85},
			ValidLoss:      []float64{1.6, 1.3, 1.1, 0.95, 0.90},
			Epochs:         5,
			TrainingTime:   trainingTime,
			FinalTrainLoss: 0.85,
			FinalValidLoss: 0.90,
		},
	}, nil
}
func (mcb *MockContentBased) Predict(ctx context.Context, model *TrainedModel, input *PredictionInput) (*PredictionOutput, error) {
	return &PredictionOutput{UserID: input.UserID, ModelID: model.ID, Timestamp: time.Now()}, nil
}
func (mcb *MockContentBased) GetHyperParameters() map[string]HyperParameter {
	return make(map[string]HyperParameter)
}
func (mcb *MockContentBased) GetMetrics() *AlgorithmMetrics { return &AlgorithmMetrics{} }

func (mmf *MockMatrixFactorization) Name() string { return "matrix_factorization" }
func (mmf *MockMatrixFactorization) Train(ctx context.Context, trainingData *TrainingData, params *TrainingParams) (*TrainedModel, error) {
	start := time.Now()
	time.Sleep(time.Millisecond * 100)
	trainingTime := time.Since(start)

	return &TrainedModel{
		ID:           "mf_model",
		Algorithm:    "matrix_factorization",
		Status:       ModelStatusTrained,
		CreatedAt:    time.Now(),
		TrainingTime: trainingTime,
		TrainingMetrics: &TrainingMetrics{
			TrainLoss:      []float64{1.8, 1.4, 1.1, 0.95, 0.88},
			ValidLoss:      []float64{1.9, 1.5, 1.2, 1.0, 0.92},
			Epochs:         5,
			TrainingTime:   trainingTime,
			FinalTrainLoss: 0.88,
			FinalValidLoss: 0.92,
		},
	}, nil
}
func (mmf *MockMatrixFactorization) Predict(ctx context.Context, model *TrainedModel, input *PredictionInput) (*PredictionOutput, error) {
	return &PredictionOutput{UserID: input.UserID, ModelID: model.ID, Timestamp: time.Now()}, nil
}
func (mmf *MockMatrixFactorization) GetHyperParameters() map[string]HyperParameter {
	return make(map[string]HyperParameter)
}
func (mmf *MockMatrixFactorization) GetMetrics() *AlgorithmMetrics { return &AlgorithmMetrics{} }

func (mdl *MockDeepLearning) Name() string { return "deep_learning" }
func (mdl *MockDeepLearning) Train(ctx context.Context, trainingData *TrainingData, params *TrainingParams) (*TrainedModel, error) {
	start := time.Now()
	time.Sleep(time.Millisecond * 100)
	trainingTime := time.Since(start)

	return &TrainedModel{
		ID:           "dl_model",
		Algorithm:    "deep_learning",
		Status:       ModelStatusTrained,
		CreatedAt:    time.Now(),
		TrainingTime: trainingTime,
		TrainingMetrics: &TrainingMetrics{
			TrainLoss:      []float64{2.1, 1.6, 1.2, 0.9, 0.78},
			ValidLoss:      []float64{2.2, 1.7, 1.3, 0.95, 0.82},
			Epochs:         5,
			TrainingTime:   trainingTime,
			FinalTrainLoss: 0.78,
			FinalValidLoss: 0.82,
		},
	}, nil
}
func (mdl *MockDeepLearning) Predict(ctx context.Context, model *TrainedModel, input *PredictionInput) (*PredictionOutput, error) {
	return &PredictionOutput{UserID: input.UserID, ModelID: model.ID, Timestamp: time.Now()}, nil
}
func (mdl *MockDeepLearning) GetHyperParameters() map[string]HyperParameter {
	return make(map[string]HyperParameter)
}
func (mdl *MockDeepLearning) GetMetrics() *AlgorithmMetrics { return &AlgorithmMetrics{} }
