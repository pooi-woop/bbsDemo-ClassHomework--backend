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
		&models.VerificationCode{},
		&models.RefreshToken{},
		&models.Post{},
		&models.Comment{},
		&models.Like{},
		&models.Block{},
		&models.FavoriteFolder{},
		&models.Favorite{},
	); err != nil {
		logger.Error("Failed to auto migrate", zap.Error(err))
		return err
	}

	logger.Info("Auto migrate completed successfully")
	return nil
}
