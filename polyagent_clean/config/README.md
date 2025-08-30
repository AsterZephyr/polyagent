# Config Module - Configuration Management

## Overview

Config模块负责PolyAgent系统的配置管理，遵循Unix哲学中"配置即环境"的原则。支持环境变量、YAML配置文件和默认值的多层配置体系，确保系统的灵活性和可维护性。

## Architecture

```mermaid
graph TB
    subgraph "Configuration Sources (Priority Order)"
        Env[🔧 Environment Variables] --> |Highest Priority|Layer1[Configuration Layer 1]
        YAML[📄 YAML Files] --> |Medium Priority|Layer2[Configuration Layer 2]
        Default[⚙️ Default Values] --> |Lowest Priority|Layer3[Configuration Layer 3]
    end
    
    subgraph "Configuration Loading"
        Layer1 --> Merge[Configuration Merger]
        Layer2 --> Merge
        Layer3 --> Merge
        Merge --> Validate[Configuration Validator]
        Validate --> Final[Final Configuration]
    end
    
    subgraph "Configuration Usage"
        Final --> Core[Core Module]
        Final --> Gateway[Gateway Module]
        Final --> Tools[Tools Module]
        Final --> Scripts[Scripts Module]
    end
```

## Configuration Flow

```mermaid
sequenceDiagram
    participant App as Application
    participant Loader as Config Loader
    participant Env as Environment
    participant File as YAML File
    participant Val as Validator
    
    App->>Loader: Load Configuration
    Loader->>Env: Read Environment Variables
    Env->>Loader: Environment Config
    
    Loader->>File: Read YAML Files
    File->>Loader: File Config
    
    Loader->>Loader: Merge Configurations
    Loader->>Val: Validate Final Config
    Val->>Loader: Validation Result
    
    alt Validation Success
        Loader->>App: Valid Configuration
    else Validation Failed
        Loader->>App: Configuration Error
    end
```

## File Structure

```
config/
├── README.md           # This documentation
├── env.example         # Environment variables template
├── models.yaml         # AI models configuration
├── logging.yaml        # Logging configuration
├── security.yaml       # Security settings
└── environments/       # Environment-specific configs
    ├── development.yaml
    ├── staging.yaml
    └── production.yaml
```

## Environment Variables

### Core API Keys

```bash
# AI Provider API Keys (至少需要一个)
OPENAI_API_KEY=sk-your-openai-key-here
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key-here
OPENROUTER_API_KEY=sk-or-your-openrouter-key-here
GLM_API_KEY=your-glm-key-here

# Proxy Settings (Optional)
OPENAI_BASE_URL=https://api.proxy.com/v1/
ANTHROPIC_BASE_URL=https://api.proxy.com/v1/
```

### Behavior Configuration

```bash
# System Behavior
POLYAGENT_VERBOSE=true              # Enable verbose output
POLYAGENT_TOOLS=true                # Enable tool calling
POLYAGENT_MEDICAL_SAFETY=true       # Enable medical safety checks
POLYAGENT_LOG_LEVEL=INFO            # Logging level: DEBUG, INFO, WARN, ERROR

# Document Configuration
POLYAGENT_DOCS=/path/to/docs1,/path/to/docs2  # Document search paths
POLYAGENT_DOC_MAX_SIZE=10MB         # Maximum document size
POLYAGENT_DOC_FORMATS=txt,md,pdf    # Supported formats

# Performance Settings
POLYAGENT_TIMEOUT=30                # Request timeout in seconds
POLYAGENT_MAX_RETRIES=3             # Maximum retry attempts
POLYAGENT_CACHE_TTL=3600            # Cache TTL in seconds
```

### Gateway Configuration

```bash
# Gateway Settings
GATEWAY_PORT=8080                   # HTTP server port
GATEWAY_HOST=0.0.0.0               # Bind address
GATEWAY_TIMEOUT=30s                # Request timeout
GATEWAY_MAX_REQUEST_SIZE=10MB      # Maximum request size

# Authentication
AUTH_ENABLED=true                  # Enable authentication
AUTH_SECRET_KEY=your-secret-key    # JWT secret key
AUTH_TOKEN_TTL=24h                 # Token validity period

# Rate Limiting
RATE_LIMIT_ENABLED=true           # Enable rate limiting
RATE_LIMIT_RPM=60                 # Requests per minute
RATE_LIMIT_BURST=10               # Burst capacity
```

