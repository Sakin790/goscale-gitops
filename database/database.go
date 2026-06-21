package database

import (
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {

		dsn = "postgres://postgres:password@localhost:5432/my_project_db?sslmode=disable"
	}

	var err error
	DB, err = sqlx.Connect("pgx", dsn)
	if err != nil {
		log.Fatalf("[❌ DATABASE ERROR] Connection failed: %v\n", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	fmt.Println("✅ [DATABASE] Connected successfully via sqlx & pooled!")
}
