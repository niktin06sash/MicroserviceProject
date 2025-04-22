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
	CreateSession(ctx context.Context, userID string) (*pb.CreateSessionResponse, error)
	DeleteSession(ctx context.Context, sessionID string) (*pb.DeleteSessionResponse, error)
	Close() error
}
type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(cfg configs.SessionServiceConfig) *GrpcClient {
	conn, err := grpc.Dial(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
func (g *GrpcClient) CreateSession(ctx context.Context, userd string) (*pb.CreateSessionResponse, error) {
	traceid := ctx.Value("traceID").(string)
	req := &pb.CreateSessionRequest{UserID: userd}
	resp, err := g.client.CreateSession(ctx, req)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] CreateSession: Request error to GRPC-Session Client: %v", traceid, err)
		return nil, err
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s] CreateSession: Successful request to GRPC-Session Client", traceid)
	return resp, nil
}
func (g *GrpcClient) DeleteSession(ctx context.Context, sessionid string) (*pb.DeleteSessionResponse, error) {
	traceid := ctx.Value("traceID").(string)
	req := &pb.DeleteSessionRequest{SessionID: sessionid}
	resp, err := g.client.DeleteSession(ctx, req)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] [TraceID: %s] DeleteSession: Request error to GRPC-Session Client(DeleteSession): %v", traceid, err)
		return nil, err
	}
	log.Printf("[INFO] [UserManagement] [TraceID: %s] DeleteSession: Successful request to GRPC-Session Client(DeleteSession)", traceid)
	return resp, nil
}
