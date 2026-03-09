package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func Connect(ctx context.Context, dbUrl string) (*pgx.Conn, error) {

	// config, err := pgx.ParseConfig(dbUrl)
	// if err != nil {
	// 	log.Printf("Unable to parse DATABASE_URL: %v", err)
	// 	return nil, err
	// }

	conn, err := pgx.Connect(ctx, "host=localhost user=postgres password=1234 dbname=testing sslmode=disable")
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
