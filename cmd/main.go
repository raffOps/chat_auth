package main

import (
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
	authController "github.com/raffops/chat_auth/internal/app/auth/controller"
	authService "github.com/raffops/chat_auth/internal/app/auth/service"
	sessionRepository "github.com/raffops/chat_auth/internal/app/sessionManager/repository"
	sessionService "github.com/raffops/chat_auth/internal/app/sessionManager/service"
	user "github.com/raffops/chat_auth/internal/app/user/repository"
	"github.com/raffops/chat_auth/internal/server"
	"github.com/raffops/chat_commons/pkg/database/postgres"
	"github.com/raffops/chat_commons/pkg/database/redis"
	"github.com/raffops/chat_commons/pkg/encryptor"
	"github.com/raffops/chat_commons/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal("cannot load .env file", zap.Error(err))
	}

	userDatabase, err := postgres.GetPostgresConn(true)
	if err != nil {
		logger.Fatal("cannot connect to database", zap.Error(err))
	}
	userRepo := user.NewPostgresUserRepository(userDatabase)
	sessionTimeout, err := time.ParseDuration(os.Getenv("SESSION_TIMEOUT"))
	if err != nil {
		logger.Fatal("cannot parse session timeout", zap.Error(err))
	}

	defaultEncryptor := encryptor.NewDefaultEncryptor()

	sessionRepo := sessionRepository.NewRedisRepository(redis.GetRedisConn(ctx), defaultEncryptor)
	sessionSrv := sessionService.NewDefaultService(
		sessionRepo,
		sessionTimeout,
		os.Getenv("SESSION_MANAGER_SECRET"),
	)
	authSrv := authService.NewDefaultService(userRepo, sessionRepo, sessionSrv)
	controller := authController.NewController(userRepo, sessionSrv, authSrv)

	s := server.NewServer(controller, sessionSrv)

	logger.Info("server started")
	err = s.ListenAndServe()
	if err != nil {
		logger.Fatal("cannot start server", zap.Error(err))
	}
}
