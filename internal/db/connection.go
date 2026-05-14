package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool
var QueriesInstance *Queries

func Connect() {
	dsn := "postgres://postgres:Shinas@localhost:5432/billing_db?sslmode=disable"

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	Pool = pool
	QueriesInstance = New(pool)

	log.Println("Connected to billing_db")
}