package repositories

import (
	"feed-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/* ---------- helper ---------- */

func setupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test db")
	}

	db.AutoMigrate(&models.Feed{})

	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			panic("failed to get raw db")
		}
		sqlDB.Close()
	}

	return db, cleanup
}

func setupFeedRepoForTest() (*FeedRepository, func()) {
	db, dbCleanUp := setupTestDB()
	repo := &FeedRepository{DB: db}

	return repo, dbCleanUp
}

func generateFeeds(db *gorm.DB) {
	for i := 0; i < 100; i++ {
		feed := models.Feed{
			AuthorName: "name",
			AuthorID:   1,
			Title:      "title",
			Content:    "content",
		}
		db.Create(&feed)
	}
}

/* ---------- Tests ---------- */

func TestCreateFeedSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	feed := models.Feed{AuthorName: "name", AuthorID: 1, Title: "title", Content: "content"}
	err := repo.CreateFeed(&feed)

	assert.Nil(t, err)
}

func TestGetFeedsSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	feed := models.Feed{AuthorName: "name", AuthorID: 1, Title: "title", Content: "content"}
	err := repo.CreateFeed(&feed)
	assert.Nil(t, err)

	feeds, err := repo.GetFeeds()
	assert.Nil(t, err)
	assert.Equal(t, len(feeds), 1)
}

func TestPaginatedFeedsSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	offset := 0
	limit := 10
	generateFeeds(repo.DB)
	response, err := repo.PaginatedFeeds(offset, limit)

	assert.Nil(t, err)
	assert.Equal(t, 10, len(response.Data))
	assert.Equal(t, true, response.Meta.HasMore)
	assert.Equal(t, 1, response.Meta.Page)
}

func TestGetFeedByIDSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	mockFeed := models.Feed{AuthorName: "name", AuthorID: 1, Title: "title", Content: "content"}
	feedId := uint(99)
	mockFeed.ID = uint(feedId)

	assert.Nil(t, repo.CreateFeed(&mockFeed))

	feed, err := repo.GetFeedByID(feedId)
	assert.Nil(t, err)
	assert.Equal(t, feedId, feed.ID)
	assert.Equal(t, mockFeed.AuthorName, feed.AuthorName)
	assert.Equal(t, mockFeed.AuthorID, feed.AuthorID)
}

func TestUpdateFeedSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	mockFeed := models.Feed{AuthorName: "name", AuthorID: 1, Title: "title", Content: "content"}
	feedId := uint(99)
	mockFeed.ID = uint(feedId)

	assert.Nil(t, repo.CreateFeed(&mockFeed))

	updatedContent := "new content"
	updatedTitle := "new title"
	mockFeed.Content = updatedContent
	mockFeed.Title = updatedTitle

	err := repo.UpdateFeed(&mockFeed)
	assert.Nil(t, err)

	updatedFeed, err := repo.GetFeedByID(mockFeed.ID)
	assert.Nil(t, err)

	assert.Equal(t, feedId, mockFeed.ID)
	assert.Equal(t, mockFeed.AuthorName, updatedFeed.AuthorName)
	assert.Equal(t, mockFeed.AuthorID, updatedFeed.AuthorID)
	assert.Equal(t, updatedContent, updatedFeed.Content)
	assert.Equal(t, updatedTitle, updatedFeed.Title)
}

func TestDeleteFeedSuccess(t *testing.T) {
	repo, dbCleanUp := setupFeedRepoForTest()
	defer dbCleanUp()

	mockFeed := models.Feed{AuthorName: "name", AuthorID: 1, Title: "title", Content: "content"}
	feedId := uint(99)
	mockFeed.ID = uint(feedId)

	assert.Nil(t, repo.CreateFeed(&mockFeed))

	err := repo.DeleteFeed(mockFeed.ID)
	assert.Nil(t, err)
}
