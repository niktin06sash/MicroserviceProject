API_Service_DIR := API_service
Session_Service_DIR := SessionManagement_service
User_Service_DIR := UserManagement_service
Kafka_Service_DIR := Kafka_service
Swagger_DIR := $(API_Service_DIR)/docs

.PHONY: all start stop clean swagger redis kafka run shutdown

all: start

swagger:
	@echo "Generating Swagger documentation for API Service..."
	cd $(API_Service_DIR) && swag init --pd -d ./cmd,./internal/handlers,./internal/handlers/response
	@echo "Swagger documentation generated in $(Swagger_DIR)."
kafka:
	@echo "Starting Zookeeper..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd C:\Users\nikit\kafka; .\bin\windows\zookeeper-server-start.bat .\config\zookeeper.properties'"
	@echo "Waiting for Zookeeper to initialize..."
	timeout /T 5 >nul
	@echo "Starting Kafka..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd C:\Users\nikit\kafka; .\bin\windows\kafka-server-start.bat .\config\server.properties'"
redis:
	@echo "Starting Redis CLI..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'wsl redis-cli'"
start:
	@echo "Starting Kafka Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Kafka_Service_DIR); go run cmd/main.go'"
	@echo "Starting API Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(API_Service_DIR); go run cmd/main.go'"
	@echo "Starting Session Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Session_Service_DIR); go run cmd/main.go'"
	@echo "Starting User Service..."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(User_Service_DIR); go run cmd/main.go'"
	@echo "All services started."

stop:
	@echo "Stopping all services..."
	taskkill /F /IM go.exe || true
	taskkill /F /IM java.exe || true
	wsl pkill redis-cli || true
	@echo "All services stopped."

clean:
	@echo "Cleaning up..."
	powershell -Command "Remove-Item -Recurse -Force $(Swagger_DIR)" || true
	@echo "Cleanup complete."

run: swagger kafka redis start

shutdown: stop clean