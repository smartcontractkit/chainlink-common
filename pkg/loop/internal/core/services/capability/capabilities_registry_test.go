package capability

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/hashicorp/go-plugin"
	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
)

var _ capabilities.BaseCapability = (*mockBaseCapability)(nil)

type mockBaseCapability struct {
	info capabilities.CapabilityInfo
}

func (f *mockBaseCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return f.info, nil
}

var _ capabilities.TriggerExecutable = (*mockTriggerExecutable)(nil)

type mockTriggerExecutable struct {
	mu       sync.RWMutex
	callback chan capabilities.TriggerResponse
}

func (f *mockTriggerExecutable) XXXTestingPushToCallbackChan(cr capabilities.TriggerResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callback <- cr
}

func (f *mockTriggerExecutable) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.callback, nil
}

func (f *mockTriggerExecutable) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callback = nil
	return nil
}

func (f *mockTriggerExecutable) AckEvent(ctx context.Context, triggerId string, eventId string, method string) error {
	return nil
}

var _ capabilities.Executable = (*mockExecutableCapability)(nil)

type mockExecutableCapability struct {
	registeredWorkflowRequest *capabilities.RegisterToWorkflowRequest
	callback                  chan capabilities.CapabilityResponse
}

func (f *mockExecutableCapability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	f.registeredWorkflowRequest = &request
	return nil
}

func (f *mockExecutableCapability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	f.registeredWorkflowRequest = nil
	return nil
}

func (f *mockExecutableCapability) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	return capabilities.CapabilityResponse{
		Value: nil,
	}, nil
}

var _ capabilities.TriggerCapability = (*mockTriggerCapability)(nil)

type mockTriggerCapability struct {
	*mockBaseCapability
	*mockTriggerExecutable
}

type mockNonTriggerCapability struct {
	*mockBaseCapability
	*mockExecutableCapability
}

type testRegistryPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerExt *net.BrokerExt
	impl      *mocks.CapabilitiesRegistry
}

func (r *testRegistryPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker
	return NewCapabilitiesRegistryClient(client, r.brokerExt), nil
}

func (r *testRegistryPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker
	pb.RegisterCapabilitiesRegistryServer(server, NewCapabilitiesRegistryServer(r.brokerExt, r.impl))
	return nil
}

