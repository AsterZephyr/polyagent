package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/polyagent/go-services/internal/config"
	"github.com/polyagent/go-services/internal/models"
)

type PostgresStorage struct {
	db *sql.DB
}

// NewPostgresStorage 创建 PostgreSQL 存储实例
func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.MaxLifetime) * time.Second)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &PostgresStorage{db: db}

	// 初始化数据库表
	if err := storage.InitTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

// InitTables 初始化数据库表
func (ps *PostgresStorage) InitTables() error {
	schema := `
	-- 用户表
	CREATE TABLE IF NOT EXISTS users (
		id VARCHAR(36) PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		config JSONB DEFAULT '{}',
		status INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		last_login TIMESTAMP WITH TIME ZONE
	);

	-- 智能体表
	CREATE TABLE IF NOT EXISTS agents (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(100) NOT NULL,
		description TEXT DEFAULT '',
		instructions TEXT DEFAULT '',
		tools JSONB DEFAULT '[]',
		config JSONB DEFAULT '{}',
		status INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- 任务表
	CREATE TABLE IF NOT EXISTS tasks (
		task_id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		session_id VARCHAR(255) NOT NULL,
		agent_type VARCHAR(100) NOT NULL,
		input TEXT NOT NULL,
		context JSONB DEFAULT '{}',
		tools JSONB DEFAULT '[]',
		memory JSONB,
		status INTEGER DEFAULT 0,
		priority INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		completed_at TIMESTAMP WITH TIME ZONE
	);

	-- 对话记忆表
	CREATE TABLE IF NOT EXISTS conversation_memory (
		session_id VARCHAR(255) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		messages JSONB DEFAULT '[]',
		summary TEXT DEFAULT '',
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- 工具表
	CREATE TABLE IF NOT EXISTS tools (
		name VARCHAR(255) PRIMARY KEY,
		description TEXT NOT NULL,
		parameters JSONB DEFAULT '{}',
		handler VARCHAR(255) NOT NULL,
		category VARCHAR(100) DEFAULT 'general',
		enabled BOOLEAN DEFAULT true,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	-- 文档表
	CREATE TABLE IF NOT EXISTS documents (
		id VARCHAR(36) PRIMARY KEY,
		user_id VARCHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		filename VARCHAR(255) NOT NULL,
		content TEXT,
		chunks JSONB DEFAULT '[]',
		metadata JSONB DEFAULT '{}',
		status INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		indexed_at TIMESTAMP WITH TIME ZONE
	);

	-- 创建索引
	CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
	CREATE INDEX IF NOT EXISTS idx_agents_user_id ON agents(user_id);
	CREATE INDEX IF NOT EXISTS idx_agents_type ON agents(type);
	CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);
	CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status);
	CREATE INDEX IF NOT EXISTS idx_conversation_memory_user_id ON conversation_memory(user_id);

	-- 更新时间触发器
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = NOW();
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	DROP TRIGGER IF EXISTS update_users_updated_at ON users;
	CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_agents_updated_at ON agents;
	CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_tasks_updated_at ON tasks;
	CREATE TRIGGER update_tasks_updated_at BEFORE UPDATE ON tasks
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_tools_updated_at ON tools;
	CREATE TRIGGER update_tools_updated_at BEFORE UPDATE ON tools
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

	DROP TRIGGER IF EXISTS update_documents_updated_at ON documents;
	CREATE TRIGGER update_documents_updated_at BEFORE UPDATE ON documents
		FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
	`

	_, err := ps.db.Exec(schema)
	return err
}

// Task相关方法

