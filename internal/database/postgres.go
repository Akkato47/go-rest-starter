package database

import (
	"context"
	"fmt"
	"go-starter/internal/config"
	"log"

	"github.com/jackc/pgx/v5"
)

func Connect(ctx context.Context, config *config.Config) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, fmt.Sprintf("host=%v port=%v user=%v password=%v dbname=%v sslmode=%v", config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbName, config.DbSslMode))
	if err != nil {
		log.Printf("Unable to create conn: %v", err)
		return nil, err
	}

	err = conn.Ping(ctx)
	if err != nil {
		log.Printf("Unable to ping database: %v", err)
		conn.Close(ctx)
		return nil, err
	}

	log.Println("Succefuly connected to PSQL db")
	return conn, nil
}