## YAML Configuration Files

### models.yaml - AI Models Configuration

```yaml
# AI Models Configuration
models:
  # Production models (high quality, higher cost)
  production:
    reasoning: "claude-3-5-sonnet-20241022"     # Best for reasoning tasks
    coding: "qwen/qwen-2.5-coder-32b-instruct" # Best for coding
    multimodal: "gpt-4o"                       # Best for images/vision
    general: "claude-3-5-sonnet-20241022"      # Default choice
    
  # Free models (cost-effective)
  free:
    coding: "qwen/qwen-2.5-coder-32b-instruct"  # OpenRouter free
    general: "microsoft/wizardlm-2-8x22b"       # OpenRouter free
    chinese: "glm-4-plus"                       # GLM free tier (2M tokens)

# Model routing rules
routing:
  keywords:
    code: ["code", "python", "javascript", "programming", "function", "class", "algorithm"]
    reasoning: ["analyze", "reason", "think", "complex", "logic", "problem"]
    image: ["image", "photo", "picture", "visual", "diagram", "chart"]
    chinese: ["中文", "汉语", "中国", "翻译"]

# Cost limits and preferences
cost_control:
  max_cost_per_request: 0.10      # Maximum cost per request ($)
  prefer_free: false              # Prefer free models when available
  cost_tracking: true             # Track and report costs

# Provider endpoints
providers:
  openai:
    base_url: "https://api.openai.com/v1"
    timeout: 30
    max_tokens: 4000
    
  anthropic:
    base_url: "https://api.anthropic.com"
    version: "2023-06-01"
    timeout: 30
    max_tokens: 4000
    
  openrouter:
    base_url: "https://openrouter.ai/api/v1"
    timeout: 30
    max_tokens: 4000
    
  glm:
    base_url: "https://open.bigmodel.cn/api/paas/v4"
    timeout: 30
    max_tokens: 4000
```

### logging.yaml - Logging Configuration

```yaml
# Logging Configuration
logging:
  # Global settings
  level: "INFO"                    # DEBUG, INFO, WARN, ERROR
  format: "structured"             # structured, simple
  timestamp: true                  # Include timestamps
  color: true                      # Colorized output (for terminal)
  
  # Output destinations
  outputs:
    console:
      enabled: true
      level: "INFO"
      format: "simple"
      
    file:
      enabled: true
      level: "DEBUG"
      format: "structured"
      path: "./logs/polyagent.log"
      max_size: "100MB"
      max_files: 7
      compress: true
      
    syslog:
      enabled: false
      facility: "local0"
      tag: "polyagent"

  # Component-specific logging
  components:
    core:
      level: "INFO"
      include_stack_trace: false
      
    gateway:
      level: "INFO"
      access_log: true
      
    ai_calls:
      level: "DEBUG"
      log_requests: true
      log_responses: false          # Sensitive data
      log_tokens: true
      
    tools:
      level: "INFO"
      log_executions: true
      
    security:
      level: "WARN"
      log_auth_attempts: true
      log_rate_limits: true

# Sensitive data filtering
filters:
  api_keys: true                   # Mask API keys in logs
  user_data: true                  # Mask user content
  system_prompts: false            # System prompts are OK to log
```

### security.yaml - Security Configuration

```yaml
# Security Configuration
security:
  # Authentication settings
  authentication:
    enabled: true
    method: "bearer"               # bearer, basic, apikey
    token_validation: "jwt"        # jwt, database, redis
    
    jwt:
      algorithm: "HS256"
      secret_key: "${AUTH_SECRET_KEY}"  # From environment
      expiration: "24h"
      issuer: "polyagent"
      
  # Authorization settings  
  authorization:
    enabled: true
    default_role: "user"
    admin_roles: ["admin", "superuser"]
    
    permissions:
      chat: ["user", "admin"]
      health: ["user", "admin"]
      models: ["user", "admin"]
      admin: ["admin"]

  # Rate limiting
  rate_limiting:
    enabled: true
    algorithm: "token_bucket"
    
    rules:
      default:
        requests_per_minute: 60
        burst: 10
        
      premium:
        requests_per_minute: 300
        burst: 50
        
      admin:
        requests_per_minute: 1000
        burst: 100

  # Medical safety
  medical_safety:
    enabled: true
    strict_mode: true              # Block vs warn
    patterns_file: "medical_patterns.yaml"
    disclaimer_required: true
    
    blocked_patterns:
      - "诊断为|确诊为"             # Diagnosis claims
      - "建议.*服用.*药"            # Medication recommendations
      - "不需要看医生"             # Discouraging medical care
      
  # Content filtering
  content_filtering:
    enabled: true
    
    blocked_categories:
      - "violence"
      - "hate_speech" 
      - "illegal_activities"
      
    response_filtering: true
    request_filtering: true

  # API security
  api_security:
    cors:
      enabled: true
      origins: ["*"]               # Restrict in production
      methods: ["GET", "POST", "OPTIONS"]
      headers: ["Content-Type", "Authorization"]
      
    headers:
      x_frame_options: "DENY"
      x_content_type_options: "nosniff"
      x_xss_protection: "1; mode=block"
      strict_transport_security: "max-age=31536000"
      
    request_limits:
      max_request_size: "10MB"
      max_header_size: "8KB"
      timeout: "30s"
```

