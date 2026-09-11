// Package flags binds a Go config struct to a cobra command, so one struct definition is the
// single description of a program's configuration and every source agrees with it: a CLI flag,
// an environment variable, a config file, and the struct's own compiled-in values, resolved in
// that order of precedence.
//
// Start at [RegisterCommandFlags], and at [Options], which is where the tag conventions, env var
// prefixes, namespacing and decoding are configured. Register a struct and the decode step is
// wired into the command's PreRun for you; by the time RunE is called the struct is populated
// and its `validate` tags have been checked.
//
// The examples directory has runnable programs for the common shapes.
package flags
