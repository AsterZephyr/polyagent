package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/polyagent/eino-polyagent/internal/config"
	"github.com/sirupsen/logrus"
)

type RoutingStrategy string

const (
	StrategyBalanced    RoutingStrategy = "balanced"
	StrategyCostOptimal RoutingStrategy = "cost_optimal"
	StrategyPerformance RoutingStrategy = "performance"
	StrategyRoundRobin  RoutingStrategy = "round_robin"
	StrategyFailover    RoutingStrategy = "failover"
)

type ModelHealth struct {
	Available    bool    `json:"available"`
	Latency      float64 `json:"latency"`
	ErrorRate    float64 `json:"error_rate"`
	Requests     int64   `json:"requests"`
	LastCheck    time.Time `json:"last_check"`
	CostPer1K    float64 `json:"cost_per_1k"`
	Priority     int     `json:"priority"`
}

type RouteRequest struct {
	Messages        []schema.Message `json:"messages"`
	ModelPreference string          `json:"model_preference,omitempty"`
	Strategy        RoutingStrategy `json:"strategy"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
	Temperature     float32         `json:"temperature,omitempty"`
	UserID          string          `json:"user_id"`
	SessionID       string          `json:"session_id"`
}

type RouteResponse struct {
	SelectedModel string           `json:"selected_model"`
	Response      *schema.Message  `json:"response"`
	Latency       time.Duration    `json:"latency"`
	Cost          float64          `json:"cost"`
	TokensUsed    int              `json:"tokens_used"`
}

type ModelRouter struct {
	models          map[string]model.ChatModel
	healthStatus    map[string]*ModelHealth
	config          *config.Config
	logger          *logrus.Logger
	mu              sync.RWMutex
	rrCounter       int64
	healthCheckTicker *time.Ticker
	stopChan        chan struct{}
}

func NewModelRouter(cfg *config.Config, logger *logrus.Logger) *ModelRouter {
	router := &ModelRouter{
		models:       make(map[string]model.ChatModel),
		healthStatus: make(map[string]*ModelHealth),
		config:       cfg,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}

	router.initializeModels()
	router.startHealthCheck()
	
	return router
}

func (r *ModelRouter) initializeModels() {
	for name, modelConfig := range r.config.AI.Models {
		chatModel, err := r.createChatModel(modelConfig)
		if err != nil {
			r.logger.WithFields(logrus.Fields{
				"model": name,
				"error": err,
			}).Error("Failed to initialize model")
			continue
		}

		r.models[name] = chatModel
		r.healthStatus[name] = &ModelHealth{
			Available:    true,
			Latency:      0,
			ErrorRate:    0,
			Requests:     0,
			LastCheck:    time.Now(),
			CostPer1K:    modelConfig.CostPer1K.Input,
			Priority:     modelConfig.Priority,
		}

		r.logger.WithField("model", name).Info("Model initialized successfully")
	}
}

func (r *ModelRouter) createChatModel(cfg config.ModelConfig) (model.ChatModel, error) {
	switch cfg.Provider {
	case "openai":
		return model.NewOpenAIChatModel(&model.OpenAIChatModelConfig{
			Model:       cfg.ModelName,
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
		})
	case "anthropic":
		return model.NewAnthropicChatModel(&model.AnthropicChatModelConfig{
			Model:       cfg.ModelName,
			APIKey:      cfg.APIKey,
			BaseURL:     cfg.BaseURL,
			MaxTokens:   cfg.MaxTokens,
			Temperature: cfg.Temperature,
		})
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

func (r *ModelRouter) Route(ctx context.Context, req *RouteRequest) (*RouteResponse, error) {
	start := time.Now()

	selectedModel, err := r.selectModel(req)
	if err != nil {
		return nil, fmt.Errorf("model selection failed: %w", err)
	}

	chatModel, exists := r.models[selectedModel]
	if !exists {
		return nil, fmt.Errorf("selected model not found: %s", selectedModel)
	}

	response, err := chatModel.Generate(ctx, req.Messages, &model.GenerateOptions{
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	})
	if err != nil {
		r.updateModelHealth(selectedModel, false, time.Since(start))
		return nil, fmt.Errorf("model generation failed: %w", err)
	}

	latency := time.Since(start)
	r.updateModelHealth(selectedModel, true, latency)

	cost := r.calculateCost(selectedModel, response.TokensUsed)
	
	return &RouteResponse{
		SelectedModel: selectedModel,
		Response:      response,
		Latency:       latency,
		Cost:          cost,
		TokensUsed:    response.TokensUsed,
	}, nil
}

func (r *ModelRouter) selectModel(req *RouteRequest) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	availableModels := r.getAvailableModels()
	if len(availableModels) == 0 {
		return "", fmt.Errorf("no available models")
	}

	if req.ModelPreference != "" {
		if _, exists := r.models[req.ModelPreference]; exists {
			if r.healthStatus[req.ModelPreference].Available {
				return req.ModelPreference, nil
			}
		}
	}

	switch req.Strategy {
	case StrategyCostOptimal:
		return r.selectByCost(availableModels), nil
	case StrategyPerformance:
		return r.selectByPerformance(availableModels), nil
	case StrategyRoundRobin:
		return r.selectRoundRobin(availableModels), nil
	case StrategyFailover:
		return r.selectByPriority(availableModels), nil
	default:
		return r.selectBalanced(availableModels), nil
	}
}

func (r *ModelRouter) getAvailableModels() []string {
	var available []string
	for name, health := range r.healthStatus {
		if health.Available {
			available = append(available, name)
		}
	}
	return available
}

func (r *ModelRouter) selectByCost(models []string) string {
	if len(models) == 0 {
		return ""
	}

	bestModel := models[0]
	lowestCost := r.healthStatus[bestModel].CostPer1K

	for _, model := range models[1:] {
		if cost := r.healthStatus[model].CostPer1K; cost < lowestCost {
			lowestCost = cost
			bestModel = model
		}
	}

	return bestModel
}

func (r *ModelRouter) selectByPerformance(models []string) string {
	if len(models) == 0 {
		return ""
	}

	bestModel := models[0]
	lowestLatency := r.healthStatus[bestModel].Latency

	for _, model := range models[1:] {
		if latency := r.healthStatus[model].Latency; latency < lowestLatency {
			lowestLatency = latency
			bestModel = model
		}
	}

	return bestModel
}

func (r *ModelRouter) selectRoundRobin(models []string) string {
	if len(models) == 0 {
		return ""
	}

	r.rrCounter++
	index := int(r.rrCounter-1) % len(models)
	return models[index]
}

func (r *ModelRouter) selectByPriority(models []string) string {
	if len(models) == 0 {
		return ""
	}

	bestModel := models[0]
	highestPriority := r.healthStatus[bestModel].Priority

	for _, model := range models[1:] {
		if priority := r.healthStatus[model].Priority; priority > highestPriority {
			highestPriority = priority
			bestModel = model
		}
	}

	return bestModel
}

func (r *ModelRouter) selectBalanced(models []string) string {
	type modelScore struct {
		name  string
		score float64
	}

	var scores []modelScore
	for _, model := range models {
		health := r.healthStatus[model]
		score := float64(health.Priority) * 0.4
		score += (1.0 - health.ErrorRate) * 0.3
		score += (1000.0 - health.Latency) / 1000.0 * 0.2
		score += (1.0 / (health.CostPer1K + 0.001)) * 0.1
		
		scores = append(scores, modelScore{name: model, score: score})
	}

	bestModel := scores[0]
	for _, s := range scores[1:] {
		if s.score > bestModel.score {
			bestModel = s
		}
	}

	return bestModel.name
}

func (r *ModelRouter) updateModelHealth(model string, success bool, latency time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	health, exists := r.healthStatus[model]
	if !exists {
		return
	}

	health.Requests++
	health.Latency = (health.Latency + latency.Seconds()) / 2.0
	health.LastCheck = time.Now()

	if !success {
		health.ErrorRate = (health.ErrorRate*0.9 + 0.1)
		if health.ErrorRate > 0.5 {
			health.Available = false
		}
	} else {
		health.ErrorRate = health.ErrorRate * 0.95
		if health.ErrorRate < 0.1 {
			health.Available = true
		}
	}
}

func (r *ModelRouter) calculateCost(model string, tokens int) float64 {
	health, exists := r.healthStatus[model]
	if !exists {
		return 0
	}

	return health.CostPer1K * float64(tokens) / 1000.0
}

func (r *ModelRouter) GetHealth(model string) (*ModelHealth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if model == "" {
		healthMap := make(map[string]*ModelHealth)
		for name, health := range r.healthStatus {
			healthMap[name] = health
		}
		return &ModelHealth{}, nil
	}

	health, exists := r.healthStatus[model]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", model)
	}

	return health, nil
}

func (r *ModelRouter) startHealthCheck() {
	r.healthCheckTicker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-r.healthCheckTicker.C:
				r.performHealthCheck()
			case <-r.stopChan:
				return
			}
		}
	}()
}

func (r *ModelRouter) performHealthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for name, chatModel := range r.models {
		go func(modelName string, model model.ChatModel) {
			start := time.Now()
			
			testMessages := []schema.Message{
				{Role: "user", Content: "test"},
			}

			_, err := model.Generate(ctx, testMessages, &model.GenerateOptions{
				MaxTokens:   1,
				Temperature: 0.1,
			})

			latency := time.Since(start)
			r.updateModelHealth(modelName, err == nil, latency)

			r.logger.WithFields(logrus.Fields{
				"model":   modelName,
				"latency": latency,
				"healthy": err == nil,
			}).Debug("Health check completed")
		}(name, chatModel)
	}
}

func (r *ModelRouter) Stop() {
	if r.healthCheckTicker != nil {
		r.healthCheckTicker.Stop()
	}
	close(r.stopChan)
}