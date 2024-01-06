package models

import (
	"fmt"
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Name string
}

func NewTag(db *gorm.DB, name string) (newTag *Tag, err error) {
	newTag = &Tag{
		Name: name,
	}

	if err = db.Create(newTag).Error; err != nil {
		return nil, err
	}

	return newTag, nil
}

func GetTag(db *gorm.DB, name string) (tag *Tag, err error) {
	err = db.First(&tag, "name = ?", name).Error

	return tag, err
}

func (t *Tag) GetLink() string {
	return "/tag/" + t.Name
}

func (t *Tag) GetDeleteLink(postId uint) string {
	return fmt.Sprintf("/admin/post/%d/tag/%d", postId, t.ID)
}
