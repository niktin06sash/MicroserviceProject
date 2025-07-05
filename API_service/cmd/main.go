package main

import (
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/app"
	"github.com/niktin06sash/MicroserviceProject/API_service/internal/configs"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/middleware"
	_ "github.com/niktin06sash/MicroserviceProject/API_service/internal/handlers/response"
)

// @title API-Gateway
// @version 1.0
// @description This is a sample API-server for managing users.
// @host localhost:8083
// @BasePath /
// @schemes http

func main() {
	config := configs.LoadConfig()
	application := app.NewAPIApplication(config)
	if err := application.Start(); err != nil {
		return
	}
}
