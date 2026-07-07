package models

import (
	"fmt"
	"time"
)

type Comment struct {
	ID         int64
	CreatedAt  time.Time
	BlogPostId int64
	Author     string
	Comment    string
}

func (c *Comment) GetHtmlId() string {
	return fmt.Sprintf("comment-%d", c.ID)
}
