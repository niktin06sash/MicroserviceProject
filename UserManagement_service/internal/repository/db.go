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
	Ping(db *sql.DB) error
	Close(db *sql.DB)
}

type DBObject struct{}

func (d *DBObject) Open(driverName, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Sql-Open error: %v", err)
		return nil, erro.ErrorDbOpen
	}
	return db, nil
}

func (d *DBObject) Ping(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Sql-Ping error: %v", err)
		return erro.ErrorDbPing
	}
	return nil
}

func (d *DBObject) Close(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Printf("[ERROR] [UserManagement] Sql-Close error: %v", err)
	}
}

func BuildConnectionString(cfg configs.DatabaseConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
}

func ConnectToDb(cfg configs.Config, dbInterface DBInterface) (*sql.DB, error) {
	connectionString := BuildConnectionString(cfg.Database)

	db, err := dbInterface.Open(cfg.Database.Driver, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = dbInterface.Ping(db)
	if err != nil {
		dbInterface.Close(db)
		return nil, fmt.Errorf("failed to establish database connection: %w", err)
	}

	log.Println("[INFO] [UserManagement] Successful connect to Postgre-Client!")
	return db, nil
}
