package internal

import (
	"regexp"
	"strings"

	"circular-api/internal/utils"
)

// Network-related constants that don't belong in utils
const DEFAULT_NAG = "https://nag.circular.io/"
var NETWORK_URL = "https://network.circular.io/"
const DEFAULT_CHAIN = "Circular"

// Re-export commonly used constants and functions for backward compatibility
const LIB_VERSION = utils.LIB_VERSION

// Re-export utility functions for backward compatibility
func PadNumber(num int) string {
	return utils.PadNumber(num)
}

func GetFormattedTimestamp() string {
	return utils.GetFormattedTimestamp()
}

func HexFix(word string) string {
	return utils.HexFix(word)
}

func StringToHex(str string) string {
	return utils.StringToHex(str)
}

func HexToString(hexStr string) string {
	return utils.HexToString(hexStr)
}

// IsValidAddress validates if the given address has the correct format
func IsValidAddress(address string) bool {
	// Check if address starts with 0x
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	
	// Remove 0x prefix
	hexPart := strings.TrimPrefix(address, "0x")
	
	// Check if it's exactly 40 characters (20 bytes in hex)
	if len(hexPart) != 40 {
		return false
	}
	
	// Check if all characters are valid hex characters
	matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart)
	if !matched {
		return false
	}
	
	return true
}

// IsValidPrivateKey validates if the given private key has the correct format
func IsValidPrivateKey(privateKey string) bool {
	// Check if empty
	if privateKey == "" {
		return false
	}
	
	// Remove 0x prefix if present
	hexPart := strings.TrimPrefix(privateKey, "0x")
	
	// Check if it's exactly 64 characters (32 bytes in hex)
	if len(hexPart) != 64 {
		return false
	}
	
	// Check if all characters are valid hex characters
	matched, _ := regexp.MatchString("^[0-9a-fA-F]+$", hexPart)
	if !matched {
		return false
	}
	
	// Check for invalid private keys (all zeros, all ones, repetitive patterns, etc.)
	if hexPart == strings.Repeat("0", 64) || 
	   hexPart == strings.Repeat("f", 64) || 
	   hexPart == strings.Repeat("F", 64) ||
	   hexPart == strings.Repeat("1", 64) ||
	   hexPart == strings.Repeat("a", 64) ||
	   hexPart == strings.Repeat("A", 64) ||
	   hexPart == strings.Repeat("9", 64) {
		return false
	}
	
	return true
}

