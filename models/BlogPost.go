package models

import (
	"fmt"
	"time"
)

type BlogPost struct {
	ID          int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Author      string
	Slug        string
	Content     string
	Description string
	Tags        []Tag
	Draft       bool
	PublishedAt *time.Time
}

func (p *BlogPost) GetEditLink(adminRoute string) string {
	return fmt.Sprintf("%s/edit/%d", adminRoute, p.ID)
}

func (p *BlogPost) GetCommentPostLink() string {
	return fmt.Sprintf("/comment/%d", p.ID)
}

func (p *BlogPost) GetCommentsHtmlId() string {
	return fmt.Sprintf("comments-%d", p.ID)
}

func (p *BlogPost) Publish() {
	now := time.Now()
	p.PublishedAt = &now
}
