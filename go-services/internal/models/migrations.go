package models

import (
	"fmt"
	"log"
	"gorm.io/gorm"
)

// Migration 数据库迁移结构
type Migration struct {
	Version     string
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// MigrationRecord 迁移记录
type MigrationRecord struct {
	ID          uint   `gorm:"primarykey"`
	Version     string `gorm:"uniqueIndex;size:50;not null"`
	Description string `gorm:"size:200"`
	AppliedAt   int64  `gorm:"not null"`
	Checksum    string `gorm:"size:64"`
}

// GetMigrations 获取所有迁移
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     "001_initial_schema",
			Description: "Create initial database schema",
			Up:          migration001Up,
			Down:        migration001Down,
		},
		{
			Version:     "002_add_vector_extension",
			Description: "Add pgvector extension for embeddings",
			Up:          migration002Up,
			Down:        migration002Down,
		},
		{
			Version:     "003_add_indexes",
			Description: "Add performance indexes",
			Up:          migration003Up,
			Down:        migration003Down,
		},
		{
			Version:     "004_add_fulltext_search",
			Description: "Add full-text search capabilities",
			Up:          migration004Up,
			Down:        migration004Down,
		},
		{
			Version:     "005_add_partitioning",
			Description: "Add table partitioning for logs and metrics",
			Up:          migration005Up,
			Down:        migration005Down,
		},
	}
}

// migration001Up 初始数据库结构
func migration001Up(db *gorm.DB) error {
	log.Println("Applying migration 001: Initial schema")
	
	// 自动迁移所有模型
	err := db.AutoMigrate(
		&User{},
		&Agent{},
		&Session{},
		&Message{},
		&Document{},
		&DocumentChunk{},
		&Task{},
		&Memory{},
		&Tool{},
		&ToolCall{},
		&SystemLog{},
		&ApiKey{},
		&Config{},
		&Webhook{},
	)
	
	if err != nil {
		return fmt.Errorf("failed to auto-migrate models: %w", err)
	}
	
	// 创建默认配置
	defaultConfigs := []Config{
		{Key: "system.name", Value: "PolyAgent", Category: "system", Description: "System name"},
		{Key: "system.version", Value: "1.0.0", Category: "system", Description: "System version"},
		{Key: "ai.default_model", Value: "gpt-3.5-turbo", Category: "ai", Description: "Default AI model"},
		{Key: "ai.max_tokens", Value: "2000", Category: "ai", Description: "Maximum tokens per request"},
		{Key: "rag.chunk_size", Value: "512", Category: "rag", Description: "Default chunk size for documents"},
		{Key: "rag.overlap_size", Value: "50", Category: "rag", Description: "Overlap size between chunks"},
	}
	
	for _, config := range defaultConfigs {
		var existing Config
		if err := db.Where("key = ?", config.Key).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&config).Error; err != nil {
					return fmt.Errorf("failed to create default config %s: %w", config.Key, err)
				}
			}
		}
	}
	
	// 创建默认工具
	defaultTools := []Tool{
		{
			Name:        "web_search",
			DisplayName: "Web Search",
			Description: "Search the web for information",
			Category:    "search",
			Status:      "active",
			Config: JSONMap{
				"api_endpoint": "https://api.search.com",
				"timeout":      30,
			},
			Schema: JSONMap{
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
						"required":    true,
					},
				},
			},
		},
		{
			Name:        "calculator",
			DisplayName: "Calculator",
			Description: "Perform mathematical calculations",
			Category:    "math",
			Status:      "active",
			Config: JSONMap{
				"precision": 10,
			},
			Schema: JSONMap{
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type":        "string",
						"description": "Mathematical expression",
						"required":    true,
					},
				},
			},
		},
	}
	
	for _, tool := range defaultTools {
		var existing Tool
		if err := db.Where("name = ?", tool.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tool).Error; err != nil {
					return fmt.Errorf("failed to create default tool %s: %w", tool.Name, err)
				}
			}
		}
	}
	
	log.Println("Migration 001 completed successfully")
	return nil
}

func migration001Down(db *gorm.DB) error {
	log.Println("Rolling back migration 001")
	
	tables := []string{
		"webhooks", "configs", "api_keys", "system_logs", "tool_calls", "tools",
		"memories", "tasks", "document_chunks", "documents", "messages",
		"sessions", "agents", "users",
	}
	
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}
	
	log.Println("Migration 001 rollback completed")
	return nil
}

