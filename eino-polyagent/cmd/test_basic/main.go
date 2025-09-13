package main

import (
	"context"
	"fmt"
	"log"
	
	"github.com/polyagent/eino-polyagent/internal/recommendation"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🎯 Basic Recommendation System Test")
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	// Test 1: SQLite Storage
	fmt.Println("\n📊 Testing SQLite Storage...")
	dbPath := "/tmp/test_basic.db"
	storage, err := recommendation.NewSQLiteStorage(dbPath, logger)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()
	
	stats := storage.GetStorageStats()
	fmt.Printf("✅ Storage created - Users: %d, Movies: %d, Ratings: %d\n", 
		stats.UserCount, stats.MovieCount, stats.RatingCount)
	
	// Test 2: Collaborative Filtering
	fmt.Println("\n🧠 Testing Collaborative Filtering...")
	cf := recommendation.NewCollaborativeFiltering()
	if cf == nil {
		log.Fatal("Failed to create collaborative filtering")
	}
	
	fmt.Printf("✅ Algorithm: %s\n", cf.Name())
	hyperParams := cf.GetHyperParameters()
	fmt.Printf("✅ Hyperparameters: %d configured\n", len(hyperParams))
	
	// Test training
	ctx := context.Background()
	err = cf.Train(ctx, nil)
	if err != nil {
		log.Fatalf("Training failed: %v", err)
	}
	fmt.Println("✅ Training completed successfully")
	
	// Test prediction
	input := &recommendation.PredictionInput{
		UserID:  "user_1",
		ItemIDs: []string{"item_1", "item_2", "item_3"},
		TopK:    3,
	}
	
	output, err := cf.Predict(ctx, input)
	if err != nil {
		log.Fatalf("Prediction failed: %v", err)
	}
	
	fmt.Printf("✅ Prediction completed - %d predictions, %d recommendations\n", 
		len(output.Predictions), len(output.Recommendations))
	
	// Test 3: MovieLens Collector
	fmt.Println("\n📥 Testing MovieLens Collector...")
	collector := recommendation.NewMovieLensCollector(storage, logger)
	if collector == nil {
		log.Fatal("Failed to create collector")
	}
	
	fmt.Printf("✅ Collector: %s\n", collector.Name())
	schema := collector.GetSchema()
	if schema != nil {
		fmt.Printf("✅ Schema fields: %d\n", len(schema.Fields))
	}
	
	// Test collection (mock)
	err = collector.Collect(ctx, map[string]interface{}{})
	if err != nil {
		log.Fatalf("Collection failed: %v", err)
	}
	fmt.Println("✅ Data collection completed")
	
	fmt.Println("\n🎉 All basic tests passed!")
	fmt.Println("\n📝 Summary:")
	fmt.Printf("  • Database: %s\n", dbPath)
	fmt.Printf("  • Algorithm: %s\n", cf.Name())
	fmt.Printf("  • Collector: %s\n", collector.Name())
	fmt.Printf("  • Predictions generated: %d\n", len(output.Predictions))
	
	fmt.Println("\n✨ Basic recommendation system is working!")
}