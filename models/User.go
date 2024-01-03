package models

import (
	"blog.simoni.dev/auth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex"`
	Password string `gorm:"type:varchar(302)"`
	Admin    bool
	Theme    string `gorm:"type:varchar(100);default:dark"`
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
		UserId:   u.ID,
		Theme:    u.Theme,
	}

	jwtToken, refreshToken, err := auth.GenerateTokens(payload)
	if err != nil {
		return nil, err
	}

	auth.AddAuthCookies(ctx, jwtToken, refreshToken)
	return payload, nil
}
