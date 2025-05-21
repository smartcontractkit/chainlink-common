package relayerset

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
	mocks2 "github.com/smartcontractkit/chainlink-common/pkg/types/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func Test_RelayerSet(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayer2 := mocks.NewRelayer(t)

	relayers := map[types.RelayID]core.Relayer{
		{
			Network: "N1",
			ChainID: "C1",
		}: relayer1,
		{
			Network: "N2",
			ChainID: "C2",
		}: relayer2,
	}

	pluginName := "relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	relayerClient, err := rc.Get(ctx, types.RelayID{
		Network: "N1",
		ChainID: "C1",
	})

	require.NoError(t, err)

	relayer1.On("Start", mock.Anything).Return(nil)
	err = relayerClient.Start(ctx)
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Start", mock.Anything)

	relayer1.On("Close").Return(nil)
	err = relayerClient.Close()
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Close")

	relayer1.On("Ready").Return(nil)
	err = relayerClient.Ready()
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Ready")

	relayer1.On("HealthReport").Return(map[string]error{"stat1": errors.New("error1")})
	healthReport := relayerClient.HealthReport()
	require.Len(t, healthReport, 1)
	require.Equal(t, "error1", healthReport["stat1"].Error())
	relayer1.AssertCalled(t, "HealthReport")

	relayer1.On("Name").Return("test-relayer")
	name := relayerClient.Name()
	require.Equal(t, "test-relayer", name)
	relayer1.AssertCalled(t, "Name")
}

func Test_RelayerSet_ContractReader(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayer2 := mocks.NewRelayer(t)

	relayers := map[types.RelayID]core.Relayer{
		{
			Network: "N1",
			ChainID: "C1",
		}: relayer1,
		{
			Network: "N2",
			ChainID: "C2",
		}: relayer2,
	}

	pluginName := "relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{
		Network: "N1",
		ChainID: "C1",
	})

	require.NoError(t, err)

	reader := &TestContractReader{mockedContractReader: mocks2.NewContractReader(t)}
	relayer1.On("NewContractReader", mock.Anything, mock.Anything).Return(reader, nil)

	fetchedReader, err := retrievedRelayer.NewContractReader(ctx, []byte("config"))
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().Start(mock.Anything).Return(nil)
	err = fetchedReader.Start(ctx)
	require.NoError(t, err)

	returnVal := map[any]any{}
	reader.mockedContractReader.EXPECT().GetLatestValue(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = fetchedReader.GetLatestValue(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().GetLatestValueWithHeadData(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	_, err = fetchedReader.GetLatestValueWithHeadData(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().QueryKey(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	_, err = fetchedReader.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().QueryKeys(mock.Anything, mock.Anything, mock.Anything).Return(func(yield func(string, types.Sequence) bool) {}, nil)
	_, err = fetchedReader.QueryKeys(ctx, []types.ContractKeyFilter{}, query.LimitAndSort{})
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().Close().Return(nil)
	err = fetchedReader.Close()
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().GetLatestValue(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = fetchedReader.GetLatestValue(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.ErrorContains(t, err, "contract reader not found")
}

type TestContractReader struct {
	types.UnimplementedContractReader
	mockedContractReader *mocks2.ContractReader
}

func (t *TestContractReader) Start(ctx context.Context) error {
	return t.mockedContractReader.Start(ctx)
}

func (t *TestContractReader) Close() error {
	return t.mockedContractReader.Close()
}

func (t *TestContractReader) Ready() error {
	return t.mockedContractReader.Ready()
}
func (t *TestContractReader) HealthReport() map[string]error {
	return t.mockedContractReader.HealthReport()
}

func (t *TestContractReader) Name() string {
	return t.mockedContractReader.Name()
}

func (t *TestContractReader) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	return t.mockedContractReader.GetLatestValue(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (t *TestContractReader) GetLatestValueWithHeadData(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) (*types.Head, error) {
	return t.mockedContractReader.GetLatestValueWithHeadData(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (t *TestContractReader) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return t.mockedContractReader.BatchGetLatestValues(ctx, request)
}

func (t *TestContractReader) Bind(ctx context.Context, bindings []types.BoundContract) error {
	return t.mockedContractReader.Bind(ctx, bindings)
}

func (t *TestContractReader) Unbind(ctx context.Context, bindings []types.BoundContract) error {
	return t.mockedContractReader.Unbind(ctx, bindings)
}

func (t *TestContractReader) QueryKey(ctx context.Context, boundContract types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	return t.mockedContractReader.QueryKey(ctx, boundContract, filter, limitAndSort, sequenceDataType)
}

func (t *TestContractReader) QueryKeys(ctx context.Context, keyQueries []types.ContractKeyFilter, limitAndSort query.LimitAndSort) (iter.Seq2[string, types.Sequence], error) {
	return t.mockedContractReader.QueryKeys(ctx, keyQueries, limitAndSort)
}

type TestRelayerSet struct {
	relayers map[types.RelayID]core.Relayer
}

func (t *TestRelayerSet) Get(ctx context.Context, relayID types.RelayID) (core.Relayer, error) {
	if relayer, ok := t.relayers[relayID]; ok {
		return relayer, nil
	}

	return nil, fmt.Errorf("relayer with id %s not found", relayID)
}

func (t *TestRelayerSet) List(ctx context.Context, relayIDs ...types.RelayID) (map[types.RelayID]core.Relayer, error) {
	return t.relayers, nil
}

type testRelaySetPlugin struct {
	log logger.Logger
	plugin.NetRPCUnsupportedPlugin
	brokerExt *net.BrokerExt
	impl      core.RelayerSet
}

func (r *testRelaySetPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker

	return NewRelayerSetClient(logger.Nop(), r.brokerExt, client), nil
}

func (r *testRelaySetPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker

	rs, _ := NewRelayerSetServer(r.log, r.impl, r.brokerExt)
	relayerset.RegisterRelayerSetServer(server, rs)
	return nil
}
