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
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrUserExists              = errors.New("user already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrInvalidPassword         = errors.New("invalid password")
	ErrInvalidCode             = errors.New("invalid verification code")
	ErrCodeExpired             = errors.New("verification code expired")
	ErrEmailNotVerified        = errors.New("email not verified")
	ErrInvalidToken            = errors.New("invalid token")
	ErrTokenExpired            = errors.New("token expired")
	ErrInvalidFileType         = errors.New("invalid file type")
	ErrFileTooLarge            = errors.New("file too large")
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeUsed    = errors.New("verification code already used")
	ErrVerificationCodeExpired = errors.New("verification code expired")
)

type verificationCode struct {
	Code      string
	Type      string
	ExpiresAt time.Time
	IsUsed    bool
}

type UserService struct {
	emailConfig  config.EmailConfig
	uploadConfig config.UploadConfig
	codes        map[string]verificationCode
	codesMutex   sync.RWMutex
}

func NewUserService(emailConfig config.EmailConfig, uploadConfig config.UploadConfig) *UserService {
	service := &UserService{
		emailConfig:  emailConfig,
		uploadConfig: uploadConfig,
		codes:        make(map[string]verificationCode),
	}

	// 启动定时清理过期验证码的协程
	go service.cleanupExpiredCodes()

	return service
}

func (s *UserService) cleanupExpiredCodes() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.codesMutex.Lock()
		for key, code := range s.codes {
			if time.Now().After(code.ExpiresAt) || code.IsUsed {
				delete(s.codes, key)
			}
		}
		s.codesMutex.Unlock()
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
	Type  string `json:"type" binding:"required,oneof=register reset delete"`
}

type DeleteAccountRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

