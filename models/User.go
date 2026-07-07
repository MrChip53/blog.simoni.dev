package models

import (
	"time"

	"blog.simoni.dev/auth"
	"github.com/gin-gonic/gin"
)

type User struct {
	ID        int64
	CreatedAt time.Time
	Username  string
	Password  string
	Admin     bool
	Theme     string
}

func (u *User) IsAdmin() bool {
	return u.Admin
}

func (u *User) VerifyPassword(password string) (bool, error) {
	return auth.VerifyPassword(password, u.Password)
}

func (u *User) NewAuthTokens(ctx *gin.Context) (*auth.JwtPayload, error) {
	payload := &auth.JwtPayload{
		Username: u.Username,
		Admin:    u.Admin,
		UserId:   uint(u.ID),
		Theme:    u.Theme,
	}

	jwtToken, refreshToken, err := auth.GenerateTokens(payload)
	if err != nil {
		return nil, err
	}

	auth.AddAuthCookies(ctx, jwtToken, refreshToken)
	return payload, nil
}
