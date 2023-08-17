package server

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
)

func NewServer(db *gorm.DB) (*gin.Engine, error) {
	router := NewRouter(db)

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = log.Writer()

	engine := gin.Default()

	if err := engine.SetTrustedProxies(nil); err != nil {
		return nil, err
	}

	engine.Static("/css", "css")
	engine.HTMLRender = createRenderer("./templates")

	engine.NoRoute(router.HandleNotFound)

	engine.GET("/", router.HandleIndex)
	engine.GET("/post/:month/:day/:year/:slug", router.HandlePost)
	engine.GET("/tag/:tag", router.HandleTag)
	engine.GET("/hp", router.HandleHealth)

	return engine, nil
}
