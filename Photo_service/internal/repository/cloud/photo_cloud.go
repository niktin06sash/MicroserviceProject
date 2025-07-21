package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/repository"
	"github.com/t3rm1n4l/go-mega"
)

type PhotoCloud struct {
	cloudclient *CloudObject
}

func NewPhotoCloud(cl *CloudObject) *PhotoCloud {
	return &PhotoCloud{
		cloudclient: cl,
	}
}
func (client *PhotoCloud) UploadFile(ctx context.Context, localfilepath string, photoid string, ext string) *repository.RepositoryResponse {
	const place = UploadFile
	select {
	case <-ctx.Done():
		return repository.BadResponse(erro.ServerError(erro.ContextCanceled), place)
	default:
	}
	filename := photoid + ext
	progresschan := make(chan int)
	go func() {
		totalbytes := 0
		for data := range progresschan {
			totalbytes += data
		}
		client.cloudclient.progressChan <- totalbytes
	}()
	uploadedFile, err := client.cloudclient.connect.UploadFile(localfilepath, client.cloudclient.mainfolder, filename, &progresschan)
	if err != nil {
		fmterr := fmt.Sprintf("File upload with id = %s error: %v", photoid, err)
		return repository.BadResponse(erro.ServerError(fmterr), place)
	}
	link, err := client.cloudclient.connect.Link(uploadedFile, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error getting a public link to file with id = %s: %v", photoid, err)
		return repository.BadResponse(erro.ServerError(fmterr), place)
	}
	select {
	case <-ctx.Done():
		return repository.BadResponse(erro.ServerError(erro.ContextCanceled), place)
	case tb := <-client.cloudclient.progressChan:
		return &repository.RepositoryResponse{Success: true,
			Data:           repository.Data{Photo: &model.Photo{ID: photoid, ContentType: ext, Size: uploadedFile.GetSize(), CreatedAt: time.Now(), URL: link}},
			Place:          place,
			SuccessMessage: fmt.Sprintf("Photo was successfully uploaded to the cloud (%v bytes uploaded)", tb),
		}
	}
}
func (client *PhotoCloud) DeleteFile(ctx context.Context, id, ext string) *repository.RepositoryResponse {
	const place = DeleteFile
	filename := id + ext
	file, err := client.findFileByName(ctx, client.cloudclient.mainfolder, filename)
	if err != nil {
		fmterr := fmt.Sprintf("Error when receiving a file with id = %s: %v", id, err)
		return repository.BadResponse(erro.ServerError(fmterr), place)
	}
	select {
	case <-ctx.Done():
		return repository.BadResponse(erro.ServerError(erro.ContextCanceled), place)
	default:
	}
	err = client.cloudclient.connect.Delete(file, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error file deleted with id = %s: %v", id, err)
		return repository.BadResponse(erro.ServerError(fmterr), place)
	}
	select {
	case <-ctx.Done():
		return repository.BadResponse(erro.ServerError(erro.ContextCanceled), place)
	default:
		return &repository.RepositoryResponse{Success: true, SuccessMessage: "Photo was successfully deleted from cloud", Place: place}
	}
}
func (client *PhotoCloud) findFileByName(ctx context.Context, node *mega.Node, name string) (*mega.Node, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf(erro.ContextCanceled)
	default:
	}
	children, err := client.cloudclient.connect.FS.GetChildren(node)
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
