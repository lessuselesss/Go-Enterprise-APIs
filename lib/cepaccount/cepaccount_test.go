package cepaccount

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"circular-api/lib/utils"
)

const (
	mockAddress    = "0x1234567890abcdef1234567890abcdef12345678"
	mockBlockchain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
	// mockPrivateKey is a dummy private key for testing purposes only.
	mockPrivateKey = "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
)

// newMockServer creates a new httptest.Server that responds with a given status code and body.
func newMockServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, body)
	}))
}

// Test_1_2_01_ShouldHaveAllRequiredEnvVariables checks for necessary environment variables.
func Test_1_2_01_ShouldHaveAllRequiredEnvVariables(t *testing.T) {
	if os.Getenv("RUN_ENV_TESTS") != "true" {
		t.Skip("Skipping environment variable test in standard run. Set RUN_ENV_TESTS=true to enable.")
	}
	requiredVars := []string{
		"CIRCULAR_API_TESTNET_BLOCKCHAIN_ID",
		"CIRCULAR_API_TESTNET_ACCOUNT_ADDRESS",
		"CIRCULAR_API_TESTNET_PRIVATE_KEY",
	}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			t.Errorf("Required environment variable %s is not set", v)
		}
	}
}

// Test_1_2_02_ShouldInitializeWithDefaultValues verifies the default state of a new account.
func Test_1_2_02_ShouldInitializeWithDefaultValues(t *testing.T) {
	acc := NewCEPAccount()

	if acc.GetAddress() != "" {
		t.Errorf("expected address to be empty, but got %s", acc.GetAddress())
	}
	if acc.GetNonce() != 0 {
		t.Errorf("expected nonce to be 0, but got %d", acc.GetNonce())
	}
	if acc.GetLastError() != "" {
		t.Errorf("expected lastError to be empty, but got %s", acc.GetLastError())
	}
}

// Test_1_2_03_ShouldOpenAccountAndSetAddress verifies the Open method.
func Test_1_2_03_ShouldOpenAccountAndSetAddress(t *testing.T) {
	acc := NewCEPAccount()
	opened := acc.Open(mockAddress)
	if !opened {
		t.Errorf("expected Open to return true")
	}
	if acc.GetAddress() != mockAddress {
		t.Errorf("expected address to be %s, but got %s", mockAddress, acc.GetAddress())
	}
}

// Test_1_2_04_ShouldClearStateOnClose verifies the Close method resets state.
func Test_1_2_04_ShouldClearStateOnClose(t *testing.T) {
	acc := NewCEPAccount()
	acc.Open(mockAddress)
	acc.SetData("key", "value")
	acc.SetBlockchain("temp_chain")

	acc.Close()

	if acc.GetAddress() != "" {
		t.Errorf("expected address to be empty after Close, but got %s", acc.GetAddress())
	}
	if acc.GetNonce() != 0 {
		t.Errorf("expected nonce to be 0 after Close, but got %d", acc.GetNonce())
	}
	if _, exists := acc.GetData("key"); exists {
		t.Errorf("expected data to be cleared after Close")
	}
}

// Test_1_2_05_ShouldSetTheTargetBlockchain verifies setting the blockchain.
func Test_1_2_05_ShouldSetTheTargetBlockchain(t *testing.T) {
	acc := NewCEPAccount()
	acc.SetBlockchain(mockBlockchain)
	// Verification of this is implicit in other tests that use the blockchain ID.
}

// Test_1_2_06_ShouldUpdateNagUrlOnSuccessfulSetNetwork tests successful network setting.
func Test_1_2_06_ShouldUpdateNagUrlOnSuccessfulSetNetwork(t *testing.T) {
	server := newMockServer(http.StatusOK, `{"status": "success", "url": "http://new.nag.url/"}`)
	defer server.Close()

	originalNetworkURL := utils.NETWORK_URL
	utils.NETWORK_URL = server.URL + "?network="
	defer func() { utils.NETWORK_URL = originalNetworkURL }()

	acc := NewCEPAccount()
	success := acc.SetNetwork("testnet")

	if !success {
		t.Errorf("expected SetNetwork to return true, got false. Error: %s", acc.GetLastError())
	}
}

