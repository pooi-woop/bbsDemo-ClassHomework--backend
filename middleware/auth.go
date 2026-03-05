package middleware

import (
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"bbsDemo/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired 认证中间件，验证JWT令牌并设置用户信息
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		var tokenString string
		if _, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		if claims.Type != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token type"})
			c.Abort()
			return
		}

		// 获取用户信息，包括是否为管理员
		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("isAdmin", user.IsAdmin)
		c.Next()
	}
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("userID")
	if !exists {
		logger.Error("Failed to get userID from context")
		return 0
	}
	return userID.(int64)
}

// GetIsAdmin 从上下文中获取管理员状态
func GetIsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("isAdmin")
	if !exists {
		return false
	}
	return isAdmin.(bool)
}

// OptionalAuth 可选认证中间件，验证JWT令牌但不强制要求
// 如果令牌有效，设置用户信息；如果无效或不存在，继续处理请求
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		var tokenString string
		if _, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenString); err != nil {
			c.Next()
			return
		}

		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		if claims.Type != "access" {
			c.Next()
			return
		}

		// 获取用户信息
		var user models.User
		if err := database.DB.First(&user, claims.UserID).Error; err != nil {
			c.Next()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("isAdmin", user.IsAdmin)
		c.Next()
	}
}
