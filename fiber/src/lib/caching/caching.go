package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"api/src/config"
	"api/src/constants"
	"api/src/lib/general"
	"api/src/models"
)

var ttlMinutes = time.Duration(general.GetEnv("CACHE_TTL", 900)) * time.Second // Default 15 minutes

// Helper function to get Redis client context with timeout
func GetRedisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), time.Duration(constants.DEFAULT_TIMEOUT)*time.Second)
}

// CacheSession stores a session in Redis
func CacheSession(sid string, session models.Sessions) error {
	ctx, cancel := GetRedisContext()
	defer cancel()

	// Convert session to JSON
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Store in Redis with TTL
	cacheKey := fmt.Sprintf("session:%s", sid)
	return config.RedisClient.Set(ctx, cacheKey, sessionJSON, ttlMinutes).Err()
}

// GetCachedSession retrieves a session from Redis
func GetCachedSession(sid string) (*models.Sessions, error) {
	ctx, cancel := GetRedisContext()
	defer cancel()

	cacheKey := fmt.Sprintf("session:%s", sid)
	sessionJSON, err := config.RedisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err // Could be redis.Nil (cache miss) or connection error
	}

	// Parse JSON back to session struct
	var session models.Sessions
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached session: %w", err)
	}

	return &session, nil
}

// CacheUser stores a user in Redis
func CacheUser(uid string, user models.Users) error {
	ctx, cancel := GetRedisContext()
	defer cancel()

	// Convert user to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Store in Redis with TTL
	cacheKey := fmt.Sprintf("user:%s", uid)
	return config.RedisClient.Set(ctx, cacheKey, userJSON, ttlMinutes).Err()
}

// GetCachedUser retrieves a user from Redis
func GetCachedUser(uid string) (*models.Users, error) {
	ctx, cancel := GetRedisContext()
	defer cancel()

	cacheKey := fmt.Sprintf("user:%s", uid)
	userJSON, err := config.RedisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, err // Could be redis.Nil (cache miss) or connection error
	}

	// Parse JSON back to user struct
	var user models.Users
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached user: %w", err)
	}

	return &user, nil
}

// DropCachedUser removes a user from Redis cache
func DropCachedUser(uid string) error {
	ctx, cancel := GetRedisContext()
	defer cancel()

	cacheKey := fmt.Sprintf("user:%s", uid)
	return config.RedisClient.Del(ctx, cacheKey).Err()
}

// DropCachedSession removes a session from Redis cache
func DropCachedSession(sid string) error {
	ctx, cancel := GetRedisContext()
	defer cancel()

	cacheKey := fmt.Sprintf("session:%s", sid)
	return config.RedisClient.Del(ctx, cacheKey).Err()
}
