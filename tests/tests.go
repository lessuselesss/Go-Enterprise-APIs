package circular_test // Convention: tests are in package_test

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	// Import the package being tested.
	// The import path depends on your go.mod file.
	// If go.mod is `module example.com/my_circular_project`, then:
	"example.com/my_circular_project/circular"
	// If your go.mod is `module my_circular_project`, then:
	// "my_circular_project/circular"
)

// Constants from the library (for verification, ensure they match the actual library)
const (
	testLibVersion      = "1.0.13"
	testDefaultNAGBase  = "https://nag.circularlabs.io"
	testDefaultNAGPath  = "/NAG.php?cep="
	testDefaultNAG      = testDefaultNAGBase + testDefaultNAGPath
	testDefaultChain    = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
	testNetworkInfoBase = "https://circularlabs.io"
	testNetworkInfoPath = "/network/getNAG"
)

// Global test variables for keys and address
var (
	testPrivateKeyHex string
	testPublicKeyHex  string
	testAccountAddr   string
)

// init generates a test key pair and address once for all tests
func init() {
	// Equivalent to JS: const ec = new EC('secp256k1'); const testKeyPair = ec.genKeyPair();
	// Go uses elliptic.P256() for secp256k1 equivalent by common crypto libraries.
	// Note: elliptic.P256() is NIST P-256. If the JS 'secp256k1' refers to the Bitcoin curve,
	// you'd need a specific library like `github.com/btcsuite/btcd/btcec`.
	// For now, assuming elliptic.P256() is acceptable or that the library's crypto matches it.
	// If it must be the specific secp256k1 curve used by Bitcoin:
	// import "github.com/decred/dcrd/dcrec/secp256k1/v4"
	// privKey, _ := secp256k1.GeneratePrivateKey()
	// pubKeyBytes := privKey.PubKey().SerializeCompressed() or Uncompressed
	// For this example, let's use standard crypto/elliptic P256
	// If exact secp256k1 (Bitcoin one) is needed, replace this with a btcec or similar import.

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate test key pair: %v", err)
	}
	testPrivateKeyHex = hex.EncodeToString(privKey.D.Bytes())
	// Public key in uncompressed format (0x04 + X + Y)
	pubKeyBytes := elliptic.Marshal(elliptic.P256(), privKey.X, privKey.Y)
	testPublicKeyHex = hex.EncodeToString(pubKeyBytes)

	// Address generation: sha256(publicKey).substring(0, 40)
	// Note: The JS sha256 often takes a string and returns a hex string.
	// Go's sha256.Sum256 takes []byte and returns [32]byte.
	// The JS version probably hashes the *hex string* of the public key. Let's clarify this.
	// Assuming it hashes the *bytes* of the public key.
	// If it hashes the hex string of the public key, the logic would be:
	// addrHash := sha256.Sum256([]byte(testPublicKeyHex))
	addrHash := sha256.Sum256(pubKeyBytes)
	testAccountAddr = "0x" + hex.EncodeToString(addrHash[:])[:40] // First 40 hex chars (20 bytes)
}

// Helper function to create a mock HTTP server
// Returns the server, its URL, and a function to close it
func startMockServer() (server *httptest.Server, serverURL string, mockResponses map[string]func(w http.ResponseWriter, r *http.Request)) {
	mockResponses = make(map[string]func(w http.ResponseWriter, r *http.Request))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathWithQuery := r.URL.Path
		if r.URL.RawQuery != "" {
			pathWithQuery += "?" + r.URL.RawQuery
		}
		// Generic matcher: first by full path + query, then by path only
		if respFunc, ok := mockResponses[pathWithQuery]; ok {
			respFunc(w, r)
			return
		}
		if respFunc, ok := mockResponses[r.URL.Path]; ok {
			respFunc(w, r)
			return
		}
		// Fallback for NAG.php?cep= style calls if only base path is registered
		if strings.HasPrefix(r.URL.Path, testDefaultNAGPath) {
			basePath := testDefaultNAGPath // The part registered in mockResponses might be just this
			if respFunc, ok := mockResponses[basePath+r.URL.Query().Get("cep")]; ok { // e.g. /NAG.php?cep=Circular_GetWalletNonce_
				respFunc(w,r)
				return
			}
		}

		log.Printf("Unhandled mock request: %s %s", r.Method, pathWithQuery)
		http.NotFound(w, r)
	})
	server = httptest.NewServer(handler)
	serverURL = server.URL
	return
}

// TestCCertificate mirrors the "C_CERTIFICATE Class" describe block
func TestCCertificate(t *testing.T) {
	t.Run("should initialize with default values", func(t *testing.T) {
		cert := circular.NewCCertificate()
		if cert.Data != "" { // In Go, uninitialized string is "", JS null for object properties
			t.Errorf("Expected cert.Data to be empty, got %s", cert.Data)
		}
		if cert.PreviousTxID != "" {
			t.Errorf("Expected cert.PreviousTxID to be empty, got %s", cert.PreviousTxID)
		}
		if cert.PreviousBlock != "" {
			t.Errorf("Expected cert.PreviousBlock to be empty, got %s", cert.PreviousBlock)
		}
		if cert.Version != testLibVersion {
			t.Errorf("Expected cert.Version to be %s, got %s", testLibVersion, cert.Version)
		}
	})

	t.Run("setData should store data as hex", func(t *testing.T) {
		cert := circular.NewCCertificate()
		testData := "test data is a string"
		cert.SetData(testData)
		expectedHex := hex.EncodeToString([]byte(testData))
		if cert.Data != expectedHex {
			t.Errorf("Expected cert.Data to be %s, got %s", expectedHex, cert.Data)
		}
	})

	t.Run("getData", func(t *testing.T) {
		t.Run("should retrieve original data for simple strings", func(t *testing.T) {
			cert := circular.NewCCertificate()
			originalData := "another test"
			cert.SetData(originalData)
			if cert.GetData() != originalData {
				t.Errorf("Expected cert.GetData() to be %s, got %s", originalData, cert.GetData())
			}
		})

		t.Run("should return empty string if data is null or empty hex", func(t *testing.T) {
			cert := circular.NewCCertificate() // Data is "" (empty) by default
			if cert.GetData() != "" {
				t.Errorf("Expected GetData() to be empty for nil data, got %s", cert.GetData())
			}
			cert.Data = ""
			if cert.GetData() != "" {
				t.Errorf("Expected GetData() to be empty for empty hex data, got %s", cert.GetData())
			}
		})

		t.Run("should return empty string if data is 0x", func(t *testing.T) {
			cert := circular.NewCCertificate()
			cert.Data = "0x"
			if cert.GetData() != "" {
				// The Go hexToString(hexFix("0x")) will result in empty string.
				t.Errorf("Expected GetData() to be empty for '0x' data, got %s", cert.GetData())
			}
		})

		// This test highlights the UTF-8 handling difference mentioned in JS tests.
		// Go's `hex.EncodeToString([]byte(s))` and `hex.DecodeString()` are UTF-8 correct.
		// The original JS `stringToHex` was not. So this test should PASS in Go.
		t.Run("should correctly retrieve multi-byte unicode data", func(t *testing.T) {
			cert := circular.NewCCertificate()
			unicodeData := "你好世界 😊"
			cert.SetData(unicodeData)
			if cert.GetData() != unicodeData {
				t.Errorf("Expected GetData() for unicode to be '%s', got '%s'", unicodeData, cert.GetData())
			}
		})
	})

	t.Run("getJSONCertificate should return a valid JSON string", func(t *testing.T) {
		cert := circular.NewCCertificate()
		testData := "json test"
		cert.SetData(testData)
		cert.PreviousTxID = "tx123"
		cert.PreviousBlock = "block456"

		jsonCertStr, err := cert.GetJSONCertificate()
		if err != nil {
			t.Fatalf("GetJSONCertificate returned an error: %v", err)
		}

		var parsedCert map[string]interface{}
		err = json.Unmarshal([]byte(jsonCertStr), &parsedCert)
		if err != nil {
			t.Fatalf("Failed to parse JSON certificate string: %v", err)
		}

		expectedHexData := hex.EncodeToString([]byte(testData))
		expectedMap := map[string]interface{}{
			"data":          expectedHexData,
			"previousTxID":  "tx123",
			"previousBlock": "block456",
			"version":       testLibVersion,
		}

		if !reflect.DeepEqual(parsedCert, expectedMap) {
			t.Errorf("Parsed JSON certificate mismatch.\nExpected: %+v\nGot:      %+v", expectedMap, parsedCert)
		}
	})

	t.Run("getCertificateSize should return correct byte length", func(t *testing.T) {
		cert := circular.NewCCertificate()
		testData := "size test"
		cert.SetData(testData)
		cert.PreviousTxID = "txIDForSize"
		cert.PreviousBlock = "blockIDForSize"

		expectedHexData := hex.EncodeToString([]byte(testData))
		expectedJSON := fmt.Sprintf(`{"data":"%s","previousTxID":"txIDForSize","previousBlock":"blockIDForSize","version":"%s"}`,
			expectedHexData, testLibVersion)
		// Need to ensure key order for exact size match if JSON keys are reordered by Go's marshaller.
		// However, C_CERTIFICATE struct has fixed fields, so order should be consistent.
		// A more robust way is to marshal the expected struct.
		expectedStruct := circular.CCertificate{
			Data: expectedHexData,
			PreviousTxID: "txIDForSize",
			PreviousBlock: "blockIDForSize",
			Version: testLibVersion,
		}
		expectedBytes, _ := json.Marshal(expectedStruct)
		expectedSize := len(expectedBytes)


		size, err := cert.GetCertificateSize()
		if err != nil {
			t.Fatalf("GetCertificateSize returned an error: %v", err)
		}
		if size != expectedSize {
			t.Errorf("Expected certificate size to be %d, got %d. Expected JSON: %s, Got JSON: %s",
				expectedSize, size, string(expectedBytes), cert.GetJSONCertificate())
		}
	})
}

