package service

import (
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"bbsDemo/utils"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrPostNotFound     = errors.New("post not found")
	ErrCommentNotFound  = errors.New("comment not found")
	ErrAlreadyLiked     = errors.New("already liked")
	ErrNotLiked         = errors.New("not liked")
	ErrAlreadyBlocked   = errors.New("already blocked")
	ErrNotBlocked       = errors.New("not blocked")
	ErrAlreadyFavorited = errors.New("already favorited")
	ErrNotFavorited     = errors.New("not favorited")
	ErrCannotBlockSelf  = errors.New("cannot block yourself")
	ErrFolderNotFound   = errors.New("folder not found")
	ErrFolderExists     = errors.New("folder already exists")
	ErrDefaultFolder    = errors.New("cannot delete default folder")
	ErrFolderNotYours   = errors.New("folder not yours")
)

type PostService struct{}

type FavoriteWithFolder struct {
	Post       models.Post `json:"post"`
	FolderID   uint        `json:"folder_id"`
	FolderName string      `json:"folder_name"`
	CreatedAt  string      `json:"created_at"`
}

func NewPostService() *PostService {
	return &PostService{}
}

type CreatePostRequest struct {
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"max=200"`
	Content string `json:"content"`
}

type CreateCommentRequest struct {
	PostID    *uint  `json:"post_id"`
	CommentID *uint  `json:"comment_id"`
	Content   string `json:"content" binding:"required"`
}

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required,max=50"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" binding:"required,max=50"`
}

type FavoritePostRequest struct {
	PostID   int64 `json:"post_id" binding:"required"`
	FolderID uint  `json:"folder_id" binding:"required"`
}

func (s *PostService) CreatePost(userID int64, req CreatePostRequest) (*models.Post, error) {
	post := models.Post{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Views:   0,
	}

	if err := database.DB.Create(&post).Error; err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return nil, err
	}

	logger.Info("Post created", zap.Int64("post_id", post.ID), zap.Int64("user_id", userID))
	return &post, nil
}

func (s *PostService) UpdatePost(userID int64, postID int64, req UpdatePostRequest) (*models.Post, error) {
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if post.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	if err := database.DB.Save(&post).Error; err != nil {
		logger.Error("Failed to update post", zap.Error(err))
		return nil, err
	}

	logger.Info("Post updated", zap.Int64("post_id", post.ID))
	return &post, nil
}

func (s *PostService) DeletePost(userID int64, postID int64) error {
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	if post.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := database.DB.Delete(&post).Error; err != nil {
		logger.Error("Failed to delete post", zap.Error(err))
		return err
	}

	logger.Info("Post deleted", zap.Int64("post_id", post.ID))
	return nil
}

func (s *PostService) GetPost(postID int64) (*models.Post, error) {
	var post models.Post
	if err := database.DB.Preload("User").First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPostNotFound
		}
		return nil, err
	}

	if err := database.PushViewCount(postID); err != nil {
		logger.Error("Failed to push view count to queue", zap.Error(err))
	}

	return &post, nil
}

