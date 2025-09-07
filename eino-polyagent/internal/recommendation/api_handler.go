package recommendation

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// APIHandler handles HTTP requests for recommendation system
type APIHandler struct {
	orchestrator *RecommendationOrchestrator
	logger       *logrus.Logger
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(orchestrator *RecommendationOrchestrator, logger *logrus.Logger) *APIHandler {
	return &APIHandler{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// RegisterRoutes registers all recommendation API routes
func (h *APIHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1/recommendation")
	{
		// Data operations
		api.POST("/data/collect", h.handleDataCollection)
		api.POST("/data/features", h.handleFeatureEngineering)
		api.POST("/data/validate", h.handleDataValidation)
		
		// Model operations
		api.POST("/models/train", h.handleModelTraining)
		api.POST("/models/evaluate", h.handleModelEvaluation)
		api.POST("/models/optimize", h.handleHyperParameterTuning)
		api.POST("/models/deploy", h.handleModelDeployment)
		api.GET("/models", h.handleListModels)
		api.GET("/models/:id", h.handleGetModel)
		
		// Prediction
		api.POST("/predict", h.handlePredict)
		api.POST("/recommend", h.handleRecommend)
		
		// System monitoring
		api.GET("/agents", h.handleListAgents)
		api.GET("/agents/:id/stats", h.handleGetAgentStats)
		api.GET("/system/metrics", h.handleGetSystemMetrics)
		api.GET("/health", h.handleHealthCheck)
	}
}

// Data Collection API
func (h *APIHandler) handleDataCollection(c *gin.Context) {
	var req DataCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskDataCollection,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"collector": req.Collector,
			"timerange": req.TimeRange,
			"filters":   req.Filters,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		h.logger.WithError(err).Error("Data collection failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":     result.TaskID,
		"success":     result.Success,
		"data":        result.Data,
		"metrics":     result.Metrics,
		"created_at":  result.CreatedAt,
	})
}

// Feature Engineering API
func (h *APIHandler) handleFeatureEngineering(c *gin.Context) {
	var req FeatureEngineeringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskFeatureEngineering,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"feature_types": req.FeatureTypes,
			"data_sources":  req.DataSources,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":    result.TaskID,
		"success":    result.Success,
		"features":   result.Data,
		"created_at": result.CreatedAt,
	})
}

// Model Training API
func (h *APIHandler) handleModelTraining(c *gin.Context) {
	var req ModelTrainingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskModelTraining,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"algorithm":        req.Algorithm,
			"hyperparameters":  req.HyperParameters,
			"training_config":  req.TrainingConfig,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":       result.TaskID,
		"success":       result.Success,
		"model_id":      result.Data["model_id"],
		"algorithm":     result.Data["algorithm"],
		"training_time": result.Data["training_time"],
		"created_at":    result.CreatedAt,
	})
}

// Model Evaluation API
func (h *APIHandler) handleModelEvaluation(c *gin.Context) {
	var req ModelEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskModelEvaluation,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"model_id": req.ModelID,
			"metrics":  req.Metrics,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":            result.TaskID,
		"success":            result.Success,
		"model_id":           req.ModelID,
		"evaluation_metrics": result.Data["evaluation_metrics"],
		"created_at":         result.CreatedAt,
	})
}

// Hyperparameter Tuning API
func (h *APIHandler) handleHyperParameterTuning(c *gin.Context) {
	var req HyperParameterTuningRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskHyperParamTuning,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"algorithm": req.Algorithm,
			"budget":    req.Budget,
			"method":    req.Method,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":         result.TaskID,
		"success":         result.Success,
		"algorithm":       req.Algorithm,
		"best_parameters": result.Data["best_parameters"],
		"best_score":      result.Data["best_score"],
		"created_at":      result.CreatedAt,
	})
}

