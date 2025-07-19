package circular_enterprise_apis

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"circular_enterprise_apis/pkg/utils"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// CEPAccount represents a client-side interface for interacting with the Circular Enterprise Protocol blockchain.
// It encapsulates all necessary account information and provides methods for managing account state,
// interacting with the Network Access Gateway (NAG), and performing blockchain operations such as
// submitting certificates and querying transaction outcomes.
type CEPAccount struct {
	Address     string      // The blockchain address of the account.
	PublicKey   string      // The public key associated with the account.
	Info        interface{} // General information or metadata about the account.
	CodeVersion string      // The version of the client library being used.
	LastError   string      // Stores the last encountered error message, aligning with Java API behavior.
	NAGURL      string      // The URL of the Network Access Gateway (NAG) for the currently configured network.
	NetworkNode string      // Identifier for the specific network node being used (e.g., "testnet", "mainnet").
	Blockchain  string      // The identifier of the blockchain being interacted with.
	LatestTxID  string      // The ID of the most recently submitted transaction by this account.
	Nonce       int64       // A unique, incrementing number used to prevent transaction replay attacks.
	IntervalSec int         // The polling interval in seconds for transaction outcome checks.
	NetworkURL  string      // The base URL for discovering network access gateways.
}

// NewCEPAccount is a factory function that creates and initializes a new CEPAccount instance.
// It sets up the account with default values for the library version, network URLs,
// blockchain, nonce, and transaction polling interval. This function should be used
// to obtain a properly configured CEPAccount object before performing any operations.
//
// Returns:
//
//	A pointer to a newly initialized CEPAccount struct.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		CodeVersion: LibVersion,
		NetworkURL:  NetworkURL,
		NAGURL:      DefaultNAG,
		Blockchain:  DefaultChain,
		Nonce:       0,
		IntervalSec: 2, // Default polling interval
	}
}

// GetLastError retrieves the last error message that occurred during an operation
// performed by the CEPAccount instance. This method is provided for compatibility
// and consistent error reporting across different API implementations (e.g., Java).
//
// Returns:
//
//	A string containing the last error message. Returns an empty string if no error
//	has occurred since the last operation or since the account was initialized.
func (a *CEPAccount) GetLastError() string {
	return a.LastError
}

// Open initializes the CEPAccount with a specified blockchain address.
// This method is a prerequisite for most other account operations.
//
// Parameters:
//   - address: The blockchain address to associate with this account.
//
// Returns:
//
//	`true` if the address is successfully set, and `false` otherwise.
//	If the address is empty, an error message is stored in `a.LastError`.
func (a *CEPAccount) Open(address string) bool {
	if address == "" {
		a.LastError = "invalid address format"
		return false
	}
	a.Address = address
	return true
}

// Close securely clears all sensitive and operational data from the CEPAccount instance.
// This includes the blockchain address, public key, network configurations,
// and any cached transaction IDs or nonces. After calling Close, the account
// must be re-opened using the Open method before it can be used again for
// blockchain operations. This ensures data privacy and resets the account state.
func (a *CEPAccount) Close() {
	a.Address = ""
	a.PublicKey = ""
	a.Info = nil
	a.NAGURL = ""
	a.NetworkNode = ""
	a.Blockchain = ""
	a.LatestTxID = ""
	a.Nonce = 0
	a.IntervalSec = 0
}

// SetNetwork configures the CEPAccount to operate on a specific blockchain network.
// It achieves this by querying a public endpoint to discover the appropriate
// Network Access Gateway (NAG) URL for the given network identifier (e.g., "testnet", "mainnet").
// The discovered NAG URL is then stored internally for subsequent API calls.
//
// Parameters:
//   - network: A string identifier for the desired network (e.g., "devnet", "testnet", "mainnet").
//
// Returns:
//
//	The resolved NAG URL as a string if successful, or an empty string
//	if there's an error during the network discovery process, with the error
//	details stored in `a.LastError`.
func (a *CEPAccount) SetNetwork(network string) string {
	url, err := GetNAG(network)
	if err != nil {
		a.LastError = fmt.Sprintf("network discovery failed: %v", err)
		return ""
	}

	a.NAGURL = url
	a.NetworkNode = network
	return url
}

