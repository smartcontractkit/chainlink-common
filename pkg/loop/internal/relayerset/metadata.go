package relayerset

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"google.golang.org/grpc/metadata"
)

type relayerMetadata struct {
	chain   string
	network string
}

func (rm *relayerMetadata) appendRelayID(ctx context.Context, id types.RelayID) context.Context {
	return metadata.AppendToOutgoingContext(ctx, rm.network, id.Network, rm.chain, id.ChainID)
}

func (rm *relayerMetadata) readRelayID(ctx context.Context) (types.RelayID, error) {
	network, err := readContextValue(ctx, rm.network)
	if err != nil {
		return types.RelayID{}, err
	}
	chainID, err := readContextValue(ctx, rm.chain)
	if err != nil {
		return types.RelayID{}, err
	}
	return types.RelayID{
		Network: network, ChainID: chainID,
	}, nil
}