// migration002Up 添加 pgvector 扩展
func migration002Up(db *gorm.DB) error {
	log.Println("Applying migration 002: Add pgvector extension")
	
	// 启用 pgvector 扩展
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}
	
	// 为向量字段创建索引
	queries := []string{
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding ON document_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
		"CREATE INDEX IF NOT EXISTS idx_memories_embedding ON memories USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100)",
	}
	
	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute query '%s': %v", query, err)
		}
	}
	
	log.Println("Migration 002 completed successfully")
	return nil
}

func migration002Down(db *gorm.DB) error {
	log.Println("Rolling back migration 002")
	
	// 删除向量索引
	queries := []string{
		"DROP INDEX IF EXISTS idx_document_chunks_embedding",
		"DROP INDEX IF EXISTS idx_memories_embedding",
	}
	
	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute query '%s': %v", query, err)
		}
	}
	
	// 注意：不删除 pgvector 扩展，因为可能被其他应用使用
	
	log.Println("Migration 002 rollback completed")
	return nil
}

// migration003Up 添加性能索引
func migration003Up(db *gorm.DB) error {
	log.Println("Applying migration 003: Add performance indexes")
	
	indexes := []string{
		// 用户相关索引
		"CREATE INDEX IF NOT EXISTS idx_users_status ON users (status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_users_last_login ON users (last_login_at DESC) WHERE deleted_at IS NULL",
		
		// 智能体相关索引
		"CREATE INDEX IF NOT EXISTS idx_agents_type_status ON agents (type, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_agents_last_used ON agents (last_used_at DESC) WHERE deleted_at IS NULL",
		
		// 会话相关索引
		"CREATE INDEX IF NOT EXISTS idx_sessions_user_status ON sessions (user_id, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_sessions_last_message ON sessions (last_message DESC) WHERE deleted_at IS NULL",
		
		// 消息相关索引
		"CREATE INDEX IF NOT EXISTS idx_messages_session_created ON messages (session_id, created_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_messages_user_role ON messages (user_id, role, created_at DESC) WHERE deleted_at IS NULL",
		
		// 文档相关索引
		"CREATE INDEX IF NOT EXISTS idx_documents_user_status ON documents (user_id, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_documents_type_indexed ON documents (type, indexed_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_documents_content_hash ON documents (content_hash) WHERE deleted_at IS NULL",
		
		// 任务相关索引
		"CREATE INDEX IF NOT EXISTS idx_tasks_status_priority ON tasks (status, priority DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tasks_user_status ON tasks (user_id, status) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tasks_scheduled ON tasks (scheduled_for ASC) WHERE scheduled_for IS NOT NULL AND deleted_at IS NULL",
		
		// 工具调用索引
		"CREATE INDEX IF NOT EXISTS idx_tool_calls_tool_status ON tool_calls (tool_id, status, created_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tool_calls_user_created ON tool_calls (user_id, created_at DESC) WHERE deleted_at IS NULL",
		
		// 系统日志索引
		"CREATE INDEX IF NOT EXISTS idx_system_logs_level_service ON system_logs (level, service, created_at DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_system_logs_user_created ON system_logs (user_id, created_at DESC) WHERE user_id IS NOT NULL AND deleted_at IS NULL",
		
		// 记忆相关索引
		"CREATE INDEX IF NOT EXISTS idx_memories_user_type ON memories (user_id, type, importance DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_memories_content_hash ON memories (content_hash) WHERE deleted_at IS NULL",
	}
	
	for _, query := range indexes {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to create index: %s, error: %v", query, err)
		}
	}
	
	log.Println("Migration 003 completed successfully")
	return nil
}

