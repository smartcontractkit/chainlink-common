package main

import (
	"fmt"
	"regexp"
	"strings"
)

// validateModuleURI ensures the module URI follows Go module conventions
func validateModuleURI(uri string) error {
	// Check for valid Go module path format
	if !regexp.MustCompile(`^[\w./-]+(/[\w./-]+)*$`).MatchString(uri) {
		return fmt.Errorf("invalid module URI format: %s", uri)
	}
	return nil
}

// validateGitRef ensures the git reference is safe
func validateGitRef(ref string) error {
	// Git refs should only contain alphanumeric, dot, dash, slash, underscore
	if !regexp.MustCompile(`^[\w./-]+$`).MatchString(ref) {
		return fmt.Errorf("invalid git reference: %s", ref)
	}
	return nil
}

// validateInstallPath ensures the install path is safe
func validateInstallPath(path string) error {
	// Validate as Go import path
	if !regexp.MustCompile(`^[.\w/-]+(/[\w./-]+)*$`).MatchString(path) {
		return fmt.Errorf("invalid install path: %s", path)
	}
	return nil
}

// validateGoFlags ensures flags are safe and prevents command injection
func validateGoFlags(flags string) error {
	// Check for potentially dangerous characters that could enable command injection
	dangerousPatterns := []string{
		";", "&&", "||", "`", "$", "|", "<", ">", "#", "//",
		"shutdown", "reboot", "rm -", "format", "mkfs", "dd",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(flags, pattern) {
			return fmt.Errorf("potentially unsafe pattern in go flags: %s", pattern)
		}
	}

	return nil
}