func TestCapabilitiesRegistry(t *testing.T) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	reg := mocks.NewCapabilitiesRegistry(t)

	capabilityResponse := capabilities.CapabilityResponse{
		Value: values.EmptyMap(),
		Metadata: capabilities.ResponseMetadata{
			Metering: []capabilities.MeteringNodeDetail{},
		},
	}

	pluginName := "registry-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRegistryPlugin{
				impl: reg,
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: logger,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := regClient.(*capabilitiesRegistryClient)
	require.True(t, ok)

	// No capabilities in register
	reg.On("Get", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.Get(t.Context(), "some-id")
	require.ErrorContains(t, err, "capability not found")

	pid := p2ptypes.PeerID([32]byte{0: 3})
	expectedNode := capabilities.Node{
		PeerID: &pid,
		WorkflowDON: capabilities.DON{
			ID:   11,
			Name: "test-workflow-don",
			Members: []p2ptypes.PeerID{
				[32]byte{0: 2},
				[32]byte{0: 3},
			},
			F:             2,
			ConfigVersion: 1,
			Families:      []string{"a"},
			Config:        []byte("test-config"),
		},
		CapabilityDONs: []capabilities.DON{
			{
				ID:            22,
				Name:          "test-capability-don-1",
				Members:       []p2ptypes.PeerID{},
				F:             1,
				ConfigVersion: 2,
				Families:      []string{"a"},
				Config:        []byte("test-config"),
			},
			{
				ID:   33,
				Name: "test-capability-don-2",
				Members: []p2ptypes.PeerID{
					[32]byte{0: 4},
					[32]byte{0: 5},
					[32]byte{0: 6},
				},
				F:             3,
				ConfigVersion: 3,
				Families:      []string{"a"},
				Config:        []byte("test-config"),
			},
		},
	}

	reg.On("LocalNode", mock.Anything).Once().Return(expectedNode, nil)

	actualNode, err := rc.LocalNode(t.Context())
	require.NoError(t, err)
	ensureEqual(t, expectedNode, actualNode)

	// Check zero values for empty node
	emptyNode := capabilities.Node{}
	reg.On("LocalNode", mock.Anything).Once().Return(emptyNode, nil)
	actualNode, err = rc.LocalNode(t.Context())
	require.NoError(t, err)
	require.Nil(t, actualNode.PeerID)
	require.Equal(t, capabilities.DON{
		ID:            0,
		Members:       nil,
		F:             0,
		ConfigVersion: 0,
	}, actualNode.WorkflowDON)
	require.Empty(t, actualNode.CapabilityDONs)

	reg.On("NodeByPeerID", mock.Anything, pid).Once().Return(expectedNode, nil)

	actualNode, err = rc.NodeByPeerID(t.Context(), pid)
	require.NoError(t, err)
	ensureEqual(t, expectedNode, actualNode)

	// Check zero values for empty node
	reg.On("NodeByPeerID", mock.Anything, pid).Once().Return(emptyNode, nil)
	actualNode, err = rc.NodeByPeerID(t.Context(), pid)
	require.NoError(t, err)
	require.Nil(t, actualNode.PeerID)
	require.Equal(t, capabilities.DON{
		ID:            0,
		Members:       nil,
		F:             0,
		ConfigVersion: 0,
	}, actualNode.WorkflowDON)
	require.Empty(t, actualNode.CapabilityDONs)

	reg.On("GetTrigger", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.GetTrigger(t.Context(), "some-id")
	require.ErrorContains(t, err, "capability not found")

	reg.On("List", mock.Anything).Return([]capabilities.BaseCapability{}, nil)
	list, err := rc.List(t.Context())
	require.NoError(t, err)
	require.Len(t, list, 0)

	// Add capability Trigger
	triggerInfo := capabilities.CapabilityInfo{
		ID:             "trigger-1@1.0.0",
		CapabilityType: capabilities.CapabilityTypeTrigger,
		Description:    "trigger-1-description",
	}
	testTrigger := mockTriggerCapability{
		mockBaseCapability:    &mockBaseCapability{info: triggerInfo},
		mockTriggerExecutable: &mockTriggerExecutable{callback: make(chan capabilities.TriggerResponse, 10)},
	}

	// After adding the trigger, we'll expect something wrapped by the internal client type below.
	reg.On("Add", mock.Anything, mock.AnythingOfType("*capability.TriggerCapabilityClient")).Return(nil)
	err = rc.Add(t.Context(), testTrigger)
	require.NoError(t, err)

	reg.On("GetTrigger", mock.Anything, "trigger-1@1.0.0").Return(testTrigger, nil)
	triggerCap, err := rc.GetTrigger(t.Context(), "trigger-1@1.0.0")
	require.NoError(t, err)

	// Test trigger Info()
	testCapabilityInfo(t, triggerInfo, triggerCap)

	// Test TriggerExecutable
	triggerChan, err := triggerCap.RegisterTrigger(t.Context(), capabilities.TriggerRegistrationRequest{})
	require.NoError(t, err)

	triggerResponse := capabilities.TriggerResponse{
		Event: capabilities.TriggerEvent{
			Outputs: values.EmptyMap(),
		},
		Err: errors.New("some-error"),
	}
	testTrigger.XXXTestingPushToCallbackChan(triggerResponse)
	require.Equal(t, triggerResponse, <-triggerChan)

	err = triggerCap.UnregisterTrigger(t.Context(), capabilities.TriggerRegistrationRequest{})
	require.NoError(t, err)
	testTrigger.mu.RLock()
	require.Nil(t, testTrigger.callback)
	testTrigger.mu.RUnlock()

	// Add capability Trigger
	actionInfo := capabilities.CapabilityInfo{
		ID:             "action-1@2.0.0",
		CapabilityType: capabilities.CapabilityTypeAction,
		Description:    "action-1-description",
	}

	actionCallbackChan := make(chan capabilities.CapabilityResponse, 10)
	testAction := mockNonTriggerCapability{
		mockBaseCapability:       &mockBaseCapability{info: actionInfo},
		mockExecutableCapability: &mockExecutableCapability{callback: actionCallbackChan},
	}
	reg.On("GetExecutable", mock.Anything, "action-1@2.0.0").Return(testAction, nil)
	actionCap, err := rc.GetExecutable(t.Context(), "action-1@2.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, actionInfo, actionCap)

	// Test Executable
	workflowRequest := capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: "workflow-ID",
		},
	}

	err = actionCap.RegisterToWorkflow(t.Context(), workflowRequest)
	require.NoError(t, err)
	require.Equal(t, workflowRequest.Metadata.WorkflowID, testAction.registeredWorkflowRequest.Metadata.WorkflowID)

	actionCallbackChan <- capabilityResponse
	callbackChan, err := actionCap.Execute(t.Context(), capabilities.CapabilityRequest{})
	require.NoError(t, err)
	require.Equal(t, capabilityResponse, callbackChan)
	err = actionCap.UnregisterFromWorkflow(t.Context(), capabilities.UnregisterFromWorkflowRequest{})
	require.NoError(t, err)
	require.Nil(t, testAction.registeredWorkflowRequest)

	// Add capability Consensus
	consensusInfo := capabilities.CapabilityInfo{
		ID:             "consensus-1@3.0.0",
		CapabilityType: capabilities.CapabilityTypeConsensus,
		Description:    "consensus-1-description",
	}
	testConsensus := mockNonTriggerCapability{
		mockBaseCapability:       &mockBaseCapability{info: consensusInfo},
		mockExecutableCapability: &mockExecutableCapability{},
	}
	reg.On("GetExecutable", mock.Anything, "consensus-1@3.0.0").Return(testConsensus, nil)
	consensusCap, err := rc.GetExecutable(t.Context(), "consensus-1@3.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, consensusInfo, consensusCap)

	// Add capability Target
	targetInfo := capabilities.CapabilityInfo{
		ID:             "target-1@1.0.0",
		CapabilityType: capabilities.CapabilityTypeTarget,
		Description:    "target-1-description",
	}
	testTarget := mockNonTriggerCapability{
		mockBaseCapability:       &mockBaseCapability{info: targetInfo},
		mockExecutableCapability: &mockExecutableCapability{},
	}
	reg.On("Get", mock.Anything, "target-1@1.0.0").Return(testTarget, nil)
	targetCap, err := rc.Get(t.Context(), "target-1@1.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, targetInfo, targetCap)
}

func testCapabilityInfo(t *testing.T, expectedInfo capabilities.CapabilityInfo, cap capabilities.BaseCapability) {
	gotInfo, err := cap.Info(t.Context())
	require.NoError(t, err)
	require.Equal(t, expectedInfo.ID, gotInfo.ID)
	require.Equal(t, expectedInfo.CapabilityType, gotInfo.CapabilityType)
	require.Equal(t, expectedInfo.Description, gotInfo.Description)
	require.Equal(t, expectedInfo.Version(), gotInfo.Version())
}

func TestToDON(t *testing.T) {
	don := &pb.DON{
		Id:   0,
		Name: "test-don",
		Members: [][]byte{
			{0: 4, 31: 0},
			{0: 5, 31: 0},
		},
		F:             2,
		ConfigVersion: 1,
		Families:      []string{"a"},
		Config:        []byte("test-config"),
	}

	expected := capabilities.DON{
		ID:   0,
		Name: "test-don",
		Members: []p2ptypes.PeerID{
			[32]byte{0: 4},
			[32]byte{0: 5},
		},
		F:             2,
		ConfigVersion: 1,
		Families:      []string{"a"},
		Config:        []byte("test-config"),
	}

	actual := toDON(don)

	require.Equal(t, expected, actual)
}

func TestCapabilitiesRegistry_ConfigForCapabilities_IncludingV2Methods(t *testing.T) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	reg := mocks.NewCapabilitiesRegistry(t)

	pluginName := "registry-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRegistryPlugin{
				impl: reg,
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: logger,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := regClient.(*capabilitiesRegistryClient)
	require.True(t, ok)

	capID := "some-cap@1.0.0"
	donID := uint32(1)
	wm, err := values.WrapMap(map[string]any{"hello": "world"})
	require.NoError(t, err)

	var rtc capabilities.RemoteTriggerConfig
	rtc.ApplyDefaults()
	var rec capabilities.RemoteExecutableConfig
	rec.ApplyDefaults()
	expectedCapConfig := capabilities.CapabilityConfiguration{
		DefaultConfig:       wm,
		RemoteTriggerConfig: &rtc,
		CapabilityMethodConfig: map[string]capabilities.CapabilityMethodConfig{
			"trigger-method": {
				RemoteTriggerConfig: &rtc,
				AggregatorConfig:    &capabilities.AggregatorConfig{AggregatorType: capabilities.AggregatorType_SignedReport},
			},
			"executable-method": {
				RemoteExecutableConfig: &rec,
			},
		},
		LocalOnly: true,
	}
	reg.On("ConfigForCapability", mock.Anything, capID, donID).Once().Return(expectedCapConfig, nil)

	capConf, err := rc.ConfigForCapability(t.Context(), capID, donID)
	require.NoError(t, err)
	assert.Equal(t, expectedCapConfig, capConf)
}

func TestCapabilitiesRegistry_ConfigForCapability_RemoteExecutableConfig(t *testing.T) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	reg := mocks.NewCapabilitiesRegistry(t)

	pluginName := "registry-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRegistryPlugin{
				impl: reg,
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: logger,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := regClient.(*capabilitiesRegistryClient)
	require.True(t, ok)

	capID := "some-cap@1.0.0"
	donID := uint32(1)
	wm, err := values.WrapMap(map[string]any{"hello": "world"})
	require.NoError(t, err)

	var rec capabilities.RemoteExecutableConfig
	rec.ApplyDefaults()
	expectedCapConfig := capabilities.CapabilityConfiguration{
		DefaultConfig:          wm,
		RemoteExecutableConfig: &rec,
	}
	reg.On("ConfigForCapability", mock.Anything, capID, donID).Once().Return(expectedCapConfig, nil)

	capConf, err := rc.ConfigForCapability(t.Context(), capID, donID)
	require.NoError(t, err)
	assert.Equal(t, expectedCapConfig, capConf)
}

