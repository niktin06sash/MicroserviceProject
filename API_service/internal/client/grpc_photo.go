package client

import (
	"context"
	"log"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	pb "github.com/niktin06sash/MicroserviceProject/Photo_service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcPhotoClient struct {
	client pb.PhotoServiceClient
	conn   *grpc.ClientConn
}

func NewGrpcPhotoClient(config configs.PhotoServiceConfig) (*GrpcPhotoClient, error) {
	conn, err := grpc.Dial(config.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("[DEBUG] [API-Service] Failed to connect to GRPC-Photo Client: %v", err)
		return nil, err
	}
	client := pb.NewPhotoServiceClient(conn)
	log.Println("[DEBUG] [API-Service] Successful connect to GRPC-Photo Client")
	return &GrpcPhotoClient{client: client, conn: conn}, nil
}
func (g *GrpcPhotoClient) Close() {
	g.conn.Close()
	log.Println("[DEBUG] [API-Service] Successful close GRPC-Photo Client")
}
func (g *GrpcPhotoClient) LoadPhoto(ctx context.Context, in *pb.LoadPhotoRequest, opts ...grpc.CallOption) (*pb.LoadPhotoResponse, error) {
	return &pb.LoadPhotoResponse{}, nil
}
func (g *GrpcPhotoClient) GetPhoto(ctx context.Context, in *pb.GetPhotoRequest, opts ...grpc.CallOption) (*pb.GetPhotoResponse, error) {
	return &pb.GetPhotoResponse{}, nil
}
func (g *GrpcPhotoClient) GetPhotos(ctx context.Context, in *pb.GetPhotosRequest, opts ...grpc.CallOption) (*pb.GetPhotosResponse, error) {
	return &pb.GetPhotosResponse{}, nil
}
func (g *GrpcPhotoClient) DeletePhoto(ctx context.Context, in *pb.DeletePhotoRequest, opts ...grpc.CallOption) (*pb.DeletePhotoResponse, error) {
	return &pb.DeletePhotoResponse{}, nil
}
