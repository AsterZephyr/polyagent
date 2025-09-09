package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/eino-polyagent/internal/recommendation"
	"github.com/sirupsen/logrus"
)

// RecommendationServer - 专门的推荐业务服务器
type RecommendationServer struct {
	recOrchestrator *recommendation.RecommendationOrchestrator
	recAPIHandler   *recommendation.APIHandler
	logger          *logrus.Logger
}

func main() {
	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Setup recommendation orchestrator
	recOrchestrator := recommendation.NewRecommendationOrchestrator(nil, logger)

	// Create and register recommendation agents
	dataAgent, err := recommendation.NewDataAgent("rec-data-agent-001", nil, logger)
	if err != nil {
		log.Fatalf("Failed to create data agent: %v", err)
	}

	modelAgent, err := recommendation.NewModelAgent("rec-model-agent-001", nil, logger)
	if err != nil {
		log.Fatalf("Failed to create model agent: %v", err)
	}

	if err := recOrchestrator.RegisterAgent(dataAgent); err != nil {
		log.Fatalf("Failed to register data agent: %v", err)
	}

	if err := recOrchestrator.RegisterAgent(modelAgent); err != nil {
		log.Fatalf("Failed to register model agent: %v", err)
	}

	// Create recommendation API handler
	recAPIHandler := recommendation.NewAPIHandler(recOrchestrator, logger)

	server := &RecommendationServer{
		recOrchestrator: recOrchestrator,
		recAPIHandler:   recAPIHandler,
		logger:          logger,
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

	// Register recommendation API routes
	server.recAPIHandler.RegisterRoutes(r)

	// Health check endpoint
	r.GET("/health", server.handleHealth)

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	logger.WithField("port", port).Info("Starting Recommendation Agent Server")
	log.Fatal(http.ListenAndServe(port, r))
}

func (s *RecommendationServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "recommendation-agent-server",
		"timestamp": "now",
	})
}
