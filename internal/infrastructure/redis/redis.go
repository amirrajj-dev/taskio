package redis

import (
	"context"
	"log"
	"time"
	"strings"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectToRedis() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse Redis URL to extract host:port
	redisAddr := configs.Configs.REDIS_URL
	
	redisAddr = strings.TrimPrefix(redisAddr, "redis://")
	redisAddr = strings.TrimPrefix(redisAddr, "rediss://")
	
	// Remove any path
	if idx := strings.Index(redisAddr, "/"); idx != -1 {
		redisAddr = redisAddr[:idx]
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		DB:       0,
	})

	var err error
	for i := 0; i < 3; i++ {
		err = RedisClient.Ping(ctx).Err()
		if err == nil {
			log.Println("connected to redis successfully")
			return
		}
		log.Printf("Redis ping attempt %d failed: %v, retrying...", i+1, err)
		time.Sleep(1 * time.Second)
	}

	log.Fatalf("failed to ping redis db after 3 attempts: %v", err)
}