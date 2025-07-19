package circular_enterprise_apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// httpClient is the default HTTP client used for making network requests within the Circular Enterprise APIs.
// It is configured with standard settings and is utilized by functions that communicate with external services,
// such as the Network Access Gateway (NAG) for network discovery and transaction submission.
var httpClient *http.Client = http.DefaultClient

// Constants define fundamental parameters and metadata for the Circular Enterprise APIs.
const (
	// LibVersion specifies the current semantic version of the Go client library.
	// This version is included in various API requests to ensure compatibility
	// and for tracking purposes on the Circular Protocol network.
	LibVersion = "1.0.13"

	// DefaultChain represents the blockchain identifier for the default public network.
	// This hexadecimal string uniquely identifies the primary blockchain that the
	// Circular Enterprise APIs will interact with by default, unless explicitly overridden.
	DefaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"

	// DefaultNAG is the base URL for the default public Network Access Gateway (NAG).
	// The NAG serves as the primary entry point for client applications to interact
	// with the Circular Protocol network, facilitating operations like transaction
	// submission and account nonce retrieval.
	DefaultNAG = "https://nag.circularlabs.io/NAG.php?cep="
)

// NetworkURL is the base endpoint used for discovering and resolving the appropriate
// Network Access Gateway (NAG) for a given network. This URL points to a service
// that provides the specific NAG endpoint based on the network identifier provided.
var NetworkURL = "https://circularlabs.io/network/getNAG?network="

// GetNAG is a utility function responsible for discovering the Network Access Gateway (NAG) URL
// for a specified network. It performs an HTTP GET request to the `NetworkURL` endpoint,
// appending the `network` identifier as a query parameter.
//
// Parameters:
//   - network: A string identifier for the desired network (e.g., "testnet", "mainnet").
//
// Returns:
//   - The resolved NAG URL as a string if the request is successful and a valid URL is returned.
//   - An error if the network identifier is empty, the HTTP request fails, the network
//     discovery service returns a non-OK status, or the response cannot be parsed
//     or indicates an error.
func GetNAG(network string) (string, error) {
	if network == "" {
		return "", fmt.Errorf("network identifier cannot be empty")
	}

	resp, err := httpClient.Get(NetworkURL + network)
	if err != nil {
		return "", fmt.Errorf("failed to fetch NAG URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("network discovery failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// The response is expected to be a JSON object like {"status":"success", "url":"..."}
	var nagResponse struct {
		Status  string `json:"status"`
		URL     string `json:"url"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &nagResponse); err != nil {
		return "", fmt.Errorf("failed to unmarshal NAG response: %w", err)
	}

	fmt.Printf("NAG Response Status: %s\n", nagResponse.Status)
	fmt.Printf("NAG Response Message: %s\n", nagResponse.Message)

	if nagResponse.Status == "error" {
		return "", fmt.Errorf("failed to get valid NAG URL from response: %s", nagResponse.Message)
	}

	if nagResponse.Status != "success" || nagResponse.URL == "" {
		return "", fmt.Errorf("failed to get valid NAG URL from response: %s", nagResponse.Message)
	}

	return nagResponse.URL, nil
}
