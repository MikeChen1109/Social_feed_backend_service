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

func TestStoreRefreshToken(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanUp, redisCleanUp  := setupTokenRepoForTest(t)
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
	repo, dbCleanUp, redisCleanUp  := setupTokenRepoForTest(t)
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

func TestDeleteRefreshToken(t *testing.T) {
	assert := assert.New(t)
	repo, dbCleanUp, redisCleanUp  := setupTokenRepoForTest(t)
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
