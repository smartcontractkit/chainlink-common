package prometheus

import (
	"strings"
	"unicode"
)

var dropSanitizationGateEnabled = true

// Normalizes the specified label to follow Prometheus label names standard
//
// See rules at https://prometheus.io/docs/concepts/data_model/#metric-names-and-labels
//
// Labels that start with non-letter rune will be prefixed with "key_"
//
// Exception is made for double-underscores which are allowed
func NormalizeLabel(label string) string {
	// Trivial case
	if len(label) == 0 {
		return label
	}

	// Replace all non-alphanumeric runes with underscores
	label = strings.Map(sanitizeRune, label)

	// If label starts with a number, prepend with "key_"
	if unicode.IsDigit(rune(label[0])) {
		label = "key_" + label
	} else if strings.HasPrefix(label, "_") && !strings.HasPrefix(label, "__") && !dropSanitizationGateEnabled {
		label = "key" + label
	}

	return label
}

// Return '_' for anything non-alphanumeric
func sanitizeRune(r rune) rune {
	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return r
	}
	return '_'
}