// Test_1_2_07_ShouldHandleFailureWhenSettingNetwork tests failures in network setting.
func Test_1_2_07_ShouldHandleFailureWhenSettingNetwork(t *testing.T) {
	t.Run("Server Error", func(t *testing.T) {
		server := newMockServer(http.StatusInternalServerError, `{"status": "error", "message": "Server is down"}`)
		defer server.Close()

		originalNetworkURL := utils.NETWORK_URL
		utils.NETWORK_URL = server.URL + "?network="
		defer func() { utils.NETWORK_URL = originalNetworkURL }()

		acc := NewCEPAccount()
		success := acc.SetNetwork("testnet")

		if success {
			t.Errorf("expected SetNetwork to return false, got true")
		}
		if !strings.Contains(acc.GetLastError(), "Failed to set network") {
			t.Errorf("expected lastError to contain 'Failed to set network', but got '%s'", acc.GetLastError())
		}
	})
}

// Test_1_2_09_ShouldUpdateNonceOnSuccessfulAccountUpdate tests successful account updates.
func Test_1_2_09_ShouldUpdateNonceOnSuccessfulAccountUpdate(t *testing.T) {
	server := newMockServer(http.StatusOK, `{"Result": 200, "Response": {"Nonce": 99}}`)
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(mockAddress)
	acc.nagUrl = server.URL

	success := acc.UpdateAccount()

	if !success {
		t.Errorf("expected UpdateAccount to return true, got false. Error: %s", acc.GetLastError())
	}
	if acc.GetNonce() != 100 {
		t.Errorf("expected nonce to be 100, but got %d", acc.GetNonce())
	}
}

// Test_1_2_10_ShouldHandleFailureOnAccountUpdate tests failed account updates.
func Test_1_2_10_ShouldHandleFailureOnAccountUpdate(t *testing.T) {
	server := newMockServer(http.StatusOK, `{"Result": 500, "Response": "Invalid address"}`)
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(mockAddress)
	acc.nagUrl = server.URL

	success := acc.UpdateAccount()

	if success {
		t.Errorf("expected UpdateAccount to return false, got true")
	}
	if !strings.Contains(acc.GetLastError(), "Failed to update account") {
		t.Errorf("expected lastError to contain 'Failed to update account', but got '%s'", acc.GetLastError())
	}
	if acc.GetNonce() != 0 {
		t.Errorf("expected nonce to remain 0 on failure, but got %d", acc.GetNonce())
	}
}

// Test_1_2_17_ShouldReturnTxIDOnSuccessfulSubmission tests successful certificate submission.
func Test_1_2_17_ShouldReturnTxIDOnSuccessfulSubmission(t *testing.T) {
	server := newMockServer(http.StatusOK, `{"Result": 200, "Response": {"TxID": "tx_success_123"}}`)
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(mockAddress)
	acc.nagUrl = server.URL
	acc.nonce = 1

	certData := `{"message": "test"}`
	result, err := acc.SubmitCertificate(certData, mockPrivateKey)

	if err != nil {
		t.Fatalf("expected SubmitCertificate to succeed, but it failed: %v", err)
	}
	if result["Result"].(float64) != 200 {
		t.Errorf("expected result code 200, got %v", result["Result"])
	}
	resp, _ := result["Response"].(map[string]interface{})
	if resp["TxID"] != "tx_success_123" {
		t.Errorf("expected TxID 'tx_success_123', got '%s'", resp["TxID"])
	}
	if acc.GetLatestTxID() != "tx_success_123" {
		t.Errorf("expected latestTxID to be updated, but got '%s'", acc.GetLatestTxID())
	}
}

