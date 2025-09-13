package recommendation

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// RecommendationAgentType defines specialized agent roles for recommendation business
type RecommendationAgentType string

const (
	AgentTypeData    RecommendationAgentType = "data_agent"    // 数据采集和特征工程
	AgentTypeModel   RecommendationAgentType = "model_agent"   // 模型训练和优化
	AgentTypeService RecommendationAgentType = "service_agent" // 实时推荐服务
	AgentTypeEval    RecommendationAgentType = "eval_agent"    // A/B测试和效果评估
)

// RecommendationAgent defines the interface for recommendation business agents
type RecommendationAgent interface {
	// 基础Agent接口
	GetID() string
	GetType() RecommendationAgentType
	GetStatus() AgentStatus
	GetMetrics() *AgentMetrics

	// 推荐业务专用接口
	Process(ctx context.Context, task *RecommendationTask) (*RecommendationResult, error)
	GetCapabilities() []string
	UpdateConfiguration(config map[string]interface{}) error

	// 健康检查和监控
	HealthCheck() error
	GetPerformanceStats() *PerformanceStats
}

// AgentStatus represents agent operational status
type AgentStatus string

const (
	StatusIdle        AgentStatus = "idle"
	StatusProcessing  AgentStatus = "processing"
	StatusError       AgentStatus = "error"
	StatusMaintenance AgentStatus = "maintenance"
)

// RecommendationTask represents a task for recommendation agents
type RecommendationTask struct {
	ID         string                 `json:"id"`
	Type       TaskType               `json:"type"`
	Priority   TaskPriority           `json:"priority"`
	Parameters map[string]interface{} `json:"parameters"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	Deadline   *time.Time             `json:"deadline,omitempty"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
}

// TaskType defines different types of recommendation tasks
type TaskType string

const (
	// DataAgent tasks
	TaskDataCollection     TaskType = "data_collection"     // 数据采集
	TaskFeatureEngineering TaskType = "feature_engineering" // 特征工程
	TaskDataCleaning       TaskType = "data_cleaning"       // 数据清洗
	TaskDataValidation     TaskType = "data_validation"     // 数据验证

	// ModelAgent tasks
	TaskModelTraining    TaskType = "model_training"    // 模型训练
	TaskModelEvaluation  TaskType = "model_evaluation"  // 模型评估
	TaskHyperParamTuning TaskType = "hyperparam_tuning" // 超参数调优
	TaskModelDeployment  TaskType = "model_deployment"  // 模型部署

	// ServiceAgent tasks
	TaskRealTimeInference TaskType = "realtime_inference" // 实时推理
	TaskCacheManagement   TaskType = "cache_management"   // 缓存管理
	TaskLoadBalancing     TaskType = "load_balancing"     // 负载均衡
	TaskServiceMonitoring TaskType = "service_monitoring" // 服务监控

	// EvalAgent tasks
	TaskABTesting        TaskType = "ab_testing"        // A/B测试
	TaskMetricsAnalysis  TaskType = "metrics_analysis"  // 指标分析
	TaskEffectEvaluation TaskType = "effect_evaluation" // 效果评估
	TaskReportGeneration TaskType = "report_generation" // 报告生成
)

// TaskPriority defines task execution priority
type TaskPriority int

const (
	PriorityLow      TaskPriority = 1
	PriorityMedium   TaskPriority = 2
	PriorityHigh     TaskPriority = 3
	PriorityCritical TaskPriority = 4
)

// RecommendationResult represents the result of a recommendation task
type RecommendationResult struct {
	TaskID    string                 `json:"task_id"`
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data"`
	Error     string                 `json:"error,omitempty"`
	Metrics   *TaskMetrics           `json:"metrics"`
	CreatedAt time.Time              `json:"created_at"`
}

// AgentMetrics tracks agent performance and health
type AgentMetrics struct {
	TasksProcessed int64          `json:"tasks_processed"`
	SuccessRate    float64        `json:"success_rate"`
	AverageLatency time.Duration  `json:"average_latency"`
	ErrorCount     int64          `json:"error_count"`
	LastActiveTime time.Time      `json:"last_active_time"`
	ResourceUsage  *ResourceUsage `json:"resource_usage"`
}

// ResourceUsage tracks resource consumption
type ResourceUsage struct {
	CPUPercent   float64 `json:"cpu_percent"`
	MemoryMB     int64   `json:"memory_mb"`
	DiskUsageMB  int64   `json:"disk_usage_mb"`
	NetworkInMB  int64   `json:"network_in_mb"`
	NetworkOutMB int64   `json:"network_out_mb"`
}

// TaskMetrics tracks individual task performance
type TaskMetrics struct {
	ExecutionTime time.Duration  `json:"execution_time"`
	ResourceUsed  *ResourceUsage `json:"resource_used"`
	QualityScore  float64        `json:"quality_score"`
	ThroughputQPS float64        `json:"throughput_qps"`
}

