// Package network lends a process that runs an OCR oracle what only the node it runs for has: rage networking over the
// node's peer. The node serves its BinaryNetworkEndpointFactory with [RegisterServer], and the process gets one from
// [Connect].
package network

import (
	"fmt"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"
)

// MaxMessageSize is the largest gRPC message between a node and a process: an OCR message as large as libocr allows,
// with room for its framing. gRPC's 4 MiB default would break an endpoint's stream on a large message, such as a
// leader's proposal carrying many observations; the endpoint on the node still enforces the oracle's own limits. The
// server [RegisterServer] serves on must receive messages this large.
const MaxMessageSize = ragetypes.MaxMessageLength + 1024

// Version picks how [Connect] reaches the node.
type Version int32

const (
	// VersionLatest is the latest version this package knows.
	VersionLatest Version = iota
	VersionV1
)

const latestVersion = VersionV1

// Config is how to reach the node.
type Config struct {
	Version Version
}

func resolveVersion(v Version) (Version, error) {
	if v == VersionLatest {
		return latestVersion, nil
	}

	if v < VersionLatest || v > latestVersion {
		return 0, fmt.Errorf("unknown OCR network version %d", v)
	}

	return v, nil
}
