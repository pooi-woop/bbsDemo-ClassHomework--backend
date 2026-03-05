package handler

import (
	"bbsDemo/logger"
	"bbsDemo/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandler(userService *service.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) SendCode(c *gin.Context) {
	var req service.SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.SendVerificationCode(req); err != nil {
		switch err {
		case service.ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		default:
			logger.Error("Failed to send verification code", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send code"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent"})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(req)
	if err != nil {
		switch err {
		case service.ErrInvalidCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		case service.ErrCodeExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code expired"})
		default:
			logger.Error("Failed to register", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	tokenPair, err := h.userService.Login(req, ip, userAgent)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case service.ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		case service.ErrEmailNotVerified:
			c.JSON(http.StatusForbidden, gin.H{"error": "Email not verified"})
		default:
			logger.Error("Failed to login", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"tokens":  tokenPair,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()

	tokenPair, err := h.userService.RefreshToken(req.RefreshToken, ip, userAgent)
	if err != nil {
		switch err {
		case service.ErrInvalidToken:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		case service.ErrTokenExpired:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		default:
			logger.Error("Failed to refresh token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token refresh failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed",
		"tokens":  tokenPair,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.Logout(userID.(int64), req.RefreshToken); err != nil {
		logger.Error("Failed to logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, _ := c.Get("userID")

	if err := h.userService.LogoutAll(userID.(int64)); err != nil {
		logger.Error("Failed to logout all", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out from all devices"})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")

	user, err := h.userService.GetUserByID(userID.(int64))
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			logger.Error("Failed to get user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) UpdateNickname(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Nickname string `json:"nickname" binding:"required,max=50"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateNickname(userID.(int64), req.Nickname)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			logger.Error("Failed to update nickname", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update nickname"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) UpdateBio(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req struct {
		Bio string `json:"bio" binding:"max=500"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateBio(userID.(int64), req.Bio)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			logger.Error("Failed to update bio", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bio"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, _ := c.Get("userID")

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get file"})
		return
	}
	defer file.Close()

	avatarURL, err := h.userService.UploadAvatar(userID.(int64), header.Filename, header.Size, file)
	if err != nil {
		switch err {
		case service.ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large"})
		case service.ErrInvalidFileType:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
		default:
			logger.Error("Failed to upload avatar", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Avatar uploaded successfully",
		"avatar":  avatarURL,
	})
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	var req service.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.DeleteAccount(req); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case service.ErrInvalidVerificationCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		case service.ErrVerificationCodeUsed:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code already used"})
		case service.ErrVerificationCodeExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code expired"})
		default:
			logger.Error("Failed to delete account", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.ResetPassword(req); err != nil {
		switch err {
		case service.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case service.ErrInvalidVerificationCode:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
		case service.ErrVerificationCodeUsed:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code already used"})
		case service.ErrVerificationCodeExpired:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Verification code expired"})
		default:
			logger.Error("Failed to reset password", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
