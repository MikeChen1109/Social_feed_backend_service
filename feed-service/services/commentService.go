package services

import (
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"feed-service/repositories"
)

type CommentService struct {
	CommentRepo repositories.CommentRepositoryInterface
}

func (s *CommentService) CreateComment(feedId uint, content string, userId uint, userName string) (*models.Comment, *appErrors.AppError) {
	comment := &models.Comment{FeedID: feedId, Content: content, AuthorName: userName, AuthorID: userId}

	err := s.CommentRepo.CreateComment(comment)
	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return comment, nil
}

func (s *CommentService) PaginatedComments(offset int, limit int, feedId uint) (*models.PaginatedCommentsResponse, *appErrors.AppError) {
	response, err := s.CommentRepo.PaginatedComments(offset, limit, feedId)

	if err != nil {
		return nil, appErrors.DatabaseError
	}

	return response, nil
}
