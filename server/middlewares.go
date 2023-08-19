package server

import (
	"blog.simoni.dev/auth"
	"github.com/gin-gonic/gin"
	"log"
)

func IsHXRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("isHXRequest", ctx.Request.Header.Get("HX-Request") == "true")
		ctx.Next()
	}
}

func ExtractAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authToken, err := auth.ExtractAuth(ctx)
		if err != nil {
			log.Printf("Failed to extract auth: %v\n", err)
			if ctx.Request.URL.Path[:6] == "/admin" && ctx.Request.URL.Path != "/admin/login" {
				ctx.Redirect(302, "/admin/login?redirect="+ctx.Request.URL.Path)
				ctx.Abort()
			}
			return
		}

		ctx.Set("authToken", authToken)
		ctx.Next()
	}
}
