package llm

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

// ConfigManager manages LLM configurations
type ConfigManager struct {
	configs map[string]*LLMAdapterConfig
	logger  *logrus.Logger
}

// NewConfigManager creates a new config manager
func NewConfigManager(logger *logrus.Logger) *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]*LLMAdapterConfig),
		logger:  logger,
	}
}

// LoadConfig loads configuration from file or environment
func (cm *ConfigManager) LoadConfig(configPath string) (*LLMAdapterConfig, error) {
	// Check if config exists in cache
	if config, exists := cm.configs[configPath]; exists {
		return config, nil
	}

	var config *LLMAdapterConfig
	var err error

	if configPath != "" && fileExists(configPath) {
		// Load from file
		config, err = cm.loadFromFile(configPath)
	} else {
		// Load from environment variables
		config, err = cm.loadFromEnv()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Validate and set defaults
	if err := cm.validateAndSetDefaults(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	// Cache the config
	cm.configs[configPath] = config
	cm.logger.Infof("Loaded LLM configuration from %s", configPath)

	return config, nil
}

// loadFromFile loads configuration from JSON file
func (cm *ConfigManager) loadFromFile(configPath string) (*LLMAdapterConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config LLMAdapterConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

// loadFromEnv loads configuration from environment variables
func (cm *ConfigManager) loadFromEnv() (*LLMAdapterConfig, error) {
	config := &LLMAdapterConfig{
		LoadBalancing:    getEnvBool("LLM_LOAD_BALANCING", false),
		CostOptimization: getEnvBool("LLM_COST_OPTIMIZATION", true),
	}

	// Load primary configuration
	primaryConfig, err := cm.loadProviderConfigFromEnv("PRIMARY")
	if err != nil {
		return nil, fmt.Errorf("failed to load primary config: %w", err)
	}
	config.Primary = *primaryConfig

	// Load fallback configurations
	for i := 1; i <= 5; i++ { // Support up to 5 fallback providers
		prefix := fmt.Sprintf("FALLBACK_%d", i)
		if !hasEnvPrefix(prefix) {
			break
		}

		fallbackConfig, err := cm.loadProviderConfigFromEnv(prefix)
		if err != nil {
			cm.logger.Warnf("Failed to load fallback config %d: %v", i, err)
			continue
		}
		config.Fallback = append(config.Fallback, *fallbackConfig)
	}

	// Load budget configuration if available
	if hasEnvPrefix("BUDGET") {
		budgetConfig, err := cm.loadProviderConfigFromEnv("BUDGET")
		if err != nil {
			cm.logger.Warnf("Failed to load budget config: %v", err)
		} else {
			config.Budget = budgetConfig
		}
	}

	return config, nil
}

// loadProviderConfigFromEnv loads a single provider config from environment
func (cm *ConfigManager) loadProviderConfigFromEnv(prefix string) (*LLMConfig, error) {
	provider := getEnvString(fmt.Sprintf("LLM_%s_PROVIDER", prefix), "")
	if provider == "" {
		return nil, fmt.Errorf("provider not specified for %s", prefix)
	}

	apiKey := getEnvString(fmt.Sprintf("LLM_%s_API_KEY", prefix), "")
	if apiKey == "" {
		return nil, fmt.Errorf("API key not specified for %s", prefix)
	}

	config := &LLMConfig{
		Provider:    LLMProvider(provider),
		Model:       getEnvString(fmt.Sprintf("LLM_%s_MODEL", prefix), cm.getDefaultModel(LLMProvider(provider))),
		APIKey:      apiKey,
		BaseURL:     getEnvString(fmt.Sprintf("LLM_%s_BASE_URL", prefix), ""),
		Timeout:     time.Duration(getEnvInt(fmt.Sprintf("LLM_%s_TIMEOUT", prefix), 30)) * time.Second,
		MaxRetries:  getEnvInt(fmt.Sprintf("LLM_%s_MAX_RETRIES", prefix), 3),
		Temperature: getEnvFloat(fmt.Sprintf("LLM_%s_TEMPERATURE", prefix), 0.7),
		MaxTokens:   getEnvInt(fmt.Sprintf("LLM_%s_MAX_TOKENS", prefix), 4096),
	}

	return config, nil
}

// validateAndSetDefaults validates config and sets defaults
func (cm *ConfigManager) validateAndSetDefaults(config *LLMAdapterConfig) error {
	// Validate primary config
	if err := cm.validateProviderConfig(&config.Primary); err != nil {
		return fmt.Errorf("primary config validation failed: %w", err)
	}

	// Validate fallback configs
	for i, fallback := range config.Fallback {
		if err := cm.validateProviderConfig(&fallback); err != nil {
			return fmt.Errorf("fallback config %d validation failed: %w", i, err)
		}
	}

	// Validate budget config if present
	if config.Budget != nil {
		if err := cm.validateProviderConfig(config.Budget); err != nil {
			return fmt.Errorf("budget config validation failed: %w", err)
		}
	}

	return nil
}

// validateProviderConfig validates a single provider config
func (cm *ConfigManager) validateProviderConfig(config *LLMConfig) error {
	if config.Provider == "" {
		return fmt.Errorf("provider cannot be empty")
	}

	if config.Model == "" {
		config.Model = cm.getDefaultModel(config.Provider)
	}

	if config.APIKey == "" {
		return fmt.Errorf("API key cannot be empty for provider %s", config.Provider)
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if config.MaxTokens == 0 {
		config.MaxTokens = 4096
	}

	return nil
}

// getDefaultModel returns the default model for a provider
func (cm *ConfigManager) getDefaultModel(provider LLMProvider) string {
	switch provider {
	case ProviderOpenAI:
		return "gpt-4o"
	case ProviderClaude:
		return "claude-3-5-sonnet-20241022"
	case ProviderQwen:
		return "qwen2.5-72b-instruct"
	case ProviderK2, ProviderOpenRouter:
		return "liquid/lfm-40b"
	default:
		return ""
	}
}

// SaveConfig saves configuration to file
func (cm *ConfigManager) SaveConfig(config *LLMAdapterConfig, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Update cache
	cm.configs[configPath] = config
	cm.logger.Infof("Saved LLM configuration to %s", configPath)

	return nil
}

// CreateDefaultConfig creates a default configuration
func (cm *ConfigManager) CreateDefaultConfig() *LLMAdapterConfig {
	return &LLMAdapterConfig{
		Primary: LLMConfig{
			Provider:    ProviderOpenAI,
			Model:       "gpt-4o",
			APIKey:      "", // Must be provided
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		Fallback: []LLMConfig{
			{
				Provider:    ProviderClaude,
				Model:       "claude-3-5-sonnet-20241022",
				APIKey:      "", // Must be provided
				Timeout:     30 * time.Second,
				MaxRetries:  3,
				Temperature: 0.7,
				MaxTokens:   4096,
			},
		},
		Budget: &LLMConfig{
			Provider:    ProviderQwen,
			Model:       "qwen2.5-72b-instruct",
			APIKey:      "", // Must be provided
			Timeout:     30 * time.Second,
			MaxRetries:  3,
			Temperature: 0.7,
			MaxTokens:   4096,
		},
		LoadBalancing:    false,
		CostOptimization: true,
	}
}

// Helper functions

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasEnvPrefix(prefix string) bool {
	key := fmt.Sprintf("LLM_%s_PROVIDER", prefix)
	return os.Getenv(key) != ""
}

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue := parseInt(value); intValue != 0 {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue := parseFloat(value); floatValue != 0 {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

// Simple parsing functions (would use strconv in real implementation)
func parseInt(s string) int {
	// TODO: Implement proper parsing
	return 0
}

func parseFloat(s string) float64 {
	// TODO: Implement proper parsing
	return 0.0
}