func (s *PostService) ListPosts(page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	if err := database.DB.Model(&models.Post{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := database.DB.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (s *PostService) CreateComment(userID int64, req CreateCommentRequest) (*models.Comment, error) {
	if req.PostID == nil && req.CommentID == nil {
		return nil, errors.New("post_id or comment_id is required")
	}

	comment := models.Comment{
		UserID:    userID,
		PostID:    req.PostID,
		CommentID: req.CommentID,
		Content:   req.Content,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		logger.Error("Failed to create comment", zap.Error(err))
		return nil, err
	}

	logger.Info("Comment created", zap.Uint("comment_id", comment.ID), zap.Int64("user_id", userID))
	return &comment, nil
}

func (s *PostService) DeleteComment(userID int64, commentID uint) error {
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if comment.UserID != userID {
		return errors.New("unauthorized")
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		logger.Error("Failed to delete comment", zap.Error(err))
		return err
	}

	logger.Info("Comment deleted", zap.Uint("comment_id", comment.ID))
	return nil
}

func (s *PostService) GetComments(postID uint, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Comment{}).Where("post_id = ?", postID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Where("comment_id IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *PostService) GetReplies(commentID uint, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Comment{}).Where("comment_id = ?", commentID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *PostService) LikePost(userID int64, postID int64) error {
	var existingLike models.Like
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingLike).Error; err == nil {
		return ErrAlreadyLiked
	}

	like := models.Like{
		UserID: userID,
		PostID: &postID,
	}

	if err := database.DB.Create(&like).Error; err != nil {
		logger.Error("Failed to like post", zap.Error(err))
		return err
	}

	if err := database.PushLikeCount(postID, 0, "like"); err != nil {
		logger.Error("Failed to push like count to queue", zap.Error(err))
	}

	logger.Info("Post liked", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) UnlikePost(userID int64, postID int64) error {
	var like models.Like
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotLiked
		}
		return err
	}

	if err := database.DB.Delete(&like).Error; err != nil {
		logger.Error("Failed to unlike post", zap.Error(err))
		return err
	}

	if err := database.PushLikeCount(postID, 0, "unlike"); err != nil {
		logger.Error("Failed to push unlike count to queue", zap.Error(err))
	}

	logger.Info("Post unliked", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) LikeComment(userID int64, commentID uint) error {
	var existingLike models.Like
	if err := database.DB.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&existingLike).Error; err == nil {
		return ErrAlreadyLiked
	}

	like := models.Like{
		UserID:    userID,
		CommentID: &commentID,
	}

	if err := database.DB.Create(&like).Error; err != nil {
		logger.Error("Failed to like comment", zap.Error(err))
		return err
	}

	if err := database.PushLikeCount(0, commentID, "like"); err != nil {
		logger.Error("Failed to push like count to queue", zap.Error(err))
	}

	logger.Info("Comment liked", zap.Uint("comment_id", commentID), zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) UnlikeComment(userID int64, commentID uint) error {
	var like models.Like
	if err := database.DB.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&like).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotLiked
		}
		return err
	}

	if err := database.DB.Delete(&like).Error; err != nil {
		logger.Error("Failed to unlike comment", zap.Error(err))
		return err
	}

	if err := database.PushLikeCount(0, commentID, "unlike"); err != nil {
		logger.Error("Failed to push unlike count to queue", zap.Error(err))
	}

	logger.Info("Comment unliked", zap.Uint("comment_id", commentID), zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) BlockUser(userID, blockedID int64) error {
	if userID == blockedID {
		return ErrCannotBlockSelf
	}

	var existingBlock models.Block
	if err := database.DB.Where("user_id = ? AND blocked_id = ?", userID, blockedID).First(&existingBlock).Error; err == nil {
		return ErrAlreadyBlocked
	}

	block := models.Block{
		UserID:    userID,
		BlockedID: blockedID,
	}

	if err := database.DB.Create(&block).Error; err != nil {
		logger.Error("Failed to block user", zap.Error(err))
		return err
	}

	logger.Info("User blocked", zap.Int64("user_id", userID), zap.Int64("blocked_id", blockedID))
	return nil
}

func (s *PostService) UnblockUser(userID, blockedID int64) error {
	var block models.Block
	if err := database.DB.Where("user_id = ? AND blocked_id = ?", userID, blockedID).First(&block).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotBlocked
		}
		return err
	}

	if err := database.DB.Delete(&block).Error; err != nil {
		logger.Error("Failed to unblock user", zap.Error(err))
		return err
	}

	logger.Info("User unblocked", zap.Int64("user_id", userID), zap.Int64("blocked_id", blockedID))
	return nil
}

func (s *PostService) GetBlockedUsers(userID int64, page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.User{}).
		Joins("JOIN blocks ON users.id = blocks.blocked_id").
		Where("blocks.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *PostService) CreateFolder(userID int64, req CreateFolderRequest) (*models.FavoriteFolder, error) {
	var existingFolder models.FavoriteFolder
	if err := database.DB.Where("user_id = ? AND name = ?", userID, req.Name).First(&existingFolder).Error; err == nil {
		return nil, ErrFolderExists
	}

	folder := models.FavoriteFolder{
		UserID: userID,
		Name:   req.Name,
	}

	if err := database.DB.Create(&folder).Error; err != nil {
		logger.Error("Failed to create folder", zap.Error(err))
		return nil, err
	}

	logger.Info("Folder created", zap.Uint("folder_id", folder.ID), zap.Int64("user_id", userID))
	return &folder, nil
}

func (s *PostService) UpdateFolder(userID int64, folderID uint, req UpdateFolderRequest) (*models.FavoriteFolder, error) {
	var folder models.FavoriteFolder
	if err := database.DB.First(&folder, folderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFolderNotFound
		}
		return nil, err
	}

	if folder.UserID != userID {
		return nil, ErrFolderNotYours
	}

	var existingFolder models.FavoriteFolder
	if err := database.DB.Where("user_id = ? AND name = ? AND id != ?", userID, req.Name, folderID).First(&existingFolder).Error; err == nil {
		return nil, ErrFolderExists
	}

	folder.Name = req.Name
	if err := database.DB.Save(&folder).Error; err != nil {
		logger.Error("Failed to update folder", zap.Error(err))
		return nil, err
	}

	logger.Info("Folder updated", zap.Uint("folder_id", folder.ID))
	return &folder, nil
}

