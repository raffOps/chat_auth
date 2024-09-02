package sessionManager

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gorilla/mux"
	authModels "github.com/raffops/chat_auth/internal/app/auth/model"
	"github.com/raffops/chat_auth/internal/app/sessionManager"
	"github.com/raffops/chat_auth/internal/app/sessionManager/service"
	userModels "github.com/raffops/chat_auth/internal/app/user/models"
	grpcMock "github.com/raffops/chat_auth/test/mocks/grpc"
	databaseDynamo "github.com/raffops/chat_commons/pkg/database/dynamodb"
	"github.com/raffops/chat_commons/pkg/encryptor"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SessionManagerDynamodbTestSuite struct {
	suite.Suite
	ctx                context.Context
	secret             string
	dynamoConn         *dynamodb.DynamoDB
	defaultEncryptor   encryptor.Encryptor
	sessionRepo        sessionManager.ReaderWriterRepository
	sessionSrv         sessionManager.Service
	johnUser           userModels.User
	johnFirstSession   string
	johnSecondSession  string
	johnSessionTimeout time.Time
}

func (s *SessionManagerDynamodbTestSuite) TestSetupSuite() {
	s.ctx = context.Background()
	dynamoDbTestContainer, err := databaseDynamo.GetDynamoDbTestContainer(s.ctx)
	if err != nil {
		s.T().Fatalf("cannot start dynamodb container: %v", err)
	}
	defer func() {
		err := dynamoDbTestContainer.Terminate(s.ctx)
		if err != nil {
			s.T().Fatalf("cannot stop dynamodb container: %v", err)
		}
	}()
	port, _ := dynamoDbTestContainer.MappedPort(s.ctx, "4566")
	os.Setenv("LOCALSTACK_PORT", port.Port())
	s.T().Logf("dynamodb container started on port %s", port.Port())

	s.dynamoConn = databaseDynamo.GetDynamodbConn(s.ctx)
	defer databaseDynamo.Close(s.dynamoConn)

	s.defaultEncryptor = encryptor.NewDefaultEncryptor()
	//s.sessionRepo, err = sessionRepository.NewDynamodbRepository(s.dynamoConn, s.defaultEncryptor)
	if err != nil {
		s.T().Fatalf("NewDynamodbRepository() error = %v", err)
	}

	timeout, _ := time.ParseDuration(os.Getenv("SESSION_TIMEOUT"))
	s.secret = os.Getenv("SESSION_MANAGER_SECRET")
	s.sessionSrv = service.NewDefaultService(s.sessionRepo, timeout, s.secret)

	s.johnUser = userModels.User{
		Id:       "1",
		Username: "John",
		Email:    "john@doe.com",
		AuthType: userModels.AuthTypeGoogle,
		Role:     authModels.RoleUser,
		Status:   userModels.StatusActive,
	}
	success := s.Run("createJohnFirstSession", s.createJohnFirstSession)
	if !success {
		s.T().Fatalf("createJohnFirstSession() failed")
	}
	success = s.Run("createJohnSecondSession", s.createJohnSecondSession)
	if !success {
		s.T().Fatalf("createJohnSecondSession() failed")
	}

	success = s.Run("checkInvalidSessionId", s.checkInvalidSessionId)
	if !success {
		s.T().Fatalf("checkInvalidSessionId() failed")
	}

	success = s.Run("checkJohnFirstSession", s.checkJohnFirstSession)
	if !success {
		s.T().Fatalf("checkJohnFirstSession() failed")
	}

	success = s.Run("checkJohnFirstSessionWithAdminRole", s.checkJohnFirstSessionWithAdminRole)
	if !success {
		s.T().Fatalf("checkJohnFirstSessionWithAdminRole() failed")
	}

	success = s.Run("refreshJohnFirstSession", s.refreshJohnFirstSession)
	if !success {
		s.T().Fatalf("refreshJohnFirstSession() failed")
	}

	success = s.Run("CheckRestSession", s.CheckRestSession)
	if !success {
		s.T().Fatalf("CheckRestSession() failed")
	}

	success = s.Run("CheckGrpcSessionValidToken", s.CheckGrpcSessionValidToken)
	if !success {
		s.T().Fatalf("CheckGrpcSessionValidToken() failed")
	}

	success = s.Run("CheckGrpcSessionValidTokenButInvalidRole", s.CheckGrpcSessionValidTokenButInvalidRole)
	if !success {
		s.T().Fatalf("CheckGrpcSessionValidTokenButInvalidRole() failed")
	}

	success = s.Run("checkJohnFirstSessionWithExpiredTTL", s.checkJohnFirstSessionWithExpiredTTL)
	if !success {
		s.T().Fatalf("checkJohnFirstSessionWithExpiredTTL() failed")
	}

	success = s.Run("UseJohnFirstSessionAfterTimeout", s.UseJohnFirstSessionAfterTimeout)
	if !success {
		s.T().Fatalf("UseJohnFirstSessionAfterTimeout() failed")
	}
	success = s.Run("checkCorruptedSession", s.checkCorruptedSession)
	if !success {
		s.T().Fatalf("checkCorruptedSession() failed")
	}

	success = s.Run("TestCheckGrpcSession_MissingToken", s.CheckGrpcSessionMissingToken)
	if !success {
		s.T().Fatalf("TestCheckGrpcSession_MissingToken() failed")
	}

	success = s.Run("TestCheckGrpcSession_InvalidToken", s.CheckGrpcSessionInvalidToken)
	if !success {
		s.T().Fatalf("TestCheckGrpcSession_InvalidToken() failed")
	}
}

