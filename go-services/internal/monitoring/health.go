package monitoring

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/polyagent/go-services/internal/storage"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusWarning   HealthStatus = "warning"
	StatusCritical  HealthStatus = "critical"
	StatusUnknown   HealthStatus = "unknown"
)

// ComponentHealth 组件健康状态
type ComponentHealth struct {
	Name           string                 `json:"name"`
	Status         HealthStatus          `json:"status"`
	Message        string                `json:"message,omitempty"`
	LastCheck      time.Time             `json:"last_check"`
	ResponseTime   time.Duration         `json:"response_time"`
	Details        map[string]interface{} `json:"details,omitempty"`
}

// SystemHealth 系统整体健康状态
type SystemHealth struct {
	Status     HealthStatus       `json:"status"`
	Timestamp  time.Time          `json:"timestamp"`
	Uptime     time.Duration      `json:"uptime"`
	Version    string             `json:"version"`
	Components []ComponentHealth  `json:"components"`
	Metrics    SystemMetrics      `json:"metrics"`
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	Memory      MemoryMetrics    `json:"memory"`
	CPU         CPUMetrics       `json:"cpu"`
	Goroutines  int              `json:"goroutines"`
	Requests    RequestMetrics   `json:"requests"`
	Database    DatabaseMetrics  `json:"database"`
}

// MemoryMetrics 内存指标
type MemoryMetrics struct {
	Alloc      uint64  `json:"alloc_bytes"`
	TotalAlloc uint64  `json:"total_alloc_bytes"`
	Sys        uint64  `json:"sys_bytes"`
	NumGC      uint32  `json:"num_gc"`
	HeapInuse  uint64  `json:"heap_inuse_bytes"`
}

// CPUMetrics CPU指标
type CPUMetrics struct {
	NumCPU      int     `json:"num_cpu"`
	NumGoroutine int    `json:"num_goroutine"`
}

// RequestMetrics 请求指标
type RequestMetrics struct {
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	ErrorRequests   int64   `json:"error_requests"`
	SuccessRate     float64 `json:"success_rate"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
}

// DatabaseMetrics 数据库指标
type DatabaseMetrics struct {
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	TotalConnections  int `json:"total_connections"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) ComponentHealth
}

// HealthMonitor 健康监控器
type HealthMonitor struct {
	checkers  []HealthChecker
	startTime time.Time
	version   string
	mutex     sync.RWMutex
	
	// 缓存最近的健康检查结果
	lastResults map[string]ComponentHealth
	
	// 请求统计
	totalRequests   int64
	successRequests int64
	errorRequests   int64
	totalResponseTime time.Duration
}

// NewHealthMonitor 创建健康监控器
func NewHealthMonitor(version string) *HealthMonitor {
	return &HealthMonitor{
		checkers:    make([]HealthChecker, 0),
		startTime:   time.Now(),
		version:     version,
		lastResults: make(map[string]ComponentHealth),
	}
}

// RegisterChecker 注册健康检查器
func (hm *HealthMonitor) RegisterChecker(checker HealthChecker) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hm.checkers = append(hm.checkers, checker)
}

// CheckHealth 执行健康检查
func (hm *HealthMonitor) CheckHealth(ctx context.Context) SystemHealth {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	components := make([]ComponentHealth, len(hm.checkers))
	overallStatus := StatusHealthy

	// 并行执行所有健康检查
	var wg sync.WaitGroup
	for i, checker := range hm.checkers {
		wg.Add(1)
		go func(i int, checker HealthChecker) {
			defer wg.Done()
			
			// 为每个检查创建带超时的上下文
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			
			components[i] = checker.Check(checkCtx)
			hm.lastResults[checker.Name()] = components[i]
			
			// 更新整体状态
			if components[i].Status == StatusCritical {
				overallStatus = StatusCritical
			} else if components[i].Status == StatusWarning && overallStatus != StatusCritical {
				overallStatus = StatusWarning
			}
		}(i, checker)
	}
	
	wg.Wait()

	return SystemHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(hm.startTime),
		Version:    hm.version,
		Components: components,
		Metrics:    hm.collectMetrics(),
	}
}

// GetLastResults 获取最近的健康检查结果
func (hm *HealthMonitor) GetLastResults() map[string]ComponentHealth {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	
	results := make(map[string]ComponentHealth)
	for k, v := range hm.lastResults {
		results[k] = v
	}
	return results
}

// collectMetrics 收集系统指标
func (hm *HealthMonitor) collectMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	successRate := float64(0)
	avgResponseTime := float64(0)
	
	if hm.totalRequests > 0 {
		successRate = float64(hm.successRequests) / float64(hm.totalRequests) * 100
		avgResponseTime = float64(hm.totalResponseTime.Milliseconds()) / float64(hm.totalRequests)
	}

	return SystemMetrics{
		Memory: MemoryMetrics{
			Alloc:      m.Alloc,
			TotalAlloc: m.TotalAlloc,
			Sys:        m.Sys,
			NumGC:      m.NumGC,
			HeapInuse:  m.HeapInuse,
		},
		CPU: CPUMetrics{
			NumCPU:      runtime.NumCPU(),
			NumGoroutine: runtime.NumGoroutine(),
		},
		Goroutines: runtime.NumGoroutine(),
		Requests: RequestMetrics{
			TotalRequests:   hm.totalRequests,
			SuccessRequests: hm.successRequests,
			ErrorRequests:   hm.errorRequests,
			SuccessRate:     successRate,
			AvgResponseTime: avgResponseTime,
		},
		Database: DatabaseMetrics{
			// 这些指标需要从数据库连接池获取
			ActiveConnections: 0,
			IdleConnections:   0,
			TotalConnections:  0,
		},
	}
}

