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
	engine.HTMLRender = createRenderer("./templates")

	engine.NoRoute(router.HandleNotFound)

	// Regular pages
	engine.GET("/", router.HandleIndex)
	engine.GET("/post/:month/:day/:year/:slug", router.HandlePost)
	engine.GET("/tag/:tag", router.HandleTag)
	engine.GET("/user/:username", router.HandleUser)

	// Admin pages
	engine.GET(adminRoute, router.HandleAdminDashboard)
	engine.GET(adminRoute+"/login", router.HandleAdminLogin)
	engine.GET(adminRoute+"/new-post", router.HandleAdminNewBlogPost)
	engine.GET(adminRoute+"/posts", router.HandleAdminPosts)

	engine.POST(adminRoute+"/login", router.HandleAdminLoginRequest)
	engine.POST(adminRoute+"/new-post", router.HandleAdminNewBlogPostRequest)

	// Authed utility endpoints
	engine.POST(adminRoute+"/generate-markdown", router.HandleAdminGenerateMarkdown)

	engine.DELETE(adminRoute+"/post/:id", router.HandleAdminPostsDelete)

	engine.GET("/hp", router.HandleHealth)

	return engine, nil
}
