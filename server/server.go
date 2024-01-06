package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

var adminRoute = "/admin"

func NewServer(db *gorm.DB) (*gin.Engine, error) {
	router := NewRouter(db)

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = log.Writer()

	engine := gin.Default()

	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	engine.Use(IsHXRequest())
	engine.Use(ExtractAuth())

	engine.Static("/css", "css")
	engine.Static("/js", "js")

	engine.NoRoute(router.HandleNotFound)

	// Regular pages
	engine.GET("/", router.HandleIndex)
	engine.GET("/post/:month/:day/:year/:slug", router.HandlePost)
	engine.GET("/tag/:tag", router.HandleTag)
	engine.GET("/user/:username", router.HandleUser)
	engine.GET("/settings", router.HandleSettings)
	engine.GET("/login", router.HandleLogin)

	engine.POST("/comment/:postId", router.HandleComment)

	engine.POST("/user/username", router.HandleUsernameChange)
	engine.POST("/user/password", router.HandlePasswordChange)

	engine.POST("/login", router.HandleLoginRequest)
	engine.GET("/logout", router.HandleLogoutRequest)

	// Admin pages
	engine.GET(adminRoute, router.HandleAdminDashboard)
	engine.GET(adminRoute+"/new-post", router.HandleAdminNewBlogPost)
	engine.GET(adminRoute+"/posts", router.HandleAdminPosts)
	engine.GET(adminRoute+"/edit/:postId", router.HandlePostEdit)

	engine.POST(adminRoute+"/edit/:postId", router.PostPostEdit)
	engine.POST(adminRoute+"/new-post", router.HandleAdminNewBlogPostRequest)

	// Authed utility endpoints
	engine.POST(adminRoute+"/generate-markdown", router.HandleAdminGenerateMarkdown)
	engine.POST(adminRoute+"/post/:id/tag", router.HandleAdminAddTagToPost)

	engine.DELETE(adminRoute+"/post/:id", router.HandleAdminPostsDelete)
	engine.DELETE(adminRoute+"/post/:id/tag/:tagId", router.HandleAdminDeleteTagFromPost)

	engine.GET("/hp", router.HandleHealth)

	return engine, nil
}
