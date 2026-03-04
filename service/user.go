package service

import (
	"bbsDemo/config"
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"bbsDemo/utils"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrUserNotFound     = errors.New("user not found")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrInvalidCode      = errors.New("invalid verification code")
	ErrCodeExpired      = errors.New("verification code expired")
	ErrEmailNotVerified = errors.New("email not verified")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrFileTooLarge     = errors.New("file too large")
)

type UserService struct {
	emailConfig  config.EmailConfig
	uploadConfig config.UploadConfig
}

func NewUserService(emailConfig config.EmailConfig, uploadConfig config.UploadConfig) *UserService {
	return &UserService{
		emailConfig:  emailConfig,
		uploadConfig: uploadConfig,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Code     string `json:"code" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register reset"`
}

func (s *UserService) SendVerificationCode(req SendCodeRequest) error {
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err == nil && req.Type == "register" {
		return ErrUserExists
	}

	code := utils.GenerateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	verificationCode := models.VerificationCode{
		Email:     req.Email,
		Code:      code,
		Type:      req.Type,
		ExpiresAt: expiresAt,
	}

	if err := database.DB.Create(&verificationCode).Error; err != nil {
		logger.Error("Failed to save verification code", zap.Error(err))
		return err
	}

	subject := "Verification Code"
	body := fmt.Sprintf("Your verification code is: %s. It will expire in 10 minutes.", code)
	if err := database.PushEmail(req.Email, subject, body); err != nil {
		logger.Error("Failed to push email to queue", zap.Error(err))
		return err
	}

	logger.Info("Verification code queued", zap.String("email", req.Email), zap.String("type", req.Type))
	return nil
}

func (s *UserService) Register(req RegisterRequest) (*models.User, error) {
	var code models.VerificationCode
	if err := database.DB.Where("email = ? AND code = ? AND type = ? AND is_used = ?",
		req.Email, req.Code, "register", false).
		Order("created_at DESC").
		First(&code).Error; err != nil {
		return nil, ErrInvalidCode
	}

	if time.Now().After(code.ExpiresAt) {
		return nil, ErrCodeExpired
	}

	code.IsUsed = true
	database.DB.Save(&code)

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := models.User{
		Email:      req.Email,
		Password:   hashedPassword,
		Nickname:   "",
		IsVerified: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	logger.Info("User registered", zap.Uint("user_id", user.ID), zap.String("email", user.Email))
	return &user, nil
}

func (s *UserService) Login(req LoginRequest, ip, userAgent string) (*utils.TokenPair, error) {
	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	if !utils.VerifyPassword(req.Password, user.Password) {
		return nil, ErrInvalidPassword
	}

	if !user.IsVerified {
		return nil, ErrEmailNotVerified
	}

	tokenPair, err := utils.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, err
	}

	refreshToken := models.RefreshToken{
		UserID:    user.ID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := database.DB.Create(&refreshToken).Error; err != nil {
		logger.Error("Failed to save refresh token", zap.Error(err))
		return nil, err
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ip
	database.DB.Save(&user)

	logger.Info("User logged in", zap.Uint("user_id", user.ID), zap.String("email", user.Email))
	return tokenPair, nil
}

func (s *UserService) RefreshToken(refreshToken string, ip, userAgent string) (*utils.TokenPair, error) {
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	var token models.RefreshToken
	if err := database.DB.Where("token = ? AND is_revoked = ?", refreshToken, false).First(&token).Error; err != nil {
		return nil, ErrInvalidToken
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	token.IsRevoked = true
	database.DB.Save(&token)

	tokenPair, err := utils.GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, err
	}

	newRefreshToken := models.RefreshToken{
		UserID:    claims.UserID,
		Token:     tokenPair.RefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		IP:        ip,
		UserAgent: userAgent,
	}
	if err := database.DB.Create(&newRefreshToken).Error; err != nil {
		logger.Error("Failed to save refresh token", zap.Error(err))
		return nil, err
	}

	logger.Info("Token refreshed", zap.Uint("user_id", claims.UserID))
	return tokenPair, nil
}

func (s *UserService) Logout(userID uint, refreshToken string) error {
	if err := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND token = ?", userID, refreshToken).
		Update("is_revoked", true).Error; err != nil {
		logger.Error("Failed to revoke token", zap.Error(err))
		return err
	}

	logger.Info("User logged out", zap.Uint("user_id", userID))
	return nil
}

func (s *UserService) LogoutAll(userID uint) error {
	if err := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error; err != nil {
		logger.Error("Failed to revoke all tokens", zap.Error(err))
		return err
	}

	logger.Info("User logged out from all devices", zap.Uint("user_id", userID))
	return nil
}

func (s *UserService) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) UpdateNickname(userID uint, nickname string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Nickname = nickname
	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to update nickname", zap.Error(err))
		return nil, err
	}

	logger.Info("Nickname updated", zap.Uint("user_id", user.ID))
	return &user, nil
}

func (s *UserService) UploadAvatar(userID uint, fileName string, fileSize int64, fileContent io.Reader) (string, error) {
	if fileSize > s.uploadConfig.MaxSize {
		return "", ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	allowedExts := strings.Split(s.uploadConfig.AllowedExt, ",")
	isAllowed := false
	for _, allowed := range allowedExts {
		if strings.TrimSpace(allowed) == ext {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return "", ErrInvalidFileType
	}

	if err := os.MkdirAll(s.uploadConfig.Path, 0755); err != nil {
		logger.Error("Failed to create upload directory", zap.Error(err))
		return "", err
	}

	newFileName := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(s.uploadConfig.Path, newFileName)

	file, err := os.Create(filePath)
	if err != nil {
		logger.Error("Failed to create file", zap.Error(err))
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, fileContent); err != nil {
		logger.Error("Failed to save file", zap.Error(err))
		return "", err
	}

	avatarURL := fmt.Sprintf("/uploads/%s", newFileName)

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return "", err
	}

	if user.Avatar != "" {
		oldPath := filepath.Join(s.uploadConfig.Path, filepath.Base(user.Avatar))
		os.Remove(oldPath)
	}

	user.Avatar = avatarURL
	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to update avatar", zap.Error(err))
		return "", err
	}

	logger.Info("Avatar uploaded", zap.Uint("user_id", userID), zap.String("avatar", avatarURL))
	return avatarURL, nil
}
