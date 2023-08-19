package models

import (
	"blog.simoni.dev/auth"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"type:varchar(100);uniqueIndex"`
	Password string `gorm:"type:varchar(302)"`
	Admin    bool
}

func (u *User) IsAdmin() bool {
	return u.Admin
}

func (u *User) VerifyPassword(password string) (bool, error) {
	return auth.VerifyPassword(password, u.Password)
}
