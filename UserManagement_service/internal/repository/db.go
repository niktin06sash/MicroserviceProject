package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"
	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/erro"

	_ "github.com/lib/pq"
)

type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type DBInterface interface {
	Open(driverName, connectionString string) (*sql.DB, error)
	Ping() error
	Close()
}

func NewDatabaseConnection(cfg configs.DatabaseConfig) (*DBObject, error) {
	dbObject := &DBObject{}
	connectionString := buildConnectionString(cfg)
	err := dbObject.Open(cfg.Driver, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = dbObject.Ping()
	if err != nil {
		dbObject.Close()
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	log.Println("[INFO] [UserManagement] Successful connect to Postgre-Client!")
	return dbObject, nil
}

type DBObject struct {
	DB *sql.DB
}

func (d *DBObject) Open(driverName, connectionString string) error {
	var err error
	d.DB, err = sql.Open(driverName, connectionString)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Postgre-Client-Open error: %v", err)
		return erro.ErrorDbOpen
	}
	return nil
}

func (d *DBObject) Close() {
	err := d.DB.Close()
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Postgre-Client-Close error: %v", err)
	}
	log.Println("[INFO] [UserManagement] Successful close Postgre-Client")
}

func (d *DBObject) Ping() error {
	err := d.DB.Ping()
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Postgre-Client-Ping error: %v", err)
		return erro.ErrorDbPing
	}
	return nil
}
func buildConnectionString(cfg configs.DatabaseConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
}
