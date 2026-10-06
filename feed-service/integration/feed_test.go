package integration

import (
	"encoding/json"
	"feed-service/controllers"
	"feed-service/models"
	"feed-service/repositories"
	"feed-service/routes"
	"feed-service/services"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func setup(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	t.Setenv("JWT_SECRET", "test-secret")
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&models.Feed{}, &models.Comment{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	r := gin.New()
	routes.RegisterFeedRoutes(r, &controllers.FeedsController{FeedsService: &services.FeedService{FeedRepo: &repositories.FeedRepository{DB: db}}})
	routes.RegisterCommentRoutes(r, &controllers.CommentsController{CommentsService: &services.CommentService{CommentRepo: &repositories.CommentRepository{DB: db}}})
	return r, db
}
func call(r *gin.Engine, method, path, body string, uid uint) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if uid != 0 {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{UserID: uid, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}})
		signed, _ := token.SignedString([]byte("test-secret"))
		req.Header.Set("Authorization", "Bearer "+signed)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
func TestAuthorOnlyMutations(t *testing.T) {
	r, db := setup(t)
	feed := models.Feed{AuthorID: 1, Title: "original", Content: "content"}
	require.NoError(t, db.Create(&feed).Error)
	path := fmt.Sprintf("/feed/%d", feed.ID)
	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		w := call(r, method, path, `{"title":"changed","content":"changed"}`, 2)
		require.Equal(t, 403, w.Code)
		w = call(r, method, path, `{}`, 0)
		require.Equal(t, 401, w.Code)
	}
	var stored models.Feed
	require.NoError(t, db.First(&stored, feed.ID).Error)
	require.Equal(t, "original", stored.Title)
	require.Equal(t, 400, call(r, "PUT", path, `{`, 1).Code)
	require.Equal(t, 200, call(r, "PUT", path, `{"title":"changed","content":"changed"}`, 1).Code)
	require.Equal(t, 200, call(r, "DELETE", path, ``, 1).Code)
	require.Equal(t, 404, call(r, "DELETE", path, ``, 1).Code)
}
func TestSecondPageDoesNotOverlapForFeedsAndComments(t *testing.T) {
	r, db := setup(t)
	// Equal timestamps verify the ID tie-breaker, not only time sorting.
	stamp := time.Now()
	for i := 0; i < 5; i++ {
		require.NoError(t, db.Create(&models.Feed{Model: gorm.Model{CreatedAt: stamp}, AuthorID: 1, Title: "feed", Content: "content"}).Error)
		require.NoError(t, db.Create(&models.Comment{Model: gorm.Model{CreatedAt: stamp}, FeedID: 1, Content: "comment"}).Error)
	}
	for _, path := range []string{"/feed/paginated?", "/comment/paginated?id=1&"} {
		ids := map[uint]bool{}
		for page := 1; page <= 3; page++ {
			w := call(r, "GET", fmt.Sprintf("%spage=%d&limit=2", path, page), "", 0)
			require.Equal(t, 200, w.Code)
			var result struct {
				Data []struct{ ID uint }
				Meta models.Meta
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
			require.Equal(t, page, result.Meta.Page)
			require.Equal(t, page < 3, result.Meta.HasMore)
			for _, item := range result.Data {
				require.False(t, ids[item.ID], "overlapping page item")
				ids[item.ID] = true
			}
		}
		require.Len(t, ids, 5)
	}
	require.Equal(t, 400, call(r, "GET", "/comment/paginated?id=-1", "", 0).Code)
	require.Equal(t, 400, call(r, "GET", "/feed/paginated?page=9223372036854775807&limit=100", "", 0).Code)
}