// SetBlockchain explicitly sets the blockchain identifier for the CEPAccount.
// This function allows overriding the default blockchain configured during initialization.
//
// Parameters:
//   - chain: A valid blockchain address or identifier (e.g., a hexadecimal string)
//     that the account will interact with for all subsequent operations.
func (a *CEPAccount) SetBlockchain(chain string) {
	a.Blockchain = chain
}

// UpdateAccount fetches the latest nonce for the account from the configured Network Access Gateway (NAG).
// The nonce is a crucial component for preventing transaction replay attacks and ensuring
// the sequential ordering of transactions from a given account. This method increments
// the internal nonce value by one after a successful fetch, preparing it for the next transaction.
//
// Returns:
//
//	`true` if the nonce is successfully updated, and `false` otherwise.
//	Any errors encountered during the network request or response parsing are stored in `a.LastError`.
func (a *CEPAccount) UpdateAccount() bool {
	if a.Address == "" {
		a.LastError = "Account not open"
		return false
	}

	requestData := map[string]string{
		"Address":    utils.HexFix(a.Address),
		"Version":    a.CodeVersion,
		"Blockchain": utils.HexFix(a.Blockchain),
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to marshal request data: %v", err)
		return false
	}

	url := a.NAGURL + "Circular_GetWalletNonce_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		a.LastError = fmt.Sprintf("failed to create request: %v", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	fmt.Printf("UpdateAccount: Request URL: %s\n", url)
	fmt.Printf("UpdateAccount: Request Headers: %v\n", req.Header)
	fmt.Printf("UpdateAccount: Request Body: %s\n", string(jsonData))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		a.LastError = fmt.Sprintf("http request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to read response body: %v", err)
		return false
	}

	fmt.Printf("UpdateAccount: Response Status: %s\n", resp.Status)
	fmt.Printf("UpdateAccount: Response Headers: %v\n", resp.Header)
	fmt.Printf("UpdateAccount: Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		a.LastError = fmt.Sprintf("network request failed with status: %s, body: %s", resp.Status, string(body))
		return false
	}

	var responseData struct {
		Result   int         `json:"Result"`
		Response interface{} `json:"Response"`
	}
	if err := json.Unmarshal(body, &responseData); err != nil {
		a.LastError = fmt.Sprintf("failed to decode response body: %v, body: %s", err, string(body))
		fmt.Printf("UpdateAccount: Failed to decode response. Error: %v, Body: %s\n", err, string(body))
		return false
	}

	fmt.Printf("UpdateAccount: Parsed Response - Result: %d, Response: %v\n", responseData.Result, responseData.Response)

	switch responseData.Result {
	case 200:
		// If Result is 200, Response should be a struct with Nonce
		var nonceResponse struct {
			Nonce int `json:"Nonce"`
		}
		responseBytes, err := json.Marshal(responseData.Response)
		if err != nil {
			a.LastError = fmt.Sprintf("failed to marshal response data: %v", err)
			return false
		}
		if err := json.Unmarshal(responseBytes, &nonceResponse); err != nil {
			a.LastError = fmt.Sprintf("failed to decode nonce response: %v, body: %s", err, string(responseBytes))
			return false
		}
		a.Nonce = int64(nonceResponse.Nonce) + 1
		return true
	case 114:
		a.LastError = "Rejected: Invalid Blockchain"
		return false
	case 115:
		a.LastError = "Rejected: Insufficient balance"
		return false
	default:
		// If Result is not 200, Response should be a string error message
		if errMsg, ok := responseData.Response.(string); ok {
			a.LastError = fmt.Sprintf("failed to update account: %s", errMsg)
		} else {
			a.LastError = "failed to update account: unknown error response"
		}
		return false
	}
}

// signData generates a cryptographic signature for a given message using the provided private key.
// This function is an internal helper used by other methods (e.g., SubmitCertificate)
// to ensure the authenticity and integrity of data submitted to the blockchain.
// The message is first hashed using SHA-256, and then signed using the secp256k1 elliptic curve
// digital signature algorithm.
//
// Parameters:
//   - message: The data (typically a hash or transaction ID) to be signed.
//   - privateKeyHex: The private key of the account, in hexadecimal format, used for signing.
//
// Returns:
//
//	The hexadecimal representation of the signature.
//	An error if the private key is invalid or the account is not open.
func (a *CEPAccount) signData(message string, privateKeyHex string) (string, error) {
	if a.Address == "" {
		return "", fmt.Errorf("account is not open")
	}

	privateKeyBytes, err := hex.DecodeString(utils.HexFix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid private key hex string: %w", err)
	}

	privateKey := secp256k1.PrivKeyFromBytes(privateKeyBytes)
	hash := sha256.Sum256([]byte(message))
	signature := ecdsa.Sign(privateKey, hash[:])

	return hex.EncodeToString(signature.Serialize()), nil
}

