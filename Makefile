API_Service_DIR := API_service
Session_Service_DIR := SessionManagement_service
User_Service_DIR := UserManagement_service
Logs_Service_DIR := Logs_service
Photo_Service_DIR := Photo_service
Swagger_DIR := $(API_Service_DIR)/docs

.PHONY: all start tests stop clean swagger run shutdown dockerstop dockerrun

all: start

swagger:
	@echo "Generating Swagger documentation for API Service..."
	cd $(API_Service_DIR) && swag init --pd -d ./cmd,./internal/handlers,./internal/handlers/response
	@echo "Swagger documentation generated in $(Swagger_DIR)."

dockerrun: 
	@echo "Starting all docker services..."
	docker-compose up -d
migrate-up:
	@echo "Applying User-Service migrations..."
	migrate -path ./UserManagement_service/schema -database "postgres://user:postgres@localhost:5434/users?sslmode=disable" up
	@echo "Applying Photo-Service migrations..."
	migrate -path ./Photo_service/schema -database "postgres://user:postgres@localhost:5433/photos?sslmode=disable" up
migrate-down:
	@echo "Rolling back Photo-Service migrations..."
	migrate -path ./Photo_service/schema -database "postgres://user:postgres@localhost:5433/photos?sslmode=disable" down
	@echo "Rolling back User-Service migrations..."
	migrate -path ./UserManagement_service/schema -database "postgres://user:postgres@localhost:5434/users?sslmode=disable" down
dockerstop:
	@echo "Stopping all docker services..."
	docker-compose stop kafka
	docker-compose stop zookeeper
	docker-compose stop
	@echo "All docker services stopped."
start:
	@echo "Building and starting Logs Service..."
    cd $(Logs_Service_DIR) && go build -o logs_service.exe cmd/main.go
    Start-Process powershell -ArgumentList '-NoExit', './logs_service.exe'
	@echo "Logs Service started."
	@echo "Building and starting API-Service..."
    cd $(API_Service_DIR) && go build -o api_service.exe cmd/main.go
    Start-Process powershell -ArgumentList '-NoExit', './api_service.exe'
	@echo "API-Service started."
	@echo "Building and starting Session-Service..."
    cd $(Session_Service_DIR) && go build -o session_service.exe cmd/main.go
    Start-Process powershell -ArgumentList '-NoExit', './session_service.exe'
	@echo "Session-Service started."
	@echo "Building and starting User-Service..."
    cd $(User_Service_DIR) && go build -o user_service.exe cmd/main.go
    Start-Process powershell -ArgumentList '-NoExit', './user_service.exe'
	@echo "User-Service started."
	@echo "Building and starting Photo-Service..."
    cd $(Photo_Service_DIR) && go build -o photo_service.exe cmd/main.go
    Start-Process powershell -ArgumentList '-NoExit', './photo_service.exe'
	@echo "Photo-Service started."
	@echo "All local services built and started."
stop:
	@echo "Stopping local services..."
    powershell -Command "Get-Process | Where-Object { $$_.Path -like '$$PWD\*' } | Stop-Process"
clean:
	@echo "Cleaning up..."
	powershell -Command "Remove-Item -Recurse -Force $(Swagger_DIR)" || true
	rm -f $(Logs_Service_DIR)/logs_service.exe
    rm -f $(API_Service_DIR)/api_service.exe
    rm -f $(Session_Service_DIR)/session_service.exe
    rm -f $(User_Service_DIR)/user_service.exe
    rm -f $(Photo_Service_DIR)/photo_service.exe
	@echo "Cleanup complete."
tests:
	@echo "Starting tests in User-Service...."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(User_Service_DIR)/internal/service/; go test -v ./...'"
	@echo "Starting tests in Session-Service...."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Session_Service_DIR)/internal/service/; go test -v ./...'"
	@echo "Starting tests in Photo-Service...."
	powershell -Command "Start-Process powershell -ArgumentList '-NoExit', 'cd $(Photo_Service_DIR)/internal/service/; go test -v ./...'"
run: tests migrate-up swagger start
shutdown: stop clean