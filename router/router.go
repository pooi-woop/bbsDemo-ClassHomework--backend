package router

import (
	"bbsDemo/config"
	"bbsDemo/handler"
	"bbsDemo/service"

	"github.com/gin-gonic/gin"
)

func InitRouter(userService *service.UserService, postService *service.PostService, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	authHandler := handler.NewAuthHandler(userService)
	postHandler := handler.NewPostHandler(postService)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.Static("/uploads", cfg.Upload.Path)

	auth := r.Group("/api/auth")
	{
		auth.POST("/send-code", authHandler.SendCode)
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
	}

	posts := r.Group("/api/posts")
	{
		posts.GET("", postHandler.ListPosts)
		posts.GET("/:id", postHandler.GetPost)
		posts.GET("/:id/comments", postHandler.GetComments)
	}

	comments := r.Group("/api/comments")
	{
		comments.GET("/:id/replies", postHandler.GetReplies)
	}

	authorized := r.Group("/api")
	authorized.Use(handler.AuthMiddleware())
	{
		authorized.GET("/profile", authHandler.GetProfile)
		authorized.POST("/logout", authHandler.Logout)
		authorized.POST("/logout-all", authHandler.LogoutAll)

		authorized.PUT("/profile/nickname", authHandler.UpdateNickname)
		authorized.POST("/profile/avatar", authHandler.UploadAvatar)

		authorized.GET("/my/posts", postHandler.GetMyPosts)
		authorized.GET("/my/favorites", postHandler.GetFavorites)
		authorized.GET("/my/blocked", postHandler.GetBlockedUsers)

		authorized.POST("/posts", postHandler.CreatePost)
		authorized.PUT("/posts/:id", postHandler.UpdatePost)
		authorized.DELETE("/posts/:id", postHandler.DeletePost)

		authorized.POST("/comments", postHandler.CreateComment)
		authorized.DELETE("/comments/:id", postHandler.DeleteComment)

		authorized.POST("/posts/:id/like", postHandler.LikePost)
		authorized.DELETE("/posts/:id/like", postHandler.UnlikePost)

		authorized.POST("/comments/:id/like", postHandler.LikeComment)
		authorized.DELETE("/comments/:id/like", postHandler.UnlikeComment)

		authorized.POST("/users/:id/block", postHandler.BlockUser)
		authorized.DELETE("/users/:id/block", postHandler.UnblockUser)

		authorized.POST("/favorites", postHandler.FavoritePost)
		authorized.DELETE("/posts/:id/favorite", postHandler.UnfavoritePost)
		authorized.PUT("/posts/:id/favorite", postHandler.MoveFavorite)

		authorized.GET("/folders", postHandler.GetFolders)
		authorized.POST("/folders", postHandler.CreateFolder)
		authorized.PUT("/folders/:id", postHandler.UpdateFolder)
		authorized.DELETE("/folders/:id", postHandler.DeleteFolder)
		authorized.GET("/folders/:id/posts", postHandler.GetFavoritesByFolder)
	}

	return r
}
