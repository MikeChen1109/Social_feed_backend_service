package repositories

import (
	"myApp/SocialFeed/models"

	"gorm.io/gorm"
)

type FeedRepositoryInterface interface {
	CreateFeed(feed *models.Feed) error
	GetFeeds() ([]models.Feed, error)
	GetFeedByID(id uint) (*models.Feed, error)
	UpdateFeed(feed *models.Feed) error
	DeleteFeed(id uint) error
}

type FeedRepository struct {
	DB *gorm.DB
}

func (r *FeedRepository) CreateFeed(feed *models.Feed) error {
	if err := r.DB.Create(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) GetFeeds() ([]models.Feed, error) {
	var feeds []models.Feed
	if err := r.DB.Find(&feeds).Error; err != nil {
		return nil, err
	}
	return feeds, nil
}

func (r *FeedRepository) GetFeedByID(id uint) (*models.Feed, error) {
	var feed models.Feed
	if err := r.DB.First(&feed, id).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *FeedRepository) UpdateFeed(feed *models.Feed) error {
	if err := r.DB.Save(feed).Error; err != nil {
		return err
	}
	return nil
}

func (r *FeedRepository) DeleteFeed(id uint) error {
	var feed models.Feed
	if err := r.DB.First(&feed, id).Error; err != nil {
		return err
	}
	if err := r.DB.Delete(&feed).Error; err != nil {
		return err
	}
	return nil
}
