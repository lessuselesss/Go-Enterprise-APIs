package main

import (
	"fmt"
	"log"

	// Import your circular package.
	// The import path is a combination of your module path (from go.mod)
	// and the directory name of the package.
	"example.com/my_circular_project/circular"
)

func main() {
	log.Println("--- Testing CCertificate ---")
	cert := circular.NewCCertificate()
	cert.SetData("Hello, Circular World!")

	fmt.Println("Certificate Data (Plain):", cert.GetData())
	fmt.Println("Certificate Data (Hex):", cert.Data) // Accessing the raw hex data

	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		log.Fatalf("Error getting JSON certificate: %v", err)
	}
	fmt.Println("JSON Certificate:", jsonCert)

	size, err := cert.GetCertificateSize()
	if err != nil {
		log.Fatalf("Error getting certificate size: %v", err)
	}
	fmt.Printf("Certificate Size: %d bytes\n", size)

	log.Println("\n--- Testing CEPAccount ---")
	account := circular.NewCEPAccount()

	// Open an account
	// Replace with a valid-looking (though not necessarily real for this test) address
	testAddress := "0x1234567890abcdef1234567890abcdef12345678"
	err = account.Open(testAddress)
	if err != nil {
		log.Fatalf("Error opening account: %v", err)
	}
	fmt.Println("Account Address:", account.Address)
	fmt.Println("Account Code Version:", account.CodeVersion)
	fmt.Println("Default NAG URL:", account.NAGURL)
	fmt.Println("Default Blockchain:", account.Blockchain)

	// Set Network (this makes an HTTP GET request)
	// Ensure you have internet connectivity for this to succeed.
	// It might fail if the network URL is down or the network name is invalid.
	err = account.SetNetwork("testnet") // Valid options: "devnet", "testnet", "mainnet"
	if err != nil {
		// Log as a warning for this example, as other local tests can still run
		log.Printf("Warning: Could not set network to 'testnet': %v (continuing with default NAG URL for other tests)", err)
	} else {
		fmt.Println("NAG URL after SetNetwork('testnet'):", account.NAGURL)
	}

	// Update Account (this makes an HTTP POST request)
	// This will likely fail unless 'testnet' (or your chosen network) is operational
	// and the address has a presence on that network.
	log.Println("Attempting to update account (fetches nonce)...")
	updated, updateErr := account.UpdateAccount()
	if !updated || updateErr != nil {
		log.Printf("Warning: Failed to update account (this is expected if network/address isn't live): %v", updateErr)
	} else {
		fmt.Printf("Account updated successfully. New Nonce: %d\n", account.Nonce)
	}

	// Example: Submit Certificate (requires a valid private key and network setup)
	// THIS IS A PLACEHOLDER AND WILL FAIL WITHOUT A REAL PRIVATE KEY AND CORRECT NONCE
	log.Println("Attempting to submit a dummy certificate (will likely fail without real keys/nonce)...")
	// IMPORTANT: Never commit real private keys to version control.
	// This is a dummy key for demonstration structure ONLY.
	dummyPrivateKeyHex := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff" // DO NOT USE FOR REAL
	dummyPData := `{"message": "Test certificate from Go"}`

	// Ensure Nonce is set, e.g., by a successful UpdateAccount or manual setting if known
	// account.Nonce = 1 // For example, if you know the next nonce

	if account.Nonce == 0 && updateErr != nil { // If UpdateAccount failed, Nonce might be 0
		log.Println("Skipping SubmitCertificate because Nonce is 0 (UpdateAccount likely failed).")
	} else {
		submissionResponse, submitErr := account.SubmitCertificate(dummyPData, dummyPrivateKeyHex)
		if submitErr != nil {
			log.Printf("Error submitting certificate: %v", submitErr)
		} else {
			fmt.Println("Certificate Submission Response:", submissionResponse)
			// If submissionResponse indicates success, you might get a TxID
			// txID, ok := submissionResponse["TxID"].(string)
			// if ok && submissionResponse["Result"] == 200.0 { // JSON numbers can be float64
			// log.Printf("Certificate submitted. TxID: %s. Polling for outcome...", txID)
			// outcome, pollErr := account.GetTransactionOutcome(txID, 60) // 60-second timeout
			// if pollErr != nil {
			// log.Printf("Error polling for transaction outcome: %v", pollErr)
			// } else {
			// fmt.Println("Transaction Outcome:", outcome)
			// }
			// }
		}
	}

	// Close account (resets fields)
	account.Close()
	fmt.Println("\nAccount Address after Close():", account.Address) // Should be empty
	fmt.Println("NAG URL after Close():", account.NAGURL)         // Should be default

	log.Println("\n--- Basic library tests completed ---")
}
