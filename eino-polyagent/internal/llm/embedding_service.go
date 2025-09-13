package llm

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MovieEmbeddingService specializes embedding generation for movie content
type MovieEmbeddingService struct {
	baseEmbedder    EmbeddingService
	movieProcessor  *MovieContentProcessor
	cache           *EmbeddingCache
	logger          *logrus.Logger
	config          *EmbeddingConfig
	metrics         *EmbeddingMetrics
}

// EmbeddingConfig configures embedding generation
type EmbeddingConfig struct {
	Model              string        `json:"model"`
	MaxTokens          int           `json:"max_tokens"`
	BatchSize          int           `json:"batch_size"`
	Dimensions         int           `json:"dimensions"`
	CacheEnabled       bool          `json:"cache_enabled"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	RequestTimeout     time.Duration `json:"request_timeout"`
	RetryAttempts      int           `json:"retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay"`
	NormalizeVectors   bool          `json:"normalize_vectors"`
}

// EmbeddingMetrics tracks embedding performance
type EmbeddingMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	BatchRequests      int64         `json:"batch_requests"`
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`
	AverageLatency     time.Duration `json:"average_latency"`
	TotalTokens        int64         `json:"total_tokens"`
	ErrorCount         int64         `json:"error_count"`
	LastUpdated        time.Time     `json:"last_updated"`
	mu                 sync.RWMutex
}

// EmbeddingCache caches embeddings to reduce API calls
type EmbeddingCache struct {
	cache      map[string]*CacheEntry
	maxSize    int
	ttl        time.Duration
	mu         sync.RWMutex
}

// CacheEntry represents a cached embedding
type CacheEntry struct {
	Vector    []float64 `json:"vector"`
	CreatedAt time.Time `json:"created_at"`
	Hits      int       `json:"hits"`
}

// MovieContentProcessor processes movie content for embedding
type MovieContentProcessor struct {
	logger *logrus.Logger
}

// OpenAIEmbeddingService implements EmbeddingService using OpenAI API
type OpenAIEmbeddingService struct {
	llmAdapter LLMAdapter
	model      string
	dimensions int
	logger     *logrus.Logger
	metrics    *EmbeddingMetrics
}

// LocalEmbeddingService implements basic local embedding service
type LocalEmbeddingService struct {
	dimensions int
	logger     *logrus.Logger
	metrics    *EmbeddingMetrics
}

