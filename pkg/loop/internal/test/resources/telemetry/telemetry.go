package telemetry_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ grpc.ClientConnInterface = (*mockClientConn)(nil)

type StaticTelemetryConfig struct {
	ChainID    string
	ContractID string
	Network    string
	Payload    []byte
	TelemType  string
}

type staticEndpoint struct {
	network    string
	chainID    string
	contractID string
	telemType  string

	StaticTelemetry
}

func (s staticEndpoint) SendLog(ctx context.Context, log []byte) error {
	return s.StaticTelemetry.Send(ctx, s.network, s.chainID, s.contractID, s.telemType, log)
}

type StaticTelemetry struct {
	StaticTelemetryConfig
}

func (s StaticTelemetry) NewEndpoint(ctx context.Context, network string, chainID string, contractID string, telemType string) (types.TelemetryClientEndpoint, error) {
	return staticEndpoint{
		network:         network,
		chainID:         chainID,
		contractID:      contractID,
		telemType:       telemType,
		StaticTelemetry: s,
	}, nil
}

func (s StaticTelemetry) Send(ctx context.Context, n string, chid string, conid string, t string, p []byte) error {
	if n != s.Network {
		return fmt.Errorf("expected %s but got %s", s.Network, n)
	}
	if chid != s.ChainID {
		return fmt.Errorf("expected %s but got %s", s.ChainID, chid)
	}
	if conid != s.ContractID {
		return fmt.Errorf("expected %s but got %s", s.ContractID, conid)
	}
	if t != s.TelemType {
		return fmt.Errorf("expected %s but got %s", s.TelemType, t)
	}
	if !bytes.Equal(p, s.Payload) {
		return fmt.Errorf("expected %s but got %s", s.Payload, p)
	}
	return nil
}

type mockClientConn struct{}

func (m mockClientConn) Invoke(ctx context.Context, method string, args any, reply any, opts ...grpc.CallOption) error {
	return nil
}

func (m mockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func TestTelemetry(t *testing.T) {
	tsc := internal.NewTelemetryServiceClient(mockClientConn{})
	c := internal.NewTelemetryClient(tsc)

	type sendTest struct {
		contractID    string
		telemetryType string
		network       string
		chainID       string
		payload       []byte

		shouldError bool
		error       string
	}

	sendTests := []sendTest{
		{
			contractID:    "",
			telemetryType: "",
			network:       "",
			chainID:       "",
			payload:       nil,
			shouldError:   true,
			error:         "contractID cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "",
			network:       "",
			chainID:       "",
			payload:       nil,
			shouldError:   true,
			error:         "telemetryType cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "",
			chainID:       "",
			payload:       nil,
			shouldError:   true,
			error:         "network cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "some-network",
			chainID:       "",
			payload:       nil,
			shouldError:   true,
			error:         "chainId cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "some-network",
			chainID:       "some-chainID",
			payload:       nil,
			shouldError:   true,
			error:         "payload cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "some-network",
			chainID:       "some-chainID",
			payload:       []byte("some-data"),
			shouldError:   false,
		},
	}

	for _, test := range sendTests {
		err := c.Send(context.Background(), test.network, test.chainID, test.contractID, test.telemetryType, test.payload)
		if test.shouldError {
			require.ErrorContains(t, err, test.error)
		} else {
			require.NoError(t, err)
		}
	}

	type genMonitoringEndpointTest struct {
		contractID    string
		telemetryType string
		network       string
		chainID       string

		shouldError bool
		error       string
	}

	genMonitoringEndpointTests := []genMonitoringEndpointTest{
		{
			contractID:    "",
			telemetryType: "",
			network:       "",
			chainID:       "",
			shouldError:   true,
			error:         "contractID cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "",
			network:       "",
			chainID:       "",
			shouldError:   true,
			error:         "telemetryType cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "",
			chainID:       "",
			shouldError:   true,
			error:         "network cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "some-network",
			chainID:       "",
			shouldError:   true,
			error:         "chainId cannot be empty",
		},
		{
			contractID:    "some-contractID",
			telemetryType: "some-telemetryType",
			network:       "some-network",
			chainID:       "some-chainID",
			shouldError:   false,
		},
	}

	for _, test := range genMonitoringEndpointTests {
		e, err := c.NewEndpoint(context.Background(), test.network, test.chainID, test.contractID, test.telemetryType)
		if test.shouldError {
			require.Nil(t, e)
			require.ErrorContains(t, err, test.error)
		} else {
			require.NotNil(t, e)
			require.Nil(t, err)
			require.Nil(t, e.SendLog(context.Background(), []byte("some-data")))
		}
	}
}
