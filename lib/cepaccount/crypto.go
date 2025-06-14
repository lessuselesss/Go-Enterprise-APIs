package cepaccount

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// sign creates a mock signature for the transaction.
// In a real implementation, this would use secp256k1.
func sign(privateKey, message string) (string, error) {
	// This is a placeholder and not a secure signature.
	hash := sha256.Sum256([]byte(message))
	return hex.EncodeToString(hash[:]), nil
}

// getPublicKey derives a mock public key from a private key.
// In a real implementation, this would use secp256k1.
func getPublicKey(privateKey string) (string, error) {
	// This is a placeholder and not a real public key derivation.
	return fmt.Sprintf("pubkey_for_%s", privateKey), nil
}
