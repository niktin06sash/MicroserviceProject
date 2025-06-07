package repository

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/erro"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/model"
	"github.com/t3rm1n4l/go-mega"
)

type MegaClient struct {
	connect      *mega.Mega
	mainfolder   *mega.Node
	wg           *sync.WaitGroup
	progressChan chan int
}

func NewMegaClient(config configs.MegaConfig) (*MegaClient, error) {
	client := mega.New()
	err := client.Login(config.Email, config.Password)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to establish Mega-Client connection: %v", err)
		return nil, err
	}
	node := client.FS.GetRoot()
	var targetFolder *mega.Node
	files, err := client.FS.GetChildren(node)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to get the main directory: %v", err)
		return nil, err
	}
	for _, file := range files {
		if file.GetName() == config.MainDirectory {
			targetFolder = file
			break
		}
	}
	if targetFolder == nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to get the main directory: %v", err)
		return nil, err
	}
	return &MegaClient{
		connect:      client,
		progressChan: make(chan int),
		wg:           &sync.WaitGroup{},
		mainfolder:   targetFolder,
	}, nil
}

const UploadFile = "Repository-UploadFile"
const DeleteFile = "Repository-DeleteFile"

func (client *MegaClient) UploadFile(localfilepath string, photoid string, ext string) *RepositoryResponse {
	const place = UploadFile
	filename := photoid + ext
	client.wg.Add(1)
	go func() {
		defer client.wg.Done()
		for _ = range client.progressChan {
		}
	}()
	uploadedFile, err := client.connect.UploadFile(localfilepath, client.mainfolder, filename, &client.progressChan)
	if err != nil {
		fmterr := fmt.Sprintf("File upload with id = %s error: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	link, err := client.connect.Link(uploadedFile, true)
	if err != nil {
		fmterr := fmt.Sprintf("Error getting a public link to file with id = %s: %v", photoid, err)
		return &RepositoryResponse{Success: false, Errors: map[string]string{erro.ErrorType: erro.ServerErrorType, erro.ErrorMessage: fmterr}, Place: place}
	}
	client.wg.Wait()
	return &RepositoryResponse{Success: true, Data: map[string]any{KeyPhoto: &model.Photo{ID: photoid, ContentType: ext, Size: uploadedFile.GetSize(), CreatedAt: time.Now(), URL: link}}, Place: place}
}
func (client *MegaClient) DeleteFile(id, ext string) *RepositoryResponse {
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
	return &RepositoryResponse{Success: true}
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
