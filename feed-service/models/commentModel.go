package models

import "gorm.io/gorm"

type Comment struct {
	gorm.Model
	FeedID     uint
	AuthorName string
	AuthorID   uint
	Content    string
}

func (feed *Comment) ToCommentResponse() *CommentResponse {
	return &CommentResponse{
		ID:         feed.ID,
		Content:    feed.Content,
		AuthorID:   feed.AuthorID,
		AuthorName: feed.AuthorName,
		CreatedAt:  feed.CreatedAt,
		UpdatedAt:  feed.UpdatedAt,
		FeedID:     feed.FeedID,
	}
}
