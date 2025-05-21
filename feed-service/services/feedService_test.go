package services

import (
	"errors"
	appErrors "feed-service/common/appErrors"
	"feed-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/* ---------- Mocks ---------- */

type mockFeedRepo struct{ mock.Mock }

func (m *mockFeedRepo) CreateFeed(feed *models.Feed) error {
	args := m.Called(feed)
	return args.Error(0)
}

func (m *mockFeedRepo) GetFeeds() ([]models.Feed, error) {
	args := m.Called()
	return args.Get(0).([]models.Feed), args.Error(1)
}

func (m *mockFeedRepo) PaginatedFeeds(offset int, limit int) (*models.PaginatedFeedsResponse, error) {
	args := m.Called(offset, limit)
	return args.Get(0).(*models.PaginatedFeedsResponse), args.Error(1)
}

func (m *mockFeedRepo) GetFeedByID(id uint) (*models.Feed, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Feed), args.Error(1)
}

func (m *mockFeedRepo) UpdateFeed(feed *models.Feed) error {
	args := m.Called(feed)
	return args.Error(0)
}

func (m *mockFeedRepo) DeleteFeed(id uint) error {
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

	result, err := svc.CreateFeed(feedTitle, feedContent, userId, username)

	assert.Nil(t, err)
	assert.Equal(t, feedTitle, result.Title)
	assert.Equal(t, username, result.AuthorName)
	assert.Equal(t, feedContent, result.Content)
	assert.Equal(t, userId, result.AuthorID)
	repo.AssertExpectations(t)
}

func TestCreateFeedWhenEmptyTitle(t *testing.T) {
	svc := &FeedService{}
	result, err := svc.CreateFeed("", "content", 1, "Mike")
	assert.Nil(t, result)
	assert.Equal(t, appErrors.ErrFeedInvalidContentOrTitle, err)
}

func TestGetFeedsSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	expected := []models.Feed{{AuthorName: "mike"}, {AuthorName: "mike"}}
	repo.On("GetFeeds").Return(expected, nil)

	result, err := svc.GetFeeds()

	assert.Nil(t, err)
	assert.Equal(t, expected, result)
	repo.AssertExpectations(t)
}

func TestGetFeedsWhenDBError(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	repo.On("GetFeeds").Return([]models.Feed(nil), errors.New("db error"))

	result, err := svc.GetFeeds()
	assert.Nil(t, result)
	assert.Equal(t, appErrors.DatabaseError, err)
}

func TestPaginatedFeedsSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}

	paginated := &models.PaginatedFeedsResponse{Data: []models.Feed{{AuthorName: "mikde"}}, Meta: models.Meta{Page: 1, Limit: 10, HasMore: false}}
	repo.On("PaginatedFeeds", 0, 10).Return(paginated, nil)

	resp, err := svc.PaginatedFeeds(0, 10)
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

	result, err := svc.GetFeedByID(feedId)
	assert.Nil(t, err)
	assert.Equal(t, feedId, result.ID)
}

func TestUpdateFeedSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}

	old := &models.Feed{AuthorName: "mike", Title: "Old", Content: "Old"}
	feedId := uint(1)
	repo.On("GetFeedByID", feedId).Return(old, nil)
	repo.On("UpdateFeed", mock.AnythingOfType("*models.Feed")).Return(nil)


	newContent := "new content"
	newTitle := "new title"
	err := svc.UpdateFeed(feedId, newTitle, newContent)
	assert.Nil(t, err)
}

func TestUpdateFeedWhenFieldsEmpty(t *testing.T) {
	svc := &FeedService{}
	err := svc.UpdateFeed(1, "", "")

	assert.Equal(t, appErrors.ErrFeedInvalidContentOrTitle, err)
}

func TestDeleteFeedSuccess(t *testing.T) {
	repo := new(mockFeedRepo)
	svc := &FeedService{FeedRepo: repo}
	repo.On("DeleteFeed", uint(10)).Return(nil)

	err := svc.DeleteFeed(10)
	assert.Nil(t, err)
}
