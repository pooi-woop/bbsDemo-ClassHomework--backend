package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID        int64          `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID  int64  `gorm:"index;not null" json:"user_id"`
	Title   string `gorm:"size:200;not null" json:"title"`
	Content string `gorm:"type:text;not null" json:"content"`
	Views   int    `gorm:"default:0" json:"views"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Comment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	PostID    *uint  `gorm:"index" json:"post_id"`
	CommentID *uint  `gorm:"index" json:"comment_id"`
	UserID    int64  `gorm:"index;not null" json:"user_id"`
	Content   string `gorm:"type:text;not null" json:"content"`
	IsDeleted bool   `gorm:"default:false" json:"is_deleted"`

	Post    *Post    `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Comment *Comment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Like struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    int64  `gorm:"index;not null" json:"user_id"`
	PostID    *int64 `gorm:"index" json:"post_id"`
	CommentID *uint  `gorm:"index" json:"comment_id"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post    *Post    `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Comment *Comment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
}

type Block struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    int64 `gorm:"index;not null" json:"user_id"`
	BlockedID int64 `gorm:"index;not null" json:"blocked_id"`

	User    *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Blocked *User `gorm:"foreignKey:BlockedID" json:"blocked,omitempty"`
}

type FavoriteFolder struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID    int64  `gorm:"index;not null" json:"user_id"`
	Name      string `gorm:"size:50;not null" json:"name"`
	IsDefault bool   `gorm:"default:false" json:"is_default"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

type Favorite struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID   int64 `gorm:"index;not null" json:"user_id"`
	PostID   int64 `gorm:"index;not null" json:"post_id"`
	FolderID uint  `gorm:"index;not null" json:"folder_id"`

	User   *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Post   *Post           `gorm:"foreignKey:PostID" json:"post,omitempty"`
	Folder *FavoriteFolder `gorm:"foreignKey:FolderID" json:"folder,omitempty"`
}

func (Post) TableName() string {
	return "posts"
}

func (Comment) TableName() string {
	return "comments"
}

func (Like) TableName() string {
	return "likes"
}

func (Block) TableName() string {
	return "blocks"
}

func (FavoriteFolder) TableName() string {
	return "favorite_folders"
}

func (Favorite) TableName() string {
	return "favorites"
}
