package data_feeds

import (
	"encoding/base64"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	mercury_v3 "github.com/smartcontractkit/chainlink-common/pkg/capabilities/write_target/report/mercury/v3"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/write_target/report/platform"
)

func TestDecodeReportV3(t *testing.T) {
	// Base64-encoded report data (example)
	// Test case sourced from runtime logs
	// version | workflow_execution_id | timestamp | don_id | config_version | ... | data
	encoded := "AYFtgPpLuLNQysw6LjlSNrzGuBOwVoth7qC9PmunIY3TZvW/cAAAAAEAAAABvAbzAOeX1ahXVjehSq4T4/hQgAjR/FT0xGEf/xemjLAwMDAwRk9PQkFSAAAAAAAAAAAAAAAAAAAAAAAAAKoAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAHAAAMREREREREREQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEgAAMREREREREREQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZvW/aQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABm9b9pAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAElCUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASUJQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABnBQGpAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAElCUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASUJQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABJQlAAMiIiIiIiIiIgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEgAAMiIiIiIiIiIgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAZvW/aQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABm9b9pAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAElCUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASUJQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABnBQGpAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAElCUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAASUJQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABJQl"

	// Decode the base64 data
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	require.NoError(t, err)

	// Decode the report
	report, err := platform.Decode(decoded)
	require.NoError(t, err)
	t.Log(fmt.Sprintf("Decoded as report: %+v", report))

	expectedFeedID := [][32]uint8{
		[32]uint8{0x0, 0x3, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		[32]uint8{0x0, 0x3, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	}

	expectedData := []mercury_v3.Report{
		mercury_v3.Report{
			FeedId:                [32]uint8{0x0, 0x3, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			ObservationsTimestamp: 0x66f5bf69,
			BenchmarkPrice:        big.NewInt(300069),
			Bid:                   big.NewInt(300069),
			Ask:                   big.NewInt(300069),
			ValidFromTimestamp:    0x66f5bf69,
			ExpiresAt:             0x670501a9,
			LinkFee:               big.NewInt(300069),
			NativeFee:             big.NewInt(300069),
		},
		mercury_v3.Report{
			FeedId:                [32]uint8{0x0, 0x3, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			ObservationsTimestamp: 0x66f5bf69,
			BenchmarkPrice:        big.NewInt(300069),
			Bid:                   big.NewInt(300069),
			Ask:                   big.NewInt(300069),
			ValidFromTimestamp:    0x66f5bf69,
			ExpiresAt:             0x670501a9,
			LinkFee:               big.NewInt(300069),
			NativeFee:             big.NewInt(300069),
		},
	}

	reports, err := Decode(report.Data)
	require.NoError(t, err)
	t.Log(fmt.Sprintf("Decoded as DF reports: %+v", reports))
	require.Equal(t, len(expectedFeedID), len(*reports))

	for i, report := range *reports {
		require.Equal(t, expectedFeedID[i], report.FeedId)
		require.True(t, len(report.Data) > 0)

		m, err := mercury_v3.Decode(report.Data)
		require.NoError(t, err)
		require.Equal(t, expectedData[i], *m)
	}
}
