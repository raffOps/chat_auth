package auth

import (
	"context"
	"net/http"

	auth "github.com/raffops/chat_auth/internal/app/auth/model"
	user "github.com/raffops/chat_auth/internal/app/user/models"
	"github.com/raffops/chat_commons/pkg/errs"
)

type Controller interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
	DeleteUser(w http.ResponseWriter, r *http.Request)
}

type Service interface {
	SignUp(
		ctx context.Context,
		username, email string,
		authType user.AuthTypeId,
		role auth.RoleId,
	) (string, errs.ChatError)
	Login(ctx context.Context, username, email string) (string, errs.ChatError)
	Refresh(ctx context.Context, sessionId string) errs.ChatError
	Logout(ctx context.Context, sessionId string) errs.ChatError
	DeleteUser(ctx context.Context, userToDelete user.User) errs.ChatError
}
