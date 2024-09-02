package service

import (
	"fmt"
	"slices"
	"time"

	auth "github.com/raffops/chat_auth/internal/app/auth/model"
	"github.com/raffops/chat_commons/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// wrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedStream struct {
	grpc.ServerStream
}

func (w *wrappedStream) RecvMsg(m any) error {
	logger.Debug("Receive a message",
		zap.String("Type", fmt.Sprintf("%T", m)),
		zap.String("Time", time.Now().Format(time.RFC3339)),
	)
	return w.ServerStream.RecvMsg(m)
}

func (w *wrappedStream) SendMsg(m any) error {
	logger.Debug("Send a message",
		zap.String("Type", fmt.Sprintf("%T", m)),
		zap.String("Time", time.Now().Format(time.RFC3339)),
	)
	return w.ServerStream.SendMsg(m)
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s}
}

func (s service) CheckGrpcSession(
	srv any,
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	// authentication (token verification)
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		return status.Errorf(codes.PermissionDenied, "missing metadata")
	}
	token := md["authorization"]
	if len(token) == 0 {
		return status.Errorf(codes.PermissionDenied, "missing token")
	}
	result, err := s.repo.HashGetEncrypted(ss.Context(), "session", token[0], s.secret)
	if err != nil {
		return status.Errorf(codes.PermissionDenied, "invalid token")
	}

	if slices.Contains(
		s.mapMethodsToRoles[info.FullMethod],
		auth.RoleId(int(result["role"].(float64))),
	) {
		err := handler(srv, newWrappedStream(ss))
		return err
	}
	return status.Errorf(codes.PermissionDenied, "invalid role")
}