func (s *PostService) DeleteFolder(userID int64, folderID uint) error {
	var folder models.FavoriteFolder
	if err := database.DB.First(&folder, folderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return err
	}

	if folder.UserID != userID {
		return ErrFolderNotYours
	}

	if folder.IsDefault {
		return ErrDefaultFolder
	}

	if err := database.DB.Delete(&folder).Error; err != nil {
		logger.Error("Failed to delete folder", zap.Error(err))
		return err
	}

	logger.Info("Folder deleted", zap.Uint("folder_id", folder.ID))
	return nil
}

func (s *PostService) GetFolders(userID int64) ([]models.FavoriteFolder, error) {
	var folders []models.FavoriteFolder
	if err := database.DB.Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&folders).Error; err != nil {
		return nil, err
	}
	return folders, nil
}

func (s *PostService) GetOrCreateDefaultFolder(userID int64) (*models.FavoriteFolder, error) {
	var folder models.FavoriteFolder
	if err := database.DB.Where("user_id = ? AND is_default = ?", userID, true).First(&folder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			folder = models.FavoriteFolder{
				UserID:    userID,
				Name:      "默认收藏夹",
				IsDefault: true,
			}
			if err := database.DB.Create(&folder).Error; err != nil {
				return nil, err
			}
			logger.Info("Default folder created", zap.Int64("user_id", userID))
			return &folder, nil
		}
		return nil, err
	}
	return &folder, nil
}

func (s *PostService) FavoritePost(userID int64, req FavoritePostRequest) error {
	var folder models.FavoriteFolder
	if err := database.DB.First(&folder, req.FolderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return err
	}

	if folder.UserID != userID {
		return ErrFolderNotYours
	}

	var existingFavorite models.Favorite
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, req.PostID).First(&existingFavorite).Error; err == nil {
		return ErrAlreadyFavorited
	}

	favorite := models.Favorite{
		UserID:   userID,
		PostID:   req.PostID,
		FolderID: req.FolderID,
	}

	if err := database.DB.Create(&favorite).Error; err != nil {
		logger.Error("Failed to favorite post", zap.Error(err))
		return err
	}

	logger.Info("Post favorited", zap.Int64("post_id", req.PostID), zap.Int64("user_id", userID), zap.Uint("folder_id", req.FolderID))
	return nil
}

func (s *PostService) UnfavoritePost(userID int64, postID int64) error {
	var favorite models.Favorite
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFavorited
		}
		return err
	}

	if err := database.DB.Delete(&favorite).Error; err != nil {
		logger.Error("Failed to unfavorite post", zap.Error(err))
		return err
	}

	logger.Info("Post unfavorited", zap.Int64("post_id", postID), zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) MoveFavorite(userID int64, postID int64, folderID uint) error {
	var favorite models.Favorite
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFavorited
		}
		return err
	}

	var folder models.FavoriteFolder
	if err := database.DB.First(&folder, folderID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFolderNotFound
		}
		return err
	}

	if folder.UserID != userID {
		return ErrFolderNotYours
	}

	favorite.FolderID = folderID
	if err := database.DB.Save(&favorite).Error; err != nil {
		logger.Error("Failed to move favorite", zap.Error(err))
		return err
	}

	logger.Info("Favorite moved", zap.Int64("post_id", postID), zap.Uint("folder_id", folderID))
	return nil
}

func (s *PostService) GetFavorites(userID int64, page, pageSize int) ([]FavoriteWithFolder, int64, error) {
	var favorites []models.Favorite
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Favorite{}).
		Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("Post").Preload("Post.User").Preload("Folder").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&favorites).Error; err != nil {
		return nil, 0, err
	}

	result := make([]FavoriteWithFolder, len(favorites))
	for i, f := range favorites {
		result[i] = FavoriteWithFolder{
			Post:       *f.Post,
			FolderID:   f.FolderID,
			FolderName: f.Folder.Name,
			CreatedAt:  f.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return result, total, nil
}

func (s *PostService) GetFavoritesByFolder(userID int64, folderID uint, page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Post{}).
		Joins("JOIN favorites ON posts.id = favorites.post_id").
		Where("favorites.user_id = ? AND favorites.folder_id = ?", userID, folderID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Order("favorites.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (s *PostService) GetMyPosts(userID int64, page, pageSize int) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Model(&models.Post{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}
