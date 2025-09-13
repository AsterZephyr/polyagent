package llm

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// VectorSearchEngine provides semantic search capabilities using vector embeddings
type VectorSearchEngine struct {
	vectorStore      VectorStore
	embeddingService EmbeddingService
	indexManager     *IndexManager
	searchConfig     *SearchConfig
	logger           *logrus.Logger
	metrics          *VectorSearchMetrics
	mu               sync.RWMutex
}

// VectorStore defines the interface for vector storage backends
type VectorStore interface {
	// Store stores vectors with metadata
	Store(ctx context.Context, vectors []VectorDocument) error

	// Search performs similarity search
	Search(ctx context.Context, query []float64, k int, filters map[string]interface{}) ([]SearchResult, error)

	// Delete removes vectors by IDs
	Delete(ctx context.Context, ids []string) error

	// Update updates existing vectors
	Update(ctx context.Context, vectors []VectorDocument) error

	// GetStats returns storage statistics
	GetStats(ctx context.Context) (*VectorStoreStats, error)

	// CreateIndex creates a new search index
	CreateIndex(ctx context.Context, config *IndexConfig) error

	// Close closes the vector store
	Close() error
}

// EmbeddingService generates vector embeddings from text
type EmbeddingService interface {
	// GenerateEmbedding generates embedding for a single text
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

	// GenerateBatchEmbeddings generates embeddings for multiple texts
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error)

	// GetDimensions returns embedding dimensions
	GetDimensions() int

	// GetModel returns the embedding model name
	GetModel() string
}

