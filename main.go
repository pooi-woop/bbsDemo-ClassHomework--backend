package main

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/router"
	"bbsDemo/service"
	"bbsDemo/utils"
	"fmt"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	if err := logger.InitLogger(cfg.Logger); err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Application starting", zap.String("config", "config.yaml"))

	utils.InitJWT(cfg.JWT.Secret)

	if err := database.InitMySQL(cfg.MySQL); err != nil {
		logger.Fatal("Failed to initialize MySQL", zap.Error(err))
	}
	defer database.CloseMySQL()

	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	userService := service.NewUserService(cfg.Email, cfg.Upload)
	postService := service.NewPostService()

	r := router.InitRouter(userService, postService, cfg)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Server starting", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
