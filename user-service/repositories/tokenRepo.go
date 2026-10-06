package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Keep the replacement write before deletion: if SET fails, the old token remains.
var rotateRefreshTokenScript = redis.NewScript(`
local oldTokenKey = KEYS[1]
local newTokenKey = KEYS[2]
local expectedUserID = ARGV[1]
local ttlMilliseconds = ARGV[2]

local storedUserID = redis.call('GET', oldTokenKey)
if storedUserID ~= expectedUserID then
    return 0
end

redis.call('SET', newTokenKey, expectedUserID, 'PX', ttlMilliseconds)
redis.call('DEL', oldTokenKey)
return 1
`)

type TokenRepositoryInterface interface {
	StoreRefreshToken(token string, userID uint, expiration time.Duration) error
	GetUserIDByRefreshToken(token string) (uint, error)
	DeleteRefreshToken(token string) error
	RotateRefreshToken(oldToken, newToken string, userID uint, expiration time.Duration) error
}

type TokenRepository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (r *TokenRepository) StoreRefreshToken(token string, userID uint, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := r.Redis.Set(ctx, "refresh:"+token, userID, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *TokenRepository) GetUserIDByRefreshToken(token string) (uint, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := "refresh:" + token
	val, err := r.Redis.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func (r *TokenRepository) DeleteRefreshToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	key := "refresh:" + token
	return r.Redis.Del(ctx, key).Err()
}

// The check, replacement write, and old-token deletion run as one Redis operation.
// Standalone Redis is required; these keys do not use a Redis Cluster hash tag.
func (r *TokenRepository) RotateRefreshToken(oldToken, newToken string, userID uint, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keys := []string{"refresh:" + oldToken, "refresh:" + newToken}
	result, err := rotateRefreshTokenScript.Run(
		ctx, r.Redis, keys, userID, expiration.Milliseconds(),
	).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return redis.Nil
	}
	return nil
}
