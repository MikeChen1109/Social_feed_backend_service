package services

import (
	appErrors "myApp/SocialFeed/common/appErrors"
	"myApp/SocialFeed/models"
	"myApp/SocialFeed/repositories"
)

type FeedService struct {
	FeedRepo repositories.FeedRepositoryInterface
}

func (s *FeedService) CreateFeed(title string, content string, user models.User) (*models.Feed, *appErrors.AppError) {
	if title == "" || content == "" {
		return nil, appErrors.ErrFeedInvalidContentOrTitle
	}

	feed := models.Feed{AuthorName: user.Username,
		AuthorID: user.ID,
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
