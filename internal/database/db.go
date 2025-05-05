package database

import (
	"database/sql"
	"log"
	"strings"

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
		strings.ToUpper(b.CountryCode), // instead of b.CountryCode
		strings.ToUpper(b.CountryName), // instead of b.CountryName
		b.IsHeadquarter,
		strings.ToUpper(b.SwiftCode), // instead of b.SwiftCode
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

func GetBankBySwiftCode(db *sql.DB, swiftCode string) (*model.Bank, error) {
	row := db.QueryRow("SELECT bank_name, address, country_code, country_name, swift_code, is_headquarter FROM banks WHERE swift_code = $1", swiftCode)

	var b model.Bank
	err := row.Scan(&b.Name, &b.Address, &b.CountryCode, &b.CountryName, &b.SwiftCode, &b.IsHeadquarter)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func GetBranchesForHeadquarter(db *sql.DB, hqSwift string) ([]model.Bank, error) {
	rows, err := db.Query("SELECT bank_name, address, country_code, swift_code, is_headquarter FROM banks WHERE swift_code LIKE $1 AND swift_code != $2", hqSwift[:8]+"%", hqSwift)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []model.Bank
	for rows.Next() {
		var b model.Bank
		err := rows.Scan(&b.Name, &b.Address, &b.CountryCode, &b.SwiftCode, &b.IsHeadquarter)
		if err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}
