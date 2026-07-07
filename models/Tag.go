package models

import (
	"fmt"
	"time"
)

type Tag struct {
	ID        int64
	CreatedAt time.Time
	Name      string
}

func (t *Tag) GetLink() string {
	return "/tag/" + t.Name
}

func (t *Tag) GetDeleteLink(postId int64) string {
	return fmt.Sprintf("/admin/post/%d/tag/%d", postId, t.ID)
}
