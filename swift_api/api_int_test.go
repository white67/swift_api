package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"database/sql"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/white67/swift_api/internal/config"
	"github.com/white67/swift_api/internal/database"
	"github.com/white67/swift_api/internal/handler"
	"github.com/white67/swift_api/internal/model"
	"github.com/white67/swift_api/internal/parser"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	// Setup for integration tests
	gin.SetMode(gin.TestMode)

	// Connect to test database
	db := config.ConnectToDB()
	config.SetDB(db)

	// Initialize schema
	config.InitSchema(db)

	// Clear and seed database with test data
	setupIntegrationTestData(db)

	// Create and start a test server
	router := setupRouter()
	testServer = httptest.NewServer(router)

	// Run tests
	code := m.Run()

	// Cleanup
	testServer.Close()
	db.Exec("DELETE FROM banks")
	db.Close()

	os.Exit(code)
}

func setupRouter() http.Handler {
	router := gin.Default()
	router.GET("/v1/swift-codes/:swiftCode", handler.GetSwiftCodeDetails)
	router.GET("/v1/swift-codes/country/:countryISO2code", handler.GetCountryDetails)
	router.POST("/v1/swift-codes", handler.AddSwiftCode)
	router.DELETE("/v1/swift-codes/:swift-code", handler.DeleteSwiftCode)
	return router
}
func setupIntegrationTestData(db *sql.DB) {
	// Clear existing data
	db.Exec("DELETE FROM banks")

	// Parse and insert sample test data
	banks, _ := parser.ParseSwiftCSV("../../data/test_swift_codes.csv")
	database.InsertAllBanks(db, banks)

	// Add some additional test banks directly
	testBanks := []model.Bank{
		{
			Address:       "Integration Test HQ",
			Name:          "Integration Bank",
			CountryCode:   "IT",
			CountryName:   "ITALY",
			IsHeadquarter: true,
			SwiftCode:     "INTEITRMXXX",
		},
		{
			Address:       "Integration Test Branch",
			Name:          "Integration Bank Branch",
			CountryCode:   "IT",
			CountryName:   "ITALY",
			IsHeadquarter: false,
			SwiftCode:     "INTEITRM123",
		},
	}

	for _, bank := range testBanks {
		database.InsertBank(db, bank)
	}
}

// Helper function to make API requests
func makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, testServer.URL+url, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	return client.Do(req)
}

func TestIntegration_FullAPIFlow(t *testing.T) {
	// 1. Get a country's banks
	t.Run("Get country details", func(t *testing.T) {
		resp, err := makeRequest("GET", "/v1/swift-codes/country/IT", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, "IT", response["countryISO2"])
		assert.Equal(t, "ITALY", response["countryName"])

		banks := response["swiftCodes"].([]interface{})
		assert.GreaterOrEqual(t, len(banks), 2)
	})

	// 2. Add a new bank
	t.Run("Add new bank", func(t *testing.T) {
		newBank := model.Bank{
			Address:       "New Integration Test Bank",
			Name:          "New Test Bank",
			CountryCode:   "IT",
			CountryName:   "Italy", // should convert to uppercase
			IsHeadquarter: false,
			SwiftCode:     "NEWITTES123",
		}
		jsonValue, _ := json.Marshal(newBank)

		resp, err := makeRequest("POST", "/v1/swift-codes", bytes.NewBuffer(jsonValue))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify we can now get the new bank
		resp, err = makeRequest("GET", "/v1/swift-codes/NEWITTES123", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var bankResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&bankResponse)
		assert.NoError(t, err)

		assert.Equal(t, "New Test Bank", bankResponse["bankName"])
		assert.Equal(t, "IT", bankResponse["countryISO2"])
		assert.Equal(t, "ITALY", bankResponse["countryName"]) // Should be uppercase
	})

	// 3. Get headquarters with branches
	t.Run("Get headquarters with branches", func(t *testing.T) {
		resp, err := makeRequest("GET", "/v1/swift-codes/INTEITRMXXX", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)

		assert.Equal(t, "Integration Bank", response["bankName"])
		assert.Equal(t, true, response["isHeadquarter"])

		branches := response["branches"].([]interface{})
		assert.GreaterOrEqual(t, len(branches), 1)

		// Check that one of the branches is what we expect
		foundExpectedBranch := false
		for _, branchInterface := range branches {
			branch := branchInterface.(map[string]interface{})
			if branch["swiftCode"] == "INTEITRM123" {
				foundExpectedBranch = true
				break
			}
		}
		assert.True(t, foundExpectedBranch, "Should find expected branch")
	})

	// 4. Delete a bank
	t.Run("Delete bank", func(t *testing.T) {
		// First verify bank exists
		resp, err := makeRequest("GET", "/v1/swift-codes/NEWITTES123", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Delete the bank
		resp, err = makeRequest("DELETE", "/v1/swift-codes/NEWITTES123", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify bank no longer exists
		resp, err = makeRequest("GET", "/v1/swift-codes/NEWITTES123", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// 5. Check error cases
	t.Run("Error handling", func(t *testing.T) {
		// Non-existent SWIFT code
		resp, err := makeRequest("GET", "/v1/swift-codes/NONEXISTENT", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Delete non-existent SWIFT code
		resp, err = makeRequest("DELETE", "/v1/swift-codes/NONEXISTENT", nil)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
