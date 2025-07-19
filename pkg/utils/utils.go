package utils

import (
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// PadNumber converts a single-digit integer (0-9) into a two-digit string by prepending a "0".
// For example, 5 becomes "05". Numbers 10 and greater are converted directly to their string representation.
// This function is useful for formatting time components like months, days, hours, minutes, and seconds
// to ensure a consistent two-digit display.
//
// Parameters:
//   - num: The integer to be padded.
//
// Returns:
//
//	The padded string representation of the number.
func PadNumber(num int) string {
	if num >= 0 && num < 10 {
		return "0" + strconv.Itoa(num)
	}
	return strconv.Itoa(num)
}

// GetFormattedTimestamp generates a UTC timestamp string in the format "YYYY:MM:DD-HH:MM:SS".
// This format is specifically designed for internal use within the Circular Enterprise APIs
// to ensure consistent time representation across various operations, such as transaction
// timestamping and certificate creation.
//
// Returns:
//
//	A string representing the current UTC timestamp in "YYYY:MM:DD-HH:MM:SS" format.
func GetFormattedTimestamp() string {
	return time.Now().UTC().Format("2006:01:02-15:04:05")
}

// HexFix normalizes and sanitizes a given hexadecimal string to a consistent format.
// It performs the following operations:
//   - Handles empty or null input strings by returning an empty string.
//   - Removes common "0x" or "0X" prefixes if present.
//   - Converts all hexadecimal characters to lowercase for uniformity.
//   - Ensures the resulting hexadecimal string has an even number of characters
//     by prepending a "0" if its length is odd. This is crucial for correct
//     byte-level decoding.
//
// Parameters:
//   - hexStr: The input string to be normalized, which may or may not be a valid hexadecimal string.
//
// Returns:
//
//	The cleaned, normalized, and lowercase hexadecimal string, ready for further processing or decoding.
func HexFix(hexStr string) string {
	if hexStr == "" {
		return ""
	}

	// Remove "0x" or "0X" prefix
	if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}

	// Convert to lower
	hexStr = strings.ToLower(hexStr)

	// Pad with '0' if length is odd
	if len(hexStr)%2 != 0 {
		hexStr = "0" + hexStr
	}

	return hexStr
}

// StringToHex converts a standard UTF-8 string into its hexadecimal representation.
// Each character in the input string is first converted to its UTF-8 byte sequence,
// and then each byte is encoded as two hexadecimal characters (0-F).
// The resulting hexadecimal string is always in uppercase.
// This function is essential for preparing string data for cryptographic operations
// or for storage in systems that require hexadecimal encoding.
//
// Parameters:
//   - s: The input string to be converted to hexadecimal.
//
// Returns:
//
//	The hexadecimal representation of the input string in uppercase.
//
// Example:
//
//	"Hello" -> "48656C6C6F"
func StringToHex(s string) string {
	// Use standard library to encode bytes to hex string
	hexEncoded := make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(hexEncoded, []byte(s))

	// Convert to uppercase
	return strings.ToUpper(string(hexEncoded))
}

// HexToString converts a hexadecimal string back into its original UTF-8 string representation.
// The function is robust, handling various input formats:
//   - It is case-insensitive (e.g., "48" and "4h" are treated the same).
//   - It gracefully handles optional "0x" or "0X" prefixes, removing them before decoding.
//   - If the input hexadecimal string contains invalid characters or has an odd length
//     (which would result in an incomplete byte), the function will return an empty string.
//
// This behavior aligns with the error handling in the corresponding Java implementation,
// ensuring consistency across different API versions.
//
// Parameters:
//   - hexStr: The hexadecimal string to be converted back to a regular string.
//
// Returns:
//
//	The original string representation of the hexadecimal input. Returns an empty string
//	if the input is invalid or cannot be decoded.
//
// Example:
//
//	"48656C6C6F" -> "Hello"
//	"0x48656c6c6f" -> "Hello"
func HexToString(hexStr string) string {
	// Handle empty string after stripping prefix
	if hexStr == "" {
		return ""
	}

	// Remove "0x" or "0X" prefix if present
	if strings.HasPrefix(hexStr, "0x") || strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}

	// Convert to lowercase to ensure consistent input for hex.DecodeString
	// (though hex.DecodeString is case-insensitive, this can prevent subtle issues)
	hexStr = strings.ToLower(hexStr)

	// Decode the hex string to bytes
	decodedBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return "" // Return empty string on error, matching Java's behavior
	}

	return string(decodedBytes)
}
