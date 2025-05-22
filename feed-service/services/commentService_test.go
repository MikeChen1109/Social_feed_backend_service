package services

import (
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/* ---------- Mocks ---------- */

type mockCommentRepo struct{ mock.Mock }

func (m *mockCommentRepo) CreateComment(comment *models.Comment) error {
	args := m.Called(comment)
	return args.Error(0)
}

func (m *mockCommentRepo) PaginatedComments(offset int, limit int, feedId uint) (*models.PaginatedCommentsResponse, error) {
	args := m.Called(offset, limit, feedId)
	return args.Get(0).(*models.PaginatedCommentsResponse), args.Error(1)
}

/* ---------- Tests ---------- */

func TestCreateCommentSuccess(t *testing.T) {
	repo := new(mockCommentRepo)
	svc := &CommentService{CommentRepo: repo}

	userId := uint(100)
	username := "mike"
	feedId := uint(1)
	feedContent := "content"
	repo.On("CreateComment", mock.AnythingOfType("*models.Comment")).Return(nil)

	result, err := svc.CreateComment(feedId, feedContent, userId, username)

	assert.Nil(t, err)
	assert.Equal(t, feedId, result.FeedID)
	assert.Equal(t, username, result.AuthorName)
	assert.Equal(t, feedContent, result.Content)
	assert.Equal(t, userId, result.AuthorID)
	repo.AssertExpectations(t)
}

func TestCreateCommentWhenEmptyContent(t *testing.T) {
	svc := &CommentService{}
	result, err := svc.CreateComment(1, "", 1, "Mike")
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrCommentIvalidContentOrFeedId, err)
}

func TestPaginatedCommentsSuccess(t *testing.T) {
	repo := new(mockCommentRepo)
	svc := &CommentService{CommentRepo: repo}

	paginated := &models.PaginatedCommentsResponse{Data: []models.Comment{{AuthorName: "mikde"}}, Meta: models.Meta{Page: 1, Limit: 10, HasMore: false}}
	repo.On("PaginatedComments", 0, 10, uint(1)).Return(paginated, nil)

	resp, err := svc.PaginatedComments(0, 10, uint(1))
	assert.Nil(t, err)
	assert.Equal(t, len(paginated.Data), len(resp.Data))
	assert.Equal(t, false, resp.Meta.HasMore)
}
