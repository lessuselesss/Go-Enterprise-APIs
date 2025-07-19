//go:build e2e

package e2e_tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	cep "circular_enterprise_apis/pkg"

	"github.com/joho/godotenv"
)

var (
	privateKeyHex string
	address       string
)

func TestMain(m *testing.M) {
	// Load .env file from the project root
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Println("Error loading .env file, tests requiring env vars will be skipped.")
	}

	privateKeyHex = os.Getenv("CIRCULAR_PRIVATE_KEY")
	address = os.Getenv("CIRCULAR_ADDRESS")

	// Run the tests
	os.Exit(m.Run())
}

func TestE2ECircularOperations(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.GetLastError())
	}

	if nagURL := acc.SetNetwork("testnet"); nagURL == "" {
		t.Fatalf("acc.SetNetwork() failed: %v", acc.GetLastError())
	}
	acc.SetBlockchain(cep.DefaultChain)

	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	acc.SubmitCertificate("test message from Go E2E test", privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	txHash := acc.LatestTxID
	if txHash == "" {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, _ := outcome["Status"].(string); status != "Executed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}

func TestE2ECertificateOperations(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.GetLastError())
	}

	if nagURL := acc.SetNetwork("testnet"); nagURL == "" {
		t.Fatalf("acc.SetNetwork() failed: %v", acc.GetLastError())
	}
	acc.SetBlockchain(cep.DefaultChain)

	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	acc.SubmitCertificate("{\"test\":\"data\"}", privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	txHash := acc.LatestTxID
	if txHash == "" {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, _ := outcome["Status"].(string); status != "Executed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}

func TestE2EHelloWorldCertification(t *testing.T) {
	if privateKeyHex == "" || address == "" {
		t.Skip("Skipping E2E test: CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS environment variables must be set")
	}

	acc := cep.NewCEPAccount()

	if !acc.Open(address) {
		t.Fatalf("acc.Open() failed: %v", acc.GetLastError())
	}

	if nagURL := acc.SetNetwork("testnet"); nagURL == "" {
		t.Fatalf("acc.SetNetwork() failed: %v", acc.GetLastError())
	}
	acc.SetBlockchain(cep.DefaultChain)

	if !acc.UpdateAccount() {
		t.Fatalf("acc.UpdateAccount() failed: %v", acc.GetLastError())
	}

	acc.SubmitCertificate("Hello World", privateKeyHex)
	if acc.GetLastError() != "" {
		t.Fatalf("acc.SubmitCertificate() failed: %v", acc.GetLastError())
	}

	txHash := acc.LatestTxID
	if txHash == "" {
		t.Fatal("txHash not found in response")
	}

	// Poll for transaction outcome
	var outcome map[string]interface{}
	for i := 0; i < 10; i++ {
		outcome = acc.GetTransactionOutcome(txHash, 10, acc.IntervalSec)
		if outcome != nil {
			if status, ok := outcome["Status"].(string); ok && status == "Confirmed" {
				break
			}
		}
		time.Sleep(2 * time.Second)
	}

	if acc.GetLastError() != "" {
		t.Fatalf("acc.GetTransactionOutcome() failed: %v", acc.GetLastError())
	}

	if status, _ := outcome["Status"].(string); status != "Executed" {
		t.Errorf("Expected transaction status to be 'Confirmed', but got '%s'", status)
	}
}
