package jwtManager

import (
	"github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/user/models"
	"github.com/raffops/chat/pkg/errs"
	"net/http"
)

type JwtManager interface {
	GenerateToken(user user.User, sessionId string) (string, errs.ChatError)
	VerifyToken(tokenString string) (*auth.Claims, errs.ChatError)
	CheckIfAuthorized(f http.HandlerFunc, roles []auth.RoleId) http.HandlerFunc
}
