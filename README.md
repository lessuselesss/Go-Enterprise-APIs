# Circular Enterprise APIs - Go Implementation

Official Circular Protocol Enterprise APIs for Data Certification - Go Implementation

## Features

- Account management and blockchain interaction
- Certificate creation and submission
- Transaction tracking and verification
- Secure digital signatures using ECDSA (secp256k1)
- RFC 6979 compliant deterministic signatures

## Requirements

- Go 1.20 or higher

## Dependencies

- `github.com/decred/dcrd/dcrec/secp256k1/v4` for secp256k1 elliptic curve operations
- `github.com/joho/godotenv` for loading environment variables

## Installation

1. Clone the repository
2. Navigate to the project directory:
   ```bash
   cd Circular-Protocol/Enterprise-APIs/Go-CEP-APIs
   ```
3. Download dependencies:
   ```bash
   go mod tidy
   ```

## Usage Example

See `examples/simple_certificate_submission.go` for a basic example of how to use the API to submit a certificate.

## API Documentation

### CEPAccount Struct

Main struct for interacting with the Circular blockchain:

- `NewCEPAccount() *CEPAccount` - Factory function to create a new `CEPAccount` instance.
- `Open(address string) bool` - Initializes the account with a specified blockchain address.
- `Close()` - Clears all sensitive and operational data from the account.
- `SetNetwork(network string) string` - Configures the account to operate on a specific blockchain network.
- `SetBlockchain(chain string)` - Explicitly sets the blockchain identifier for the account.
- `UpdateAccount() bool` - Fetches the latest nonce for the account from the NAG.
- `SubmitCertificate(pdata string, privateKeyHex string)` - Creates, signs, and submits a data certificate to the blockchain.
- `GetTransaction(blockID string, transactionID string) map[string]interface{}` - Retrieves transaction details by block and transaction ID.
- `GetTransactionOutcome(txID string, timeoutSec int, intervalSec int) map[string]interface{}` - Polls for the final status of a transaction.
- `GetLastError() string` - Retrieves the last error message.

### CCertificate Struct

Struct for managing certificates:

- `NewCCertificate() *CCertificate` - Factory function to create a new `CCertificate` instance.
- `SetData(data string)` - Sets the primary data content of the certificate.
- `GetData() string` - Retrieves the primary data content from the certificate.
- `GetJSONCertificate() string` - Serializes the certificate object into a JSON string.
- `GetCertificateSize() int` - Calculates the size of the JSON-serialized certificate in bytes.
- `SetPreviousTxID(txID string)` - Sets the transaction ID of the preceding certificate.
- `SetPreviousBlock(block string)` - Sets the block identifier of the preceding certificate.
- `GetPreviousTxID() string` - Retrieves the transaction ID of the preceding certificate.
- `GetPreviousBlock() string` - Retrieves the block identifier of the preceding certificate.

## Testing

To run the tests, you need to set up the following environment variables in a `.env` file in the project root:

```
CIRCULAR_PRIVATE_KEY="your_64_character_private_key_here"
CIRCULAR_ADDRESS="your_wallet_address_here"
```

The private key should be a 64-character (32-byte) hex string, and the address should be a valid Ethereum-style address (40 characters + 0x prefix).

### Running Tests

```bash
go test ./...
```

## Building

```bash
go build -o circular-apis main.go
```

## License

MIT License - see LICENSE file for details

## Credits

CIRCULAR GLOBAL LEDGERS, INC. - USA

- Original JS Version: Gianluca De Novi, PhD
- Go Implementation: Danny De Novi