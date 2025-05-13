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
type GrpcClientService interface {
	CreateSession(ctx context.Context, userID string) (*pb.CreateSessionResponse, error)
	DeleteSession(ctx context.Context, sessionID string) (*pb.DeleteSessionResponse, error)
}
type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(cfg configs.SessionServiceConfig) *GrpcClient {
	conn, err := grpc.Dial(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[ERROR] [User-Service] Failed to connect to GRPC-Session Client: %v", err)
		return nil
	}
	client := pb.NewSessionServiceClient(conn)
	log.Println("[INFO] [User-Service] Successful connect to GRPC-Session Client")
	return &GrpcClient{client: client, conn: conn}
}
func (g *GrpcClient) Close() {
	g.conn.Close()
	log.Println("[INFO] [User-Service] Successful close GRPC-Session Client")
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
