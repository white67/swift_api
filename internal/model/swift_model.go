package model

type Bank struct {
	Address       string `json:"address"`
	Name          string `json:"bankName"`
	CountryCode   string `json:"countryISO2"`
	CountryName   string `json:"countryName,omitempty"`
	IsHeadquarter bool   `json:"isHeadquarter"`
	SwiftCode     string `json:"swiftCode"`
}

// last 3 letters in Code = branch code (if not XXX)
func TypeHeadquarters(s string) bool {
	if s[8:] == "XXX" {
		return true
	} else {
		return false
	}
}