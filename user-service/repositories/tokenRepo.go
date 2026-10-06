package repositories

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
	"gorm.io/gorm"
)

// Keep the replacement write before deletion: if SET fails, the old token remains.
var rotateRefreshTokenScript = redis.NewScript(2, `
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
	StoreRefreshToken(ctx context.Context, token string, userID uint, expiration time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (uint, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	RotateRefreshToken(ctx context.Context, oldToken, newToken string, userID uint, expiration time.Duration) error
}

type TokenRepository struct {
	DB    *gorm.DB
	Redis *redis.Pool
}

func (r *TokenRepository) StoreRefreshToken(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := r.Redis.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = redis.DoContext(conn, ctx, "SET", "refresh:"+token, userID, "PX", expiration.Milliseconds())
	return err
}
func (r *TokenRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (uint, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := r.Redis.GetContext(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	id, err := redis.Uint64(redis.DoContext(conn, ctx, "GET", "refresh:"+token))
	return uint(id), err
}
func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := r.Redis.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = redis.DoContext(conn, ctx, "DEL", "refresh:"+token)
	return err
}

// The check, replacement write, and old-token deletion run as one Redis operation.
// Standalone Redis is required; these keys do not use a Redis Cluster hash tag.
func (r *TokenRepository) RotateRefreshToken(ctx context.Context, oldToken, newToken string, userID uint, expiration time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := r.Redis.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	result, err := redis.Int(rotateRefreshTokenScript.DoContext(ctx, conn, "refresh:"+oldToken, "refresh:"+newToken, userID, expiration.Milliseconds()))
	if err != nil {
		return err
	}
	if result != 1 {
		return redis.ErrNil
	}
	return nil
}
