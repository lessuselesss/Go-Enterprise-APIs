package circular_enterprise_apis

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSetNetwork(t *testing.T) {
	testCases := []struct {
		name           string
		network        string
		mockResponse   string
		mockStatusCode int
		expectedNAGURL string
		expectError    bool
	}{
		{
			name:           "testnet",
			network:        "testnet",
			mockResponse:   "{\"status\":\"success\", \"url\":\"https://nag.circularlabs.io/NAG.php?cep=\", \"message\":\"OK\"}",
			mockStatusCode: http.StatusOK,
			expectedNAGURL: "https://nag.circularlabs.io/NAG.php?cep=",
			expectError:    false,
		},
		{
            name:           "devnet",
            network:        "devnet",
            mockResponse:   "{\"status\":\"success\", \"url\":\"https://nag.circularlabs.io/NAG_DevNet.php?cep=\", \"message\":\"OK\"}",
            mockStatusCode: http.StatusOK,
            expectedNAGURL: "https://nag.circularlabs.io/NAG_DevNet.php?cep=",
            expectError:    false,
        },
		{
            name:           "mainnet",
            network:        "mainnet",
            mockResponse:   "{\"status\":\"success\", \"url\":\"https://nag.circularlabs.io/NAG.php?cep=\", \"message\":\"OK\"}",
            mockStatusCode: http.StatusOK,
            expectedNAGURL: "https://nag.circularlabs.io/NAG_Mainnet.php?cep=",
            expectError:    false,
        },
		{
			name:           "unsupported",
			network:        "unsupported",
			mockResponse:   "Unsupported Network",
			mockStatusCode: http.StatusNotFound,
			expectedNAGURL: "",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.mockStatusCode)
				fmt.Fprint(w, tc.mockResponse)
			}))
			defer server.Close()

			// Point the account to the mock server
			acc := NewCEPAccount()
			acc.NetworkURL = server.URL

			// Call the function
			returnedURL := acc.SetNetwork(tc.network)

			if tc.expectError {
				if returnedURL != "" {
					t.Errorf("Expected empty URL on error, but got %s", returnedURL)
				}
				if acc.GetLastError() == "" {
					t.Error("Expected an error, but LastError was not set")
				}
			} else {
				if acc.GetLastError() != "" {
					t.Errorf("Expected no error, but got: %s", acc.GetLastError())
				}
				if returnedURL != tc.expectedNAGURL {
					t.Errorf("Expected returned URL to be %s, but got %s", tc.expectedNAGURL, returnedURL)
				}
				if acc.NAGURL != tc.expectedNAGURL {
					t.Errorf("Expected acc.NAGURL to be set to %s, but got %s", tc.expectedNAGURL, acc.NAGURL)
				}
				if acc.NetworkNode != tc.network {
					t.Errorf("Expected acc.NetworkNode to be %s, but got %s", tc.network, acc.NetworkNode)
				}
			}
		})
	}
}