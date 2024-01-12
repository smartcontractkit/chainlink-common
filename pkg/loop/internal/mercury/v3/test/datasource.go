package mercury_v3_test

import (
	"context"

	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type StaticDataSource struct{}

var _ mercury_v3_types.DataSource = StaticDataSource{}

func (StaticDataSource) Observe(ctx context.Context, repts ocrtypes.ReportTimestamp, fetchMaxFinalizedTimestamp bool) (mercury_v3_types.Observation, error) {

	return mercury_v3_types.Observation{}, nil
}
