package services

import (
	"context"
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"feed-service/repositories"
)

type CommentService struct {
	CommentRepo repositories.CommentRepositoryInterface
}

func (s *CommentService) CreateComment(ctx context.Context, feedId uint, content string, userId uint, userName string) (*models.CommentResponse, *appErrors.AppError) {
	if content == "" || feedId == uint(0) {
		return nil, appErrors.ErrCommentIvalidContentOrFeedId
	}

	comment := &models.Comment{FeedID: feedId, Content: content, AuthorName: userName, AuthorID: userId}

	err := s.CommentRepo.CreateComment(ctx, comment)
	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return comment.ToCommentResponse(), nil
}

func (s *CommentService) PaginatedComments(ctx context.Context, cursor *models.Cursor, limit int, feedId uint) (*models.PaginatedCommentsResponse, *appErrors.AppError) {
	response, err := s.CommentRepo.PaginatedComments(ctx, cursor, limit, feedId)

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return response, nil
}
