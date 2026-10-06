package integration

import (
	"context"
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
	"net/url"
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

func TestCursorPaginationSurvivesInsertDeleteAndTimestampsTies(t *testing.T) {
	for _, kind := range []string{"feeds", "comments"} {
		t.Run(kind, func(t *testing.T) {
			r, db := setup(t)
			stamp := time.Now().UTC().Truncate(time.Microsecond)
			for i := 0; i < 5; i++ {
				if kind == "feeds" {
					require.NoError(t, db.Create(&models.Feed{Model: gorm.Model{CreatedAt: stamp}, AuthorID: 1, Title: "feed", Content: "content"}).Error)
				} else {
					require.NoError(t, db.Create(&models.Comment{Model: gorm.Model{CreatedAt: stamp}, FeedID: 1, Content: "comment"}).Error)
				}
			}
			path := "/feed/paginated?limit=2"
			if kind == "comments" {
				path = "/comment/paginated?id=1&limit=2"
			}
			fetch := func(cursor string) ([]uint, models.Meta) {
				requestPath := path
				if cursor != "" {
					requestPath += "&cursor=" + url.QueryEscape(cursor)
				}
				w := call(r, "GET", requestPath, "", 0)
				require.Equal(t, 200, w.Code, w.Body.String())
				var result struct {
					Data []struct{ ID uint }
					Meta models.Meta
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
				ids := []uint{}
				for _, item := range result.Data {
					ids = append(ids, item.ID)
				}
				return ids, result.Meta
			}
			first, meta := fetch("")
			require.Equal(t, []uint{5, 4}, first)
			require.True(t, meta.HasMore)
			require.NotEmpty(t, meta.NextCursor)
			// Remove both already-read records, including the record encoded in the cursor.
			// Insert one newer item with the same timestamp and one with a later timestamp.
			if kind == "feeds" {
				require.NoError(t, db.Delete(&models.Feed{}, []uint{5, 4}).Error)
				require.NoError(t, db.Create(&models.Feed{Model: gorm.Model{CreatedAt: stamp}, AuthorID: 1, Title: "new"}).Error)
				require.NoError(t, db.Create(&models.Feed{Model: gorm.Model{CreatedAt: stamp.Add(time.Second)}, AuthorID: 1, Title: "newer"}).Error)
			} else {
				require.NoError(t, db.Delete(&models.Comment{}, []uint{5, 4}).Error)
				require.NoError(t, db.Create(&models.Comment{Model: gorm.Model{CreatedAt: stamp}, FeedID: 1, Content: "new"}).Error)
				require.NoError(t, db.Create(&models.Comment{Model: gorm.Model{CreatedAt: stamp.Add(time.Second)}, FeedID: 1, Content: "newer"}).Error)
				require.NoError(t, db.Create(&models.Comment{FeedID: 2, Content: "different feed"}).Error)
			}
			second, meta := fetch(meta.NextCursor)
			require.Equal(t, []uint{3, 2}, second)
			require.True(t, meta.HasMore)
			last, meta := fetch(meta.NextCursor)
			require.Equal(t, []uint{1}, last)
			require.False(t, meta.HasMore)
			require.Empty(t, meta.NextCursor)
			for _, query := range []string{"&page=2", "&cursor=bad", "&limit=0"} {
				require.Equal(t, 400, call(r, "GET", path+query, "", 0).Code)
			}
			wrong := models.EncodeCursor("wrong-scope", stamp, 4)
			require.Equal(t, 400, call(r, "GET", path+"&cursor="+url.QueryEscape(wrong), "", 0).Code)
		})
	}
}

func TestHTTPDisconnectCancelsRepositoryWork(t *testing.T) {
	r, db := setup(t)
	entered := make(chan struct{})
	stopped := make(chan struct{})
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("wait_for_disconnect", func(tx *gorm.DB) {
		close(entered)
		<-tx.Statement.Context.Done()
		tx.AddError(tx.Statement.Context.Err())
		close(stopped)
	}))
	server := httptest.NewServer(r)
	defer server.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/feed/", nil)
	require.NoError(t, err)
	clientDone := make(chan error, 1)
	go func() {
		resp, err := server.Client().Do(req)
		if resp != nil {
			resp.Body.Close()
		}
		clientDone <- err
	}()
	<-entered
	cancel()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("HTTP cancellation did not reach repository")
	}
	require.ErrorIs(t, <-clientDone, context.Canceled)
}
