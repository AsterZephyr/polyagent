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

type RecommendationServer struct {
	storage *recommendation.SQLiteStorage
	cf      *recommendation.CollaborativeFiltering
	logger  *logrus.Logger
}

type RecommendRequest struct {
	UserID string   `json:"user_id"`
	TopK   int      `json:"top_k"`
	Items  []string `json:"items,omitempty"`
}

type RecommendResponse struct {
	UserID          string                             `json:"user_id"`
	Recommendations []recommendation.RecommendationItem `json:"recommendations"`
	Algorithm       string                             `json:"algorithm"`
	GeneratedAt     time.Time                          `json:"generated_at"`
	Stats           *recommendation.StorageStats        `json:"stats"`
}

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
	
	server := &RecommendationServer{
		storage: storage,
		cf:      cf,
		logger:  logger,
	}
	
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
	
	// Routes
	r.GET("/health", server.handleHealth)
	r.GET("/api/v1/stats", server.handleStats)
	r.POST("/api/v1/recommend", server.handleRecommend)
	
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = ":" + envPort
	}
	
	log.Printf("🌐 Server running on http://localhost%s", port)
	log.Println("📋 Available endpoints:")
	log.Println("  GET  /health              - Health check")
	log.Println("  GET  /api/v1/stats        - System statistics")
	log.Println("  POST /api/v1/recommend    - Generate recommendations")
	
	if err := r.Run(port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func (s *RecommendationServer) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "recommendation-server",
		"algorithm": s.cf.Name(),
		"timestamp": time.Now(),
	})
}

func (s *RecommendationServer) handleStats(c *gin.Context) {
	stats := s.storage.GetStorageStats()
	c.JSON(http.StatusOK, gin.H{
		"stats":     stats,
		"algorithm": s.cf.Name(),
		"timestamp": time.Now(),
	})
}

func (s *RecommendationServer) handleRecommend(c *gin.Context) {
	var req RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Validate request
	if req.UserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	
	if req.TopK <= 0 {
		req.TopK = 10
	}
	
	// Default items if not provided
	if len(req.Items) == 0 {
		req.Items = []string{"1", "2", "3", "4", "5", "10", "15", "20", "25", "30", 
						   "35", "40", "45", "50", "55", "60", "65", "70", "75", "80"}
	}
	
	// Generate recommendations
	input := &recommendation.PredictionInput{
		UserID:  req.UserID,
		ItemIDs: req.Items,
		TopK:    req.TopK,
	}
	
	ctx := context.Background()
	output, err := s.cf.Predict(ctx, input)
	if err != nil {
		s.logger.Errorf("Prediction failed for user %s: %v", req.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate recommendations"})
		return
	}
	
	response := RecommendResponse{
		UserID:          req.UserID,
		Recommendations: output.Recommendations,
		Algorithm:       s.cf.Name(),
		GeneratedAt:     time.Now(),
		Stats:           s.storage.GetStorageStats(),
	}
	
	s.logger.Infof("Generated %d recommendations for user %s", len(output.Recommendations), req.UserID)
	c.JSON(http.StatusOK, response)
}
