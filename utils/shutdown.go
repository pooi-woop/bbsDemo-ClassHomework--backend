package utils

import (
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/queue"
	"bbsDemo/scheduler"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type ShutdownManager struct {
	server       *http.Server
	worker       *queue.Worker
	scheduler    *scheduler.Scheduler
	shutdownMu   sync.Mutex
	shutdownChan chan struct{}
}

func NewShutdownManager(server *http.Server, worker *queue.Worker, sched *scheduler.Scheduler) *ShutdownManager {
	return &ShutdownManager{
		server:       server,
		worker:       worker,
		scheduler:    sched,
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

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		logger.Info("Stopping HTTP server")
		if err := sm.server.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		} else {
			logger.Info("HTTP server stopped")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Stopping workers")
		if sm.worker.StopWithTimeout(timeout) {
			logger.Info("All workers stopped gracefully")
		} else {
			logger.Warn("Workers stopped with timeout")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Stopping scheduler")
		if sm.scheduler != nil {
			sm.scheduler.Stop()
			logger.Info("Scheduler stopped")
		}
	}()

	go func() {
		defer wg.Done()
		logger.Info("Closing Kafka connections")
		if err := database.CloseKafka(); err != nil {
			logger.Error("Failed to close Kafka", zap.Error(err))
		} else {
			logger.Info("Kafka connections closed")
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Graceful shutdown completed")
		return nil
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
