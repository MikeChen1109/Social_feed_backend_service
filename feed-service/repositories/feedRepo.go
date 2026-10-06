package repositories

import (
	"context"
	"feed-service/models"
	"time"

	"gorm.io/gorm"
)

type FeedRepositoryInterface interface {
	CreateFeed(feed *models.Feed) error
	GetFeeds() ([]models.Feed, error)
	PaginatedFeeds(offset int, limit int) (*models.PaginatedFeedsResponse, error)
	GetFeedByID(id uint) (*models.Feed, error)
	UpdateFeed(feed *models.Feed) error
	DeleteFeed(id uint) error
}

type FeedRepository struct {
	DB *gorm.DB
}

func (r *FeedRepository) CreateFeed(feed *models.Feed) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Create(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) GetFeeds() ([]models.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feeds []models.Feed
	if err := db.Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *FeedRepository) PaginatedFeeds(offset int, limit int) (*models.PaginatedFeedsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feeds []models.Feed
	err := db.Offset(offset).Limit(limit + 1).Order("created_at DESC, id DESC").Find(&feeds).Error
	if err != nil {
		return nil, err
	}

	hasMore := false
	if len(feeds) > limit {
		hasMore = true
		feeds = feeds[:limit] // 只取回前 limit 筆
	}

	var meta = models.Meta{HasMore: hasMore, Page: offset/limit + 1, Limit: limit}
	var response = models.PaginatedFeedsResponse{Data: feeds, Meta: meta}
	return &response, nil
}

func (r *FeedRepository) GetFeedByID(id uint) (*models.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	var feed models.Feed
	if err := db.First(&feed, id).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *FeedRepository) UpdateFeed(feed *models.Feed) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db := r.DB.WithContext(ctx)
	if err := db.Save(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) DeleteFeed(id uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
