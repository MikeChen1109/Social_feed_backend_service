package repositories

import (
	"context"
	"github.com/alicebob/miniredis/v2/server"
	"sync"
	"testing"
	"time"
	"user-service/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/* ---------- helper ---------- */

func setupTestDB() (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test db")
	}

	db.AutoMigrate(&models.User{})

	cleanup := func() {
		sqlDB, err := db.DB()
		if err != nil {
			panic("failed to get raw db")
		}
		sqlDB.Close()
	}

	return db, cleanup
}

func setupTestRedis(t *testing.T) (*redis.Pool, func()) {
	s := miniredis.RunT(t)
	pool := &redis.Pool{MaxIdle: 10, MaxActive: 50, Wait: true, DialContext: func(ctx context.Context) (redis.Conn, error) { return redis.DialContext(ctx, "tcp", s.Addr()) }}
	return pool, func() { pool.Close(); s.Close() }
}

func setupTokenRepoForTest(t *testing.T) (*TokenRepository, func(), func()) {
	db, dbCleanUp := setupTestDB()
	redis, redisCleanUp := setupTestRedis(t)
	repo := &TokenRepository{DB: db, Redis: redis}

	return repo, dbCleanUp, redisCleanUp
}

/* ---------- Tests ---------- */

func TestStoreRefreshTokenSuccess(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanUp, redisCleanUp := setupTokenRepoForTest(t)
	expectedToken := uuid.NewString()
	expectedUserId := uint(100)
	defer dbCleanUp()
	defer redisCleanUp()

	expiration := time.Hour * 24 * 30
	err := repo.StoreRefreshToken(context.Background(), expectedToken, expectedUserId, expiration)
	assert.Nil(err)

	id, err := repo.GetUserIDByRefreshToken(context.Background(), expectedToken)
	assert.Nil(err)
	assert.Equal(id, expectedUserId)
}

func TestGetUserIDByRefreshTokenWhenTokenNotExists(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanUp, redisCleanUp := setupTokenRepoForTest(t)
	token := uuid.NewString()
	userId := uint(100)
	defer dbCleanUp()
	defer redisCleanUp()

	expiration := time.Hour * 24 * 30
	err := repo.StoreRefreshToken(context.Background(), token, userId, expiration)
	assert.Nil(err)

	fakeToken := uuid.NewString()
	id, err := repo.GetUserIDByRefreshToken(context.Background(), fakeToken)

	assert.NotNil(err)
	assert.Equal(id, uint(0))
}

func TestDeleteRefreshTokenSuccess(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanUp, redisCleanUp := setupTokenRepoForTest(t)
	token := uuid.NewString()
	userId := uint(100)
	defer dbCleanUp()
	defer redisCleanUp()

	expiration := time.Hour * 24 * 30
	storeError := repo.StoreRefreshToken(context.Background(), token, userId, expiration)
	assert.Nil(storeError)

	err := repo.DeleteRefreshToken(context.Background(), token)
	assert.Nil(err)
}

func TestConcurrentRefreshRotationHasOneWinner(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken(context.Background(), "old", 42, time.Hour))
	const workers = 20
	start := make(chan struct{})
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			<-start
			results <- repo.RotateRefreshToken(context.Background(), "old", uuid.NewString(), 42, time.Hour)
		}()
	}
	close(start)
	winners := 0
	for i := 0; i < workers; i++ {
		err := <-results
		if err == nil {
			winners++
		} else {
			assert.ErrorIs(t, err, redis.ErrNil)
		}
	}
	assert.Equal(t, 1, winners)
	_, err := repo.GetUserIDByRefreshToken(context.Background(), "old")
	assert.ErrorIs(t, err, redis.ErrNil)
}

func TestRefreshRotationPreservesOldTokenForWrongUser(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken(context.Background(), "old", 42, time.Hour))
	assert.ErrorIs(t, repo.RotateRefreshToken(context.Background(), "old", "new", 99, time.Hour), redis.ErrNil)
	id, err := repo.GetUserIDByRefreshToken(context.Background(), "old")
	assert.NoError(t, err)
	assert.Equal(t, uint(42), id)
	_, err = repo.GetUserIDByRefreshToken(context.Background(), "new")
	assert.ErrorIs(t, err, redis.ErrNil)
}

func TestFailedReplacementWritePreservesOldToken(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken(context.Background(), "old", 42, time.Hour))
	// Redis rejects SET PX 0; the script must not delete the old token first.
	assert.Error(t, repo.RotateRefreshToken(context.Background(), "old", "new", 42, 0))
	id, err := repo.GetUserIDByRefreshToken(context.Background(), "old")
	assert.NoError(t, err)
	assert.Equal(t, uint(42), id)
}

func TestCanceledContextDoesNotStoreOrRotateTokens(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	require.NoError(t, repo.StoreRefreshToken(context.Background(), "old", 42, time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	require.ErrorIs(t, repo.StoreRefreshToken(ctx, "new", 42, time.Hour), context.Canceled)
	require.ErrorIs(t, repo.RotateRefreshToken(ctx, "old", "replacement", 42, time.Hour), context.Canceled)
	require.ErrorIs(t, repo.DeleteRefreshToken(ctx, "old"), context.Canceled)
	id, err := repo.GetUserIDByRefreshToken(context.Background(), "old")
	require.NoError(t, err)
	require.Equal(t, uint(42), id)
	_, err = repo.GetUserIDByRefreshToken(context.Background(), "new")
	require.ErrorIs(t, err, redis.ErrNil)
}

func TestCancellationInterruptsInFlightRedisRead(t *testing.T) {
	s := miniredis.RunT(t)
	s.Set("refresh:old", "42")
	pool := &redis.Pool{MaxIdle: 1, MaxActive: 2, Wait: true, DialContext: func(ctx context.Context) (redis.Conn, error) {
		return redis.DialContext(ctx, "tcp", s.Addr(), redis.DialReadTimeout(5*time.Second))
	}}
	defer pool.Close()
	entered := make(chan struct{})
	release := make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	defer unblock()
	s.Server().SetPreHook(func(_ *server.Peer, command string, _ ...string) bool {
		if command == "GET" {
			close(entered)
			<-release
		}
		return false
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repo := &TokenRepository{Redis: pool}
	result := make(chan error, 1)
	go func() { _, err := repo.GetUserIDByRefreshToken(ctx, "old"); result <- err }()
	<-entered
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(time.Second):
		t.Fatal("Redis read waited for its own timeout instead of request cancellation")
	}
	s.Server().SetPreHook(nil)
	unblock()
	id, err := repo.GetUserIDByRefreshToken(context.Background(), "old")
	require.NoError(t, err)
	require.Equal(t, uint(42), id)
}
