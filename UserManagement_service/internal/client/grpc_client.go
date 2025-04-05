package client

import (
	"context"
	"log"

	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcServiceClient(cfg configs.Config) *GrpcClient {

	conn, err := grpc.Dial(cfg.SessionService.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}

	client := pb.NewSessionServiceClient(conn)
	log.Println("UserManagement: Successful Connect to GRPC-Session Client")
	return &GrpcClient{client: client, conn: conn}
}
func (g *GrpcClient) Close() error {
	return g.conn.Close()
}
func (g *GrpcClient) CreateSession(ctx context.Context, userd string) (*pb.CreateSessionResponse, error) {
	req := &pb.CreateSessionRequest{UserID: userd}
	resp, err := g.client.CreateSession(ctx, req)
	if err != nil {

		return nil, err
	}
	return resp, nil
}
