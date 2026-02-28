package main

import (
	"context"
	"fmt"
	"os"

	// Notice we use pgxpool instead of just pgx
	"github.com/jackc/pgx/v5/pgxpool"
)

// Change return type to *pgxpool.Pool
func connecDatabase() (*pgxpool.Pool, error) {
	connectionString := os.Getenv("DATABASE")
	if connectionString == "" {
		return nil, fmt.Errorf("DATABASE environment variable is not set")
	}

	// Create a connection pool (handles concurrent requests automatically)
	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		// Return the error instead of calling os.Exit()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	// Ping to verify the connection is actually alive
	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database!")
	return pool, nil
}
