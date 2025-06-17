package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"circular-api/errors" // Import the new errors package
)

// PostJSON sends a POST request with a JSON payload and returns the response body.
func PostJSON(url string, payload interface{}) ([]byte, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, &errors.NetworkError{Err: err, URL: url}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	// Basic check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Attempt to parse a potential API error message from the body
		var apiErr map[string]interface{}
		if unmarshalErr := json.Unmarshal(body, &apiErr); unmarshalErr == nil {
			if msg, ok := apiErr["message"].(string); ok {
				return body, &errors.APIError{StatusCode: resp.StatusCode, Message: msg}
			}
			if msg, ok := apiErr["error"].(string); ok {
				return body, &errors.APIError{StatusCode: resp.StatusCode, Message: msg}
			}
		}
		// If no specific message found, return a generic API error
		return body, &errors.APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("HTTP status code %d", resp.StatusCode)}
	}

	return body, nil
}

// Get sends a GET request and returns the response body.
func Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, &errors.NetworkError{Err: err, URL: url}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	// Basic check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Attempt to parse a potential API error message from the body
		var apiErr map[string]interface{}
		if unmarshalErr := json.Unmarshal(body, &apiErr); unmarshalErr == nil {
			if msg, ok := apiErr["message"].(string); ok {
				return body, &errors.APIError{StatusCode: resp.StatusCode, Message: msg}
			}
			if msg, ok := apiErr["error"].(string); ok {
				return body, &errors.APIError{StatusCode: resp.StatusCode, Message: msg}
			}
		}
		// If no specific message found, return a generic API error
		return body, &errors.APIError{StatusCode: resp.StatusCode, Message: fmt.Sprintf("HTTP status code %d", resp.StatusCode)}
	}

	return body, nil
}
