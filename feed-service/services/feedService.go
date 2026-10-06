package services

import (
	"context"
	"errors"
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"feed-service/repositories"
	"gorm.io/gorm"
)

type FeedService struct {
	FeedRepo repositories.FeedRepositoryInterface
}

func (s *FeedService) CreateFeed(ctx context.Context, title string, content string, userId uint, userName string) (*models.FeedResponse, *appErrors.AppError) {
	if title == "" || content == "" {
		return nil, appErrors.ErrFeedInvalidContentOrTitle
	}

	feed := models.Feed{AuthorName: userName,
		AuthorID: userId,
		Title:    title,
		Content:  content}

	err := s.FeedRepo.CreateFeed(ctx, &feed)
	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return feed.ToFeedResponse(), nil
}

func (s *FeedService) GetFeeds(ctx context.Context) ([]models.FeedResponse, *appErrors.AppError) {
	var feeds []models.Feed
	feeds, err := s.FeedRepo.GetFeeds(ctx)

	responses := make([]models.FeedResponse, len(feeds))
	for i, f := range feeds {
		responses[i] = *f.ToFeedResponse()
	}

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return responses, nil
}

func (s *FeedService) PaginatedFeeds(ctx context.Context, cursor *models.Cursor, limit int) (*models.PaginatedFeedsResponse, *appErrors.AppError) {
	response, err := s.FeedRepo.PaginatedFeeds(ctx, cursor, limit)

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return response, nil
}

func (s *FeedService) GetFeedByID(ctx context.Context, id uint) (*models.FeedResponse, *appErrors.AppError) {
	feed, err := s.FeedRepo.GetFeedByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, appErrors.ErrFeedNotFound
		}
		return nil, appErrors.DatabaseError
	}

	return feed.ToFeedResponse(), nil
}

func (s *FeedService) UpdateFeed(ctx context.Context, id uint, userID uint, title string, content string) *appErrors.AppError {
	if title == "" || content == "" {
		return appErrors.ErrFeedInvalidContentOrTitle
	}

	feed, err := s.FeedRepo.GetFeedByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrFeedNotFound
		}
		return appErrors.DatabaseError
	}

	if feed.AuthorID != userID {
		return appErrors.ErrForbidden
	}
	feed.Title = title
	feed.Content = content
	err = s.FeedRepo.UpdateFeed(ctx, feed)
	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}

func (s *FeedService) DeleteFeed(ctx context.Context, id uint, userID uint) *appErrors.AppError {
	feed, err := s.FeedRepo.GetFeedByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return appErrors.ErrFeedNotFound
		}
		return appErrors.DatabaseError
	}
	if feed.AuthorID != userID {
		return appErrors.ErrForbidden
	}
	err = s.FeedRepo.DeleteFeed(ctx, id)
	if err != nil {
		return appErrors.DatabaseError
	}

	return nil
}
