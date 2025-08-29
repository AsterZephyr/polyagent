package models

import (
	"time"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"github.com/lib/pq"
)

// JSONMap 自定义 JSON 类型
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	
	return json.Unmarshal(bytes, &j)
}

// Base 基础模型
type Base struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// User 用户模型
type User struct {
	Base
	Username     string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string    `json:"-" gorm:"size:255;not null"`
	DisplayName  string    `json:"display_name" gorm:"size:100"`
	Avatar       string    `json:"avatar" gorm:"size:255"`
	Role         string    `json:"role" gorm:"size:20;default:'user'"`
	Status       string    `json:"status" gorm:"size:20;default:'active'"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	Settings     JSONMap   `json:"settings" gorm:"type:jsonb"`
	
	// 关联
	Sessions     []Session     `json:"sessions,omitempty" gorm:"foreignKey:UserID"`
	Messages     []Message     `json:"messages,omitempty" gorm:"foreignKey:UserID"`
	Documents    []Document    `json:"documents,omitempty" gorm:"foreignKey:UserID"`
	Agents       []Agent       `json:"agents,omitempty" gorm:"foreignKey:UserID"`
	Tasks        []Task        `json:"tasks,omitempty" gorm:"foreignKey:UserID"`
}

// Agent 智能体模型
type Agent struct {
	Base
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	Name        string  `json:"name" gorm:"size:100;not null"`
	Description string  `json:"description" gorm:"type:text"`
	Type        string  `json:"type" gorm:"size:50;not null"`
	Mode        string  `json:"mode" gorm:"size:50;default:'auto'"`
	Status      string  `json:"status" gorm:"size:20;default:'active'"`
	Config      JSONMap `json:"config" gorm:"type:jsonb"`
	
	// 性能指标
	RequestCount    int64   `json:"request_count" gorm:"default:0"`
	SuccessCount    int64   `json:"success_count" gorm:"default:0"`
	ErrorCount      int64   `json:"error_count" gorm:"default:0"`
	AvgResponseTime float64 `json:"avg_response_time" gorm:"default:0"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	
	// 关联
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Sessions []Session `json:"sessions,omitempty" gorm:"foreignKey:AgentID"`
	Messages []Message `json:"messages,omitempty" gorm:"foreignKey:AgentID"`
	Tasks    []Task    `json:"tasks,omitempty" gorm:"foreignKey:AgentID"`
}

// Session 会话模型
type Session struct {
	Base
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	AgentID     *uint      `json:"agent_id" gorm:"index"`
	Title       string     `json:"title" gorm:"size:200"`
	Status      string     `json:"status" gorm:"size:20;default:'active'"`
	Context     JSONMap    `json:"context" gorm:"type:jsonb"`
	LastMessage *time.Time `json:"last_message"`
	MessageCount int       `json:"message_count" gorm:"default:0"`
	
	// 关联
	User     User      `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Agent    *Agent    `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Messages []Message `json:"messages,omitempty" gorm:"foreignKey:SessionID"`
}

// Message 消息模型
type Message struct {
	Base
	SessionID   uint    `json:"session_id" gorm:"not null;index"`
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	AgentID     *uint   `json:"agent_id" gorm:"index"`
	Role        string  `json:"role" gorm:"size:20;not null"`
	Content     string  `json:"content" gorm:"type:text;not null"`
	ContentType string  `json:"content_type" gorm:"size:50;default:'text'"`
	
	// AI 相关信息
	Model           string  `json:"model" gorm:"size:100"`
	TokensUsed      int     `json:"tokens_used" gorm:"default:0"`
	ProcessingTime  float64 `json:"processing_time" gorm:"default:0"`
	ConfidenceScore float64 `json:"confidence_score" gorm:"default:0"`
	
	// 扩展信息
	Metadata JSONMap `json:"metadata" gorm:"type:jsonb"`
	
	// 关联
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Agent   *Agent   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
}

// Document 文档模型
type Document struct {
	Base
	UserID       uint    `json:"user_id" gorm:"not null;index"`
	Title        string  `json:"title" gorm:"size:200;not null"`
	Description  string  `json:"description" gorm:"type:text"`
	Type         string  `json:"type" gorm:"size:50;not null"`
	SourceURL    string  `json:"source_url" gorm:"size:500"`
	FilePath     string  `json:"file_path" gorm:"size:500"`
	FileSize     int64   `json:"file_size" gorm:"default:0"`
	ContentHash  string  `json:"content_hash" gorm:"size:64;index"`
	
	// 处理状态
	Status           string    `json:"status" gorm:"size:20;default:'pending'"`
	ProcessingResult JSONMap   `json:"processing_result" gorm:"type:jsonb"`
	ProcessedAt      *time.Time `json:"processed_at"`
	
	// 索引信息
	ChunksCount     int `json:"chunks_count" gorm:"default:0"`
	EmbeddingsCount int `json:"embeddings_count" gorm:"default:0"`
	IndexedAt       *time.Time `json:"indexed_at"`
	
	// 使用统计
	AccessCount int        `json:"access_count" gorm:"default:0"`
	LastAccess  *time.Time `json:"last_access"`
	
	// 扩展信息
	Tags     pq.StringArray `json:"tags" gorm:"type:text[]"`
	Metadata JSONMap        `json:"metadata" gorm:"type:jsonb"`
	
	// 关联
	User   User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Chunks []DocumentChunk `json:"chunks,omitempty" gorm:"foreignKey:DocumentID"`
}

