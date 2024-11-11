package triggers

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	eventID    = "ev_id_1"
	rawReport1 = "abcd"
	rawReport2 = "efgh"
	capID      = "streams-trigger@3.2.1"
)

type testMercuryCodec struct {
}

func (c testMercuryCodec) Unwrap(wrapped values.Value) ([]datastreams.FeedReport, error) {
	dest := datastreams.StreamsTriggerEvent{}
	err := wrapped.UnwrapTo(&dest)
	return dest.Payload, err
}

func (c testMercuryCodec) Validate(report datastreams.FeedReport, _ [][]byte, _ int) error {
	return nil
}

func (c testMercuryCodec) Wrap(reports []datastreams.FeedReport) (values.Value, error) {
	return values.Wrap(reports)
}

func TestMercuryRemoteAggregator(t *testing.T) {
	agg := NewMercuryRemoteAggregator(testMercuryCodec{}, nil, 0, capID, logger.Nop())
	signatures := [][]byte{{1, 2, 3}}

	feed1Old := datastreams.FeedReport{
		FeedID:               feedOne,
		BenchmarkPrice:       big.NewInt(100).Bytes(),
		ObservationTimestamp: 100,
		FullReport:           []byte(rawReport1),
		ReportContext:        []byte{},
		Signatures:           signatures,
	}
	feed1New := datastreams.FeedReport{
		FeedID:               feedOne,
		BenchmarkPrice:       big.NewInt(200).Bytes(),
		ObservationTimestamp: 200,
		FullReport:           []byte(rawReport1),
		ReportContext:        []byte{},
		Signatures:           signatures,
	}
	feed2Old := datastreams.FeedReport{
		FeedID:               feedTwo,
		BenchmarkPrice:       big.NewInt(300).Bytes(),
		ObservationTimestamp: 300,
		FullReport:           []byte(rawReport2),
		ReportContext:        []byte{},
		Signatures:           signatures,
	}
	feed2New := datastreams.FeedReport{
		FeedID:               feedTwo,
		BenchmarkPrice:       big.NewInt(400).Bytes(),
		ObservationTimestamp: 400,
		FullReport:           []byte(rawReport2),
		ReportContext:        []byte{},
		Signatures:           signatures,
	}

	rawNode1Resp := getRawResponse(t, []datastreams.FeedReport{feed1Old, feed2New}, 400)
	rawNode2Resp := getRawResponse(t, []datastreams.FeedReport{feed1New, feed2Old}, 300)

	// aggregator should return latest value for each feedID
	aggResponse, err := agg.Aggregate(eventID, [][]byte{rawNode1Resp, rawNode2Resp})
	require.NoError(t, err)
	aggEvent := aggResponse.Event
	decodedReports, err := testMercuryCodec{}.Unwrap(aggEvent.Outputs)
	require.NoError(t, err)

	require.Len(t, decodedReports, 2)
	require.Equal(t, feed1New, decodedReports[0])
	require.Equal(t, feed2New, decodedReports[1])

	// never roll back to an older report
	rawNode3Resp := getRawResponse(t, []datastreams.FeedReport{feed1Old, feed2Old}, 400)
	aggResponse, err = agg.Aggregate(eventID, [][]byte{rawNode3Resp})
	require.NoError(t, err)
	decodedReports, err = testMercuryCodec{}.Unwrap(aggResponse.Event.Outputs)
	require.NoError(t, err)

	require.Len(t, decodedReports, 2)
	require.Equal(t, feed1New, decodedReports[0])
	require.Equal(t, feed2New, decodedReports[1])
}

func getRawResponse(t *testing.T, reports []datastreams.FeedReport, timestamp int64) []byte {
	resp, err := wrapReports(reports, eventID, timestamp, datastreams.Metadata{}, capID)
	require.NoError(t, err)
	rawResp, err := pb.MarshalTriggerResponse(resp)
	require.NoError(t, err)
	return rawResp
}
