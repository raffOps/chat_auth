package server

import (
	"fmt"
	"github.com/raffops/auth/internal/app/auth"
	"github.com/raffops/auth/internal/app/sessionManager"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	database "github.com/raffops/chat/pkg/database/postgres"
	"github.com/raffops/chat/pkg/logger"
)

type Server struct {
	port int

	db database.Service
}

func NewServer(authController auth.Controller, sessionMgr sessionManager.Service) *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	NewServer := &Server{
		port: port,

		db: database.New(),
	}

	handler := NewServer.RegisterRoutes(authController, sessionMgr)
	loggedHandler := logger.LoggingMiddleware()(handler)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      loggedHandler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
