package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"

	_ "github.com/lib/pq"
)

type DBInterface interface {
	Open(driverName, connectionString string) (*sql.DB, error)
	Ping(db *sql.DB) error
	Close(db *sql.DB)
	SetConfig(cfg DBConfig)
}

type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type DBObject struct {
	dbConfig DBConfig
}

func (d *DBObject) SetConfig(cfg DBConfig) {
	d.dbConfig = cfg
}

func (d *DBObject) Open(driverName string, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		log.Printf("Sql-Open error %v", err)
		return nil, err
	}
	return db, nil
}

func (d *DBObject) Ping(db *sql.DB) error {
	err := db.Ping()
	if err != nil {
		log.Printf("Sql-Ping error %v", err)
		return err
	}
	return nil
}

func (d *DBObject) Close(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Printf("Sql-Close error %v", err)
	}
}

func ConnectToDb(cfg configs.Config) (*sql.DB, DBInterface, error) {
	dbInterface := &DBObject{}

	dbConfig := DBConfig{
		Driver:   cfg.Database.Driver,
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}
	dbInterface.SetConfig(dbConfig)

	connectionString := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name, dbConfig.SSLMode)

	db, err := dbInterface.Open(dbConfig.Driver, connectionString)
	if err != nil {

		return nil, nil, err
	}
	err = dbInterface.Ping(db)
	if err != nil {
		dbInterface.Close(db)

		return nil, nil, err
	}
	log.Println("UserManagement: Successful connect to Postgre-Client!")
	return db, dbInterface, nil
}
