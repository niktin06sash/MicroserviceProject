package client

import (
	"context"
	"log"

	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/metrics"
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
func (g *GrpcPhotoClient) LoadPhoto(ctx context.Context, userid string, filedata []byte) (*pb.LoadPhotoResponse, error) {
	req := &pb.LoadPhotoRequest{UserId: userid, FileData: filedata}
	metrics.APIBackendRequestsTotal.WithLabelValues("Photo-Service").Inc()
	return g.client.LoadPhoto(ctx, req)
}
func (g *GrpcPhotoClient) GetPhoto(ctx context.Context, userid string, photoid string) (*pb.GetPhotoResponse, error) {
	req := &pb.GetPhotoRequest{UserId: userid, PhotoId: photoid}
	metrics.APIBackendRequestsTotal.WithLabelValues("Photo-Service").Inc()
	return g.client.GetPhoto(ctx, req)
}
func (g *GrpcPhotoClient) GetPhotos(ctx context.Context, userid string) (*pb.GetPhotosResponse, error) {
	req := &pb.GetPhotosRequest{UserId: userid}
	metrics.APIBackendRequestsTotal.WithLabelValues("Photo-Service").Inc()
	return g.client.GetPhotos(ctx, req)
}
func (g *GrpcPhotoClient) DeletePhoto(ctx context.Context, userid string, photoid string) (*pb.DeletePhotoResponse, error) {
	req := &pb.DeletePhotoRequest{UserId: userid, PhotoId: photoid}
	metrics.APIBackendRequestsTotal.WithLabelValues("Photo-Service").Inc()
	return g.client.DeletePhoto(ctx, req)
}
