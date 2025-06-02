package circular

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"context"
)

const (
	libVersion   = "1.0.13"
	networkURL   = "https://circularlabs.io/network/getNAG?network="
	defaultChain = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
	defaultNAG   = "https://nag.circularlabs.io/NAG.php?cep="
)

// httpClient is a shared HTTP client for the package.
// It's good practice to reuse http.Client.
var httpClient = &http.Client{
	Timeout: 30 * time.Second, // Default timeout for HTTP requests
}

/* HELPER FUNCTIONS ***********************************************************/

// padNumber adds a leading zero to numbers less than 10.
func padNumber(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	}
	return strconv.Itoa(num)
}

// getFormattedTimestamp returns the current UTC time in "YYYY:MM:DD-HH:MM:SS" format.
func getFormattedTimestamp() string {
	now := time.Now().UTC()
	return fmt.Sprintf("%d:%s:%s-%s:%s:%s",
		now.Year(),
		padNumber(int(now.Month())),
		padNumber(now.Day()),
		padNumber(now.Hour()),
		padNumber(now.Minute()),
		padNumber(now.Second()),
	)
}

// hexFix removes "0x" prefix from a hexadecimal string if present.
func hexFix(word string) string {
	if strings.HasPrefix(strings.ToLower(word), "0x") {
		return word[2:]
	}
	return word
}

// stringToHex converts a string to its hexadecimal representation (without "0x" prefix).
func stringToHex(s string) string {
	return hex.EncodeToString([]byte(s))
}

// hexToString converts a hexadecimal string (without "0x" prefix) to a regular string.
// It skips null bytes.
func hexToString(h string) string {
	h = hexFix(h)
	decoded, err := hex.DecodeString(h)
	if err != nil {
		// In JS, parseInt would return NaN for invalid hex, and fromCharCode(NaN) is not appended.
		// Here, we return an empty string or handle error as appropriate.
		// For now, log error and return what's parsable or empty.
		log.Printf("hexToString: error decoding hex: %v", err)
		// Attempt to convert valid parts if any, or return empty.
		// The original JS code iterates and converts char by char, effectively skipping invalid ones.
		// This Go version using hex.DecodeString is more all-or-nothing for the whole string.
		// To mimic JS more closely, one would iterate and parse hex.DecodeString on 2-char substrings.
		// However, usually, a valid hex string is expected.
		var sb strings.Builder
		for i := 0; i < len(h); i += 2 {
			if i+2 > len(h) {
				break // Not enough characters for a full byte
			}
			byteStr := h[i : i+2]
			val, err := strconv.ParseUint(byteStr, 16, 8)
			if err == nil && val != 0 { // Skip null bytes as in JS
				sb.WriteByte(byte(val))
			}
		}
		return sb.String()

	}
	// Remove null bytes, as in the original JS logic String.fromCharCode(0) is often skipped.
	var result []byte
	for _, b := range decoded {
		if b != 0 {
			result = append(result, b)
		}
	}
	return string(result)
}

/*******************************************************************************
 * Circular Certificate Class for certificate chaining
 */

// CCertificate represents a Circular Certificate.
type CCertificate struct {
	Data          string `json:"data"`          // Hex encoded data
	PreviousTxID  string `json:"previousTxID"`  // Hex encoded
	PreviousBlock string `json:"previousBlock"` // Hex encoded
	Version       string `json:"version"`
}

// NewCCertificate creates a new CCertificate instance.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		Version: libVersion,
	}
}

// SetData inserts application data into the certificate (stores as hex).
func (c *CCertificate) SetData(data string) {
	c.Data = stringToHex(data)
}

// GetData extracts application data from the certificate (decodes from hex).
func (c *CCertificate) GetData() string {
	return hexToString(c.Data)
}

// GetJSONCertificate returns the certificate in JSON string format.
func (c *CCertificate) GetJSONCertificate() (string, error) {
	jsonData, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetCertificateSize extracts certificate size in bytes (of its JSON representation).
func (c *CCertificate) GetCertificateSize() (int, error) {
	jsonString, err := c.GetJSONCertificate()
	if err != nil {
		return 0, err
	}
	return len([]byte(jsonString)), nil
}

/*******************************************************************************
 * Circular Account Class
 */

// CEPAccount represents a Circular Enterprise Account.
type CEPAccount struct {
	Address      string
	PublicKey    string // Not actively used in the provided JS, but kept for structure
	Info         interface{} // Not actively used
	CodeVersion  string
	LastError    string
	NAGURL       string
	NetworkNode  string
	Blockchain   string
	LatestTxID   string
	Nonce        int64 // JS uses number, can be large
	Data         map[string]interface{}
	IntervalSec  int
	httpClient   *http.Client // Allow custom client for testing or specific needs
}

// NewCEPAccount creates a new CEPAccount instance.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		CodeVersion: libVersion,
		NAGURL:      defaultNAG,
		Blockchain:  defaultChain,
		Data:        make(map[string]interface{}),
		IntervalSec: 2,
		httpClient:  httpClient, // Use shared client by default
	}
}

