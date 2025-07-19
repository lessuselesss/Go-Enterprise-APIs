package circular_enterprise_apis

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetNAG(t *testing.T) {
	// Save original values to restore after tests
	originalNetworkURL := NetworkURL
	originalHTTPClient := httpClient
	defer func() {
		NetworkURL = originalNetworkURL
		httpClient = originalHTTPClient
	}()

	tests := []struct {
		name           string
		network        string
		mockStatusCode int
		mockBody       string
		expectedURL    string
		expectedError  string
	}{
		{
			name:           "Success",
			network:        "testnet",
			mockStatusCode: http.StatusOK,
			mockBody:       `{"status":"success", "url":"https://mock.nag.url/NAG.php?cep="}`,
			expectedURL:    "https://mock.nag.url/NAG.php?cep=",
			expectedError:  "",
		},
		{
			name:           "EmptyNetworkIdentifier",
			network:        "",
			mockStatusCode: http.StatusOK, // This won't be hit due to early exit
			mockBody:       "",
			expectedURL:    "",
			expectedError:  "network identifier cannot be empty",
		},
		{
			name:           "NonOKStatus",
			network:        "testnet",
			mockStatusCode: http.StatusInternalServerError,
			mockBody:       `{"status":"error", "message":"Internal server error"}`,
			expectedURL:    "",
			expectedError:  "network discovery failed with status: 500 Internal Server Error",
		},
		{
			name:           "InvalidJSON",
			network:        "testnet",
			mockStatusCode: http.StatusOK,
			mockBody:       `invalid json`,
			expectedURL:    "",
			expectedError:  "failed to unmarshal NAG response: invalid character 'i' looking for beginning of value",
		},
		{
			name:           "ErrorStatusInJSON",
			network:        "testnet",
			mockStatusCode: http.StatusOK,
			mockBody:       `{"status":"error", "message":"Network not found"}`,
			expectedURL:    "",
			expectedError:  "failed to get valid NAG URL from response: Network not found",
		},
		{
			name:           "SuccessStatusEmptyURL",
			network:        "testnet",
			mockStatusCode: http.StatusOK,
			mockBody:       `{"status":"success", "url":""}`,
			expectedURL:    "",
			expectedError:  "failed to get valid NAG URL from response: ", // Message is empty as per mockBody
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server for each test case
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				_, err := w.Write([]byte(tt.mockBody))
				if err != nil {
					t.Fatalf("Failed to write mock response: %v", err)
				}
			}))
			defer server.Close()

			// Set the NetworkURL to the mock server's URL
			NetworkURL = server.URL + "/getNAG?network="
			httpClient = server.Client() // Use the mock server's client

			url, err := GetNAG(tt.network)

			// Check for expected URL
			if url != tt.expectedURL {
				t.Errorf("GetNAG() got URL = %q, want %q", url, tt.expectedURL)
			}

			// Check for expected error
			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("GetNAG() unexpected error: %v", err)
				}
			} else {
				if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("GetNAG() got error = %v, want error containing %q", err, tt.expectedError)
				}
			}
		})
	}
}