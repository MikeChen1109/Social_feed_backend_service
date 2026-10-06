package repositories

import (
	"testing"
	"time"
	"user-service/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		rdb.Close()
		s.Close()
	}

	return rdb, cleanup
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
	err := repo.StoreRefreshToken(expectedToken, expectedUserId, expiration)
	assert.Nil(err)

	id, err := repo.GetUserIDByRefreshToken(expectedToken)
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
	err := repo.StoreRefreshToken(token, userId, expiration)
	assert.Nil(err)

	fakeToken := uuid.NewString()
	id, err := repo.GetUserIDByRefreshToken(fakeToken)

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
	storeError := repo.StoreRefreshToken(token, userId, expiration)
	assert.Nil(storeError)

	err := repo.DeleteRefreshToken(token)
	assert.Nil(err)
}

func TestConcurrentRefreshRotationHasOneWinner(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken("old", 42, time.Hour))
	const workers = 20
	start := make(chan struct{})
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() { <-start; results <- repo.RotateRefreshToken("old", uuid.NewString(), 42, time.Hour) }()
	}
	close(start)
	winners := 0
	for i := 0; i < workers; i++ {
		err := <-results
		if err == nil {
			winners++
		} else {
			assert.ErrorIs(t, err, redis.Nil)
		}
	}
	assert.Equal(t, 1, winners)
	_, err := repo.GetUserIDByRefreshToken("old")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestRefreshRotationPreservesOldTokenForWrongUser(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken("old", 42, time.Hour))
	assert.ErrorIs(t, repo.RotateRefreshToken("old", "new", 99, time.Hour), redis.Nil)
	id, err := repo.GetUserIDByRefreshToken("old")
	assert.NoError(t, err)
	assert.Equal(t, uint(42), id)
	_, err = repo.GetUserIDByRefreshToken("new")
	assert.ErrorIs(t, err, redis.Nil)
}

func TestFailedReplacementWritePreservesOldToken(t *testing.T) {
	repo, dbCleanup, redisCleanup := setupTokenRepoForTest(t)
	defer dbCleanup()
	defer redisCleanup()
	assert.NoError(t, repo.StoreRefreshToken("old", 42, time.Hour))
	// Redis rejects SET PX 0; the script must not delete the old token first.
	assert.Error(t, repo.RotateRefreshToken("old", "new", 42, 0))
	id, err := repo.GetUserIDByRefreshToken("old")
	assert.NoError(t, err)
	assert.Equal(t, uint(42), id)
}