// VectorDocument represents a document with its vector embedding
type VectorDocument struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Vector    []float64              `json:"vector"`
	Metadata  map[string]interface{} `json:"metadata"`
	IndexedAt time.Time              `json:"indexed_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Document   VectorDocument `json:"document"`
	Score      float64        `json:"score"`
	Rank       int            `json:"rank"`
	Highlights []string       `json:"highlights,omitempty"`
}

// SearchConfig configures vector search behavior
type SearchConfig struct {
	DefaultK           int     `json:"default_k"`
	MaxK               int     `json:"max_k"`
	DefaultSimilarity  string  `json:"default_similarity"` // "cosine", "euclidean", "dot_product"
	MinSimilarityScore float64 `json:"min_similarity_score"`
	EnableReranking    bool    `json:"enable_reranking"`
	RerankingModel     string  `json:"reranking_model"`
	CacheEnabled       bool    `json:"cache_enabled"`
	CacheTTL           int     `json:"cache_ttl_seconds"`
	SearchTimeout      int     `json:"search_timeout_ms"`
	EmbeddingBatchSize int     `json:"embedding_batch_size"`
}

// IndexConfig configures vector index creation
type IndexConfig struct {
	Name       string                 `json:"name"`
	Dimensions int                    `json:"dimensions"`
	IndexType  string                 `json:"index_type"`  // "flat", "ivf", "hnsw"
	MetricType string                 `json:"metric_type"` // "cosine", "l2", "ip"
	Parameters map[string]interface{} `json:"parameters"`
	Shards     int                    `json:"shards"`
	Replicas   int                    `json:"replicas"`
}

// VectorStoreStats provides statistics about vector storage
type VectorStoreStats struct {
	TotalDocuments int64     `json:"total_documents"`
	TotalVectors   int64     `json:"total_vectors"`
	IndexSize      int64     `json:"index_size_bytes"`
	Dimensions     int       `json:"dimensions"`
	LastUpdated    time.Time `json:"last_updated"`
	MemoryUsage    int64     `json:"memory_usage_bytes"`
	DiskUsage      int64     `json:"disk_usage_bytes"`
}

// VectorSearchMetrics tracks search performance metrics
type VectorSearchMetrics struct {
	TotalSearches    int64         `json:"total_searches"`
	AverageLatency   time.Duration `json:"average_latency"`
	CacheHitRate     float64       `json:"cache_hit_rate"`
	IndexingRate     float64       `json:"indexing_rate_docs_per_sec"`
	EmbeddingLatency time.Duration `json:"embedding_latency"`
	LastUpdated      time.Time     `json:"last_updated"`
	TopQueries       []string      `json:"top_queries"`
	mu               sync.RWMutex
}

// IndexManager manages vector indices
type IndexManager struct {
	indices map[string]*IndexConfig
	logger  *logrus.Logger
	mu      sync.RWMutex
}

// MovieVectorSearcher specializes vector search for movie recommendations
type MovieVectorSearcher struct {
	engine        *VectorSearchEngine
	movieEmbedder *MovieEmbeddingService
	logger        *logrus.Logger
}

// NewVectorSearchEngine creates a new vector search engine
func NewVectorSearchEngine(vectorStore VectorStore, embeddingService EmbeddingService, config *SearchConfig, logger *logrus.Logger) (*VectorSearchEngine, error) {
	if config == nil {
		config = createDefaultSearchConfig()
	}

	indexManager := &IndexManager{
		indices: make(map[string]*IndexConfig),
		logger:  logger,
	}

	engine := &VectorSearchEngine{
		vectorStore:      vectorStore,
		embeddingService: embeddingService,
		indexManager:     indexManager,
		searchConfig:     config,
		logger:           logger,
		metrics: &VectorSearchMetrics{
			LastUpdated: time.Now(),
		},
	}

	logger.Info("Vector search engine initialized successfully")
	return engine, nil
}

// IndexDocuments indexes documents for semantic search
func (vse *VectorSearchEngine) IndexDocuments(ctx context.Context, documents []VectorDocument) error {
	startTime := time.Now()

	vse.logger.WithField("document_count", len(documents)).Info("Starting document indexing")

	// Generate embeddings for documents that don't have them
	var textsToEmbed []string
	var indicesToUpdate []int

	for i, doc := range documents {
		if len(doc.Vector) == 0 {
			textsToEmbed = append(textsToEmbed, doc.Content)
			indicesToUpdate = append(indicesToUpdate, i)
		}
	}

	if len(textsToEmbed) > 0 {
		embeddings, err := vse.embeddingService.GenerateBatchEmbeddings(ctx, textsToEmbed)
		if err != nil {
			return fmt.Errorf("failed to generate embeddings: %w", err)
		}

		// Update documents with embeddings
		for i, embedding := range embeddings {
			docIndex := indicesToUpdate[i]
			documents[docIndex].Vector = embedding
			documents[docIndex].IndexedAt = time.Now()
		}
	}

	// Store in vector store
	if err := vse.vectorStore.Store(ctx, documents); err != nil {
		return fmt.Errorf("failed to store vectors: %w", err)
	}

	// Update metrics
	indexingTime := time.Since(startTime)
	vse.updateIndexingMetrics(len(documents), indexingTime)

	vse.logger.WithFields(logrus.Fields{
		"documents_indexed": len(documents),
		"indexing_time":     indexingTime,
	}).Info("Document indexing completed")

	return nil
}

// SemanticSearch performs semantic search using vector similarity
func (vse *VectorSearchEngine) SemanticSearch(ctx context.Context, query string, k int, filters map[string]interface{}) ([]SearchResult, error) {
	startTime := time.Now()

	vse.logger.WithFields(logrus.Fields{
		"query": query,
		"k":     k,
	}).Info("Performing semantic search")

	// Validate and adjust k
	if k <= 0 {
		k = vse.searchConfig.DefaultK
	}
	if k > vse.searchConfig.MaxK {
		k = vse.searchConfig.MaxK
	}

	// Generate query embedding
	queryVector, err := vse.embeddingService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform vector search
	results, err := vse.vectorStore.Search(ctx, queryVector, k, filters)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Filter by minimum similarity score
	filteredResults := make([]SearchResult, 0)
	for i, result := range results {
		if result.Score >= vse.searchConfig.MinSimilarityScore {
			result.Rank = i + 1
			filteredResults = append(filteredResults, result)
		}
	}

	// Apply reranking if enabled
	if vse.searchConfig.EnableReranking && len(filteredResults) > 1 {
		filteredResults = vse.rerankResults(ctx, query, filteredResults)
	}

	searchTime := time.Since(startTime)
	vse.updateSearchMetrics(searchTime, query)

	vse.logger.WithFields(logrus.Fields{
		"results_found": len(filteredResults),
		"search_time":   searchTime,
	}).Info("Semantic search completed")

	return filteredResults, nil
}

// HybridSearch combines vector search with traditional keyword search
func (vse *VectorSearchEngine) HybridSearch(ctx context.Context, query string, k int, filters map[string]interface{}, weights map[string]float64) ([]SearchResult, error) {
	// Default weights: 70% semantic, 30% keyword
	if weights == nil {
		weights = map[string]float64{
			"semantic": 0.7,
			"keyword":  0.3,
		}
	}

	// Perform semantic search
	semanticResults, err := vse.SemanticSearch(ctx, query, k*2, filters) // Get more results for fusion
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}

	// Perform keyword search (simplified implementation)
	keywordResults := vse.performKeywordSearch(ctx, query, k*2, filters)

	// Fuse results using reciprocal rank fusion
	fusedResults := vse.fuseSearchResults(semanticResults, keywordResults, weights)

	// Return top k results
	if len(fusedResults) > k {
		fusedResults = fusedResults[:k]
	}

	return fusedResults, nil
}

// SimilaritySearch finds documents similar to a given document
func (vse *VectorSearchEngine) SimilaritySearch(ctx context.Context, documentID string, k int) ([]SearchResult, error) {
	// This would retrieve the document vector and search for similar ones
	// Implementation depends on vector store capabilities
	vse.logger.WithFields(logrus.Fields{
		"document_id": documentID,
		"k":           k,
	}).Info("Performing similarity search")

	// Mock implementation
	return []SearchResult{}, fmt.Errorf("similarity search not implemented yet")
}

// UpdateDocuments updates existing documents in the vector store
func (vse *VectorSearchEngine) UpdateDocuments(ctx context.Context, documents []VectorDocument) error {
	vse.logger.WithField("document_count", len(documents)).Info("Updating documents")

	// Generate new embeddings if content changed
	for i := range documents {
		if documents[i].Content != "" {
			embedding, err := vse.embeddingService.GenerateEmbedding(ctx, documents[i].Content)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for document %s: %w", documents[i].ID, err)
			}
			documents[i].Vector = embedding
			documents[i].UpdatedAt = time.Now()
		}
	}

	return vse.vectorStore.Update(ctx, documents)
}

// DeleteDocuments removes documents from the vector store
func (vse *VectorSearchEngine) DeleteDocuments(ctx context.Context, documentIDs []string) error {
	vse.logger.WithField("document_count", len(documentIDs)).Info("Deleting documents")
	return vse.vectorStore.Delete(ctx, documentIDs)
}

// GetSearchMetrics returns current search metrics
func (vse *VectorSearchEngine) GetSearchMetrics() *VectorSearchMetrics {
	vse.metrics.mu.RLock()
	defer vse.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	metricsCopy := *vse.metrics
	return &metricsCopy
}

// GetStoreStats returns vector store statistics
func (vse *VectorSearchEngine) GetStoreStats(ctx context.Context) (*VectorStoreStats, error) {
	return vse.vectorStore.GetStats(ctx)
}

// performKeywordSearch performs simple keyword-based search
func (vse *VectorSearchEngine) performKeywordSearch(ctx context.Context, query string, k int, filters map[string]interface{}) []SearchResult {
	// Simplified keyword search implementation
	// In a real implementation, this would use a text search engine like Elasticsearch

	vse.logger.Debug("Performing keyword search")

	// Mock results for demonstration
	mockResults := []SearchResult{
		{
			Document: VectorDocument{
				ID:      "keyword_1",
				Content: "Action movie with great special effects",
				Metadata: map[string]interface{}{
					"title":  "Action Hero",
					"genre":  "Action",
					"rating": 4.2,
				},
			},
			Score: 0.8,
		},
		{
			Document: VectorDocument{
				ID:      "keyword_2",
				Content: "Sci-fi thriller with amazing visuals",
				Metadata: map[string]interface{}{
					"title":  "Future Wars",
					"genre":  "Sci-Fi",
					"rating": 4.5,
				},
			},
			Score: 0.7,
		},
	}

	return mockResults
}

// fuseSearchResults combines semantic and keyword search results
func (vse *VectorSearchEngine) fuseSearchResults(semanticResults, keywordResults []SearchResult, weights map[string]float64) []SearchResult {
	// Create a map to combine results by document ID
	resultMap := make(map[string]*SearchResult)

	// Process semantic results
	for i, result := range semanticResults {
		fusedScore := result.Score * weights["semantic"]
		// Apply reciprocal rank fusion
		fusedScore += weights["semantic"] / float64(i+1)

		resultMap[result.Document.ID] = &SearchResult{
			Document: result.Document,
			Score:    fusedScore,
		}
	}

	// Process keyword results
	for i, result := range keywordResults {
		fusedScore := result.Score * weights["keyword"]
		fusedScore += weights["keyword"] / float64(i+1)

		if existing, exists := resultMap[result.Document.ID]; exists {
			// Combine scores
			existing.Score += fusedScore
		} else {
			resultMap[result.Document.ID] = &SearchResult{
				Document: result.Document,
				Score:    fusedScore,
			}
		}
	}

	// Convert map to slice and sort by score
	var fusedResults []SearchResult
	for _, result := range resultMap {
		fusedResults = append(fusedResults, *result)
	}

	sort.Slice(fusedResults, func(i, j int) bool {
		return fusedResults[i].Score > fusedResults[j].Score
	})

	// Update ranks
	for i := range fusedResults {
		fusedResults[i].Rank = i + 1
	}

	return fusedResults
}

// rerankResults applies reranking to improve result quality
func (vse *VectorSearchEngine) rerankResults(ctx context.Context, query string, results []SearchResult) []SearchResult {
	// Simplified reranking implementation
	// In practice, this would use a cross-encoder model

	vse.logger.Debug("Applying result reranking")

	// Apply simple text matching bonus
	for i := range results {
		content := results[i].Document.Content
		if title, exists := results[i].Document.Metadata["title"]; exists {
			if titleStr, ok := title.(string); ok {
				content += " " + titleStr
			}
		}

		// Simple text overlap scoring
		overlapScore := vse.calculateTextOverlap(query, content)
		results[i].Score = results[i].Score*0.8 + overlapScore*0.2
	}

	// Re-sort by new scores
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Update ranks
	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

// calculateTextOverlap calculates simple text overlap score
func (vse *VectorSearchEngine) calculateTextOverlap(query, content string) float64 {
	// Very simplified implementation
	// In practice, use more sophisticated text matching algorithms

	queryWords := splitWords(query)
	contentWords := splitWords(content)

	queryMap := make(map[string]bool)
	for _, word := range queryWords {
		queryMap[word] = true
	}

	overlap := 0
	for _, word := range contentWords {
		if queryMap[word] {
			overlap++
		}
	}

	if len(queryWords) == 0 {
		return 0.0
	}

	return float64(overlap) / float64(len(queryWords))
}

// splitWords splits text into words (simplified)
func splitWords(text string) []string {
	// Very simplified word splitting
	// In practice, use proper tokenization
	words := make([]string, 0)
	current := ""

	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

// updateSearchMetrics updates search performance metrics
func (vse *VectorSearchEngine) updateSearchMetrics(latency time.Duration, query string) {
	vse.metrics.mu.Lock()
	defer vse.metrics.mu.Unlock()

	vse.metrics.TotalSearches++

	// Update average latency
	if vse.metrics.TotalSearches == 1 {
		vse.metrics.AverageLatency = latency
	} else {
		vse.metrics.AverageLatency = (vse.metrics.AverageLatency + latency) / 2
	}

	// Track top queries (simplified)
	if len(vse.metrics.TopQueries) < 10 {
		vse.metrics.TopQueries = append(vse.metrics.TopQueries, query)
	}

	vse.metrics.LastUpdated = time.Now()
}

// updateIndexingMetrics updates indexing performance metrics
func (vse *VectorSearchEngine) updateIndexingMetrics(docCount int, duration time.Duration) {
	vse.metrics.mu.Lock()
	defer vse.metrics.mu.Unlock()

	rate := float64(docCount) / duration.Seconds()
	if vse.metrics.IndexingRate == 0 {
		vse.metrics.IndexingRate = rate
	} else {
		vse.metrics.IndexingRate = (vse.metrics.IndexingRate + rate) / 2
	}

	vse.metrics.LastUpdated = time.Now()
}

// createDefaultSearchConfig creates default search configuration
func createDefaultSearchConfig() *SearchConfig {
	return &SearchConfig{
		DefaultK:           10,
		MaxK:               100,
		DefaultSimilarity:  "cosine",
		MinSimilarityScore: 0.3,
		EnableReranking:    true,
		RerankingModel:     "cross_encoder",
		CacheEnabled:       true,
		CacheTTL:           300,  // 5 minutes
		SearchTimeout:      5000, // 5 seconds
		EmbeddingBatchSize: 32,
	}
}

// CosineSimilarity calculates cosine similarity between two vectors
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EuclideanDistance calculates Euclidean distance between two vectors
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}

	var sum float64
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return math.Sqrt(sum)
}

// DotProduct calculates dot product between two vectors
func DotProduct(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var product float64
	for i := range a {
		product += a[i] * b[i]
	}

	return product
}