// PerformanceStats provides detailed performance analytics
type PerformanceStats struct {
	Uptime           time.Duration `json:"uptime"`
	TotalTasks       int64         `json:"total_tasks"`
	SuccessfulTasks  int64         `json:"successful_tasks"`
	FailedTasks      int64         `json:"failed_tasks"`
	AverageLatency   time.Duration `json:"average_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	ThroughputQPS    float64       `json:"throughput_qps"`
	ErrorRate        float64       `json:"error_rate"`
	LastErrorTime    *time.Time    `json:"last_error_time,omitempty"`
	LastErrorMessage string        `json:"last_error_message,omitempty"`
}

// RecommendationOrchestrator manages the lifecycle and coordination of recommendation agents
type RecommendationOrchestrator struct {
	agents      map[string]RecommendationAgent
	taskQueue   chan *RecommendationTask
	resultQueue chan *RecommendationResult
	logger      *logrus.Logger
	config      *OrchestratorConfig
}

// OrchestratorConfig defines orchestrator configuration
type OrchestratorConfig struct {
	MaxConcurrentTasks  int           `json:"max_concurrent_tasks"`
	TaskTimeout         time.Duration `json:"task_timeout"`
	RetryPolicy         *RetryPolicy  `json:"retry_policy"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval     time.Duration `json:"metrics_interval"`
}

// RetryPolicy defines task retry behavior
type RetryPolicy struct {
	MaxRetries        int           `json:"max_retries"`
	InitialDelay      time.Duration `json:"initial_delay"`
	BackoffMultiplier float64       `json:"backoff_multiplier"`
	MaxDelay          time.Duration `json:"max_delay"`
}

// PredictionInput represents input for prediction
type PredictionInput struct {
	UserID  string                 `json:"user_id"`
	ItemIDs []string               `json:"item_ids"`
	Context map[string]interface{} `json:"context"`
	TopK    int                    `json:"top_k"`
}

// PredictionOutput represents prediction results
type PredictionOutput struct {
	UserID          string               `json:"user_id"`
	Predictions     []Prediction         `json:"predictions"`
	Recommendations []RecommendationItem `json:"recommendations"`
	Algorithm       string               `json:"algorithm"`
	ModelID         string               `json:"model_id"`
	Timestamp       time.Time            `json:"timestamp"`
}

// Prediction represents a single item prediction
type Prediction struct {
	ItemID     string  `json:"item_id"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// RecommendationItem represents a recommended item
type RecommendationItem struct {
	ItemID     string                 `json:"item_id"`
	Score      float64                `json:"score"`
	Rank       int                    `json:"rank"`
	Reason     string                 `json:"reason,omitempty"`
	Features   map[string]interface{} `json:"features,omitempty"`
	Confidence float64                `json:"confidence"`
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalSize     int64     `json:"total_size"`
	UsedSize      int64     `json:"used_size"`
	AvailableSize int64     `json:"available_size"`
	UserCount     int64     `json:"user_count"`
	MovieCount    int64     `json:"movie_count"`
	RatingCount   int64     `json:"rating_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// DeleteCriteria represents criteria for data deletion
type DeleteCriteria struct {
	Type   string                 `json:"type"`
	Filter map[string]interface{} `json:"filter"`
}

// DataQuery represents a query for data retrieval
type DataQuery struct {
	Type      string                 `json:"type"`
	Table     string                 `json:"table"`
	Fields    []string               `json:"fields"`
	Filters   map[string]interface{} `json:"filters"`
	OrderBy   []string               `json:"order_by"`
	Limit     int                    `json:"limit"`
	Offset    int                    `json:"offset"`
	TimeRange *TimeRange             `json:"time_range"`
}

// TimeRange defines time-based filtering
type TimeRange struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// ProcessingStats represents data processing statistics
type ProcessingStats struct {
	RecordsProcessed int64         `json:"records_processed"`
	RecordsFiltered  int64         `json:"records_filtered"`
	ProcessingTime   time.Duration `json:"processing_time"`
	ErrorCount       int64         `json:"error_count"`
	ThroughputRPS    float64       `json:"throughput_rps"`
}

// AlertHandler handles data quality alerts
type AlertHandler interface {
	Handle(alert *QualityAlert) error
}

// QualityAlert represents a data quality alert
type QualityAlert struct {
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}

// QualityIssue represents a data quality problem (already exists in data_agent.go)
// Moved to avoid duplication

// ModelStatus represents model training/deployment status
type ModelStatus string

const (
	ModelStatusTraining  ModelStatus = "training"
	ModelStatusTrained   ModelStatus = "trained"
	ModelStatusDeploying ModelStatus = "deploying"
	ModelStatusDeployed  ModelStatus = "deployed"
	ModelStatusFailed    ModelStatus = "failed"
	ModelStatusRetired   ModelStatus = "retired"
)

// HyperParameter represents a hyperparameter definition
type HyperParameter struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // "int", "float", "string", "bool"
	Default     interface{}   `json:"default"`
	Min         interface{}   `json:"min,omitempty"`
	Max         interface{}   `json:"max,omitempty"`
	Options     []interface{} `json:"options,omitempty"`
	Description string        `json:"description"`
}

// AlgorithmMetrics represents algorithm performance metrics
type AlgorithmMetrics struct {
	TrainingTime   time.Duration `json:"training_time"`
	PredictionTime time.Duration `json:"prediction_time"`
	MemoryUsage    int64         `json:"memory_usage"`
	ModelSize      int64         `json:"model_size"`
	Accuracy       float64       `json:"accuracy"`
	LastUpdated    time.Time     `json:"last_updated"`
}
