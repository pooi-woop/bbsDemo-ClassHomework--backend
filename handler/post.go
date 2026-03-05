package handler

import (
	"bbsDemo/logger"
	"bbsDemo/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{postService: postService}
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req service.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.postService.CreatePost(userID.(int64), req)
	if err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"post": post})
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req service.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post, err := h.postService.UpdatePost(userID.(int64), postID, req)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		default:
			logger.Error("Failed to update post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}

func (h *PostHandler) DeletePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.DeletePost(userID.(int64), postID); err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		default:
			logger.Error("Failed to delete post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

func (h *PostHandler) GetPost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	post, err := h.postService.GetPost(postID)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		default:
			logger.Error("Failed to get post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}

func (h *PostHandler) ListPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.postService.ListPosts(page, pageSize)
	if err != nil {
		logger.Error("Failed to list posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) SearchPosts(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "keyword is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.postService.SearchPosts(keyword, page, pageSize)
	if err != nil {
		logger.Error("Failed to search posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"keyword":   keyword,
	})
}

func (h *PostHandler) CreateComment(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment, err := h.postService.CreateComment(userID.(int64), req)
	if err != nil {
		logger.Error("Failed to create comment", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create comment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"comment": comment})
}

func (h *PostHandler) DeleteComment(c *gin.Context) {
	userID, _ := c.Get("userID")

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := h.postService.DeleteComment(userID.(int64), uint(commentID)); err != nil {
		switch err {
		case service.ErrCommentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		default:
			logger.Error("Failed to delete comment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

// AdminDeletePost 管理员删除帖子
func (h *PostHandler) AdminDeletePost(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.AdminDeletePost(postID); err != nil {
		switch err {
		case service.ErrPostNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		default:
			logger.Error("Failed to delete post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted"})
}

// AdminDeleteComment 管理员删除评论
func (h *PostHandler) AdminDeleteComment(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := h.postService.AdminDeleteComment(uint(commentID)); err != nil {
		switch err {
		case service.ErrCommentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found"})
		default:
			logger.Error("Failed to delete comment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted"})
}

// BanUser 禁言用户
func (h *PostHandler) BanUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.postService.BanUser(userID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			logger.Error("Failed to ban user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ban user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User banned"})
}

// UnbanUser 解除禁言
func (h *PostHandler) UnbanUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.postService.UnbanUser(userID); err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			logger.Error("Failed to unban user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unban user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unbanned"})
}

func (h *PostHandler) GetComments(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	comments, total, err := h.postService.GetComments(uint(postID), page, pageSize)
	if err != nil {
		logger.Error("Failed to get comments", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get comments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"comments":  comments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) GetReplies(c *gin.Context) {
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	comments, total, err := h.postService.GetReplies(uint(commentID), page, pageSize)
	if err != nil {
		logger.Error("Failed to get replies", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get replies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"replies":   comments,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) LikePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.LikePost(userID.(int64), postID); err != nil {
		switch err {
		case service.ErrAlreadyLiked:
			c.JSON(http.StatusConflict, gin.H{"error": "Already liked"})
		default:
			logger.Error("Failed to like post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post liked"})
}

func (h *PostHandler) UnlikePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.UnlikePost(userID.(int64), postID); err != nil {
		switch err {
		case service.ErrNotLiked:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not liked"})
		default:
			logger.Error("Failed to unlike post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post unliked"})
}

func (h *PostHandler) LikeComment(c *gin.Context) {
	userID, _ := c.Get("userID")

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := h.postService.LikeComment(userID.(int64), uint(commentID)); err != nil {
		switch err {
		case service.ErrAlreadyLiked:
			c.JSON(http.StatusConflict, gin.H{"error": "Already liked"})
		default:
			logger.Error("Failed to like comment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to like comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment liked"})
}

func (h *PostHandler) UnlikeComment(c *gin.Context) {
	userID, _ := c.Get("userID")

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	if err := h.postService.UnlikeComment(userID.(int64), uint(commentID)); err != nil {
		switch err {
		case service.ErrNotLiked:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not liked"})
		default:
			logger.Error("Failed to unlike comment", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unlike comment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment unliked"})
}

func (h *PostHandler) BlockUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	blockedID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.postService.BlockUser(userID.(int64), blockedID); err != nil {
		switch err {
		case service.ErrAlreadyBlocked:
			c.JSON(http.StatusConflict, gin.H{"error": "Already blocked"})
		case service.ErrCannotBlockSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot block yourself"})
		default:
			logger.Error("Failed to block user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User blocked"})
}

func (h *PostHandler) UnblockUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	blockedID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.postService.UnblockUser(userID.(int64), blockedID); err != nil {
		switch err {
		case service.ErrNotBlocked:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not blocked"})
		default:
			logger.Error("Failed to unblock user", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User unblocked"})
}

func (h *PostHandler) GetBlockedUsers(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.postService.GetBlockedUsers(userID.(int64), page, pageSize)
	if err != nil {
		logger.Error("Failed to get blocked users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get blocked users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) CreateFolder(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req service.CreateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.postService.CreateFolder(userID.(int64), req)
	if err != nil {
		switch err {
		case service.ErrFolderExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Folder already exists"})
		default:
			logger.Error("Failed to create folder", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"folder": folder})
}

func (h *PostHandler) UpdateFolder(c *gin.Context) {
	userID, _ := c.Get("userID")

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	var req service.UpdateFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	folder, err := h.postService.UpdateFolder(userID.(int64), uint(folderID), req)
	if err != nil {
		switch err {
		case service.ErrFolderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		case service.ErrFolderExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Folder name already exists"})
		case service.ErrFolderNotYours:
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not yours"})
		default:
			logger.Error("Failed to update folder", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update folder"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"folder": folder})
}

func (h *PostHandler) DeleteFolder(c *gin.Context) {
	userID, _ := c.Get("userID")

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	if err := h.postService.DeleteFolder(userID.(int64), uint(folderID)); err != nil {
		switch err {
		case service.ErrFolderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		case service.ErrDefaultFolder:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default folder"})
		case service.ErrFolderNotYours:
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not yours"})
		default:
			logger.Error("Failed to delete folder", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted"})
}

func (h *PostHandler) GetFolders(c *gin.Context) {
	userID, _ := c.Get("userID")

	folders, err := h.postService.GetFolders(userID.(int64))
	if err != nil {
		logger.Error("Failed to get folders", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get folders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

func (h *PostHandler) FavoritePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req service.FavoritePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.postService.FavoritePost(userID.(int64), req); err != nil {
		switch err {
		case service.ErrAlreadyFavorited:
			c.JSON(http.StatusConflict, gin.H{"error": "Already favorited"})
		case service.ErrFolderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		case service.ErrFolderNotYours:
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not yours"})
		default:
			logger.Error("Failed to favorite post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to favorite post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post favorited"})
}

func (h *PostHandler) UnfavoritePost(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := h.postService.UnfavoritePost(userID.(int64), postID); err != nil {
		switch err {
		case service.ErrNotFavorited:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not favorited"})
		default:
			logger.Error("Failed to unfavorite post", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfavorite post"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post unfavorited"})
}

func (h *PostHandler) MoveFavorite(c *gin.Context) {
	userID, _ := c.Get("userID")

	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var req struct {
		FolderID uint `json:"folder_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.postService.MoveFavorite(userID.(int64), postID, req.FolderID); err != nil {
		switch err {
		case service.ErrNotFavorited:
			c.JSON(http.StatusNotFound, gin.H{"error": "Not favorited"})
		case service.ErrFolderNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		case service.ErrFolderNotYours:
			c.JSON(http.StatusForbidden, gin.H{"error": "Folder not yours"})
		default:
			logger.Error("Failed to move favorite", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move favorite"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Favorite moved"})
}

func (h *PostHandler) GetFavorites(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	favorites, total, err := h.postService.GetFavorites(userID.(int64), page, pageSize)
	if err != nil {
		logger.Error("Failed to get favorites", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"favorites": favorites,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) GetFavoritesByFolder(c *gin.Context) {
	userID, _ := c.Get("userID")

	folderID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.postService.GetFavoritesByFolder(userID.(int64), uint(folderID), page, pageSize)
	if err != nil {
		logger.Error("Failed to get favorites by folder", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favorites"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (h *PostHandler) GetMyPosts(c *gin.Context) {
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	posts, total, err := h.postService.GetMyPosts(userID.(int64), page, pageSize)
	if err != nil {
		logger.Error("Failed to get my posts", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get my posts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts":     posts,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
