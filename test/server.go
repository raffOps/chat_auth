package test

import (
	"github.com/raffops/auth/test/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	pb.UnimplementedTestServer
}

func (s server) TestInterceptor(interceptorServer pb.Test_TestInterceptorServer) error {
	request, err := interceptorServer.Recv()
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	response := pb.TestResponse{Body: request.Body}
	err = interceptorServer.Send(&response)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	return nil
}

func NewTestServer() pb.TestServer {
	return &server{}
}