// Open an account, setting its address.
func (acc *CEPAccount) Open(address string) error {
	if address == "" {
		return errors.New("invalid address format: address cannot be empty")
	}
	acc.Address = address
	return nil
}

// UpdateAccountResponsePayload defines the structure for the "Response" part of UpdateAccount API.
type UpdateAccountResponsePayload struct {
	Nonce int64 `json:"Nonce"`
}

// UpdateAccountAPIResponse defines the structure for the UpdateAccount API response.
type UpdateAccountAPIResponse struct {
	Result   int                        `json:"Result"` // HTTP status like codes (200 for success)
	Response UpdateAccountResponsePayload `json:"Response"`
	Message  string                     `json:"Message"` // Optional message
}

// UpdateAccount retrieves account info and updates the Nonce.
func (acc *CEPAccount) UpdateAccount() (bool, error) {
	if acc.Address == "" {
		return false, errors.New("account is not open")
	}

	requestPayload := map[string]string{
		"Blockchain": hexFix(acc.Blockchain),
		"Address":    hexFix(acc.Address),
		"Version":    acc.CodeVersion,
	}
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return false, fmt.Errorf("error marshalling request: %w", err)
	}

	url := acc.NAGURL + "Circular_GetWalletNonce_" + acc.NetworkNode
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := acc.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("http error! status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var apiResponse UpdateAccountAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return false, fmt.Errorf("error decoding response: %w", err)
	}

	if apiResponse.Result == 200 { // Assuming 200 is the success code from "Result" field
		acc.Nonce = apiResponse.Response.Nonce + 1
		return true, nil
	}
	
	errMsg := "invalid response format or missing Nonce field"
	if apiResponse.Message != "" {
		errMsg = apiResponse.Message
	}
	return false, fmt.Errorf("API error (Result %d): %s", apiResponse.Result, errMsg)
}

// SetNetworkResponse defines the structure for the SetNetwork API response.
type SetNetworkResponse struct {
	Status  string `json:"status"`
	URL     string `json:"url"`
	Message string `json:"message"` // For error messages
}

// SetNetwork selects the blockchain network.
// network can be 'devnet', 'testnet', 'mainnet' or a custom one.
func (acc *CEPAccount) SetNetwork(network string) error {
	nagURL := networkURL + url.QueryEscape(network) // Use url.QueryEscape for safe param encoding

	req, err := http.NewRequest("GET", nagURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := acc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error fetching network URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error! status: %d", resp.StatusCode)
	}

	var data SetNetworkResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("error decoding network response: %w", err)
	}

	if data.Status == "success" && data.URL != "" {
		acc.NAGURL = data.URL
		return nil
	}
	
	if data.Message != "" {
		return errors.New(data.Message)
	}
	return errors.New("failed to get URL or invalid response")
}

// SetBlockchain selects the blockchain.
func (acc *CEPAccount) SetBlockchain(chain string) {
	acc.Blockchain = chain
}

// Close resets the account fields to their default values.
func (acc *CEPAccount) Close() {
	acc.Address = ""
	acc.PublicKey = ""
	acc.Info = nil
	acc.LastError = ""
	acc.NAGURL = defaultNAG
	acc.NetworkNode = ""
	acc.Blockchain = defaultChain
	acc.LatestTxID = ""
	acc.Nonce = 0
	acc.Data = make(map[string]interface{})
	acc.IntervalSec = 2
}

