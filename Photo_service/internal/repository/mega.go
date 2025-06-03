package repository

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
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
func (client *MegaClient) UploadFile(localfilepath string, id string) *model.Photo {
	ext := filepath.Ext(localfilepath)
	fileName := id + ext
	client.wg.Add(1)
	go func() {
		defer client.wg.Done()
		for _ = range client.progressChan {

		}
	}()
	uploadedFile, err := client.connect.UploadFile(localfilepath, client.mainfolder, fileName, &client.progressChan)
	if err != nil {
		log.Printf("File upload with name = %s error: %v", fileName, err)
		return nil
	}
	link, err := client.connect.Link(uploadedFile, true)
	if err != nil {
		log.Fatalf("Error getting a public link to file with name = %s: %v", fileName, err)
		return nil
	}
	client.wg.Wait()
	log.Printf("File upload with name = %s succesfully completed", fileName)
	return &model.Photo{ID: id, Size: uploadedFile.GetSize(), ContentType: ext, CreatedAt: time.Now(), URL: link}
}
func (client *MegaClient) DeleteFile(id, contenttype string) {
	filename := id + contenttype
	file := client.findFileByName(client.mainfolder, filename)
	if file == nil {
		return
	}
	err := client.connect.Delete(file, true)
	if err != nil {
		log.Printf("File deleted with name = %s error: %v", filename, err)
		return
	}
	log.Printf("File deleted with name = %s succesfully completed", filename)
}
func (client *MegaClient) findFileByName(node *mega.Node, name string) *mega.Node {
	children, err := client.connect.FS.GetChildren(node)
	if err != nil {
		fmt.Printf("Error when receiving a file with name = %s: %v", name, err)
		return nil
	}
	for _, child := range children {
		if child.GetName() == name {
			return child
		}
	}
	return nil
}
