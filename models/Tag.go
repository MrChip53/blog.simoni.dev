package models

import "gorm.io/gorm"

type Tag struct {
	gorm.Model
	Name string
}
