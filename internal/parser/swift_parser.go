package parser

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/white67/swift_api/internal/model"
)

func ParseSwiftCSV(path string) ([]model.Bank, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	var result []model.Bank

	// do not include first row
	_, err = reader.Read()
	if err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err != io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		countryCode := record[0]
		swiftCode := record[1]
		bankName := record[3]
		address := record[4]
		countryName := record[6]

		swift := model.Bank{
			Address:		address,
			Name:			bankName,
			CountryCode: 	countryCode,
			CountryName:	countryName,
			SwiftCode:		swiftCode,
			IsHeadquarter: 	model.TypeHeadquarters(swiftCode),
		}

		result = append(result, swift)
	}

	return result, nil
}