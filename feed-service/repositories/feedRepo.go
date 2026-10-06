package repositories

import (
	"context"
	"feed-service/models"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type FeedRepositoryInterface interface {
	CreateFeed(ctx context.Context, feed *models.Feed) error
	GetFeeds(ctx context.Context) ([]models.Feed, error)
	PaginatedFeeds(ctx context.Context, cursor *models.Cursor, limit int) (*models.PaginatedFeedsResponse, error)
	GetFeedByID(ctx context.Context, id uint) (*models.Feed, error)
	UpdateFeed(ctx context.Context, feed *models.Feed) error
	DeleteFeed(ctx context.Context, id uint) error
}

type FeedRepository struct {
	DB *gorm.DB
}

func (r *FeedRepository) CreateFeed(ctx context.Context, feed *models.Feed) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Create(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) GetFeeds(ctx context.Context) ([]models.Feed, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feeds []models.Feed
	if err := db.Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *FeedRepository) PaginatedFeeds(ctx context.Context, cursor *models.Cursor, limit int) (*models.PaginatedFeedsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feeds []models.Feed
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("invalid page limit")
	}
	if cursor != nil {
		db = db.Where("(created_at < ? OR (created_at = ? AND id < ?))", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	err := db.Limit(limit + 1).Order("created_at DESC, id DESC").Find(&feeds).Error
	if err != nil {
		return nil, err
	}

	hasMore := false
	if len(feeds) > limit {
		hasMore = true
		feeds = feeds[:limit] // 只取回前 limit 筆
	}

	var meta = models.Meta{HasMore: hasMore, Limit: limit}
	if hasMore {
		last := feeds[len(feeds)-1]
		meta.NextCursor = models.EncodeCursor("feeds", last.CreatedAt, last.ID)
	}
	var response = models.PaginatedFeedsResponse{Data: feeds, Meta: meta}
	return &response, nil
}

func (r *FeedRepository) GetFeedByID(ctx context.Context, id uint) (*models.Feed, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feed models.Feed
	if err := db.First(&feed, id).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *FeedRepository) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Save(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) DeleteFeed(ctx context.Context, id uint) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feed models.Feed
	if err := db.First(&feed, id).Error; err != nil {
		return err
	}
	if err := db.Delete(&feed).Error; err != nil {
		return err
	}
	return nil
}
