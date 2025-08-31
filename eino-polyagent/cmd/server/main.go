package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
	"github.com/sirupsen/logrus"
)

type Server struct {
	orchestrator *orchestration.AgentOrchestrator
	logger       *logrus.Logger
}

type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	AgentID   string `json:"agent_id"`
}

type ChatResponse struct {
	Content   string                 `json:"content"`
	AgentID   string                 `json:"agent_id"`
	SessionID string                 `json:"session_id"`
	Latency   string                 `json:"latency"`
	Metadata  map[string]interface{} `json:"metadata"`
}

type CreateAgentRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type"`
	SystemPrompt string  `json:"system_prompt"`
	Model        string  `json:"model"`
	Temperature  float32 `json:"temperature"`
}

type CreateAgentResponse struct {
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Setup configuration
	cfg := &config.Config{}

	// Setup model router
	modelConfig := &ai.ModelConfig{
		DefaultModel: "gpt-4",
	}
	modelRouter := ai.NewModelRouter(modelConfig, logger)

	// Setup orchestrator
	orchestrator := orchestration.NewAgentOrchestrator(cfg, modelRouter, logger)

	// Create default agent
	defaultAgent := &orchestration.AgentConfig{
		ID:           "default",
		Name:         "Default Assistant",
		Type:         orchestration.AgentTypeConversational,
		SystemPrompt: "You are a helpful AI assistant.",
		Model:        "gpt-4",
		Temperature:  0.7,
		MaxTokens:    2000,
	}

	_, err := orchestrator.CreateAgent(defaultAgent)
	if err != nil {
		log.Fatalf("Failed to create default agent: %v", err)
	}

	server := &Server{
		orchestrator: orchestrator,
		logger:       logger,
	}

	// Setup Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api/v1")
	{
		api.POST("/chat", server.handleChat)
		api.POST("/agents", server.handleCreateAgent)
		api.GET("/agents", server.handleGetAgents)
		api.GET("/health", server.handleHealth)
		api.POST("/workflows/:type/execute", server.handleExecuteWorkflow)
	}

	// Health check endpoint
	r.GET("/health", server.handleHealth)

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}
	
	logger.WithField("port", port).Info("Starting PolyAgent server")
	log.Fatal(http.ListenAndServe(port, r))
}

func (s *Server) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use default agent if not specified
	if req.AgentID == "" {
		req.AgentID = "default"
	}

	if req.SessionID == "" {
		req.SessionID = "default-session"
	}

	if req.UserID == "" {
		req.UserID = "anonymous"
	}

	ctx := context.Background()
	result, err := s.orchestrator.ProcessMessage(ctx, req.AgentID, req.SessionID, req.Message, req.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	response := ChatResponse{
		Content:   result.Content,
		AgentID:   result.AgentID,
		SessionID: result.SessionID,
		Latency:   result.Latency.String(),
		Metadata:  result.Metadata,
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleCreateAgent(c *gin.Context) {
	var req CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Type == "" {
		req.Type = string(orchestration.AgentTypeConversational)
	}
	if req.Model == "" {
		req.Model = "gpt-4"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.SystemPrompt == "" {
		req.SystemPrompt = "You are a helpful AI assistant."
	}

	agentConfig := &orchestration.AgentConfig{
		Name:         req.Name,
		Type:         orchestration.AgentType(req.Type),
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    2000,
	}

	agentID, err := s.orchestrator.CreateAgent(agentConfig)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	response := CreateAgentResponse{
		AgentID: agentID,
		Message: "Agent created successfully",
	}

	c.JSON(http.StatusCreated, response)
}

func (s *Server) handleGetAgents(c *gin.Context) {
	agents := s.orchestrator.GetAgents()
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "polyagent",
		"timestamp": "now",
	})
}

func (s *Server) handleExecuteWorkflow(c *gin.Context) {
	workflowType := c.Param("type")
	
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agentID := "default"
	if id, exists := requestBody["agent_id"]; exists {
		if idStr, ok := id.(string); ok {
			agentID = idStr
		}
	}

	var workflow *orchestration.SimpleWorkflow
	
	switch workflowType {
	case "coding":
		workflow = orchestration.NewCodingWorkflow(agentID)
	case "research":
		workflow = orchestration.NewResearchWorkflow(agentID)
	case "problem-solving":
		workflow = orchestration.NewProblemSolvingWorkflow(agentID)
	case "content-creation":
		workflow = orchestration.NewContentCreationWorkflow(agentID)
	case "data-analysis":
		workflow = orchestration.NewDataAnalysisWorkflow(agentID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown workflow type"})
		return
	}

	// Return workflow definition (in a real implementation, this would execute the workflow)
	response := map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"status":        workflow.Status,
		"steps":         len(workflow.Steps),
		"created_at":    workflow.CreatedAt,
		"steps_detail":  workflow.Steps,
	}

	c.JSON(http.StatusOK, response)
}