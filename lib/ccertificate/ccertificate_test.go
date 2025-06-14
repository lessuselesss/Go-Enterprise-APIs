package ccertificate

import (
	"encoding/json"
	"circular-api/lib/utils"
	"testing"
)

func TestNewCCertificate(t *testing.T) {
	cert := NewCCertificate()

	if cert.data != "" {
		t.Errorf("expected data to be empty, but got %s", cert.data)
	}
	if cert.PreviousTxID != "" {
		t.Errorf("expected previousTxID to be empty, but got %s", cert.PreviousTxID)
	}
	if cert.PreviousBlock != "" {
		t.Errorf("expected previousBlock to be empty, but got %s", cert.PreviousBlock)
	}
	if cert.codeVersion != utils.LIB_VERSION {
		t.Errorf("expected codeVersion to be %s, but got %s", utils.LIB_VERSION, cert.codeVersion)
	}
}

func TestSetAndGetData(t *testing.T) {
	cert := NewCCertificate()
	input := "test data"
	expectedHex := "746573742064617461"

	cert.SetData(input)

	if cert.data != expectedHex {
		t.Errorf("expected hex data to be %s, but got %s", expectedHex, cert.data)
	}

	retrievedData := cert.GetData()
	if retrievedData != input {
		t.Errorf("expected retrieved data to be '%s', but got '%s'", input, retrievedData)
	}
}

func TestGetJSONCertificate(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("test data")
	cert.PreviousTxID = "prevTx123"
	cert.PreviousBlock = "prevBlock456"

	jsonString, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate returned an error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["data"] != "746573742064617461" {
		t.Errorf("unexpected data in JSON: got %v", result["data"])
	}
	if result["previousTxID"] != "prevTx123" {
		t.Errorf("unexpected previousTxID in JSON: got %v", result["previousTxID"])
	}
	if result["previousBlock"] != "prevBlock456" {
		t.Errorf("unexpected previousBlock in JSON: got %v", result["previousBlock"])
	}
	if result["version"] != utils.LIB_VERSION {
		t.Errorf("unexpected version in JSON: got %v", result["version"])
	}
}

func TestGetCertificateSize(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("test data")

	jsonString, _ := cert.GetJSONCertificate()
	expectedSize := len(jsonString)

	size, err := cert.GetCertificateSize()
	if err != nil {
		t.Fatalf("GetCertificateSize returned an error: %v", err)
	}

	if size != expectedSize {
		t.Errorf("expected size %d, but got %d", expectedSize, size)
	}
}

// TestSetData_EmptyString checks that setting an empty string works correctly.
func TestSetData_EmptyString(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("")

	if cert.data != "" {
		t.Errorf("expected hex data to be empty for empty string input, but got %s", cert.data)
	}

	retrievedData := cert.GetData()
	if retrievedData != "" {
		t.Errorf("expected retrieved data to be empty, but got '%s'", retrievedData)
	}
}

// TestGetData_InvalidHex checks that GetData returns an empty string if the internal data is not valid hex.
func TestGetData_InvalidHex(t *testing.T) {
	cert := NewCCertificate()
	cert.data = "not-a-hex-string" // Manually set invalid hex data

	retrievedData := cert.GetData()
	if retrievedData != "" {
		t.Errorf("expected GetData to return an empty string for invalid hex, but got '%s'", retrievedData)
	}
}

// TestGetCertificateSize_EmptyCert tests the size of a certificate with no data set.
func TestGetCertificateSize_EmptyCert(t *testing.T) {
	cert := NewCCertificate()
	// No data is set, PreviousTxID and PreviousBlock are empty strings.

	jsonString, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate returned an error: %v", err)
	}
	expectedSize := len(jsonString)

	size, err := cert.GetCertificateSize()
	if err != nil {
		t.Fatalf("GetCertificateSize returned an error: %v", err)
	}

	if size != expectedSize {
		t.Errorf("expected size for an empty certificate to be %d, but got %d", expectedSize, size)
	}
}
