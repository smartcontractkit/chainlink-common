package cli

import "github.com/smartcontractkit/chainlink-common/x/config/markup"

// Options configures how a command tree's structs are bound, decoded, and named. New behaviour
// goes here, not in new methods. Markup has no default so importers don't link every language.
type Options struct {
	// Prefixes are env var prefixes, tried in order: "CRE", "CL" bind chain.id to CRE_CHAIN_ID
	// then CL_CHAIN_ID.
	Prefixes []string

	// Markup is the required config language, usually x/config/markup/tomlmarkup. It decodes the
	// config files, and sets the key tag, the default extension, and which types decode whole.
	Markup markup.Markup

	// BaseConfig is a default config file, usually embedded, in Markup's format and layered under
	// every --config file. Keeps defaults in a file operators can read.
	BaseConfig []byte

	// DefaultConfigPath is read when --config names no file; empty means config.<ext> in the
	// working directory.
	DefaultConfigPath string
}

// RegisterOption configures one [Binder.Register] of a T.
type RegisterOption[T any] struct {
	setup func(*targetEntry) (func(), error)
}
