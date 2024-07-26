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
	"time"
)

type service struct {
	repo              sessionManager.Repository
	timeout           time.Duration
	secret            string
	mapMethodsToRoles map[string][]authModels.RoleId
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
	sessionValues, err := s.repo.GetEncryptedHashmap(ctx, sessionId, s.secret)
	if err != nil {
		return err
	}

	userId, ok := sessionValues["userId"].(string)
	if !ok {
		return errs.NewError(errs.ErrInternal, fmt.Errorf("userId not found"))
	}

	timeoutAt := time.Now().Add(s.timeout)
	err = s.repo.ExpireAt(ctx, "session:"+sessionId, timeoutAt)
	if err != nil {
		return err
	}

	err = s.repo.Hashset(
		ctx,
		fmt.Sprintf("user:%s", userId),
		map[string]interface{}{
			"session:" + sessionId: timeoutAt.Unix(),
		},
	)
	return err
}

func (s service) CreateSession(
	ctx context.Context,
	id string,
	payload map[string]interface{},
) (string, errs.ChatError) {
	sessionId := generateRandomSessionId()
	timeoutAt := time.Now().Add(s.timeout)
	err := s.repo.SetEncryptedHashmap(ctx, "session:"+sessionId, s.secret, payload)
	if err != nil {
		return sessionId, err
	}

	err = s.repo.ExpireAt(ctx, sessionId, timeoutAt)
	if err != nil {
		return sessionId, err
	}

	err = s.repo.Hashset(
		ctx,
		fmt.Sprintf("user:%s", id),
		map[string]interface{}{
			"session:" + sessionId: timeoutAt.Unix(),
		},
	)

	return sessionId, err
}

func (s service) FinishSession(ctx context.Context, id string) errs.ChatError {
	//TODO implement me
	panic("implement me")
}

func NewDefaultService(repo sessionManager.Repository, timeout time.Duration, secret string) sessionManager.Service {
	sanityCheck()
	return &service{
		repo:              repo,
		timeout:           timeout,
		secret:            secret,
		mapMethodsToRoles: map[string][]authModels.RoleId{},
	}
}

func generateRandomSessionId() string {
	return fmt.Sprintf("%s", uuid.GenerateUUID())
}
