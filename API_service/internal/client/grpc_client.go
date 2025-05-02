package client

import (
	"context"
	"log"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	pb "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GrpcClientService interface {
	ValidateSession(ctx context.Context, sessionid string) (*pb.ValidateSessionResponse, error)
	DeleteSession(ctx context.Context, sessionID string) (*pb.DeleteSessionResponse, error)
	Close()
}
type GrpcClient struct {
	client pb.SessionServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcClient(cfg configs.SessionServiceConfig) *GrpcClient {
	conn, err := grpc.Dial(cfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[ERROR] [API-Service] [gRPC-Client] Failed to connect to GRPC-Session Client: %v", err)
		return nil
	}
	client := pb.NewSessionServiceClient(conn)
	log.Println("[INFO] [API-Service] [gRPC-Client] Successful connect to GRPC-Session Client")
	return &GrpcClient{client: client, conn: conn}
}
func (g *GrpcClient) Close() {
	g.conn.Close()
	log.Println("[INFO] [API-Service] [gRPC-Client] Successful close GRPC-Session Client")
}
func (g *GrpcClient) ValidateSession(ctx context.Context, sessionid string) (*pb.ValidateSessionResponse, error) {
	traceid := ctx.Value("traceID").(string)
	req := &pb.ValidateSessionRequest{SessionID: sessionid}
	resp, err := g.client.ValidateSession(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			log.Printf("[ERROR] [API-Service] [TraceID: %s] CreateSession: Request error to GRPC-Session Client: %v", traceid, err)
		}
		return nil, err
	}
	log.Printf("[INFO] [API-Service] [TraceID: %s] CreateSession: Successful request to GRPC-Session Client", traceid)
	return resp, nil
}
func (g *GrpcClient) DeleteSession(ctx context.Context, sessionid string) (*pb.DeleteSessionResponse, error) {
	traceid := ctx.Value("traceID").(string)
	req := &pb.DeleteSessionRequest{SessionID: sessionid}
	resp, err := g.client.DeleteSession(ctx, req)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() != codes.InvalidArgument {
			log.Printf("[ERROR] [API-Service] [TraceID: %s] DeleteSession: Request error to GRPC-Session Client: %v", traceid, err)
		}
		return nil, err
	}
	log.Printf("[INFO] [API-Service] [TraceID: %s] DeleteSession: Successful request to GRPC-Session Client", traceid)
	return resp, nil
}
