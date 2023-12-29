package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	BlogPostId uint
	Author     string
	Comment    string `gorm:"type:varchar(1000)"`
}
