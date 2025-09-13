package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polyagent/eino-polyagent/internal/recommendation"
	"github.com/sirupsen/logrus"
)

func main() {
	log.Println("🚀 Starting Real Recommendation Server...")

	// Setup logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Initialize storage with real MovieLens data
	storage, err := recommendation.NewSQLiteStorage("/tmp/server_movielens.db", logger)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Load data if not exists
	stats := storage.GetStorageStats()
	if stats.RatingCount == 0 {
		log.Println("📥 Loading MovieLens data for first time...")
		err = storage.LoadMovieLensData("100k")
		if err != nil {
			log.Fatalf("Failed to load data: %v", err)
		}
		log.Println("✅ Data loaded successfully")
	} else {
		log.Printf("✅ Using existing data: %d users, %d movies, %d ratings",
			stats.UserCount, stats.MovieCount, stats.RatingCount)
	}

	// Initialize and train algorithm
	cf := recommendation.NewCollaborativeFiltering()
	ctx := context.Background()
	err = cf.Train(ctx, storage)
	if err != nil {
		log.Fatalf("Training failed: %v", err)
	}
	log.Printf("🧠 Algorithm trained: %s", cf.Name())

	// 创建推荐系统编排器配置
	config := &recommendation.OrchestratorConfig{
		MaxConcurrentTasks:  100,
		TaskTimeout:         5 * time.Minute,
		HealthCheckInterval: 30 * time.Second,
		MetricsInterval:     1 * time.Minute,
		RetryPolicy: &recommendation.RetryPolicy{
			MaxRetries:        3,
			InitialDelay:      1 * time.Second,
			BackoffMultiplier: 2.0,
			MaxDelay:          30 * time.Second,
		},
	}

	orchestrator := recommendation.NewRecommendationOrchestrator(config, logger)

	// 创建API处理器
	apiHandler := recommendation.NewAPIHandler(orchestrator, logger)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 注册推荐系统API路由
	apiHandler.RegisterRoutes(r)

	// 添加简单的健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "recommendation-server",
			"algorithm": cf.Name(),
			"timestamp": time.Now(),
		})
	})

	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}

	log.Printf("🌐 Server running on http://localhost%s", port)
	log.Println("📋 Available endpoints:")
	log.Println("  GET  /health                              - Health check")
	log.Println("  GET  /api/v1/recommendation/health        - Health check")
	log.Println("  GET  /api/v1/recommendation/system/metrics - System metrics")
	log.Println("  GET  /api/v1/recommendation/agents        - Agent list")
	log.Println("  GET  /api/v1/recommendation/models        - Model list")
	log.Println("  POST /api/v1/recommendation/recommend     - Generate recommendations")
	log.Println("  POST /api/v1/recommendation/predict       - Generate predictions")

	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
