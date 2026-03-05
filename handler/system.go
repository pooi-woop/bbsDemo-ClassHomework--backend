package handler

import (
	"bbsDemo/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SystemHandler struct {
	shutdownFunc func(timeout time.Duration) error
}

func NewSystemHandler(shutdownFunc func(time.Duration) error) *SystemHandler {
	return &SystemHandler{
		shutdownFunc: shutdownFunc,
	}
}

func (h *SystemHandler) Shutdown(c *gin.Context) {
	logger.Info("Shutdown request received")

	go func() {
		timeout := 30 * time.Second
		if err := h.shutdownFunc(timeout); err != nil {
			logger.Error("Error during shutdown", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Server is shutting down gracefully",
		"timeout": 30.0,
	})
}