// TestCEPAccount mirrors the "CEP_Account Class" describe block
func TestCEPAccount(t *testing.T) {
	mockAddress := testAccountAddr
	mockPrivateKey := testPrivateKeyHex

	// Mock server setup for CEPAccount tests
	server, serverURL, mockResponses := startMockServer()
	defer server.Close()

	// defaultNagURLForTest will be the mock server's URL + the default path
	defaultNagURLForTest := serverURL + testDefaultNAGPath

	t.Run("should initialize with default values", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		if acc.Address != "" {
			t.Errorf("Expected acc.Address to be empty, got %s", acc.Address)
		}
		if acc.PublicKey != "" {
			t.Errorf("Expected acc.PublicKey to be empty, got %s", acc.PublicKey)
		}
		if acc.Info != nil {
			t.Errorf("Expected acc.Info to be nil, got %+v", acc.Info)
		}
		if acc.CodeVersion != testLibVersion {
			t.Errorf("Expected acc.CodeVersion to be %s, got %s", testLibVersion, acc.CodeVersion)
		}
		if acc.LastError != "" {
			t.Errorf("Expected acc.LastError to be empty, got %s", acc.LastError)
		}
		// Default NAG URL uses the const from the library, not the test const
		if acc.NAGURL != circular.DefaultNAG() { // Assuming DefaultNAG() is an exported way to get it
			t.Errorf("Expected acc.NAGURL to be %s, got %s", circular.DefaultNAG(), acc.NAGURL)
		}
		if acc.NetworkNode != "" {
			t.Errorf("Expected acc.NetworkNode to be empty, got %s", acc.NetworkNode)
		}
		if acc.Blockchain != circular.DefaultChain() { // Assuming DefaultChain()
			t.Errorf("Expected acc.Blockchain to be %s, got %s", circular.DefaultChain(), acc.Blockchain)
		}
		if acc.LatestTxID != "" {
			t.Errorf("Expected acc.LatestTxID to be empty, got %s", acc.LatestTxID)
		}
		if acc.Nonce != 0 {
			t.Errorf("Expected acc.Nonce to be 0, got %d", acc.Nonce)
		}
		if len(acc.Data) != 0 {
			t.Errorf("Expected acc.Data to be empty, got %+v", acc.Data)
		}
		if acc.IntervalSec != 2 {
			t.Errorf("Expected acc.IntervalSec to be 2, got %d", acc.IntervalSec)
		}
	})

	t.Run("open", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		t.Run("should set the account address", func(t *testing.T) {
			err := acc.Open(mockAddress)
			if err != nil {
				t.Fatalf("acc.Open failed: %v", err)
			}
			if acc.Address != mockAddress {
				t.Errorf("Expected acc.Address to be %s, got %s", mockAddress, acc.Address)
			}
		})
		t.Run("should return an error for invalid address format", func(t *testing.T) {
			err := acc.Open("") // JS used null, Go empty string is common
			if err == nil || !strings.Contains(err.Error(), "invalid address format") { // Check for specific error
				t.Errorf("Expected error for empty address, got %v", err)
			}
			// Note: Go's static typing prevents passing int/object directly like JS.
			// The JS test `account.open(123)` isn't directly translatable unless Open accepts interface{}.
			// The current Go `Open` takes `string`, so type mismatch would be a compile error.
		})
	})

	t.Run("close", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		// Modify some fields first
		acc.Open(mockAddress)
		acc.NAGURL = "http://custom.url"
		acc.Nonce = 10

		acc.Close()

		if acc.Address != "" {
			t.Errorf("Expected acc.Address to be empty after close, got %s", acc.Address)
		}
		if acc.NAGURL != circular.DefaultNAG() {
			t.Errorf("Expected acc.NAGURL to be default after close, got %s", acc.NAGURL)
		}
		if acc.Nonce != 0 {
			t.Errorf("Expected acc.Nonce to be 0 after close, got %d", acc.Nonce)
		}
		// ... check other fields reset by Close()
	})

	t.Run("setBlockchain", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		newChain := "0xmynewchain"
		acc.SetBlockchain(newChain)
		if acc.Blockchain != newChain {
			t.Errorf("Expected acc.Blockchain to be %s, got %s", newChain, acc.Blockchain)
		}
	})

	t.Run("setNetwork", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		// The library's NETWORK_URL constant will be used by acc.SetNetwork
		// We need to mock responses from that base URL.
		// The mock server setup is a bit global. For specific endpoint mocking:
		networkInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryNet := r.URL.Query().Get("network")
			var responseObj struct {
				Status  string `json:"status"`
				URL     string `json:"url,omitempty"`
				Message string `json:"message,omitempty"`
			}
			switch queryNet {
			case "mainnet":
				responseObj.Status = "success"
				responseObj.URL = serverURL + "/mainnet-nag-api/" // Mocked distinct URL
			case "testnet":
				responseObj.Status = "success"
				responseObj.URL = serverURL + "/testnet-nag-api/"
			case "devnet":
				responseObj.Status = "success"
				responseObj.URL = serverURL + "/devnet-nag-api/"
			case "brokennet":
				http.Error(w, "Server Error", http.StatusInternalServerError)
				return
			case "failednet":
				responseObj.Status = "error"
				responseObj.Message = "Invalid network specified"
			default:
				http.NotFound(w,r)
				return
			}
			json.NewEncoder(w).Encode(responseObj)
		}))
		defer networkInfoServer.Close()

		// Temporarily override the library's NETWORK_URL to point to our mock network info server
		originalNetworkURL := circular.GetNetworkURLBase() // Assume a getter for this
		circular.SetNetworkURLBase(networkInfoServer.URL + testNetworkInfoPath + "?network=") // Assume a setter
		defer circular.SetNetworkURLBase(originalNetworkURL)      // Restore original

		testCases := []struct {
			name        string
			networkArg  string
			expectedURL string // Relative to serverURL or full
			expectError bool
			errorMsg    string
		}{
			{"mainnet", "mainnet", serverURL + "/mainnet-nag-api/", false, ""},
			{"testnet", "testnet", serverURL + "/testnet-nag-api/", false, ""},
			{"devnet", "devnet", serverURL + "/devnet-nag-api/", false, ""},
			{"brokennet", "brokennet", "", true, "http error! status: 500"},
			{"failednet", "failednet", "", true, "Invalid network specified"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				acc.NAGURL = "placeholder" // ensure it changes
				err := acc.SetNetwork(tc.networkArg)
				if tc.expectError {
					if err == nil {
						t.Fatalf("Expected error for network %s, got nil", tc.networkArg)
					}
					if !strings.Contains(err.Error(), tc.errorMsg) {
						t.Errorf("Expected error message to contain '%s', got '%s'", tc.errorMsg, err.Error())
					}
				} else {
					if err != nil {
						t.Fatalf("SetNetwork for %s failed: %v", tc.networkArg, err)
					}
					if acc.NAGURL != tc.expectedURL {
						t.Errorf("Expected NAG_URL to be %s, got %s", tc.expectedURL, acc.NAGURL)
					}
				}
			})
		}
	})

	t.Run("updateAccount", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		acc.NAGURL = defaultNagURLForTest // Point to our mock server's default path
		
		// Test case: successful update
		t.Run("should update Nonce on successful API call", func(t *testing.T) {
			acc.Open(mockAddress) // Open account first
			acc.Nonce = 0 // Reset nonce
			mockAPIResponse := circular.UpdateAccountAPIResponse{
				Result:   200,
				Response: circular.UpdateAccountResponsePayload{Nonce: 5},
			}
			// Define mock response for the specific endpoint (Circular_GetWalletNonce_)
			mockResponses[defaultNagURLForTest+"Circular_GetWalletNonce_"] = func(w http.ResponseWriter, r *http.Request) {
				// Optional: verify request body
				var reqBody map[string]string
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					http.Error(w, "bad request body", http.StatusBadRequest)
					return
				}
				if reqBody["Address"] != circular.HexFix(mockAddress) {
					t.Errorf("Unexpected address in request: got %s, want %s", reqBody["Address"], circular.HexFix(mockAddress))
				}
				json.NewEncoder(w).Encode(mockAPIResponse)
			}

			result, err := acc.UpdateAccount()
			if err != nil {
				t.Fatalf("UpdateAccount failed: %v", err)
			}
			if !result {
				t.Errorf("Expected UpdateAccount result to be true, got false")
			}
			if acc.Nonce != 6 { // 5 (from API) + 1
				t.Errorf("Expected acc.Nonce to be 6, got %d", acc.Nonce)
			}
			delete(mockResponses, defaultNagURLForTest+"Circular_GetWalletNonce_") // Clean up specific mock
		})

		// Test case: API error (Result != 200)
		t.Run("should return false on API error Result!=200", func(t *testing.T) {
			acc.Open(mockAddress)
			initialNonce := acc.Nonce
			mockAPIResponse := circular.UpdateAccountAPIResponse{Result: 400, Message: "Bad Request"}
			mockResponses[defaultNagURLForTest+"Circular_GetWalletNonce_"] = func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(mockAPIResponse)
			}

			result, err := acc.UpdateAccount()
			if err == nil || !strings.Contains(err.Error(), "API error (Result 400): Bad Request") {
				// In Go, the bool is for simple success/fail, error carries details.
				// The JS version returned false. Go version returns (false, error).
				t.Errorf("Expected specific error, got: %v", err)
			}
			if result {
				t.Errorf("Expected UpdateAccount result to be false, got true")
			}
			if acc.Nonce != initialNonce {
				t.Errorf("Expected acc.Nonce to remain %d, got %d", initialNonce, acc.Nonce)
			}
			delete(mockResponses, defaultNagURLForTest+"Circular_GetWalletNonce_")
		})

		// Test case: network error
		t.Run("should return false on network error", func(t *testing.T) {
			acc.Open(mockAddress)
			initialNonce := acc.Nonce
			// No mock response set up, or make server return error
			// Forcing a network error is tricky with httptest if server is running.
			// Instead, we can simulate it by having the mock handler return an error,
			// or by pointing NAGURL to a non-existent server.
			// Here, let's just make the mock handler NOT reply for this specific path.
			// A better way: close server temporarily or use a distinct failing handler.
			// For now, let's assume the http client in the lib has a timeout.
			// This test is harder to make deterministic with httptest without more control.
			// Alternative: point to bad URL
			originalNAG := acc.NAGURL
			acc.NAGURL = "http://localhost:12345/nonexistent" // Point to something that will fail
			
			result, err := acc.UpdateAccount()
			if err == nil {
				t.Errorf("Expected network error, got nil")
			}
			if result {
				t.Errorf("Expected UpdateAccount result to be false on network error, got true")
			}
			if acc.Nonce != initialNonce {
				t.Errorf("Expected acc.Nonce to remain %d, got %d", initialNonce, acc.Nonce)
			}
			acc.NAGURL = originalNAG // Restore
		})
		
		t.Run("should return error if account is not open", func(t *testing.T) {
			freshAcc := circular.NewCEPAccount() // ensure it's closed
			_, err := freshAcc.UpdateAccount()
			if err == nil || !strings.Contains(err.Error(), "account is not open") {
				t.Errorf("Expected 'account is not open' error, got %v", err)
			}
		})

		t.Run("should return false if response malformed (missing Nonce)", func(t *testing.T) {
			acc.Open(mockAddress)
			initialNonce := acc.Nonce
			// Malformed response: Result 200, but Response object doesn't have Nonce field
			malformedAPIResponse := struct {
				Result int
				Response map[string]interface{}
			}{Result: 200, Response: map[string]interface{}{"SomeOtherField": 5}}

			mockResponses[defaultNagURLForTest+"Circular_GetWalletNonce_"] = func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(malformedAPIResponse)
			}

			result, err := acc.UpdateAccount()
			if err == nil || !strings.Contains(err.Error(), "API error (Result 200): Invalid response format or missing Nonce field") {
				// This error message comes from the library.
				t.Errorf("Expected specific error for malformed response, got %v", err)
			}
			if result {
				t.Errorf("Expected UpdateAccount result to be false, got true")
			}
			if acc.Nonce != initialNonce {
				t.Errorf("Expected acc.Nonce to remain %d, got %d", initialNonce, acc.Nonce)
			}
			delete(mockResponses, defaultNagURLForTest+"Circular_GetWalletNonce_")
		})
	})

	t.Run("signData", func(t *testing.T) {
		acc := circular.NewCEPAccount()

		t.Run("should sign data correctly", func(t *testing.T) {
			acc.Open(mockAddress)
			dataToSign := "sample data for signing"
			// In Go, signData takes hashHex. JS took raw data and hashed inside.
			// For parity, let's hash here first.
			dataHash := sha256.Sum256([]byte(dataToSign))
			dataHashHex := hex.EncodeToString(dataHash[:])

			signatureHex, err := acc.SignData(dataHashHex, mockPrivateKey)
			if err != nil {
				t.Fatalf("SignData failed: %v", err)
			}
			if signatureHex == "" {
				t.Errorf("Expected signature to be non-empty")
			}

			// Verification (requires public key corresponding to mockPrivateKey)
			// Need to parse the public key hex
			pubKeyBytes, _ := hex.DecodeString(testPublicKeyHex) // From global test setup
			// Unmarshal returns a generic crypto.PublicKey, so we type assert
			pub, err := unmarshalPublicKey(pubKeyBytes) // Helper needed if not elliptic.Unmarshal
			if err != nil {
				t.Fatalf("Failed to unmarshal public key: %v", err)
			}
			ecdsaPubKey, ok := pub.(*ecdsa.PublicKey)
			if !ok {
				t.Fatalf("Public key is not ECDSA type")
			}

			sigBytes, _ := hex.DecodeString(signatureHex)
			
			// ecdsa.VerifyASN1 expects the hash of the message, not the message itself
			verified := ecdsa.VerifyASN1(ecdsaPubKey, dataHash[:], sigBytes)
			if !verified {
				t.Errorf("Signature verification failed")
			}
		})

		t.Run("should return error if account is not open", func(t *testing.T) {
			freshAcc := circular.NewCEPAccount()
			_, err := freshAcc.SignData("dummyhash", mockPrivateKey)
			if err == nil || !strings.Contains(err.Error(), "account is not open") {
				t.Errorf("Expected 'account is not open' error, got %v", err)
			}
		})

		t.Run("should produce different signatures for different data", func(t *testing.T) {
			acc.Open(mockAddress)
			hash1 := sha256.Sum256([]byte("data1"))
			hash1Hex := hex.EncodeToString(hash1[:])
			sig1, _ := acc.SignData(hash1Hex, mockPrivateKey)

			hash2 := sha256.Sum256([]byte("data2"))
			hash2Hex := hex.EncodeToString(hash2[:])
			sig2, _ := acc.SignData(hash2Hex, mockPrivateKey)

			if sig1 == sig2 {
				t.Errorf("Expected different signatures for different data, but got same")
			}
		})
		
		t.Run("should produce different signatures for different private keys", func(t *testing.T) {
			acc.Open(mockAddress)
			data := "commondata"
			dataHash := sha256.Sum256([]byte(data))
			dataHashHex := hex.EncodeToString(dataHash[:])

			sig1, _ := acc.SignData(dataHashHex, mockPrivateKey)

			// Generate another key pair for this subtest
			otherPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			otherPrivKeyHex := hex.EncodeToString(otherPrivKey.D.Bytes())
			sig2, _ := acc.SignData(dataHashHex, otherPrivKeyHex)

			if sig1 == sig2 {
				t.Errorf("Expected different signatures for different private keys, but got same")
			}
		})
	})

	// Helper for unmarshalling public key if elliptic.Unmarshal isn't directly suitable
	// (e.g. if key was from a different secp256k1 lib or format)
	// For standard crypto/elliptic P256():
	unmarshalPublicKey := func(pubKeyBytes []byte) (interface{}, error) {
		x, y := elliptic.Unmarshal(elliptic.P256(), pubKeyBytes)
		if x == nil {
			return nil, fmt.Errorf("invalid public key bytes for P256")
		}
		return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
	}


	t.Run("getTransaction and getTransactionbyID", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		acc.NAGURL = defaultNagURLForTest // Point to mock server

		txID := "testTxID123"
		blockNum := 100

		endpointPath := defaultNagURLForTest + "Circular_GetTransactionbyID_"

		t.Run("getTransaction(BlockID, TxID) should fetch a transaction", func(t *testing.T) {
			mockResponse := map[string]interface{}{"Result": 200.0, "Response": map[string]interface{}{"id": txID, "status": "Confirmed"}}
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				var reqBody map[string]string
				json.NewDecoder(r.Body).Decode(&reqBody)
				if reqBody["ID"] != circular.HexFix(txID) || reqBody["Start"] != fmt.Sprintf("%d", blockNum) || reqBody["End"] != fmt.Sprintf("%d", blockNum) {
					t.Errorf("Unexpected request body in getTransaction: %+v", reqBody)
				}
				json.NewEncoder(w).Encode(mockResponse)
			}

			result, err := acc.GetTransaction(blockNum, txID)
			if err != nil {
				t.Fatalf("GetTransaction failed: %v", err)
			}
			if !reflect.DeepEqual(result, mockResponse) {
				t.Errorf("GetTransaction response mismatch.\nExpected: %+v\nGot:      %+v", mockResponse, result)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("getTransactionbyID should fetch a transaction within a block range", func(t *testing.T) {
			startBlock, endBlock := 100, 110
			mockResponse := map[string]interface{}{"Result": 200.0, "Response": map[string]interface{}{"id": txID, "status": "Confirmed"}}
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				var reqBody map[string]string
				json.NewDecoder(r.Body).Decode(&reqBody)
				if reqBody["ID"] != circular.HexFix(txID) || 
				   reqBody["Start"] != strconv.Itoa(startBlock) || 
				   reqBody["End"] != strconv.Itoa(endBlock) {
					t.Errorf("Unexpected request body in getTransactionbyID: %+v", reqBody)
				}
				json.NewEncoder(w).Encode(mockResponse)
			}
			result, err := acc.GetTransactionByID(txID, startBlock, endBlock)
			if err != nil {
				t.Fatalf("GetTransactionByID failed: %v", err)
			}
			if !reflect.DeepEqual(result, mockResponse) {
				t.Errorf("GetTransactionByID response mismatch.\nExpected: %+v\nGot:      %+v", mockResponse, result)
			}
			delete(mockResponses, endpointPath)
		})
		
		t.Run("getTransactionbyID should handle Transaction Not Found string response", func(t *testing.T) {
			mockResponse := map[string]interface{}{"Result": 200.0, "Response": "Transaction Not Found"}
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(mockResponse)
			}
			result, err := acc.GetTransactionByID(txID, 0, 10)
			if err != nil {
				t.Fatalf("GetTransactionByID failed: %v", err)
			}
			if !reflect.DeepEqual(result, mockResponse) {
				t.Errorf("GetTransactionByID response mismatch.\nExpected: %+v\nGot:      %+v", mockResponse, result)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("getTransactionbyID should return error on network error", func(t *testing.T) {
			// Simulate network error by having no handler or specific error response
			originalNAG := acc.NAGURL
			acc.NAGURL = "http://localhost:12345/nonexistent" // Point to failure
			_, err := acc.GetTransactionByID(txID, 0, 10)
			if err == nil {
				t.Fatalf("Expected network error, got nil")
			}
			// Check for a generic network error message part
			if !strings.Contains(err.Error(), "request failed") && !strings.Contains(err.Error(), "connection refused") {
				t.Errorf("Expected network related error, got: %v", err)
			}
			acc.NAGURL = originalNAG // Restore
		})

	})

	t.Run("submitCertificate", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		acc.NAGURL = defaultNagURLForTest // Point to mock server
		acc.Open(mockAddress)
		acc.Nonce = 1 // Pre-set nonce for the test

		certData := "my certificate data"
		endpointPath := defaultNagURLForTest + "Circular_AddTransaction_"

		t.Run("should submit a certificate successfully", func(t *testing.T) {
			mockAPIResponse := map[string]interface{}{"Result": 200.0, "TxID": "newTxID789", "Message": "Transaction Added"}
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				var reqBody map[string]string
				if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
					http.Error(w, "bad request", http.StatusBadRequest)
					return
				}
				// Basic verification from JS test
				if reqBody["From"] != circular.HexFix(mockAddress) {
					t.Errorf("From mismatch: got %s", reqBody["From"])
				}
				if reqBody["Nonce"] != "1" {
					t.Errorf("Nonce mismatch: got %s", reqBody["Nonce"])
				}
				if reqBody["Type"] != "C_TYPE_CERTIFICATE" {
					t.Errorf("Type mismatch: got %s", reqBody["Type"])
				}
				// Verify payload content
				payloadHex := circular.HexFix(reqBody["Payload"])
				payloadBytes, _ := hex.DecodeString(payloadHex)
				var payloadObject struct { Action string; Data string }
				json.Unmarshal(payloadBytes, &payloadObject)

				if payloadObject.Action != "CP_CERTIFICATE" {
					t.Errorf("Payload Action mismatch: got %s", payloadObject.Action)
				}
				expectedPDataHex := hex.EncodeToString([]byte(certData))
				if circular.HexFix(payloadObject.Data) != expectedPDataHex {
					t.Errorf("Payload Data mismatch: got %s, want %s", payloadObject.Data, expectedPDataHex)
				}
				json.NewEncoder(w).Encode(mockAPIResponse)
			}

			result, err := acc.SubmitCertificate(certData, mockPrivateKey)
			if err != nil {
				// In the Go port, SubmitCertificate can return an error for network/marshalling issues
				// or a (map, nil) where the map contains success:false for API-level handled errors.
				// The JS test expected an object `{ success: false, ...}`.
				// The Go port returns such a map if the API call itself succeeded but the server reported an issue.
				// Let's assume err != nil means a lower-level problem.
				t.Fatalf("SubmitCertificate failed unexpectedly: %v", err)
			}
			// Check map content
			if success, ok := result["success"].(bool); ok && !success {
				t.Fatalf("SubmitCertificate returned success:false map: %+v", result)
			}
			// For direct API success like the JS test:
			if !reflect.DeepEqual(result, mockAPIResponse) {
				t.Errorf("SubmitCertificate response mismatch.\nExpected: %+v\nGot:      %+v", mockAPIResponse, result)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should return error map on network failure", func(t *testing.T) {
			// Forcing a network error for submitCertificate
			// The Go lib returns a map with error details for this specific case.
			originalNAG := acc.NAGURL
			acc.NAGURL = "http://localhost:12345/nonexistent" // Point to failure
			
			result, err := acc.SubmitCertificate(certData, mockPrivateKey)
			if err != nil {
				// This implies a problem *before* the custom error map could be formed
				// e.g. marshalling the request itself failed.
				t.Fatalf("SubmitCertificate returned unexpected direct error: %v", err)
			}
			if success, ok := result["success"].(bool); !ok || success {
				t.Errorf("Expected result.success to be false, got %v (ok: %v)", result["success"], ok)
			}
			if msg, ok := result["message"].(string); !ok || msg != "Server unreachable or request failed" {
				t.Errorf("Expected message 'Server unreachable or request failed', got '%v'", result["message"])
			}
			if errMsg, ok := result["error"].(string); !ok || !strings.Contains(errMsg, "connection refused") { // Varies by OS
				t.Errorf("Expected error string to contain 'connection refused' or similar, got '%v'", result["error"])
			}
			acc.NAGURL = originalNAG // Restore
		})

		t.Run("should return error map on HTTP error status", func(t *testing.T) {
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, `{"message": "Internal Server Error"}`, http.StatusInternalServerError)
			}
			result, err := acc.SubmitCertificate(certData, mockPrivateKey)
			if err != nil {
				t.Fatalf("SubmitCertificate returned unexpected direct error: %v", err)
			}
			if success, ok := result["success"].(bool); !ok || success {
				t.Errorf("Expected result.success to be false, got %v (ok: %v)", result["success"], ok)
			}
			if msg, ok := result["message"].(string); !ok || !strings.Contains(msg, "Network response was not ok") {
				t.Errorf("Expected message to contain 'Network response was not ok', got '%v'", result["message"])
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should return error if account is not open", func(t *testing.T) {
			freshAcc := circular.NewCEPAccount() // Closed by default
			freshAcc.NAGURL = defaultNagURLForTest
			_, err := freshAcc.SubmitCertificate(certData, mockPrivateKey)
			if err == nil || !strings.Contains(err.Error(), "account is not open") {
				t.Errorf("Expected 'account is not open' error, got %v", err)
			}
		})
	})

	t.Run("getTransactionOutcome", func(t *testing.T) {
		acc := circular.NewCEPAccount()
		acc.NAGURL = defaultNagURLForTest
		acc.IntervalSec = 1 // Speed up polling for tests
		
		txID := "pollTxID456"
		shortTimeoutSec := 3
		endpointPath := defaultNagURLForTest + "Circular_GetTransactionbyID_"

		// Capture logs
		var logBuf bytes.Buffer
		originalLoggerOutput := log.Writer()
		log.SetOutput(&logBuf)
		defer func() {
			log.SetOutput(originalLoggerOutput)
			// t.Log("Captured logs for GetTransactionOutcome:\n", logBuf.String()) // Optionally print logs
		}()


		t.Run("should resolve if found and confirmed quickly", func(t *testing.T) {
			logBuf.Reset()
			confirmedPayload := map[string]interface{}{"id": txID, "Status": "Confirmed", "data": "some data"}
			mockResponse := map[string]interface{}{"Result": 200.0, "Response": confirmedPayload}
			
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(mockResponse)
			}

			outcome, err := acc.GetTransactionOutcome(txID, shortTimeoutSec)
			if err != nil {
				t.Fatalf("GetTransactionOutcome failed: %v", err)
			}
			if !reflect.DeepEqual(outcome, confirmedPayload) {
				t.Errorf("Outcome mismatch.\nExpected: %+v\nGot:      %+v", confirmedPayload, outcome)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should poll and resolve when confirmed after pending", func(t *testing.T) {
			logBuf.Reset()
			pendingResponse := map[string]interface{}{"Result": 200.0, "Response": map[string]interface{}{"id": txID, "Status": "Pending"}}
			confirmedPayload := map[string]interface{}{"id": txID, "Status": "Confirmed", "finalData": "final"}
			confirmedResponse := map[string]interface{}{"Result": 200.0, "Response": confirmedPayload}
			
			callCount := 0
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount <= 2 {
					json.NewEncoder(w).Encode(pendingResponse)
				} else {
					json.NewEncoder(w).Encode(confirmedResponse)
				}
			}

			// Test will run for acc.IntervalSec * (poll attempts) + some buffer
			// Here, 1s interval, 2 pending + 1 confirmed = 3s. Timeout of 3+2=5s.
			outcome, err := acc.GetTransactionOutcome(txID, shortTimeoutSec+2)
			if err != nil {
				t.Fatalf("GetTransactionOutcome failed: %v", err)
			}
			if !reflect.DeepEqual(outcome, confirmedPayload) {
				t.Errorf("Outcome mismatch.\nExpected: %+v\nGot:      %+v", confirmedPayload, outcome)
			}
			if callCount != 3 {
				t.Errorf("Expected 3 poll attempts, got %d", callCount)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should poll and resolve when confirmed after Transaction Not Found", func(t *testing.T) {
			logBuf.Reset()
			notFoundResponse := map[string]interface{}{"Result": 200.0, "Response": "Transaction Not Found"}
			confirmedPayload := map[string]interface{}{"id": txID, "Status": "Confirmed", "finalData": "final"}
			confirmedResponse := map[string]interface{}{"Result": 200.0, "Response": confirmedPayload}
			
			callCount := 0
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount == 1 {
					json.NewEncoder(w).Encode(notFoundResponse)
				} else {
					json.NewEncoder(w).Encode(confirmedResponse)
				}
			}
			outcome, err := acc.GetTransactionOutcome(txID, shortTimeoutSec)
			if err != nil {
				t.Fatalf("GetTransactionOutcome failed: %v", err)
			}
			if !reflect.DeepEqual(outcome, confirmedPayload) {
				t.Errorf("Outcome mismatch.\nExpected: %+v\nGot:      %+v", confirmedPayload, outcome)
			}
			if callCount != 2 {
				t.Errorf("Expected 2 poll attempts, got %d", callCount)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should reject if getTransactionbyID call fails during polling", func(t *testing.T) {
			logBuf.Reset()
			// Make the mock server return an error
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Simulated network error", http.StatusInternalServerError)
			}
			_, err := acc.GetTransactionOutcome(txID, shortTimeoutSec)
			if err == nil {
				t.Fatalf("Expected GetTransactionOutcome to fail, but it succeeded.")
			}
			// The error will be from the underlying HTTP call within the loop.
			// It might be wrapped by the "timeout exceeded" if the internal call fails repeatedly.
			// The JS test expects 'Network connection lost'. The Go client returns a different error.
			// Let's check for a part of "http error" or the timeout message.
			if !strings.Contains(err.Error(), "http error! status: 500") && !strings.Contains(err.Error(), "timeout exceeded") {
				t.Errorf("Expected error to contain 'http error! status: 500' or 'timeout exceeded', got: %v", err)
			}
			delete(mockResponses, endpointPath)
		})

		t.Run("should reject with Timeout exceeded if polling duration exceeds timeoutSec", func(t *testing.T) {
			logBuf.Reset()
			pendingResponse := map[string]interface{}{"Result": 200.0, "Response": map[string]interface{}{"id": txID, "Status": "Pending"}}
			mockResponses[endpointPath] = func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(pendingResponse) // Always pending
			}
			_, err := acc.GetTransactionOutcome(txID, 1) // Very short timeout
			if err == nil {
				t.Fatalf("Expected GetTransactionOutcome to timeout, but it succeeded.")
			}
			if !strings.Contains(err.Error(), "timeout exceeded") {
				t.Errorf("Expected error message 'timeout exceeded', got: %v", err.Error())
			}
			delete(mockResponses, endpointPath)
		})
	})


	// --- Live Network Tests ---
	// These require `go test -v -tags=live -targetNetwork=testnet` (or devnet, mainnet)
	// The -tags=live ensures these tests only run when explicitly requested.
	// The -targetNetwork is a custom flag we'd need to parse.
	// Go's testing package doesn't have built-in support for custom flags like mocha's process.env.
	// A common way is to use an environment variable directly (os.Getenv)
	// or use `flag` package if running test binary directly with flags.
	// For `go test`, environment variables are simpler.

	targetNetworkEnv := os.Getenv("CIRCULAR_TEST_NETWORK")
	if targetNetworkEnv != "" {
		if !isCI() { // Optionally skip live tests in CI if not configured
			t.Run(fmt.Sprintf("CEP_Account Live Network Tests (against %s)", targetNetworkEnv), func(t *testing.T) {
				t.Parallel() // Allow live tests to run in parallel if desired
				
				// Override default test timeout for live tests if needed (e.g., `go test -timeout 2m`)
				// Individual subtests can also manage their own timeouts with context.

				liveAccount := circular.NewCEPAccount()
				
				// For live tests, we don't use the httptest mock server.
				// We allow real HTTP requests.
				// Reset httpClient in account for live tests if it was previously mocked or configured
				// For now, assuming NewCEPAccount() gives a fresh client.

				// --- TEMPORARY HARDCODED NAG URLS (from JS test) ---
				// This logic for hardcoding needs to be considered carefully in Go.
				// It implies the library itself might need a way to accept overrides,
				// or these tests need to manipulate the library's state if possible.
				hardcodedNAGURLs := map[string]string{
					"testnet": "https://testnet-nag.circularlabs.io/API/", // Example
					"devnet":  "https://devnet-nag.circularlabs.io/API/",  // Example
				}
				
				log.Printf("[LIVE TEST/%s] Configuring account...", targetNetworkEnv)
				err := liveAccount.SetNetwork(targetNetworkEnv)
				if err != nil {
					// Log and potentially skip if SetNetwork is critical and fails
					log.Printf("[LIVE TEST/%s] WARNING: liveAccount.SetNetwork('%s') failed: %v. Tests might use default NAG.", targetNetworkEnv, targetNetworkEnv, err)
				}
				initialNagFromSetNetwork := liveAccount.NAGURL

				if hardcodedURL, ok := hardcodedNAGURLs[targetNetworkEnv]; ok && hardcodedURL != "" {
					log.Printf("[LIVE TEST/%s] TEMPORARY OVERRIDE: Using hardcoded NAG_URL: %s (was %s from SetNetwork)", targetNetworkEnv, hardcodedURL, initialNagFromSetNetwork)
					liveAccount.NAGURL = hardcodedURL
				} else {
					log.Printf("[LIVE TEST/%s] No hardcoded NAG_URL override. Using URL from setNetwork(): %s", targetNetworkEnv, initialNagFromSetNetwork)
				}

				err = liveAccount.Open(testAccountAddr) // Use the globally generated test account address
				if err != nil {
					t.Fatalf("[LIVE TEST/%s] Failed to open account: %v", targetNetworkEnv, err)
				}
				log.Printf("[LIVE TEST/%s] Account NAG_URL for tests: %s", targetNetworkEnv, liveAccount.NAGURL)
				log.Printf("[LIVE TEST/%s] Account Address for tests: %s", targetNetworkEnv, liveAccount.Address)
				log.Printf("[LIVE TEST/%s] Account PrivateKey for tests (first 8 chars): %s...", targetNetworkEnv, testPrivateKeyHex[:8])


				t.Run("should update account nonce on real network", func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
					defer cancel()
					// Make UpdateAccount accept context if possible, or rely on client timeout
					success, err := liveAccount.UpdateAccount() // Add WithContext if lib supports
					if err != nil {
						t.Fatalf("UpdateAccount returned error: %v. NAG_URL: %s", err, liveAccount.NAGURL)
					}
					if !success {
						t.Fatalf("updateAccount failed for '%s'. NAG_URL was: %s. Check if this is the correct NAG for address %s.", targetNetworkEnv, liveAccount.NAGURL, liveAccount.Address)
					}
					if liveAccount.Nonce <= 0 {
						t.Errorf("Expected Nonce to be > 0, got %d", liveAccount.Nonce)
					}
					log.Printf("[LIVE TEST/%s] Updated nonce: %d", targetNetworkEnv, liveAccount.Nonce)
				})

				t.Run("should submit a certificate and get its outcome on real network", func(t *testing.T) {
					ctxOverall, cancelOverall := context.WithTimeout(context.Background(), 120*time.Second)
					defer cancelOverall()

					// 1. Update account to get current nonce
					log.Printf("[LIVE TEST/%s] Submitting: Updating account for nonce...", targetNetworkEnv)
					success, err := liveAccount.UpdateAccount()
					if !success || err != nil {
						t.Logf("[LIVE TEST/%s] Submitting: Could not update account nonce before submitting certificate. Current Nonce: %d. Error: %v. This might cause submission failure.", targetNetworkEnv, liveAccount.Nonce, err)
						// Decide whether to fail test here or proceed. Proceeding for now.
						if err != nil {t.Fatalf("UpdateAccount failed: %v", err)}
					}
					log.Printf("[LIVE TEST/%s] Submitting: Nonce for submission: %d", targetNetworkEnv, liveAccount.Nonce)

					certData := fmt.Sprintf("Test data for %s - %s", targetNetworkEnv, time.Now().UTC().Format(time.RFC3339Nano))
					
					log.Printf("[LIVE TEST/%s] Submitting: Attempting to submit certificate...", targetNetworkEnv)
					submitResult, err := liveAccount.SubmitCertificate(certData, testPrivateKeyHex) // Use global test private key
					if err != nil { // This is for pre-API call errors in Go lib
						t.Fatalf("[LIVE TEST/%s] SubmitCertificate pre-API call error: %v. NAG_URL: %s", targetNetworkEnv, err, liveAccount.NAGURL)
					}

					// Check API result from the map
					apiResultCode, _ := submitResult["Result"].(float64) // JSON numbers are float64
					txID, _ := submitResult["TxID"].(string)
					apiMessage, _ := submitResult["Message"].(string)
					apiResponseStr, _ := submitResult["Response"].(string) // If "Response" is a string message

					if int(apiResultCode) != 200 {
						t.Fatalf("[LIVE TEST/%s] Submission failed. API Result: %f, Message: '%s', Response: '%s', TxID: '%s'. NAG_URL: %s.",
							targetNetworkEnv, apiResultCode, apiMessage, apiResponseStr, txID, liveAccount.NAGURL)
					}
					if txID == "" {
						t.Fatalf("[LIVE TEST/%s] Submission succeeded (Result %f) but TxID is empty. Message: '%s'. NAG_URL: %s.",
							targetNetworkEnv, apiResultCode, apiMessage, liveAccount.NAGURL)
					}
					log.Printf("[LIVE TEST/%s] Submitting: Certificate submitted. TxID: %s. Waiting for outcome...", targetNetworkEnv, txID)

					outcomeTimeoutSec := 60
					outcome, err := liveAccount.GetTransactionOutcome(txID, outcomeTimeoutSec)
					if err != nil {
						t.Fatalf("[LIVE TEST/%s] GetTransactionOutcome for TxID %s failed: %v", targetNetworkEnv, txID, err)
					}
					if outcome == nil {
						t.Fatalf("[LIVE TEST/%s] GetTransactionOutcome for TxID %s returned nil outcome", targetNetworkEnv, txID)
					}
					
					status, _ := outcome["Status"].(string)
					if status != "Confirmed" {
						t.Errorf("[LIVE TEST/%s] Expected outcome Status 'Confirmed', got '%s'. Full outcome: %+v", targetNetworkEnv, status, outcome)
					}
					
					outcomeID, _ := outcome["id"].(string)
					// Clean TxIDs for comparison (remove 0x if present)
					cleanedOriginalTxID := strings.TrimPrefix(strings.ToLower(txID), "0x")
					cleanedOutcomeTxID := strings.TrimPrefix(strings.ToLower(outcomeID), "0x")

					if cleanedOutcomeTxID != cleanedOriginalTxID {
						t.Errorf("[LIVE TEST/%s] Outcome TxID mismatch. Expected original (cleaned) '%s', got outcome (cleaned) '%s'", targetNetworkEnv, cleanedOriginalTxID, cleanedOutcomeTxID)
					}
					log.Printf("[LIVE TEST/%s] Submitting: Transaction outcome confirmed for TxID: %s", targetNetworkEnv, txID)
				})
				
				// Add live tests for getTransactionbyID and getTransaction here, similar structure.
				// They will require placeholder TxIDs to be replaced with actual known TxIDs on the target network.
				// Example structure:
				/*
				t.Run("should fetch a transaction by ID on real network", func(t *testing.T) {
					knownTxID := "0xYOUR_KNOWN_TX_ID_ON_THIS_NAG_FOR_" + targetNetworkEnv
					if strings.HasPrefix(knownTxID, "0xYOUR_KNOWN_TX_ID") {
						t.Skipf("[LIVE TEST/%s] Skipping getTransactionbyID: Replace placeholder with a real TxID for %s", targetNetworkEnv, targetNetworkEnv)
						return
					}
					// ... actual test logic using liveAccount.GetTransactionByID ...
				})
				*/
				t.Run("should correctly reflect network URL configuration status", func(t *testing.T) {
					log.Printf("[LIVE TEST/%s] Verifying NAG_URL configuration. Current effective test NAG_URL: %s", targetNetworkEnv, liveAccount.NAGURL)
					
					freshAccForSetNetwork := circular.NewCEPAccount()
					var urlFromSetNetworkService string
					err := freshAccForSetNetwork.SetNetwork(targetNetworkEnv)
					if err != nil {
						log.Printf("[LIVE TEST/%s] SetNetwork for fresh account failed: %v. Cannot compare to service URL.", targetNetworkEnv, err)
						urlFromSetNetworkService = "ERROR_FETCHING_FROM_SERVICE"
					} else {
						urlFromSetNetworkService = freshAccForSetNetwork.NAGURL
					}

					hardcodedURLForTarget, hasHardcoded := hardcodedNAGURLs[targetNetworkEnv]

					if hasHardcoded && hardcodedURLForTarget != "" {
						if liveAccount.NAGURL != hardcodedURLForTarget {
							t.Errorf("[LIVE TEST/%s] Test is using NAG_URL %s, but hardcoded is %s. Mismatch in test setup.", targetNetworkEnv, liveAccount.NAGURL, hardcodedURLForTarget)
						}
						log.Printf("[LIVE TEST/%s] Test is using a hardcoded NAG_URL: %s", targetNetworkEnv, hardcodedURLForTarget)
						if hardcodedURLForTarget != urlFromSetNetworkService && urlFromSetNetworkService != "ERROR_FETCHING_FROM_SERVICE" {
							log.Printf("[LIVE TEST/%s] MISMATCH: Hardcoded URL (%s) differs from what getNAG service provides (%s). Service may need update.", targetNetworkEnv, hardcodedURLForTarget, urlFromSetNetworkService)
						}
						// ... other checks from JS ...
					} else {
						log.Printf("[LIVE TEST/%s] No hardcoded NAG_URL. Test is using URL from setNetwork() via service: %s", targetNetworkEnv, liveAccount.NAGURL)
						if liveAccount.NAGURL != urlFromSetNetworkService && urlFromSetNetworkService != "ERROR_FETCHING_FROM_SERVICE" {
							t.Errorf("[LIVE TEST/%s] Mismatch: liveAccount.NAG_URL (%s) not same as fresh SetNetwork call (%s). This shouldn't happen.", targetNetworkEnv, liveAccount.NAGURL, urlFromSetNetworkService)
						}
						if liveAccount.NAGURL == circular.DefaultNAG() && targetNetworkEnv != "mainnet" { // Assuming mainnet is default
							log.Printf("[LIVE TEST/%s] CURRENT CONFIG: For non-mainnet '%s', setNetwork() resulted in DEFAULT_NAG. Service might point to default.", targetNetworkEnv, targetNetworkEnv)
						}
						// ... other checks from JS ...
					}
					log.Printf("[LIVE TEST/%s] Final effective NAG_URL for this test run was: %s", targetNetworkEnv, liveAccount.NAGURL)
				})


			})
		} else {
			t.Logf("Skipping live network tests for targetNetwork '%s' because CIRCULAR_TEST_NETWORK env var is set but running in CI or CI check failed.", targetNetworkEnv)
		}
	} else {
		t.Log("Skipping live network tests: CIRCULAR_TEST_NETWORK environment variable not set.")
	}

}

