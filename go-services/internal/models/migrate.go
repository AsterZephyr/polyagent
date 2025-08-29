package models

import (
	"crypto/sha256"
	"fmt"
	"log"
	"time"
	"gorm.io/gorm"
)

// Migrator 数据库迁移管理器
type Migrator struct {
	db *gorm.DB
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Initialize 初始化迁移系统
func (m *Migrator) Initialize() error {
	// 创建迁移记录表
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migration_records table: %w", err)
	}
	
	log.Println("Migration system initialized successfully")
	return nil
}

// Migrate 执行所有待执行的迁移
func (m *Migrator) Migrate() error {
	log.Println("Starting database migration...")
	
	migrations := GetMigrations()
	
	for _, migration := range migrations {
		applied, err := m.isMigrationApplied(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}
		
		if applied {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}
		
		log.Printf("Applying migration %s: %s", migration.Version, migration.Description)
		
		// 在事务中执行迁移
		err = m.db.Transaction(func(tx *gorm.DB) error {
			// 执行迁移
			if err := migration.Up(tx); err != nil {
				return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
			}
			
			// 记录迁移
			record := MigrationRecord{
				Version:     migration.Version,
				Description: migration.Description,
				AppliedAt:   time.Now().Unix(),
				Checksum:    m.calculateMigrationChecksum(migration),
			}
			
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
			}
			
			return nil
		})
		
		if err != nil {
			return err
		}
		
		log.Printf("Migration %s applied successfully", migration.Version)
	}
	
	log.Println("Database migration completed successfully")
	return nil
}

// Rollback 回滚指定数量的迁移
func (m *Migrator) Rollback(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("rollback steps must be positive")
	}
	
	log.Printf("Rolling back %d migration(s)...", steps)
	
	// 获取已应用的迁移记录（按应用时间倒序）
	var appliedRecords []MigrationRecord
	if err := m.db.Order("applied_at DESC").Limit(steps).Find(&appliedRecords).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	if len(appliedRecords) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}
	
	migrations := GetMigrations()
	migrationMap := make(map[string]Migration)
	for _, migration := range migrations {
		migrationMap[migration.Version] = migration
	}
	
	// 按应用顺序回滚（最新的先回滚）
	for _, record := range appliedRecords {
		migration, exists := migrationMap[record.Version]
		if !exists {
			log.Printf("Warning: Migration %s not found in code, skipping rollback", record.Version)
			continue
		}
		
		log.Printf("Rolling back migration %s: %s", record.Version, record.Description)
		
		// 在事务中执行回滚
		err := m.db.Transaction(func(tx *gorm.DB) error {
			// 执行回滚
			if err := migration.Down(tx); err != nil {
				return fmt.Errorf("failed to rollback migration %s: %w", record.Version, err)
			}
			
			// 删除迁移记录
			if err := tx.Delete(&record).Error; err != nil {
				return fmt.Errorf("failed to delete migration record %s: %w", record.Version, err)
			}
			
			return nil
		})
		
		if err != nil {
			return err
		}
		
		log.Printf("Migration %s rolled back successfully", record.Version)
	}
	
	log.Printf("Rolled back %d migration(s) successfully", len(appliedRecords))
	return nil
}

// Status 显示迁移状态
func (m *Migrator) Status() error {
	log.Println("Migration Status:")
	log.Println("================")
	
	migrations := GetMigrations()
	appliedMap := make(map[string]MigrationRecord)
	
	// 获取所有已应用的迁移
	var appliedRecords []MigrationRecord
	if err := m.db.Find(&appliedRecords).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	for _, record := range appliedRecords {
		appliedMap[record.Version] = record
	}
	
	// 显示每个迁移的状态
	for _, migration := range migrations {
		if record, applied := appliedMap[migration.Version]; applied {
			appliedTime := time.Unix(record.AppliedAt, 0)
			log.Printf("✓ %s - %s (Applied: %s)", 
				migration.Version, 
				migration.Description, 
				appliedTime.Format("2006-01-02 15:04:05"))
		} else {
			log.Printf("✗ %s - %s (Pending)", 
				migration.Version, 
				migration.Description)
		}
	}
	
	return nil
}

// isMigrationApplied 检查迁移是否已应用
func (m *Migrator) isMigrationApplied(version string) (bool, error) {
	var record MigrationRecord
	err := m.db.Where("version = ?", version).First(&record).Error
	
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	
	if err != nil {
		return false, err
	}
	
	return true, nil
}