func TestCapabilitiesRegistry_DONsForCapability(t *testing.T) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	reg := mocks.NewCapabilitiesRegistry(t)

	pluginName := "registry-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRegistryPlugin{
				impl: reg,
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: logger,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := regClient.(*capabilitiesRegistryClient)
	require.True(t, ok)

	capID := "some-cap@1.0.0"

	donID := uint32(1)
	expectedDON := capabilities.DON{
		ID: donID,
		F:  1,
		Members: []p2ptypes.PeerID{
			[32]byte{0: 1},
			[32]byte{0: 2},
		},
	}
	expectedNodes := []capabilities.Node{
		{
			PeerID:              &p2ptypes.PeerID{0: 1},
			NodeOperatorID:      1,
			EncryptionPublicKey: [32]byte{0: 1},
			CapabilityDONs:      []capabilities.DON{},
		},
		{
			PeerID:              &p2ptypes.PeerID{0: 2},
			NodeOperatorID:      2,
			EncryptionPublicKey: [32]byte{0: 2},
			CapabilityDONs:      []capabilities.DON{},
		},
	}
	expectedDONs := []capabilities.DONWithNodes{
		{
			DON:   expectedDON,
			Nodes: expectedNodes,
		},
	}
	reg.On("DONsForCapability", mock.Anything, capID).Once().Return(expectedDONs, nil)

	dons, err := rc.DONsForCapability(t.Context(), capID)
	require.NoError(t, err)
	assert.Equal(t, expectedDONs, dons)
}