// isCI checks for common CI environment variables.
func isCI() bool {
	return os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" || os.Getenv("TRAVIS") != "" || os.Getenv("CIRCLECI") != "" || os.Getenv("JENKINS_URL") != ""
}

// Helper functions to allow overriding consts for testing SetNetwork
// These would need to be implemented in the main circular package if you want this pattern
// For example, in circular.go:
/*
var (
	networkURLBase = "https://circularlabs.io/network/getNAG?network=" // actual default
	defaultNAGConst = "https://nag.circularlabs.io/NAG.php?cep="
	defaultChainConst = "0x..."
)
func GetNetworkURLBase() string { return networkURLBase }
func SetNetworkURLBase(url string) { networkURLBase = url } // BE CAREFUL with global state changes in tests
func DefaultNAG() string { return defaultNAGConst }
func DefaultChain() string { return defaultChainConst }
*/
// For this test, I'm assuming such functions exist or will be added to the circular package.
// If not, testing SetNetwork's dependency on a hardcoded const is harder without build tags or other tricks.
// Let's simulate them here for test compilation.
// This should be in the actual `circular` package for real use.
// ---- Start of simulated setters/getters for library constants ----
var (
	simulatedNetworkURL string
	networkURLOnce      sync.Once
)

func getSimulatedNetworkURLBase() string {
	networkURLOnce.Do(func() {
		// This would be the actual constant from the circular package
		simulatedNetworkURL = "https://circularlabs.io/network/getNAG?network="
	})
	return simulatedNetworkURL
}