## Environment-Specific Configurations

### development.yaml

```yaml
# Development Environment Configuration
environment: "development"

core:
  verbose: true
  debug_mode: true
  hot_reload: true

gateway:
  port: 8080
  cors_origins: ["http://localhost:3000", "http://localhost:8080"]
  
logging:
  level: "DEBUG"
  console_output: true
  file_output: false

security:
  authentication:
    enabled: false                # Disabled for development
  rate_limiting:
    enabled: false               # Disabled for development
    
performance:
  cache_enabled: false           # Disable cache for development
  timeout: 60                    # Longer timeout for debugging
```

### production.yaml

```yaml
# Production Environment Configuration
environment: "production"

core:
  verbose: false
  debug_mode: false
  performance_monitoring: true

gateway:
  port: 8080
  cors_origins: ["https://yourdomain.com"]
  
logging:
  level: "INFO"
  console_output: false
  file_output: true
  structured_format: true

security:
  authentication:
    enabled: true
  rate_limiting:
    enabled: true
    strict_mode: true
    
  content_filtering:
    enabled: true
    strict_mode: true
    
performance:
  cache_enabled: true
  timeout: 30
  connection_pooling: true

monitoring:
  metrics_enabled: true
  health_checks: true
  alerting: true
```

## Configuration Loading Process

```mermaid
flowchart TD
    Start[Application Start] --> LoadEnv[Load Environment Variables]
    LoadEnv --> DetectEnv{Detect Environment}
    
    DetectEnv -->|POLYAGENT_ENV=development| DevConfig[Load development.yaml]
    DetectEnv -->|POLYAGENT_ENV=staging| StagingConfig[Load staging.yaml]
    DetectEnv -->|POLYAGENT_ENV=production| ProdConfig[Load production.yaml]
    DetectEnv -->|Not Set| DefaultConfig[Load Default Configuration]
    
    DevConfig --> LoadModels[Load models.yaml]
    StagingConfig --> LoadModels
    ProdConfig --> LoadModels
    DefaultConfig --> LoadModels
    
    LoadModels --> LoadLogging[Load logging.yaml]
    LoadLogging --> LoadSecurity[Load security.yaml]
    LoadSecurity --> MergeConfig[Merge All Configurations]
    
    MergeConfig --> ValidateConfig{Validate Configuration}
    ValidateConfig -->|Valid| FinalConfig[Final Configuration Ready]
    ValidateConfig -->|Invalid| ConfigError[Configuration Error]
    
    FinalConfig --> StartServices[Start Services]
    ConfigError --> Exit[Exit Application]
```

## Configuration Validation

### Required Fields Validation

```yaml
# Validation schema
validation:
  required_fields:
    - "models.production.general"
    - "logging.level"
    - "security.authentication.enabled"
    
  api_keys:
    minimum_required: 1           # At least one API key
    valid_formats:
      openai: "^sk-[a-zA-Z0-9]{48}$"
      anthropic: "^sk-ant-[a-zA-Z0-9-]{95}$"
      
  constraints:
    port_range: [1024, 65535]
    timeout_range: [1, 300]
    log_level: ["DEBUG", "INFO", "WARN", "ERROR"]
```

### Configuration Health Check

