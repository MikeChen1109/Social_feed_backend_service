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
}

type TokenRepository struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (r *TokenRepository) StoreRefreshToken(token string, userID uint, expiration time.Duration) error {
	ctx := context.Background()
	err := r.Redis.Set(ctx, "refresh:"+token, userID, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *TokenRepository) GetUserIDByRefreshToken(token string) (uint, error) {
	ctx := context.Background()
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
	ctx := context.Background()
	key := "refresh:" + token
	return r.Redis.Del(ctx, key).Err()
}