func setSimulatedNetworkURLBase(newURL string) {
	// This is a HACK for testing. The library needs to expose this.
	// If `circular.networkURL` (the const) is truly const, it can't be changed.
	// This implies the library's SetNetwork might need to take the base URL as a parameter
	// for easier testing, or use an unexported var that can be changed via build tags.
	// For now, this test will assume a hypothetical way to change it.
	// In a real scenario, you'd mock the HTTP client used by SetNetwork or ensure
	// the library's networkURL variable is modifiable for tests (not ideal for a const).
	
	// Let's assume the library `circular.go` was modified to use a variable:
	// var NetworkURL = "..." // instead of const
	// Then in test: circular.NetworkURL = newURL (if exported)
	// Or the library's SetNetwork uses a configurable http.Client, and we configure that.
	
	// For the purpose of this test file standalone compilation, we use a local sim var.
	// This specific mock won't affect the actual library code when running tests unless
	// the library code is modified to read from such a mutable source.
	// The `circular.SetNetworkURLBase` in the test code above is calling this.
	simulatedNetworkURL = newURL
}

// We'll need to make sure the `circular` package in `circular.go` provides these for the test to be fully correct:
// circular.GetNetworkURLBase()
// circular.SetNetworkURLBase()
// circular.DefaultNAG()
// circular.DefaultChain()

