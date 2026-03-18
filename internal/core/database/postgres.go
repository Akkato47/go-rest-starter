package database

import (
	"context"
	"fmt"
	"go-starter/internal/core/config"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, config *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v", config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbName, config.DbSslMode))
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to create conn: %v", err))
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to ping database: %v", err))
		pool.Close()
		return nil, err
	}

	slog.Info("Successfully connected to PSQL db")
	return pool, nil
}