// Model Deployment API
func (h *APIHandler) handleModelDeployment(c *gin.Context) {
	var req ModelDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskModelDeployment,
		Priority: PriorityCritical, // Deployment is always critical
		Parameters: map[string]interface{}{
			"model_id": req.ModelID,
			"strategy": req.Strategy,
			"config":   req.Config,
		},
		MaxRetries: 1,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":           result.TaskID,
		"success":           result.Success,
		"model_id":          req.ModelID,
		"deployment_status": result.Data["deployment_status"],
		"created_at":        result.CreatedAt,
	})
}

// Recommendation API
func (h *APIHandler) handleRecommend(c *gin.Context) {
	var req RecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find a deployed model agent
	agents := h.orchestrator.GetAgents()
	var modelAgent RecommendationAgent
	for _, agent := range agents {
		if agent.GetType() == AgentTypeModel && agent.GetStatus() == StatusIdle {
			modelAgent = agent
			break
		}
	}

	if modelAgent == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "no model agent available"})
		return
	}

	// Mock recommendation generation
	recommendations := h.generateMockRecommendations(req)

	c.JSON(http.StatusOK, gin.H{
		"user_id":         req.UserID,
		"recommendations": recommendations,
		"model_id":        "active_model",
		"timestamp":       time.Now(),
		"request_id":      c.GetHeader("X-Request-ID"),
	})
}

// Prediction API
func (h *APIHandler) handlePredict(c *gin.Context) {
	var req PredictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Mock prediction
	prediction := map[string]interface{}{
		"user_id":    req.UserID,
		"item_id":    req.ItemID,
		"score":      0.85,
		"confidence": 0.92,
		"model_id":   "active_model",
		"timestamp":  time.Now(),
	}

	c.JSON(http.StatusOK, prediction)
}

