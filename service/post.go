package service

import (
	"bbsDemo/database"
	"bbsDemo/logger"
	"bbsDemo/models"
	"bbsDemo/utils"
	"errors"
	"strconv"
	"time"

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

type BlockedUserWithStatus struct {
	ID          int64      `json:"id,string"`
	Email       string     `json:"email"`
	Nickname    string     `json:"nickname"`
	Bio         string     `json:"bio"`
	Avatar      string     `json:"avatar"`
	IsAdmin     bool       `json:"is_admin"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	BlockedAt   time.Time  `json:"blocked_at"`
	UnblockedAt *time.Time `json:"unblocked_at,omitempty"`
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
	PostID    interface{} `json:"post_id"`
	CommentID interface{} `json:"comment_id"`
	Content   string      `json:"content" binding:"required"`
}

type PostWithStatus struct {
	models.Post
	IsLiked     bool `json:"is_liked"`
	IsFavorited bool `json:"is_favorited"`
}

type CreateFolderRequest struct {
	Name string `json:"name" binding:"required,max=50"`
}

type UpdateFolderRequest struct {
	Name string `json:"name" binding:"required,max=50"`
}

type FavoritePostRequest struct {
	PostID   string `json:"post_id" binding:"required"`
	FolderID uint   `json:"folder_id" binding:"required"`
}

func (s *PostService) CreatePost(userID int64, req CreatePostRequest) (*models.Post, error) {
	// 检查用户是否被禁言
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if user.Status == 0 {
		return nil, errors.New("user is banned")
	}

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

	if err := database.DB.Delete(&models.Post{}, postID).Error; err != nil {
		logger.Error("Failed to delete post", zap.Error(err))
		return err
	}

	logger.Info("Post deleted", zap.Int64("post_id", post.ID))
	return nil
}

func (s *PostService) DeletePostWithAdminCheck(userID int64, postID int64, isAdmin bool) error {
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	if post.UserID != userID && !isAdmin {
		return errors.New("unauthorized")
	}

	if err := database.DB.Delete(&models.Post{}, postID).Error; err != nil {
		logger.Error("Failed to delete post", zap.Error(err))
		return err
	}

	logger.Info("Post deleted", zap.Int64("post_id", post.ID), zap.Bool("is_admin", isAdmin))
	return nil
}

type PostWithStatusAndComments struct {
	PostWithStatus
	Comments []models.Comment `json:"comments"`
}

func (s *PostService) GetPost(postID, userID int64) (*PostWithStatusAndComments, error) {
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

	// 创建带状态的帖子
	postWithStatus := PostWithStatus{
		Post:        post,
		IsLiked:     false,
		IsFavorited: false,
	}

	// 如果用户已登录，检查点赞和收藏状态
	if userID > 0 {
		// 检查是否点赞
		var like models.Like
		if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&like).Error; err == nil {
			postWithStatus.IsLiked = true
		}

		// 检查是否收藏
		var favorite models.Favorite
		err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&favorite).Error
		logger.Info("Check favorite status",
			zap.Int64("user_id", userID),
			zap.Int64("post_id", post.ID),
			zap.Bool("found", err == nil),
			zap.Error(err))
		if err == nil {
			postWithStatus.IsFavorited = true
		}
	}

	// 查询帖子的所有评论
	var comments []models.Comment
	if err := database.DB.Where("post_id = ?", postID).
		Preload("User").
		Preload("Comment"). // 加载回复的评论
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		logger.Error("Failed to get comments", zap.Error(err))
		// 不返回错误，只记录日志
	}

	// 为评论添加用户的点赞状态
	if userID > 0 {
		for i := range comments {
			var like models.Like
			if err := database.DB.Where("user_id = ? AND comment_id = ?", userID, comments[i].ID).First(&like).Error; err == nil {
				// 这里可以添加评论的点赞状态，需要在 Comment 模型中添加字段
			}
		}
	}

	return &PostWithStatusAndComments{
		PostWithStatus: postWithStatus,
		Comments:       comments,
	}, nil
}

func (s *PostService) ListPosts(userID int64, keyword string, page, pageSize int) ([]PostWithStatus, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	logger.Info("List posts service",
		zap.Int64("user_id", userID),
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// 构建基础查询
	query := database.DB.Model(&models.Post{})

	// 处理搜索关键字
	if keyword != "" {
		logger.Info("Searching posts with keyword", zap.String("keyword", keyword))
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 如果用户已登录，过滤被拉黑用户的内容
	if userID > 0 {
		var blockedIDs []int64
		database.DB.Model(&models.Block{}).Where("user_id = ? AND unblocked_at IS NULL", userID).Pluck("blocked_id", &blockedIDs)
		if len(blockedIDs) > 0 {
			query = query.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 重新构建查询以包含 Preload
	findQuery := database.DB.Model(&models.Post{}).Preload("User")
	if keyword != "" {
		findQuery = findQuery.Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if userID > 0 {
		var blockedIDs []int64
		database.DB.Model(&models.Block{}).Where("user_id = ? AND unblocked_at IS NULL", userID).Pluck("blocked_id", &blockedIDs)
		if len(blockedIDs) > 0 {
			findQuery = findQuery.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	if err := findQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	// 转换为带状态的帖子
	postsWithStatus := make([]PostWithStatus, len(posts))
	for i, post := range posts {
		postsWithStatus[i] = PostWithStatus{
			Post:        post,
			IsLiked:     false,
			IsFavorited: false,
		}

		// 如果用户已登录，检查点赞和收藏状态
		if userID > 0 {
			// 检查是否点赞
			var like models.Like
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&like).Error; err == nil {
				postsWithStatus[i].IsLiked = true
			}

			// 检查是否收藏
			var favorite models.Favorite
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&favorite).Error; err == nil {
				postsWithStatus[i].IsFavorited = true
			}
		}
	}

	return postsWithStatus, total, nil
}

func (s *PostService) SearchPosts(userID int64, keyword string, page, pageSize int) ([]PostWithStatus, int64, error) {
	var posts []models.Post
	var total int64

	offset := (page - 1) * pageSize

	logger.Info("Search posts",
		zap.Int64("user_id", userID),
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// 如果关键词为空，返回空结果
	if keyword == "" {
		return []PostWithStatus{}, 0, nil
	}

	query := database.DB.Model(&models.Post{}).Preload("User").
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 如果用户已登录，过滤被拉黑用户的内容
	if userID > 0 {
		var blockedIDs []int64
		database.DB.Model(&models.Block{}).Where("user_id = ? AND unblocked_at IS NULL AND deleted_at IS NULL", userID).Pluck("blocked_id", &blockedIDs)
		if len(blockedIDs) > 0 {
			query = query.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	// 转换为带状态的帖子
	postsWithStatus := make([]PostWithStatus, len(posts))
	for i, post := range posts {
		postsWithStatus[i] = PostWithStatus{
			Post:        post,
			IsLiked:     false,
			IsFavorited: false,
		}

		// 如果用户已登录，检查点赞和收藏状态
		if userID > 0 {
			// 检查是否点赞
			var like models.Like
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&like).Error; err == nil {
				postsWithStatus[i].IsLiked = true
			}

			// 检查是否收藏
			var favorite models.Favorite
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&favorite).Error; err == nil {
				postsWithStatus[i].IsFavorited = true
			}
		}
	}

	return postsWithStatus, total, nil
}

// SearchResult 综合搜索结果
type SearchResult struct {
	Posts []PostWithStatus `json:"posts"`
	Users []models.User    `json:"users"`
	Total struct {
		Posts int64 `json:"posts"`
		Users int64 `json:"users"`
	} `json:"total"`
}

// Search 综合搜索，同时搜索用户和帖子
func (s *PostService) Search(userID int64, keyword string, page, pageSize int) (*SearchResult, error) {
	logger.Info("Search request",
		zap.Int64("user_id", userID),
		zap.String("keyword", keyword),
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	result := &SearchResult{
		Posts: []PostWithStatus{},
		Users: []models.User{},
	}

	// 如果关键词为空，返回空结果
	if keyword == "" {
		return result, nil
	}

	offset := (page - 1) * pageSize

	// 搜索帖子
	var posts []models.Post
	var postTotal int64

	postQuery := database.DB.Model(&models.Post{}).Preload("User").
		Where("title LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 如果用户已登录，过滤被拉黑用户的内容
	if userID > 0 {
		var blockedIDs []int64
		database.DB.Model(&models.Block{}).Where("user_id = ? AND unblocked_at IS NULL AND deleted_at IS NULL", userID).Pluck("blocked_id", &blockedIDs)
		if len(blockedIDs) > 0 {
			postQuery = postQuery.Where("user_id NOT IN ?", blockedIDs)
		}
	}

	if err := postQuery.Count(&postTotal).Error; err != nil {
		logger.Error("Failed to count posts", zap.Error(err))
		return nil, err
	}

	if err := postQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		logger.Error("Failed to search posts", zap.Error(err))
		return nil, err
	}

	// 转换为带状态的帖子
	postsWithStatus := make([]PostWithStatus, len(posts))
	for i, post := range posts {
		postsWithStatus[i] = PostWithStatus{
			Post:        post,
			IsLiked:     false,
			IsFavorited: false,
		}

		// 如果用户已登录，检查点赞和收藏状态
		if userID > 0 {
			// 检查是否点赞
			var like models.Like
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&like).Error; err == nil {
				postsWithStatus[i].IsLiked = true
			}

			// 检查是否收藏
			var favorite models.Favorite
			if err := database.DB.Where("user_id = ? AND post_id = ?", userID, post.ID).First(&favorite).Error; err == nil {
				postsWithStatus[i].IsFavorited = true
			}
		}
	}
	result.Posts = postsWithStatus
	result.Total.Posts = postTotal

	// 搜索用户
	var users []models.User
	var userTotal int64

	userQuery := database.DB.Model(&models.User{}).
		Where("nickname LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	if err := userQuery.Count(&userTotal).Error; err != nil {
		logger.Error("Failed to count users", zap.Error(err))
		return nil, err
	}

	if err := userQuery.
		Select("id, email, nickname, bio, avatar, status, is_admin, is_verified, created_at, last_login_at").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		logger.Error("Failed to search users", zap.Error(err))
		return nil, err
	}

	result.Users = users
	result.Total.Users = userTotal

	logger.Info("Search completed",
		zap.Int64("post_total", postTotal),
		zap.Int64("user_total", userTotal))

	return result, nil
}

func (s *PostService) CreateComment(userID int64, req CreateCommentRequest) (*models.Comment, error) {
	// 检查用户是否被禁言
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	if user.Status == 0 {
		return nil, errors.New("user is banned")
	}

	logger.Info("Create comment request",
		zap.Int64("user_id", userID),
		zap.Any("post_id", req.PostID),
		zap.Any("comment_id", req.CommentID),
		zap.String("content", req.Content))

	if req.PostID == nil && req.CommentID == nil {
		logger.Error("Neither post_id nor comment_id provided")
		return nil, errors.New("post_id or comment_id is required")
	}

	comment := models.Comment{
		UserID:  userID,
		Content: req.Content,
	}

	// 处理 PostID（int64 类型）
	if req.PostID != nil {
		var postIDStr string
		switch v := req.PostID.(type) {
		case string:
			postIDStr = v
		case float64:
			postIDStr = strconv.FormatInt(int64(v), 10)
		default:
			logger.Error("Invalid post_id type", zap.Any("post_id", req.PostID))
			return nil, errors.New("invalid post_id type")
		}
		// 直接解析为 int64
		postID, err := strconv.ParseInt(postIDStr, 10, 64)
		if err != nil {
			logger.Error("Invalid post_id format", zap.String("post_id_str", postIDStr), zap.Error(err))
			return nil, errors.New("invalid post_id")
		}
		// 检查帖子是否存在
		var post models.Post
		if err := database.DB.First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("Post not found", zap.Int64("post_id", postID))
				return nil, errors.New("post not found")
			}
			logger.Error("Failed to find post", zap.Int64("post_id", postID), zap.Error(err))
			return nil, err
		}
		logger.Info("Post found", zap.Int64("post_id", postID), zap.String("title", post.Title))
		// 将 int64 转换为 uint 存储
		postIDUint := uint(postID)
		comment.PostID = &postIDUint
	}

	// 处理 CommentID（uint 类型）
	if req.CommentID != nil {
		var commentIDStr string
		switch v := req.CommentID.(type) {
		case string:
			commentIDStr = v
		case float64:
			commentIDStr = strconv.FormatInt(int64(v), 10)
		default:
			logger.Error("Invalid comment_id type", zap.Any("comment_id", req.CommentID))
			return nil, errors.New("invalid comment_id type")
		}
		commentID, err := strconv.ParseUint(commentIDStr, 10, 32)
		if err != nil {
			logger.Error("Invalid comment_id format", zap.String("comment_id_str", commentIDStr), zap.Error(err))
			return nil, errors.New("invalid comment_id")
		}
		// 检查评论是否存在
		var parentComment models.Comment
		if err := database.DB.First(&parentComment, commentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				logger.Error("Parent comment not found", zap.Uint64("comment_id", commentID))
				return nil, errors.New("parent comment not found")
			}
			logger.Error("Failed to find parent comment", zap.Uint64("comment_id", commentID), zap.Error(err))
			return nil, err
		}
		logger.Info("Parent comment found", zap.Uint64("comment_id", commentID))
		commentIDUint := uint(commentID)
		comment.CommentID = &commentIDUint

		// 从父评论中获取 PostID
		if parentComment.PostID != nil {
			comment.PostID = parentComment.PostID
			logger.Info("Set post_id from parent comment", zap.Uint("post_id", *parentComment.PostID))
		} else {
			logger.Error("Parent comment has no post_id", zap.Uint64("comment_id", commentID))
			return nil, errors.New("parent comment has no post_id")
		}
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		logger.Error("Failed to create comment", zap.Error(err))
		return nil, err
	}

	// 发送收信箱消息
	if req.CommentID != nil {
		// 回复评论：给被回复的评论作者发送消息
		var parentComment models.Comment
		if err := database.DB.First(&parentComment, comment.CommentID).Error; err == nil {
			if parentComment.UserID != userID {
				msg := database.InboxMessage{
					PostID:    int64(*comment.PostID),
					CommentID: comment.ID,
					SenderID:  userID,
					Type:      "reply_comment",
				}
				if err := database.PushInboxMessage(parentComment.UserID, msg); err != nil {
					logger.Error("Failed to push inbox message for reply_comment", zap.Error(err))
				}
			}
		}
	} else if req.PostID != nil {
		// 回复帖子：给帖子作者发送消息
		var post models.Post
		postID := int64(*comment.PostID)
		if err := database.DB.First(&post, postID).Error; err == nil {
			if post.UserID != userID {
				msg := database.InboxMessage{
					PostID:   postID,
					SenderID: userID,
					Type:     "reply_post",
				}
				if err := database.PushInboxMessage(post.UserID, msg); err != nil {
					logger.Error("Failed to push inbox message for reply_post", zap.Error(err))
				}
			}
		}
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

// AdminDeletePost 管理员删除帖子
func (s *PostService) AdminDeletePost(postID int64) error {
	var post models.Post
	if err := database.DB.First(&post, postID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrPostNotFound
		}
		return err
	}

	if err := database.DB.Delete(&post).Error; err != nil {
		logger.Error("Failed to delete post", zap.Error(err))
		return err
	}

	logger.Info("Post deleted by admin", zap.Int64("post_id", postID))
	return nil
}

// AdminDeleteComment 管理员删除评论
func (s *PostService) AdminDeleteComment(commentID uint) error {
	var comment models.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCommentNotFound
		}
		return err
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		logger.Error("Failed to delete comment", zap.Error(err))
		return err
	}

	logger.Info("Comment deleted by admin", zap.Uint("comment_id", commentID))
	return nil
}

// BanUser 禁言用户
func (s *PostService) BanUser(userID int64) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	user.Status = 0 // 0 表示禁用
	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to ban user", zap.Error(err))
		return err
	}

	logger.Info("User banned", zap.Int64("user_id", userID))
	return nil
}

// UnbanUser 解除禁言
func (s *PostService) UnbanUser(userID int64) error {
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	user.Status = 1 // 1 表示正常
	if err := database.DB.Save(&user).Error; err != nil {
		logger.Error("Failed to unban user", zap.Error(err))
		return err
	}

	logger.Info("User unbanned", zap.Int64("user_id", userID))
	return nil
}

func (s *PostService) GetComments(postID int64, page, pageSize int) ([]models.Comment, int64, error) {
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

// SearchComments 根据关键词搜索评论
func (s *PostService) SearchComments(keyword string, page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	// 如果关键词为空，返回空结果
	if keyword == "" {
		return []models.Comment{}, 0, nil
	}

	query := database.DB.Model(&models.Comment{}).Preload("User").Preload("Post").
		Where("content LIKE ?", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count comments", zap.Error(err))
		return nil, 0, err
	}

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		logger.Error("Failed to search comments", zap.Error(err))
		return nil, 0, err
	}

	return comments, total, nil
}

func (s *PostService) GetAllComments(page, pageSize int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	offset := (page - 1) * pageSize

	if err := database.DB.Model(&models.Comment{}).Count(&total).Error; err != nil {
		logger.Error("Failed to count comments", zap.Error(err))
		return nil, 0, err
	}

	if err := database.DB.Preload("User").Preload("Post").
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error; err != nil {
		logger.Error("Failed to get comments", zap.Error(err))
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

// IsPostLiked 检查用户是否点赞了帖子
func (s *PostService) IsPostLiked(userID int64, postID int64) (bool, error) {
	var like models.Like
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		logger.Error("Failed to check like status", zap.Error(err))
		return false, err
	}
	return true, nil
}

// GetPostLikeCount 获取帖子的点赞数量
func (s *PostService) GetPostLikeCount(postID int64) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count).Error
	if err != nil {
		logger.Error("Failed to get like count", zap.Error(err))
		return 0, err
	}
	return count, nil
}

// GetPostFavoriteInfo 检查用户是否收藏了帖子以及收藏在哪些文件夹中
func (s *PostService) GetPostFavoriteInfo(userID int64, postID int64) (bool, []models.FavoriteFolder, error) {
	var favorites []models.Favorite
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).Find(&favorites).Error
	if err != nil {
		logger.Error("Failed to check favorite status", zap.Error(err))
		return false, nil, err
	}

	if len(favorites) == 0 {
		return false, nil, nil
	}

	// 获取所有相关的收藏夹
	var folderIDs []uint
	for _, fav := range favorites {
		folderIDs = append(folderIDs, fav.FolderID)
	}

	var folders []models.FavoriteFolder
	if len(folderIDs) > 0 {
		err = database.DB.Where("id IN ?", folderIDs).Find(&folders).Error
		if err != nil {
			logger.Error("Failed to get folders", zap.Error(err))
			return true, nil, err
		}
	}

	return true, folders, nil
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

	// 检查是否已经被拉黑且未取消
	var existingBlock models.Block
	if err := database.DB.Where("user_id = ? AND blocked_id = ? AND unblocked_at IS NULL", userID, blockedID).First(&existingBlock).Error; err == nil {
		return ErrAlreadyBlocked
	}

	// 检查是否已经有记录但已取消拉黑
	var existingBlockRecord models.Block
	if err := database.DB.Where("user_id = ? AND blocked_id = ?", userID, blockedID).First(&existingBlockRecord).Error; err == nil {
		// 如果已经取消拉黑，重新拉黑
		if err := database.DB.Model(&existingBlockRecord).Update("unblocked_at", nil).Error; err != nil {
			logger.Error("Failed to re-block user", zap.Error(err))
			return err
		}
		logger.Info("User re-blocked", zap.Int64("user_id", userID), zap.Int64("blocked_id", blockedID))
		return nil
	}

	// 创建新的拉黑记录
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
	logger.Info("Unblock user request",
		zap.Int64("user_id", userID),
		zap.Int64("blocked_id", blockedID))

	var block models.Block
	// 查询未取消拉黑的记录
	if err := database.DB.Where("user_id = ? AND blocked_id = ? AND unblocked_at IS NULL", userID, blockedID).First(&block).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Info("Block record not found", zap.Int64("user_id", userID), zap.Int64("blocked_id", blockedID))
			return ErrNotBlocked
		}
		logger.Error("Failed to find block record", zap.Error(err))
		return err
	}

	if block.UnblockedAt != nil {
		logger.Info("Found block record", zap.Uint("block_id", block.ID), zap.Time("unblocked_at", *block.UnblockedAt))
	} else {
		logger.Info("Found block record", zap.Uint("block_id", block.ID), zap.String("unblocked_at", "NULL"))
	}

	now := time.Now()
	if err := database.DB.Model(&block).Update("unblocked_at", &now).Error; err != nil {
		logger.Error("Failed to unblock user", zap.Error(err))
		return err
	}

	logger.Info("User unblocked", zap.Int64("user_id", userID), zap.Int64("blocked_id", blockedID))
	return nil
}

func (s *PostService) IsBlocked(userID, blockedID int64) (bool, error) {
	var block models.Block
	err := database.DB.Where("user_id = ? AND blocked_id = ? AND unblocked_at IS NULL", userID, blockedID).First(&block).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PostService) GetBlockedUsers(userID int64, page, pageSize int) ([]BlockedUserWithStatus, int64, error) {
	var blockedUsers []BlockedUserWithStatus
	var total int64

	offset := (page - 1) * pageSize

	query := database.DB.Table("blocks").
		Select(`
			users.id,
			users.email,
			users.nickname,
			users.bio,
			users.avatar,
			users.is_admin,
			users.created_at,
			users.updated_at,
			users.deleted_at,
			blocks.created_at as blocked_at,
			blocks.unblocked_at
		`).
		Joins("JOIN users ON users.id = blocks.blocked_id").
		Where("blocks.user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset(offset).Limit(pageSize).Find(&blockedUsers).Error; err != nil {
		return nil, 0, err
	}

	return blockedUsers, total, nil
}

func (s *PostService) CreateFolder(userID int64, req CreateFolderRequest) (*models.FavoriteFolder, error) {
	logger.Info("Service create folder called", zap.Int64("user_id", userID), zap.String("folder_name", req.Name))

	var existingFolder models.FavoriteFolder
	if err := database.DB.Where("user_id = ? AND name = ?", userID, req.Name).First(&existingFolder).Error; err == nil {
		logger.Warn("Folder already exists", zap.Int64("user_id", userID), zap.String("folder_name", req.Name))
		return nil, ErrFolderExists
	}

	folder := models.FavoriteFolder{
		UserID: userID,
		Name:   req.Name,
	}

	if err := database.DB.Create(&folder).Error; err != nil {
		logger.Error("Failed to create folder in database", zap.Error(err), zap.Int64("user_id", userID), zap.String("folder_name", req.Name))
		return nil, err
	}

	logger.Info("Folder created successfully in database", zap.Uint("folder_id", folder.ID), zap.Int64("user_id", userID))
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
	// 将字符串 PostID 转换为 int64
	postID, err := strconv.ParseInt(req.PostID, 10, 64)
	if err != nil {
		return errors.New("invalid post_id")
	}

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
	if err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&existingFavorite).Error; err == nil {
		return ErrAlreadyFavorited
	}

	favorite := models.Favorite{
		UserID:   userID,
		PostID:   postID,
		FolderID: req.FolderID,
	}

	if err := database.DB.Create(&favorite).Error; err != nil {
		logger.Error("Failed to favorite post", zap.Error(err))
		return err
	}

	logger.Info("Post favorited", zap.Int64("post_id", postID), zap.Int64("user_id", userID), zap.Uint("folder_id", req.FolderID))
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

	findQuery := database.DB.Model(&models.Post{}).Preload("User").Where("user_id = ?", userID)

	if err := findQuery.
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (s *PostService) GetInbox(userID int64, page, pageSize int) ([]database.InboxMessage, int64, error) {
	return database.GetInboxMessages(userID, page, pageSize)
}

func (s *PostService) ClearInbox(userID int64) error {
	return database.ClearInbox(userID)
}
