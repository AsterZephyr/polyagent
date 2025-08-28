package models

import (
	"time"

	"github.com/google/uuid"
)

// AgentTask 智能体任务
type AgentTask struct {
	TaskID      string                 `json:"task_id" db:"task_id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	SessionID   string                 `json:"session_id" db:"session_id"`
	AgentType   string                 `json:"agent_type" db:"agent_type"`
	Input       string                 `json:"input" db:"input"`
	Context     map[string]interface{} `json:"context" db:"context"`
	Tools       []string               `json:"tools" db:"tools"`
	Memory      *ConversationMemory    `json:"memory" db:"memory"`
	Status      TaskStatus             `json:"status" db:"status"`
	Priority    int                    `json:"priority" db:"priority"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty" db:"completed_at"`
}

// AgentResponse 智能体响应
type AgentResponse struct {
	TaskID    string                 `json:"task_id"`
	Status    string                 `json:"status"`
	Output    string                 `json:"output"`
	ToolCalls []ToolCall             `json:"tool_calls"`
	Memory    *ConversationMemory    `json:"memory"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ConversationMemory 对话记忆
type ConversationMemory struct {
	SessionID string    `json:"session_id" db:"session_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Messages  []Message `json:"messages" db:"messages"`
	Summary   string    `json:"summary" db:"summary"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Message 消息
type Message struct {
	ID        string                 `json:"id"`
	Role      string                 `json:"role"` // user, assistant, system, tool
	Content   string                 `json:"content"`
	ToolCalls []ToolCall             `json:"tool_calls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	Result     interface{}            `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Agent 智能体定义
type Agent struct {
	ID           string                 `json:"id" db:"id"`
	UserID       string                 `json:"user_id" db:"user_id"`
	Name         string                 `json:"name" db:"name"`
	Type         string                 `json:"type" db:"type"`
	Description  string                 `json:"description" db:"description"`
	Instructions string                 `json:"instructions" db:"instructions"`
	Tools        []string               `json:"tools" db:"tools"`
	Config       map[string]interface{} `json:"config" db:"config"`
	Status       AgentStatus            `json:"status" db:"status"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// Tool 工具定义
type Tool struct {
	Name        string               `json:"name" db:"name"`
	Description string               `json:"description" db:"description"`
	Parameters  map[string]Parameter `json:"parameters" db:"parameters"`
	Handler     string               `json:"handler" db:"handler"`
	Category    string               `json:"category" db:"category"`
	Enabled     bool                 `json:"enabled" db:"enabled"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at" db:"updated_at"`
}

// Parameter 参数定义
type Parameter struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
}

// Document RAG 文档
type Document struct {
	ID          string                 `json:"id" db:"id"`
	UserID      string                 `json:"user_id" db:"user_id"`
	Filename    string                 `json:"filename" db:"filename"`
	Content     string                 `json:"content" db:"content"`
	Chunks      []DocumentChunk        `json:"chunks" db:"chunks"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	Status      DocumentStatus         `json:"status" db:"status"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	IndexedAt   *time.Time             `json:"indexed_at,omitempty" db:"indexed_at"`
}

// DocumentChunk 文档块
type DocumentChunk struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Vector    []float64              `json:"vector,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
	StartPos  int                    `json:"start_pos"`
	EndPos    int                    `json:"end_pos"`
}

// User 用户
type User struct {
	ID        string                 `json:"id" db:"id"`
	Username  string                 `json:"username" db:"username"`
	Email     string                 `json:"email" db:"email"`
	Config    map[string]interface{} `json:"config" db:"config"`
	Status    UserStatus             `json:"status" db:"status"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
	LastLogin *time.Time             `json:"last_login,omitempty" db:"last_login"`
}

// 状态枚举
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
)

type AgentStatus int

const (
	AgentStatusActive AgentStatus = iota
	AgentStatusInactive
	AgentStatusArchived
)

type DocumentStatus int

const (
	DocumentStatusUploaded DocumentStatus = iota
	DocumentStatusProcessing
	DocumentStatusIndexed
	DocumentStatusFailed
)

type UserStatus int

const (
	UserStatusActive UserStatus = iota
	UserStatusInactive
	UserStatusSuspended
)

// NewTaskID 生成新的任务ID
func NewTaskID() string {
	return uuid.New().String()
}

// NewAgentID 生成新的智能体ID
func NewAgentID() string {
	return uuid.New().String()
}

// NewDocumentID 生成新的文档ID
func NewDocumentID() string {
	return uuid.New().String()
}

// NewUserID 生成新的用户ID
func NewUserID() string {
	return uuid.New().String()
}