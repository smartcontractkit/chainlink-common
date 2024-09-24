package cmd

import (
	"strings"
)

func capitalize(s string) string {
	return strings.ToUpper(string(s[0])) + s[1:]
}
