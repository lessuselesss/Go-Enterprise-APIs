package cepaccount

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"circular-api/lib/utils"
)

const (
	testAddress    = "test_address_123"
	testBlockchain = "test_chain"
)

func TestNewCEPAccount(t *testing.T) {
	acc := NewCEPAccount()

	if acc.address != "" {
		t.Errorf("expected address to be empty, but got %s", acc.address)
	}
	if acc.codeVersion != utils.LIB_VERSION {
		t.Errorf("expected codeVersion to be %s, but got %s", utils.LIB_VERSION, acc.codeVersion)
	}
	if acc.nagUrl != utils.DEFAULT_NAG {
		t.Errorf("expected nagUrl to be default, but got %s", acc.nagUrl)
	}
	if acc.blockchain != utils.DEFAULT_CHAIN {
		t.Errorf("expected blockchain to be default, but got %s", acc.blockchain)
	}
	if acc.nonce != 0 {
		t.Errorf("expected nonce to be 0, but got %d", acc.nonce)
	}
	if acc.intervalSec != 2 {
		t.Errorf("expected intervalSec to be 2, but got %d", acc.intervalSec)
	}
}

func TestOpen(t *testing.T) {
	acc := NewCEPAccount()
	opened := acc.Open(testAddress)
	if !opened {
		t.Errorf("expected Open to return true")
	}
	if acc.address != testAddress {
		t.Errorf("expected address to be %s, but got %s", testAddress, acc.address)
	}
}

func TestClose(t *testing.T) {
	acc := NewCEPAccount()
	acc.Open(testAddress)
	acc.nonce = 10
	acc.latestTxID = "some_tx_id"
	acc.Close()

	if acc.address != "" {
		t.Errorf("expected address to be empty after Close, but got %s", acc.address)
	}
	if acc.nonce != 0 {
		t.Errorf("expected nonce to be 0 after Close, but got %d", acc.nonce)
	}
	if acc.latestTxID != "" {
		t.Errorf("expected latestTxID to be empty after Close, but got %s", acc.latestTxID)
	}
}

func TestSetBlockchain(t *testing.T) {
	acc := NewCEPAccount()
	acc.SetBlockchain(testBlockchain)
	if acc.blockchain != testBlockchain {
		t.Errorf("expected blockchain to be %s, but got %s", testBlockchain, acc.blockchain)
	}
}

func TestSetNetwork_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "success", "url": "http://new.nag.url/"}`)
	}))
	defer server.Close()

	// Temporarily override the network URL to point to our mock server
	originalNetworkURL := utils.NETWORK_URL
	utils.NETWORK_URL = server.URL + "/"
	defer func() { utils.NETWORK_URL = originalNetworkURL }()

	acc := NewCEPAccount()
	success := acc.SetNetwork("testnet")

	if !success {
		t.Errorf("expected SetNetwork to return true, got false. Error: %s", acc.lastError)
	}
	if acc.nagUrl != "http://new.nag.url/" {
		t.Errorf("expected nagUrl to be updated to 'http://new.nag.url/', but got %s", acc.nagUrl)
	}
}

func TestUpdateAccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"Result": 200, "Response": {"Nonce": 99}}`)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(testAddress)
	acc.nagUrl = server.URL + "/"

	success := acc.UpdateAccount()

	if !success {
		t.Errorf("expected UpdateAccount to return true, got false. Error: %s", acc.lastError)
	}
	if acc.nonce != 100 {
		t.Errorf("expected nonce to be 100, but got %d", acc.nonce)
	}
}

func TestGettersAndSetters(t *testing.T) {
	acc := NewCEPAccount()
	acc.Open(testAddress)

	// Test LastError
	errMsg := "an error occurred"
	acc.lastError = errMsg
	if acc.GetLastError() != errMsg {
		t.Errorf("expected GetLastError to return '%s', but got '%s'", errMsg, acc.GetLastError())
	}

	// Test Data
	acc.SetData("key1", "value1")
	val, exists := acc.GetData("key1")
	if !exists || val != "value1" {
		t.Errorf("expected GetData('key1') to return 'value1', but it did not")
	}

	// Test GetAddress
	if acc.GetAddress() != testAddress {
		t.Errorf("expected GetAddress to return '%s', but got '%s'", testAddress, acc.GetAddress())
	}

	// Test GetLatestTxID
	txId := "tx_id_555"
	acc.latestTxID = txId
	if acc.GetLatestTxID() != txId {
		t.Errorf("expected GetLatestTxID to return '%s', but got '%s'", txId, acc.GetLatestTxID())
	}

	// Test GetNonce
	acc.nonce = 50
	if acc.GetNonce() != 50 {
		t.Errorf("expected GetNonce to return 50, but got %d", acc.GetNonce())
	}
}

