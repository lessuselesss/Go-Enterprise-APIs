package ccertificate

import (
	"encoding/json"
	"circular-api/lib/utils"
)

// CCertificate is a data structure used to build and format certificate payloads.
type CCertificate struct {
	data          string
	PreviousTxID  string
	PreviousBlock string
	codeVersion   string
}

// NewCCertificate creates and initializes a new CCertificate object.
func NewCCertificate() *CCertificate {
	return &CCertificate{
		codeVersion: utils.LIB_VERSION,
	}
}

// SetData converts input data to hex format and stores it.
func (c *CCertificate) SetData(input string) {
	c.data = utils.StringToHex(input)
}

// GetData converts hex data back to its original string format.
func (c *CCertificate) GetData() string {
	return utils.HexToString(c.data)
}

// GetJSONCertificate creates a JSON payload with all certificate data.
func (c *CCertificate) GetJSONCertificate() (string, error) {
	certMap := map[string]interface{}{
		"data":          c.data,
		"previousTxID":  c.PreviousTxID,
		"previousBlock": c.PreviousBlock,
		"version":       c.codeVersion,
	}

	jsonBytes, err := json.Marshal(certMap)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetCertificateSize returns the size of the JSON certificate payload in bytes.
func (c *CCertificate) GetCertificateSize() (int, error) {
	jsonString, err := c.GetJSONCertificate()
	if err != nil {
		return 0, err
	}
	return len(jsonString), nil
}
