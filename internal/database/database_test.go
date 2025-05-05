package database_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/model"
)

var testDB *sql.DB

func setupTestDB(t *testing.T) {
	// Setup test DB connection
	testDB = config.ConnectToDB()
	config.InitSchema(testDB)

	// Clear any existing data
	_, err := testDB.Exec("DELETE FROM banks")
	assert.NoError(t, err, "Failed to clear test database")
}

func teardownTestDB(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM banks")
	assert.NoError(t, err, "Failed to clear test database")
	testDB.Close()
}

func TestIsDatabaseEmpty(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// test with empty database
	empty, err := database.IsDatabaseEmpty(testDB)
	assert.NoError(t, err, "Should not error when checking empty database")
	assert.True(t, empty, "Database should be empty")

	// insert a test bank
	testBank := model.Bank{
		Address:       "Test Address",
		Name:          "Test Bank",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
		SwiftCode:     "TESTUS1XXXX",
	}
	err = database.InsertBank(testDB, testBank)
	assert.NoError(t, err, "Should not error when inserting bank")

	// Test with non-empty database
	empty, err = database.IsDatabaseEmpty(testDB)
	assert.NoError(t, err, "Should not error when checking non-empty database")
	assert.False(t, empty, "Database should not be empty")
}

func TestInsertBank(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	testBank := model.Bank{
		Address:       "Test Address",
		Name:          "Test Bank",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
		SwiftCode:     "TESTUS1XXXX",
	}

	err := database.InsertBank(testDB, testBank)
	assert.NoError(t, err, "Should not error when inserting bank")

	// Verify the bank was inserted
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM banks WHERE swift_code = $1", testBank.SwiftCode).Scan(&count)
	assert.NoError(t, err, "Should not error when querying bank")
	assert.Equal(t, 1, count, "Should have inserted exactly one bank")
}

func TestInsertAllBanks(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	testBanks := []model.Bank{
		{
			Address:       "Test Address 1",
			Name:          "Test Bank 1",
			CountryCode:   "US",
			CountryName:   "UNITED STATES",
			IsHeadquarter: true,
			SwiftCode:     "TESTUS1XXXX",
		},
		{
			Address:       "Test Address 2",
			Name:          "Test Bank 2",
			CountryCode:   "UK",
			CountryName:   "UNITED KINGDOM",
			IsHeadquarter: false,
			SwiftCode:     "TESTGB2YYYY",
		},
	}

	err := database.InsertAllBanks(testDB, testBanks)
	assert.NoError(t, err, "Should not error when inserting multiple banks")

	// Verify the banks were inserted
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM banks").Scan(&count)
	assert.NoError(t, err, "Should not error when querying banks")
	assert.Equal(t, 2, count, "Should have inserted exactly two banks")
}

func TestGetBankBySwiftCode(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	testBank := model.Bank{
		Address:       "Test Address",
		Name:          "Test Bank",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
		SwiftCode:     "TESTUS1XXXX",
	}

	err := database.InsertBank(testDB, testBank)
	assert.NoError(t, err, "Should not error when inserting bank")

	// Test getting the bank
	bank, err := database.GetBankBySwiftCode(testDB, testBank.SwiftCode)
	assert.NoError(t, err, "Should not error when getting bank by swift code")
	assert.Equal(t, testBank.SwiftCode, bank.SwiftCode, "Should return correct swift code")
	assert.Equal(t, testBank.Name, bank.Name, "Should return correct bank name")
	assert.Equal(t, testBank.CountryCode, bank.CountryCode, "Should return correct country code")
	assert.Equal(t, testBank.IsHeadquarter, bank.IsHeadquarter, "Should return correct headquarter status")

	// Test getting a non-existent bank
	_, err = database.GetBankBySwiftCode(testDB, "NONEXISTENT")
	assert.Error(t, err, "Should error when getting non-existent bank")
}

func TestGetBranchesForHeadquarter(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Insert a headquarter
	hqBank := model.Bank{
		Address:       "HQ Address",
		Name:          "Test Bank",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: true,
		SwiftCode:     "TESTUS1XXXX",
	}
	err := database.InsertBank(testDB, hqBank)
	assert.NoError(t, err, "Should not error when inserting headquarter")

	// Insert branches
	branch1 := model.Bank{
		Address:       "Branch Address 1",
		Name:          "Test Bank Branch 1",
		CountryCode:   "US",
		CountryName:   "UNITED STATES",
		IsHeadquarter: false,
		SwiftCode:     "TESTUS1AAA",
	}
	err = database.InsertBank(testDB, branch1)
	assert.NoError(t, err, "Should not error when inserting branch 1")

	// Different bank (should not be included in results)
	otherBank := model.Bank{
		Address:       "Other Bank Address",
		Name:          "Other Bank",
		CountryCode:   "UK",
		CountryName:   "UNITED KINGDOM",
		IsHeadquarter: true,
		SwiftCode:     "OTHERB1XXX",
	}
	err = database.InsertBank(testDB, otherBank)
	assert.NoError(t, err, "Should not error when inserting other bank")

}
