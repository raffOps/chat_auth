package auth

import (
	"context"
	"fmt"
	"github.com/raffops/auth/internal/app/auth"
	authModels "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/sessionManager"
	"github.com/raffops/auth/internal/app/user"
	userModels "github.com/raffops/auth/internal/app/user/models"
	"github.com/raffops/chat/pkg/errs"
)

type defaultService struct {
	userRepo   user.ReaderWriterRepository
	sessionSrv sessionManager.Service
}

func (s defaultService) SignUp(ctx context.Context, username, email, authType, role string) (string, errs.ChatError) {
	u := userModels.User{
		Username: username,
		Email:    email,
		AuthType: userModels.MapAuthTypeString[authType],
		Role:     authModels.MapRoleString[role],
		Status:   userModels.StatusActive,
	}
	tx, _ := s.userRepo.GetDB().BeginTx(ctx, nil)
	defer tx.Rollback()

	createUser, err := s.userRepo.CreateUser(ctx, tx, u)
	if err != nil {
		return "", err
	}

	sessionId, err := s.sessionSrv.CreateSession(
		ctx,
		fmt.Sprintf("user:%s", createUser.Id),
		map[string]interface{}{"role": createUser.Role, "status": createUser.Status, "auth_type": createUser.AuthType},
	)
	if err != nil {
		return "", errs.NewError(errs.ErrInternal, err)
	}
	errCommit := tx.Commit()
	if errCommit != nil {
		return "", errs.NewError(errs.ErrInternal, errCommit)
	}

	return sessionId, nil
}

func (s defaultService) Login(ctx context.Context, username, email string) (string, errs.ChatError) {
	u, err := s.userRepo.GetUser(ctx, "username", username)
	if err != nil {
		return "", err
	}
	if u.Email != email {
		return "", errs.NewError(errs.ErrNotAuthorized, nil)
	}

	return s.sessionSrv.CreateSession(
		ctx,
		u.Id,
		map[string]interface{}{"role": u.Role, "status": u.Status, "auth_type": u.AuthType},
	)
}

func NewDefaultService(
	userRepo user.ReaderWriterRepository,
	sessionSrv sessionManager.Service,
) auth.Service {
	return &defaultService{
		userRepo:   userRepo,
		sessionSrv: sessionSrv,
	}
}