```mermaid
sequenceDiagram
    participant App as Application
    participant Validator as Config Validator
    participant API as API Services
    participant File as File System
    
    App->>Validator: Validate Configuration
    
    Validator->>API: Test API Keys
    loop For each API key
        API->>API: Make test request
        API->>Validator: Key validation result
    end
    
    Validator->>File: Check file permissions
    File->>Validator: Permission status
    
    Validator->>Validator: Check required fields
    Validator->>Validator: Validate constraints
    
    Validator->>App: Validation Report
```

## Usage Examples

### Loading Configuration in Python

```python
import os
import yaml
from typing import Dict, Any

class ConfigLoader:
    def __init__(self):
        self.config = {}
        
    def load(self) -> Dict[str, Any]:
        # 1. Load default values
        self._load_defaults()
        
        # 2. Load YAML files
        self._load_yaml_files()
        
        # 3. Override with environment variables
        self._load_environment()
        
        # 4. Validate configuration
        self._validate()
        
        return self.config
        
    def _load_environment(self):
        # API Keys
        api_keys = {}
        for key in ['OPENAI_API_KEY', 'ANTHROPIC_API_KEY', 'OPENROUTER_API_KEY', 'GLM_API_KEY']:
            if os.getenv(key):
                api_keys[key] = os.getenv(key)
        
        if api_keys:
            self.config['api_keys'] = api_keys
            
        # Behavior settings
        if os.getenv('POLYAGENT_VERBOSE'):
            self.config['verbose'] = os.getenv('POLYAGENT_VERBOSE').lower() == 'true'
            
        if os.getenv('POLYAGENT_LOG_LEVEL'):
            self.config['logging']['level'] = os.getenv('POLYAGENT_LOG_LEVEL')
```

### Loading Configuration in Go

```go
package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Gateway struct {
        Port    int    `yaml:"port" env:"GATEWAY_PORT"`
        Host    string `yaml:"host" env:"GATEWAY_HOST"`
        Timeout string `yaml:"timeout" env:"GATEWAY_TIMEOUT"`
    } `yaml:"gateway"`
    
    Auth struct {
        Enabled   bool   `yaml:"enabled" env:"AUTH_ENABLED"`
        SecretKey string `yaml:"secret_key" env:"AUTH_SECRET_KEY"`
    } `yaml:"auth"`
}

func Load() (*Config, error) {
    config := &Config{}
    
    // Load YAML file
    if err := loadYAML("config.yaml", config); err != nil {
        return nil, err
    }
    
    // Override with environment variables
    loadEnvOverrides(config)
    
    return config, nil
}
```

## Best Practices

### 1. Environment Variable Naming
```bash
# Use consistent prefixing
POLYAGENT_VERBOSE=true
POLYAGENT_LOG_LEVEL=INFO
POLYAGENT_TIMEOUT=30

# Not: VERBOSE=true, LOG_LEVEL=INFO
```

### 2. Sensitive Data Handling
```yaml
# ❌ Don't put secrets in YAML files
api_keys:
  openai: "sk-actual-key-here"

# ✅ Reference environment variables
api_keys:
  openai: "${OPENAI_API_KEY}"
```

### 3. Environment-Specific Overrides
```yaml
# Base configuration
timeout: 30

---
# Development override
development:
  timeout: 60    # Longer for debugging

# Production override  
production:
  timeout: 15    # Stricter for production
```

### 4. Configuration Documentation
```yaml
# Document each setting
models:
  production:
    # Primary model for reasoning tasks
    # Higher cost but better quality
    reasoning: "claude-3-5-sonnet-20241022"
```

## Troubleshooting

### Common Configuration Issues

```mermaid
flowchart TD
    Issue[Configuration Issue] --> Type{Issue Type}
    
    Type -->|Missing API Key| APIKeyFix[Add API key to environment]
    Type -->|Invalid YAML| YAMLFix[Check YAML syntax]
    Type -->|Permission Error| PermFix[Check file permissions]
    Type -->|Port Conflict| PortFix[Change port number]
    
    APIKeyFix --> Test[Test Configuration]
    YAMLFix --> Test
    PermFix --> Test
    PortFix --> Test
    
    Test --> Success[Configuration Loaded Successfully]
```

### Debug Commands

```bash
# Check environment variables
env | grep POLYAGENT

# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('models.yaml'))"

# Test configuration loading
python -c "from core.main import load_config; print(load_config())"

# Check file permissions
ls -la config/
```

---

*Config模块体现了Unix"约定优于配置"的哲学，提供了灵活而可靠的配置管理机制，支持从开发到生产的全生命周期配置需求。*