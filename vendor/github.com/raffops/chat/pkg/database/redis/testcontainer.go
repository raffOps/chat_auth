package redis

import (
	"context"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

func GetRedisTestContainer(ctx context.Context) (*redis.RedisContainer, error) {
	return redis.Run(
		ctx,
		"redis:latest",
		testcontainers.WithHostPortAccess(6379),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections tcp").
			WithStartupTimeout(15*time.Second),
		),
	)
}
