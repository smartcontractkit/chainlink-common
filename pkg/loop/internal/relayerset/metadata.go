package relayerset

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"google.golang.org/grpc/metadata"
)

const metadataChain = "relay-chain-id"
const metadataNetwork = "relay-network"

func appendRelayID(ctx context.Context, id types.RelayID) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataNetwork, id.Network, metadataChain, id.ChainID)
}

func readRelayID(ctx context.Context) (types.RelayID, error) {
	network, err := readContextValue(ctx, metadataNetwork)
	if err != nil {
		return types.RelayID{}, err
	}
	chainID, err := readContextValue(ctx, metadataChain)
	if err != nil {
		return types.RelayID{}, err
	}
	return types.RelayID{
		Network: network, ChainID: chainID,
	}, nil
}