func (s *SessionManagerDynamodbTestSuite) createJohnFirstSession() {
	id := s.johnUser.Id
	payload := map[string]interface{}{
		"auth_type": int(s.johnUser.AuthType),
		"role":      int(s.johnUser.Role),
		"status":    int(s.johnUser.Status),
		"user_id":   id,
	}

	got, err := s.sessionSrv.CreateSession(s.ctx, id, payload)
	if got == "" {
		s.T().Fatalf("CreateSession() got = %v, want not empty", got)
	}
	if err != nil {
		s.T().Fatalf("CreateSession() error = %v", err)
	}

	result, errEncrypted := s.sessionRepo.HashGet(s.ctx, "session", got, "encrypted_value")
	if errEncrypted != nil {
		s.T().Fatalf("CreateSession() error = %v", errEncrypted)
	}

	payloadString, _ := json.Marshal(payload)
	payloadEncrypted, _ := s.defaultEncryptor.Encrypt(string(payloadString), s.secret)
	encryptedSession := result["encrypted_value"].(string)
	if encryptedSession != payloadEncrypted {
		s.T().Fatalf("CreateSession() got = %v, want %v", encryptedSession, payloadEncrypted)
	}

	s.johnFirstSession = got
}

func (s *SessionManagerDynamodbTestSuite) createJohnSecondSession() {
	id := s.johnUser.Id
	payload := map[string]interface{}{
		"auth_type": int(s.johnUser.AuthType),
		"role":      int(s.johnUser.Role),
		"status":    int(s.johnUser.Status),
		"user_id":   id,
	}
	payloadString, _ := json.Marshal(payload)
	got, err := s.sessionSrv.CreateSession(s.ctx, id, payload)
	if got == "" {
		s.T().Fatalf("CreateSession() got = %v, want not empty", got)
	}
	if err != nil {
		s.T().Fatalf("CreateSession() error = %v", err)
	}

	result, errEncrypted := s.sessionRepo.HashGet(s.ctx, "session", got, "encrypted_value")
	if errEncrypted != nil {
		s.T().Fatalf("CreateSession() error = %v", errEncrypted)
	}

	encryptedSession := result["encrypted_value"].(string)
	payloadEncrypted, _ := s.defaultEncryptor.Encrypt(string(payloadString), s.secret)
	if encryptedSession != payloadEncrypted {
		s.T().Fatalf("CreateSession() got = %v, want %v", encryptedSession, payloadEncrypted)
	}

	s.johnSecondSession = got
}

func (s *SessionManagerDynamodbTestSuite) checkInvalidSessionId() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "invalid")
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world")) //nolint:errcheck
	}
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(f, []authModels.RoleId{s.johnUser.Role}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "Unauthorized\n" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Unauthorized\n")
	}
}

func (s *SessionManagerDynamodbTestSuite) checkJohnFirstSession() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnFirstSession)
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world")) //nolint:errcheck
	}
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(f, []authModels.RoleId{s.johnUser.Role}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "Hello world" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Hello world")
	}
}

func (s *SessionManagerDynamodbTestSuite) checkJohnFirstSessionWithAdminRole() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnFirstSession)
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world")) //nolint:errcheck
	}
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(f, []authModels.RoleId{authModels.RoleAdmin}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusOK)
	}

	if w.Body.String() != "Unauthorized\n" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Hello world")
	}
}

func (s *SessionManagerDynamodbTestSuite) checkJohnFirstSessionWithExpiredTTL() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnFirstSession)
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world")) //nolint:errcheck
	}
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(f, []authModels.RoleId{authModels.RoleUser}))
	time.Sleep(time.Until(s.johnSessionTimeout))

	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusUnauthorized)
	}

	if w.Body.String() != "Unauthorized\n" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Hello world")
	}
}

