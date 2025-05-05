package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectToDB() *sql.DB {
	// get credentials from .env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlInfo := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Cannot establish connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	fmt.Println("Connected to database")
	return db
}

func InitSchema(db *sql.DB) {
	query := `
	CREATE TABLE IF NOT EXISTS banks (
		id SERIAL PRIMARY KEY,
		address TEXT,
		bank_name TEXT,
		country_code VARCHAR(2),
		country_name TEXT,
		is_headquarter BOOLEAN,
		swift_code VARCHAR(11) UNIQUE
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("Error creating a table: %v", err)
	}
	fmt.Println("Table has been created")
}