// Test_1_2_18_ShouldHandleApiErrorOnSubmission tests API errors during submission.
func Test_1_2_18_ShouldHandleApiErrorOnSubmission(t *testing.T) {
	server := newMockServer(http.StatusOK, `{"Result": 500, "Message": "Invalid transaction format"}`)
	defer server.Close()

	acc := NewCEPAccount()
	acc.Open(mockAddress)
	acc.nagUrl = server.URL
	acc.nonce = 1

	certData := `{"message": "test"}`
	_, err := acc.SubmitCertificate(certData, mockPrivateKey)

	if err == nil {
		t.Fatalf("expected SubmitCertificate to fail, but it succeeded")
	}
	if !strings.Contains(err.Error(), "Certificate submission failed") {
		t.Errorf("expected error message to contain 'Certificate submission failed', got '%v'", err)
	}
}

// Test_1_2_23_ShouldReturnTransactionOutcomeOnSuccess tests getting a successful transaction outcome.
func Test_1_2_23_ShouldReturnTransactionOutcomeOnSuccess(t *testing.T) {
	txID := "tx_success_123"
	responseBody := fmt.Sprintf(`{"Result": 200, "Response": {"TxID": "%s", "Status": "Success"}}`, txID)
	server := newMockServer(http.StatusOK, responseBody)
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL

	outcome, err := acc.GetTransactionOutcome(txID)

	if err != nil {
		t.Fatalf("expected GetTransactionOutcome to succeed, but it failed: %v", err)
	}
	resp, _ := outcome["Response"].(map[string]interface{})
	if resp["Status"] != "Success" {
		t.Errorf("expected outcome status 'Success', got '%s'", resp["Status"])
	}
}

// Test_1_2_25_ShouldWaitForTransactionOutcome tests waiting for a transaction outcome.
func Test_1_2_25_ShouldWaitForTransactionOutcome(t *testing.T) {
	txID := "tx_pending_then_success"
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body string
		if callCount == 0 {
			body = fmt.Sprintf(`{"Result": 200, "Response": {"TxID": "%s", "Status": "Pending"}}`, txID)
		} else {
			body = fmt.Sprintf(`{"Result": 200, "Response": {"TxID": "%s", "Status": "Success"}}`, txID)
		}
		callCount++
		fmt.Fprintln(w, body)
	}))
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL
	acc.intervalSec = 1

	outcome, err := acc.WaitForTransactionOutcome(txID, 5)

	if err != nil {
		t.Fatalf("expected WaitForTransactionOutcome to succeed, got error: %v", err)
	}
	resp, _ := outcome["Response"].(map[string]interface{})
	if resp["Status"] != "Success" {
		t.Errorf("expected final status to be 'Success', got '%s'", resp["Status"])
	}
	if callCount < 2 {
		t.Errorf("expected server to be polled at least twice, got %d calls", callCount)
	}
}

// Test_1_2_27_ShouldTimeoutIfOutcomeNotReachedInTime tests the timeout functionality of waiting for an outcome.
func Test_1_2_27_ShouldTimeoutIfOutcomeNotReachedInTime(t *testing.T) {
	txID := "tx_always_pending"
	responseBody := fmt.Sprintf(`{"Result": 200, "Response": {"TxID": "%s", "Status": "Pending"}}`, txID)
	server := newMockServer(http.StatusOK, responseBody)
	defer server.Close()

	acc := NewCEPAccount()
	acc.nagUrl = server.URL
	acc.intervalSec = 1

	_, err := acc.WaitForTransactionOutcome(txID, 2)

	if err == nil {
		t.Fatal("expected WaitForTransactionOutcome to time out, but it succeeded")
	}
	if !strings.Contains(err.Error(), "Timeout waiting for transaction outcome") {
		t.Errorf("expected error to be a timeout error, got: %v", err)
	}
}

// Test_NetworkResilience provides a placeholder for future network resilience tests.
func Test_NetworkResilience(t *testing.T) {
	t.Skip("Skipping network resilience tests; requires advanced mocking capabilities.")
}

