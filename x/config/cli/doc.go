// Package cli fills Go config structs from a cobra command's flags, env vars, and config files, in that order of
// precedence, over the struct's own values.
//
// Create a [Binder] with [New] and attach structs with [Binder.Register]. They are decoded and
// validated before the command's hooks run, so RunE sees populated structs.
//
// Help text comes from the DocComments that [github.com/smartcontractkit/chainlink-common/x/config/commentparsing]
// generates.
//
// See the examples directory for runnable programs.
package cli
