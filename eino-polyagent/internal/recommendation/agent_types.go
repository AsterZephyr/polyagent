package recommendation

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// RecommendationAgentType defines specialized agent roles for recommendation business
type RecommendationAgentType string

const (
	AgentTypeData    RecommendationAgentType = "data_agent"     // 数据采集和特征工程
	AgentTypeModel   RecommendationAgentType = "model_agent"    // 模型训练和优化
	AgentTypeService RecommendationAgentType = "service_agent"  // 实时推荐服务
	AgentTypeEval    RecommendationAgentType = "eval_agent"     // A/B测试和效果评估
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
	StatusIdle       AgentStatus = "idle"
	StatusProcessing AgentStatus = "processing"
	StatusError      AgentStatus = "error"
	StatusMaintenance AgentStatus = "maintenance"
)

// RecommendationTask represents a task for recommendation agents
type RecommendationTask struct {
	ID          string                 `json:"id"`
	Type        TaskType               `json:"type"`
	Priority    TaskPriority           `json:"priority"`
	Parameters  map[string]interface{} `json:"parameters"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	Deadline    *time.Time             `json:"deadline,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
}

// TaskType defines different types of recommendation tasks
type TaskType string

const (
	// DataAgent tasks
	TaskDataCollection   TaskType = "data_collection"   // 数据采集
	TaskFeatureEngineering TaskType = "feature_engineering" // 特征工程
	TaskDataCleaning     TaskType = "data_cleaning"     // 数据清洗
	TaskDataValidation   TaskType = "data_validation"   // 数据验证
	
	// ModelAgent tasks  
	TaskModelTraining    TaskType = "model_training"    // 模型训练
	TaskModelEvaluation  TaskType = "model_evaluation"  // 模型评估
	TaskHyperParamTuning TaskType = "hyperparam_tuning" // 超参数调优
	TaskModelDeployment  TaskType = "model_deployment"  // 模型部署
	
	// ServiceAgent tasks
	TaskRealTimeInference TaskType = "realtime_inference" // 实时推理
	TaskCacheManagement  TaskType = "cache_management"   // 缓存管理
	TaskLoadBalancing    TaskType = "load_balancing"     // 负载均衡
	TaskServiceMonitoring TaskType = "service_monitoring" // 服务监控
	
	// EvalAgent tasks
	TaskABTesting        TaskType = "ab_testing"         // A/B测试
	TaskMetricsAnalysis  TaskType = "metrics_analysis"   // 指标分析
	TaskEffectEvaluation TaskType = "effect_evaluation"  // 效果评估
	TaskReportGeneration TaskType = "report_generation"  // 报告生成
)

// TaskPriority defines task execution priority
type TaskPriority int

const (
	PriorityLow    TaskPriority = 1
	PriorityMedium TaskPriority = 2
	PriorityHigh   TaskPriority = 3
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
	TasksProcessed   int64         `json:"tasks_processed"`
	SuccessRate      float64       `json:"success_rate"`
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorCount       int64         `json:"error_count"`
	LastActiveTime   time.Time     `json:"last_active_time"`
	ResourceUsage    *ResourceUsage `json:"resource_usage"`
}

// ResourceUsage tracks resource consumption
type ResourceUsage struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryMB      int64   `json:"memory_mb"`
	DiskUsageMB   int64   `json:"disk_usage_mb"`
	NetworkInMB   int64   `json:"network_in_mb"`
	NetworkOutMB  int64   `json:"network_out_mb"`
}

// TaskMetrics tracks individual task performance
type TaskMetrics struct {
	ExecutionTime time.Duration `json:"execution_time"`
	ResourceUsed  *ResourceUsage `json:"resource_used"`
	QualityScore  float64       `json:"quality_score"`
	ThroughputQPS float64       `json:"throughput_qps"`
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
	MaxConcurrentTasks int           `json:"max_concurrent_tasks"`
	TaskTimeout        time.Duration `json:"task_timeout"`
	RetryPolicy        *RetryPolicy  `json:"retry_policy"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	MetricsInterval    time.Duration `json:"metrics_interval"`
}

// RetryPolicy defines task retry behavior
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	BackoffMultiplier float64     `json:"backoff_multiplier"`
	MaxDelay        time.Duration `json:"max_delay"`
}