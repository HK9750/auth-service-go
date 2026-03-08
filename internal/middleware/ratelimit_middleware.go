package middleware

import (
	"github.com/redis/go-redis/v9"
)

type RateLimitConfig struct {
	MaxLimit   int
	WindowSize int
}

func NewRateLimitConfig(MaxLimit int, WindowSize int) *RateLimitConfig {
	return &RateLimitConfig{
		MaxLimit:   MaxLimit,
		WindowSize: WindowSize,
	}
}

type RateLimiter struct {
	Client *redis.Client
	Config *RateLimitConfig
}

func NewRateLimiter(Client *redis.Client, Config *RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		Client: Client,
		Config: Config,
	}
}
