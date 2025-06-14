package utils

import (
	"regexp"
	"testing"
)

func TestPadNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"SingleDigit", 5, "05"},
		{"DoubleDigit", 10, "10"},
		{"Zero", 0, "00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := PadNumber(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %s, but got %s", tt.expected, actual)
			}
		})
	}
}

func TestGetFormattedTimestamp(t *testing.T) {
	// Regex to validate the timestamp format YYYY:MM:DD-HH:MM:SS
	// This is more robust than checking a static timestamp.
	timestampRegex := regexp.MustCompile(`^\d{4}:(0[1-9]|1[0-2]):(0[1-9]|[12]\d|3[01])-(2[0-3]|[01]\d):[0-5]\d:[0-5]\d$`)

	timestamp := GetFormattedTimestamp()

	if !timestampRegex.MatchString(timestamp) {
		t.Errorf("timestamp format is incorrect. got: %s", timestamp)
	}
}

func TestHexFix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"WithPrefix", "0x1a2b3c", "1a2b3c"},
		{"WithoutPrefix", "1a2b3c", "1a2b3c"},
		{"EmptyString", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := HexFix(tt.input)
			if actual != tt.expected {
				t.Errorf("expected %s, but got %s", tt.expected, actual)
			}
		})
	}
}

func TestStringToHexAndHexToString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		hexpectd string
	}{
		{"SimpleString", "hello world", "68656c6c6f20776f726c64"},
		{"WithNumbers", "12345", "3132333435"},
		{"EmptyString", "", ""},
		{"SpecialChars", "!@#$%^&*()", "21402324255e262a2829"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hex := StringToHex(tt.input)
			if hex != tt.hexpectd {
				t.Errorf("StringToHex: expected %s, but got %s", tt.hexpectd, hex)
			}

			original := HexToString(hex)
			if original != tt.input {
				t.Errorf("HexToString: expected %s, but got %s", tt.input, original)
			}
		})
	}
}
