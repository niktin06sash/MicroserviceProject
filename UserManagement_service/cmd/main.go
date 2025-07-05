package main

import (
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/app"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
)

func main() {
	config := configs.LoadConfig()
	application := app.NewUserApplication(config)
	if err := application.Start(); err != nil {
		return
	}
}
