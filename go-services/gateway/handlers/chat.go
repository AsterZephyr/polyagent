package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/polyagent/go-services/internal/ai"
	"github.com/polyagent/go-services/internal/models"
	"github.com/polyagent/go-services/internal/scheduler"
	"github.com/polyagent/go-services/internal/storage"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	scheduler *scheduler.TaskScheduler
	postgres  *storage.PostgresStorage
	redis     *storage.RedisStorage
	aiClient  *ai.PythonAIClient
	upgrader  websocket.Upgrader
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(scheduler *scheduler.TaskScheduler, postgres *storage.PostgresStorage, redis *storage.RedisStorage, aiClient *ai.PythonAIClient) *ChatHandler {
	return &ChatHandler{
		scheduler: scheduler,
		postgres:  postgres,
		redis:     redis,
		aiClient:  aiClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该更严格
			},
		},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Message   string   `json:"message" binding:"required"`
	SessionID string   `json:"session_id"`
	AgentType string   `json:"agent_type"`
	Tools     []string `json:"tools"`
	Stream    bool     `json:"stream"`
	Priority  int      `json:"priority"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	TaskID    string    `json:"task_id"`
	Message   string    `json:"message"`
	SessionID string    `json:"session_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成会话ID（如果没有提供）
	if req.SessionID == "" {
		req.SessionID = generateSessionID(userID.(string))
	}

	// 设置默认值
	if req.AgentType == "" {
		req.AgentType = "general"
	}

	// 创建任务
	task := &models.AgentTask{
		TaskID:    models.NewTaskID(),
		UserID:    userID.(string),
		SessionID: req.SessionID,
		AgentType: req.AgentType,
		Input:     req.Message,
		Tools:     req.Tools,
		Priority:  req.Priority,
		Context: map[string]interface{}{
			"stream": req.Stream,
		},
	}

	// 获取对话记忆
	memory, err := h.postgres.GetConversationMemory(req.SessionID)
	if err == nil {
		task.Memory = memory
	}

	// 提交任务
	if err := h.scheduler.SubmitTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit task"})
		return
	}

	// 如果是流式响应，返回任务ID供客户端轮询或WebSocket连接
	if req.Stream {
		c.JSON(http.StatusAccepted, ChatResponse{
			TaskID:    task.TaskID,
			SessionID: req.SessionID,
			Status:    "processing",
			Timestamp: time.Now(),
		})
		return
	}

	// 同步等待结果（简化实现）
	response := h.waitForTaskCompletion(task.TaskID, 30*time.Second)
	
	c.JSON(http.StatusOK, response)
}

// StreamChat 处理流式聊天请求
func (h *ChatHandler) StreamChat(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置SSE头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// 创建并提交任务
	task := &models.AgentTask{
		TaskID:    models.NewTaskID(),
		UserID:    userID.(string),
		SessionID: req.SessionID,
		AgentType: req.AgentType,
		Input:     req.Message,
		Tools:     req.Tools,
		Priority:  req.Priority,
		Context: map[string]interface{}{
			"stream": true,
		},
	}

	if err := h.scheduler.SubmitTask(task); err != nil {
		c.SSEvent("error", gin.H{"error": "Failed to submit task"})
		return
	}

	// 流式返回结果
	h.streamTaskProgress(c, task.TaskID)
}

// WebSocketHandler WebSocket处理器
func (h *ChatHandler) WebSocketHandler(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}
	defer conn.Close()

	// WebSocket消息处理循环
	for {
		var req ChatRequest
		if err := conn.ReadJSON(&req); err != nil {
			break
		}

		// 处理消息并发送响应
		h.handleWebSocketMessage(conn, req)
	}
}

