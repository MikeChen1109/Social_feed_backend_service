package initializers

import (
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() *redis.Client {
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal("Invalid REDIS_URL")
	}
	client := redis.NewClient(opt)

	return client
}