func (s *SessionManagerDynamodbTestSuite) refreshJohnFirstSession() {
	firstTimeout, _ := s.sessionRepo.GetTTL(s.ctx, "session", s.johnFirstSession)

	time.Sleep(time.Duration(1) * time.Second)
	errRefresh := s.sessionSrv.RefreshSession(s.ctx, s.johnFirstSession)
	if errRefresh != nil {
		s.T().Fatalf("RefreshSession() error = %v", errRefresh)
	}
	secondTimeout, err := s.sessionRepo.GetTTL(s.ctx, "session", s.johnFirstSession)
	if err != nil || firstTimeout.Unix() >= secondTimeout.Unix() {
		s.T().Fatalf("RefreshSession() got = %v, want greater than %v", secondTimeout, firstTimeout)
	}

	s.johnSessionTimeout = secondTimeout
}

func (s *SessionManagerDynamodbTestSuite) CheckRestSession() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnFirstSession)
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(HelloWorldUser, []authModels.RoleId{s.johnUser.Role}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusOK)
	}
	if w.Body.String() != "Hello user" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Hello user")
	}
}

func (s *SessionManagerDynamodbTestSuite) UseJohnFirstSessionAfterTimeout() {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnFirstSession)
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(HelloWorldUser, []authModels.RoleId{s.johnUser.Role}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusUnauthorized)
	}
	if w.Body.String() != "Unauthorized\n" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Unauthorized\n")
	}
}

func (s *SessionManagerDynamodbTestSuite) checkCorruptedSession() {
	errSet := s.sessionRepo.HashSet(s.ctx, nil, "session", s.johnSecondSession, map[string]interface{}{"encrypted_value": "corrupted"})
	if errSet != nil {
		s.T().Fatalf("Set() error = %v", errSet)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+s.johnSecondSession)
	router := mux.NewRouter()
	router.HandleFunc("/", s.sessionSrv.CheckRestSession(HelloWorldUser, []authModels.RoleId{s.johnUser.Role}))
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Code, http.StatusUnauthorized)
	}
	if w.Body.String() != "Unauthorized\n" {
		s.T().Fatalf("CheckRestSession() got = %v, want %v", w.Body.String(), "Unauthorized\n")
	}
}

func (s *SessionManagerDynamodbTestSuite) CheckGrpcSessionMissingToken() {
	md := metadata.New(map[string]string{})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ss := &grpcMock.ServerStream{}
	ss.EXPECT().Context().Return(ctx)
	info := &grpc.StreamServerInfo{}
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := s.sessionSrv.CheckGrpcSession(nil, ss, info, handler)
	s.Equal(codes.PermissionDenied, status.Code(err))
	s.Equal("rpc error: code = PermissionDenied desc = missing token", err.Error())
}

func (s *SessionManagerDynamodbTestSuite) CheckGrpcSessionInvalidToken() {
	md := metadata.New(map[string]string{
		"authorization": "invalid",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ss := &grpcMock.ServerStream{}
	ss.EXPECT().Context().Return(ctx)
	info := &grpc.StreamServerInfo{}
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := s.sessionSrv.CheckGrpcSession(nil, ss, info, handler)
	s.Equal(codes.PermissionDenied, status.Code(err))
	s.Equal("rpc error: code = PermissionDenied desc = invalid token", err.Error())
}

func (s *SessionManagerDynamodbTestSuite) CheckGrpcSessionValidToken() {
	md := metadata.New(map[string]string{
		"authorization": s.johnFirstSession,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ss := &grpcMock.ServerStream{}
	ss.EXPECT().Context().Return(ctx)
	info := &grpc.StreamServerInfo{}
	info.FullMethod = "/test"
	s.sessionSrv.SetRoles("/test", []authModels.RoleId{s.johnUser.Role})
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := s.sessionSrv.CheckGrpcSession(nil, ss, info, handler)
	s.Equal(nil, err)
}

func (s *SessionManagerDynamodbTestSuite) CheckGrpcSessionValidTokenButInvalidRole() {
	md := metadata.New(map[string]string{
		"authorization": s.johnFirstSession,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ss := &grpcMock.ServerStream{}
	ss.EXPECT().Context().Return(ctx)
	info := &grpc.StreamServerInfo{}
	info.FullMethod = "/test"
	s.sessionSrv.SetRoles("/test", []authModels.RoleId{authModels.RoleAdmin})
	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := s.sessionSrv.CheckGrpcSession(nil, ss, info, handler)
	s.Equal(codes.PermissionDenied, status.Code(err))
	s.Equal("rpc error: code = PermissionDenied desc = invalid role", err.Error())
}

func TestSessionManagerDynamo(t *testing.T) {
	os.Setenv("SESSION_MANAGER_SECRET", "7CIuQStxETYG3x0qVO7TcZF7vUNnKlMz")
	os.Setenv("SESSION_TIMEOUT", "6s")
	suite.Run(t, new(SessionManagerDynamodbTestSuite))
}
