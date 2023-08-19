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

	// Admin pages
	engine.GET(adminRoute, router.HandleAdminDashboard)
	engine.GET(adminRoute+"/login", router.HandleAdminLogin)
	engine.GET(adminRoute+"/new-post", router.HandleAdminNewBlogPost)

	engine.POST(adminRoute+"/login", router.HandleAdminLoginRequest)
	engine.POST(adminRoute+"/new-post", router.HandleAdminNewBlogPostRequest)

	engine.GET("/hp", router.HandleHealth)

	return engine, nil
}
