package monitoring

import (
	"io"
	"strings"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	nodesOnlyTypeSuffix = "-nodesonly"
)

// NodesOnlyType indicates to the multiFeedMonitor that this source or exporter should not be created on a per-feed basis
func NodesOnlyType(in string) string {
	if IsNodesOnly(in) {
		return in
	}
	return in + nodesOnlyTypeSuffix
}
func IsNodesOnly(s string) bool {
	return strings.HasSuffix(s, nodesOnlyTypeSuffix)
}

// NodesParser extracts multiple nodes' configurations from the configuration server, eg. weiwatchers.com
type NodesParser func(buf io.ReadCloser) ([]NodeConfig, error)

// NodeConfig is the subset of on-chain node operator's configuration required by the OM framework.
type NodeConfig interface {
	GetName() string
	GetAccount() types.Account
}
