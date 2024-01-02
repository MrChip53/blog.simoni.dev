package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	BlogPostId uint
	Author     string
	Comment    string `gorm:"type:varchar(1000)"`
}

func (c *Comment) GetHtmlId() string {
	return fmt.Sprintf("comment-%d", c.ID)
}
