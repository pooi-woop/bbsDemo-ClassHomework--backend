package middleware

import (
	"bbsDemo/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AdminRequired 检查用户是否为管理员
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		isAdmin, exists := c.Get("isAdmin")
		if !exists || !isAdmin.(bool) {
			logger.Warn("Admin access denied", zap.Int64("user_id", userID.(int64)))
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
