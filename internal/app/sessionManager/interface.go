package sessionManager

import (
	"context"
	authModels "github.com/raffops/auth/internal/app/auth/model"
	"github.com/raffops/chat/pkg/errs"
	"google.golang.org/grpc"
	"net/http"
	"time"
)

type Repository interface {
	SetAppend(ctx context.Context, key, value string) errs.ChatError
	Hashset(ctx context.Context, key string, value map[string]interface{}) errs.ChatError
	SetEncryptedHashmap(ctx context.Context, key, secretKey string, value map[string]interface{}) errs.ChatError
	GetEncryptedHashmap(ctx context.Context, key, secretKey string) (map[string]interface{}, errs.ChatError)
	Hashget(ctx context.Context, key, field string) (string, errs.ChatError)
	ExpireAt(ctx context.Context, key string, at time.Time) errs.ChatError
}

type Service interface {
	CreateSession(ctx context.Context, id string, payload map[string]interface{}) (string, errs.ChatError)
	FinishSession(ctx context.Context, id string) errs.ChatError
	RefreshSession(ctx context.Context, sessionId string) errs.ChatError
	CheckRestSession(next http.HandlerFunc, roles []authModels.RoleId) http.HandlerFunc
	CheckGrpcSession(
		srv any,
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error
	SetRoles(method string, roles []authModels.RoleId)
	GetRoles(ctx context.Context, method string) ([]authModels.RoleId, errs.ChatError)
}
