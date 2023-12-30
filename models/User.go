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

func (u *User) NewAuthTokens(ctx *gin.Context) error {
	jwtToken, refreshToken, err := auth.GenerateTokens(&auth.JwtPayload{
		Username: u.Username,
		Admin:    u.Admin,
		UserId:   u.ID,
		Theme:    u.Theme,
	})
	if err != nil {
		return err
	}

	auth.AddAuthCookies(ctx, jwtToken, refreshToken)
	return nil
}
