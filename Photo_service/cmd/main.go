package main

import (
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/app"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
)

func main() {
	config := configs.LoadConfig()
	application := app.NewPhotoApplication(config)
	if err := application.Start(); err != nil {
		return
	}
}
