package service

import (
	"context"
	"fmt"
	authModels "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/auth/internal/app/sessionManager"
	"github.com/raffops/chat/pkg/errs"
	"github.com/raffops/chat/pkg/logger"
	"github.com/raffops/chat/pkg/uuid"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type service struct {
	repo              sessionManager.ReaderWriterRepository
	timeout           time.Duration
	secret            string
	mapMethodsToRoles map[string][]authModels.RoleId
}

func (s service) FinishUserSessions(ctx context.Context, userId string) errs.ChatError {
	pattern := fmt.Sprintf("user_session:%s:*", userId)
	userSessions, err := s.repo.GetKeys(ctx, pattern)
	if err != nil {
		return err
	}

	sessionRepoTx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer s.repo.RollbackTransaction(ctx, sessionRepoTx)

	prefix := "user_session:" + userId + ":"
	for _, userSession := range userSessions {
		session, _ := strings.CutPrefix(userSession, prefix)
		err = s.repo.Delete(ctx, sessionRepoTx, "session", session)
		if err != nil {
			return err
		}
		err = s.repo.Delete(ctx, sessionRepoTx, "user_session", userId+":"+session)
		if err != nil {
			return err
		}
	}
	return s.repo.CommitTransaction(ctx, sessionRepoTx)
}

func (s service) GetSession(ctx context.Context, sessionId string) (map[string]interface{}, errs.ChatError) {
	return s.repo.HashGetEncrypted(ctx, "session", sessionId, s.secret)
}

func (s service) SetRoles(method string, roles []authModels.RoleId) {
	s.mapMethodsToRoles[method] = roles
}

func (s service) GetRoles(ctx context.Context, method string) ([]authModels.RoleId, errs.ChatError) {
	if roles, ok := s.mapMethodsToRoles[method]; ok {
		return roles, nil
	}
	return nil, errs.NewError(errs.ErrNotFound, fmt.Errorf("method not found"))
}

func sanityCheck() {
	envVariables := []string{
		"SESSION_TIMEOUT",
		"SESSION_MANAGER_SECRET",
	}
	for _, envVariable := range envVariables {
		if _, ok := os.LookupEnv(envVariable); !ok {
			logger.Fatal("Environment variable not set", zap.String("variable", envVariable))
		}
	}
}

func (s service) RefreshSession(ctx context.Context, sessionId string) errs.ChatError {
	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer s.repo.RollbackTransaction(ctx, tx)

	sessionValues, err := s.repo.HashGetEncrypted(ctx, "session", sessionId, s.secret)
	if err != nil {
		return err
	}

	_, ok := sessionValues["user_id"].(string)
	if !ok {
		return errs.NewError(errs.ErrInternal, fmt.Errorf("user_id not found"))
	}

	timeoutAt := time.Now().Add(s.timeout)
	err = s.repo.ExpireAt(ctx, tx, "session", sessionId, timeoutAt)
	if err != nil {
		return err
	}

	return s.repo.CommitTransaction(ctx, tx)
}

func (s service) CreateSession(
	ctx context.Context,
	userId string,
	payload map[string]interface{},
) (string, errs.ChatError) {
	payload["user_id"] = userId
	sessionId := generateRandomSessionId()

	tx, err := s.repo.BeginTransaction(ctx)
	if tx == nil {
		return sessionId, errs.NewError(errs.ErrInternal, fmt.Errorf("transaction is nil"))
	}

	if err != nil {
		return sessionId, err
	}

	defer s.repo.RollbackTransaction(ctx, tx)

	err = s.repo.HashSetEncrypted(ctx, tx, "session", sessionId, s.secret, payload)
	if err != nil {
		return sessionId, err
	}
	err = s.repo.StringSet(ctx, tx, fmt.Sprintf("user_session:%s", userId), sessionId, "")
	if err != nil {
		return sessionId, err
	}
	// TODO maybe async?
	timeoutAt := time.Now().Add(s.timeout)
	err = s.repo.ExpireAt(ctx, tx, "session", sessionId, timeoutAt)
	if err != nil {
		return sessionId, err
	}
	err = s.repo.ExpireAt(ctx, tx, fmt.Sprintf("user_session:%s", userId), sessionId, timeoutAt)
	if err != nil {
		return sessionId, err
	}

	err = s.repo.CommitTransaction(ctx, tx)
	return sessionId, err
}

func (s service) FinishSession(ctx context.Context, sessionId string) errs.ChatError {

	session, err := s.repo.HashGetEncrypted(ctx, "session", sessionId, s.secret)
	if err != nil {
		return err
	}
	userId := session["user_id"].(string)

	tx, err := s.repo.BeginTransaction(ctx)
	if err != nil {
		return err
	}
	defer s.repo.RollbackTransaction(ctx, tx)

	err = s.repo.Delete(ctx, tx, "session", sessionId)
	if err != nil {
		return err
	}
	err = s.repo.Delete(ctx, tx, fmt.Sprintf("user_session:%s", userId), sessionId)
	if err != nil {
		return err
	}

	return s.repo.CommitTransaction(ctx, tx)
}

func NewDefaultService(repo sessionManager.ReaderWriterRepository, timeout time.Duration, secret string) sessionManager.Service {
	sanityCheck()
	return &service{
		repo:              repo,
		timeout:           timeout,
		secret:            secret,
		mapMethodsToRoles: map[string][]authModels.RoleId{},
	}
}

func generateRandomSessionId() string {
	return uuid.GenerateUUID()
}
