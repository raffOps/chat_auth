package sessionManager

import (
	"context"
	"github.com/raffops/auth/internal/app/sessionManager"
	"github.com/raffops/chat/pkg/errs"
	"time"
)

type stubSessionManager struct{}

func (s stubSessionManager) SetEncryptedHashmap(ctx context.Context, key, secretKey string, value map[string]interface{}) errs.ChatError {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) GetEncryptedHashmap(ctx context.Context, key, secretKey string) (map[string]interface{}, errs.ChatError) {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) Hashget(ctx context.Context, key, field string) (string, errs.ChatError) {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) SetAppend(ctx context.Context, key, value string) errs.ChatError {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) Hashset(ctx context.Context, key string, value map[string]interface{}) errs.ChatError {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) ExpireAt(ctx context.Context, key string, at time.Time) errs.ChatError {
	//TODO implement me
	panic("implement me")
}

func (s stubSessionManager) CreateSession(
	ctx context.Context,
	payload map[string]interface{},
) (string, errs.ChatError) {
	return "123", nil
}

func (s stubSessionManager) FinishSession(ctx context.Context, sessionId string) errs.ChatError {
	return nil
}

func (s stubSessionManager) CheckSession(ctx context.Context, sessionId string) (bool, errs.ChatError) {
	return true, nil
}

func NewStubSessionManager() sessionManager.Repository {
	return &stubSessionManager{}
}
