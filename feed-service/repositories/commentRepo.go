package repositories

import (
	"context"
	"feed-service/models"
	"time"

	"gorm.io/gorm"
)

type CommentRepositoryInterface interface {
	CreateComment(comment *models.Comment) error
	PaginatedComments(offset int, limit int, feedId uint) (*models.PaginatedCommentsResponse, error)
}

type CommentRepository struct {
	DB *gorm.DB
}

func (r *CommentRepository) CreateComment(comment *models.Comment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Create(comment).Error; err != nil {
		return err
	}
	return nil
}

func (r *CommentRepository) PaginatedComments(offset int, limit int, feedId uint) (*models.PaginatedCommentsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var comments []models.Comment
	err := db.
		Where("feed_id = ?", feedId).
		Offset(offset).Limit(limit + 1).
		Order("created_at DESC, id DESC").
		Find(&comments).Error

	if err != nil {
		return nil, err
	}

	hasMore := false
	if len(comments) > limit {
		hasMore = true
		comments = comments[:limit] // 只取回前 limit 筆
	}

	var meta = models.Meta{HasMore: hasMore, Page: offset/limit + 1, Limit: limit}
	var response = models.PaginatedCommentsResponse{Data: comments, Meta: meta}
	return &response, nil
}
