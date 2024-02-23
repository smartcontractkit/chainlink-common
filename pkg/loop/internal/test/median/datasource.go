package median_test

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

var _ median.DataSource = (*StaticTestDataSource)(nil)

type DataSourceTestConfig struct {
	ReportContext types.ReportContext
	Value         *big.Int
}

type StaticTestDataSource struct {
	DataSourceTestConfig
}

func DefaultTestDataSource() StaticTestDataSource {
	return StaticTestDataSource{
		DataSourceTestConfig{
			ReportContext: reportContext,
			Value:         value,
		},
	}
}

func DefaultTestJuelsPerFeeCoinDataSource() StaticTestDataSource {
	return StaticTestDataSource{
		DataSourceTestConfig{
			ReportContext: reportContext,
			Value:         juelsPerFeeCoin,
		},
	}
}

func (s StaticTestDataSource) Observe(ctx context.Context, timestamp types.ReportTimestamp) (*big.Int, error) {
	if timestamp != s.ReportContext.ReportTimestamp {
		return nil, fmt.Errorf("expected %v but got %v", s.ReportContext.ReportTimestamp, timestamp)
	}
	return s.Value, nil
}

func (s StaticTestDataSource) Evaluate(ctx context.Context, ds median.DataSource) error {
	gotVal, err := ds.Observe(ctx, s.ReportContext.ReportTimestamp)
	if err != nil {
		return fmt.Errorf("failed to observe dataSource: %w", err)
	}
	if gotVal.Cmp(s.Value) != 0 {
		return fmt.Errorf("expected Value %s but got %s", value, gotVal)
	}
	return nil
}
