package main

import (
	"fmt"
	"log"
	"os"

	circular_enterprise_apis "circular_enterprise_apis/pkg"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve environment variables
	privateKey := os.Getenv("CIRCULAR_PRIVATE_KEY")
	address := os.Getenv("CIRCULAR_ADDRESS")

	if privateKey == "" || address == "" {
		log.Fatal("Please set CIRCULAR_PRIVATE_KEY and CIRCULAR_ADDRESS in your .env file")
	}

	// Initialize CEPAccount
	account := circular_enterprise_apis.NewCEPAccount()
	if !account.Open(address) {
		log.Fatalf("Failed to open account: %s", account.LastError)
	}

	// Set network (e.g., "testnet")
	nagURL := account.SetNetwork("testnet")
	if nagURL == "" {
		log.Fatalf("Failed to set network: %s", account.LastError)
	}
	fmt.Printf("Connected to NAG: %s\n", nagURL)

	// Update account nonce
	if !account.UpdateAccount() {
		log.Fatalf("Failed to update account: %s", account.LastError)
	}
	fmt.Printf("Account nonce updated. Current Nonce: %d\n", account.Nonce)

	// Create and submit a certificate
	certificateData := "Hello, Circular Protocol!"
	account.SubmitCertificate(certificateData, privateKey)
	if account.LastError != "" {
		log.Fatalf("Failed to submit certificate: %s", account.LastError)
	}
	fmt.Printf("Certificate submitted. Latest Transaction ID: %s\n", account.LatestTxID)

	// Poll for transaction outcome
	fmt.Println("Polling for transaction outcome...")
	outcome := account.GetTransactionOutcome(account.LatestTxID, 60, 5) // 60s timeout, 5s interval
	if outcome == nil {
		log.Fatalf("Failed to get transaction outcome: %s", account.LastError)
	}
	fmt.Printf("Transaction Outcome: %+v\n", outcome)

	// Close the account
	account.Close()
	fmt.Println("Account closed.")
}