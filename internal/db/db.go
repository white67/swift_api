package database

import (
	"database/sql"
	"log"

	"github.com/white67/swift_api/internal/model"
)

func InsertBank(db *sql.DB, b model.Bank) error {
	query := `
	INSERT INTO banks (
		address,
		bank_name,
		country_code,
		country_name,
		is_headquarter,
		swift_code
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (swift_code) DO NOTHING;`

	_, err := db.Exec(query,
		b.Address,
		b.Name,
		b.CountryCode,
		b.CountryName,
		b.IsHeadquarter,
		b.SwiftCode,
	)

	if err != nil {
		log.Printf("Error when inserting new data: %v", err)
		return err
	}

	return nil
}

func InsertAllBanks(db *sql.DB, banks []model.Bank) error {
	for _, bank := range banks {
		err := InsertBank(db, bank)
		if err != nil {
			log.Printf("Error when inserting new data%s: %v", bank.SwiftCode, err)
		}
	}
	return nil
}

func IsDatabaseEmpty(db *sql.DB) (bool, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM banks").Scan(&count)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}