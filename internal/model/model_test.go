package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/white67/swift_api/internal/model"
)

func TestTypeHeadquarters(t *testing.T) {
	testCases := []struct {
		description string
		swiftCode   string
		expected    bool
	}{
		{
			description: "Headquarters SWIFT code (XXX ending)",
			swiftCode:   "TESTPLPAXXX",
			expected:    true,
		},
		{
			description: "Branch SWIFT code (not XXX ending)",
			swiftCode:   "TESTPLPA123",
			expected:    false,
		},
		{
			description: "Another branch SWIFT code ex",
			swiftCode:   "TESTPLPAABC",
			expected:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := model.TypeHeadquarters(tc.swiftCode)
			assert.Equal(t, tc.expected, result, "Should correctly identify if SWIFT code represents headquarters or not")
		})
	}
}