func TestSubmitCertificate_Success(t *testing.T) {
	// This is a complex test that mocks the server and crypto.
	// In a real scenario, more extensive crypto validation would be needed.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var submission map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Validate that TxPayload is a string (JSON in this case)
		txPayload, ok := submission["TxPayload"].(string)
		if !ok {
			http.Error(w, "TxPayload is not a string", http.StatusBadRequest)
			return
		}

		// Validate the contents of the TxPayload
		var payloadContents map[string]interface{}
		if err := json.Unmarshal([]byte(txPayload), &payloadContents); err != nil {
			http.Error(w, "Cannot unmarshal TxPayload", http.StatusBadRequest)
			return
		}

		if payloadContents["Nonce"].(float64) != 1 {
			http.Error(w, "Incorrect Nonce in payload", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"Result": 200, "Response": {"TxID": "tx_success_123"}}`)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(testAddress)
	acc.nagUrl = server.URL + "/"
	acc.nonce = 1

	// NOTE: This is a dummy private key for testing purposes only.
	dummyPrivateKey := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
	certData := `{"message": "test"}`

	resp, err := acc.SubmitCertificate(certData, dummyPrivateKey)

	if err != nil {
		t.Fatalf("SubmitCertificate returned an error: %v", err)
	}

	if acc.GetLatestTxID() != "tx_success_123" {
		t.Errorf("expected latestTxID to be 'tx_success_123', but got '%s'", acc.GetLatestTxID())
	}

	if resp["Result"].(float64) != 200 {
		t.Errorf("expected result to be 200, but got %v", resp["Result"])
	}
}

func TestGetTransactionOutcome_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Circular_GetTransactionOutcome_" {
			http.NotFound(w, r)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if payload["TxID"] != "tx_123" {
			http.Error(w, "Incorrect TxID", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"Result": 200, "Response": {"Status": "Confirmed"}}`)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL + "/"

	outcome, err := acc.GetTransactionOutcome("tx_123")
	if err != nil {
		t.Fatalf("GetTransactionOutcome returned an error: %v", err)
	}

	if outcome["Result"].(float64) != 200 {
		t.Errorf("expected result to be 200, but got %v", outcome["Result"])
	}

	resp, ok := outcome["Response"].(map[string]interface{})
	if !ok {
		t.Fatal("Response field is not a map")
	}

	if resp["Status"] != "Confirmed" {
		t.Errorf("expected status to be 'Confirmed', but got '%s'", resp["Status"])
	}
}

func TestSetNetwork_ServerDown(t *testing.T) {
	// This test ensures that SetNetwork returns false when the server is not reachable.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // Close the server immediately to simulate a down network

	originalNetworkURL := utils.NETWORK_URL
	utils.NETWORK_URL = server.URL + "/"
	defer func() { utils.NETWORK_URL = originalNetworkURL }()

	acc := NewCEPAccount()
	success := acc.SetNetwork("testnet")

	if success {
		t.Errorf("expected SetNetwork to return false when the server is down, but it returned true")
	}
	if acc.GetLastError() == "" {
		t.Errorf("expected a lastError message, but it was empty")
	}
}

func TestUpdateAccount_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"Result": 200, "Response": {"Nonce": "not-a-number"}}`) // Invalid nonce type
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(testAddress)
	acc.nagUrl = server.URL + "/"

	success := acc.UpdateAccount()

	if success {
		t.Errorf("expected UpdateAccount to return false for invalid JSON, but it returned true")
	}
	if acc.GetLastError() == "" {
		t.Errorf("expected a lastError message for invalid JSON, but it was empty")
	}
}

func TestSubmitCertificate_ApiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"Result": 500, "Response": "Internal Server Error"}`)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(testAddress)
	acc.nagUrl = server.URL + "/"
	acc.nonce = 1

	dummyPrivateKey := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
	certData := `{"message": "test"}`

	_, err := acc.SubmitCertificate(certData, dummyPrivateKey)

	if err == nil {
		t.Errorf("expected SubmitCertificate to return an error for API failure, but it did not")
	}
}

func TestWaitForTransactionOutcome_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"Result": 200, "Response": {"Status": "Failed"}}`)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL + "/"

	outcome, err := acc.WaitForTransactionOutcome("tx_failed_123", 5)
	if err != nil {
		t.Fatalf("WaitForTransactionOutcome returned an unexpected error: %v", err)
	}

	resp, _ := outcome["Response"].(map[string]interface{})
	if resp["Status"] != "Failed" {
		t.Errorf("expected final status to be 'Failed', but got '%s'", resp["Status"])
	}
}

func TestWaitForTransactionOutcome_Success(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount < 3 {
			fmt.Fprintln(w, `{"Result": 200, "Response": {"Status": "Pending"}}`)
		} else {
			fmt.Fprintln(w, `{"Result": 200, "Response": {"Status": "Confirmed"}}`)
		}
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL + "/"
	acc.intervalSec = 1 // Shorten interval for testing

	outcome, err := acc.WaitForTransactionOutcome("tx_wait_123", 5)
	if err != nil {
		t.Fatalf("WaitForTransactionOutcome returned an error: %v", err)
	}

	if requestCount < 3 {
		t.Errorf("expected at least 3 requests, but got %d", requestCount)
	}

	resp, ok := outcome["Response"].(map[string]interface{})
	if !ok {
		t.Fatal("Response field is not a map")
	}

	if resp["Status"] != "Confirmed" {
		t.Errorf("expected final status to be 'Confirmed', but got '%s'", resp["Status"])
	}
}

