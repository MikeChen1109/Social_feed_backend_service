package models

import "gorm.io/gorm"

type Feed struct {
	gorm.Model
	AuthorName string
	AuthorID   uint
	Upvotes    int32
	Title      string
	Content    string
	Comments   []Comment `gorm:"foreignKey:FeedID"`
}

type Comment struct {
	gorm.Model
	FeedID     uint
	AuthorName string
	AuthorID   uint
	Content    string
}
