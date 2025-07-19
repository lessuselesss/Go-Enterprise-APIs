//go:build integration_tests

package integration_tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	cep "circular_enterprise_apis/pkg"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// Load .env file from the current directory (tests/integration)
	// This makes the environment variables available to the tests
	err := godotenv.Load("../../.env") // Load .env from project root
	if err != nil {
		fmt.Println("Error loading .env file from project root, tests requiring env vars will be skipped.")
	}
	// Run the tests
	os.Exit(m.Run())
}

func TestCircularOperations(t *testing.T) {
	privateKeyHex := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()
	acc.SetNetwork("testnet")
	acc.SetBlockchain("8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2")

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.LastError)
	}

	t.Logf("NAGURL: %s", acc.NAGURL)
	t.Logf("Blockchain: %s", acc.Blockchain)
	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	acc.SubmitCertificate("test message", privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	txHash := acc.LatestTxID
	if txHash == "" {
		t.Fatal("txHash not found in response")
	}

	outcome := acc.GetTransactionOutcome(txHash, 30, 2) // Increased timeout and interval to match Java
	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, ok := outcome["Status"].(string); !ok || status != "Executed" { // Changed to "Executed" to match Java
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}

	// Query the transaction
	var blockID string
	if resp, ok := outcome["Response"].(map[string]interface{}); ok {
		if bID, ok := resp["BlockID"].(string); ok {
			blockID = bID
		}
	}

	if blockID != "" {
		txData := acc.GetTransaction(blockID, txHash)
		if acc.GetLastError() != "" {
			t.Fatalf("acc.GetTransaction() failed: %v", acc.GetLastError())
		}
		if result, ok := txData["Result"].(float64); !ok || result != 200 {
			t.Errorf("Expected transaction query result to be 200, but got %v", result)
		}
	} else {

	}
}

func TestCertificateOperations(t *testing.T) {
	privateKeyHex := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()
	acc.SetNetwork("testnet")
	acc.SetBlockchain("8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2")

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.LastError)
	}

	t.Logf("NAGURL: %s", acc.NAGURL)
	t.Logf("Blockchain: %s", acc.Blockchain)
	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	certificateData := "test data"
	acc.SubmitCertificate(certificateData, privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	txHash := acc.LatestTxID
	if txHash == "" {
		t.Fatal("txHash not found in response")
	}

	outcome := acc.GetTransactionOutcome(txHash, 30, 2) // Increased timeout and interval to match Java
	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, ok := outcome["Status"].(string); !ok || status != "Executed" { // Changed to "Executed" to match Java
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}

	// If a transaction ID is available, test transaction query
	if txHash != "" {
		var blockID string
		if resp, ok := outcome["Response"].(map[string]interface{}); ok {
			if bID, ok := resp["BlockID"].(string); ok {
				blockID = bID
			}
		}
		if blockID != "" {
			txData := acc.GetTransaction(blockID, txHash) // Changed to GetTransaction
			if acc.GetLastError() != "" {
				t.Fatalf("acc.GetTransaction() failed: %v", acc.GetLastError())
			}
			if result, ok := txData["Result"].(float64); !ok || result != 200 {
				t.Errorf("Expected transaction query result to be 200, but got %v", result)
			}
		} else {

		}
	}
}

func TestHelloWorldCertification(t *testing.T) {
	privateKeyHex := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()
	acc.SetNetwork("testnet")
	acc.SetBlockchain("8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2")

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.LastError)
	}

	t.Logf("NAGURL: %s", acc.NAGURL)
	t.Logf("Blockchain: %s", acc.Blockchain)
	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	// Create and submit the certificate with timestamp
	message := "Hello World"
	certificateData := fmt.Sprintf(
		`{"message":"%s","timestamp":%d}`,
		message,
		time.Now().UnixNano()/int64(time.Millisecond),
	)

	acc.SubmitCertificate(certificateData, privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	// Get and save the transaction ID with retries
	txHash := ""
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		txHash = acc.LatestTxID
		if txHash != "" {
			break
		}
		t.Logf("Transaction ID not available yet, retrying in 2 seconds... (Attempt %d/%d)", i+1, maxRetries)
		time.Sleep(2 * time.Second)
	}

	if txHash == "" {
		t.Fatal("Failed to get transaction ID after retries")
	}

	// Wait for transaction to be processed and get outcome (increased timeout to 120 seconds)
	outcome := acc.GetTransactionOutcome(txHash, 120, 5)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, ok := outcome["Status"].(string); !ok || status != "Executed" {
		t.Errorf("Expected transaction status to be 'Executed', but got '%s'", status)
	}

	// Query the transaction (Java test also queries by blockID and txId)
	var blockID string
	if resp, ok := outcome["Response"].(map[string]interface{}); ok {
		if bID, ok := resp["BlockID"].(string); ok {
			blockID = bID
		}
	}

	if blockID != "" {
		txData := acc.GetTransaction(blockID, txHash)
		if acc.GetLastError() != "" {
			t.Fatalf("acc.GetTransaction() failed: %v", acc.GetLastError())
		}
		if result, ok := txData["Result"].(float64); !ok || result != 200 {
			t.Errorf("Expected transaction query result to be 200, but got %v", result)
		}
	} else {

	}
}
