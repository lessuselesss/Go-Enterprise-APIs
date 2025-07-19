package circular_enterprise_apis

import (
	"encoding/json"
	"testing"
)

func TestSetData(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("test")
	if cert.Data != "74657374" {
		t.Errorf("Expected data to be '74657374', but got '%s'", cert.Data)
	}
}

func TestGetData(t *testing.T) {
	cert := NewCCertificate()
	cert.Data = "74657374"
	if cert.GetData() != "test" {
		t.Errorf("Expected data to be 'test', but got '%s'", cert.GetData())
	}
}

func TestGetJSONCertificate(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("test")
	cert.SetPreviousTxID("0x123")
	cert.SetPreviousBlock("0x456")

	jsonString := cert.GetJSONCertificate()

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if data["data"] != "74657374" {
		t.Errorf("Expected data to be '74657374', but got '%s'", data["data"])
	}
	if data["previousTxID"] != "0x123" {
		t.Errorf("Expected previousTxID to be '0x123', but got '%s'", data["previousTxID"])
	}
	if data["previousBlock"] != "0x456" {
		t.Errorf("Expected previousBlock to be '0x456', but got '%s'", data["previousBlock"])
	}
	if data["version"] != LibVersion {
		t.Errorf("Expected version to be '%s', but got '%s'", LibVersion, data["version"])
	}
}

func TestGetCertificateSize(t *testing.T) {
	cert := NewCCertificate()
	cert.SetData("test")
	cert.SetPreviousTxID("0x123")
	cert.SetPreviousBlock("0x456")

	jsonString := cert.GetJSONCertificate()
	expectedSize := len(jsonString)
	actualSize := cert.GetCertificateSize()

	if actualSize != expectedSize {
		t.Errorf("Expected size to be %d, but got %d", expectedSize, actualSize)
	}
}