// CreateTask 创建任务
func (ps *PostgresStorage) CreateTask(task *models.AgentTask) error {
	query := `
		INSERT INTO tasks (task_id, user_id, session_id, agent_type, input, context, tools, memory, status, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := ps.db.Exec(query,
		task.TaskID, task.UserID, task.SessionID, task.AgentType,
		task.Input, task.Context, task.Tools, task.Memory,
		task.Status, task.Priority,
	)
	return err
}

// GetTask 获取任务
func (ps *PostgresStorage) GetTask(taskID string) (*models.AgentTask, error) {
	task := &models.AgentTask{}
	query := `
		SELECT task_id, user_id, session_id, agent_type, input, context, tools, memory,
			   status, priority, created_at, updated_at, completed_at
		FROM tasks WHERE task_id = $1
	`
	err := ps.db.QueryRow(query, taskID).Scan(
		&task.TaskID, &task.UserID, &task.SessionID, &task.AgentType,
		&task.Input, &task.Context, &task.Tools, &task.Memory,
		&task.Status, &task.Priority, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateTaskStatus 更新任务状态
func (ps *PostgresStorage) UpdateTaskStatus(taskID string, status models.TaskStatus) error {
	query := `UPDATE tasks SET status = $1 WHERE task_id = $2`
	_, err := ps.db.Exec(query, status, taskID)
	return err
}

// CompleteTask 完成任务
func (ps *PostgresStorage) CompleteTask(taskID string) error {
	now := time.Now()
	query := `UPDATE tasks SET status = $1, completed_at = $2 WHERE task_id = $3`
	_, err := ps.db.Exec(query, models.TaskStatusCompleted, now, taskID)
	return err
}

// GetPendingTasks 获取待处理任务
func (ps *PostgresStorage) GetPendingTasks(limit int) ([]*models.AgentTask, error) {
	query := `
		SELECT task_id, user_id, session_id, agent_type, input, context, tools, memory,
			   status, priority, created_at, updated_at, completed_at
		FROM tasks 
		WHERE status = $1 
		ORDER BY priority DESC, created_at ASC 
		LIMIT $2
	`
	
	rows, err := ps.db.Query(query, models.TaskStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*models.AgentTask
	for rows.Next() {
		task := &models.AgentTask{}
		err := rows.Scan(
			&task.TaskID, &task.UserID, &task.SessionID, &task.AgentType,
			&task.Input, &task.Context, &task.Tools, &task.Memory,
			&task.Status, &task.Priority, &task.CreatedAt, &task.UpdatedAt, &task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

// Agent相关方法

// CreateAgent 创建智能体
func (ps *PostgresStorage) CreateAgent(agent *models.Agent) error {
	query := `
		INSERT INTO agents (id, user_id, name, type, description, instructions, tools, config, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := ps.db.Exec(query,
		agent.ID, agent.UserID, agent.Name, agent.Type,
		agent.Description, agent.Instructions, agent.Tools, agent.Config, agent.Status,
	)
	return err
}

// GetAgent 获取智能体
func (ps *PostgresStorage) GetAgent(agentID string) (*models.Agent, error) {
	agent := &models.Agent{}
	query := `
		SELECT id, user_id, name, type, description, instructions, tools, config,
			   status, created_at, updated_at
		FROM agents WHERE id = $1
	`
	err := ps.db.QueryRow(query, agentID).Scan(
		&agent.ID, &agent.UserID, &agent.Name, &agent.Type,
		&agent.Description, &agent.Instructions, &agent.Tools, &agent.Config,
		&agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return agent, nil
}

// GetUserAgents 获取用户智能体列表
func (ps *PostgresStorage) GetUserAgents(userID string) ([]*models.Agent, error) {
	query := `
		SELECT id, user_id, name, type, description, instructions, tools, config,
			   status, created_at, updated_at
		FROM agents 
		WHERE user_id = $1 AND status = $2
		ORDER BY created_at DESC
	`
	
	rows, err := ps.db.Query(query, userID, models.AgentStatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		agent := &models.Agent{}
		err := rows.Scan(
			&agent.ID, &agent.UserID, &agent.Name, &agent.Type,
			&agent.Description, &agent.Instructions, &agent.Tools, &agent.Config,
			&agent.Status, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	return agents, nil
}

// Memory相关方法

// SaveConversationMemory 保存对话记忆
func (ps *PostgresStorage) SaveConversationMemory(memory *models.ConversationMemory) error {
	query := `
		INSERT INTO conversation_memory (session_id, user_id, messages, summary)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (session_id) 
		DO UPDATE SET messages = $3, summary = $4, updated_at = NOW()
	`
	_, err := ps.db.Exec(query, memory.SessionID, memory.UserID, memory.Messages, memory.Summary)
	return err
}

// GetConversationMemory 获取对话记忆
func (ps *PostgresStorage) GetConversationMemory(sessionID string) (*models.ConversationMemory, error) {
	memory := &models.ConversationMemory{}
	query := `
		SELECT session_id, user_id, messages, summary, updated_at
		FROM conversation_memory WHERE session_id = $1
	`
	err := ps.db.QueryRow(query, sessionID).Scan(
		&memory.SessionID, &memory.UserID, &memory.Messages, &memory.Summary, &memory.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return memory, nil
}

// Close 关闭数据库连接
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}