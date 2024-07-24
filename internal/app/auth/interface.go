package auth

import (
	"context"
	"github.com/raffops/chat/pkg/errs"
	"net/http"
)

type Controller interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	Callback(w http.ResponseWriter, r *http.Request)
}

type Service interface {
	SignUp(ctx context.Context, username, email, authType, role string) (string, errs.ChatError)
	Login(ctx context.Context, username, email string) (string, errs.ChatError)
}
