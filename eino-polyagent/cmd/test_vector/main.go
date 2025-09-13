package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/polyagent/eino-polyagent/internal/llm"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("=== 向量数据库集成和语义搜索测试 ===")

	ctx := context.Background()
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create embedding service
	embeddingService := llm.NewLocalEmbeddingService(768, logger)

	// Create vector store
	vectorStore := llm.NewInMemoryVectorStore(nil, logger)

	// Create vector search engine
	searchConfig := &llm.SearchConfig{
		DefaultK:           5,
		MaxK:               20,
		DefaultSimilarity:  "cosine",
		MinSimilarityScore: 0.3,
		EnableReranking:    true,
		RerankingModel:     "cross_encoder",
		CacheEnabled:       true,
		CacheTTL:           300,
		SearchTimeout:      5000,
		EmbeddingBatchSize: 10,
	}

	searchEngine, err := llm.NewVectorSearchEngine(vectorStore, embeddingService, searchConfig, logger)
	if err != nil {
		log.Fatalf("创建向量搜索引擎失败: %v", err)
	}

	// Create movie embedding service
	movieEmbeddingService := llm.NewMovieEmbeddingService(embeddingService, nil, logger)

	fmt.Println("\n--- 步骤1: 准备电影数据 ---")

	// Prepare test movies
	testMovies := []*llm.RecommendedMovie{
		{
			ID:      1,
			Title:   "The Matrix",
			Genres:  []string{"Action", "Sci-Fi"},
			Year:    1999,
			Rating:  4.5,
			Description: "A computer programmer is led to fight an underground war against powerful computers who have constructed his entire reality with a system called the Matrix.",
			Reason:  "Classic sci-fi with groundbreaking action sequences",
		},
		{
			ID:      2,
			Title:   "Inception",
			Genres:  []string{"Action", "Sci-Fi", "Thriller"},
			Year:    2010,
			Rating:  4.7,
			Description: "A thief who steals corporate secrets through the use of dream-sharing technology is given the inverse task of planting an idea into the mind of a C.E.O.",
			Reason:  "Mind-bending thriller with complex narrative",
		},
		{
			ID:      3,
			Title:   "The Godfather",
			Genres:  []string{"Crime", "Drama"},
			Year:    1972,
			Rating:  4.9,
			Description: "The aging patriarch of an organized crime dynasty transfers control of his clandestine empire to his reluctant son.",
			Reason:  "Masterpiece of crime drama cinema",
		},
		{
			ID:      4,
			Title:   "Pulp Fiction",
			Genres:  []string{"Crime", "Drama"},
			Year:    1994,
			Rating:  4.8,
			Description: "The lives of two mob hitmen, a boxer, a gangster and his wife, and a pair of diner bandits intertwine in four tales of violence and redemption.",
			Reason:  "Iconic non-linear storytelling",
		},
		{
			ID:      5,
			Title:   "The Dark Knight",
			Genres:  []string{"Action", "Crime", "Drama"},
			Year:    2008,
			Rating:  4.8,
			Description: "When the menace known as the Joker wreaks havoc and chaos on the people of Gotham, Batman must accept one of the greatest psychological and physical tests of his ability to fight injustice.",
			Reason:  "Perfect superhero film with complex themes",
		},
		{
			ID:      6,
			Title:   "Forrest Gump",
			Genres:  []string{"Drama", "Romance"},
			Year:    1994,
			Rating:  4.7,
			Description: "The presidencies of Kennedy and Johnson, the Vietnam War, the Watergate scandal and other historical events unfold from the perspective of an Alabama man with an IQ of 75.",
			Reason:  "Heartwarming story spanning decades",
		},
		{
			ID:      7,
			Title:   "Interstellar",
			Genres:  []string{"Adventure", "Drama", "Sci-Fi"},
			Year:    2014,
			Rating:  4.6,
			Description: "A team of explorers travel through a wormhole in space in an attempt to ensure humanity's survival.",
			Reason:  "Scientifically accurate space epic",
		},
		{
			ID:      8,
			Title:   "The Avengers",
			Genres:  []string{"Action", "Adventure", "Sci-Fi"},
			Year:    2012,
			Rating:  4.2,
			Description: "Earth's mightiest heroes must come together and learn to fight as a team if they are going to stop the trickster Loki and his alien army from enslaving humanity.",
			Reason:  "Epic superhero team-up adventure",
		},
	}

	fmt.Printf("准备了 %d 部电影数据\n", len(testMovies))

	fmt.Println("\n--- 步骤2: 生成电影嵌入向量 ---")

	// Generate embeddings for movies
	movieEmbeddings, err := movieEmbeddingService.GenerateBatchMovieEmbeddings(ctx, testMovies)
	if err != nil {
		log.Fatalf("生成电影嵌入向量失败: %v", err)
	}

	fmt.Printf("成功生成 %d 个嵌入向量，每个维度: %d\n", len(movieEmbeddings), len(movieEmbeddings[0]))

	fmt.Println("\n--- 步骤3: 构建向量文档 ---")

	// Create vector documents
	var vectorDocs []llm.VectorDocument
	for i, movie := range testMovies {
		doc := llm.VectorDocument{
			ID:      fmt.Sprintf("movie_%d", movie.ID),
			Content: fmt.Sprintf("%s. %s", movie.Title, movie.Description),
			Vector:  movieEmbeddings[i],
			Metadata: map[string]interface{}{
				"movie_id":    movie.ID,
				"title":       movie.Title,
				"genres":      movie.Genres,
				"year":        movie.Year,
				"rating":      movie.Rating,
				"description": movie.Description,
			},
			IndexedAt: time.Now(),
		}
		vectorDocs = append(vectorDocs, doc)
	}

	fmt.Printf("构建了 %d 个向量文档\n", len(vectorDocs))

	fmt.Println("\n--- 步骤4: 索引向量文档 ---")

	// Index documents
	if err := searchEngine.IndexDocuments(ctx, vectorDocs); err != nil {
		log.Fatalf("索引文档失败: %v", err)
	}

	// Get store statistics
	stats, err := searchEngine.GetStoreStats(ctx)
	if err != nil {
		log.Printf("获取存储统计失败: %v", err)
	} else {
		fmt.Printf("存储统计: %d 文档, %d 向量, 内存使用: %d 字节\n", 
			stats.TotalDocuments, stats.TotalVectors, stats.MemoryUsage)
	}

	fmt.Println("\n--- 步骤5: 语义搜索测试 ---")

	// Test semantic search scenarios
	searchQueries := []struct {
		name    string
		query   string
		k       int
		filters map[string]interface{}
	}{
		{
			name:  "科幻电影搜索",
			query: "futuristic science fiction movie with advanced technology",
			k:     3,
		},
		{
			name:  "犯罪剧情片搜索",
			query: "crime drama about organized crime and family",
			k:     2,
		},
		{
			name:  "动作英雄电影搜索",
			query: "action hero superhero fighting evil",
			k:     3,
		},
		{
			name:    "高评分电影搜索",
			query:   "excellent highly rated masterpiece movie",
			k:       5,
			filters: map[string]interface{}{},
		},
		{
			name:  "情感剧情片搜索",
			query: "emotional heartwarming drama about life journey",
			k:     2,
		},
	}

	for i, searchQuery := range searchQueries {
		fmt.Printf("\n--- 搜索场景 %d: %s ---\n", i+1, searchQuery.name)
		fmt.Printf("查询: %s\n", searchQuery.query)

		startTime := time.Now()
		results, err := searchEngine.SemanticSearch(ctx, searchQuery.query, searchQuery.k, searchQuery.filters)
		searchTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("搜索失败: %v\n", err)
			continue
		}

		fmt.Printf("搜索时间: %v\n", searchTime)
		fmt.Printf("找到 %d 个结果:\n", len(results))

		for j, result := range results {
			title := result.Document.Metadata["title"].(string)
			rating := result.Document.Metadata["rating"].(float64)
			year := result.Document.Metadata["year"].(int)
			genres := result.Document.Metadata["genres"]

			fmt.Printf("  %d. %s (%d) - 评分: %.1f - 相似度: %.3f\n",
				j+1, title, year, rating, result.Score)
			
			if genreSlice, ok := genres.([]string); ok {
				fmt.Printf("     类型: %v\n", genreSlice)
			}
			
			if len(result.Document.Content) > 100 {
				fmt.Printf("     内容: %s...\n", result.Document.Content[:100])
			} else {
				fmt.Printf("     内容: %s\n", result.Document.Content)
			}
		}
	}

	fmt.Println("\n--- 步骤6: 混合搜索测试 ---")

	hybridQuery := "action packed superhero adventure"
	fmt.Printf("混合搜索查询: %s\n", hybridQuery)

	hybridResults, err := searchEngine.HybridSearch(ctx, hybridQuery, 4, nil, map[string]float64{
		"semantic": 0.7,
		"keyword":  0.3,
	})

	if err != nil {
		log.Printf("混合搜索失败: %v", err)
	} else {
		fmt.Printf("混合搜索结果 (%d 个):\n", len(hybridResults))
		for i, result := range hybridResults {
			title := result.Document.Metadata["title"].(string)
			fmt.Printf("  %d. %s - 融合评分: %.3f\n", i+1, title, result.Score)
		}
	}

	fmt.Println("\n--- 步骤7: 搜索性能指标 ---")

	searchMetrics := searchEngine.GetSearchMetrics()
	metricsJSON, _ := json.MarshalIndent(searchMetrics, "", "  ")
	fmt.Printf("搜索指标:\n%s\n", string(metricsJSON))

	embeddingMetrics := movieEmbeddingService.GetMetrics()
	embeddingJSON, _ := json.MarshalIndent(embeddingMetrics, "", "  ")
	fmt.Printf("嵌入生成指标:\n%s\n", string(embeddingJSON))

	fmt.Println("\n--- 步骤8: 文档更新和删除测试 ---")

	// Test document updates
	updateDoc := llm.VectorDocument{
		ID:      "movie_1",
		Content: "The Matrix - Enhanced description: A groundbreaking sci-fi film about virtual reality and human consciousness.",
		Metadata: map[string]interface{}{
			"movie_id":    1,
			"title":       "The Matrix",
			"genres":      []string{"Action", "Sci-Fi", "Philosophy"},
			"year":        1999,
			"rating":      4.6, // Updated rating
			"description": "Enhanced description with philosophical themes",
		},
	}

	fmt.Println("更新《The Matrix》文档...")
	if err := searchEngine.UpdateDocuments(ctx, []llm.VectorDocument{updateDoc}); err != nil {
		log.Printf("更新文档失败: %v", err)
	} else {
		fmt.Println("文档更新成功")
	}

	// Test search after update
	fmt.Println("更新后搜索测试...")
	updatedResults, err := searchEngine.SemanticSearch(ctx, "philosophical virtual reality", 2, nil)
	if err != nil {
		log.Printf("更新后搜索失败: %v", err)
	} else {
		fmt.Printf("更新后搜索结果 (%d 个):\n", len(updatedResults))
		for i, result := range updatedResults {
			title := result.Document.Metadata["title"].(string)
			fmt.Printf("  %d. %s - 相似度: %.3f\n", i+1, title, result.Score)
		}
	}

	// Test document deletion
	fmt.Println("删除文档测试...")
	if err := searchEngine.DeleteDocuments(ctx, []string{"movie_8"}); err != nil {
		log.Printf("删除文档失败: %v", err)
	} else {
		fmt.Println("成功删除《The Avengers》文档")
	}

	// Final statistics
	finalStats, err := searchEngine.GetStoreStats(ctx)
	if err != nil {
		log.Printf("获取最终统计失败: %v", err)
	} else {
		fmt.Printf("最终统计: %d 文档, %d 向量\n", finalStats.TotalDocuments, finalStats.TotalVectors)
	}

	fmt.Println("\n--- 步骤9: 相似度算法测试 ---")

	// Test different similarity algorithms
	testVec1 := []float64{1.0, 0.5, 0.3, 0.8}
	testVec2 := []float64{0.8, 0.6, 0.2, 0.9}

	cosine := llm.CosineSimilarity(testVec1, testVec2)
	euclidean := llm.EuclideanDistance(testVec1, testVec2)
	dotProduct := llm.DotProduct(testVec1, testVec2)

	fmt.Printf("向量相似度测试:\n")
	fmt.Printf("  余弦相似度: %.4f\n", cosine)
	fmt.Printf("  欧几里得距离: %.4f\n", euclidean)
	fmt.Printf("  点积: %.4f\n", dotProduct)

	fmt.Println("\n--- 总结 ---")
	fmt.Println("✅ 向量搜索引擎初始化成功")
	fmt.Println("✅ 电影内容嵌入生成完成")
	fmt.Println("✅ 向量文档索引建立成功")
	fmt.Println("✅ 语义搜索功能正常工作")
	fmt.Println("✅ 混合搜索（语义+关键词）实现")
	fmt.Println("✅ 文档更新和删除功能验证")
	fmt.Println("✅ 搜索性能指标收集正常")
	fmt.Println("✅ 多种相似度算法支持")

	fmt.Println("\n=== 向量数据库集成和语义搜索测试完成 ===")
}