API_Service_DIR := API_service
Session_Service_DIR := SessionManagement_service
User_Service_DIR := UserManagement_service
Logs_Service_DIR := Logs_service
Photo_Service_DIR := Photo_service
Swagger_DIR := $(API_Service_DIR)/docs

.PHONY: all start tests stop clean swagger redis kafka run shutdown dockerstart dockerstop dockerrun dockershutdown

all: start

swagger:
	@echo "Generating Swagger documentation for API Service..."
	cd $(API_Service_DIR) && swag init --pd -d ./cmd,./internal/handlers,./internal/handlers/response
	@echo "Swagger documentation generated in $(Swagger_DIR)."

dockerstart: 
	@echo "Starting all services..."
	docker-compose up -d

dockerstop:
	@echo "Stopping all services..."
	docker-compose stop kafka
	docker-compose stop zookeeper
	docker-compose stop
	@echo "All services stopped."
dockerrun: swagger dockerstart
dockershutdown: dockerstop clean

kafka:
	@echo "Starting Zookeeper..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd C:\Users\nikit\kafka; .\bin\windows\zookeeper-server-start.bat .\config\zookeeper.properties'"
	@echo "Waiting for Zookeeper to initialize..."
	powershell -Command "Start-Sleep -Seconds 5"
	@echo "Starting Kafka..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd C:\Users\nikit\kafka; .\bin\windows\kafka-server-start.bat .\config\server.properties'"
	powershell -Command "Start-Sleep -Seconds 5"
prometheus:
	@echo "Starting Prometheus..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd C:\Users\nikit\prometheus; .\prometheus.exe --config.file=prometheus.yml'"
redis:
	@echo "Starting Redis CLI..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'wsl redis-cli'"
start:
	@echo "Starting Logs Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Logs_Service_DIR); go run cmd/main.go'"
	@echo "Starting API-Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(API_Service_DIR); go run cmd/main.go'"
	@echo "Starting Session-Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Session_Service_DIR); go run cmd/main.go'"
	@echo "Starting User-Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(User_Service_DIR); go run cmd/main.go'"
	@echo "Starting Photo-Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Photo_Service_DIR); go run cmd/main.go'"
	@echo "All services started."

stop:
	@echo "Stopping all services..."
	taskkill /F /IM go.exe || true
	taskkill /F /IM java.exe || true
	taskkill /IM prometheus.exe /F
	wsl pkill redis-cli || true
	@echo "All services stopped."

clean:
	@echo "Cleaning up..."
	powershell -Command "Remove-Item -Recurse -Force $(Swagger_DIR)" || true
	@echo "Cleanup complete."
tests:
	@echo "Starting tests in User-Service...."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(User_Service_DIR)/internal/service/; go test -v ./...'"
run: kafka swagger prometheus redis start
shutdown: stop clean
