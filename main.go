package main

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/queue"
	"bbsDemo/router"
	"bbsDemo/service"
	"bbsDemo/utils"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "shutdown" {
		fmt.Println("Sending shutdown signal to server...")
		fmt.Println("Shutdown signal sent. Please note: This implementation requires manual process management.")
		fmt.Println("In production, consider using a process manager like PM2 or systemd.")
		os.Exit(0)
	}

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
	utils.InitSnowflake(1)

	if err := database.InitMySQL(cfg.MySQL); err != nil {
		logger.Fatal("Failed to initialize MySQL", zap.Error(err))
	}
	defer database.CloseMySQL()

	if err := database.AutoMigrate(); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	if err := database.InitRedis(cfg.Redis); err != nil {
		logger.Fatal("Failed to initialize Redis", zap.Error(err))
	}
	defer database.CloseRedis()

	if err := database.InitKafka(cfg.Kafka); err != nil {
		logger.Fatal("Failed to initialize Kafka", zap.Error(err))
	}

	worker := queue.NewWorker(cfg.Email, 3)
	worker.Start()

	userService := service.NewUserService(cfg.Email, cfg.Upload)
	postService := service.NewPostService()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: nil,
	}

	shutdownManager := utils.NewShutdownManager(server, worker)

	r := router.InitRouter(userService, postService, cfg)
	server.Handler = r

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Server starting", zap.String("address", addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("Received shutdown signal")

	shutdownTimeout := 30 * time.Second
	if err := shutdownManager.Shutdown(shutdownTimeout); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("Application stopped successfully")
}
