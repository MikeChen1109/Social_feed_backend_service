package services

import (
	"context"
	"errors"
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/* ---------- Mocks ---------- */

type mockFeedRepo struct{ mock.Mock }

func (m *mockFeedRepo) CreateFeed(ctx context.Context, feed *models.Feed) error {
	args := m.Called(feed)
	return args.Error(0)
}

func (m *mockFeedRepo) GetFeeds(ctx context.Context) ([]models.Feed, error) {
	args := m.Called()
	return args.Get(0).([]models.Feed), args.Error(1)
}

func (m *mockFeedRepo) PaginatedFeeds(ctx context.Context, cursor *models.Cursor, limit int) (*models.PaginatedFeedsResponse, error) {
	args := m.Called(cursor, limit)
	return args.Get(0).(*models.PaginatedFeedsResponse), args.Error(1)
}

func (m *mockFeedRepo) GetFeedByID(ctx context.Context, id uint) (*models.Feed, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Feed), args.Error(1)
}

func (m *mockFeedRepo) UpdateFeed(ctx context.Context, feed *models.Feed) error {
	args := m.Called(feed)
	return args.Error(0)
}

func (m *mockFeedRepo) DeleteFeed(ctx context.Context, id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

/* ---------- Tests ---------- */

func TestCreateFeedSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}

	userId := uint(100)
	username := "mike"
	feedTitle := "title"
	feedContent := "content"
	repo.On("CreateFeed", mock.AnythingOfType("*models.Feed")).Return(nil)

	result, err := svc.CreateFeed(context.Background(), feedTitle, feedContent, userId, username)

	assert.Nil(t, err)
	assert.Equal(t, feedTitle, result.Title)
	assert.Equal(t, username, result.AuthorName)
	assert.Equal(t, feedContent, result.Content)
	assert.Equal(t, userId, result.AuthorID)
	repo.AssertExpectations(t)
}

func TestCreateFeedWhenEmptyTitle(t *testing.T) {
	svc := &FeedService{}
	result, err := svc.CreateFeed(context.Background(), "", "content", 1, "Mike")
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrFeedInvalidContentOrTitle, err)
}

func TestGetFeedsSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	expected := []models.Feed{{AuthorName: "mike"}, {AuthorName: "mike"}}
	repo.On("GetFeeds").Return(expected, nil)

	result, err := svc.GetFeeds(context.Background())

	assert.Nil(t, err)
	assert.Equal(t, expected[0].ToFeedResponse(), &result[0])
	repo.AssertExpectations(t)
}

func TestGetFeedsWhenDBError(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	repo.On("GetFeeds").Return([]models.Feed(nil), errors.New("db error"))

	result, err := svc.GetFeeds(context.Background())
	assert.Nil(t, result)
	assert.Equal(t, appErrors.DatabaseError, err)
}

func TestPaginatedFeedsSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}

	paginated := &models.PaginatedFeedsResponse{Data: []models.Feed{{AuthorName: "mikde"}}, Meta: models.Meta{Limit: 10, HasMore: false}}
	repo.On("PaginatedFeeds", (*models.Cursor)(nil), 10).Return(paginated, nil)

	resp, err := svc.PaginatedFeeds(context.Background(), nil, 10)
	assert.Nil(t, err)
	assert.Equal(t, len(paginated.Data), len(resp.Data))
	assert.Equal(t, false, resp.Meta.HasMore)
}

func TestGetFeedByIDSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	feed := &models.Feed{AuthorName: "mike"}
	feedId := uint(5)
	feed.ID = feedId
	repo.On("GetFeedByID", feed.ID).Return(feed, nil)

	result, err := svc.GetFeedByID(context.Background(), feedId)
	assert.Nil(t, err)
	assert.Equal(t, feedId, result.ID)
}

func TestUpdateFeedSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}

	old := &models.Feed{AuthorID: 1, AuthorName: "mike", Title: "Old", Content: "Old"}
	feedId := uint(1)
	repo.On("GetFeedByID", feedId).Return(old, nil)
	repo.On("UpdateFeed", mock.AnythingOfType("*models.Feed")).Return(nil)

	newContent := "new content"
	newTitle := "new title"
	err := svc.UpdateFeed(context.Background(), feedId, 1, newTitle, newContent)
	assert.Nil(t, err)
}

func TestUpdateFeedWhenFieldsEmpty(t *testing.T) {
	svc := &FeedService{}
	err := svc.UpdateFeed(context.Background(), 1, 1, "", "")

	assert.Equal(t, appErrors.ErrFeedInvalidContentOrTitle, err)
}

func TestDeleteFeedSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	repo.On("GetFeedByID", uint(10)).Return(&models.Feed{AuthorID: 1}, nil)
	repo.On("DeleteFeed", uint(10)).Return(nil)

	err := svc.DeleteFeed(context.Background(), 10, 1)
	assert.Nil(t, err)
}
