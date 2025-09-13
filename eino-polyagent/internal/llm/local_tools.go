package llm

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// LocalToolManager manages local computation tools
type LocalToolManager struct {
	tools   map[string]LocalTool
	logger  *logrus.Logger
	metrics *LocalToolManagerMetrics
}

// LocalToolManagerMetrics tracks manager metrics
type LocalToolManagerMetrics struct {
	TotalExecutions  int64         `json:"total_executions"`
	SuccessfulRuns   int64         `json:"successful_runs"`
	FailedRuns       int64         `json:"failed_runs"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastExecution    time.Time     `json:"last_execution"`
}

// NewLocalToolManager creates a new local tool manager
func NewLocalToolManager(logger *logrus.Logger) *LocalToolManager {
	return &LocalToolManager{
		tools:  make(map[string]LocalTool),
		logger: logger,
		metrics: &LocalToolManagerMetrics{},
	}
}

// ExecuteTask executes a task using local tools
func (ltm *LocalToolManager) ExecuteTask(ctx context.Context, req *HybridExecutionRequest) (*LocalExecutionResult, error) {
	startTime := time.Now()
	
	ltm.logger.WithFields(logrus.Fields{
		"task_type": req.TaskType,
		"task_id":   req.TaskID,
	}).Info("Executing task with local tools")

	// Find appropriate tool
	tool := ltm.findBestTool(req.TaskType)
	if tool == nil {
		return nil, fmt.Errorf("no local tool available for task type: %s", req.TaskType)
	}

	// Prepare local execution request
	localReq := &LocalExecutionRequest{
		TaskType: req.TaskType,
		Data:     req.Data,
		Context:  req.Context,
	}

	// Execute with timeout
	result, err := tool.Execute(ctx, localReq)
	if err != nil {
		ltm.updateMetrics(false, time.Since(startTime))
		return nil, fmt.Errorf("local tool execution failed: %w", err)
	}

	ltm.updateMetrics(true, time.Since(startTime))
	
	ltm.logger.WithFields(logrus.Fields{
		"tool_name":       tool.GetName(),
		"processing_time": time.Since(startTime),
		"confidence":      result.Confidence,
	}).Info("Local tool execution completed")

	return result, nil
}

// findBestTool finds the best tool for a given task type
func (ltm *LocalToolManager) findBestTool(taskType TaskType) LocalTool {
	var bestTool LocalTool
	bestScore := 0.0

	for _, tool := range ltm.tools {
		capabilities := tool.GetCapabilities()
		for _, cap := range capabilities {
			if cap == taskType {
				metrics := tool.GetPerformanceMetrics()
				score := metrics.SuccessRate - (float64(metrics.AverageLatency.Milliseconds()) / 1000.0)
				if score > bestScore {
					bestScore = score
					bestTool = tool
				}
				break
			}
		}
	}

	return bestTool
}

// RegisterTool registers a local tool
func (ltm *LocalToolManager) RegisterTool(tool LocalTool) {
	ltm.tools[tool.GetName()] = tool
	ltm.logger.WithField("tool_name", tool.GetName()).Info("Local tool registered")
}

// updateMetrics updates manager metrics
func (ltm *LocalToolManager) updateMetrics(success bool, duration time.Duration) {
	ltm.metrics.TotalExecutions++
	if success {
		ltm.metrics.SuccessfulRuns++
	} else {
		ltm.metrics.FailedRuns++
	}
	
	// Update average latency
	if ltm.metrics.TotalExecutions == 1 {
		ltm.metrics.AverageLatency = duration
	} else {
		ltm.metrics.AverageLatency = (ltm.metrics.AverageLatency + duration) / 2
	}
	
	ltm.metrics.LastExecution = time.Now()
}

// CollaborativeFilteringTool implements local collaborative filtering
type CollaborativeFilteringTool struct {
	name    string
	logger  *logrus.Logger
	metrics *LocalToolMetrics
}

// NewCollaborativeFilteringTool creates a new collaborative filtering tool
func NewCollaborativeFilteringTool(logger *logrus.Logger) *CollaborativeFilteringTool {
	return &CollaborativeFilteringTool{
		name:   "collaborative_filtering",
		logger: logger,
		metrics: &LocalToolMetrics{
			SuccessRate: 0.85,
			AverageLatency: 50 * time.Millisecond,
			LastUpdated: time.Now(),
		},
	}
}

// GetName returns the tool name
func (cf *CollaborativeFilteringTool) GetName() string {
	return cf.name
}

// GetCapabilities returns supported task types
func (cf *CollaborativeFilteringTool) GetCapabilities() []TaskType {
	return []TaskType{TaskMovieRecommendation, TaskSimilarityCalc}
}

// Execute performs collaborative filtering
func (cf *CollaborativeFilteringTool) Execute(ctx context.Context, req *LocalExecutionRequest) (*LocalExecutionResult, error) {
	startTime := time.Now()
	
	cf.logger.WithField("task_type", req.TaskType).Debug("Executing collaborative filtering")

	// Extract user data
	userID, ok := req.Data["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_id not found in request data")
	}

	topK := 10
	if k, exists := req.Data["top_k"]; exists {
		if kInt, ok := k.(int); ok {
			topK = kInt
		}
	}

	// Simulate collaborative filtering computation
	recommendations := cf.computeCollaborativeFiltering(userID, topK)
	
	cf.updateMetrics(true, time.Since(startTime))

	return &LocalExecutionResult{
		Result:         recommendations,
		Confidence:     0.85,
		ProcessingTime: time.Since(startTime),
		ToolsUsed:      []string{cf.name},
		Metadata: map[string]interface{}{
			"algorithm":    "user_based_cf",
			"similarity":   "pearson",
			"neighborhood": 50,
		},
	}, nil
}

// computeCollaborativeFiltering simulates collaborative filtering computation
func (cf *CollaborativeFilteringTool) computeCollaborativeFiltering(userID string, topK int) []RecommendedMovie {
	// Simulate computation with mock data
	movies := []RecommendedMovie{
		{
			ID: 1,
			Title: "The Matrix",
			Rating: 4.5,
			Genres: []string{"Action", "Sci-Fi"},
			Year: 1999,
			Reason: "Users with similar preferences enjoyed this movie",
		},
		{
			ID: 2, 
			Title: "Inception",
			Rating: 4.7,
			Genres: []string{"Action", "Sci-Fi", "Thriller"},
			Year: 2010,
			Reason: "High correlation with your viewing history",
		},
		{
			ID: 3,
			Title: "Interstellar",
			Rating: 4.6,
			Genres: []string{"Drama", "Sci-Fi"},
			Year: 2014,
			Reason: "Similar users rated this highly",
		},
	}

	// Simulate additional recommendations if topK > 3
	for i := len(movies); i < topK && i < 10; i++ {
		movies = append(movies, RecommendedMovie{
			ID: i+1,
			Title: fmt.Sprintf("Movie %d", i+1),
			Rating: 4.0 + float64(i%10)/10.0,
			Genres: []string{"Action", "Drama"},
			Year: 2020 - i,
			Reason: "Generated based on collaborative filtering",
		})
	}

	// Sort by rating
	sort.Slice(movies, func(i, j int) bool {
		return movies[i].Rating > movies[j].Rating
	})

	if len(movies) > topK {
		movies = movies[:topK]
	}

	return movies
}

// GetPerformanceMetrics returns performance metrics
func (cf *CollaborativeFilteringTool) GetPerformanceMetrics() *LocalToolMetrics {
	return cf.metrics
}

// CanHandle checks if tool can handle the task
func (cf *CollaborativeFilteringTool) CanHandle(taskType TaskType, complexity int) bool {
	capabilities := cf.GetCapabilities()
	for _, cap := range capabilities {
		if cap == taskType {
			// Can handle complexity up to 7
			return complexity <= 7
		}
	}
	return false
}

// updateMetrics updates tool metrics
func (cf *CollaborativeFilteringTool) updateMetrics(success bool, duration time.Duration) {
	cf.metrics.TotalExecutions++
	if success {
		cf.metrics.SuccessRate = (cf.metrics.SuccessRate*float64(cf.metrics.TotalExecutions-1) + 1.0) / float64(cf.metrics.TotalExecutions)
	} else {
		cf.metrics.SuccessRate = (cf.metrics.SuccessRate*float64(cf.metrics.TotalExecutions-1) + 0.0) / float64(cf.metrics.TotalExecutions)
	}
	
	// Update average latency
	if cf.metrics.TotalExecutions == 1 {
		cf.metrics.AverageLatency = duration
	} else {
		cf.metrics.AverageLatency = (cf.metrics.AverageLatency + duration) / 2
	}
	
	cf.metrics.LastUpdated = time.Now()
}

// ContentFilteringTool implements local content-based filtering
type ContentFilteringTool struct {
	name    string
	logger  *logrus.Logger
	metrics *LocalToolMetrics
}

// NewContentFilteringTool creates a new content filtering tool
func NewContentFilteringTool(logger *logrus.Logger) *ContentFilteringTool {
	return &ContentFilteringTool{
		name:   "content_filtering",
		logger: logger,
		metrics: &LocalToolMetrics{
			SuccessRate: 0.90,
			AverageLatency: 30 * time.Millisecond,
			LastUpdated: time.Now(),
		},
	}
}

// GetName returns the tool name
func (ct *ContentFilteringTool) GetName() string {
	return ct.name
}

// GetCapabilities returns supported task types
func (ct *ContentFilteringTool) GetCapabilities() []TaskType {
	return []TaskType{TaskContentFiltering, TaskSimilarityCalc}
}

// Execute performs content-based filtering
func (ct *ContentFilteringTool) Execute(ctx context.Context, req *LocalExecutionRequest) (*LocalExecutionResult, error) {
	startTime := time.Now()
	
	ct.logger.WithField("task_type", req.TaskType).Debug("Executing content filtering")

	// Extract filtering criteria
	criteria, ok := req.Data["criteria"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("criteria not found in request data")
	}

	// Apply content filtering
	filteredMovies := ct.applyContentFiltering(criteria)
	
	ct.updateMetrics(true, time.Since(startTime))

	return &LocalExecutionResult{
		Result:         filteredMovies,
		Confidence:     0.90,
		ProcessingTime: time.Since(startTime),
		ToolsUsed:      []string{ct.name},
		Metadata: map[string]interface{}{
			"algorithm":      "content_based",
			"features_used":  []string{"genre", "year", "rating"},
			"filters_applied": len(criteria),
		},
	}, nil
}

// applyContentFiltering applies content-based filtering
func (ct *ContentFilteringTool) applyContentFiltering(criteria map[string]interface{}) []RecommendedMovie {
	// Mock movie database
	allMovies := []RecommendedMovie{
		{ID: 1, Title: "The Matrix", Rating: 4.5, Genres: []string{"Action", "Sci-Fi"}, Year: 1999},
		{ID: 2, Title: "Inception", Rating: 4.7, Genres: []string{"Action", "Sci-Fi", "Thriller"}, Year: 2010},
		{ID: 3, Title: "Interstellar", Rating: 4.6, Genres: []string{"Drama", "Sci-Fi"}, Year: 2014},
		{ID: 4, Title: "The Godfather", Rating: 4.9, Genres: []string{"Crime", "Drama"}, Year: 1972},
		{ID: 5, Title: "Pulp Fiction", Rating: 4.8, Genres: []string{"Crime", "Drama"}, Year: 1994},
		{ID: 6, Title: "The Dark Knight", Rating: 4.8, Genres: []string{"Action", "Crime", "Drama"}, Year: 2008},
		{ID: 7, Title: "Forrest Gump", Rating: 4.7, Genres: []string{"Drama", "Romance"}, Year: 1994},
		{ID: 8, Title: "The Avengers", Rating: 4.2, Genres: []string{"Action", "Adventure", "Sci-Fi"}, Year: 2012},
	}

	var filtered []RecommendedMovie

	// Apply filters
	for _, movie := range allMovies {
		include := true

		// Genre filter
		if genres, exists := criteria["genres"]; exists {
			if genreList, ok := genres.([]interface{}); ok {
				hasGenre := false
				for _, g := range genreList {
					if genreStr, ok := g.(string); ok {
						for _, movieGenre := range movie.Genres {
							if movieGenre == genreStr {
								hasGenre = true
								break
							}
						}
						if hasGenre {
							break
						}
					}
				}
				if !hasGenre {
					include = false
				}
			}
		}

		// Year range filter
		if yearRange, exists := criteria["year_range"]; exists {
			if yr, ok := yearRange.(map[string]interface{}); ok {
				if minYear, exists := yr["min"]; exists {
					if min, ok := minYear.(float64); ok && movie.Year < int(min) {
						include = false
					}
				}
				if maxYear, exists := yr["max"]; exists {
					if max, ok := maxYear.(float64); ok && movie.Year > int(max) {
						include = false
					}
				}
			}
		}

		// Rating filter
		if minRating, exists := criteria["min_rating"]; exists {
			if rating, ok := minRating.(float64); ok && movie.Rating < rating {
				include = false
			}
		}

		if include {
			filtered = append(filtered, movie)
		}
	}

	// Sort by rating
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Rating > filtered[j].Rating
	})

	return filtered
}

// GetPerformanceMetrics returns performance metrics
func (ct *ContentFilteringTool) GetPerformanceMetrics() *LocalToolMetrics {
	return ct.metrics
}

// CanHandle checks if tool can handle the task
func (ct *ContentFilteringTool) CanHandle(taskType TaskType, complexity int) bool {
	capabilities := ct.GetCapabilities()
	for _, cap := range capabilities {
		if cap == taskType {
			// Can handle any complexity for content filtering
			return true
		}
	}
	return false
}

// updateMetrics updates tool metrics
func (ct *ContentFilteringTool) updateMetrics(success bool, duration time.Duration) {
	ct.metrics.TotalExecutions++
	if success {
		ct.metrics.SuccessRate = (ct.metrics.SuccessRate*float64(ct.metrics.TotalExecutions-1) + 1.0) / float64(ct.metrics.TotalExecutions)
	} else {
		ct.metrics.SuccessRate = (ct.metrics.SuccessRate*float64(ct.metrics.TotalExecutions-1) + 0.0) / float64(ct.metrics.TotalExecutions)
	}
	
	// Update average latency
	if ct.metrics.TotalExecutions == 1 {
		ct.metrics.AverageLatency = duration
	} else {
		ct.metrics.AverageLatency = (ct.metrics.AverageLatency + duration) / 2
	}
	
	ct.metrics.LastUpdated = time.Now()
}

// SimilarityCalculationTool implements local similarity calculations
type SimilarityCalculationTool struct {
	name    string
	logger  *logrus.Logger
	metrics *LocalToolMetrics
}

// NewSimilarityCalculationTool creates a new similarity calculation tool
func NewSimilarityCalculationTool(logger *logrus.Logger) *SimilarityCalculationTool {
	return &SimilarityCalculationTool{
		name:   "similarity_calculation",
		logger: logger,
		metrics: &LocalToolMetrics{
			SuccessRate: 0.95,
			AverageLatency: 20 * time.Millisecond,
			LastUpdated: time.Now(),
		},
	}
}

// GetName returns the tool name
func (sc *SimilarityCalculationTool) GetName() string {
	return sc.name
}

// GetCapabilities returns supported task types
func (sc *SimilarityCalculationTool) GetCapabilities() []TaskType {
	return []TaskType{TaskSimilarityCalc}
}

// Execute performs similarity calculations
func (sc *SimilarityCalculationTool) Execute(ctx context.Context, req *LocalExecutionRequest) (*LocalExecutionResult, error) {
	startTime := time.Now()
	
	sc.logger.WithField("task_type", req.TaskType).Debug("Executing similarity calculation")

	// Extract similarity request data
	sourceItem, ok := req.Data["source_item"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("source_item not found in request data")
	}

	targetItems, ok := req.Data["target_items"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("target_items not found in request data")
	}

	// Calculate similarities
	similarities := sc.calculateSimilarities(sourceItem, targetItems)
	
	sc.updateMetrics(true, time.Since(startTime))

	return &LocalExecutionResult{
		Result:         similarities,
		Confidence:     0.95,
		ProcessingTime: time.Since(startTime),
		ToolsUsed:      []string{sc.name},
		Metadata: map[string]interface{}{
			"algorithm":      "cosine_similarity",
			"features_used":  []string{"genre_vector", "rating", "year_normalized"},
			"items_compared": len(targetItems),
		},
	}, nil
}

// calculateSimilarities calculates similarity scores between items
func (sc *SimilarityCalculationTool) calculateSimilarities(source map[string]interface{}, targets []interface{}) []map[string]interface{} {
	var similarities []map[string]interface{}

	// Extract source features
	sourceGenres := sc.extractGenres(source)
	sourceRating := sc.extractRating(source)
	sourceYear := sc.extractYear(source)

	for i, target := range targets {
		if targetMap, ok := target.(map[string]interface{}); ok {
			// Extract target features
			targetGenres := sc.extractGenres(targetMap)
			targetRating := sc.extractRating(targetMap)
			targetYear := sc.extractYear(targetMap)

			// Calculate similarity components
			genreSim := sc.calculateGenreSimilarity(sourceGenres, targetGenres)
			ratingSim := sc.calculateRatingSimilarity(sourceRating, targetRating)
			yearSim := sc.calculateYearSimilarity(sourceYear, targetYear)

			// Weighted combination
			overallSim := (genreSim*0.5 + ratingSim*0.3 + yearSim*0.2)

			similarities = append(similarities, map[string]interface{}{
				"item_id":         fmt.Sprintf("item_%d", i),
				"similarity":      overallSim,
				"genre_sim":       genreSim,
				"rating_sim":      ratingSim,
				"year_sim":        yearSim,
				"target_item":     targetMap,
			})
		}
	}

	// Sort by similarity score
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i]["similarity"].(float64) > similarities[j]["similarity"].(float64)
	})

	return similarities
}

// extractGenres extracts genres from item data
func (sc *SimilarityCalculationTool) extractGenres(item map[string]interface{}) []string {
	if genres, exists := item["genres"]; exists {
		if genreList, ok := genres.([]interface{}); ok {
			var result []string
			for _, g := range genreList {
				if genreStr, ok := g.(string); ok {
					result = append(result, genreStr)
				}
			}
			return result
		}
	}
	return []string{}
}

// extractRating extracts rating from item data
func (sc *SimilarityCalculationTool) extractRating(item map[string]interface{}) float64 {
	if rating, exists := item["rating"]; exists {
		if r, ok := rating.(float64); ok {
			return r
		}
	}
	return 0.0
}

// extractYear extracts year from item data
func (sc *SimilarityCalculationTool) extractYear(item map[string]interface{}) int {
	if year, exists := item["year"]; exists {
		if y, ok := year.(float64); ok {
			return int(y)
		}
	}
	return 0
}

// calculateGenreSimilarity calculates genre similarity using Jaccard index
func (sc *SimilarityCalculationTool) calculateGenreSimilarity(genres1, genres2 []string) float64 {
	if len(genres1) == 0 || len(genres2) == 0 {
		return 0.0
	}

	// Convert to sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, g := range genres1 {
		set1[g] = true
	}
	for _, g := range genres2 {
		set2[g] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := make(map[string]bool)
	
	for g := range set1 {
		union[g] = true
		if set2[g] {
			intersection++
		}
	}
	for g := range set2 {
		union[g] = true
	}

	// Jaccard index
	return float64(intersection) / float64(len(union))
}

// calculateRatingSimilarity calculates rating similarity
func (sc *SimilarityCalculationTool) calculateRatingSimilarity(rating1, rating2 float64) float64 {
	maxRating := 5.0
	diff := math.Abs(rating1 - rating2)
	return 1.0 - (diff / maxRating)
}

// calculateYearSimilarity calculates year similarity
func (sc *SimilarityCalculationTool) calculateYearSimilarity(year1, year2 int) float64 {
	if year1 == 0 || year2 == 0 {
		return 0.5 // Neutral similarity if year unknown
	}
	
	diff := math.Abs(float64(year1 - year2))
	maxDiff := 50.0 // Movies 50 years apart have 0 similarity
	
	if diff >= maxDiff {
		return 0.0
	}
	
	return 1.0 - (diff / maxDiff)
}

// GetPerformanceMetrics returns performance metrics
func (sc *SimilarityCalculationTool) GetPerformanceMetrics() *LocalToolMetrics {
	return sc.metrics
}

// CanHandle checks if tool can handle the task
func (sc *SimilarityCalculationTool) CanHandle(taskType TaskType, complexity int) bool {
	capabilities := sc.GetCapabilities()
	for _, cap := range capabilities {
		if cap == taskType {
			// Can handle any complexity for similarity calculation
			return true
		}
	}
	return false
}

// updateMetrics updates tool metrics
func (sc *SimilarityCalculationTool) updateMetrics(success bool, duration time.Duration) {
	sc.metrics.TotalExecutions++
	if success {
		sc.metrics.SuccessRate = (sc.metrics.SuccessRate*float64(sc.metrics.TotalExecutions-1) + 1.0) / float64(sc.metrics.TotalExecutions)
	} else {
		sc.metrics.SuccessRate = (sc.metrics.SuccessRate*float64(sc.metrics.TotalExecutions-1) + 0.0) / float64(sc.metrics.TotalExecutions)
	}
	
	// Update average latency
	if sc.metrics.TotalExecutions == 1 {
		sc.metrics.AverageLatency = duration
	} else {
		sc.metrics.AverageLatency = (sc.metrics.AverageLatency + duration) / 2
	}
	
	sc.metrics.LastUpdated = time.Now()
}