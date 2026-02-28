package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func connecDatabase() (*pgx.Conn, error) {
	connectionString := os.Getenv("DATABASE")

	fmt.Println(connectionString)

	conn, err := pgx.Connect(context.Background(), connectionString)
	if err != nil {
		_, err2 := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err2 != nil {
			return nil, err2
		}
		os.Exit(1)
	}

	// ping
	err = conn.Ping(context.Background())
	if err != nil {
		_, err3 := fmt.Fprintf(os.Stderr, "Unable to ping database: %v\n", err)
		if err3 != nil {
			return nil, err3
		}
		os.Exit(1)
	}

	fmt.Println("Successfully connected to database!")

	return conn, nil
}
