package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP 请求指标
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// AI 请求指标
	aiRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "Total number of AI requests",
		},
		[]string{"provider", "model", "status"},
	)

	aiRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "AI request duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
		},
		[]string{"provider", "model"},
	)

	aiTokensUsed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_tokens_used_total",
			Help: "Total number of AI tokens used",
		},
		[]string{"provider", "model", "type"},
	)

	// 数据库指标
	dbConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"database"},
	)

	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"database", "operation", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"database", "operation"},
	)

	// WebSocket 指标
	websocketConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)

	websocketMessages = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"type", "status"},
	)

	// 任务调度器指标
	tasksQueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "tasks_queue_size",
			Help: "Number of tasks in queue",
		},
		[]string{"queue"},
	)

	tasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_processed_total",
			Help: "Total number of processed tasks",
		},
		[]string{"queue", "status"},
	)

	taskProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_processing_duration_seconds",
			Help:    "Task processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 300},
		},
		[]string{"queue", "type"},
	)

	// RAG 系统指标
	ragQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rag_queries_total",
			Help: "Total number of RAG queries",
		},
		[]string{"retriever", "status"},
	)

	ragRetrievalDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rag_retrieval_duration_seconds",
			Help:    "RAG retrieval duration in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"retriever"},
	)

	ragDocumentsIndexed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rag_documents_indexed_total",
			Help: "Total number of documents indexed",
		},
	)

	// 系统资源指标
	memoryUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
	)

	cpuUsage = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "cpu_usage_percent",
			Help: "CPU usage percentage",
		},
	)

	goroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Number of active goroutines",
		},
	)
)

// InitMetrics 初始化指标收集
func InitMetrics() {
	// 注册自定义指标收集器
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
	prometheus.MustRegister(prometheus.NewGoCollector())
}

// PrometheusMiddleware Gin中间件用于收集HTTP指标
func PrometheusMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Observe(duration.Seconds())
	})
}

// RecordAIRequest 记录AI请求指标
func RecordAIRequest(provider, model, status string, duration time.Duration, inputTokens, outputTokens int) {
	aiRequestsTotal.WithLabelValues(provider, model, status).Inc()
	aiRequestDuration.WithLabelValues(provider, model).Observe(duration.Seconds())
	
	if inputTokens > 0 {
		aiTokensUsed.WithLabelValues(provider, model, "input").Add(float64(inputTokens))
	}
	if outputTokens > 0 {
		aiTokensUsed.WithLabelValues(provider, model, "output").Add(float64(outputTokens))
	}
}

// RecordDBQuery 记录数据库查询指标
func RecordDBQuery(database, operation, status string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(database, operation, status).Inc()
	dbQueryDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// UpdateDBConnections 更新数据库连接数指标
func UpdateDBConnections(database string, count int) {
	dbConnectionsActive.WithLabelValues(database).Set(float64(count))
}

// IncWebSocketConnections 增加WebSocket连接数
func IncWebSocketConnections() {
	websocketConnections.Inc()
}

// DecWebSocketConnections 减少WebSocket连接数
func DecWebSocketConnections() {
	websocketConnections.Dec()
}

// RecordWebSocketMessage 记录WebSocket消息指标
func RecordWebSocketMessage(msgType, status string) {
	websocketMessages.WithLabelValues(msgType, status).Inc()
}

// UpdateTaskQueueSize 更新任务队列大小
func UpdateTaskQueueSize(queue string, size int) {
	tasksQueueSize.WithLabelValues(queue).Set(float64(size))
}

// RecordTaskProcessed 记录任务处理指标
func RecordTaskProcessed(queue, status string, duration time.Duration, taskType string) {
	tasksProcessed.WithLabelValues(queue, status).Inc()
	taskProcessingDuration.WithLabelValues(queue, taskType).Observe(duration.Seconds())
}

// RecordRAGQuery 记录RAG查询指标
func RecordRAGQuery(retriever, status string, duration time.Duration) {
	ragQueriesTotal.WithLabelValues(retriever, status).Inc()
	ragRetrievalDuration.WithLabelValues(retriever).Observe(duration.Seconds())
}

// IncRAGDocumentsIndexed 增加已索引文档数
func IncRAGDocumentsIndexed() {
	ragDocumentsIndexed.Inc()
}

// UpdateSystemMetrics 更新系统资源指标
func UpdateSystemMetrics(memUsage uint64, cpuPercent float64, goroutineCount int) {
	memoryUsage.Set(float64(memUsage))
	cpuUsage.Set(cpuPercent)
	goroutines.Set(float64(goroutineCount))
}