// SubmitCertificate creates a data certificate, signs it with the provided private key,
// and then submits it to the blockchain via the configured Network Access Gateway (NAG).
// This function encapsulates the entire process of preparing the certificate payload,
// generating a unique transaction ID, signing the transaction, and sending it to the network.
// It updates the account's `LatestTxID` upon successful submission and increments the nonce.
//
// Parameters:
//   - pdata: The primary data content of the certificate to be submitted.
//   - privateKeyHex: The private key of the account, in hexadecimal format, used for signing the transaction.
//
// Returns:
//
//	This function does not explicitly return a value. Any errors during the process
//	(e.g., account not open, signing failure, network issues, or non-200 response from the server)
//	are captured and stored in `a.LastError`.
func (a *CEPAccount) SubmitCertificate(pdata string, privateKeyHex string) {
	if a.Address == "" {
		a.LastError = "Account is not open"
		return
	}

	payloadObject := map[string]string{
		"Action": "CP_CERTIFICATE",
		"Data":   utils.StringToHex(pdata),
	}
	jsonStr, _ := json.Marshal(payloadObject)
	payload := utils.StringToHex(string(jsonStr))
	timestamp := utils.GetFormattedTimestamp()

	strToHash := utils.HexFix(a.Blockchain) + utils.HexFix(a.Address) + utils.HexFix(a.Address) + payload + fmt.Sprintf("%d", a.Nonce) + timestamp
	hash := sha256.Sum256([]byte(strToHash))
	id := hex.EncodeToString(hash[:])

	signature, err := a.signData(id, privateKeyHex)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to sign data: %v", err)
		return
	}

	requestData := map[string]string{
		"ID":         id,
		"From":       utils.HexFix(a.Address),
		"To":         utils.HexFix(a.Address),
		"Timestamp":  timestamp,
		"Payload":    payload,
		"Nonce":      fmt.Sprintf("%d", a.Nonce),
		"Signature":  signature,
		"Blockchain": utils.HexFix(a.Blockchain),
		"Type":       "C_TYPE_CERTIFICATE",
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to marshal request data: %v", err)
		return
	}

	url := a.NAGURL + "Circular_AddTransaction_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		a.LastError = fmt.Sprintf("failed to submit certificate: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to read response body: %v", err)
		return
	}

	fmt.Printf("SubmitCertificate: Response Status: %s\n", resp.Status)
	fmt.Printf("SubmitCertificate: Response Headers: %v\n", resp.Header)
	fmt.Printf("SubmitCertificate: Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		a.LastError = fmt.Sprintf("network returned an error - status: %s, body: %s", resp.Status, string(body))
		return
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(body, &responseMap); err != nil {
		a.LastError = fmt.Sprintf("failed to decode response JSON: %v", err)
		return
	}

	if result, ok := responseMap["Result"].(float64); ok && result == 200 {
		// Save our generated transaction ID
		a.LatestTxID = id
		a.Nonce++ // Increment nonce for the next transaction
	} else {
		// Extract the error message from the response if available
		if errMsg, ok := responseMap["Response"].(string); ok {
			a.LastError = fmt.Sprintf("certificate submission failed: %s", errMsg)
		} else {
			a.LastError = "certificate submission failed with non-200 result code"
		}
	}

}

