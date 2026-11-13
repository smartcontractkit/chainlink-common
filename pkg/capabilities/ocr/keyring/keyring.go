// Package keyring lends a process that runs an OCR oracle what only the node it runs for has: signing with the node's
// OCR keys. The node serves its [Keyrings] with [RegisterServer], and the process gets the same libocr types from
// [Connect].
package keyring

import (
	"fmt"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// MaxMessageSize is the largest gRPC message between a node and a process: a report as large as libocr allows, with
// room for its framing, which is past gRPC's 4 MiB default. The server [RegisterServer] serves on must receive messages
// this large.
const MaxMessageSize = ocr3types.MaxMaxReportLength + 1024

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

// Keyrings are the node's OCR keys a process signs with.
type Keyrings struct {
	OffchainKeyring ocrtypes.OffchainKeyring
	// OnchainKeyring signs a report with the key bundle its info names, such as evm.
	OnchainKeyring ocr3types.OnchainKeyring[string]
}

func resolveVersion(v Version) (Version, error) {
	if v == VersionLatest {
		return latestVersion, nil
	}

	if v < VersionLatest || v > latestVersion {
		return 0, fmt.Errorf("unknown OCR keyring version %d", v)
	}

	return v, nil
}
