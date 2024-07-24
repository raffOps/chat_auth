package server

import (
	"encoding/json"
	"github.com/raffops/auth/internal/app/auth"
	authModel "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/sessionManager"
	"github.com/raffops/chat/pkg/logger"
	"go.uber.org/zap"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes(authController auth.Controller, sessionMgr sessionManager.Service) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", s.HelloWorldHandler)
	r.HandleFunc(
		"/session_id",
		sessionMgr.CheckRestSession(s.HelloWorldUser, []authModel.RoleId{authModel.RoleAdmin, authModel.RoleUser}),
	)
	r.HandleFunc("/health", s.healthHandler)

	r.HandleFunc("/login/{provider}", authController.Login)
	r.HandleFunc("/login/{provider}/callback", authController.Callback)
	r.HandleFunc("/signUp", authController.SignUp)
	return r
}

func (s *Server) HelloWorldUser(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello user"))
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		logger.Debug("error handling JSON marshal", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		logger.Debug("error handling JSON marshal", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
	}

	_, _ = w.Write(jsonResp)
}
