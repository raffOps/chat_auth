package sessionManager

import (
	"context"
	"net/http"
	"time"

	authModels "github.com/raffops/chat_auth/internal/app/auth/model"
	"github.com/raffops/chat_commons/pkg/errs"
	"google.golang.org/grpc"
)

type ReaderRepository interface {
	HashGetEncrypted(ctx context.Context, tableName, key, secret string) (map[string]interface{}, errs.ChatError)
	HashGet(ctx context.Context, tableName, key string, columns ...string) (map[string]interface{}, errs.ChatError)
	StringGet(ctx context.Context, tableName, key string) (string, errs.ChatError)
	GetTTL(ctx context.Context, tableName, key string) (time.Time, errs.ChatError)
	GetKeys(ctx context.Context, pattern string) ([]string, errs.ChatError)
}

type WriterRepository interface {
	HashSetEncrypted(ctx context.Context, tx interface{}, tableName, key, secret string, values map[string]interface{}) errs.ChatError
	HashSet(ctx context.Context, tx interface{}, tableName, key string, values map[string]interface{}) errs.ChatError
	StringSet(ctx context.Context, tx interface{}, tableName, key, value string) errs.ChatError
	ExpireAt(ctx context.Context, tx interface{}, tableName string, key string, at time.Time) errs.ChatError
	Delete(ctx context.Context, tx interface{}, tableName, key string) errs.ChatError
	BeginTransaction(ctx context.Context) (interface{}, errs.ChatError)
	CommitTransaction(ctx context.Context, tx interface{}) errs.ChatError
	RollbackTransaction(ctx context.Context, tx interface{}) errs.ChatError
}

type ReaderWriterRepository interface {
	ReaderRepository
	WriterRepository
}

type Service interface {
	CreateSession(ctx context.Context, userId string, payload map[string]interface{}) (string, errs.ChatError)
	GetSession(ctx context.Context, sessionId string) (map[string]interface{}, errs.ChatError)
	FinishSession(ctx context.Context, sessionId string) errs.ChatError
	FinishUserSessions(ctx context.Context, userId string) errs.ChatError
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
