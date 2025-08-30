package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置结构
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	AI       AIConfig       `mapstructure:"ai"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	Mode         string        `mapstructure:"mode"` // debug, release
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	SSLMode      string `mapstructure:"sslmode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// AIConfig AI模型配置
type AIConfig struct {
	Models      map[string]ModelConfig `mapstructure:"models"`
	DefaultRoute string                `mapstructure:"default_route"`
	Timeout     time.Duration         `mapstructure:"timeout"`
}

// ModelConfig 单个模型配置
type ModelConfig struct {
	Provider    string            `mapstructure:"provider"`
	APIKey      string            `mapstructure:"api_key"`
	BaseURL     string            `mapstructure:"base_url"`
	ModelName   string            `mapstructure:"model_name"`
	MaxTokens   int               `mapstructure:"max_tokens"`
	Temperature float32           `mapstructure:"temperature"`
	Timeout     time.Duration     `mapstructure:"timeout"`
	RateLimit   int               `mapstructure:"rate_limit"`
	Priority    int               `mapstructure:"priority"`
	Capabilities []string         `mapstructure:"capabilities"`
	CostPer1K   CostConfig        `mapstructure:"cost_per_1k"`
	Metadata    map[string]string `mapstructure:"metadata"`
}

// CostConfig 成本配置
type CostConfig struct {
	Input  float64 `mapstructure:"input"`
	Output float64 `mapstructure:"output"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	JWT          JWTConfig      `mapstructure:"jwt"`
	RateLimit    RateLimitConfig `mapstructure:"rate_limit"`
	CORS         CORSConfig     `mapstructure:"cors"`
	Medical      MedicalConfig  `mapstructure:"medical"`
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	SecretKey      string        `mapstructure:"secret_key"`
	ExpirationTime time.Duration `mapstructure:"expiration_time"`
	Issuer         string        `mapstructure:"issuer"`
	Algorithm      string        `mapstructure:"algorithm"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	RequestsPerMin int           `mapstructure:"requests_per_min"`
	BurstSize      int           `mapstructure:"burst_size"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// MedicalConfig 医疗安全配置
type MedicalConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	StrictMode  bool     `mapstructure:"strict_mode"`
	Patterns    []string `mapstructure:"patterns"`
	Disclaimer  string   `mapstructure:"disclaimer"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, text
	Output     string `mapstructure:"output"` // stdout, file
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

// MetricsConfig 监控配置
type MetricsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Port       int    `mapstructure:"port"`
	Path       string `mapstructure:"path"`
	Namespace  string `mapstructure:"namespace"`
}

// Load 加载配置文件
func Load(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	
	// 设置默认值
	setDefaults()
	
	// 环境变量绑定
	bindEnvVars()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	// 验证配置
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	
	return &config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// 服务器默认配置
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "120s")
	viper.SetDefault("server.mode", "debug")
	
	// 数据库默认配置
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 20)
	viper.SetDefault("database.max_idle_conns", 10)
	
	// Redis默认配置
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	
	// AI配置默认值
	viper.SetDefault("ai.default_route", "balanced")
	viper.SetDefault("ai.timeout", "30s")
	
	// 安全配置默认值
	viper.SetDefault("security.jwt.expiration_time", "24h")
	viper.SetDefault("security.jwt.issuer", "polyagent")
	viper.SetDefault("security.jwt.algorithm", "HS256")
	viper.SetDefault("security.rate_limit.enabled", true)
	viper.SetDefault("security.rate_limit.requests_per_min", 60)
	viper.SetDefault("security.rate_limit.burst_size", 10)
	viper.SetDefault("security.rate_limit.cleanup_interval", "1m")
	
	// 医疗安全默认配置
	viper.SetDefault("security.medical.enabled", true)
	viper.SetDefault("security.medical.strict_mode", false)
	
	// 日志默认配置
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
	
	// 监控默认配置
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.port", 9090)
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.namespace", "polyagent")
}

// bindEnvVars 绑定环境变量
func bindEnvVars() {
	// AI模型API密钥
	viper.BindEnv("ai.models.openai.api_key", "OPENAI_API_KEY")
	viper.BindEnv("ai.models.anthropic.api_key", "ANTHROPIC_API_KEY")
	viper.BindEnv("ai.models.openrouter.api_key", "OPENROUTER_API_KEY")
	viper.BindEnv("ai.models.glm.api_key", "GLM_API_KEY")
	
	// 数据库配置
	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	
	// Redis配置
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	
	// 安全配置
	viper.BindEnv("security.jwt.secret_key", "JWT_SECRET_KEY")
	
	// 服务器配置
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("server.mode", "GIN_MODE")
}

// validate 验证配置
func validate(config *Config) error {
	// 验证必需的配置项
	if config.Security.JWT.SecretKey == "" {
		return fmt.Errorf("JWT密钥不能为空")
	}
	
	if config.Database.Host == "" {
		return fmt.Errorf("数据库主机不能为空")
	}
	
	// 验证AI模型配置
	if len(config.AI.Models) == 0 {
		return fmt.Errorf("至少需要配置一个AI模型")
	}
	
	for name, model := range config.AI.Models {
		if model.Provider == "" {
			return fmt.Errorf("模型 %s 的提供商不能为空", name)
		}
		if model.APIKey == "" {
			return fmt.Errorf("模型 %s 的API密钥不能为空", name)
		}
	}
	
	return nil
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr 获取Redis连接地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器监听地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsProduction 是否生产环境
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "release" || os.Getenv("GO_ENV") == "production"
}

// LoadFromEnv 从环境变量加载配置（用于简化部署）
func LoadFromEnv() (*Config, error) {
	config := &Config{}
	
	// 基本服务器配置
	config.Server.Host = getEnvOrDefault("HOST", "0.0.0.0")
	config.Server.Port = getIntEnvOrDefault("PORT", 8080)
	config.Server.Mode = getEnvOrDefault("GIN_MODE", "debug")
	
	// AI配置
	config.AI.Models = make(map[string]ModelConfig)
	
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.AI.Models["openai"] = ModelConfig{
			Provider:  "openai",
			APIKey:    apiKey,
			BaseURL:   getEnvOrDefault("OPENAI_BASE_URL", "https://api.openai.com/v1"),
			ModelName: getEnvOrDefault("OPENAI_MODEL", "gpt-4"),
			MaxTokens: getIntEnvOrDefault("OPENAI_MAX_TOKENS", 2000),
		}
	}
	
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.AI.Models["anthropic"] = ModelConfig{
			Provider:  "anthropic",
			APIKey:    apiKey,
			BaseURL:   getEnvOrDefault("ANTHROPIC_BASE_URL", "https://api.anthropic.com"),
			ModelName: getEnvOrDefault("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),
			MaxTokens: getIntEnvOrDefault("ANTHROPIC_MAX_TOKENS", 2000),
		}
	}
	
	// JWT配置
	config.Security.JWT.SecretKey = os.Getenv("JWT_SECRET_KEY")
	if config.Security.JWT.SecretKey == "" {
		config.Security.JWT.SecretKey = "default-secret-key-change-in-production"
	}
	
	return config, validate(config)
}

// 辅助函数
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := fmt.Scanf(value, "%d"); err == nil {
			return intValue
		}
	}
	return defaultValue
}