// List Models API
func (h *APIHandler) handleListModels(c *gin.Context) {
	// Get all model agents and their models
	models := make([]map[string]interface{}, 0)
	
	agents := h.orchestrator.GetAgents()
	for _, agent := range agents {
		if modelAgent, ok := agent.(*ModelAgent); ok {
			agentModels := modelAgent.GetModels()
			for _, model := range agentModels {
				models = append(models, map[string]interface{}{
					"id":           model.ID,
					"name":         model.Name,
					"algorithm":    model.Algorithm,
					"version":      model.Version,
					"status":       model.Status,
					"created_at":   model.CreatedAt,
					"training_time": model.TrainingTime,
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
		"total":  len(models),
	})
}

// Get Model API
func (h *APIHandler) handleGetModel(c *gin.Context) {
	modelID := c.Param("id")
	if modelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model ID is required"})
		return
	}

	// Find model in agents
	agents := h.orchestrator.GetAgents()
	for _, agent := range agents {
		if modelAgent, ok := agent.(*ModelAgent); ok {
			models := modelAgent.GetModels()
			if model, exists := models[modelID]; exists {
				c.JSON(http.StatusOK, model)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "model not found"})
}

// List Agents API
func (h *APIHandler) handleListAgents(c *gin.Context) {
	agents := h.orchestrator.GetAgents()
	agentList := make([]map[string]interface{}, 0, len(agents))

	for _, agent := range agents {
		agentInfo := map[string]interface{}{
			"id":           agent.GetID(),
			"type":         agent.GetType(),
			"status":       agent.GetStatus(),
			"capabilities": agent.GetCapabilities(),
			"metrics":      agent.GetMetrics(),
		}
		agentList = append(agentList, agentInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agentList,
		"total":  len(agentList),
	})
}

// Get Agent Stats API
func (h *APIHandler) handleGetAgentStats(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent ID is required"})
		return
	}

	agents := h.orchestrator.GetAgents()
	agent, exists := agents[agentID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	stats := agent.GetPerformanceStats()
	c.JSON(http.StatusOK, stats)
}

// System Metrics API
func (h *APIHandler) handleGetSystemMetrics(c *gin.Context) {
	metrics := h.orchestrator.GetSystemMetrics()
	c.JSON(http.StatusOK, metrics)
}

// Health Check API
func (h *APIHandler) handleHealthCheck(c *gin.Context) {
	agents := h.orchestrator.GetAgents()
	healthStatus := map[string]interface{}{
		"status":        "healthy",
		"total_agents":  len(agents),
		"active_agents": 0,
		"timestamp":     time.Now(),
		"agents":        make(map[string]string),
	}

	for agentID, agent := range agents {
		status := agent.GetStatus()
		healthStatus["agents"].(map[string]string)[agentID] = string(status)
		if status != StatusError && status != StatusMaintenance {
			healthStatus["active_agents"] = healthStatus["active_agents"].(int) + 1
		}
	}

	c.JSON(http.StatusOK, healthStatus)
}

// Helper methods

func (h *APIHandler) getPriority(priority string) TaskPriority {
	switch priority {
	case "low":
		return PriorityLow
	case "medium":
		return PriorityMedium
	case "high":
		return PriorityHigh
	case "critical":
		return PriorityCritical
	default:
		return PriorityMedium
	}
}

func (h *APIHandler) generateMockRecommendations(req RecommendationRequest) []map[string]interface{} {
	recommendations := make([]map[string]interface{}, req.TopK)
	
	for i := 0; i < req.TopK; i++ {
		recommendations[i] = map[string]interface{}{
			"item_id":    "item_" + strconv.Itoa(i+1),
			"score":      0.9 - float64(i)*0.05,
			"rank":       i + 1,
			"reason":     "Based on user preferences and behavior",
			"confidence": 0.8 + float64(i)*0.02,
		}
	}
	
	return recommendations
}

// Request/Response structures

type DataCollectionRequest struct {
	Collector  string                 `json:"collector" binding:"required"`
	TimeRange  string                 `json:"time_range,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Priority   string                 `json:"priority,omitempty"`
	MaxRetries int                    `json:"max_retries,omitempty"`
}

type FeatureEngineeringRequest struct {
	FeatureTypes []string               `json:"feature_types" binding:"required"`
	DataSources  []string               `json:"data_sources,omitempty"`
	Priority     string                 `json:"priority,omitempty"`
	MaxRetries   int                    `json:"max_retries,omitempty"`
}

type ModelTrainingRequest struct {
	Algorithm       string                 `json:"algorithm" binding:"required"`
	HyperParameters map[string]interface{} `json:"hyperparameters,omitempty"`
	TrainingConfig  map[string]interface{} `json:"training_config,omitempty"`
	Priority        string                 `json:"priority,omitempty"`
	MaxRetries      int                    `json:"max_retries,omitempty"`
}

type ModelEvaluationRequest struct {
	ModelID    string   `json:"model_id" binding:"required"`
	Metrics    []string `json:"metrics,omitempty"`
	Priority   string   `json:"priority,omitempty"`
	MaxRetries int      `json:"max_retries,omitempty"`
}

type HyperParameterTuningRequest struct {
	Algorithm  string `json:"algorithm" binding:"required"`
	Budget     int    `json:"budget,omitempty"`
	Method     string `json:"method,omitempty"`
	Priority   string `json:"priority,omitempty"`
	MaxRetries int    `json:"max_retries,omitempty"`
}

type ModelDeploymentRequest struct {
	ModelID  string                 `json:"model_id" binding:"required"`
	Strategy string                 `json:"strategy,omitempty"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

type RecommendationRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	TopK        int                    `json:"top_k,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
}

type PredictionRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	ItemID      string                 `json:"item_id" binding:"required"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// DataValidation request handler
func (h *APIHandler) handleDataValidation(c *gin.Context) {
	var req struct {
		Rules      []string `json:"rules" binding:"required"`
		Priority   string   `json:"priority,omitempty"`
		MaxRetries int      `json:"max_retries,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &RecommendationTask{
		Type:     TaskDataValidation,
		Priority: h.getPriority(req.Priority),
		Parameters: map[string]interface{}{
			"validation_rules": req.Rules,
		},
		MaxRetries: req.MaxRetries,
	}

	result, err := h.orchestrator.ProcessTask(c.Request.Context(), task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"task_id":    result.TaskID,
		"success":    result.Success,
		"validation": result.Data,
		"created_at": result.CreatedAt,
	})
}