func migration003Down(db *gorm.DB) error {
	log.Println("Rolling back migration 003")
	
	indexes := []string{
		"DROP INDEX IF EXISTS idx_users_status",
		"DROP INDEX IF EXISTS idx_users_last_login",
		"DROP INDEX IF EXISTS idx_agents_type_status",
		"DROP INDEX IF EXISTS idx_agents_last_used",
		"DROP INDEX IF EXISTS idx_sessions_user_status",
		"DROP INDEX IF EXISTS idx_sessions_last_message",
		"DROP INDEX IF EXISTS idx_messages_session_created",
		"DROP INDEX IF EXISTS idx_messages_user_role",
		"DROP INDEX IF EXISTS idx_documents_user_status",
		"DROP INDEX IF EXISTS idx_documents_type_indexed",
		"DROP INDEX IF EXISTS idx_documents_content_hash",
		"DROP INDEX IF EXISTS idx_tasks_status_priority",
		"DROP INDEX IF EXISTS idx_tasks_user_status",
		"DROP INDEX IF EXISTS idx_tasks_scheduled",
		"DROP INDEX IF EXISTS idx_tool_calls_tool_status",
		"DROP INDEX IF EXISTS idx_tool_calls_user_created",
		"DROP INDEX IF EXISTS idx_system_logs_level_service",
		"DROP INDEX IF EXISTS idx_system_logs_user_created",
		"DROP INDEX IF EXISTS idx_memories_user_type",
		"DROP INDEX IF EXISTS idx_memories_content_hash",
	}
	
	for _, query := range indexes {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to drop index: %s, error: %v", query, err)
		}
	}
	
	log.Println("Migration 003 rollback completed")
	return nil
}