// Example of how they might look in `circular.go` (simplified):
/*
package circular

var (
	// Use vars instead of consts if they need to be test-modifiable
	// Or provide functions that can be swapped out for testing.
	networkURL     = "https://circularlabs.io/network/getNAG?network="
	defaultNAG     = "https://nag.circularlabs.io/NAG.php?cep="
	defaultChain   = "0x8a20baa40c45dc5055aeb26197c203e576ef389d9acb171bd62da11dc5ad72b2"
)
// GetNetworkURLBase might be better if SetNetwork uses a client whose transport can be mocked
func GetNetworkURLBase() string { return networkURL }
// SetNetworkURLBase is generally not good practice for production code to modify global consts/vars
// Prefer dependency injection or client mocking.
func SetNetworkURLBase(url string) { networkURL = url }
func DefaultNAG() string { return defaultNAG }
func DefaultChain() string { return defaultChain }
*/

// This part is to make the circular_test.go compile by providing placeholders
// for the functions it expects from the `circular` package for constant access/modification.
// In reality, these should be implemented in `circular.go`.
// This demonstrates a common challenge when testing unexported or hardcoded dependencies.
var circularSetNetworkURLBaseFunc func(string)
var circularGetNetworkURLBaseFunc func() string
var circularDefaultNAGFunc func() string
var circularDefaultChainFunc func() string

