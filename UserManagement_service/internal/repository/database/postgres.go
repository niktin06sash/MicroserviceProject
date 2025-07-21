package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const CreateUser = "Repository-CreateUser"
const GetUser = "Repository-GetUser"
const DeleteUser = "Repository-DeleteUser"
const UpdateName = "Repository-UpdateName"
const UpdatePassword = "Repository-UpdatePassword"
const UpdateEmail = "Repository-UpdateEmail"
const GetMyProfile = "Repository-GetMyProfile"
const GetProfileById = "Repository-GetProfileById"
const UpdateUserData = "Repository-UpdateUserData"

func NewPostgresConnection(cfg configs.DatabaseConfig) (*DBObject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connectionString := buildConnectionString(cfg)
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to parse Postgres-connection string: %v", err)
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Printf("[DEBUG] [User-Service] Failed to create Postgres-connection pool: %v", err)
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	log.Println("[DEBUG] [User-Service] Successful connect to Postgres-Client")
	return &DBObject{pool: pool}, nil
}

type DBObject struct {
	pool *pgxpool.Pool
}

func (db *DBObject) Close() {
	db.pool.Close()
	log.Println("[DEBUG] [User-Service] Successful close Postgre-Client")
}
func (db *DBObject) Ping(ctx context.Context) error {
	err := db.pool.Ping(ctx)
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
