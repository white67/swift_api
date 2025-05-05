package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var dbInstance *sql.DB

func ConnectToDB() *sql.DB {

	// load env variables
	if os.Getenv("ENV") != "production" {
		cwd, _ := os.Getwd() // current working directory
		rootPath := findRootEnvPath(cwd)
		err := godotenv.Load(filepath.Join(rootPath, ".env"))
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

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

	SetDB(db) // set the global database connection
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

func GetDB() *sql.DB {
	return dbInstance
}

// sets the global database connection
func SetDB(database *sql.DB) {
	dbInstance = database
}

func findRootEnvPath(startPath string) string {
	current := startPath
	for {
		if _, err := os.Stat(filepath.Join(current, ".env")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			log.Fatal(".env file not found in any parent directories")
		}
		current = parent
	}
}
