package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/t3rm1n4l/go-mega"
)

type MegaClientRepo struct {
	megaclient *CloudObject
}

func NewMegaClientRepo(cl *CloudObject) *MegaClientRepo {
	return &MegaClientRepo{
		megaclient: cl,
	}
}
func (client *MegaClientRepo) UploadFile(ctx context.Context, localfilepath string, photoid string, ext string) *RepositoryResponse {
	const place = UploadFile
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(erro.ContextCanceled)}
	default:
	}
	filename := photoid + ext
	progresschan := make(chan int)
	go func() {
		totalbytes := 0
		for data := range progresschan {
			totalbytes += data
		}
		client.megaclient.progressChan <- totalbytes
	}()
	uploadedFile, err := client.megaclient.connect.UploadFile(localfilepath, client.megaclient.mainfolder, filename, &progresschan)
	if err != nil {
		fmterr := fmt.Sprintf("File upload with id = %s error: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmterr), Place: place}
	}
	link, err := client.megaclient.connect.Link(uploadedFile, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error getting a public link to file with id = %s: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmterr), Place: place}
	}
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(erro.ContextCanceled)}
	case tb := <-client.megaclient.progressChan:
		return &RepositoryResponse{Success: true,
			Data:           Data{Photo: &model.Photo{ID: photoid, ContentType: ext, Size: uploadedFile.GetSize(), CreatedAt: time.Now(), URL: link}},
			Place:          place,
			SuccessMessage: fmt.Sprintf("Photo was successfully uploaded to the cloud (%v bytes uploaded)", tb),
		}
	}
}
func (client *MegaClientRepo) DeleteFile(ctx context.Context, id, ext string) *RepositoryResponse {
	const place = DeleteFile
	filename := id + ext
	file, err := client.findFileByName(ctx, client.megaclient.mainfolder, filename)
	if err != nil {
		fmterr := fmt.Sprintf("Error when receiving a file with id = %s: %v", id, err)
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmterr), Place: place}
	}
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(erro.ContextCanceled)}
	default:
	}
	err = client.megaclient.connect.Delete(file, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error file deleted with id = %s: %v", id, err)
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(fmterr), Place: place}
	}
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: erro.ServerError(erro.ContextCanceled)}
	default:
		return &RepositoryResponse{Success: true, SuccessMessage: "Photo was successfully deleted from cloud", Place: place}
	}
}
func (client *MegaClientRepo) findFileByName(ctx context.Context, node *mega.Node, name string) (*mega.Node, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf(erro.ContextCanceled)
	default:
	}
	children, err := client.megaclient.connect.FS.GetChildren(node)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if child.GetName() == name {
			return child, nil
		}
	}
	return nil, fmt.Errorf("Photo file was not found in directory")
}
