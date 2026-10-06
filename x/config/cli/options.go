package cli

import "github.com/smartcontractkit/chainlink-common/x/config/markup"

// Options configures a [Binder]. Markup has no default, so a program only links the config languages it uses.
type Options struct {
	// Prefixes are tried in order: with "CRE" and "CL", chain.id reads CRE_CHAIN_ID, then CL_CHAIN_ID. The prefix ""
	// adds nothing, so chain.id reads CHAIN_ID. A nil or empty list is the same as {""}.
	Prefixes []string

	// Markup is required, usually x/config/markup/tomlmarkup. It decodes config files and decides config keys.
	Markup markup.Markup

	// BaseConfig is layered under every --config file, usually embedded, so defaults live in a file operators can read.
	BaseConfig []byte

	// DefaultConfigPath is read, if it exists, when no --config is given.
	// Empty means config.<extension> in the working directory.
	DefaultConfigPath string
}

// RegisterOption customizes one [Binder.Register] call. None exist yet; the parameter lets options be added without
// breaking callers.
type RegisterOption[T any] struct {
	setup func(*typedEntry[T]) (func(), error)
}