func ensureEqual(t *testing.T, expectedNode, actualNode capabilities.Node) {
	// check local node struct
	require.Equal(t, expectedNode.PeerID, actualNode.PeerID)

	// check workflow DON
	require.Len(t, expectedNode.WorkflowDON.Members, len(actualNode.WorkflowDON.Members))
	require.ElementsMatch(t, expectedNode.WorkflowDON.Members, actualNode.WorkflowDON.Members)
	require.Equal(t, expectedNode.WorkflowDON.ID, actualNode.WorkflowDON.ID)
	require.Equal(t, expectedNode.WorkflowDON.F, actualNode.WorkflowDON.F)
	require.Equal(t, expectedNode.WorkflowDON.ConfigVersion, actualNode.WorkflowDON.ConfigVersion)
	require.Equal(t, expectedNode.WorkflowDON.Name, actualNode.WorkflowDON.Name)
	require.Equal(t, expectedNode.WorkflowDON.Families, actualNode.WorkflowDON.Families)
	require.Equal(t, expectedNode.WorkflowDON.Config, actualNode.WorkflowDON.Config)

	// check capability DONs
	require.Len(t, expectedNode.CapabilityDONs, len(actualNode.CapabilityDONs))
	for i := range expectedNode.CapabilityDONs {
		require.Equal(t, expectedNode.CapabilityDONs[i].ID, actualNode.CapabilityDONs[i].ID)
		require.Len(t, expectedNode.CapabilityDONs[i].Members, len(actualNode.CapabilityDONs[i].Members))
		require.ElementsMatch(t, expectedNode.CapabilityDONs[i].Members, actualNode.CapabilityDONs[i].Members)
		require.Equal(t, expectedNode.CapabilityDONs[i].F, actualNode.CapabilityDONs[i].F)
		require.Equal(t, expectedNode.CapabilityDONs[i].ConfigVersion, actualNode.CapabilityDONs[i].ConfigVersion)
		require.Equal(t, expectedNode.CapabilityDONs[i].Name, actualNode.CapabilityDONs[i].Name)
		require.Equal(t, expectedNode.CapabilityDONs[i].Families, actualNode.CapabilityDONs[i].Families)
		require.Equal(t, expectedNode.CapabilityDONs[i].Config, actualNode.CapabilityDONs[i].Config)
	}
}
