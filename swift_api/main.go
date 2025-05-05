package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/parser"
)

func main() {

	db := config.ConnectToDB()
	defer db.Close()

	config.InitSchema(db)

	empty, err := database.IsDatabaseEmpty(db)
	if err != nil {
		log.Fatal("Error when checking if database is empty:", err)
	}

	if empty {
		fmt.Println("Add new data from .csv file as database is empty")
		banks, err := parser.ParseSwiftCSV("data/Interns_2025_SWIFT_CODES - Sheet1.csv")
		if err != nil {
			log.Fatal(err)
		}
		err = database.InsertAllBanks(db, banks)
		if err != nil {
			log.Fatal("Error when inserting new items:", err)
		}
	}

	// create gin router
	router := gin.Default()

	router.GET("/v1/swift-codes/{swift-code}")
}
