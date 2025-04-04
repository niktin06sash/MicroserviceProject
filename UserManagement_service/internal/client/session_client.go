package client

import (
	pb "SessionManagement_service/proto"

	"google.golang.org/grpc"
)

type SessionClient struct {
	client  pb.SessionServiceClient
	conn    *grpc.ClientConn
	address string
}
