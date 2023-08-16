package models

import "gorm.io/gorm"

type BlogPost struct {
	gorm.Model
	Title   string
	Author  string
	Slug    string `gorm:"type:varchar(100);unique_index"`
	Content string `gorm:"type:text"`
	Tags    []Tag  `gorm:"many2many:blog_post_tags;"`
}
