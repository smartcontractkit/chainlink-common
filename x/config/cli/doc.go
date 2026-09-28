// Package cli binds a Go config struct to a cobra command. Sources, highest precedence first:
// CLI flag, env var, config file, the struct's own values.
//
// Scalars, durations, text unmarshalers, and lists and maps of those get flags. More structured
// fields are config file only. [New] registers a repeatable --config; files layer in order, each
// decoded by the markup itself.
//
// Start at [New], [Binder.Register] and [Options]. Decoding runs in PreRun, so RunE sees a
// populated, validated struct.
//
// Help text is each field's doc comment, via the generated DocComments of
// [github.com/smartcontractkit/chainlink-common/x/config/commentparsing]; ungenerated packages
// bind without help.
//
// See the examples directory for runnable programs.
package cli
