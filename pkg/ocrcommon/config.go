package ocrcommon

import (
	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// Config is the networking configuration a process needs to run a local libocr rage p2p V2
// peer. Shaped for chainlink-common's x/config/cli, so a binary can bind it as flags, env vars
// and config-file keys directly.
//
// It says nothing about delegating networking elsewhere instead of creating a peer: that is a
// property of the process, not of the peer, so a binary offering that mode embeds this struct in
// one of its own that adds it. Nor does it say where the peer's identity comes from: a process
// may unlock it from a keystore, derive it deterministically, or be handed one, so the settings
// that choice needs (a keystore password, say) belong to that process and not here. Defaults are
// the binary's too - the zero value of this struct is just a zero value, and the binary binding
// it supplies the instance to decode into.
type Config struct {
	// ListenAddresses are the rage p2p V2 listen addresses (host:port), such as 127.0.0.1:1234; setting any creates a
	// local peer.
	ListenAddresses []string
	// AnnounceAddresses are the rage p2p V2 announce addresses (host:port); defaults to the listen addresses.
	AnnounceAddresses []string `validate:"excluded_without=ListenAddresses"`
	// DeltaReconcile is the rage p2p V2 delta reconcile interval.
	DeltaReconcile config.Duration
	// DeltaDial is the rage p2p V2 minimum interval between dial attempts.
	DeltaDial config.Duration

	// IncomingBufferSize is the per-remote incoming message buffer size.
	IncomingBufferSize int
	// OutgoingBufferSize is the per-remote outgoing message buffer size.
	OutgoingBufferSize int
}
