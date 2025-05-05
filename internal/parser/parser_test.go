package parser_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/white67/swift_api/internal/parser"
)

func TestParseSwiftCSV(t *testing.T) {
	// Create a temporary test CSV file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test_swift_codes.csv")

	// Create test data
	csvContent := `Country,SWIFT Code,Code Type,Bank Name,Address,Town,Country Name
PL,TESTPLPWXXX,BIC11,Test Bank Poland,Test Address 1,Mielno,Poland
US,TESTUSNYABC,BIC11,Test Bank USA,USA Address 1,Dallas,United States
FR,TESTFRPPXXX,BIC11,Test Bank France,France Address,Nice,France
`

	err := os.WriteFile(tempFile, []byte(csvContent), 0644)
	assert.NoError(t, err, "Failed to create test CSV file")

	// Test the parser
	banks, err := parser.ParseSwiftCSV(tempFile)
	assert.NoError(t, err, "Parser should not return an error")
	assert.Len(t, banks, 3, "Parser should return 3 bank entries")

	// Check first bank
	assert.Equal(t, "PL", banks[0].CountryCode)
	assert.Equal(t, "TESTPLPWXXX", banks[0].SwiftCode)
	assert.Equal(t, "Test Bank Poland", banks[0].Name)
	assert.Equal(t, "Test Address 1", banks[0].Address)
	assert.Equal(t, "POLAND", banks[0].CountryName)
	assert.True(t, banks[0].IsHeadquarter)

	// Check second bank
	assert.Equal(t, "US", banks[1].CountryCode)
	assert.Equal(t, "TESTUSNYABC", banks[1].SwiftCode)
	assert.False(t, banks[1].IsHeadquarter)

	// Check third bank
	assert.Equal(t, "FR", banks[2].CountryCode)
	assert.Equal(t, "TESTFRPPXXX", banks[2].SwiftCode)
	assert.True(t, banks[2].IsHeadquarter)
}

func TestParseSwiftCSV_FileNotFound(t *testing.T) {
	_, err := parser.ParseSwiftCSV("non_existent_file.csv")
	assert.Error(t, err, "Parser should return an error for non-existent file")
}
