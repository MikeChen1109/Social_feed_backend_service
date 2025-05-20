package models

import "gorm.io/gorm"

type Feed struct {
	gorm.Model
	AuthorName    string
	AuthorID      uint
	Upvotes       int `gorm:"default:0"`
	Title         string
	Content       string
	CommentsCount int `gorm:"default:0"`
}
