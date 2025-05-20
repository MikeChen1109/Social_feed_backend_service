package services

import (
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"feed-service/repositories"
)

type FeedService struct {
	FeedRepo repositories.FeedRepositoryInterface
}

func (s *FeedService) CreateFeed(title string, content string, userId uint, userName string) (*models.Feed, *appErrors.AppError) {
	if title == "" || content == "" {
		return nil, appErrors.ErrFeedInvalidContentOrTitle
	}

	feed := models.Feed{AuthorName: userName,
		AuthorID: userId,
		Title:    title,
		Content:  content}

	err := s.FeedRepo.CreateFeed(&feed)
	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return &feed, nil
}

func (s *FeedService) GetFeeds() ([]models.Feed, *appErrors.AppError) {
	var feeds []models.Feed
	feeds, err := s.FeedRepo.GetFeeds()

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return feeds, nil
}

func (s *FeedService) PaginatedFeeds(offset int, limit int) (*models.PaginatedFeedsResponse, *appErrors.AppError) {
	response, err := s.FeedRepo.PaginatedFeeds(offset, limit)

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return response, nil
}

func (s *FeedService) GetFeedByID(id uint) (*models.Feed, *appErrors.AppError) {
	feed, err := s.FeedRepo.GetFeedByID(id)
	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return feed, nil
}

func (s *FeedService) UpdateFeed(id uint, title string, content string) *appErrors.AppError {
	if title == "" || content == "" {
		return appErrors.ErrFeedInvalidContentOrTitle
	}

	feed, err := s.FeedRepo.GetFeedByID(id)
	if err != nil {
		return appErrors.DatabaseError
	}

	feed.Title = title
	feed.Content = content
	err = s.FeedRepo.UpdateFeed(feed)
	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (s *FeedService) DeleteFeed(id uint) *appErrors.AppError {
	err := s.FeedRepo.DeleteFeed(id)
	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}
