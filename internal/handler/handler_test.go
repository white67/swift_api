package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/handler"
	"github.com/white67/swift_api/internal/model"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Use test mode for Gin
	gin.SetMode(gin.TestMode)

	// setup test DB connection
	testDB = config.ConnectToDB() // make sure this connects to test DB
	// config.SetDB(testDB)          // set global DB for handlers

	// initialize table schema
	config.InitSchema(testDB)

	// seed test data
	seedTestDatabase()

	code := m.Run()

	// Cleanup DB
	testDB.Exec("DELETE FROM banks")
	testDB.Close()

	os.Exit(code)
}

func seedTestDatabase() {
	// Clear any existing data
	testDB.Exec("DELETE FROM banks")

	// test headquarter
	database.InsertBank(testDB, model.Bank{
		Address:       "Address Test #1",
		Name:          "Bank Test Name",
		CountryCode:   "PL",
		CountryName:   "Poland",
		IsHeadquarter: true,
		SwiftCode:     "TESTPLPWXXX",
	})

	// test branch
	database.InsertBank(testDB, model.Bank{
		Address:       "Branch Address",
		Name:          "Bank Test Name Branch",
		CountryCode:   "PL",
		CountryName:   "Poland",
		IsHeadquarter: false,
		SwiftCode:     "TESTPLPW123",
	})

	// bank from another country
	database.InsertBank(testDB, model.Bank{
		Address:       "German Address",
		Name:          "German Bank",
		CountryCode:   "DE",
		CountryName:   "Germany",
		IsHeadquarter: true,
		SwiftCode:     "TESTDEPWXXX",
	})
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/v1/swift-codes/:swiftCode", handler.GetSwiftCodeDetails)
	router.GET("/v1/swift-codes/country/:countryISO2code", handler.GetCountryDetails)
	router.POST("/v1/swift-codes", handler.AddSwiftCode)
	router.DELETE("/v1/swift-codes/:swiftcode", handler.DeleteSwiftCode)
	return router
}

func TestGetSwiftCodeDetails_Headquarter(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/swift-codes/TESTPLPWXXX", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bank Test Name", response["bankName"])
	assert.Equal(t, "POLAND", response["countryName"])
	assert.Equal(t, true, response["isHeadquarter"])

	// check if branches are included
	branches, ok := response["branches"].([]interface{})
	assert.True(t, ok, "Response should include branches array")
	assert.Len(t, branches, 1, "Should have 1 branch")

	branch := branches[0].(map[string]interface{})
	assert.Equal(t, "TESTPLPW123", branch["swiftCode"])
}

func TestGetSwiftCodeDetails_Branch(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/swift-codes/TESTPLPW123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Bank Test Name Branch", response["bankName"])
	assert.Equal(t, "POLAND", response["countryName"])
	assert.Equal(t, false, response["isHeadquarter"])

	// branches should not be included
	_, branchesExist := response["branches"]
	assert.False(t, branchesExist, "Response should not include branches for a branch")
}

func TestGetSwiftCodeDetails_NotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/swift-codes/NONEXISTENTCODE", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestGetCountryDetails_Success(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/swift-codes/country/PL", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "PL", response["countryISO2"])
	assert.Equal(t, "POLAND", response["countryName"])

	banks, ok := response["swiftCodes"].([]interface{})
	assert.True(t, ok, "Response should include swiftCodes array")
	assert.Len(t, banks, 2, "Should have 2 banks for Poland")
}

func TestGetCountryDetails_NotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/v1/swift-codes/country/FR", nil) // No French banks
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAddSwiftCode_Success(t *testing.T) {
	router := setupRouter()

	newBank := model.Bank{
		Address:       "New Address #2",
		Name:          "New Bank #2",
		CountryCode:   "US",            // Should be converted to uppercase
		CountryName:   "United States", // Should be converted to uppercase
		IsHeadquarter: false,
		SwiftCode:     "NEWUS999ABC",
	}
	jsonValue, _ := json.Marshal(newBank)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/swift-codes", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the bank was added to the database
	bank, err := database.GetBankBySwiftCode(testDB, "NEWUS999ABC")
	assert.NoError(t, err)
	assert.Equal(t, "UNITED STATES", bank.CountryName) // Should be uppercase
}

func TestAddSwiftCode_InvalidJSON(t *testing.T) {
	router := setupRouter()

	invalidJSON := []byte(`{"address": "Invalid JSON`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/swift-codes", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteSwiftCode_Success(t *testing.T) {
	router := setupRouter()

	// First check the bank exists
	_, err := database.GetBankBySwiftCode(testDB, "TESTPLPW123")
	assert.NoError(t, err, "Bank should exist before deletion")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/swift-codes/TESTPLPW123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the bank was deleted
	_, err = database.GetBankBySwiftCode(testDB, "TESTPLPW123")
	assert.Error(t, err, "Bank should no longer exist after deletion")
}

func TestDeleteSwiftCode_NotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/v1/swift-codes/NONEXISTENTCODE", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
