package main

import (
	"fmt"
	"log"
	"os"
	
	"github.com/polyagent/eino-polyagent/internal/recommendation"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("🎯 Testing MovieLens Recommendation System")
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	
	// Create database path
	dbPath := "/tmp/test_movielens.db"
	fmt.Printf("Using database: %s\n", dbPath)
	
	// Create SQLite storage
	storage, err := recommendation.NewSQLiteStorage(dbPath, logger)
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()
	
	fmt.Println("✅ SQLite storage created successfully")
	
	// Test initial stats
	stats := storage.GetStorageStats()
	fmt.Printf("Initial stats - Users: %d, Movies: %d, Ratings: %d\n", 
		stats.UserCount, stats.MovieCount, stats.RatingCount)
	
	// Try to load MovieLens data
	fmt.Println("\n📊 Attempting to load MovieLens 100K dataset...")
	
	// Check if data exists (adjust path based on where we run from)
	dataPath := "../data/movielens/ml-100k"
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		dataPath = "../../data/movielens/ml-100k"
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			fmt.Printf("❌ MovieLens data not found at %s\n", dataPath)
			fmt.Println("Please ensure MovieLens datasets are downloaded to data/movielens/")
			return
		}
	}
	
	// Load data
	err = storage.LoadMovieLensData("100k")
	if err != nil {
		fmt.Printf("❌ Failed to load MovieLens data: %v\n", err)
		return
	}
	
	fmt.Println("✅ MovieLens 100K dataset loaded successfully")
	
	// Show updated stats
	stats = storage.GetStorageStats()
	fmt.Printf("📈 Updated stats - Users: %d, Movies: %d, Ratings: %d\n", 
		stats.UserCount, stats.MovieCount, stats.RatingCount)
	
	// Test collaborative filtering
	fmt.Println("\n🧠 Testing Collaborative Filtering Algorithm...")
	
	cf := recommendation.NewCollaborativeFiltering()
	if cf == nil {
		fmt.Println("❌ Failed to create collaborative filtering algorithm")
		return
	}
	
	fmt.Printf("✅ Algorithm: %s\n", cf.Name())
	
	// Show hyperparameters
	hyperParams := cf.GetHyperParameters()
	fmt.Printf("🔧 Available hyperparameters (%d):\n", len(hyperParams))
	for name, param := range hyperParams {
		fmt.Printf("  • %s: %s (default: %v)\n", name, param.Type, param.Default)
	}
	
	// Test MovieLens collector
	fmt.Println("\n📥 Testing MovieLens Data Collector...")
	
	collector := recommendation.NewMovieLensCollector(storage, logger)
	if collector == nil {
		fmt.Println("❌ Failed to create MovieLens collector")
		return
	}
	
	fmt.Printf("✅ Collector: %s\n", collector.Name())
	
	// Show schema
	schema := collector.GetSchema()
	if schema != nil {
		fmt.Printf("📋 Data schema has %d fields:\n", len(schema.Fields))
		for name, field := range schema.Fields {
			fmt.Printf("  • %s: %s (required: %v)\n", name, field.Type, field.Required)
		}
	}
	
	fmt.Println("\n🎉 Basic recommendation system components tested successfully!")
	fmt.Println("\n📝 Summary:")
	fmt.Printf("  • Database: %s\n", dbPath)
	fmt.Printf("  • Users loaded: %d\n", stats.UserCount)
	fmt.Printf("  • Movies loaded: %d\n", stats.MovieCount) 
	fmt.Printf("  • Ratings loaded: %d\n", stats.RatingCount)
	fmt.Printf("  • Algorithm: %s\n", cf.Name())
	fmt.Printf("  • Data collector: %s\n", collector.Name())
	
	fmt.Println("\n✨ Ready for real recommendation training and prediction!")
}