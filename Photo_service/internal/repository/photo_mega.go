package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/t3rm1n4l/go-mega"
)

func (client *MegaClient) UploadFile(ctx context.Context, localfilepath string, photoid string, ext string) *RepositoryResponse {
	const place = UploadFile
	filename := photoid + ext
	progresschan := make(chan int)
	client.wg.Add(1)
	go func() {
		defer client.wg.Done()
		totalbytes := 0
		for data := range progresschan {
			totalbytes += data
		}
		client.progressChan <- totalbytes
	}()
	uploadedFile, err := client.connect.UploadFile(localfilepath, client.mainfolder, filename, &progresschan)
	if err != nil {
		fmterr := fmt.Sprintf("File upload with id = %s error: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	link, err := client.connect.Link(uploadedFile, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error getting a public link to file with id = %s: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: "Context canceled or timeout"}}
	case tb := <-client.progressChan:
		return &RepositoryResponse{Success: true,
			Data:           map[string]any{KeyPhoto: &model.Photo{ID: photoid, ContentType: ext, Size: uploadedFile.GetSize(), CreatedAt: time.Now(), URL: link}},
			Place:          place,
			SuccessMessage: fmt.Sprintf("Photo was successfully uploaded to the cloud (%v bytes uploaded)", tb),
		}
	}
}
func (client *MegaClient) DeleteFile(ctx context.Context, id, ext string) *RepositoryResponse {
	const place = DeleteFile
	filename := id + ext
	file, err := client.findFileByName(client.mainfolder, filename)
	if err != nil {
		fmterr := fmt.Sprintf("Error when receiving a file with id = %s: %v", id, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	err = client.connect.Delete(file, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error file deleted with id = %s: %v", id, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	select {
	case <-ctx.Done():
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: "Context canceled or timeout"}}
	default:
		return &RepositoryResponse{Success: true, SuccessMessage: "Photo was successfully deleted from cloud"}
	}
}
func (client *MegaClient) findFileByName(node *mega.Node, name string) (*mega.Node, error) {
	children, err := client.connect.FS.GetChildren(node)
	if err != nil {
		return nil, err
	}
	for _, child := range children {
		if child.GetName() == name {
			return child, nil
		}
	}
	return nil, err
}
