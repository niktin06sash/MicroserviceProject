package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/niktin06sash/MicroserviceProject/Photo_service/internal/configs"
)

const LoadPhoto = "Repository-LoadPhoto"
const DeletePhoto = "Repository-DeletePhoto"
const GetPhotos = "Repository-GetPhotos"
const GetPhoto = "Repository-GetPhoto"
const DeleteUserData = "Repository-DeleteUserData"
const AddUserId = "Repository-AddUserId"

func NewPostgresConnection(cfg configs.DatabaseConfig) (*DBObject, error) {
	dbObject := &DBObject{}
	connectionString := buildConnectionString(cfg)
	err := dbObject.Open(cfg.Driver, connectionString)
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Failed to connect to Postgre-Client: %v", err)
		return nil, err
	}

	err = dbObject.Ping()
	if err != nil {
		dbObject.Close()
		return nil, err
	}

	log.Println("[DEBUG] [Photo-Service] Successful connect to Postgre-Client")
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
		log.Printf("[DEBUG] [Photo-Service] Postgre-Client-Open error: %v", err)
		return err
	}
	db.mapstmt = make(map[string]*sql.Stmt)
	queries := map[string]string{
		insertUserQuery:        "Prepare insert userid",
		deletePhotoQuery:       "Prepare delete photo",
		insertPhotoQuery:       "Prepare insert photo",
		deleteUserQuery:        "Prepare delete userid",
		selectPhotoQuery:       "Prepare select photo",
		selectPhotosQuery:      "Prepare select photos",
		selectContentTypeQuery: "Prepare select content_type",
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
	log.Println("[DEBUG] [Photo-Service] Successful close Postgre-Client")
}

func (db *DBObject) Ping() error {
	err := db.connect.Ping()
	if err != nil {
		log.Printf("[DEBUG] [Photo-Service] Postgre-Client-Ping error: %v", err)
		return err
	}
	return nil
}
func buildConnectionString(cfg configs.DatabaseConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
}
