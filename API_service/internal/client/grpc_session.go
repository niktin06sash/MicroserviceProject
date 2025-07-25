package client

import (
	"context"
	"log"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcSessionClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcSessionClient(cfg configs.SessionServiceConfig) (*GrpcSessionClient, error) {
	conn, err := grpc.Dial(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[DEBUG] [API-Service] Failed to connect to GRPC-Session Client: %v", err)
		return nil, err
	}
	client := pb.NewSessionServiceClient(conn)
	log.Println("[DEBUG] [API-Service] Successful connect to GRPC-Session Client")
	return &GrpcSessionClient{client: client, conn: conn}, nil
}
func (g *GrpcSessionClient) Close() {
	g.conn.Close()
	log.Println("[DEBUG] [API-Service] Successful close GRPC-Session Client")
}
func (g *GrpcSessionClient) ValidateSession(ctx context.Context, sessionid string) (*pb.ValidateSessionResponse, error) {
	req := &pb.ValidateSessionRequest{SessionID: sessionid}
	metrics.APIBackendRequestsTotal.WithLabelValues("Session-Service").Inc()
	return g.client.ValidateSession(ctx, req)
}
