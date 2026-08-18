package config

import (
	"github.com/smartcontractkit/libocr/commontypes"
)

// BootstrapperLocator is a bootstrap peer as it is configured: a peer ID and the
// addresses it can be reached at, written peerID@host:port[/host:port...].
//
// It exists so that a binary taking bootstrappers as configuration declares them
// as what they are rather than as strings it parses itself. Parsing is the same
// parsing either way; doing it here means a malformed entry is reported when the
// configuration is read, against the setting that holds it, instead of somewhere
// later that happens to look.
type BootstrapperLocator commontypes.BootstrapperLocator

// UnmarshalText parses the peerID@addr[/addr...] form, exactly as libocr does:
// the two must agree, and there is no second way to spell one.
func (b *BootstrapperLocator) UnmarshalText(text []byte) error {
	return (*commontypes.BootstrapperLocator)(b).UnmarshalText(text)
}

// MarshalText writes the form UnmarshalText reads, so a configuration this was
// decoded from can be written back out - which is what the generated docs and
// example configs do.
func (b BootstrapperLocator) MarshalText() ([]byte, error) {
	return (&b).toOCR().MarshalText()
}

// ToBootstrapperLocator returns the libocr form, for handing to an oracle.
func (b BootstrapperLocator) ToBootstrapperLocator() commontypes.BootstrapperLocator {
	return *(&b).toOCR()
}

func (b *BootstrapperLocator) toOCR() *commontypes.BootstrapperLocator {
	return (*commontypes.BootstrapperLocator)(b)
}

// BootstrapperLocators is a configured list of bootstrap peers.
//
// Each entry parses itself, so a list arrives as a comma-separated string from a
// flag or an env var, and as a list of strings from a config file, without this
// type having to know which it was.
type BootstrapperLocators []BootstrapperLocator

// ToBootstrapperLocators returns the libocr form of every entry, which is what
// an oracle is configured with.
func (b BootstrapperLocators) ToBootstrapperLocators() []commontypes.BootstrapperLocator {
	locators := make([]commontypes.BootstrapperLocator, 0, len(b))
	for _, locator := range b {
		locators = append(locators, locator.ToBootstrapperLocator())
	}
	return locators
}
