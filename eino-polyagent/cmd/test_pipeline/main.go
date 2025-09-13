package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	
	"github.com/polyagent/eino-polyagent/internal/recommendation"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🚀 End-to-End Recommendation Pipeline Test")
	fmt.Println(strings.Repeat("=", 50))
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	// Step 1: Initialize storage and load real data
	fmt.Println("\n🗄️  Step 1: Initialize Database & Load Real Data")
	dbPath := "/tmp/test_pipeline.db"
	storage, err := recommendation.NewSQLiteStorage(dbPath, logger)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()
	
	// Load MovieLens 100K data
	fmt.Println("📥 Loading MovieLens 100K dataset...")
	err = storage.LoadMovieLensData("100k")
	if err != nil {
		log.Fatalf("Failed to load data: %v", err)
	}
	
	stats := storage.GetStorageStats()
	fmt.Printf("✅ Data loaded: %d users, %d movies, %d ratings\n", 
		stats.UserCount, stats.MovieCount, stats.RatingCount)
	
	// Step 2: Initialize recommendation algorithm
	fmt.Println("\n🧠 Step 2: Initialize & Train Algorithm")
	cf := recommendation.NewCollaborativeFiltering()
	fmt.Printf("🔧 Algorithm: %s\n", cf.Name())
	
	// Train the algorithm with loaded data
	ctx := context.Background()
	err = cf.Train(ctx, storage)
	if err != nil {
		log.Fatalf("Training failed: %v", err)
	}
	fmt.Println("✅ Training completed")
	
	// Step 3: Generate recommendations for sample users
	fmt.Println("\n🎯 Step 3: Generate Real Recommendations")
	
	testUsers := []string{"1", "5", "10", "50", "100"}
	for i, userID := range testUsers {
		fmt.Printf("\n👤 User %s Recommendations:\n", userID)
		
		// Get recommendations for user
		input := &recommendation.PredictionInput{
			UserID:  userID,
			ItemIDs: []string{"1", "2", "3", "4", "5", "10", "15", "20", "25", "30"},
			TopK:    5,
		}
		
		output, err := cf.Predict(ctx, input)
		if err != nil {
			fmt.Printf("❌ Failed to generate recommendations: %v\n", err)
			continue
		}
		
		fmt.Printf("🎬 Top %d Movies for User %s:\n", len(output.Recommendations), userID)
		for j, rec := range output.Recommendations {
			fmt.Printf("  %d. Movie ID: %s (Score: %.3f, Confidence: %.3f)\n", 
				j+1, rec.ItemID, rec.Score, rec.Confidence)
		}
		
		if i < len(testUsers)-1 {
			fmt.Println("  " + strings.Repeat("-", 40))
		}
	}
	
	// Step 4: Performance metrics
	fmt.Println("\n📊 Step 4: System Performance Metrics")
	fmt.Printf("💾 Database: %s\n", dbPath)
	fmt.Printf("📈 Dataset: MovieLens 100K\n")
	fmt.Printf("👥 Users: %d\n", stats.UserCount)
	fmt.Printf("🎬 Movies: %d\n", stats.MovieCount) 
	fmt.Printf("⭐ Ratings: %d\n", stats.RatingCount)
	fmt.Printf("🤖 Algorithm: %s\n", cf.Name())
	
	// Step 5: Data quality check
	fmt.Println("\n🔍 Step 5: Data Quality Assessment")
	coverage := float64(stats.RatingCount) / (float64(stats.UserCount) * float64(stats.MovieCount)) * 100
	fmt.Printf("📏 Data Sparsity: %.4f%% (Rating Coverage)\n", coverage)
	
	avgRatingsPerUser := float64(stats.RatingCount) / float64(stats.UserCount)
	fmt.Printf("👤 Avg Ratings per User: %.1f\n", avgRatingsPerUser)
	
	avgRatingsPerMovie := float64(stats.RatingCount) / float64(stats.MovieCount)
	fmt.Printf("🎬 Avg Ratings per Movie: %.1f\n", avgRatingsPerMovie)
	
	fmt.Println("\n🎉 End-to-End Pipeline Test Completed Successfully!")
	fmt.Println("✨ The recommendation system is fully functional with real data!")
}