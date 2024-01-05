package mercury_v3

import (
	"context"
	"math/big"

	"google.golang.org/grpc"

	mercury_common_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury"
	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"

	//ocr_types "github.com/smartcontractkit/libocr/offchainreporting/types"
	ocr2plus_types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	mercury_v3_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v3"
)

var _ mercury_v3_types.DataSource = (*DataSourceClient)(nil)

type DataSourceClient struct {
	grpc mercury_v3_pb.DataSourceClient
}

func newDataSourceClient(cc grpc.ClientConnInterface) *DataSourceClient {
	return &DataSourceClient{grpc: mercury_v3_pb.NewDataSourceClient(cc)}
}

func (d *DataSourceClient) Observe(ctx context.Context, timestamp ocr2plus_types.ReportTimestamp, fetchMaxFinalizedTimestamp bool) (mercury_v3_types.Observation, error) {
	reply, err := d.grpc.Observe(ctx, &mercury_v3_pb.ObserveRequest{
		ReportTimestamp: internal.PbReportTimestamp(timestamp),
	})
	if err != nil {
		return mercury_v3_types.Observation{}, err
	}
	panic("fetchMaxFinalizedTimestamp not implemented")
	return observation(reply), nil
}

var _ mercury_v3_pb.DataSourceServer = (*dataSourceServer)(nil)

type dataSourceServer struct {
	mercury_v3_pb.UnimplementedDataSourceServer

	impl mercury_v3_types.DataSource
}

func (d *dataSourceServer) Observe(ctx context.Context, request *mercury_v3_pb.ObserveRequest) (*mercury_v3_pb.ObserveResponse, error) {
	timestamp, err := internal.ReportTimestamp(request.ReportTimestamp)
	if err != nil {
		return nil, err
	}
	val, err := d.impl.Observe(ctx, timestamp, request.FetchMaxFinalizedBlockNum)
	if err != nil {
		return nil, err
	}
	return &mercury_v3_pb.ObserveResponse{Observation: pbObservation(val)}, nil
}

func observation(resp *mercury_v3_pb.ObserveResponse) mercury_v3_types.Observation {
	// TODO: figure out what to do with the Err field. should it be the resp error? that seems wrong b/c
	// the Err field is one all the Observation fields.
	return mercury_v3_types.Observation{
		BenchmarkPrice:        mercury_common_types.ObsResult[*big.Int]{Val: resp.Observation.BenchmarkPrice.Int()},
		Bid:                   mercury_common_types.ObsResult[*big.Int]{Val: resp.Observation.Bid.Int()},
		Ask:                   mercury_common_types.ObsResult[*big.Int]{Val: resp.Observation.Ask.Int()},
		MaxFinalizedTimestamp: mercury_common_types.ObsResult[int64]{Val: resp.Observation.MaxFinalizedTimestamp},
		LinkPrice:             mercury_common_types.ObsResult[*big.Int]{Val: resp.Observation.LinkPrice.Int()},
		NativePrice:           mercury_common_types.ObsResult[*big.Int]{Val: resp.Observation.NativePrice.Int()},
	}
}

func pbObservation(obs mercury_v3_types.Observation) *mercury_v3_pb.Observation {
	return &mercury_v3_pb.Observation{
		BenchmarkPrice:        pb.NewBigIntFromInt(obs.BenchmarkPrice.Val),
		Bid:                   pb.NewBigIntFromInt(obs.Bid.Val),
		Ask:                   pb.NewBigIntFromInt(obs.Ask.Val),
		MaxFinalizedTimestamp: obs.MaxFinalizedTimestamp.Val,
		LinkPrice:             pb.NewBigIntFromInt(obs.LinkPrice.Val),
		NativePrice:           pb.NewBigIntFromInt(obs.NativePrice.Val),
	}
}
