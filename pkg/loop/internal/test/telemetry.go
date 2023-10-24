package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
)

var _ grpc.ClientConnInterface = (*mockClientConn)(nil)

type mockClientConn struct {
}

func (m mockClientConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return nil
}

func (m mockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func TelemetryClient(t *testing.T) {
	mcc := mockClientConn{}
	lggr, ol := logger.TestObserved(t, zapcore.ErrorLevel)
	c := internal.NewTelemetryClient(&mcc, lggr)

	err := c.Send(context.Background(), "", "", "", "", nil)
	require.ErrorContains(t, err, "contractID cannot be empty")

	err = c.Send(context.Background(), uuid.New().String(), "", "", "", nil)
	require.ErrorContains(t, err, "telemetryType cannot be empty")

	err = c.Send(context.Background(), uuid.New().String(), "some-type", "", "", nil)
	require.ErrorContains(t, err, "network cannot be empty")

	err = c.Send(context.Background(), uuid.New().String(), "some-type", "some-network", "", nil)
	require.ErrorContains(t, err, "chainId cannot be empty")

	err = c.Send(context.Background(), uuid.New().String(), "some-type", "some-network", "some-chain-id", nil)
	require.ErrorContains(t, err, "payload cannot be empty")

	err = c.Send(context.Background(), uuid.New().String(), "some-type", "some-network", "some-chain-id", []byte("some-data"))
	require.NoError(t, err)

	e := c.GenMonitoringEndpoint("", "", "", "")
	require.Nil(t, e)
	require.Equal(t, 1, ol.Len())
	require.Contains(t, ol.TakeAll()[0].Message, "cannot generate monitoring endpoint, contractID is empty")

	e = c.GenMonitoringEndpoint("some-contractID", "", "", "")
	require.Nil(t, e)
	require.Equal(t, 1, ol.Len())
	require.Contains(t, ol.TakeAll()[0].Message, "cannot generate monitoring endpoint, telemetryType is empty")

	e = c.GenMonitoringEndpoint("some-contractID", "some-type", "", "")
	require.Nil(t, e)
	require.Equal(t, 1, ol.Len())
	require.Contains(t, ol.TakeAll()[0].Message, "cannot generate monitoring endpoint, network is empty")

	e = c.GenMonitoringEndpoint("some-contractID", "some-type", "some-network", "")
	require.Nil(t, e)
	require.Equal(t, 1, ol.Len())
	require.Contains(t, ol.TakeAll()[0].Message, "cannot generate monitoring endpoint, chainID is empty")

	e = c.GenMonitoringEndpoint("some-contractID", "some-type", "some-network", "some-chainID")
	require.NotNil(t, e)
	require.Equal(t, 0, ol.Len())

	e.SendLog([]byte("some-data"))
	require.Equal(t, 0, ol.Len())
}

type staticTelemetry struct {
	endpoints map[string]staticEndpoint
}

type staticEndpoint struct {
}

func (s staticEndpoint) SendLog(log []byte) {}

func (s staticTelemetry) GenMonitoringEndpoint(contractID string, telemType string, network string, chainID string) commontypes.MonitoringEndpoint {
	s.endpoints[fmt.Sprintf("%s_%s_%s_%s", contractID, telemType, network, chainID)] = staticEndpoint{}
	return s.endpoints[fmt.Sprintf("%s_%s_%s_%s", contractID, telemType, network, chainID)]
}

func TelemetryServer(t *testing.T) {
	st := staticTelemetry{
		endpoints: make(map[string]staticEndpoint),
	}
	s := internal.NewTelemetryServer(st)
	_, err := s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID:       nil,
		ContractID:    "",
		TelemetryType: "",
		Payload:       nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "contractID cannot be empty")

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID:       nil,
		ContractID:    "some-contractID",
		TelemetryType: "",
		Payload:       nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "telemetryType cannot be empty")

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID:       nil,
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "RelayID cannot be nil")

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID: &pb.RelayID{
			Network: "",
			ChainId: "",
		},
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "RelayID.Network cannot be empty")

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID: &pb.RelayID{
			Network: "some-network",
			ChainId: "",
		},
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "RelayID.ChainId cannot be empty")

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID: &pb.RelayID{
			Network: "some-network",
			ChainId: "some-chainID",
		},
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.NoError(t, err)
	require.Len(t, st.endpoints, 1)

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID: &pb.RelayID{
			Network: "some-network",
			ChainId: "some-chainID",
		},
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.NoError(t, err)
	require.Len(t, st.endpoints, 1)

	_, err = s.Send(context.Background(), &pb.TelemetryMessage{
		RelayID: &pb.RelayID{
			Network: "some-other-network",
			ChainId: "some-chainID",
		},
		ContractID:    "some-contractID",
		TelemetryType: "some-type",
		Payload:       nil,
	})
	require.NoError(t, err)
	require.Len(t, st.endpoints, 2)
}
