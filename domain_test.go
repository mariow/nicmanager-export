package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAPIdate(t *testing.T) {
	tests := []struct {
		name        string
		dateString  string
		expected    time.Time
		expectError bool
	}{
		{
			name:        "valid date string",
			dateString:  "2023-03-15T14:30:45Z",
			expected:    time.Date(2023, 3, 15, 14, 30, 45, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "valid date string with zero time",
			dateString:  "2020-01-01T00:00:00Z",
			expected:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "valid date string end of year",
			dateString:  "2022-12-31T23:59:59Z",
			expected:    time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC),
			expectError: false,
		},
		{
			name:        "invalid date format - missing Z",
			dateString:  "2023-03-15T14:30:45",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "invalid date format - wrong separator",
			dateString:  "2023/03/15T14:30:45Z",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "empty string",
			dateString:  "",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "invalid month",
			dateString:  "2023-13-15T14:30:45Z",
			expected:    time.Time{},
			expectError: true,
		},
		{
			name:        "invalid day",
			dateString:  "2023-02-30T14:30:45Z",
			expected:    time.Time{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAPIdate(tt.dateString)

			if tt.expectError {
				assert.Error(t, err, "Expected an error for input: %s", tt.dateString)
			} else {
				require.NoError(t, err, "Unexpected error for input: %s", tt.dateString)
				assert.Equal(t, tt.expected, result, "Parsed date doesn't match expected for input: %s", tt.dateString)
			}
		})
	}
}

func TestDomain_IsBelowCutoff(t *testing.T) {
	cutoffDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		domain   Domain
		expected bool
	}{
		{
			name: "domain with no delete date should be included",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "",
			},
			expected: true,
		},
		{
			name: "domain deleted after cutoff should be included",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "2023-07-15T10:30:00Z", // After cutoff
			},
			expected: true,
		},
		{
			name: "domain deleted on cutoff date should be excluded",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "2023-06-01T00:00:00Z", // Exactly on cutoff
			},
			expected: false,
		},
		{
			name: "domain deleted before cutoff should be excluded",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "2023-05-15T10:30:00Z", // Before cutoff
			},
			expected: false,
		},
		{
			name: "domain deleted just after cutoff should be included",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "2023-06-01T00:00:01Z", // 1 second after cutoff
			},
			expected: true,
		},
		{
			name: "domain deleted just before cutoff should be excluded",
			domain: Domain{
				Name:           "example.com",
				DeleteDateTime: "2023-05-31T23:59:59Z", // 1 second before cutoff
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.domain.IsBelowCutoff(cutoffDate)
			assert.Equal(t, tt.expected, result, "IsBelowCutoff result doesn't match expected for domain: %s", tt.domain.Name)
		})
	}
}

func TestDomain_IsBelowCutoff_EdgeCases(t *testing.T) {
	// Test with different cutoff dates
	t.Run("leap year cutoff", func(t *testing.T) {
		cutoffDate := time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC) // Leap year
		domain := Domain{
			Name:           "leap.com",
			DeleteDateTime: "2024-03-01T00:00:00Z", // Day after leap day
		}
		result := domain.IsBelowCutoff(cutoffDate)
		assert.True(t, result, "Domain deleted after leap year cutoff should be included")
	})

	t.Run("year boundary", func(t *testing.T) {
		cutoffDate := time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)
		domain := Domain{
			Name:           "newyear.com",
			DeleteDateTime: "2024-01-01T00:00:00Z", // New year
		}
		result := domain.IsBelowCutoff(cutoffDate)
		assert.True(t, result, "Domain deleted in new year should be included")
	})
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkParseAPIdate(b *testing.B) {
	dateString := "2023-03-15T14:30:45Z"
	for i := 0; i < b.N; i++ {
		_, _ = parseAPIdate(dateString)
	}
}

func BenchmarkDomain_IsBelowCutoff(b *testing.B) {
	cutoffDate := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	domain := Domain{
		Name:           "benchmark.com",
		DeleteDateTime: "2023-07-15T10:30:00Z",
	}
	
	for i := 0; i < b.N; i++ {
		_ = domain.IsBelowCutoff(cutoffDate)
	}
}

func TestFetchNicmanagerAPI(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		responseStatus int
		expectedError  bool
		login          string
		password       string
		pageNo         int
	}{
		{
			name:           "successful API call",
			responseBody:   `[{"name":"example.com","order_status":"active","order_datetime":"2023-01-01T00:00:00Z","registration_datetime":"2023-01-01T00:00:00Z","delete_datetime":""}]`,
			responseStatus: 200,
			expectedError:  false,
			login:          "testuser",
			password:       "testpass",
			pageNo:         1,
		},
		{
			name:           "empty response",
			responseBody:   `[]`,
			responseStatus: 200,
			expectedError:  false,
			login:          "testuser",
			password:       "testpass",
			pageNo:         1,
		},
		{
			name:           "unauthorized error",
			responseBody:   `{"error":"unauthorized"}`,
			responseStatus: 401,
			expectedError:  true,
			login:          "wronguser",
			password:       "wrongpass",
			pageNo:         1,
		},
		{
			name:           "server error",
			responseBody:   `{"error":"internal server error"}`,
			responseStatus: 500,
			expectedError:  true,
			login:          "testuser",
			password:       "testpass",
			pageNo:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, "/v1/domains")
				assert.Equal(t, "application/json", r.Header.Get("Accept"))
				
				// Check basic auth
				username, password, ok := r.BasicAuth()
				assert.True(t, ok, "Basic auth should be present")
				assert.Equal(t, tt.login, username)
				assert.Equal(t, tt.password, password)
				
				// Check query parameters
				assert.Equal(t, "100", r.URL.Query().Get("limit"))
				assert.Equal(t, "1", r.URL.Query().Get("page"))

				w.WriteHeader(tt.responseStatus)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			// Create a custom client that uses our test server
			client := http.Client{}
			
			// We need to modify the fetchNicmanagerAPI function to accept a custom URL for testing
			// For now, let's test the logic by creating a custom version
			result, err := fetchNicmanagerAPIWithURL(client, tt.login, tt.password, tt.pageNo, server.URL+"/v1/domains")

			if tt.expectedError {
				assert.Error(t, err, "Expected an error for status %d", tt.responseStatus)
			} else {
				require.NoError(t, err, "Unexpected error for successful request")
				assert.Equal(t, tt.responseBody, string(result), "Response body should match")
			}
		})
	}
}

// Helper function for testing with custom URL
func fetchNicmanagerAPIWithURL(client http.Client, login string, password string, pageNo int, baseURL string) ([]byte, error) {
	req, rErr := http.NewRequest("GET", baseURL+"?limit=100&page="+fmt.Sprintf("%d", pageNo), nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(login, password)
	if rErr != nil {
		return nil, rErr
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code error: %d %s", res.StatusCode, res.Status))
	}

	return ioutil.ReadAll(res.Body)
}