func TestMain(m *testing.M) {
	// Simulate linking the test helpers to the actual (or placeholder) functions
	// In a real scenario, these functions would be part of the `circular` package.
	circularSetNetworkURLBaseFunc = setSimulatedNetworkURLBase // If lib had this
	circularGetNetworkURLBaseFunc = getSimulatedNetworkURLBase // If lib had this
	circularDefaultNAGFunc = func() string { return testDefaultNAG } // Simulate lib func
	circularDefaultChainFunc = func() string { return testDefaultChain } // Simulate lib func

	// Make sure the circular package uses these test-controlled functions
	// This might involve editing circular.go to use these functions IF this pattern is chosen.
	// Example: in circular.go, SetNetwork uses circular.GetNetworkURLBase()
	// For the test to compile, the test package needs to believe these exist in 'circular'
	// This is why I defined placeholders earlier for `circular.GetNetworkURLBase` etc.
	// This setup is complex because of the desire to modify what might be a const in the lib.
	// The better way is for SetNetwork to use an HTTP client that can be replaced in tests.
	
	// If the library `circular` itself provides these setters/getters, this TestMain is simpler:
	// e.g. circular.InternalSetNetworkURLForTesting = func(...)
	
	exitCode := m.Run()
	os.Exit(exitCode)
}
