package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	FeedID     uint
	AuthorName string
	AuthorID   uint
	Content    string
}
