package cepaccount

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"circular-api/lib/ccertificate"
	"circular-api/lib/utils"
)

// CEPAccount manages a user's account on the Circular network.
type CEPAccount struct {
	address       string
	publicKey     string
	info          interface{}
	codeVersion   string
	lastError     string
	nagUrl        string
	networkNode   string
	blockchain    string
	latestTxID    string
	nonce         int
	data          map[string]interface{}
	intervalSec   int
}

// NewCEPAccount creates and initializes a new CEPAccount object.
func NewCEPAccount() *CEPAccount {
	return &CEPAccount{
		codeVersion: utils.LIB_VERSION,
		nagUrl:      utils.DEFAULT_NAG,
		blockchain:  utils.DEFAULT_CHAIN,
		data:        make(map[string]interface{}),
		intervalSec: 2,
	}
}

// Open initializes the account with a given address.
func (a *CEPAccount) Open(accountAddress string) bool {
	a.address = accountAddress
	return true
}

// Close clears the account state.
func (a *CEPAccount) Close() {
	a.address = ""
	a.publicKey = ""
	a.info = nil
	a.lastError = ""
	a.nagUrl = utils.DEFAULT_NAG
	a.networkNode = ""
	a.blockchain = utils.DEFAULT_CHAIN
	a.latestTxID = ""
	a.data = make(map[string]interface{})
	a.intervalSec = 2
	a.nonce = 0
}

// SetNetwork configures the network environment.
func (a *CEPAccount) SetNetwork(network string) bool {
	endpoint := utils.NETWORK_URL + url.QueryEscape(network)
	resp, err := http.Get(endpoint)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to set network: %v", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		a.lastError = fmt.Sprintf("Failed to set network: HTTP error! status: %d", resp.StatusCode)
		return false
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to set network: could not read response body: %v", err)
		return false
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		a.lastError = fmt.Sprintf("Failed to set network: could not parse JSON response: %v", err)
		return false
	}

	if status, ok := result["status"].(string); !ok || status != "success" {
		if msg, ok := result["message"].(string); ok {
			a.lastError = fmt.Sprintf("Failed to set network: API error: %s", msg)
		} else {
			a.lastError = "Failed to set network: API returned an error"
		}
		return false
	}

	if newURL, ok := result["url"].(string); ok {
		a.nagUrl = newURL
		return true
	}

	a.lastError = "Failed to set network: URL not found in response"
	return false
}

// SetBlockchain sets the target blockchain address.
func (a *CEPAccount) SetBlockchain(chain string) {
	a.blockchain = chain
}

// UpdateAccount fetches the latest nonce from the network.
func (a *CEPAccount) UpdateAccount() bool {
	payload := map[string]interface{}{
		"Blockchain": utils.HexFix(a.blockchain),
		"Address":    utils.HexFix(a.address),
		"Version":    a.codeVersion,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to update account: could not create payload: %v", err)
		return false
	}

	endpoint := a.nagUrl + "/Circular_GetWalletNonce_" + a.networkNode
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to update account: network error: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to update account: could not read response body: %v", err)
		return false
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		a.lastError = fmt.Sprintf("Failed to update account: could not parse JSON response: %v", err)
		return false
	}

	if res, ok := result["Result"].(float64); !ok || res != 200 {
		a.lastError = "Failed to update account: Invalid response format or error result"
		return false
	}

	if response, ok := result["Response"].(map[string]interface{}); ok {
		if nonce, ok := response["Nonce"].(float64); ok {
			a.nonce = int(nonce) + 1
			return true
		}
	}

	a.lastError = "Failed to update account: Invalid response format or missing Nonce field"
	return false
}

// GetLastError returns the last error message.
func (a *CEPAccount) GetLastError() string {
	return a.lastError
}

// SetData sets a key-value pair in the account's data store.
func (a *CEPAccount) SetData(key string, value interface{}) {
	a.data[key] = value
}

// GetData retrieves a value from the account's data store.
func (a *CEPAccount) GetData(key string) (interface{}, bool) {
	val, exists := a.data[key]
	return val, exists
}

// GetAddress returns the account's address.
func (a *CEPAccount) GetAddress() string {
	return a.address
}

// GetLatestTxID returns the last successful transaction ID.
func (a *CEPAccount) GetLatestTxID() string {
	return a.latestTxID
}

// GetNonce returns the current nonce for the account.
func (a *CEPAccount) GetNonce() int {
	return a.nonce
}

