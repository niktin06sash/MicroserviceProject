package main

import (
	"github.com/niktin06sash/MicroserviceProject/Logs_service/internal/app"
	configs "github.com/niktin06sash/MicroserviceProject/Logs_service/internal/configs"
)

func main() {
	config := configs.LoadConfig()
	application := app.NewLogsApplication(config)
	if err := application.Start(); err != nil {
		return
	}
}
