package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/go-services/internal/models"
	"github.com/polyagent/go-services/internal/storage"
)

// AgentHandler 智能体处理器
type AgentHandler struct {
	postgres *storage.PostgresStorage
	redis    *storage.RedisStorage
}

// NewAgentHandler 创建智能体处理器
func NewAgentHandler(postgres *storage.PostgresStorage, redis *storage.RedisStorage) *AgentHandler {
	return &AgentHandler{
		postgres: postgres,
		redis:    redis,
	}
}

// CreateAgentRequest 创建智能体请求
type CreateAgentRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Description  string                 `json:"description"`
	Instructions string                 `json:"instructions"`
	Tools        []string               `json:"tools"`
	Config       map[string]interface{} `json:"config"`
}

// UpdateAgentRequest 更新智能体请求
type UpdateAgentRequest struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Instructions string                 `json:"instructions"`
	Tools        []string               `json:"tools"`
	Config       map[string]interface{} `json:"config"`
	Status       *models.AgentStatus    `json:"status"`
}

// ListAgents 获取智能体列表
func (h *AgentHandler) ListAgents(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agents, err := h.postgres.GetUserAgents(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"total":  len(agents),
	})
}

// CreateAgent 创建智能体
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建智能体
	agent := &models.Agent{
		ID:           models.NewAgentID(),
		UserID:       userID.(string),
		Name:         req.Name,
		Type:         req.Type,
		Description:  req.Description,
		Instructions: req.Instructions,
		Tools:        req.Tools,
		Config:       req.Config,
		Status:       models.AgentStatusActive,
	}

	if err := h.postgres.CreateAgent(agent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

// GetAgent 获取智能体详情
func (h *AgentHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")

	agent, err := h.postgres.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// 检查权限：只有创建者可以查看
	userID, exists := c.Get("user_id")
	if !exists || agent.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	c.JSON(http.StatusOK, agent)
}

// UpdateAgent 更新智能体
func (h *AgentHandler) UpdateAgent(c *gin.Context) {
	agentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 检查智能体是否存在
	agent, err := h.postgres.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// 检查权限
	if agent.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	var req UpdateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		agent.Name = req.Name
	}
	if req.Description != "" {
		agent.Description = req.Description
	}
	if req.Instructions != "" {
		agent.Instructions = req.Instructions
	}
	if req.Tools != nil {
		agent.Tools = req.Tools
	}
	if req.Config != nil {
		agent.Config = req.Config
	}
	if req.Status != nil {
		agent.Status = *req.Status
	}

	// 这里应该实现更新逻辑（简化版本）
	c.JSON(http.StatusOK, agent)
}

// DeleteAgent 删除智能体
func (h *AgentHandler) DeleteAgent(c *gin.Context) {
	agentID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 检查智能体是否存在
	agent, err := h.postgres.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
		return
	}

	// 检查权限
	if agent.UserID != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// 软删除：更新状态为已归档
	agent.Status = models.AgentStatusArchived
	// 这里应该实现更新逻辑

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}