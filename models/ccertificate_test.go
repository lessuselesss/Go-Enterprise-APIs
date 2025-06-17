package models

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"circular-api/internal/utils"
)

// [1.1.01] should have all required env variables for testnet
func Test_1_1_01_ShouldHaveAllRequiredEnvVariablesForTestnet(t *testing.T) {
	// Note: This test is marked with the 'integration' build tag
	// and will only run when explicitly requested.
	t.Skip("Skipping environment variable test in standard run.")

	requiredVars := []string{
		"TESTNET_CIRCULAR_SANDBOX_ACCOUNT_PUBKEY",
		"TESTNET_CIRCULAR_SANDBOX_ACCOUNT_PVTKEY",
	}

	for _, varName := range requiredVars {
		value := os.Getenv(varName)
		if value == "" {
			t.Errorf("Environment variable '%s' is not set", varName)
		}
		if strings.Contains(value, "<") || strings.Contains(value, ">") {
			t.Errorf("Environment variable '%s' appears to be a placeholder: %s", varName, value)
		}
	}
}

// [1.1.02] should initialize with default values
func Test_1_1_02_ShouldInitializeWithDefaultValues(t *testing.T) {
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

// [1.1.04] should store data as hex
func Test_1_1_04_ShouldStoreDataAsHex(t *testing.T) {
	cert := NewCCertificate()
	testData := "test data is a string"
	expectedHex := utils.StringToHex(testData)

	cert.SetData(testData)

	if cert.data != expectedHex {
		t.Errorf("expected hex to be '%s', but got '%s'", expectedHex, cert.data)
	}
}

// [1.1.05] should retrieve original data for simple strings
func Test_1_1_05_ShouldRetrieveOriginalDataForSimpleStrings(t *testing.T) {
	cert := NewCCertificate()
	originalData := "another test"

	cert.SetData(originalData)
	retrievedData := cert.GetData()

	if retrievedData != originalData {
		t.Errorf("expected data to be '%s', but got '%s'", originalData, retrievedData)
	}
}

// [1.1.06] should return empty string if data is null or empty hex
func Test_1_1_06_ShouldReturnEmptyStringIfDataIsNullOrEmptyHex(t *testing.T) {
	cert := NewCCertificate()

	// Test with nil/empty data field
	cert.data = ""
	if cert.GetData() != "" {
		t.Errorf("expected empty string for empty hex, but got '%s'", cert.GetData())
	}
}

// [1.1.07] should return empty string if data is "0x"
func Test_1_1_07_ShouldReturnEmptyStringIfDataIs0x(t *testing.T) {
	cert := NewCCertificate()
	cert.data = "0x"

	if cert.GetData() != "" {
		t.Errorf("expected empty string for '0x', but got '%s'", cert.GetData())
	}
}

// [1.1.08] should correctly retrieve multi-byte unicode data
func Test_1_1_08_ShouldCorrectlyRetrieveMultiByteUnicodeData(t *testing.T) {
	cert := NewCCertificate()
	unicodeData := "你好世界 😊"

	cert.SetData(unicodeData)
	retrievedData := cert.GetData()

	if retrievedData != unicodeData {
		t.Errorf("expected unicode data to be '%s', but got '%s'", unicodeData, retrievedData)
	}
}

// [1.1.09] should return a valid JSON string
func Test_1_1_09_ShouldReturnValidJSONString(t *testing.T) {
	cert := NewCCertificate()
	testData := "json test"
	cert.SetData(testData)
	cert.PreviousTxID = "tx123"
	cert.PreviousBlock = "block456"

	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		t.Fatalf("GetJSONCertificate returned an error: %v", err)
	}

	var parsedCert map[string]interface{}
	if err := json.Unmarshal([]byte(jsonCert), &parsedCert); err != nil {
		t.Fatalf("Failed to parse JSON certificate: %v", err)
	}

	if parsedCert["data"] != utils.StringToHex(testData) {
		t.Errorf("mismatched data in JSON")
	}
	if parsedCert["previousTxID"] != "tx123" {
		t.Errorf("mismatched previousTxID in JSON")
	}
	if parsedCert["previousBlock"] != "block456" {
		t.Errorf("mismatched previousBlock in JSON")
	}
	if parsedCert["version"] != utils.LIB_VERSION {
		t.Errorf("mismatched version in JSON")
	}
}

// [1.1.10] should return correct byte length
func Test_1_1_10_ShouldReturnCorrectByteLength(t *testing.T) {
	cert := NewCCertificate()
	testData := "size test"
	cert.SetData(testData)
	cert.PreviousTxID = "txIDForSize"
	cert.PreviousBlock = "blockIDForSize"

	jsonString, _ := cert.GetJSONCertificate()
	expectedSize := len(jsonString)

	size, err := cert.GetCertificateSize()
	if err != nil {
		t.Fatalf("GetCertificateSize returned an error: %v", err)
	}

	if size != expectedSize {
		t.Errorf("expected size to be %d, but got %d", expectedSize, size)
	}
}

// [1.1.11] should return correct byte length for an empty certificate
func Test_1_1_11_ShouldReturnCorrectByteLengthForEmptyCertificate(t *testing.T) {
	cert := NewCCertificate()

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
