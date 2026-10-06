package initializers

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"log"
	"net/url"
	"os"
	"time"
)

func ConnectRedis() *redis.Pool {
	address := os.Getenv("REDIS_URL")
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "redis" && parsed.Scheme != "rediss") {
		log.Fatal("Invalid REDIS_URL")
	}
	return &redis.Pool{MaxIdle: 10, MaxActive: 50, Wait: true, IdleTimeout: 5 * time.Minute,
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURLContext(ctx, address, redis.DialConnectTimeout(5*time.Second), redis.DialReadTimeout(5*time.Second), redis.DialWriteTimeout(5*time.Second))
		},
	}
}
