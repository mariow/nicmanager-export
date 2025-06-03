package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainJSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected []Domain
	}{
		{
			name:     "single domain",
			jsonData: `[{"name":"example.com","order_status":"active","order_datetime":"2023-01-01T00:00:00Z","registration_datetime":"2023-01-01T00:00:00Z","delete_datetime":""}]`,
			expected: []Domain{
				{
					Name:                 "example.com",
					OrderStatus:          "active",
					OrderDateTime:        "2023-01-01T00:00:00Z",
					RegistrationDateTime: "2023-01-01T00:00:00Z",
					DeleteDateTime:       "",
				},
			},
		},
		{
			name:     "multiple domains",
			jsonData: `[{"name":"example.com","order_status":"active","order_datetime":"2023-01-01T00:00:00Z","registration_datetime":"2023-01-01T00:00:00Z","delete_datetime":""},{"name":"test.com","order_status":"deleted","order_datetime":"2022-01-01T00:00:00Z","registration_datetime":"2022-01-01T00:00:00Z","delete_datetime":"2023-06-01T00:00:00Z"}]`,
			expected: []Domain{
				{
					Name:                 "example.com",
					OrderStatus:          "active",
					OrderDateTime:        "2023-01-01T00:00:00Z",
					RegistrationDateTime: "2023-01-01T00:00:00Z",
					DeleteDateTime:       "",
				},
				{
					Name:                 "test.com",
					OrderStatus:          "deleted",
					OrderDateTime:        "2022-01-01T00:00:00Z",
					RegistrationDateTime: "2022-01-01T00:00:00Z",
					DeleteDateTime:       "2023-06-01T00:00:00Z",
				},
			},
		},
		{
			name:     "empty array",
			jsonData: `[]`,
			expected: []Domain{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var domains []Domain
			err := json.Unmarshal([]byte(tt.jsonData), &domains)
			require.NoError(t, err, "Failed to unmarshal JSON")
			assert.Equal(t, tt.expected, domains, "Unmarshaled domains don't match expected")
		})
	}
}

func TestCSVWriting(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_export_*.csv")
	require.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tmpFile.Name()) // Clean up

	// Test data
	domains := []Domain{
		{
			Name:                 "example.com",
			OrderDateTime:        "2023-01-01T00:00:00Z",
			RegistrationDateTime: "2023-01-01T00:00:00Z",
			DeleteDateTime:       "",
		},
		{
			Name:                 "test.com",
			OrderDateTime:        "2022-01-01T00:00:00Z",
			RegistrationDateTime: "2022-01-01T00:00:00Z",
			DeleteDateTime:       "2023-06-01T00:00:00Z",
		},
	}

	cutoffDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)

	// Write CSV data
	csvWriter := csv.NewWriter(tmpFile)
	headerWritten := false

	for _, domain := range domains {
		if !headerWritten {
			csvWriter.Write([]string{
				"Domain",
				"Order Date",
				"Reg Date",
				"Close Date",
			})
			headerWritten = true
		}

		if domain.IsBelowCutoff(cutoffDate) {
			// Parse dates
			dateOrd, _ := parseAPIdate(domain.OrderDateTime)
			dateReg, _ := parseAPIdate(domain.RegistrationDateTime)

			// Format Delete date for output
			dateDelFmt := ""
			if domain.DeleteDateTime != "" {
				parsedDate, _ := parseAPIdate(domain.DeleteDateTime)
				dateDelFmt = parsedDate.Format("2006-01-02")
			}

			csvWriter.Write([]string{
				domain.Name,
				dateOrd.Format("2006-01-02"),
				dateReg.Format("2006-01-02"),
				dateDelFmt,
			})
		}
	}
	csvWriter.Flush()
	tmpFile.Close()

	// Read and verify the CSV content
	content, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err, "Failed to read temp file")

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 2, "Expected header + 1 data row") // Only example.com should be included

	// Verify header
	assert.Equal(t, "Domain,Order Date,Reg Date,Close Date", lines[0])

	// Verify data row (only example.com should be included as test.com is deleted before cutoff)
	assert.Equal(t, "example.com,2023-01-01,2023-01-01,", lines[1])
}

func TestDomainFilteringWithCutoff(t *testing.T) {
	cutoffDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)

	domains := []Domain{
		{
			Name:           "active.com",
			DeleteDateTime: "", // No delete date - should be included
		},
		{
			Name:           "deleted-after.com",
			DeleteDateTime: "2023-07-01T00:00:00Z", // Deleted after cutoff - should be included
		},
		{
			Name:           "deleted-before.com",
			DeleteDateTime: "2023-05-01T00:00:00Z", // Deleted before cutoff - should be excluded
		},
		{
			Name:           "deleted-on-cutoff.com",
			DeleteDateTime: "2023-06-01T00:00:00Z", // Deleted exactly on cutoff - should be excluded
		},
	}

	var includedDomains []string
	for _, domain := range domains {
		if domain.IsBelowCutoff(cutoffDate) {
			includedDomains = append(includedDomains, domain.Name)
		}
	}

	expected := []string{"active.com", "deleted-after.com"}
	assert.Equal(t, expected, includedDomains, "Filtered domains don't match expected")
}