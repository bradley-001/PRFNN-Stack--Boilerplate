package config

import (
	"context"
	"fmt"
	"time"

	"api/src/constants"
	"api/src/lib/general"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectToRedis() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		Log("Could not find .env file", 3, true, false)
		return
	}

	// Redis configuration
	redisAddr := general.GetEnv("REDIS_ADDRESS", "localhost")
	redisPort := general.GetEnv("REDIS_PORT", 6379)
	redisDB := general.GetEnv("REDIS_DB", 0)

	// Create Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", redisAddr, redisPort),
		DB:           redisDB, // use default DB
		DialTimeout:  time.Duration(constants.DEFAULT_TIMEOUT*2) * time.Second,
		ReadTimeout:  time.Duration(constants.DEFAULT_TIMEOUT) * time.Second,
		WriteTimeout: time.Duration(constants.DEFAULT_TIMEOUT) * time.Second,
		PoolSize:     10,
		MinIdleConns: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(constants.DEFAULT_TIMEOUT)*time.Second)
	defer cancel()

	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		Log("Could not connect to Redis Service", 3, true, false)
		return
	}

	notice := fmt.Sprintf("Redis connection successful at %s:%d | Response: %s",
		redisAddr,
		redisPort,
		pong,
	)
	Log(notice, 1, false, false)
}

func CloseRedisConnection() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
