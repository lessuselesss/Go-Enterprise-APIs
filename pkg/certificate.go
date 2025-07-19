package circular_enterprise_apis

import (
	"encoding/json"

	"circular_enterprise_apis/pkg/utils"
)

// CCertificate represents a data structure for a Circular Protocol certificate.
// It encapsulates the core data content, references to previous transactions and blocks
// for chaining purposes, and the version of the library used to create it.
// This structure is fundamental for ensuring data integrity and traceability on the blockchain.
type CCertificate struct {
	Data          string `json:"data"`          // The primary data content of the certificate, typically in hexadecimal format.
	PreviousTxID  string `json:"previousTxID"`  // The transaction ID of the preceding certificate in a chain, if applicable.
	PreviousBlock string `json:"previousBlock"` // The block identifier of the preceding certificate in a chain, if applicable.
	Version       string `json:"version"`       // The version of the Circular Enterprise APIs library used to generate the certificate.
}

// NewCCertificate creates and initializes a new CCertificate instance with default empty values.
// The `Version` field is automatically populated with the current library version (`LibVersion`).
// This factory function ensures that a new certificate object is properly structured
// before its data and chaining references are set.
//
// Returns:
//
//	A pointer to a newly initialized CCertificate struct.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		Data:          "",
		PreviousTxID:  "",
		PreviousBlock: "",
		Version:       LibVersion,
	}
}

// SetData sets the primary data content of the certificate.
// The input `data` string is automatically converted into its hexadecimal representation
// and stored in the `Data` field of the CCertificate. This ensures that the certificate
// data is consistently stored in a blockchain-compatible format.
//
// Parameters:
//   - data: The string content to be set as the certificate's data.
func (c *CCertificate) SetData(data string) {
	c.Data = utils.StringToHex(data)
}

// GetData retrieves the primary data content from the certificate.
// The hexadecimal data stored in the `Data` field of the CCertificate is
// automatically converted back into its original string representation.
// This function allows for easy access to the human-readable form of the
// certificate's payload.
//
// Returns:
//
//	The original string representation of the certificate's data.
func (c *CCertificate) GetData() string {
	return utils.HexToString(c.Data)
}

// GetJSONCertificate serializes the entire CCertificate object into a JSON string.
// This function is crucial for preparing the certificate for submission to the blockchain
// or for external consumption, ensuring a standardized and interoperable format.
// It includes all fields of the CCertificate: `Data`, `PreviousTxID`, `PreviousBlock`, and `Version`.
//
// Returns:
//
//	A JSON string representation of the CCertificate. In case of a marshaling error,
//	an empty string is returned, aligning with the behavior of the corresponding Java API.
func (c *CCertificate) GetJSONCertificate() string {
	certificateMap := map[string]interface{}{
		"data":          c.Data,
		"previousTxID":  c.PreviousTxID,
		"previousBlock": c.PreviousBlock,
		"version":       c.Version,
	}
	jsonBytes, err := json.Marshal(certificateMap)
	if err != nil {
		return "" // Return empty string on error, matching Java's behavior
	}
	return string(jsonBytes)
}

// GetCertificateSize calculates the size of the JSON-serialized representation of the certificate in bytes.
// This function is useful for estimating the payload size before submission to the blockchain
// or for network transfer considerations. It first converts the certificate to its JSON string
// representation and then measures the byte length of that string.
//
// Returns:
//
//	The size of the JSON-serialized certificate in bytes. If the JSON serialization fails,
//	0 is returned, maintaining consistency with the Java API.
func (c *CCertificate) GetCertificateSize() int {
	jsonString := c.GetJSONCertificate()
	if jsonString == "" {
		return 0 // Return 0 on error, matching Java's behavior
	}
	return len(jsonString)
}

// SetPreviousTxID sets the transaction ID of the certificate that immediately precedes
// the current certificate in a chain. This is a critical component for establishing
// a verifiable lineage of data on the blockchain, enabling historical tracking and integrity checks.
//
// Parameters:
//   - txID: The transaction ID (string) of the previous certificate.
func (c *CCertificate) SetPreviousTxID(txID string) {
	c.PreviousTxID = txID
}

// SetPreviousBlock sets the block identifier of the certificate that immediately precedes
// the current certificate in a chain. Similar to `SetPreviousTxID`, this helps in
// establishing a verifiable and traceable history of certificates on the blockchain.
//
// Parameters:
//   - block: The block identifier (string) of the previous certificate.
func (c *CCertificate) SetPreviousBlock(block string) {
	c.PreviousBlock = block
}

// GetPreviousTxID retrieves the transaction ID of the certificate that immediately precedes
// the current certificate in a chain. This value is crucial for reconstructing the
// historical lineage of data and verifying the integrity of the certificate chain.
//
// Returns:
//
//	The transaction ID (string) of the previous certificate.
func (c *CCertificate) GetPreviousTxID() string {
	return c.PreviousTxID
}

// GetPreviousBlock retrieves the block identifier of the certificate that immediately precedes
// the current certificate in a chain. This value, in conjunction with `GetPreviousTxID`,
// allows for comprehensive tracing and validation of the certificate's history on the blockchain.
//
// Returns:
//
//	The block identifier (string) of the previous certificate.
func (c *CCertificate) GetPreviousBlock() string {
	return c.PreviousBlock
}