// SubmitCertificate creates and submits a transaction to the network.
func (a *CEPAccount) SubmitCertificate(data string, privateKey string) (map[string]interface{}, error) {
	// 1. Create the certificate
	cert := &ccertificate.CCertificate{}
	cert.SetData(data)
	cert.PreviousTxID = a.latestTxID

	jsonCert, err := cert.GetJSONCertificate()
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to create JSON certificate: %v", err)
		return nil, err
	}

	// 2. Get public key
	pubKey, err := getPublicKey(privateKey)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to get public key: %v", err)
		return nil, err
	}
	a.publicKey = pubKey

	// 3. Create the transaction payload
	txPayload := map[string]interface{}{
		"Certificate": jsonCert,
		"Blockchain":  utils.HexFix(a.blockchain),
		"Address":     utils.HexFix(a.address),
		"PublicKey":   utils.HexFix(a.publicKey),
		"Nonce":       a.nonce,
		"Version":     a.codeVersion,
		"Timestamp":   utils.GetFormattedTimestamp(),
	}

	// 4. Sign the transaction
	jsonPayload, err := json.Marshal(txPayload)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to marshal transaction payload: %v", err)
		return nil, err
	}

	signature, err := sign(privateKey, string(jsonPayload))
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to sign transaction: %v", err)
		return nil, err
	}

	// 5. Create the final submission payload
	submissionPayload := map[string]interface{}{
		"TxPayload": string(jsonPayload),
		"Signature": signature,
	}

	jsonSubmission, err := json.Marshal(submissionPayload)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to marshal submission payload: %v", err)
		return nil, err
	}

	// 6. Submit the transaction
	endpoint := a.nagUrl + "/Circular_SubmitCertificate_" + a.networkNode
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonSubmission))
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to submit certificate: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to read response body: %v", err)
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		a.lastError = fmt.Sprintf("Failed to parse JSON response: %v", err)
		return nil, err
	}

	if res, ok := result["Result"].(float64); !ok || res != 200 {
		a.lastError = "Certificate submission failed"
		return result, fmt.Errorf(a.lastError)
	}

	if response, ok := result["Response"].(map[string]interface{}); ok {
		if txID, ok := response["TxID"].(string); ok {
			a.latestTxID = txID
			a.nonce++
		}
	}

	return result, nil
}

// GetTransactionOutcome retrieves the outcome of a specific transaction.
func (a *CEPAccount) GetTransactionOutcome(txID string) (map[string]interface{}, error) {
	// 1. Create the payload
	payload := map[string]interface{}{
		"TxID":      txID,
		"Version":   a.codeVersion,
		"Timestamp": utils.GetFormattedTimestamp(),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to marshal transaction outcome payload: %v", err)
		return nil, err
	}

	// 2. Make the request
	endpoint := a.nagUrl + "/Circular_GetTransactionOutcome_" + a.networkNode
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to get transaction outcome: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		a.lastError = fmt.Sprintf("Failed to read response body: %v", err)
		return nil, err
	}

	// 3. Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		a.lastError = fmt.Sprintf("Failed to parse JSON response: %v", err)
		return nil, err
	}

	return result, nil
}

// WaitForTransactionOutcome polls for the outcome of a transaction until it's confirmed or fails.
func (a *CEPAccount) WaitForTransactionOutcome(txID string, timeoutSec int) (map[string]interface{}, error) {
	startTime := time.Now()
	for {
		outcome, err := a.GetTransactionOutcome(txID)
		if err == nil {
			if response, ok := outcome["Response"].(map[string]interface{}); ok {
				if status, ok := response["Status"].(string); ok {
					if status == "Confirmed" || status == "Failed" {
						return outcome, nil
					}
				}
			}
		}

		if time.Since(startTime).Seconds() > float64(timeoutSec) {
			a.lastError = "Timeout waiting for transaction outcome"
			return nil, fmt.Errorf(a.lastError)
		}

		time.Sleep(time.Duration(a.intervalSec) * time.Second)
	}
}

// SetNagURL is a test helper to set the nagUrl field.
func (a *CEPAccount) SetNagURL(url string) {
	a.nagUrl = url
}

// SetNonceForTest is a test helper to set the nonce field.
func (a *CEPAccount) SetNonceForTest(nonce int) {
	a.nonce = nonce
}

// SetIntervalSec is a test helper to set the intervalSec field.
func (a *CEPAccount) SetIntervalSec(sec int) {
	a.intervalSec = sec
}