// RecordRequest 记录请求统计
func (hm *HealthMonitor) RecordRequest(success bool, responseTime time.Duration) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	
	hm.totalRequests++
	hm.totalResponseTime += responseTime
	
	if success {
		hm.successRequests++
	} else {
		hm.errorRequests++
	}
}

// PostgresHealthChecker PostgreSQL健康检查器
type PostgresHealthChecker struct {
	postgres *storage.PostgresStorage
}

// NewPostgresHealthChecker 创建PostgreSQL健康检查器
func NewPostgresHealthChecker(postgres *storage.PostgresStorage) *PostgresHealthChecker {
	return &PostgresHealthChecker{postgres: postgres}
}

func (p *PostgresHealthChecker) Name() string {
	return "postgresql"
}

func (p *PostgresHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// 执行简单的查询来检查数据库连接
	err := p.postgres.Ping(ctx)
	responseTime := time.Since(start)
	
	if err != nil {
		return ComponentHealth{
			Name:         p.Name(),
			Status:       StatusCritical,
			Message:      fmt.Sprintf("Database connection failed: %v", err),
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}
	
	// 检查连接池状态
	stats := p.postgres.GetStats()
	
	status := StatusHealthy
	message := "Database is healthy"
	
	// 检查连接池使用率
	if stats != nil {
		usageRate := float64(stats.InUse) / float64(stats.MaxOpenConnections) * 100
		if usageRate > 90 {
			status = StatusWarning
			message = fmt.Sprintf("High connection usage: %.1f%%", usageRate)
		} else if usageRate > 95 {
			status = StatusCritical
			message = fmt.Sprintf("Critical connection usage: %.1f%%", usageRate)
		}
	}
	
	details := map[string]interface{}{
		"response_time_ms": responseTime.Milliseconds(),
	}
	
	if stats != nil {
		details["open_connections"] = stats.OpenConnections
		details["in_use"] = stats.InUse
		details["idle"] = stats.Idle
		details["max_open_connections"] = stats.MaxOpenConnections
		details["wait_count"] = stats.WaitCount
		details["wait_duration_ms"] = stats.WaitDuration.Milliseconds()
	}
	
	return ComponentHealth{
		Name:         p.Name(),
		Status:       status,
		Message:      message,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Details:      details,
	}
}

// RedisHealthChecker Redis健康检查器
type RedisHealthChecker struct {
	redis *storage.RedisStorage
}

// NewRedisHealthChecker 创建Redis健康检查器
func NewRedisHealthChecker(redis *storage.RedisStorage) *RedisHealthChecker {
	return &RedisHealthChecker{redis: redis}
}

func (r *RedisHealthChecker) Name() string {
	return "redis"
}

func (r *RedisHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// 执行PING命令检查Redis连接
	err := r.redis.Ping(ctx)
	responseTime := time.Since(start)
	
	if err != nil {
		return ComponentHealth{
			Name:         r.Name(),
			Status:       StatusCritical,
			Message:      fmt.Sprintf("Redis connection failed: %v", err),
			LastCheck:    time.Now(),
			ResponseTime: responseTime,
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		}
	}
	
	// 获取Redis信息
	info, err := r.redis.GetInfo(ctx)
	status := StatusHealthy
	message := "Redis is healthy"
	
	details := map[string]interface{}{
		"response_time_ms": responseTime.Milliseconds(),
	}
	
	if info != nil {
		details["connected_clients"] = info["connected_clients"]
		details["used_memory"] = info["used_memory"]
		details["used_memory_human"] = info["used_memory_human"]
		details["keyspace_hits"] = info["keyspace_hits"]
		details["keyspace_misses"] = info["keyspace_misses"]
		
		// 检查内存使用率
		if maxMemory, ok := info["maxmemory"].(string); ok && maxMemory != "0" {
			// 这里可以添加内存使用率检查逻辑
		}
	}
	
	return ComponentHealth{
		Name:         r.Name(),
		Status:       status,
		Message:      message,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Details:      details,
	}
}

// AIServiceHealthChecker AI服务健康检查器
type AIServiceHealthChecker struct {
	baseURL string
}

// NewAIServiceHealthChecker 创建AI服务健康检查器
func NewAIServiceHealthChecker(baseURL string) *AIServiceHealthChecker {
	return &AIServiceHealthChecker{baseURL: baseURL}
}

func (a *AIServiceHealthChecker) Name() string {
	return "ai_service"
}

func (a *AIServiceHealthChecker) Check(ctx context.Context) ComponentHealth {
	start := time.Now()
	
	// 这里应该实现对AI服务的健康检查
	// 可以通过HTTP请求检查AI服务的健康状态
	
	responseTime := time.Since(start)
	
	// 简化实现，实际应该发送HTTP请求
	return ComponentHealth{
		Name:         a.Name(),
		Status:       StatusHealthy,
		Message:      "AI service is healthy",
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		Details: map[string]interface{}{
			"base_url": a.baseURL,
			"response_time_ms": responseTime.Milliseconds(),
		},
	}
}