// SignData signs data using the account's private key.
// dataHexHash: The hex string of the hash of the data to be signed.
// privateKeyHex: The private key in hex format.
func (acc *CEPAccount) SignData(dataHexHash string, privateKeyHex string) (string, error) {
	if acc.Address == "" {
		return "", errors.New("account is not open")
	}

	privKeyBytes, err := hex.DecodeString(hexFix(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("invalid private key hex: %w", err)
	}

	curve := elliptic.P256() // secp256k1
	priv := new(ecdsa.PrivateKey)
	priv.D = new(big.Int).SetBytes(privKeyBytes)
	priv.PublicKey.Curve = curve
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(privKeyBytes)

	hashBytes, err := hex.DecodeString(dataHexHash) // dataHexHash is already a hash
	if err != nil {
		return "", fmt.Errorf("invalid data hex hash: %w", err)
	}

	signatureDER, err := ecdsa.SignASN1(rand.Reader, priv, hashBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	return hex.EncodeToString(signatureDER), nil
}

// makeTransactionAPICall is a helper for GetTransaction and GetTransactionByID
func (acc *CEPAccount) makeTransactionAPICall(endpointSuffix string, payload map[string]string) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	url := acc.NAGURL + endpointSuffix + acc.NetworkNode
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := acc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http error! status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return result, nil
}


// GetTransaction searches a transaction by its ID within a specific block.
// blockNum: block where the transaction was saved.
// txID: transaction ID.
func (acc *CEPAccount) GetTransaction(blockNum int, txID string) (map[string]interface{}, error) {
	payload := map[string]string{
		"Blockchain": hexFix(acc.Blockchain),
		"ID":         hexFix(txID),
		"Start":      strconv.Itoa(blockNum),
		"End":        strconv.Itoa(blockNum),
		"Version":    acc.CodeVersion,
	}
	return acc.makeTransactionAPICall("Circular_GetTransactionbyID_", payload)
}


// GetTransactionByID searches a transaction by its ID within a range of blocks.
// txID: transaction ID.
// startBlock: starting block.
// endBlock: end block. If endBlock = 0, startBlock indicates the number of blocks from the last minted block.
func (acc *CEPAccount) GetTransactionByID(txID string, startBlock int, endBlock int) (map[string]interface{}, error) {
	payload := map[string]string{
		"Blockchain": hexFix(acc.Blockchain),
		"ID":         hexFix(txID),
		"Start":      strconv.Itoa(startBlock),
		"End":        strconv.Itoa(endBlock),
		"Version":    acc.CodeVersion,
	}
	return acc.makeTransactionAPICall("Circular_GetTransactionbyID_", payload)
}

// SubmitCertificate submits data to the blockchain.
// pdata: data string to be certified.
// privateKey: private key associated to the account.
func (acc *CEPAccount) SubmitCertificate(pdata string, privateKeyHex string) (map[string]interface{}, error) {
	if acc.Address == "" {
		return nil, errors.New("account is not open")
	}
	
	// Ensure Nonce is up-to-date before submitting
	// The JS version does not explicitly call updateAccount here, it assumes Nonce is current.
	// For robustness, one might add:
	// if _, err := acc.UpdateAccount(); err != nil {
	// 	 return nil, fmt.Errorf("failed to update account nonce before submitting: %w", err)
	// }
	// However, to match JS, we use current acc.Nonce and expect it to be managed correctly by the user.

	payloadObject := map[string]string{
		"Action": "CP_CERTIFICATE",
		"Data":   stringToHex(pdata),
	}
	jsonPayloadObject, err := json.Marshal(payloadObject)
	if err != nil {
		return nil, fmt.Errorf("error marshalling payload object: %w", err)
	}
	
	payloadHex := stringToHex(string(jsonPayloadObject))
	timestamp := getFormattedTimestamp()
	
	// Construct string for ID generation: Blockchain + FromAddr + ToAddr + PayloadHex + Nonce + Timestamp
	// Ensure Nonce is stringified
	strToHash := hexFix(acc.Blockchain) +
		hexFix(acc.Address) +
		hexFix(acc.Address) + // "To" is same as "From" for certificates
		payloadHex +
		strconv.FormatInt(acc.Nonce, 10) +
		timestamp

	hash := sha256.Sum256([]byte(strToHash))
	idHex := hex.EncodeToString(hash[:])

	signature, err := acc.SignData(idHex, privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("error signing data: %w", err)
	}

	finalPayload := map[string]string{
		"ID":         idHex,
		"From":       hexFix(acc.Address),
		"To":         hexFix(acc.Address),
		"Timestamp":  timestamp,
		"Payload":    payloadHex, // This should be string(Payload) as per JS, not stringToHex again. Payload is already hex.
		"Nonce":      strconv.FormatInt(acc.Nonce, 10),
		"Signature":  signature,
		"Blockchain": hexFix(acc.Blockchain),
		"Type":       "C_TYPE_CERTIFICATE",
		"Version":    acc.CodeVersion,
	}

	jsonData, err := json.Marshal(finalPayload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling final payload: %w", err)
	}

	url := acc.NAGURL + "Circular_AddTransaction_" + acc.NetworkNode
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := acc.httpClient.Do(req)
	if err != nil {
		// Mimic JS error structure for this specific case
		return map[string]interface{}{
			"success": false,
			"message": "Server unreachable or request failed",
			"error":   err.Error(),
		}, nil // Return nil error for the function itself, as the map contains the error info
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		// Mimic JS error structure
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Network response was not ok, status: %d", resp.StatusCode),
			"error":   string(bodyBytes),
		}, nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return result, nil
}


// GetTransactionOutcome polls for transaction finality.
// txID: Transaction ID.
// timeoutSec: Waiting timeout in seconds.
func (acc *CEPAccount) GetTransactionOutcome(txID string, timeoutSec int) (map[string]interface{}, error) {
	overallTimeout := time.Duration(timeoutSec) * time.Second
	ctx, cancelOverallTimeout := context.WithTimeout(context.Background(), overallTimeout)
	defer cancelOverallTimeout()

	ticker := time.NewTicker(time.Duration(acc.IntervalSec) * time.Second)
	defer ticker.Stop()

	// Perform an initial check almost immediately, then rely on ticker.
	// Or, to match JS, the first check happens after `intervalSec`.
	// The JS code sets timeout for first check *after* interval: `setTimeout(checkTransaction, interval);`
	// So we let the ticker do its first tick.

	log.Printf("GetTransactionOutcome: Starting to poll for TxID %s with timeout %v and interval %ds", txID, overallTimeout, acc.IntervalSec)

	for {
		select {
		case <-ctx.Done(): // Overall timeout exceeded
			log.Printf("GetTransactionOutcome: Timeout exceeded for TxID %s", txID)
			return nil, errors.New("timeout exceeded while polling for transaction outcome")
		case <-ticker.C:
			log.Printf("GetTransactionOutcome: Checking transaction %s...", txID)

			// Use a shorter timeout for each individual API call within the polling loop
			callCtx, callCancel := context.WithTimeout(ctx, time.Duration(acc.IntervalSec*2)*time.Second) // e.g., twice the interval or a fixed value
			
			// The JS used (0, 10) for Start, End. Assuming this is a sensible default for polling recent transactions.
			apiResponse, err := acc.getTransactionByIDWithContext(callCtx, txID, 0, 10)
			callCancel() // Release context resources for the API call

			if err != nil {
				log.Printf("GetTransactionOutcome: Error fetching transaction %s: %v. Continuing to poll.", txID, err)
				// Continue polling on error, relying on overall timeout
				continue
			}

			log.Printf("GetTransactionOutcome: Data received for TxID %s: %v", txID, apiResponse)

			resultVal, ok := apiResponse["Result"].(float64) // JSON numbers are often float64
			if !ok || int(resultVal) != 200 {
				log.Printf("GetTransactionOutcome: API call for TxID %s did not return Result 200 (got %v). Polling again.", txID, resultVal)
				continue
			}

			responseField, hasResponseField := apiResponse["Response"]
			if !hasResponseField {
				log.Printf("GetTransactionOutcome: Response field missing for TxID %s. Polling again.", txID)
				continue
			}

			// Case 1: Response is a string "Transaction Not Found"
			if responseStr, isStr := responseField.(string); isStr {
				if responseStr == "Transaction Not Found" {
					log.Printf("GetTransactionOutcome: Transaction %s Not Found. Polling again.", txID)
					continue
				}
				// Some other unexpected string
				log.Printf("GetTransactionOutcome: TxID %s - Unexpected string in Response field: %s. Polling again.", txID, responseStr)
				continue
			}
			
			// Case 2: Response is an object (map[string]interface{})
			if responseMap, isMap := responseField.(map[string]interface{}); isMap {
				statusVal, hasStatus := responseMap["Status"]
				if !hasStatus {
					log.Printf("GetTransactionOutcome: TxID %s - Status field missing in Response object. Polling again.", txID)
					continue
				}
				statusStr, isStatusStr := statusVal.(string)
				if !isStatusStr {
					log.Printf("GetTransactionOutcome: TxID %s - Status field is not a string. Polling again.", txID)
					continue
				}

				if statusStr != "Pending" {
					log.Printf("GetTransactionOutcome: Transaction %s confirmed with status '%s'. Response: %v", txID, statusStr, responseMap)
					return responseMap, nil // Success! Transaction is finalized (not pending)
				}
				// Status is "Pending"
				log.Printf("GetTransactionOutcome: Transaction %s status is Pending. Polling again.", txID)
				continue
			}

			// If Response field is neither a known string nor a map that leads to resolution
			log.Printf("GetTransactionOutcome: TxID %s - Unexpected type or structure for Response field. Polling again. Type: %T, Value: %v", txID, responseField, responseField)
		}
	}
}

// getTransactionByIDWithContext is a version of GetTransactionByID that accepts a context for cancellation/timeout.
func (acc *CEPAccount) getTransactionByIDWithContext(ctx context.Context, txID string, startBlock int, endBlock int) (map[string]interface{}, error) {
	payload := map[string]string{
		"Blockchain": hexFix(acc.Blockchain),
		"ID":         hexFix(txID),
		"Start":      strconv.Itoa(startBlock),
		"End":        strconv.Itoa(endBlock),
		"Version":    acc.CodeVersion,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	url := acc.NAGURL + "Circular_GetTransactionbyID_" + acc.NetworkNode
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request with context: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := acc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("http error! status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return result, nil
}