// DocumentChunk 文档块模型
type DocumentChunk struct {
	Base
	DocumentID uint    `json:"document_id" gorm:"not null;index"`
	ChunkIndex int     `json:"chunk_index" gorm:"not null"`
	Content    string  `json:"content" gorm:"type:text;not null"`
	ContentHash string `json:"content_hash" gorm:"size:64;index"`
	StartOffset int    `json:"start_offset" gorm:"default:0"`
	EndOffset   int    `json:"end_offset" gorm:"default:0"`
	
	// 向量嵌入
	EmbeddingModel string    `json:"embedding_model" gorm:"size:100"`
	Embedding      []float32 `json:"-" gorm:"type:vector(1536)"` // 使用 pgvector
	EmbeddingHash  string    `json:"embedding_hash" gorm:"size:64;index"`
	
	// 扩展信息
	Metadata JSONMap `json:"metadata" gorm:"type:jsonb"`
	
	// 关联
	Document Document `json:"document,omitempty" gorm:"foreignKey:DocumentID"`
}

// Task 任务模型
type Task struct {
	Base
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	AgentID     *uint   `json:"agent_id" gorm:"index"`
	SessionID   *uint   `json:"session_id" gorm:"index"`
	Type        string  `json:"type" gorm:"size:50;not null"`
	Status      string  `json:"status" gorm:"size:20;default:'pending'"`
	Priority    int     `json:"priority" gorm:"default:5"`
	
	// 任务内容
	Title       string  `json:"title" gorm:"size:200;not null"`
	Description string  `json:"description" gorm:"type:text"`
	Input       JSONMap `json:"input" gorm:"type:jsonb"`
	Output      JSONMap `json:"output" gorm:"type:jsonb"`
	
	// 执行信息
	StartedAt     *time.Time `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at"`
	ExecutionTime float64    `json:"execution_time" gorm:"default:0"`
	RetryCount    int        `json:"retry_count" gorm:"default:0"`
	MaxRetries    int        `json:"max_retries" gorm:"default:3"`
	
	// 错误信息
	ErrorMessage string  `json:"error_message" gorm:"type:text"`
	ErrorCode    string  `json:"error_code" gorm:"size:50"`
	
	// 调度信息
	ScheduledFor *time.Time `json:"scheduled_for"`
	Timeout      int        `json:"timeout" gorm:"default:300"` // 超时时间（秒）
	
	// 扩展信息
	Tags     pq.StringArray `json:"tags" gorm:"type:text[]"`
	Metadata JSONMap        `json:"metadata" gorm:"type:jsonb"`
	
	// 关联
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Agent   *Agent   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// Memory 记忆模型
type Memory struct {
	Base
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	SessionID   *uint   `json:"session_id" gorm:"index"`
	Type        string  `json:"type" gorm:"size:50;not null;index"`
	Content     string  `json:"content" gorm:"type:text;not null"`
	ContentHash string  `json:"content_hash" gorm:"size:64;index"`
	
	// 重要性和相关性
	Importance     float64 `json:"importance" gorm:"default:0.5"`
	AccessCount    int     `json:"access_count" gorm:"default:0"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	
	// 向量嵌入
	EmbeddingModel string    `json:"embedding_model" gorm:"size:100"`
	Embedding      []float32 `json:"-" gorm:"type:vector(1536)"`
	EmbeddingHash  string    `json:"embedding_hash" gorm:"size:64;index"`
	
	// 关联记忆
	RelatedMemories pq.Int64Array `json:"related_memories" gorm:"type:bigint[]"`
	
	// 扩展信息
	Tags     pq.StringArray `json:"tags" gorm:"type:text[]"`
	Metadata JSONMap        `json:"metadata" gorm:"type:jsonb"`
	
	// 关联
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
}

// Tool 工具模型
type Tool struct {
	Base
	Name        string  `json:"name" gorm:"uniqueIndex;size:100;not null"`
	DisplayName string  `json:"display_name" gorm:"size:100;not null"`
	Description string  `json:"description" gorm:"type:text"`
	Category    string  `json:"category" gorm:"size:50;not null"`
	Version     string  `json:"version" gorm:"size:20;default:'1.0.0'"`
	Status      string  `json:"status" gorm:"size:20;default:'active'"`
	
	// 工具配置
	Config       JSONMap `json:"config" gorm:"type:jsonb"`
	Schema       JSONMap `json:"schema" gorm:"type:jsonb"`
	Permissions  JSONMap `json:"permissions" gorm:"type:jsonb"`
	
	// 使用统计
	CallCount   int64      `json:"call_count" gorm:"default:0"`
	ErrorCount  int64      `json:"error_count" gorm:"default:0"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	AvgExecTime float64    `json:"avg_exec_time" gorm:"default:0"`
	
	// 关联
	ToolCalls []ToolCall `json:"tool_calls,omitempty" gorm:"foreignKey:ToolID"`
}

// ToolCall 工具调用记录
type ToolCall struct {
	Base
	ToolID      uint    `json:"tool_id" gorm:"not null;index"`
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	AgentID     *uint   `json:"agent_id" gorm:"index"`
	SessionID   *uint   `json:"session_id" gorm:"index"`
	MessageID   *uint   `json:"message_id" gorm:"index"`
	
	// 调用信息
	Input          JSONMap `json:"input" gorm:"type:jsonb"`
	Output         JSONMap `json:"output" gorm:"type:jsonb"`
	Status         string  `json:"status" gorm:"size:20;not null"`
	ExecutionTime  float64 `json:"execution_time" gorm:"default:0"`
	
	// 错误信息
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	ErrorCode    string `json:"error_code" gorm:"size:50"`
	
	// 关联
	Tool    Tool     `json:"tool,omitempty" gorm:"foreignKey:ToolID"`
	User    User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Agent   *Agent   `json:"agent,omitempty" gorm:"foreignKey:AgentID"`
	Session *Session `json:"session,omitempty" gorm:"foreignKey:SessionID"`
	Message *Message `json:"message,omitempty" gorm:"foreignKey:MessageID"`
}

// SystemLog 系统日志模型
type SystemLog struct {
	Base
	Level     string  `json:"level" gorm:"size:20;not null;index"`
	Service   string  `json:"service" gorm:"size:50;not null;index"`
	Component string  `json:"component" gorm:"size:100;index"`
	Message   string  `json:"message" gorm:"type:text;not null"`
	
	// 关联信息
	UserID    *uint `json:"user_id" gorm:"index"`
	SessionID *uint `json:"session_id" gorm:"index"`
	AgentID   *uint `json:"agent_id" gorm:"index"`
	TaskID    *uint `json:"task_id" gorm:"index"`
	
	// 扩展信息
	StackTrace string  `json:"stack_trace" gorm:"type:text"`
	Metadata   JSONMap `json:"metadata" gorm:"type:jsonb"`
	
	// IP和用户代理
	ClientIP  string `json:"client_ip" gorm:"size:45"`
	UserAgent string `json:"user_agent" gorm:"size:500"`
}

// ApiKey API密钥模型
type ApiKey struct {
	Base
	UserID      uint       `json:"user_id" gorm:"not null;index"`
	Name        string     `json:"name" gorm:"size:100;not null"`
	KeyHash     string     `json:"-" gorm:"size:255;not null;uniqueIndex"`
	KeyPrefix   string     `json:"key_prefix" gorm:"size:10;not null"`
	Status      string     `json:"status" gorm:"size:20;default:'active'"`
	Permissions JSONMap    `json:"permissions" gorm:"type:jsonb"`
	
	// 使用限制
	RateLimit   int        `json:"rate_limit" gorm:"default:1000"`
	UsageCount  int64      `json:"usage_count" gorm:"default:0"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	
	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// Config 配置模型
type Config struct {
	Base
	Key         string  `json:"key" gorm:"uniqueIndex;size:100;not null"`
	Value       string  `json:"value" gorm:"type:text;not null"`
	Type        string  `json:"type" gorm:"size:20;default:'string'"`
	Category    string  `json:"category" gorm:"size:50;not null;index"`
	Description string  `json:"description" gorm:"type:text"`
	IsSecret    bool    `json:"is_secret" gorm:"default:false"`
	IsReadonly  bool    `json:"is_readonly" gorm:"default:false"`
	
	// 验证规则
	ValidationRule string `json:"validation_rule" gorm:"type:text"`
	DefaultValue   string `json:"default_value" gorm:"type:text"`
	
	// 扩展信息
	Metadata JSONMap `json:"metadata" gorm:"type:jsonb"`
}

// Webhook 模型
type Webhook struct {
	Base
	UserID      uint    `json:"user_id" gorm:"not null;index"`
	Name        string  `json:"name" gorm:"size:100;not null"`
	URL         string  `json:"url" gorm:"size:500;not null"`
	Events      pq.StringArray `json:"events" gorm:"type:text[];not null"`
	Secret      string  `json:"-" gorm:"size:255"`
	Status      string  `json:"status" gorm:"size:20;default:'active'"`
	
	// 配置
	Headers     JSONMap `json:"headers" gorm:"type:jsonb"`
	Timeout     int     `json:"timeout" gorm:"default:30"`
	RetryCount  int     `json:"retry_count" gorm:"default:3"`
	
	// 统计
	CallCount    int64      `json:"call_count" gorm:"default:0"`
	ErrorCount   int64      `json:"error_count" gorm:"default:0"`
	LastCalledAt *time.Time `json:"last_called_at"`
	
	// 关联
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}