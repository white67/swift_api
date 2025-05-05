package handler

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/model"
)

func GetSwiftCodeDetails(c *gin.Context) {
	swiftCode := c.Param("swiftCode")

	db := config.GetDB()

	bank, err := database.GetBankBySwiftCode(db, swiftCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "SWIFT code not found"})
		return
	}

	if bank.IsHeadquarter {
		branches, err := database.GetBranchesForHeadquarter(db, bank.SwiftCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching branches"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"address":       bank.Address,
			"bankName":      bank.Name,
			"countryISO2":   bank.CountryCode,
			"countryName":   bank.CountryName,
			"isHeadquarter": bank.IsHeadquarter,
			"swiftCode":     bank.SwiftCode,
			"branches":      branches,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"address":       bank.Address,
			"bankName":      bank.Name,
			"countryISO2":   bank.CountryCode,
			"countryName":   bank.CountryName,
			"isHeadquarter": bank.IsHeadquarter,
			"swiftCode":     bank.SwiftCode,
		})
	}
}

func GetCountryDetails(c *gin.Context) {
	countryCode := c.Param("countryISO2code")
	db := config.GetDB()

	rows, err := db.Query(`
		SELECT bank_name, address, country_code, country_name, is_headquarter, swift_code
		FROM banks
		WHERE country_code = $1
	`, countryCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query error"})
		return
	}
	defer rows.Close()

	var banks []model.Bank
	var countryName string

	for rows.Next() {
		var b model.Bank
		err := rows.Scan(&b.Name, &b.Address, &b.CountryCode, &countryName, &b.IsHeadquarter, &b.SwiftCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error scanning row"})
			return
		}
		banks = append(banks, b)
	}

	if len(banks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No banks found for given country code"})
		return
	}

	response := gin.H{
		"countryISO2": countryCode,
		"countryName": countryName,
		"swiftCodes":  banks,
	}

	c.JSON(http.StatusOK, response)
}

func AddSwiftCode(c *gin.Context) {
	var bank model.Bank
	if err := c.ShouldBindJSON(&bank); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid JSON format"})
		return
	}

	bank.CountryCode = strings.ToUpper(bank.CountryCode)
	bank.CountryName = strings.ToUpper(bank.CountryName)

	db := config.GetDB()
	_, err := db.Exec(`
		INSERT INTO banks (address, bank_name, country_code, country_name, is_headquarter, swift_code)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,
		bank.Address,
		bank.Name,
		bank.CountryCode,
		bank.CountryName,
		bank.IsHeadquarter,
		bank.SwiftCode,
	)

	if err != nil {
		log.Printf("DB error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to insert SWIFT code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SWIFT code successfully added"})
}

func DeleteSwiftCode(c *gin.Context) {
	swiftCode := c.Param("swiftCode")

	db := config.GetDB()
	result, err := db.Exec("DELETE FROM banks WHERE swift_code = $1", swiftCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete SWIFT code"})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "SWIFT code not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SWIFT code successfully deleted"})
}
