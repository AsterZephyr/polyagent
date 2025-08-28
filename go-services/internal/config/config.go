package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	PythonAI  PythonAIConfig  `mapstructure:"python_ai"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Log       LogConfig       `mapstructure:"log"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	MaxBodySize  int64  `mapstructure:"max_body_size"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	SSLMode      string `mapstructure:"ssl_mode"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxLifetime  int    `mapstructure:"max_lifetime"`
}

type RedisConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Password    string `mapstructure:"password"`
	Database    int    `mapstructure:"database"`
	MaxRetries  int    `mapstructure:"max_retries"`
	PoolSize    int    `mapstructure:"pool_size"`
	MinIdleConn int    `mapstructure:"min_idle_conn"`
}

type PythonAIConfig struct {
	BaseURL        string `mapstructure:"base_url"`
	Timeout        int    `mapstructure:"timeout"`
	MaxConcurrency int    `mapstructure:"max_concurrency"`
	RetryAttempts  int    `mapstructure:"retry_attempts"`
}

type AuthConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	TokenExpiry   int    `mapstructure:"token_expiry"`
	RefreshExpiry int    `mapstructure:"refresh_expiry"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

var AppConfig *Config

// Load 加载配置
func Load() (*Config, error) {
	config := &Config{}

	// 设置默认值
	setDefaults()

	// 从环境变量加载
	loadFromEnv()

	// 尝试从配置文件加载
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using defaults and env vars: %v", err)
	}

	// 映射到结构体
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	AppConfig = config
	return config, nil
}

// setDefaults 设置默认配置
func setDefaults() {
	// Server
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("server.max_body_size", 32*1024*1024) // 32MB

	// Database
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "user")
	viper.SetDefault("database.password", "pass")
	viper.SetDefault("database.database", "polyagent")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_lifetime", 300)

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conn", 2)

	// Python AI
	viper.SetDefault("python_ai.base_url", "http://localhost:8000")
	viper.SetDefault("python_ai.timeout", 60)
	viper.SetDefault("python_ai.max_concurrency", 10)
	viper.SetDefault("python_ai.retry_attempts", 3)

	// Auth
	viper.SetDefault("auth.jwt_secret", "your-jwt-secret")
	viper.SetDefault("auth.token_expiry", 3600)    // 1 hour
	viper.SetDefault("auth.refresh_expiry", 86400) // 24 hours

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.max_age", 30)
}

// loadFromEnv 从环境变量加载配置
func loadFromEnv() {
	// Server
	if host := os.Getenv("SERVER_HOST"); host != "" {
		viper.Set("server.host", host)
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			viper.Set("server.port", p)
		}
	}

	// Database
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		// 解析 DATABASE_URL
		parsePostgresURL(dbURL)
	} else {
		// 单独的环境变量
		if host := os.Getenv("DB_HOST"); host != "" {
			viper.Set("database.host", host)
		}
		if port := os.Getenv("DB_PORT"); port != "" {
			if p, err := strconv.Atoi(port); err == nil {
				viper.Set("database.port", p)
			}
		}
		if user := os.Getenv("DB_USER"); user != "" {
			viper.Set("database.user", user)
		}
		if password := os.Getenv("DB_PASSWORD"); password != "" {
			viper.Set("database.password", password)
		}
		if database := os.Getenv("DB_DATABASE"); database != "" {
			viper.Set("database.database", database)
		}
	}

	// Redis
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		parseRedisURL(redisURL)
	} else {
		if host := os.Getenv("REDIS_HOST"); host != "" {
			viper.Set("redis.host", host)
		}
		if port := os.Getenv("REDIS_PORT"); port != "" {
			if p, err := strconv.Atoi(port); err == nil {
				viper.Set("redis.port", p)
			}
		}
		if password := os.Getenv("REDIS_PASSWORD"); password != "" {
			viper.Set("redis.password", password)
		}
	}

	// Python AI
	if url := os.Getenv("PYTHON_AI_URL"); url != "" {
		viper.Set("python_ai.base_url", url)
	}

	// Auth
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		viper.Set("auth.jwt_secret", secret)
	}

	// Log
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		viper.Set("log.level", level)
	}
}

// parsePostgresURL 解析 PostgreSQL 连接字符串
func parsePostgresURL(url string) {
	// postgres://user:pass@localhost:5432/dbname
	if strings.HasPrefix(url, "postgres://") {
		url = strings.TrimPrefix(url, "postgres://")
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			// 解析用户名密码
			userPass := strings.Split(parts[0], ":")
			if len(userPass) == 2 {
				viper.Set("database.user", userPass[0])
				viper.Set("database.password", userPass[1])
			}

			// 解析主机端口数据库
			hostDB := strings.Split(parts[1], "/")
			if len(hostDB) == 2 {
				hostPort := strings.Split(hostDB[0], ":")
				if len(hostPort) == 2 {
					viper.Set("database.host", hostPort[0])
					if port, err := strconv.Atoi(hostPort[1]); err == nil {
						viper.Set("database.port", port)
					}
				}
				viper.Set("database.database", hostDB[1])
			}
		}
	}
}

// parseRedisURL 解析 Redis 连接字符串
func parseRedisURL(url string) {
	// redis://localhost:6379/0 或 redis://:password@localhost:6379/0
	if strings.HasPrefix(url, "redis://") {
		url = strings.TrimPrefix(url, "redis://")
		
		// 检查是否有密码
		if strings.Contains(url, "@") {
			parts := strings.Split(url, "@")
			if len(parts) == 2 && strings.HasPrefix(parts[0], ":") {
				viper.Set("redis.password", strings.TrimPrefix(parts[0], ":"))
				url = parts[1]
			}
		}

		// 解析主机端口数据库
		parts := strings.Split(url, "/")
		if len(parts) >= 1 {
			hostPort := strings.Split(parts[0], ":")
			if len(hostPort) >= 1 {
				viper.Set("redis.host", hostPort[0])
			}
			if len(hostPort) == 2 {
				if port, err := strconv.Atoi(hostPort[1]); err == nil {
					viper.Set("redis.port", port)
				}
			}
		}
		if len(parts) == 2 {
			if db, err := strconv.Atoi(parts[1]); err == nil {
				viper.Set("redis.database", db)
			}
		}
	}
}