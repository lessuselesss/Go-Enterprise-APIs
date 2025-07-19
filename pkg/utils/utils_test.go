package utils

import (
	"testing"
)

func TestPadNumber(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{5, "05"},
		{10, "10"},
		{0, "00"},
		{123, "123"},
		{-1, "-1"},
	}

	for _, test := range tests {
		actual := PadNumber(test.num)
		if actual != test.expected {
			t.Errorf("PadNumber(%d): Expected %s, Got %s", test.num, test.expected, actual)
		}
	}
}

/*
func TestGetFormattedTimestamp(t *testing.T) {
	expectedFormat := "2006:01:02-15:04:05"
	// Get the current time and format it as expected
	now := time.Now()
	expected := now.Format(expectedFormat)

	actual := GetFormattedTimestamp()

	// We can't directly compare to a fixed string because time.Now() changes.
	// Instead, we check if the format is correct and if it's close to the current time.
	// A simple check for length and a basic format match is sufficient for this test.
	if len(actual) != len(expected) {
		t.Errorf("GetFormattedTimestamp(): Expected length %d, Got length %d. Actual: %s", len(expected), len(actual), actual)
	}

	// Further check: try parsing the actual string to ensure it's a valid time in the expected format
	parsedTime, err := time.Parse(expectedFormat, actual)
	if err != nil {
		t.Errorf("GetFormattedTimestamp(): Could not parse actual timestamp %s with format %s: %v", actual, expectedFormat, err)
	}

	// Check if the parsed time is within a reasonable range of 'now'
	if now.Sub(parsedTime) > 5*time.Second || parsedTime.Sub(now) > 5*time.Second {
		t.Errorf("GetFormattedTimestamp(): Actual timestamp %s is not close to current time %s", actual, now.Format(expectedFormat))
	}
}
*/

func TestHexFix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0xabc", "0abc"},
		{"1a2b3c", "1a2b3c"},
		{"DEF", "0def"},
		{"0XfF", "ff"},
		{"", ""},
		{"0x12345", "012345"},
		{"abcDEF", "abcdef"},
		{"AbCdEf", "abcdef"},
		{"0x", ""},
		{"0X", ""},
		{"0", "00"},
		{"A", "0a"},
		// Add a nil case (Go string zero value is empty string)
		{func() string { var s string; return s }(), ""},
	}

	for _, test := range tests {
		actual := HexFix(test.input)
		if actual != test.expected {
			t.Errorf("hexFix(\"%s\"): Expected %s, Got %s",
				test.input, test.expected, actual)
		}
	}
}

func TestStringToHex(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"Hello", "48656C6C6F"},
		{"你好", "E4BDA0E5A5BD"},
		{"😊", "F09F988A"},
		{"Test\n123", "546573740A313233"},
		{"\u0000Byte", "0042797465"},
		{"abc", "616263"},
		{"ABC", "414243"},
		{func() string { var s string; return s }(), ""},
	}

	for _, test := range tests {
		actual := StringToHex(test.input)
		if actual != test.expected {
			t.Errorf("stringToHex(\"%s\"): Expected %s, Got %s",
				test.input, test.expected, actual)
		}
	}
}

func TestHexToString(t *testing.T) {
	tests := []struct {
		input       string
		expectedStr string
	}{
		{"48656c6c6f", "Hello"},
		{"e4bda0e5a5bd", "你好"},
		{"", ""},
		{"48656C6C6F", "Hello"},
		{"48656c6C6f", "Hello"},
		{"0x48656c6c6f", "Hello"},
		{"0X48656c6c6f", "Hello"},
		{"48656c6c6fg", ""}, // Invalid hex, should return empty string
		{"48656c 6c6f", ""}, // Invalid hex, should return empty string
		{"48656c6c6", ""},   // Odd length, should return empty string
		{"0x48656c6c6", ""}, // Odd length, should return empty string
		{"000000", "\x00\x00\x00"},
		{func() string { var s string; return s }(), ""},
	}

	for _, test := range tests {
		actualStr := HexToString(test.input)
		if actualStr != test.expectedStr {
			t.Errorf("HexToString(\"%s\"): Expected string %q, Got %q",
				test.input, test.expectedStr, actualStr)
		}
	}
}