// NewMovieEmbeddingService creates a new movie embedding service
func NewMovieEmbeddingService(baseEmbedder EmbeddingService, config *EmbeddingConfig, logger *logrus.Logger) *MovieEmbeddingService {
	if config == nil {
		config = createDefaultEmbeddingConfig()
	}

	var cache *EmbeddingCache
	if config.CacheEnabled {
		cache = NewEmbeddingCache(10000, config.CacheTTL) // Cache up to 10k embeddings
	}

	return &MovieEmbeddingService{
		baseEmbedder:   baseEmbedder,
		movieProcessor: NewMovieContentProcessor(logger),
		cache:         cache,
		logger:        logger,
		config:        config,
		metrics: &EmbeddingMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// GenerateMovieEmbedding generates embedding for movie content
func (mes *MovieEmbeddingService) GenerateMovieEmbedding(ctx context.Context, movie *RecommendedMovie) ([]float64, error) {
	startTime := time.Now()
	
	// Process movie content for embedding
	content := mes.movieProcessor.ProcessMovieContent(movie)
	
	// Check cache first
	if mes.cache != nil {
		if cached := mes.cache.Get(content); cached != nil {
			mes.updateMetrics(time.Since(startTime), true, len(content))
			return cached, nil
		}
	}
	
	// Generate embedding
	embedding, err := mes.baseEmbedder.GenerateEmbedding(ctx, content)
	if err != nil {
		mes.metrics.mu.Lock()
		mes.metrics.ErrorCount++
		mes.metrics.mu.Unlock()
		return nil, fmt.Errorf("failed to generate movie embedding: %w", err)
	}
	
	// Normalize if configured
	if mes.config.NormalizeVectors {
		embedding = normalizeVector(embedding)
	}
	
	// Cache the result
	if mes.cache != nil {
		mes.cache.Set(content, embedding)
	}
	
	mes.updateMetrics(time.Since(startTime), false, len(content))
	
	return embedding, nil
}

// GenerateBatchMovieEmbeddings generates embeddings for multiple movies
func (mes *MovieEmbeddingService) GenerateBatchMovieEmbeddings(ctx context.Context, movies []*RecommendedMovie) ([][]float64, error) {
	startTime := time.Now()
	
	mes.logger.WithField("movie_count", len(movies)).Info("Generating batch movie embeddings")
	
	var contents []string
	var cachedResults = make(map[int][]float64)
	var uncachedIndices []int
	
	// Process movie content and check cache
	for i, movie := range movies {
		content := mes.movieProcessor.ProcessMovieContent(movie)
		contents = append(contents, content)
		
		if mes.cache != nil {
			if cached := mes.cache.Get(content); cached != nil {
				cachedResults[i] = cached
				mes.metrics.mu.Lock()
				mes.metrics.CacheHits++
				mes.metrics.mu.Unlock()
				continue
			}
		}
		
		uncachedIndices = append(uncachedIndices, i)
	}
	
	// Generate embeddings for uncached content
	var embeddings [][]float64
	if len(uncachedIndices) > 0 {
		uncachedContents := make([]string, len(uncachedIndices))
		for i, idx := range uncachedIndices {
			uncachedContents[i] = contents[idx]
		}
		
		var err error
		embeddings, err = mes.baseEmbedder.GenerateBatchEmbeddings(ctx, uncachedContents)
		if err != nil {
			mes.metrics.mu.Lock()
			mes.metrics.ErrorCount++
			mes.metrics.mu.Unlock()
			return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
		}
		
		// Normalize if configured
		if mes.config.NormalizeVectors {
			for i := range embeddings {
				embeddings[i] = normalizeVector(embeddings[i])
			}
		}
		
		// Cache the results
		if mes.cache != nil {
			for i, embedding := range embeddings {
				content := uncachedContents[i]
				mes.cache.Set(content, embedding)
			}
		}
	}
	
	// Combine cached and new results
	results := make([][]float64, len(movies))
	embeddingIdx := 0
	
	for i := range movies {
		if cached, exists := cachedResults[i]; exists {
			results[i] = cached
		} else {
			results[i] = embeddings[embeddingIdx]
			embeddingIdx++
		}
	}
	
	// Update metrics
	totalTokens := 0
	for _, content := range contents {
		totalTokens += len(content)
	}
	
	mes.metrics.mu.Lock()
	mes.metrics.BatchRequests++
	mes.metrics.CacheMisses += int64(len(uncachedIndices))
	mes.metrics.TotalTokens += int64(totalTokens)
	mes.metrics.mu.Unlock()
	
	mes.updateMetrics(time.Since(startTime), false, totalTokens)
	
	mes.logger.WithFields(logrus.Fields{
		"total_movies":    len(movies),
		"cached_results":  len(cachedResults),
		"new_embeddings": len(uncachedIndices),
		"processing_time": time.Since(startTime),
	}).Info("Batch movie embeddings completed")
	
	return results, nil
}

// GetMetrics returns embedding service metrics
func (mes *MovieEmbeddingService) GetMetrics() *EmbeddingMetrics {
	mes.metrics.mu.RLock()
	defer mes.metrics.mu.RUnlock()
	
	metricsCopy := *mes.metrics
	return &metricsCopy
}

// NewMovieContentProcessor creates a new movie content processor
func NewMovieContentProcessor(logger *logrus.Logger) *MovieContentProcessor {
	return &MovieContentProcessor{
		logger: logger,
	}
}

// ProcessMovieContent processes movie data into embedding-ready text
func (mcp *MovieContentProcessor) ProcessMovieContent(movie *RecommendedMovie) string {
	var contentParts []string
	
	// Title (weighted heavily)
	if movie.Title != "" {
		contentParts = append(contentParts, fmt.Sprintf("Title: %s", movie.Title))
		// Repeat title to give it more weight
		contentParts = append(contentParts, movie.Title)
	}
	
	// Genres
	if len(movie.Genres) > 0 {
		genreStr := strings.Join(movie.Genres, ", ")
		contentParts = append(contentParts, fmt.Sprintf("Genres: %s", genreStr))
		// Add genres again for emphasis
		contentParts = append(contentParts, genreStr)
	}
	
	// Year
	if movie.Year > 0 {
		decade := (movie.Year / 10) * 10
		contentParts = append(contentParts, fmt.Sprintf("Year: %d", movie.Year))
		contentParts = append(contentParts, fmt.Sprintf("Decade: %ds", decade))
	}
	
	// Rating (convert to descriptive text)
	if movie.Rating > 0 {
		ratingDesc := mcp.getRatingDescription(movie.Rating)
		contentParts = append(contentParts, fmt.Sprintf("Rating: %.1f (%s)", movie.Rating, ratingDesc))
		contentParts = append(contentParts, ratingDesc)
	}
	
	// Description if available
	if movie.Description != "" {
		contentParts = append(contentParts, fmt.Sprintf("Description: %s", movie.Description))
	}
	
	// Reason/explanation if available
	if movie.Reason != "" {
		contentParts = append(contentParts, fmt.Sprintf("Context: %s", movie.Reason))
	}
	
	// Combine all parts
	content := strings.Join(contentParts, ". ")
	
	// Add semantic enrichment
	content = mcp.enrichContent(movie, content)
	
	return content
}

// getRatingDescription converts numeric rating to descriptive text
func (mcp *MovieContentProcessor) getRatingDescription(rating float64) string {
	switch {
	case rating >= 4.5:
		return "excellent highly rated masterpiece"
	case rating >= 4.0:
		return "very good highly recommended"
	case rating >= 3.5:
		return "good solid choice"
	case rating >= 3.0:
		return "decent average quality"
	case rating >= 2.5:
		return "mediocre mixed reviews"
	default:
		return "poor low rated"
	}
}

// enrichContent adds semantic context to improve embeddings
func (mcp *MovieContentProcessor) enrichContent(movie *RecommendedMovie, content string) string {
	enrichments := []string{}
	
	// Add genre-based enrichments
	for _, genre := range movie.Genres {
		switch strings.ToLower(genre) {
		case "action":
			enrichments = append(enrichments, "exciting thrilling fast-paced adventure")
		case "comedy":
			enrichments = append(enrichments, "funny humorous entertaining lighthearted")
		case "drama":
			enrichments = append(enrichments, "emotional serious compelling character-driven")
		case "horror":
			enrichments = append(enrichments, "scary frightening suspenseful dark")
		case "sci-fi", "science fiction":
			enrichments = append(enrichments, "futuristic technology science space")
		case "romance":
			enrichments = append(enrichments, "love romantic relationship heartwarming")
		case "thriller":
			enrichments = append(enrichments, "suspenseful tense gripping mystery")
		case "fantasy":
			enrichments = append(enrichments, "magical fantasy adventure otherworldly")
		case "crime":
			enrichments = append(enrichments, "criminal investigation police detective")
		case "documentary":
			enrichments = append(enrichments, "factual educational informative real-life")
		}
	}
	
	// Add year-based context
	if movie.Year > 0 {
		switch {
		case movie.Year >= 2020:
			enrichments = append(enrichments, "recent modern contemporary current")
		case movie.Year >= 2010:
			enrichments = append(enrichments, "modern recent 2010s")
		case movie.Year >= 2000:
			enrichments = append(enrichments, "2000s millennium early 2000s")
		case movie.Year >= 1990:
			enrichments = append(enrichments, "1990s nineties classic")
		case movie.Year >= 1980:
			enrichments = append(enrichments, "1980s eighties retro")
		default:
			enrichments = append(enrichments, "classic vintage old-school")
		}
	}
	
	if len(enrichments) > 0 {
		content += ". " + strings.Join(enrichments, " ")
	}
	
	return content
}

// NewEmbeddingCache creates a new embedding cache
func NewEmbeddingCache(maxSize int, ttl time.Duration) *EmbeddingCache {
	return &EmbeddingCache{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

// Get retrieves embedding from cache
func (ec *EmbeddingCache) Get(key string) []float64 {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	entry, exists := ec.cache[key]
	if !exists {
		return nil
	}
	
	// Check TTL
	if time.Since(entry.CreatedAt) > ec.ttl {
		// Delete expired entry (in a real implementation, use a background cleaner)
		delete(ec.cache, key)
		return nil
	}
	
	entry.Hits++
	return entry.Vector
}

// Set stores embedding in cache
func (ec *EmbeddingCache) Set(key string, vector []float64) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	// Evict old entries if at capacity
	if len(ec.cache) >= ec.maxSize {
		ec.evictOldest()
	}
	
	ec.cache[key] = &CacheEntry{
		Vector:    vector,
		CreatedAt: time.Now(),
		Hits:      0,
	}
}

// evictOldest removes the oldest cache entry
func (ec *EmbeddingCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, entry := range ec.cache {
		if oldestKey == "" || entry.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.CreatedAt
		}
	}
	
	if oldestKey != "" {
		delete(ec.cache, oldestKey)
	}
}

// NewOpenAIEmbeddingService creates OpenAI-based embedding service
func NewOpenAIEmbeddingService(llmAdapter LLMAdapter, model string, logger *logrus.Logger) *OpenAIEmbeddingService {
	dimensions := 1536 // Default for text-embedding-ada-002
	if model == "text-embedding-3-small" {
		dimensions = 1536
	} else if model == "text-embedding-3-large" {
		dimensions = 3072
	}
	
	return &OpenAIEmbeddingService{
		llmAdapter: llmAdapter,
		model:      model,
		dimensions: dimensions,
		logger:     logger,
		metrics: &EmbeddingMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// GenerateEmbedding generates embedding using OpenAI API
func (oes *OpenAIEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	startTime := time.Now()
	
	// For demonstration, return mock embedding
	// In real implementation, call OpenAI embedding API
	embedding := generateMockEmbedding(text, oes.dimensions)
	
	oes.updateMetrics(time.Since(startTime), len(text))
	
	return embedding, nil
}

// GenerateBatchEmbeddings generates multiple embeddings
func (oes *OpenAIEmbeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	startTime := time.Now()
	
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embeddings[i] = generateMockEmbedding(text, oes.dimensions)
	}
	
	totalTokens := 0
	for _, text := range texts {
		totalTokens += len(text)
	}
	
	oes.metrics.mu.Lock()
	oes.metrics.BatchRequests++
	oes.metrics.TotalTokens += int64(totalTokens)
	oes.metrics.mu.Unlock()
	
	oes.updateMetrics(time.Since(startTime), totalTokens)
	
	return embeddings, nil
}

// GetDimensions returns embedding dimensions
func (oes *OpenAIEmbeddingService) GetDimensions() int {
	return oes.dimensions
}

// GetModel returns embedding model name
func (oes *OpenAIEmbeddingService) GetModel() string {
	return oes.model
}

// updateMetrics updates OpenAI embedding service metrics
func (oes *OpenAIEmbeddingService) updateMetrics(latency time.Duration, tokenCount int) {
	oes.metrics.mu.Lock()
	defer oes.metrics.mu.Unlock()
	
	oes.metrics.TotalRequests++
	oes.metrics.TotalTokens += int64(tokenCount)
	
	if oes.metrics.TotalRequests == 1 {
		oes.metrics.AverageLatency = latency
	} else {
		oes.metrics.AverageLatency = (oes.metrics.AverageLatency + latency) / 2
	}
	
	oes.metrics.LastUpdated = time.Now()
}

// NewLocalEmbeddingService creates a local embedding service
func NewLocalEmbeddingService(dimensions int, logger *logrus.Logger) *LocalEmbeddingService {
	return &LocalEmbeddingService{
		dimensions: dimensions,
		logger:     logger,
		metrics: &EmbeddingMetrics{
			LastUpdated: time.Now(),
		},
	}
}

// GenerateEmbedding generates simple local embedding
func (les *LocalEmbeddingService) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	startTime := time.Now()
	
	// Generate deterministic embedding based on text
	embedding := generateMockEmbedding(text, les.dimensions)
	
	les.updateMetrics(time.Since(startTime), len(text))
	
	return embedding, nil
}

// GenerateBatchEmbeddings generates multiple local embeddings
func (les *LocalEmbeddingService) GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
	startTime := time.Now()
	
	embeddings := make([][]float64, len(texts))
	for i, text := range texts {
		embeddings[i] = generateMockEmbedding(text, les.dimensions)
	}
	
	totalTokens := 0
	for _, text := range texts {
		totalTokens += len(text)
	}
	
	les.updateMetrics(time.Since(startTime), totalTokens)
	
	return embeddings, nil
}

// GetDimensions returns embedding dimensions
func (les *LocalEmbeddingService) GetDimensions() int {
	return les.dimensions
}

// GetModel returns embedding model name
func (les *LocalEmbeddingService) GetModel() string {
	return "local-embedding-model"
}

// updateMetrics updates local embedding service metrics
func (les *LocalEmbeddingService) updateMetrics(latency time.Duration, tokenCount int) {
	les.metrics.mu.Lock()
	defer les.metrics.mu.Unlock()
	
	les.metrics.TotalRequests++
	les.metrics.TotalTokens += int64(tokenCount)
	
	if les.metrics.TotalRequests == 1 {
		les.metrics.AverageLatency = latency
	} else {
		les.metrics.AverageLatency = (les.metrics.AverageLatency + latency) / 2
	}
	
	les.metrics.LastUpdated = time.Now()
}

// updateMetrics updates movie embedding service metrics
func (mes *MovieEmbeddingService) updateMetrics(latency time.Duration, cacheHit bool, tokenCount int) {
	mes.metrics.mu.Lock()
	defer mes.metrics.mu.Unlock()
	
	mes.metrics.TotalRequests++
	mes.metrics.TotalTokens += int64(tokenCount)
	
	if cacheHit {
		mes.metrics.CacheHits++
	} else {
		mes.metrics.CacheMisses++
	}
	
	if mes.metrics.TotalRequests == 1 {
		mes.metrics.AverageLatency = latency
	} else {
		mes.metrics.AverageLatency = (mes.metrics.AverageLatency + latency) / 2
	}
	
	mes.metrics.LastUpdated = time.Now()
}

// generateMockEmbedding generates a deterministic mock embedding
func generateMockEmbedding(text string, dimensions int) []float64 {
	// Create deterministic embedding based on text content
	// This is for demonstration - real embeddings would come from trained models
	
	rand.Seed(int64(hashString(text)))
	embedding := make([]float64, dimensions)
	
	for i := range embedding {
		embedding[i] = rand.Float64()*2 - 1 // Range [-1, 1]
	}
	
	return normalizeVector(embedding)
}

// hashString creates a simple hash of a string
func hashString(s string) int {
	hash := 0
	for _, char := range s {
		hash = hash*31 + int(char)
	}
	return hash
}

// normalizeVector normalizes a vector to unit length
func normalizeVector(vector []float64) []float64 {
	var norm float64
	for _, val := range vector {
		norm += val * val
	}
	norm = math.Sqrt(norm)
	
	if norm == 0 {
		return vector
	}
	
	normalized := make([]float64, len(vector))
	for i, val := range vector {
		normalized[i] = val / norm
	}
	
	return normalized
}

// createDefaultEmbeddingConfig creates default embedding configuration
func createDefaultEmbeddingConfig() *EmbeddingConfig {
	return &EmbeddingConfig{
		Model:            "text-embedding-ada-002",
		MaxTokens:        8191,
		BatchSize:        100,
		Dimensions:       1536,
		CacheEnabled:     true,
		CacheTTL:         24 * time.Hour,
		RequestTimeout:   30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       time.Second,
		NormalizeVectors: true,
	}
}