// GetSessions 获取用户会话列表
func (h *ChatHandler) GetSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	_ = userID // 避免unused变量警告
	
	// 这里应该从数据库获取会话列表
	// 简化实现
	sessions := []map[string]interface{}{
		{
			"session_id": "session_1",
			"title":      "示例对话1",
			"updated_at": time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// GetSessionMessages 获取会话消息
func (h *ChatHandler) GetSessionMessages(c *gin.Context) {
	sessionID := c.Param("session_id")
	
	memory, err := h.postgres.GetConversationMemory(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": memory.Messages})
}

// GetTasks 获取用户任务列表
func (h *ChatHandler) GetTasks(c *gin.Context) {
	userID, _ := c.Get("user_id")
	_ = userID // 避免unused变量警告
	
	// 从数据库获取用户任务
	// 这里应该实现分页和过滤
	c.JSON(http.StatusOK, gin.H{
		"tasks": []interface{}{}, // 简化实现
		"total": 0,
	})
}

// GetTask 获取任务详情
func (h *ChatHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	
	task, err := h.postgres.GetTask(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CancelTask 取消任务
func (h *ChatHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	
	// 更新任务状态为已取消
	if err := h.postgres.UpdateTaskStatus(taskID, models.TaskStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled successfully"})
}

// ListTools 获取可用工具列表
func (h *ChatHandler) ListTools(c *gin.Context) {
	tools, err := h.aiClient.GetAvailableTools()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tools"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

// ExecuteTool 执行工具
func (h *ChatHandler) ExecuteTool(c *gin.Context) {
	var toolCall models.ToolCall
	if err := c.ShouldBindJSON(&toolCall); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.aiClient.ExecuteTool(&toolCall)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute tool"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetTaskStats 获取任务统计
func (h *ChatHandler) GetTaskStats(c *gin.Context) {
	date := c.Query("date")
	
	stats, err := h.scheduler.GetTaskStats(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// 辅助函数

// generateSessionID 生成会话ID
func generateSessionID(userID string) string {
	return userID + "_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// waitForTaskCompletion 等待任务完成
func (h *ChatHandler) waitForTaskCompletion(taskID string, timeout time.Duration) ChatResponse {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		task, err := h.postgres.GetTask(taskID)
		if err != nil {
			return ChatResponse{
				TaskID: taskID,
				Status: "error",
				Message: "Task not found",
				Timestamp: time.Now(),
			}
		}

		if task.Status == models.TaskStatusCompleted {
			return ChatResponse{
				TaskID: taskID,
				Status: "completed",
				Message: "Task completed", // 这里应该从AI响应获取
				SessionID: task.SessionID,
				Timestamp: time.Now(),
			}
		}

		if task.Status == models.TaskStatusFailed {
			return ChatResponse{
				TaskID: taskID,
				Status: "failed",
				Message: "Task failed",
				Timestamp: time.Now(),
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return ChatResponse{
		TaskID: taskID,
		Status: "timeout",
		Message: "Task timeout",
		Timestamp: time.Now(),
	}
}

// streamTaskProgress 流式返回任务进度
func (h *ChatHandler) streamTaskProgress(c *gin.Context, taskID string) {
	// 这里应该实现真正的流式响应
	// 可以通过Redis的发布订阅机制监听任务进度
	c.SSEvent("start", gin.H{"task_id": taskID})
	
	// 简化实现：等待任务完成并发送结果
	result := h.waitForTaskCompletion(taskID, 30*time.Second)
	c.SSEvent("message", result)
	c.SSEvent("end", gin.H{"task_id": taskID})
}

// handleWebSocketMessage 处理WebSocket消息
func (h *ChatHandler) handleWebSocketMessage(conn *websocket.Conn, req ChatRequest) {
	// 创建任务
	task := &models.AgentTask{
		TaskID:    models.NewTaskID(),
		SessionID: req.SessionID,
		AgentType: req.AgentType,
		Input:     req.Message,
		Tools:     req.Tools,
		Context: map[string]interface{}{
			"stream": true,
		},
	}

	// 提交任务
	if err := h.scheduler.SubmitTask(task); err != nil {
		conn.WriteJSON(gin.H{"error": "Failed to submit task"})
		return
	}

	// 发送任务开始通知
	conn.WriteJSON(gin.H{
		"type":    "task_start",
		"task_id": task.TaskID,
	})

	// 等待并发送结果
	result := h.waitForTaskCompletion(task.TaskID, 30*time.Second)
	conn.WriteJSON(result)
}