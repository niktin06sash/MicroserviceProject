package client

import (
	"context"
	"log"

	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source=grpc_client.go -destination=mocks/mock.go
type GrpcClientService interface {
	CreateSession(ctx context.Context, userID string, reqid string) (*pb.CreateSessionResponse, error)
	DeleteSession(ctx context.Context, sessionID string, reqid string) (*pb.DeleteSessionResponse, error)
	Close() error
}
type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(cfg configs.Config) *GrpcClient {
	conn, err := grpc.Dial(cfg.SessionService.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Failed to connect to GRPC-Session Client: %v", err)
		return nil
	}
	client := pb.NewSessionServiceClient(conn)
	log.Println("[INFO] [UserManagement] Successful connect to GRPC-Session Client")
	return &GrpcClient{client: client, conn: conn}
}
func (g *GrpcClient) Close() error {
	return g.conn.Close()
}
func (g *GrpcClient) CreateSession(ctx context.Context, userd string, reqid string) (*pb.CreateSessionResponse, error) {
	req := &pb.CreateSessionRequest{UserID: userd}
	resp, err := g.client.CreateSession(ctx, req)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] CreateSession: Request error to GRPC-Session Client: %v", reqid, err)
		return nil, err
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s] CreateSession: Successful request to GRPC-Session Client", reqid)
	return resp, nil
}
func (g *GrpcClient) DeleteSession(ctx context.Context, sessionid string, reqid string) (*pb.DeleteSessionResponse, error) {
	req := &pb.DeleteSessionRequest{SessionID: sessionid}
	resp, err := g.client.DeleteSession(ctx, req)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [RequestID: %s] DeleteSession: Request error to GRPC-Session Client(DeleteSession): %v", reqid, err)
		return nil, err
	}
	log.Printf("[INFO] [UserManagement] [RequestID: %s] DeleteSession: Successful request to GRPC-Session Client(DeleteSession)", reqid)
	return resp, nil
}
