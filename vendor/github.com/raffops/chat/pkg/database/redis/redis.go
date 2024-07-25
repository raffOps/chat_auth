package redis

import (
	"context"
	"github.com/raffops/chat/pkg"
	"github.com/raffops/chat/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"os"
)

func GetRedisConn(ctx context.Context) *redis.Client {
	envVariables := []string{"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD"}
	pkg.SanityCheck(logger.Logger, envVariables)
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
	})

	err := rdb.Ping(ctx).Err()
	if err != nil {
		logger.Fatal("error pinging redis", zap.Error(err))
	}
	return rdb
}
