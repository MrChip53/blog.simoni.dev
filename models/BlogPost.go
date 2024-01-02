package models

import (
	"fmt"
	"gorm.io/gorm"
)

type BlogPost struct {
	gorm.Model
	Title       string
	Author      string
	Slug        string `gorm:"type:varchar(100);unique_index"`
	Content     string `gorm:"type:text"`
	Description string `gorm:"type:varchar(100)"`
	Tags        []Tag  `gorm:"many2many:blog_post_tags;"`
}

func NewBlogPost(db *gorm.DB, title string, author string, slug string, content string, description string) (newPost *BlogPost, err error) {
	newPost = &BlogPost{
		Title:       title,
		Slug:        slug,
		Content:     content,
		Author:      author,
		Description: description,
	}

	if err := db.Create(newPost).Error; err != nil {
		return nil, err
	}

	return newPost, nil
}

func (p *BlogPost) AddTag(tag *Tag) {
	p.Tags = append(p.Tags, *tag)
}

func (p *BlogPost) UpdateTags(tx *gorm.DB) error {
	err := tx.Model(p).Association("Tags").Replace(p.Tags)
	if err != nil {
		return err
	}

	return nil
}

func (p *BlogPost) GetCommentPostLink() string {
	return fmt.Sprintf("/comment/%d", p.ID)
}

func (p *BlogPost) GetCommentsHtmlId() string {
	return fmt.Sprintf("comments-%d", p.ID)
}
