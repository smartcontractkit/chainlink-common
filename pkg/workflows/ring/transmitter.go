package ring

import (
	"context"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/ring/pb"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ ocr3types.ContractTransmitter[[]byte] = (*Transmitter)(nil)

// Transmitter handles transmission of shard orchestration outcomes
type Transmitter struct {
	lggr          logger.Logger
	store         *Store
	arbiterScaler pb.ArbiterScalerClient
	fromAccount   types.Account
}

func NewTransmitter(lggr logger.Logger, store *Store, arbiterScaler pb.ArbiterScalerClient, fromAccount types.Account) *Transmitter {
	return &Transmitter{lggr: lggr, store: store, arbiterScaler: arbiterScaler, fromAccount: fromAccount}
}

func (t *Transmitter) Transmit(ctx context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[[]byte], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		t.lggr.Errorf("failed to unmarshal report")
		return err
	}

	if err := t.notifyArbiter(ctx, outcome.State); err != nil {
		t.lggr.Errorf("failed to notify arbiter", "err", err)
		return err
	}

	for workflowID, route := range outcome.Routes {
		t.store.SetShardForWorkflow(workflowID, route.Shard)
		t.lggr.Debugw("Updated workflow shard mapping", "workflowID", workflowID, "shard", route.Shard)
	}

	return nil
}

func (t *Transmitter) notifyArbiter(ctx context.Context, state *pb.RoutingState) error {
	if state == nil {
		return nil
	}

	var nShards uint32
	switch s := state.State.(type) {
	case *pb.RoutingState_RoutableShards:
		nShards = s.RoutableShards
		t.lggr.Infow("Transmitting shard routing", "routableShards", nShards)
	case *pb.RoutingState_Transition:
		nShards = s.Transition.WantShards
		t.lggr.Infow("Transmitting shard routing (in transition)", "wantShards", nShards)
	}

	if t.arbiterScaler != nil && nShards > 0 {
		if _, err := t.arbiterScaler.ConsensusWantShards(ctx, &pb.ConsensusWantShardsRequest{NShards: nShards}); err != nil {
			return err
		}
	}

	return nil
}

func (t *Transmitter) FromAccount(ctx context.Context) (types.Account, error) {
	return t.fromAccount, nil
}
