package main

import (
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/app"
	"github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/configs"
)

func main() {
	config := configs.LoadConfig()
	application := app.NewSessionApplication(config)
	if err := application.Start(); err != nil {
		return
	}
}
