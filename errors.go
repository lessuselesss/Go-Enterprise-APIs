package circular

import "fmt"

// NetworkError represents an error during a network request.
type NetworkError struct {
	Err error
	URL string
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error accessing %s: %v", e.URL, e.Err)
}

// APIError represents an error returned by the Circular API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status %d, message: %s", e.StatusCode, e.Message)
}

// InvalidResponseError represents an error when the API response format is unexpected.
type InvalidResponseError struct {
	Message string
}

func (e *InvalidResponseError) Error() string {
	return fmt.Sprintf("invalid API response: %s", e.Message)
}

// Other potential custom errors can be added here as needed.
