package main

import (
	"fmt"
	"log"

	// "os"

	"github.com/gin-gonic/gin"
	// "github.com/joho/godotenv"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/handler"
	"github.com/white67/swift_api/internal/parser"
)

func main() {

	// load env variables
	// if os.Getenv("ENV") != "production" {
	// 	err := godotenv.Load("internal/config/.env")
	// 	if err != nil {
	// 		log.Fatal("Error loading .env file")
	// 	}
	// }

	// connect to database
	db := config.ConnectToDB()
	defer db.Close()

	config.InitSchema(db)

	// check if database is empty
	empty, err := database.IsDatabaseEmpty(db)
	if err != nil {
		log.Fatal("Error when checking if database is empty:", err)
	}

	// parse data from .csv file to database if empty
	if empty {
		fmt.Println("Add new data from .csv file as database is empty")
		banks, err := parser.ParseSwiftCSV("data/2025_SWIFT_CODES.csv")
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
	router.GET("/v1/swift-codes/:swiftCode", handler.GetSwiftCodeDetails)
	router.GET("/v1/swift-codes/country/:countryISO2code", handler.GetCountryDetails)
	router.POST("/v1/swift-codes", handler.AddSwiftCode)
	router.DELETE("/v1/swift-codes/:swiftCode", handler.DeleteSwiftCode)
	router.Run(":8080")
}