// GetTransaction retrieves the details of a specific transaction using its block ID and transaction ID.
// This function acts as a convenience wrapper around `getTransactionByID`, specifically
// searching for the transaction within the confines of a single, designated block.
// It is useful when the exact block where a transaction was recorded is known.
//
// Parameters:
//   - blockID: The identifier of the block where the transaction is expected to be found.
//   - transactionID: The unique identifier of the transaction.
//
// Returns:
//
//	A map[string]interface{} containing the transaction details if found.
//	Returns `nil` if the `blockID` is empty or invalid, or if the transaction cannot be retrieved.
//	An error message is stored in `a.LastError` in case of failure.
func (a *CEPAccount) GetTransaction(blockID string, transactionID string) map[string]interface{} {
	if blockID == "" {
		a.LastError = "blockID cannot be empty"
		return nil
	}
	// This function is a convenience wrapper around getTransactionByID,
	// searching within a single, specific block.
	startBlock, err := strconv.ParseInt(blockID, 10, 64)
	if err != nil {
		a.LastError = fmt.Sprintf("invalid blockID: %v", err)
		return nil
	}
	result, err := a.getTransactionByID(transactionID, startBlock, startBlock)
	if err != nil {
		a.LastError = fmt.Sprintf("failed to get transaction by ID: %v", err)
		return nil
	}
	return result
}

// getTransactionByID retrieves the detailed information for a specific transaction by its ID.
// It allows for searching within a specified range of blocks (`startBlock` to `endBlock`).
// This function communicates with the configured Network Access Gateway (NAG) to fetch
// transaction data.
//
// Parameters:
//   - transactionID: The unique identifier of the transaction to retrieve.
//   - startBlock: The starting block number for the search range.
//   - endBlock: The ending block number for the search range.
//
// Returns:
//
//	A map containing the transaction details if successful.
//	An error if the network is not set, the request data cannot be marshaled,
//	the HTTP request fails, the network returns a non-OK status, or the response
//	JSON cannot be decoded.
func (a *CEPAccount) getTransactionByID(transactionID string, startBlock, endBlock int64) (map[string]interface{}, error) {
	if a.NAGURL == "" {
		return nil, fmt.Errorf("network is not set")
	}

	requestData := map[string]string{
		"Blockchain": utils.HexFix(a.Blockchain),
		"ID":         utils.HexFix(transactionID),
		"Start":      fmt.Sprintf("%d", startBlock),
		"End":        fmt.Sprintf("%d", endBlock),
		"Version":    a.CodeVersion,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	url := a.NAGURL + "Circular_GetTransactionbyID_"
	if a.NetworkNode != "" {
		url += a.NetworkNode
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("http post request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("getTransactionByID: Response Status: %s\n", resp.Status)
	fmt.Printf("getTransactionByID: Response Headers: %v\n", resp.Header)
	fmt.Printf("getTransactionByID: Response Body: %s\n", string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("network request failed with status: %s, body: %s", resp.Status, string(body))
	}

	var transactionDetails map[string]interface{}
	if err := json.Unmarshal(body, &transactionDetails); err != nil {
		return nil, fmt.Errorf("failed to decode transaction JSON: %w, body: %s", err, string(body))
	}

	fmt.Printf("getTransactionByID: Parsed Response: %v\n", transactionDetails)

	return transactionDetails, nil
}

// GetTransactionOutcome polls the blockchain for the final status of a transaction
// identified by `txID`. It repeatedly queries the Network Access Gateway (NAG)
// until the transaction's status is no longer "Pending" or a specified timeout is reached.
// The polling interval is determined by `intervalSec`.
//
// Parameters:
//   - txID: The unique identifier of the transaction to monitor.
//   - timeoutSec: The maximum time (in seconds) to wait for the transaction to finalize.
//   - intervalSec: The delay (in seconds) between consecutive polling attempts.
//
// Returns:
//
//	A map[string]interface{} containing the finalized transaction details if successful.
//	Returns `nil` if the timeout is exceeded or if any error occurs during polling,
//	with the specific error message stored in `a.LastError`.
func (a *CEPAccount) GetTransactionOutcome(txID string, timeoutSec int, intervalSec int) map[string]interface{} {
	if a.NAGURL == "" {
		a.LastError = "network is not set"
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSec)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.LastError = "timeout exceeded while waiting for transaction outcome"
			return nil
		case <-ticker.C:
			data, err := a.getTransactionByID(txID, 0, 10) // Search recent blocks
			if err != nil {
				// Log non-critical errors and continue polling
				
				continue
			}

			if result, ok := data["Result"].(float64); ok && result == 200 {
				if response, ok := data["Response"].(map[string]interface{}); ok {
					if status, ok := response["Status"].(string); ok && status != "Pending" {
						return response // Transaction finalized
					}
				}
			}
		}
	}
}
