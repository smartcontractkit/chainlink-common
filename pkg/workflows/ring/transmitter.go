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
	lggr        logger.Logger
	store       *Store
	fromAccount types.Account
}

func NewTransmitter(lggr logger.Logger, store *Store, fromAccount types.Account) *Transmitter {
	return &Transmitter{lggr: lggr, store: store, fromAccount: fromAccount}
}

func (t *Transmitter) Transmit(_ context.Context, _ types.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[[]byte], _ []types.AttributedOnchainSignature) error {
	outcome := &pb.Outcome{}
	if err := proto.Unmarshal(r.Report, outcome); err != nil {
		t.lggr.Errorf("failed to unmarshal report")
		return err
	}

	if outcome.State != nil {
		if routableShards, ok := outcome.State.State.(*pb.RoutingState_RoutableShards); ok {
			t.lggr.Infow("Transmitting shard routing", "routableShards", routableShards.RoutableShards)
		}
	}

	for workflowID, route := range outcome.Routes {
		t.store.SetShardForWorkflow(workflowID, route.Shard)
		t.lggr.Debugw("Updated workflow shard mapping", "workflowID", workflowID, "shard", route.Shard)
	}

	return nil
}

func (t *Transmitter) FromAccount(ctx context.Context) (types.Account, error) {
	return t.fromAccount, nil
}
