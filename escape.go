package zpl

import (
	"strings"
)

// EscapeFieldData escapes special ZPL characters in field data.
// This prevents ZPL injection when using untrusted input.
//
// The following characters are escaped:
//   - ^ (caret) - ZPL command prefix
//   - ~ (tilde) - ZPL control command prefix
//
// Use this function when building field data from user input:
//
//	label.TextField(x, y, zpl.Font0, 30, 30, zpl.EscapeFieldData(userInput))
func EscapeFieldData(s string) string {
	// Replace ^ and ~ with their safe representations
	// Using underscore prefix convention (similar to hex mode)
	s = strings.ReplaceAll(s, "^", "_5E")
	s = strings.ReplaceAll(s, "~", "_7E")
	return s
}

// MustEscapeFieldData returns true if the string contains characters
// that should be escaped before use in ZPL field data.
func MustEscapeFieldData(s string) bool {
	return strings.ContainsAny(s, "^~")
}
