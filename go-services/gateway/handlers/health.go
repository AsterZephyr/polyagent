package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/go-services/internal/ai"
	"github.com/polyagent/go-services/internal/storage"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	postgres *storage.PostgresStorage
	redis    *storage.RedisStorage
	aiClient *ai.PythonAIClient
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(postgres *storage.PostgresStorage, redis *storage.RedisStorage, aiClient *ai.PythonAIClient) *HealthHandler {
	return &HealthHandler{
		postgres: postgres,
		redis:    redis,
		aiClient: aiClient,
	}
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
	System    SystemInfo        `json:"system"`
}

// SystemInfo 系统信息
type SystemInfo struct {
	GoVersion   string `json:"go_version"`
	NumCPU      int    `json:"num_cpu"`
	NumGoroutine int   `json:"num_goroutine"`
	MemAlloc    uint64 `json:"mem_alloc"`
	MemSys      uint64 `json:"mem_sys"`
}

// MetricsInfo 指标信息
type MetricsInfo struct {
	Uptime        time.Duration     `json:"uptime"`
	RequestCount  int64             `json:"request_count"`
	ErrorCount    int64             `json:"error_count"`
	ActiveSessions int64            `json:"active_sessions"`
	QueueLength   int64             `json:"queue_length"`
	Services      map[string]string `json:"services"`
	System        SystemInfo        `json:"system"`
}

var startTime = time.Now()

// HealthCheck 健康检查
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0", // 这个应该从构建时注入
		Services:  make(map[string]string),
		System:    h.getSystemInfo(),
	}

	// 检查 PostgreSQL
	if err := h.checkPostgreSQL(); err != nil {
		status.Services["postgresql"] = "unhealthy"
		status.Status = "degraded"
	} else {
		status.Services["postgresql"] = "healthy"
	}

	// 检查 Redis
	if err := h.checkRedis(); err != nil {
		status.Services["redis"] = "unhealthy"
		status.Status = "degraded"
	} else {
		status.Services["redis"] = "healthy"
	}

	// 检查 Python AI 服务
	if err := h.checkPythonAI(); err != nil {
		status.Services["python_ai"] = "unhealthy"
		status.Status = "degraded"
	} else {
		status.Services["python_ai"] = "healthy"
	}

	// 根据整体状态返回相应的HTTP状态码
	httpStatus := http.StatusOK
	if status.Status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, status)
}

// Metrics 系统指标
func (h *HealthHandler) Metrics(c *gin.Context) {
	// 获取队列长度
	queueLength, _ := h.redis.GetQueueLength()

	// 获取请求统计
	today := time.Now().Format("2006-01-02")
	requestCount, _ := h.redis.GetCounter("stats:requests:total:" + today)
	errorCount, _ := h.redis.GetCounter("stats:requests:error:" + today)

	metrics := MetricsInfo{
		Uptime:         time.Since(startTime),
		RequestCount:   requestCount,
		ErrorCount:     errorCount,
		ActiveSessions: 0, // 这里应该实现活跃会话统计
		QueueLength:    queueLength,
		Services:       make(map[string]string),
		System:         h.getSystemInfo(),
	}

	// 服务状态
	metrics.Services["postgresql"] = h.getServiceStatus(h.checkPostgreSQL())
	metrics.Services["redis"] = h.getServiceStatus(h.checkRedis())
	metrics.Services["python_ai"] = h.getServiceStatus(h.checkPythonAI())

	c.JSON(http.StatusOK, metrics)
}

// checkPostgreSQL 检查 PostgreSQL 连接
func (h *HealthHandler) checkPostgreSQL() error {
	// 这里应该实现数据库连接检查
	// 例如：执行一个简单的查询
	return nil // 简化实现
}

// checkRedis 检查 Redis 连接
func (h *HealthHandler) checkRedis() error {
	// 执行 PING 命令检查连接
	_, err := h.redis.GetCounter("health_check")
	if err != nil {
		return err
	}
	return nil
}

// checkPythonAI 检查 Python AI 服务
func (h *HealthHandler) checkPythonAI() error {
	return h.aiClient.HealthCheck()
}

// getSystemInfo 获取系统信息
func (h *HealthHandler) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		GoVersion:    runtime.Version(),
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemAlloc:     m.Alloc,
		MemSys:       m.Sys,
	}
}

// getServiceStatus 获取服务状态字符串
func (h *HealthHandler) getServiceStatus(err error) string {
	if err != nil {
		return "unhealthy"
	}
	return "healthy"
}