package client

import (
	"context"
	"log"

	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source=grpc_client.go -destination=mocks/mock.go
type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(cfg configs.SessionServiceConfig) (*GrpcClient, error) {
	conn, err := grpc.Dial(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to connect to GRPC-Session Client: %v", err)
		return nil, err
	}
	client := pb.NewSessionServiceClient(conn)
	log.Println("[DEBUG] [User-Service] Successful connect to GRPC-Session Client")
	return &GrpcClient{client: client, conn: conn}, nil
}
func (g *GrpcClient) Close() {
	g.conn.Close()
	log.Println("[DEBUG] [User-Service] Successful close GRPC-Session Client")
}
func (g *GrpcClient) CreateSession(ctx context.Context, userd string) (*pb.CreateSessionResponse, error) {
	req := &pb.CreateSessionRequest{UserID: userd}
	metrics.UserBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
	return g.client.CreateSession(ctx, req)
}
func (g *GrpcClient) DeleteSession(ctx context.Context, sessionid string) (*pb.DeleteSessionResponse, error) {
	req := &pb.DeleteSessionRequest{SessionID: sessionid}
	metrics.UserBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
	return g.client.DeleteSession(ctx, req)
}
