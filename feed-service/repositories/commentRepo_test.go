package repositories

import (
	"feed-service/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

/* ---------- helper ---------- */

func setupCommentRepoForTest() (*CommentRepository, func()) {
	db, dbCleanUp := setupTestDB()
	repo := &CommentRepository{DB: db}

	return repo, dbCleanUp
}

func generateComments(db *gorm.DB) {
	for i := 0; i < 100; i++ {
		feed := models.Comment{
			AuthorName: "name",
			AuthorID:   1,
			FeedID:     1,
			Content:    "content",
		}
		db.Create(&feed)
	}
}

/* ---------- Tests ---------- */

func TestCreateCommentSuccess(t *testing.T) {
	repo, dbCleanUp := setupCommentRepoForTest()
	defer dbCleanUp()

	comment := models.Comment{AuthorName: "name", AuthorID: 1, FeedID: 1, Content: "content"}
	err := repo.CreateComment(&comment)

	assert.Nil(t, err)
}

func TestPaginatedCommentsSuccess(t *testing.T) {
	repo, dbCleanUp := setupCommentRepoForTest()
	defer dbCleanUp()

	offset := 0
	limit := 10
	feedId := uint(1)
	generateComments(repo.DB)
	response, err := repo.PaginatedComments(offset, limit, feedId)

	assert.Nil(t, err)
	assert.Equal(t, 10, len(response.Data))
	assert.Equal(t, true, response.Meta.HasMore)
	assert.Equal(t, 1, response.Meta.Page)
}
