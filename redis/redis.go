package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Upstash Redis configuration
type RedisConfig struct {
	UpstashURL   string // UPSTASH_REDIS_REST_URL
	UpstashToken string // UPSTASH_REDIS_REST_TOKEN
	PoolSize     int    // Default: 10
}

// RedisService manages Redis connection and operations
type RedisService struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisService creates a new Redis service instance
func NewRedisService(config *RedisConfig) (*RedisService, error) {
	// Parse Upstash URL to get host and port
	parsedURL, err := url.Parse(config.UpstashURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Upstash URL: %w", err)
	}

	// Extract host and port from Upstash URL
	host := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		port = "6379" // Default Redis port
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: config.UpstashToken,
		DB:       0, // Upstash uses single database
		PoolSize: config.PoolSize,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	})

	// Test connection
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Upstash Redis: %w", err)
	}

	return &RedisService{
		client: client,
		config: config,
	}, nil
}

// GetClient returns the Redis client
func (r *RedisService) GetClient() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Ping tests the Redis connection
func (r *RedisService) Ping() error {
	_, err := r.client.Ping(context.Background()).Result()
	return err
}

// HealthCheck performs a health check on the Redis service
func (r *RedisService) HealthCheck() error {
	return r.Ping()
}
