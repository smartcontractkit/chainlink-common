package datastreams_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	feedIDAStr         = "0x1111111111111111111100000000000000000000000000000000000000000000"
	feedIDBStr         = "0x2222222222222222222200000000000000000000000000000000000000000000"
	testFullReportAHex = "0x1111aabbccddeeff"
	testFullReportBHex = "0x2222aabbccddeeff"
)

func TestFeedID_Validate(t *testing.T) {
	_, err := datastreams.NewFeedID("012345678901234567890123456789012345678901234567890123456789000000")
	require.Error(t, err)

	_, err = datastreams.NewFeedID("0x1234")
	require.Error(t, err)

	_, err = datastreams.NewFeedID("0x123zzz")
	require.Error(t, err)

	_, err = datastreams.NewFeedID("0x0001013ebd4ed3f5889FB5a8a52b42675c60c1a8c42bc79eaa72dcd922ac4292")
	require.Error(t, err)

	_, err = datastreams.NewFeedID(feedIDAStr)
	require.NoError(t, err)
}

func Test_UnwrapStreamsTriggerEventToFeedReportList(t *testing.T) {
	feedReports := []datastreams.FeedReport{
		{
			FeedID:        feedIDAStr,
			FullReport:    randomByteArray(t, 1000),
			ReportContext: randomByteArray(t, 96),
			Signatures:    [][]byte{randomByteArray(t, 65), randomByteArray(t, 65)},
		},
		{
			FeedID:        feedIDBStr,
			FullReport:    randomByteArray(t, 1000),
			ReportContext: randomByteArray(t, 96),
			Signatures:    [][]byte{randomByteArray(t, 65), randomByteArray(t, 65)},
		},
	}

	payload := datastreams.StreamsTriggerEvent{
		Payload: feedReports,
	}
	wrapped, err := values.Wrap(payload)
	require.NoError(t, err)

	unwrapped, err := datastreams.UnwrapStreamsTriggerEventToFeedReportList(wrapped)
	require.NoError(t, err)
	require.Equal(t, feedReports, unwrapped)
}

func randomByteArray(t *testing.T, n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b[:])
	require.NoError(t, err)
	return b
}
