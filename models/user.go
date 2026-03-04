package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email          string `gorm:"uniqueIndex;not null" json:"email"`
	Password       string `gorm:"not null" json:"-"`
	Nickname       string `json:"nickname"`
	Avatar         string `json:"avatar"`
	Status         int    `gorm:"default:1" json:"status"`
	IsVerified     bool   `gorm:"default:false" json:"is_verified"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastLoginIP    string `json:"last_login_ip"`
}

type VerificationCode struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email     string    `gorm:"index;not null" json:"email"`
	Code      string    `gorm:"not null" json:"-"`
	Type      string    `gorm:"not null" json:"type"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
}

type RefreshToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"-"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsRevoked bool      `gorm:"default:false" json:"is_revoked"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
}

func (User) TableName() string {
	return "users"
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}

func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
