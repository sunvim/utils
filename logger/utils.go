package logger

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EncodeJSON encodes an interface to JSON bytes
func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// DecodeJSON decodes JSON bytes to an interface
func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// EncodeJSONString encodes an interface to JSON string
func EncodeJSONString(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

// DecodeJSONString decodes JSON string to an interface
func DecodeJSONString(str string, v interface{}) error {
	return json.Unmarshal([]byte(str), v)
}

// ParseChainID parses a chain ID into namespace and reference
func ParseChainID(chainID string) []string {
	return strings.Split(chainID, ":")
}

// FormatChainID formats namespace and reference into a chain ID
func FormatChainID(namespace, reference string) string {
	return fmt.Sprintf("%s:%s", namespace, reference)
}

// Contains checks if a slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Remove removes a string from a slice
func Remove(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// Unique removes duplicate strings from a slice
func Unique(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

// TruncateString truncates a string to a maximum length
func TruncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen-3] + "..."
}
