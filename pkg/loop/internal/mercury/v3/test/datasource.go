package v3_test

import (
	"context"
	"fmt"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"

	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
)

var DataSourceImpl = staticDataSource{}

type DataSourceEvaluator interface {
	mercury_v3_types.DataSource
	// Evaluate runs the other DataSource and checks that
	// the results are equal to this one
	Evaluate(ctx context.Context, other mercury_v3_types.DataSource) error
}

type staticDataSource struct{}

var _ DataSourceEvaluator = staticDataSource{}

func (staticDataSource) Observe(ctx context.Context, repts ocrtypes.ReportTimestamp, fetchMaxFinalizedTimestamp bool) (mercury_v3_types.Observation, error) {
	return Fixtures.Observation, nil
}

func (staticDataSource) Evaluate(ctx context.Context, other mercury_v3_types.DataSource) error {
	gotVal, err := other.Observe(ctx, Fixtures.ReportTimestamp, false)
	if err != nil {
		return fmt.Errorf("failed to observe dataSource: %w", err)
	}
	if !assert.ObjectsAreEqual(Fixtures.Observation, gotVal) {
		return fmt.Errorf("expected Value %v but got %v", Fixtures.Observation, gotVal)
	}
	return nil
}