// calculateMigrationChecksum 计算迁移的校验和
func (m *Migrator) calculateMigrationChecksum(migration Migration) string {
	data := fmt.Sprintf("%s_%s", migration.Version, migration.Description)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Validate 验证迁移完整性
func (m *Migrator) Validate() error {
	log.Println("Validating migration integrity...")
	
	migrations := GetMigrations()
	
	// 获取所有已应用的迁移记录
	var appliedRecords []MigrationRecord
	if err := m.db.Find(&appliedRecords).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	migrationMap := make(map[string]Migration)
	for _, migration := range migrations {
		migrationMap[migration.Version] = migration
	}
	
	// 验证每个已应用的迁移
	for _, record := range appliedRecords {
		migration, exists := migrationMap[record.Version]
		if !exists {
			log.Printf("Warning: Applied migration %s not found in code", record.Version)
			continue
		}
		
		expectedChecksum := m.calculateMigrationChecksum(migration)
		if record.Checksum != expectedChecksum {
			return fmt.Errorf("migration %s has been modified (checksum mismatch)", record.Version)
		}
	}
	
	log.Println("Migration validation completed successfully")
	return nil
}

// Reset 重置数据库（谨慎使用）
func (m *Migrator) Reset() error {
	log.Println("WARNING: This will reset the entire database!")
	log.Println("Resetting database...")
	
	migrations := GetMigrations()
	
	// 获取所有已应用的迁移记录（按应用时间倒序）
	var appliedRecords []MigrationRecord
	if err := m.db.Order("applied_at DESC").Find(&appliedRecords).Error; err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	
	migrationMap := make(map[string]Migration)
	for _, migration := range migrations {
		migrationMap[migration.Version] = migration
	}
	
	// 回滚所有迁移
	for _, record := range appliedRecords {
		migration, exists := migrationMap[record.Version]
		if !exists {
			log.Printf("Warning: Migration %s not found in code, skipping rollback", record.Version)
			continue
		}
		
		log.Printf("Rolling back migration %s", record.Version)
		
		if err := migration.Down(m.db); err != nil {
			log.Printf("Warning: Failed to rollback migration %s: %v", record.Version, err)
		}
	}
	
	// 删除所有迁移记录
	if err := m.db.Where("1=1").Delete(&MigrationRecord{}).Error; err != nil {
		log.Printf("Warning: Failed to delete migration records: %v", err)
	}
	
	log.Println("Database reset completed")
	return nil
}

// Fresh 清空数据库并重新执行所有迁移
func (m *Migrator) Fresh() error {
	log.Println("Performing fresh migration (reset + migrate)...")
	
	if err := m.Reset(); err != nil {
		return fmt.Errorf("failed to reset database: %w", err)
	}
	
	if err := m.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize migration system: %w", err)
	}
	
	if err := m.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	
	log.Println("Fresh migration completed successfully")
	return nil
}

// CreateMigration 创建新的迁移文件（辅助功能）
func (m *Migrator) CreateMigration(name, description string) error {
	timestamp := time.Now().Format("20060102150405")
	version := fmt.Sprintf("%s_%s", timestamp, name)
	
	log.Printf("Creating new migration: %s", version)
	log.Printf("Description: %s", description)
	log.Println("Please implement the Up and Down functions in migrations.go")
	
	// 这里可以生成迁移文件模板
	template := fmt.Sprintf(`
// migration%sUp %s
func migration%sUp(db *gorm.DB) error {
	log.Println("Applying migration %s: %s")
	
	// TODO: Implement migration logic here
	
	log.Println("Migration %s completed successfully")
	return nil
}

func migration%sDown(db *gorm.DB) error {
	log.Println("Rolling back migration %s")
	
	// TODO: Implement rollback logic here
	
	log.Println("Migration %s rollback completed")
	return nil
}
`, version, description, version, version, description, version, version, version)
	
	log.Printf("Migration template:\n%s", template)
	
	return nil
}

// GetAppliedMigrations 获取已应用的迁移列表
func (m *Migrator) GetAppliedMigrations() ([]MigrationRecord, error) {
	var records []MigrationRecord
	if err := m.db.Order("applied_at ASC").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}
	return records, nil
}

// GetPendingMigrations 获取待应用的迁移列表
func (m *Migrator) GetPendingMigrations() ([]Migration, error) {
	migrations := GetMigrations()
	var pending []Migration
	
	for _, migration := range migrations {
		applied, err := m.isMigrationApplied(migration.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to check migration status: %w", err)
		}
		
		if !applied {
			pending = append(pending, migration)
		}
	}
	
	return pending, nil
}