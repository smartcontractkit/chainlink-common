package securemint

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/securemint"
	sm "github.com/smartcontractkit/chainlink-common/pkg/types/core/securemint"
)

// externalAdapterClient is a protobuf client that implements the securemint.ExternalAdapter interface.
// It's basically a wrapper around the protobuf external adapter client so that it can be used as a securemint.ExternalAdapter.
type externalAdapterClient struct {
	lggr logger.Logger
	grpc pb.ExternalAdapterClient
}

var _ sm.ExternalAdapter = (*externalAdapterClient)(nil)

func newExternalAdapterClient(lggr logger.Logger, cc grpc.ClientConnInterface) *externalAdapterClient {
	return &externalAdapterClient{lggr: logger.Named(lggr, "ExternalAdapterClient"), grpc: pb.NewExternalAdapterClient(cc)}
}

func (d *externalAdapterClient) GetPayload(ctx context.Context, blocks sm.Blocks) (sm.ExternalAdapterPayload, error) {
	d.lggr.Infof("GetPayload request pb client: %+v", blocks)

	request := &pb.Blocks{
		Value: make(map[uint64]uint64, len(blocks)),
	}
	for chainSelector, blockNumber := range blocks {
		request.Value[uint64(chainSelector)] = uint64(blockNumber)
	}

	reply, err := d.grpc.GetPayload(ctx, request)
	if err != nil {
		return sm.ExternalAdapterPayload{}, err
	}

	mintables := make(map[sm.ChainSelector]sm.BlockMintablePair, len(reply.Mintables))
	for chainSelector, blockMintablePair := range reply.Mintables {
		mintable, err := stringToBigInt(blockMintablePair.Mintable)
		if err != nil {
			return sm.ExternalAdapterPayload{}, err
		}
		mintables[sm.ChainSelector(chainSelector)] = sm.BlockMintablePair{
			Block:    sm.BlockNumber(blockMintablePair.BlockNumber),
			Mintable: mintable,
		}
	}

	reserveAmount, err := stringToBigInt(reply.ReserveInfo.ReserveAmount)
	if err != nil {
		return sm.ExternalAdapterPayload{}, err
	}
	reserveInfo := sm.ReserveInfo{
		ReserveAmount: reserveAmount,
		Timestamp:     reply.ReserveInfo.Timestamp.AsTime(),
	}

	latestBlocks := make(sm.Blocks, len(reply.LatestBlocks.Value))
	for chainSelector, blockNumber := range reply.LatestBlocks.Value {
		latestBlocks[sm.ChainSelector(chainSelector)] = sm.BlockNumber(blockNumber)
	}

	result := sm.ExternalAdapterPayload{
		Mintables:    mintables,
		ReserveInfo:  reserveInfo,
		LatestBlocks: latestBlocks,
	}

	d.lggr.Infof("GetPayload response pb client: %+v", result)
	return result, nil
}

var _ pb.ExternalAdapterServer = (*externalAdapterServer)(nil)

// externalAdapterServer is a protobuf server that implements the pb.ExternalAdapterServer interface.
// It's basically a protobuf wrapper around the securemint.ExternalAdapter implementation.
type externalAdapterServer struct {
	pb.UnimplementedExternalAdapterServer

	lggr logger.Logger
	impl sm.ExternalAdapter
}

func newExternalAdapterServer(lggr logger.Logger, impl sm.ExternalAdapter) *externalAdapterServer {
	return &externalAdapterServer{lggr: logger.Named(lggr, "ExternalAdapterServer"), impl: impl}
}

func (d *externalAdapterServer) GetPayload(ctx context.Context, request *pb.Blocks) (*pb.ExternalAdapterPayload, error) {
	d.lggr.Infof("GetPayload request pb server: %+v", request)

	blocks := make(sm.Blocks, len(request.Value))
	for chainSelector, blockNumber := range request.Value {
		blocks[sm.ChainSelector(chainSelector)] = sm.BlockNumber(blockNumber)
	}

	val, err := d.impl.GetPayload(ctx, blocks)
	if err != nil {
		return nil, fmt.Errorf("failed to get payload from external adapter for request %v: %w", request, err)
	}

	mintables := make(map[uint64]*pb.BlockMintablePair, len(val.Mintables))
	for chainSelector, blockMintablePair := range val.Mintables {
		mintables[uint64(chainSelector)] = &pb.BlockMintablePair{
			BlockNumber: uint64(blockMintablePair.Block),
			Mintable:    blockMintablePair.Mintable.String(),
		}
	}

	reserveInfo := &pb.ReserveInfo{
		ReserveAmount: val.ReserveInfo.ReserveAmount.String(),
		Timestamp:     timestamppb.New(val.ReserveInfo.Timestamp),
	}

	valLatestBlocks := make(map[uint64]uint64, len(val.LatestBlocks))
	for chainSelector, blockNumber := range val.LatestBlocks {
		valLatestBlocks[uint64(chainSelector)] = uint64(blockNumber)
	}
	latestBlocks := &pb.Blocks{
		Value: valLatestBlocks,
	}

	response := &pb.ExternalAdapterPayload{
		Mintables:    mintables,
		ReserveInfo:  reserveInfo,
		LatestBlocks: latestBlocks,
	}

	d.lggr.Infof("GetPayload response pb server: %+v", response)
	return response, nil
}

func stringToBigInt(s string) (*big.Int, error) {
	z := new(big.Int)
	_, ok := z.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid integer %q", s)
	}
	return z, nil
}
