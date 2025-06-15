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
	connect         *sql.DB
	insertUserStmt  *sql.Stmt
	selectUserAStmt *sql.Stmt
	selectUserDStmt *sql.Stmt
	selectUserPStmt *sql.Stmt
	selectUserEStmt *sql.Stmt
	deleteUserStmt  *sql.Stmt
	updateUserNStmt *sql.Stmt
	updateUserEStmt *sql.Stmt
	updateUserPStmt *sql.Stmt
}

func (db *DBObject) Open(driverName, connectionString string) error {
	var err error
	db.connect, err = sql.Open(driverName, connectionString)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Postgre-Client-Open error: %v", err)
		return err
	}
	insertStmt, err := db.connect.Prepare(insertUserQuery)
	if err != nil {
		return fmt.Errorf("Prepare insert user: %w", err)
	}
	selectAStmt, err := db.connect.Prepare(selectUserAuthenticateQuery)
	if err != nil {
		return fmt.Errorf("Prepare selectA user: %w", err)
	}
	selectDStmt, err := db.connect.Prepare(selectUserPasswordQuery)
	if err != nil {
		return fmt.Errorf("Prepare selectD user: %w", err)
	}
	deleteStmt, err := db.connect.Prepare(deleteUserQuery)
	if err != nil {
		return fmt.Errorf("Prepare delete user: %w", err)
	}
	selectPStmt, err := db.connect.Prepare(selectUserGetProfile)
	if err != nil {
		return fmt.Errorf("Prepare selectP user: %w", err)
	}
	updateNStmt, err := db.connect.Prepare(updateUserName)
	if err != nil {
		return fmt.Errorf("Prepare updateN user: %w", err)
	}
	updateEStmt, err := db.connect.Prepare(updateUserEmail)
	if err != nil {
		return fmt.Errorf("Prepare updateE user: %w", err)
	}
	updatePStmt, err := db.connect.Prepare(updateUserPassword)
	if err != nil {
		return fmt.Errorf("Prepare updateP user: %w", err)
	}
	selectEStmt, err := db.connect.Prepare(selectEmailCount)
	if err != nil {
		return fmt.Errorf("Prepare selectE user: %w", err)
	}
	db.insertUserStmt = insertStmt
	db.selectUserAStmt = selectAStmt
	db.selectUserDStmt = selectDStmt
	db.deleteUserStmt = deleteStmt
	db.selectUserPStmt = selectPStmt
	db.updateUserNStmt = updateNStmt
	db.updateUserEStmt = updateEStmt
	db.selectUserEStmt = selectEStmt
	db.updateUserPStmt = updatePStmt
	return nil
}

func (db *DBObject) Close() {
	db.insertUserStmt.Close()
	db.selectUserAStmt.Close()
	db.selectUserDStmt.Close()
	db.deleteUserStmt.Close()
	db.selectUserPStmt.Close()
	db.updateUserNStmt.Close()
	db.updateUserEStmt.Close()
	db.selectUserEStmt.Close()
	db.updateUserPStmt.Close()
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
