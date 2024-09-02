package auth

import (
	"context"

	"github.com/raffops/chat_auth/internal/app/auth"
	authModels "github.com/raffops/chat_auth/internal/app/auth/model"
	"github.com/raffops/chat_auth/internal/app/sessionManager"
	"github.com/raffops/chat_auth/internal/app/user"
	userModels "github.com/raffops/chat_auth/internal/app/user/models"
	"github.com/raffops/chat_commons/pkg/errs"
)

type defaultService struct {
	userRepo    user.ReaderWriterRepository
	sessionRepo sessionManager.ReaderRepository
	sessionSrv  sessionManager.Service
}

func (s defaultService) DeleteUser(ctx context.Context, userToDelete userModels.User) errs.ChatError {
	err := s.sessionSrv.FinishUserSessions(ctx, userToDelete.Id)
	if err != nil {
		return err
	}

	return nil
}

func (s defaultService) Logout(ctx context.Context, sessionId string) errs.ChatError {
	return s.sessionSrv.FinishSession(ctx, sessionId)
}

func (s defaultService) SignUp(
	ctx context.Context,
	username, email string,
	authType userModels.AuthTypeId,
	role authModels.RoleId,
) (string, errs.ChatError) {
	u := userModels.User{
		Username: username,
		Email:    email,
		AuthType: authType,
		Role:     role,
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
		createUser.Id,
		map[string]interface{}{
			"role":      createUser.Role,
			"status":    createUser.Status,
			"auth_type": createUser.AuthType},
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

func (s defaultService) Refresh(ctx context.Context, sessionId string) errs.ChatError {
	return s.sessionSrv.RefreshSession(ctx, sessionId)
}

func NewDefaultService(
	userRepo user.ReaderWriterRepository,
	sessionRepo sessionManager.ReaderRepository,
	sessionSrv sessionManager.Service,
) auth.Service {
	return &defaultService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		sessionSrv:  sessionSrv,
	}
}
