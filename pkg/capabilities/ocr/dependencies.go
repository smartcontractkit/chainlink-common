// Package ocr lends a process that runs an OCR oracle what only the node it runs for has: rage networking over the
// node's peer, and signing with the node's OCR keys. The node serves its [Node] with [RegisterServer], and the process
// gets the same libocr types from [Connect].
package ocr

import (
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

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

// Node is what a node lends a process that runs an OCR oracle.
type Node struct {
	BinaryNetworkEndpointFactory ocrtypes.BinaryNetworkEndpointFactory
	OffchainKeyring              ocrtypes.OffchainKeyring
	// OnchainKeyring signs a report with the key bundle its info names, such as evm.
	OnchainKeyring ocr3types.OnchainKeyring[string]
}