// migration004Up 添加全文搜索
func migration004Up(db *gorm.DB) error {
	log.Println("Applying migration 004: Add full-text search")
	
	// 为文档和消息添加全文搜索索引
	queries := []string{
		// 文档全文搜索
		"ALTER TABLE documents ADD COLUMN IF NOT EXISTS search_vector tsvector",
		"CREATE INDEX IF NOT EXISTS idx_documents_search_vector ON documents USING gin(search_vector)",
		
		// 文档块全文搜索
		"ALTER TABLE document_chunks ADD COLUMN IF NOT EXISTS search_vector tsvector",
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_search_vector ON document_chunks USING gin(search_vector)",
		
		// 消息全文搜索
		"ALTER TABLE messages ADD COLUMN IF NOT EXISTS search_vector tsvector",
		"CREATE INDEX IF NOT EXISTS idx_messages_search_vector ON messages USING gin(search_vector)",
		
		// 记忆全文搜索
		"ALTER TABLE memories ADD COLUMN IF NOT EXISTS search_vector tsvector",
		"CREATE INDEX IF NOT EXISTS idx_memories_search_vector ON memories USING gin(search_vector)",
	}
	
	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute query '%s': %v", query, err)
		}
	}
	
	// 创建更新搜索向量的触发器函数
	triggerFunctions := []string{
		`CREATE OR REPLACE FUNCTION update_documents_search_vector() RETURNS trigger AS $$
		BEGIN
			NEW.search_vector := setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
								setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
		
		`CREATE OR REPLACE FUNCTION update_document_chunks_search_vector() RETURNS trigger AS $$
		BEGIN
			NEW.search_vector := to_tsvector('english', COALESCE(NEW.content, ''));
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
		
		`CREATE OR REPLACE FUNCTION update_messages_search_vector() RETURNS trigger AS $$
		BEGIN
			NEW.search_vector := to_tsvector('english', COALESCE(NEW.content, ''));
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
		
		`CREATE OR REPLACE FUNCTION update_memories_search_vector() RETURNS trigger AS $$
		BEGIN
			NEW.search_vector := to_tsvector('english', COALESCE(NEW.content, ''));
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`,
	}
	
	for _, fn := range triggerFunctions {
		if err := db.Exec(fn).Error; err != nil {
			log.Printf("Warning: Failed to create trigger function: %v", err)
		}
	}
	
	// 创建触发器
	triggers := []string{
		"DROP TRIGGER IF EXISTS trig_update_documents_search_vector ON documents",
		"CREATE TRIGGER trig_update_documents_search_vector BEFORE INSERT OR UPDATE ON documents FOR EACH ROW EXECUTE FUNCTION update_documents_search_vector()",
		
		"DROP TRIGGER IF EXISTS trig_update_document_chunks_search_vector ON document_chunks",
		"CREATE TRIGGER trig_update_document_chunks_search_vector BEFORE INSERT OR UPDATE ON document_chunks FOR EACH ROW EXECUTE FUNCTION update_document_chunks_search_vector()",
		
		"DROP TRIGGER IF EXISTS trig_update_messages_search_vector ON messages",
		"CREATE TRIGGER trig_update_messages_search_vector BEFORE INSERT OR UPDATE ON messages FOR EACH ROW EXECUTE FUNCTION update_messages_search_vector()",
		
		"DROP TRIGGER IF EXISTS trig_update_memories_search_vector ON memories",
		"CREATE TRIGGER trig_update_memories_search_vector BEFORE INSERT OR UPDATE ON memories FOR EACH ROW EXECUTE FUNCTION update_memories_search_vector()",
	}
	
	for _, trigger := range triggers {
		if err := db.Exec(trigger).Error; err != nil {
			log.Printf("Warning: Failed to create trigger: %s, error: %v", trigger, err)
		}
	}
	
	log.Println("Migration 004 completed successfully")
	return nil
}

func migration004Down(db *gorm.DB) error {
	log.Println("Rolling back migration 004")
	
	// 删除触发器
	triggers := []string{
		"DROP TRIGGER IF EXISTS trig_update_documents_search_vector ON documents",
		"DROP TRIGGER IF EXISTS trig_update_document_chunks_search_vector ON document_chunks",
		"DROP TRIGGER IF EXISTS trig_update_messages_search_vector ON messages",
		"DROP TRIGGER IF EXISTS trig_update_memories_search_vector ON memories",
	}
	
	for _, trigger := range triggers {
		if err := db.Exec(trigger).Error; err != nil {
			log.Printf("Warning: Failed to drop trigger: %v", err)
		}
	}
	
	// 删除触发器函数
	functions := []string{
		"DROP FUNCTION IF EXISTS update_documents_search_vector()",
		"DROP FUNCTION IF EXISTS update_document_chunks_search_vector()",
		"DROP FUNCTION IF EXISTS update_messages_search_vector()",
		"DROP FUNCTION IF EXISTS update_memories_search_vector()",
	}
	
	for _, fn := range functions {
		if err := db.Exec(fn).Error; err != nil {
			log.Printf("Warning: Failed to drop function: %v", err)
		}
	}
	
	// 删除搜索列
	queries := []string{
		"ALTER TABLE documents DROP COLUMN IF EXISTS search_vector",
		"ALTER TABLE document_chunks DROP COLUMN IF EXISTS search_vector",
		"ALTER TABLE messages DROP COLUMN IF EXISTS search_vector",
		"ALTER TABLE memories DROP COLUMN IF EXISTS search_vector",
	}
	
	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute query: %v", err)
		}
	}
	
	log.Println("Migration 004 rollback completed")
	return nil
}

// migration005Up 添加表分区
func migration005Up(db *gorm.DB) error {
	log.Println("Applying migration 005: Add table partitioning")
	
	// 为系统日志创建分区表
	partitionQueries := []string{
		// 创建系统日志分区表
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'system_logs_partitioned') THEN
				CREATE TABLE system_logs_partitioned (LIKE system_logs INCLUDING ALL);
				ALTER TABLE system_logs_partitioned ADD CONSTRAINT system_logs_partitioned_pkey PRIMARY KEY (id, created_at);
			END IF;
		END $$`,
		
		// 为每个月创建分区（示例）
		`DO $$
		DECLARE
			start_date DATE := DATE_TRUNC('month', CURRENT_DATE - INTERVAL '6 months');
			end_date DATE := DATE_TRUNC('month', CURRENT_DATE + INTERVAL '6 months');
			partition_date DATE := start_date;
			partition_name TEXT;
		BEGIN
			WHILE partition_date < end_date LOOP
				partition_name := 'system_logs_' || TO_CHAR(partition_date, 'YYYY_MM');
				
				IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = partition_name) THEN
					EXECUTE FORMAT('CREATE TABLE %I PARTITION OF system_logs_partitioned
									FOR VALUES FROM (%L) TO (%L)',
								   partition_name,
								   partition_date,
								   partition_date + INTERVAL '1 month');
				END IF;
				
				partition_date := partition_date + INTERVAL '1 month';
			END LOOP;
		END $$`,
	}
	
	for _, query := range partitionQueries {
		if err := db.Exec(query).Error; err != nil {
			log.Printf("Warning: Failed to execute partitioning query: %v", err)
		}
	}
	
	log.Println("Migration 005 completed successfully")
	return nil
}

func migration005Down(db *gorm.DB) error {
	log.Println("Rolling back migration 005")
	
	// 删除分区表
	if err := db.Exec("DROP TABLE IF EXISTS system_logs_partitioned CASCADE").Error; err != nil {
		log.Printf("Warning: Failed to drop partitioned table: %v", err)
	}
	
	log.Println("Migration 005 rollback completed")
	return nil
}