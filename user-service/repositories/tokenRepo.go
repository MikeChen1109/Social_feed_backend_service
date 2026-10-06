package repositories

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

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
	script := redis.NewScript(`
 if redis.call('GET', KEYS[1]) ~= ARGV[1] then return 0 end
 redis.call('SET', KEYS[2], ARGV[1], 'PX', ARGV[2])
 redis.call('DEL', KEYS[1])
 return 1`)
	result, err := script.Run(ctx, r.Redis, []string{"refresh:" + oldToken, "refresh:" + newToken}, userID, expiration.Milliseconds()).Int()
	if err != nil {
		return err
	}
	if result != 1 {
		return redis.Nil
	}
	return nil
}
