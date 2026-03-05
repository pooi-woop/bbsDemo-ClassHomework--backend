package database

import (
	"bbsDemo/config"
	"bbsDemo/logger"
	"bbsDemo/models"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMySQL(cfg config.MySQLConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		logger.Error("Failed to connect database",
			zap.Error(err),
			zap.String("host", cfg.Host),
			zap.Int("port", cfg.Port),
			zap.String("database", cfg.Database))
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get sql.DB", zap.Error(err))
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	logger.Info("MySQL connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database))
	return nil
}

func CloseMySQL() error {
	sqlDB, err := DB.DB()
	if err != nil {
		logger.Error("Failed to get sql.DB for closing", zap.Error(err))
		return err
	}

	logger.Info("MySQL connection closed")
	return sqlDB.Close()
}

func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Post{},
		&models.Like{},
		&models.Comment{},

		&models.Block{},
		&models.FavoriteFolder{},
		&models.Favorite{},
	); err != nil {
		logger.Error("Failed to auto migrate", zap.Error(err))
		return err
	}

	// 创建条件唯一索引：只在 deleted_at IS NULL 时强制 email 唯一
	// 这样已删除的用户不会阻止新用户使用相同的邮箱注册
	if err := createUniqueIndexIfNotExists(); err != nil {
		logger.Error("Failed to create unique index", zap.Error(err))
		return err
	}

	logger.Info("Auto migrate completed successfully")
	return nil
}

func createUniqueIndexIfNotExists() error {
	// 检查索引是否存在
	var count int64
	result := DB.Raw(`
		SELECT COUNT(*) FROM information_schema.STATISTICS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'users' 
		AND INDEX_NAME = 'idx_users_email_active'
	`).Scan(&count)

	if result.Error != nil {
		return result.Error
	}

	// 如果索引不存在，创建它
	if count == 0 {
		// 先删除可能存在的旧唯一索引（使用 MySQL 兼容的语法）
		DB.Exec(`DROP INDEX idx_users_email ON users`)

		// MySQL 8.0.13+ 支持函数索引，使用 (email) 作为表达式
		// 对于旧版本 MySQL，我们需要使用不同的策略
		// 这里使用生成列的方式实现条件唯一索引

		// 首先检查是否存在 active_email 列
		var colCount int64
		DB.Raw(`
			SELECT COUNT(*) FROM information_schema.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'users' 
			AND COLUMN_NAME = 'active_email'
		`).Scan(&colCount)

		if colCount == 0 {
			// 添加生成列：当 deleted_at 为 NULL 时，active_email = email，否则为 NULL
			DB.Exec(`
				ALTER TABLE users 
				ADD COLUMN active_email VARCHAR(255) AS (
					CASE WHEN deleted_at IS NULL THEN email ELSE NULL END
				) STORED
			`)

			// 在生成列上创建唯一索引
			result = DB.Exec(`
				CREATE UNIQUE INDEX idx_users_email_active 
				ON users(active_email)
			`)
			if result.Error != nil {
				// 如果创建失败，尝试删除生成列并返回错误
				DB.Exec(`ALTER TABLE users DROP COLUMN active_email`)
				return result.Error
			}
			logger.Info("Created unique index idx_users_email_active using generated column")
		}
	}

	return nil
}
