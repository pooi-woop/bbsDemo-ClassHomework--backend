package main

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/queue"
	"bbsDemo/router"
	"bbsDemo/service"
	"bbsDemo/utils"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
)

type ShutdownManager struct {
	server       *http.Server
	worker       *queue.Worker
	shutdownMu   sync.Mutex
	shutdownChan chan struct{}
}

func NewShutdownManager(server *http.Server, worker *queue.Worker) *ShutdownManager {
	return &ShutdownManager{
		server:       server,
		worker:       worker,
		shutdownChan: make(chan struct{}),
	}
}

func (sm *ShutdownManager) Shutdown(timeout time.Duration) error {
	sm.shutdownMu.Lock()
	defer sm.shutdownMu.Unlock()

	select {
	case <-sm.shutdownChan:
		return nil
	default:
		close(sm.shutdownChan)
	}

	logger.Info("Starting graceful shutdown", zap.Duration("timeout", timeout))

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	shutdownDone := make(chan error, 1)

	go func() {
		logger.Info("Stopping HTTP server")
		if err := sm.server.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown HTTP server", zap.Error(err))
			shutdownDone <- err
			return
		}
		logger.Info("HTTP server stopped")
		shutdownDone <- nil
	}()

	go func() {
		logger.Info("Stopping workers")
		if sm.worker.StopWithTimeout(timeout) {
			logger.Info("All workers stopped gracefully")
		} else {
			logger.Warn("Workers stopped with timeout")
		}
	}()

	select {
	case err := <-shutdownDone:
		logger.Info("Graceful shutdown completed")
		return err
	case <-time.After(timeout + time.Second):
		logger.Warn("Shutdown timeout exceeded, forcing exit")
		return fmt.Errorf("shutdown timeout")
	}
}

func (sm *ShutdownManager) IsShuttingDown() bool {
	select {
	case <-sm.shutdownChan:
		return true
	default:
		return false
	}
}

var shutdownManager *ShutdownManager

func main() {
	if len(os.Args) > 1 && os.Args[1] == "shutdown" {
		// 处理 shutdown 命令
		// 向服务器进程发送 SIGINT 信号
		fmt.Println("Sending shutdown signal to server...")

		// 这里需要找到服务器进程并发送信号
		// 暂时使用简单的实现，实际生产环境可能需要更复杂的进程管理
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

	worker := queue.NewWorker(cfg.Email, 3)
	worker.Start()

	userService := service.NewUserService(cfg.Email, cfg.Upload)
	postService := service.NewPostService()

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: nil,
	}

	shutdownManager = NewShutdownManager(server, worker)

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
