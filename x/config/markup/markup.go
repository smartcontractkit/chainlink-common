// Package markup describes the languages a configuration can be written in.
//
// Nothing outside a language's own package names its library, so a config in another language
// is a value passed in rather than a second code path - and a program linking one language does
// not carry the others. A language lives beside this package, as x/config/markup/tomlmarkup does.
// [Decode] reads one document without a command line; the pieces it shares with x/config/cli
// live in x/config/internal/decode.
package markup

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/x/config/internal/decode"
)

// Markup is a configuration language, described as its own decoder reads a document.
type Markup = decode.Markup

// Err is what a caller reports without a Markup: no language is chosen silently on its behalf.
var Err = errors.New("no markup language set: see x/config/markup/tomlmarkup for an example")