func (s *UserService) SendVerificationCode(req SendCodeRequest) error {
	var user models.User
	err := database.DB.Where("email = ?", req.Email).First(&user).Error

	if req.Type == "register" {
		// 注册场景：用户已存在则报错，用户不存在则继续发送验证码
		if err == nil {
			return ErrUserExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		// 用户不存在，继续发送验证码
	} else {
		// 其他场景（重置密码、删除账号）：用户必须存在
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrUserNotFound
			}
			return err
		}
	}

	code := utils.GenerateVerificationCode()
	expiresAt := time.Now().Add(10 * time.Minute)

	// 检查是否有未使用且未过期的验证码
	s.codesMutex.RLock()
	for key, existingCode := range s.codes {
		if strings.Contains(key, req.Email) && existingCode.Type == req.Type && !existingCode.IsUsed && time.Now().Before(existingCode.ExpiresAt) {
			s.codesMutex.RUnlock()
			return errors.New("verification code already sent")
		}
	}
	s.codesMutex.RUnlock()

	// 生成唯一键
	key := fmt.Sprintf("%s:%s", req.Email, req.Type)

	// 存储验证码到内存
	s.codesMutex.Lock()
	s.codes[key] = verificationCode{
		Code:      code,
		Type:      req.Type,
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}
	s.codesMutex.Unlock()

	subject := "Verification Code - EyuForum（恶雨论坛）"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f4f4f4;
            padding: 20px;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo {
            font-size: 28px;
            font-weight: bold;
            color: #4a90e2;
        }
        .code-box {
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #ffffff;
            font-size: 36px;
            font-weight: bold;
            letter-spacing: 8px;
            padding: 20px;
            text-align: center;
            border-radius: 8px;
            margin: 30px 0;
            box-shadow: 0 4px 6px rgba(102, 126, 234, 0.3);
        }
        .info {
            text-align: center;
            color: #666;
            margin-bottom: 30px;
        }
        .footer {
            text-align: center;
            color: #999;
            font-size: 14px;
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
        }
        .timer {
            color: #e74c3c;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">📧 EyuForum（恶雨论坛）</div>
        </div>
        
        <p style="text-align: center; font-size: 18px; margin-bottom: 20px;">
            您好！
        </p>
        
        <p style="text-align: center; color: #666;">
            感谢您注册 EyuForum（恶雨论坛），请使用以下验证码完成注册：
        </p>
        
        <div class="code-box">%s</div>
        
        <div class="info">
            <p>验证码将在 <span class="timer">10分钟</span> 后过期</p>
            <p style="font-size: 14px; margin-top: 10px;">
                如果这不是您的操作，请忽略此邮件
            </p>
        </div>
        
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
            <p>© 2024 EyuForum（恶雨论坛）. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`, code)
	if err := database.PushEmail(req.Email, subject, body); err != nil {
		logger.Error("Failed to push email to queue", zap.Error(err))
		return err
	}

	logger.Info("Verification code queued", zap.String("email", req.Email), zap.String("type", req.Type))
	return nil
}

func (s *UserService) Register(req RegisterRequest) (*models.User, error) {
	key := fmt.Sprintf("%s:%s", req.Email, "register")

	s.codesMutex.Lock()
	code, exists := s.codes[key]
	if !exists || code.Code != req.Code || code.IsUsed {
		s.codesMutex.Unlock()
		return nil, ErrInvalidCode
	}

	if time.Now().After(code.ExpiresAt) {
		s.codesMutex.Unlock()
		return nil, ErrCodeExpired
	}

	// 标记为已使用
	code.IsUsed = true
	s.codes[key] = code
	s.codesMutex.Unlock()

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := models.User{
		ID:         utils.GenerateID(),
		Email:      req.Email,
		Password:   hashedPassword,
		Nickname:   "",
		IsVerified: true,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	logger.Info("User registered", zap.Int64("user_id", user.ID), zap.String("email", user.Email))
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

	if user.Status == 0 {
		return nil, errors.New("account is banned")
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

	logger.Info("User logged in", zap.Int64("user_id", user.ID), zap.String("email", user.Email))
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

	logger.Info("Token refreshed", zap.Int64("user_id", claims.UserID))
	return tokenPair, nil
}

func (s *UserService) Logout(userID int64, refreshToken string) error {
	if err := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ? AND token = ?", userID, refreshToken).
		Update("is_revoked", true).Error; err != nil {
		logger.Error("Failed to revoke token", zap.Error(err))
		return err
	}

	logger.Info("User logged out", zap.Int64("user_id", userID))
	return nil
}

func (s *UserService) LogoutAll(userID int64) error {
	if err := database.DB.Model(&models.RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true).Error; err != nil {
		logger.Error("Failed to revoke all tokens", zap.Error(err))
		return err
	}

	logger.Info("User logged out from all devices", zap.Int64("user_id", userID))
	return nil
}

func (s *UserService) GetUserByID(userID int64) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	if user.DeletedAt.Valid {
		user.DeletedAtStr = user.DeletedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &user, nil
}

func (s *UserService) GetAllUsers(page, pageSize int, keyword string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.User{})

	// 如果有关键词，添加搜索条件
	if keyword != "" {
		query = query.Where("nickname LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count users", zap.Error(err))
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		logger.Error("Failed to get users", zap.Error(err))
		return nil, 0, err
	}

	// 处理 deleted_at 字段
	for i := range users {
		if users[i].DeletedAt.Valid {
			users[i].DeletedAtStr = users[i].DeletedAt.Time.Format("2006-01-02 15:04:05")
		}
	}

	return users, total, nil
}

func (s *UserService) UpdateNickname(userID int64, nickname string) (*models.User, error) {
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

	logger.Info("Nickname updated", zap.Int64("user_id", user.ID))
	return &user, nil
}

func (s *UserService) UpdateBio(userID int64, bio string) (*models.User, error) {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	user.Bio = bio
	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to update bio", zap.Error(err))
		return nil, err
	}

	logger.Info("Bio updated", zap.Int64("user_id", user.ID))
	return &user, nil
}

func (s *UserService) DeleteAccount(req DeleteAccountRequest) error {
	key := fmt.Sprintf("%s:%s", req.Email, "delete")
	s.codesMutex.RLock()
	codeData, exists := s.codes[key]
	s.codesMutex.RUnlock()

	if !exists {
		return ErrInvalidVerificationCode
	}

	if codeData.IsUsed {
		return ErrVerificationCodeUsed
	}

	if time.Now().After(codeData.ExpiresAt) {
		return ErrVerificationCodeExpired
	}

	if codeData.Code != req.Code {
		return ErrInvalidVerificationCode
	}

	s.codesMutex.Lock()
	s.codes[key] = verificationCode{
		Code:      codeData.Code,
		Type:      codeData.Type,
		ExpiresAt: codeData.ExpiresAt,
		IsUsed:    true,
	}
	s.codesMutex.Unlock()

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return err
	}

	if err := database.DB.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{}).Error; err != nil {
		logger.Warn("Failed to delete user refresh tokens", zap.Error(err))
	}

	logger.Info("User account deleted", zap.Int64("user_id", user.ID), zap.String("email", req.Email))
	return nil
}

func (s *UserService) ResetPassword(req ResetPasswordRequest) error {
	key := fmt.Sprintf("%s:%s", req.Email, "reset")
	s.codesMutex.RLock()
	codeData, exists := s.codes[key]
	s.codesMutex.RUnlock()

	if !exists {
		return ErrInvalidVerificationCode
	}

	if codeData.IsUsed {
		return ErrVerificationCodeUsed
	}

	if time.Now().After(codeData.ExpiresAt) {
		return ErrVerificationCodeExpired
	}

	if codeData.Code != req.Code {
		return ErrInvalidVerificationCode
	}

	s.codesMutex.Lock()
	s.codes[key] = verificationCode{
		Code:      codeData.Code,
		Type:      codeData.Type,
		ExpiresAt: codeData.ExpiresAt,
		IsUsed:    true,
	}
	s.codesMutex.Unlock()

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return err
	}

	if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
		logger.Error("Failed to reset password", zap.Error(err))
		return err
	}

	if err := database.DB.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{}).Error; err != nil {
		logger.Warn("Failed to delete user refresh tokens after password reset", zap.Error(err))
	}

	logger.Info("User password reset", zap.Int64("user_id", user.ID), zap.String("email", req.Email))
	return nil
}

func (s *UserService) UploadAvatar(userID int64, fileName string, fileSize int64, fileContent io.Reader) (string, error) {
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

	logger.Info("Avatar uploaded", zap.Int64("user_id", userID), zap.String("avatar", avatarURL))
	return avatarURL, nil
}
