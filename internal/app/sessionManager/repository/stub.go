package sessionManager

//
//import (
//	"context"
//	"github.com/raffops/auth/internal/app/sessionManager"
//	"github.com/raffops/chat/pkg/errs"
//	"time"
//)
//
//type stubSessionManager struct{}
//
//func (s stubSessionManager) StringSet(ctx context.Context, tx interface{}, tableName, key, value string) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) StringGet(ctx context.Context, tableName, key string) (string, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) Delete(ctx context.Context, tableName, key string) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) HashGet(ctx context.Context, tableName, key string, columns ...string) (map[string]interface{}, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) HashSet(ctx context.Context, tableName, key string, values map[string]interface{}) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) GetTTL(ctx context.Context, tableName, key string) (time.Time, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) SetHashAppend(ctx context.Context, tableName string, key string, value map[string]interface{}) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) HashSetEncrypted(ctx context.Context, tableName, key, secret string, values map[string]interface{}) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) HashGetEncrypted(ctx context.Context, tableName, key, secret string) (map[string]interface{}, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) Hashget(ctx context.Context, tableName, key, field string) (string, errs.ChatError) {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) SetAppend(ctx context.Context, key, value string) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) Hashset(ctx context.Context, tableName string, key string, value map[string]interface{}) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) ExpireAt(ctx context.Context, tableName string, key string, at time.Time) errs.ChatError {
//	//TODO implement me
//	panic("implement me")
//}
//
//func (s stubSessionManager) CreateSession(
//	ctx context.Context,
//	payload map[string]interface{},
//) (string, errs.ChatError) {
//	return "123", nil
//}
//
//func (s stubSessionManager) FinishSession(ctx context.Context, sessionId string) errs.ChatError {
//	return nil
//}
//
//func (s stubSessionManager) CheckSession(ctx context.Context, sessionId string) (bool, errs.ChatError) {
//	return true, nil
//}
//
//func NewStubSessionManager() sessionManager.Repository {
//	return &stubSessionManager{}
//}
