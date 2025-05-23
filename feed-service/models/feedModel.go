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

func (feed *Feed) ToFeedResponse() *FeedResponse {
	return &FeedResponse{
		ID:           feed.ID,
		Title:        feed.Title,
		Content:      feed.Content,
		AuthorID:     feed.AuthorID,
		AuthorName:   feed.AuthorName,
		Upvotes:      feed.Upvotes,
		CommentsCount: feed.CommentsCount,
		CreatedAt:    feed.CreatedAt,
		UpdatedAt:    feed.UpdatedAt,
	}
}
