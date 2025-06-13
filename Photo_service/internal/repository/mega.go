package repository

import (
	"log"
	"sync"

	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
	"github.com/t3rm1n4l/go-mega"
)

const UploadFile = "Repository-UploadFile"
const DeleteFile = "Repository-DeleteFile"

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
