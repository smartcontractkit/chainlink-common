package median_test

import (
	"context"
	"fmt"
	"math/big"

	reportingplugin_test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/ocr2/reporting_plugin"
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

func DefaultTestDataSource() median.DataSource {
	return &StaticTestDataSource{
		DataSourceTestConfig{
			ReportContext: reportingplugin_test.DefaultReportingPluginTestConfig.ReportContext,
			Value:         value,
		},
	}
}

func DefaultTestJuelsPerFeeCoinDataSource() median.DataSource {
	return &StaticTestDataSource{
		DataSourceTestConfig{
			ReportContext: reportingplugin_test.DefaultReportingPluginTestConfig.ReportContext,
			Value:         juelsPerFeeCoin,
		},
	}
}

func (s *StaticTestDataSource) Observe(ctx context.Context, timestamp types.ReportTimestamp) (*big.Int, error) {
	if timestamp != s.ReportContext.ReportTimestamp {
		return nil, fmt.Errorf("expected %v but got %v", s.ReportContext.ReportTimestamp, timestamp)
	}
	return s.Value, nil
}
