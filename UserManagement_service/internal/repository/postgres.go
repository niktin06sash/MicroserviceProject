package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"

	_ "github.com/lib/pq"
)

func NewDatabaseConnection(cfg configs.DatabaseConfig) (*DBObject, error) {
	dbObject := &DBObject{}
	connectionString := buildConnectionString(cfg)
	err := dbObject.Open(cfg.Driver, connectionString)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}

	err = dbObject.Ping()
	if err != nil {
		dbObject.Close()
		return nil, fmt.Errorf("Failed to establish database connection: %w", err)
	}

	log.Println("[DEBUG] [User-Service] Successful connect to Postgre-Client")
	return dbObject, nil
}

type DBObject struct {
	connect *sql.DB
	mapstmt map[string]*sql.Stmt
}

func (db *DBObject) Open(driverName, connectionString string) error {
	var err error
	db.connect, err = sql.Open(driverName, connectionString)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Postgre-Client-Open error: %v", err)
		return err
	}
	db.mapstmt = make(map[string]*sql.Stmt)
	queries := map[string]string{
		insertUserQuery:         "Prepare insert user",
		selectUserGetQuery:      "Prepare selectA user",
		selectUserPasswordQuery: "Prepare selectD user",
		deleteUserQuery:         "Prepare delete user",
		selectUserGetProfile:    "Prepare selectP user",
		updateUserName:          "Prepare updateN user",
		updateUserEmail:         "Prepare updateE user",
		updateUserPassword:      "Prepare updateP user",
		selectEmailCount:        "Prepare selectE user",
	}
	for query, errv := range queries {
		stmt, err := db.connect.Prepare(query)
		if err != nil {
			return fmt.Errorf("%s: %w", errv, err)
		}
		db.mapstmt[query] = stmt
	}
	return nil
}

func (db *DBObject) Close() {
	for _, smtp := range db.mapstmt {
		smtp.Close()
	}
	db.connect.Close()
	log.Println("[DEBUG] [User-Service] Successful close Postgre-Client")
}

func (db *DBObject) Ping() error {
	err := db.connect.Ping()
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Postgre-Client-Ping error: %v", err)
		return err
	}
	return nil
}
func buildConnectionString(cfg configs.DatabaseConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
}
