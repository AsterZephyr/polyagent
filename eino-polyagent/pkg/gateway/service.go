package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/polyagent/eino-polyagent/internal/ai"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/polyagent/eino-polyagent/internal/orchestration"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type ChatRequest struct {
	Message   string                 `json:"message" binding:"required"`
	SessionID string                 `json:"session_id,omitempty"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Stream    bool                   `json:"stream,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type ChatResponse struct {
	Response    string                 `json:"response"`
	SessionID   string                 `json:"session_id"`
	AgentID     string                 `json:"agent_id"`
	TokensUsed  int                    `json:"tokens_used"`
	Cost        float64                `json:"cost"`
	Latency     time.Duration          `json:"latency"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type AgentCreateRequest struct {
	Name          string                 `json:"name" binding:"required"`
	Type          string                 `json:"type" binding:"required"`
	SystemPrompt  string                 `json:"system_prompt" binding:"required"`
	Model         string                 `json:"model,omitempty"`
	Temperature   float32                `json:"temperature,omitempty"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	ToolsEnabled  bool                   `json:"tools_enabled,omitempty"`
	MemoryEnabled bool                   `json:"memory_enabled,omitempty"`
	SafetyFilters []string               `json:"safety_filters,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type UserContext struct {
	UserID string            `json:"user_id"`
	Email  string            `json:"email"`
	Roles  []string          `json:"roles"`
	Claims map[string]string `json:"claims"`
}

type GatewayService struct {
	config       *config.Config
	orchestrator *orchestration.AgentOrchestrator
	modelRouter  *ai.ModelRouter
	logger       *logrus.Logger
	limiter      *rate.Limiter
	router       *gin.Engine
}

func NewGatewayService(cfg *config.Config, orchestrator *orchestration.AgentOrchestrator, modelRouter *ai.ModelRouter, logger *logrus.Logger) *GatewayService {
	service := &GatewayService{
		config:       cfg,
		orchestrator: orchestrator,
		modelRouter:  modelRouter,
		logger:       logger,
		limiter:      rate.NewLimiter(rate.Limit(cfg.Security.RateLimit.RequestsPerMin), cfg.Security.RateLimit.BurstSize),
	}

	service.setupRouter()
	return service
}

func (s *GatewayService) setupRouter() {
	if s.config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	s.router = gin.New()
	
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.corsMiddleware())
	s.router.Use(s.rateLimitMiddleware())

	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/models", s.getModels)
		
		auth := api.Group("")
		auth.Use(s.authMiddleware())
		{
			auth.POST("/chat", s.handleChat)
			auth.POST("/chat/stream", s.handleStreamChat)
			auth.POST("/agents", s.createAgent)
			auth.GET("/agents", s.listAgents)
			auth.GET("/agents/:id", s.getAgent)
			auth.DELETE("/agents/:id", s.deleteAgent)
			auth.GET("/sessions/:id/history", s.getSessionHistory)
		}
	}
}

func (s *GatewayService) Start() error {
	addr := s.config.Server.GetAddr()
	
	server := &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    s.config.Server.ReadTimeout,
		WriteTimeout:   s.config.Server.WriteTimeout,
		IdleTimeout:    s.config.Server.IdleTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	s.logger.WithField("addr", addr).Info("Starting gateway service")
	return server.ListenAndServe()
}

func (s *GatewayService) healthCheck(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	}

	if modelHealth, err := s.modelRouter.GetHealth(""); err == nil {
		health["models"] = modelHealth
	}

	c.JSON(http.StatusOK, health)
}

func (s *GatewayService) getModels(c *gin.Context) {
	health, err := s.modelRouter.GetHealth("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"models": health})
}

func (s *GatewayService) handleChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx := s.getUserContext(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	if req.SessionID == "" {
		req.SessionID = s.generateSessionID()
	}

	if req.AgentID == "" {
		req.AgentID = "default"
	}

	result, err := s.orchestrator.ProcessMessage(c.Request.Context(), req.AgentID, req.SessionID, req.Message, userCtx.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process message"})
		return
	}

	response := &ChatResponse{
		Response:   result.Response.Content,
		SessionID:  req.SessionID,
		AgentID:    result.AgentID,
		TokensUsed: result.TokensUsed,
		Cost:       result.Cost,
		Latency:    result.Latency,
		Metadata:   result.Metadata,
	}

	c.JSON(http.StatusOK, response)
}

func (s *GatewayService) handleStreamChat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx := s.getUserContext(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	if req.SessionID == "" {
		req.SessionID = s.generateSessionID()
	}

	if req.AgentID == "" {
		req.AgentID = "default"
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	streamChan, err := s.orchestrator.StreamResponse(c.Request.Context(), req.AgentID, req.SessionID, req.Message, userCtx.UserID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start stream")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start stream"})
		return
	}

	c.Stream(func(w gin.ResponseWriter) bool {
		select {
		case chunk, ok := <-streamChan:
			if !ok {
				return false
			}
			c.SSEvent("data", chunk)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}

func (s *GatewayService) createAgent(c *gin.Context) {
	var req AgentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userCtx := s.getUserContext(c)
	if userCtx == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	agentConfig := &orchestration.AgentConfig{
		Name:          req.Name,
		Type:          orchestration.AgentType(req.Type),
		SystemPrompt:  req.SystemPrompt,
		Model:         req.Model,
		Temperature:   req.Temperature,
		MaxTokens:     req.MaxTokens,
		ToolsEnabled:  req.ToolsEnabled,
		MemoryEnabled: req.MemoryEnabled,
		SafetyFilters: req.SafetyFilters,
		Metadata:      req.Metadata,
	}

	if agentConfig.Model == "" {
		agentConfig.Model = "openai"
	}
	if agentConfig.Temperature == 0 {
		agentConfig.Temperature = 0.7
	}
	if agentConfig.MaxTokens == 0 {
		agentConfig.MaxTokens = 2000
	}

	agentID, err := s.orchestrator.CreateAgent(agentConfig)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create agent")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"agent_id": agentID,
		"config":   agentConfig,
	})
}

func (s *GatewayService) listAgents(c *gin.Context) {
	agents := s.orchestrator.GetAgents()
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

func (s *GatewayService) getAgent(c *gin.Context) {
	agentID := c.Param("id")
	agents := s.orchestrator.GetAgents()
	
	if agent, exists := agents[agentID]; exists {
		c.JSON(http.StatusOK, gin.H{"agent": agent})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
	}
}

func (s *GatewayService) deleteAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	if err := s.orchestrator.DeleteAgent(agentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "agent deleted successfully"})
}

func (s *GatewayService) getSessionHistory(c *gin.Context) {
	sessionID := c.Param("id")
	agentID := c.Query("agent_id")
	
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "agent_id is required"})
		return
	}

	agents := s.orchestrator.GetAgents()
	if _, exists := agents[agentID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id": sessionID,
		"messages":   []interface{}{},
	})
}

func (s *GatewayService) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		allowedOrigins := s.config.Security.CORS.AllowOrigins
		if len(allowedOrigins) == 0 {
			allowedOrigins = []string{"*"}
		}

		allowed := false
		for _, ao := range allowedOrigins {
			if ao == "*" || ao == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", strconv.FormatBool(s.config.Security.CORS.AllowCredentials))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func (s *GatewayService) rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.config.Security.RateLimit.Enabled {
			c.Next()
			return
		}

		if !s.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *GatewayService) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]
		userCtx, err := s.validateJWT(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user", userCtx)
		c.Next()
	}
}

func (s *GatewayService) validateJWT(tokenString string) (*UserContext, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.Security.JWT.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)
	
	var roles []string
	if rolesClaim, exists := claims["roles"]; exists {
		if rolesSlice, ok := rolesClaim.([]interface{}); ok {
			for _, role := range rolesSlice {
				if roleStr, ok := role.(string); ok {
					roles = append(roles, roleStr)
				}
			}
		}
	}

	return &UserContext{
		UserID: userID,
		Email:  email,
		Roles:  roles,
		Claims: make(map[string]string),
	}, nil
}

func (s *GatewayService) getUserContext(c *gin.Context) *UserContext {
	if userCtx, exists := c.Get("user"); exists {
		return userCtx.(*UserContext)
	}
	return nil
}

func (s *GatewayService) generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}