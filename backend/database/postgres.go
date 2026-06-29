package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ileego/go_browser_capture/backend/config"
)

var Pool *pgxpool.Pool

func Connect(cfg *config.DatabaseConfig) error {
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var err error
	Pool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return fmt.Errorf("failed to create pool: %w", err)
	}

	if err := Pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
		log.Println("PostgreSQL connection pool closed")
	}
}
