package repositories

import (
	"context"
	"feed-service/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type CommentRepositoryInterface interface {
	CreateComment(ctx context.Context, comment *models.Comment) error
	PaginatedComments(ctx context.Context, cursor *models.Cursor, limit int, feedId uint) (*models.PaginatedCommentsResponse, error)
}

type CommentRepository struct {
	DB *gorm.DB
}

func (r *CommentRepository) CreateComment(ctx context.Context, comment *models.Comment) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Create(comment).Error; err != nil {
		return err
	}
	return nil
}

func (r *CommentRepository) PaginatedComments(ctx context.Context, cursor *models.Cursor, limit int, feedId uint) (*models.PaginatedCommentsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var comments []models.Comment
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("invalid page limit")
	}
	if cursor != nil {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	err := db.
		Where("feed_id = ?", feedId).
		Limit(limit + 1).
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

	var meta = models.Meta{HasMore: hasMore, Limit: limit}
	if hasMore {
		last := comments[len(comments)-1]
		meta.NextCursor = models.EncodeCursor(models.CommentCursorScope(feedId), last.CreatedAt, last.ID)
	}
	var response = models.PaginatedCommentsResponse{Data: comments, Meta: meta}
	return &response, nil
}
