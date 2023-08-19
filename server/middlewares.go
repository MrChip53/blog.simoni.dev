package server

import (
	"github.com/gin-gonic/gin"
)

func IsHXRequest() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("isHXRequest", ctx.Request.Header.Get("HX-Request") == "true")
		ctx.Next()
	}
}

func ExtractAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || authHeader[:7] != "Bearer " {
			ctx.Set("hasAuth", false)
			path := ctx.Request.URL.Path
			adminRouteLen := len(adminRoute)

			equalLengths := len(path) == adminRouteLen

			if equalLengths {
				pathIsAdmin := path[:adminRouteLen] == adminRoute
				pathIsNotLogin := path != adminRoute+"/login"

				if pathIsAdmin && pathIsNotLogin {
					HandleAdminLogin(ctx)
					ctx.Abort()
				}
			} else {
				ctx.Next()
			}
			return
		}

		ctx.Set("authToken", authHeader[7:])

		// TODO decode token and add to context
		// TODO verify token is valid
	}
}
