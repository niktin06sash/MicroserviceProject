package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/niktin06sash/MicroserviceProject/UserManagement_service/internal/configs"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDatabaseConnection(cfg configs.DatabaseConfig) (*DBObject, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	connectionString := buildConnectionString(cfg)
	poolConfig, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	log.Println("[DEBUG] [User-Service] Successful connect to Postgre-Client")
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
	return db.pool.Ping(ctx)
}
func buildConnectionString(cfg configs.DatabaseConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)
}
