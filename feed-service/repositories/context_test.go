package repositories

import (
	"context"
	"errors"
	"feed-service/models"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestRepositoryInheritsCancellationAndDeadline(t *testing.T) {
	for _, mode := range []string{"cancel", "deadline"} {
		t.Run(mode, func(t *testing.T) {
			repo, cleanup := setupFeedRepoForTest()
			defer cleanup()
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), contextKey{}, "request"))
			if mode == "deadline" {
				cancel()
				ctx, cancel = context.WithTimeout(context.WithValue(context.Background(), contextKey{}, "request"), 50*time.Millisecond)
			}
			defer cancel()
			started := make(chan struct{})
			observed := make(chan any, 1)
			require.NoError(t, repo.DB.Callback().Query().Before("gorm:query").Register("block", func(tx *gorm.DB) {
				observed <- tx.Statement.Context.Value(contextKey{})
				close(started)
				<-tx.Statement.Context.Done()
				tx.AddError(tx.Statement.Context.Err())
			}))
			result := make(chan error, 1)
			go func() { _, err := repo.GetFeeds(ctx); result <- err }()
			<-started
			require.Equal(t, "request", <-observed)
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-result:
				require.True(t, errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded))
			case <-time.After(time.Second):
				t.Fatal("repository detached from request context")
			}
		})
	}
}

type contextKey struct{}

func TestPostgresCancelsInFlightQuery(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN for real PostgreSQL cancellation test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	defer raw.Close()
	entered := make(chan struct{})
	require.NoError(t, db.Callback().Query().Before("gorm:query").Register("slow_query", func(tx *gorm.DB) {
		close(entered)
		rows, err := tx.Statement.ConnPool.QueryContext(tx.Statement.Context, "SELECT pg_sleep(10)")
		if rows != nil {
			rows.Close()
		}
		tx.AddError(err)
	}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := (&FeedRepository{DB: db}).GetFeeds(ctx); result <- err }()
	<-entered
	time.Sleep(100 * time.Millisecond)
	cancel()
	select {
	case err := <-result:
		require.Error(t, err)
	case <-time.After(time.Second):
		t.Fatal("PostgreSQL query did not stop on request cancellation")
	}
	// A canceled create must not reach the database.
	require.Error(t, (&FeedRepository{DB: db}).CreateFeed(ctx, &models.Feed{Title: "canceled"}))
}

func TestPostgresCursorPaginationSurvivesMutations(t *testing.T) {
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("set TEST_POSTGRES_DSN for real PostgreSQL cursor tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	raw, err := db.DB()
	require.NoError(t, err)
	defer raw.Close()
	for _, kind := range []string{"feeds", "comments"} {
		t.Run(kind, func(t *testing.T) {
			tx := db.Begin()
			require.NoError(t, tx.Error)
			defer tx.Rollback()
			stamp := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond)
			anchor := models.Feed{Title: "cursor-test-anchor"}
			require.NoError(t, tx.Create(&anchor).Error)
			ids := []uint{}
			for i := 0; i < 5; i++ {
				if kind == "feeds" {
					row := models.Feed{Model: gorm.Model{CreatedAt: stamp}, Title: "cursor-test"}
					require.NoError(t, tx.Create(&row).Error)
					ids = append(ids, row.ID)
				} else {
					row := models.Comment{Model: gorm.Model{CreatedAt: stamp}, FeedID: anchor.ID, Content: "cursor-test"}
					require.NoError(t, tx.Create(&row).Error)
					ids = append(ids, row.ID)
				}
			}
			scope := "feeds"
			if kind == "comments" {
				scope = models.CommentCursorScope(anchor.ID)
			}
			fetch := func(token string) ([]uint, models.Meta) {
				cursor, _, err := models.ParsePagination(url.Values{"cursor": {token}}, scope)
				require.NoError(t, err)
				found := []uint{}
				var meta models.Meta
				if kind == "feeds" {
					page, err := (&FeedRepository{DB: tx}).PaginatedFeeds(context.Background(), cursor, 2)
					require.NoError(t, err)
					meta = page.Meta
					for _, row := range page.Data {
						found = append(found, row.ID)
					}
				} else {
					page, err := (&CommentRepository{DB: tx}).PaginatedComments(context.Background(), cursor, 2, anchor.ID)
					require.NoError(t, err)
					meta = page.Meta
					for _, row := range page.Data {
						found = append(found, row.ID)
					}
				}
				return found, meta
			}
			first, meta := fetch("")
			require.Equal(t, []uint{ids[4], ids[3]}, first)
			if kind == "feeds" {
				require.NoError(t, tx.Delete(&models.Feed{}, []uint{ids[4], ids[3]}).Error)
				require.NoError(t, tx.Create(&models.Feed{Model: gorm.Model{CreatedAt: stamp}, Title: "newer-id"}).Error)
			} else {
				require.NoError(t, tx.Delete(&models.Comment{}, []uint{ids[4], ids[3]}).Error)
				require.NoError(t, tx.Create(&models.Comment{Model: gorm.Model{CreatedAt: stamp}, FeedID: anchor.ID, Content: "newer-id"}).Error)
			}
			second, meta := fetch(meta.NextCursor)
			require.Equal(t, []uint{ids[2], ids[1]}, second)
			last, _ := fetch(meta.NextCursor)
			require.NotEmpty(t, last)
			require.Equal(t, ids[0], last[0])
		})
	}
}
