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

func AddJwtPayloadToCtx(ctx *gin.Context, jwtPayload *auth.JwtPayload) {
	ctx.Set("authToken", jwtPayload)
	ctx.Set("authed", true)
	ctx.Set("theme", jwtPayload.Theme)
	ctx.Set("isAdmin", jwtPayload.Admin)
	ctx.Set("username", jwtPayload.Username)
	ctx.Set("userId", jwtPayload.UserId)
}

func ExtractAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authToken, err := auth.ExtractAuth(ctx)
		if err != nil {
			log.Printf("Failed to extract auth: %v\n", err)
			pathLength := len(ctx.Request.URL.Path)
			if pathLength >= 6 && ctx.Request.URL.Path[:6] == "/admin" {
				ctx.Redirect(302, "/login?redirect="+ctx.Request.URL.Path)
				ctx.Abort()
			}
			return
		}

		AddJwtPayloadToCtx(ctx, authToken)

		if ctx.Request.URL.Path == "/login" {
			ctx.Redirect(302, "/")
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
