package capability

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
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
	callback chan capabilities.CapabilityResponse
}

func (f *mockTriggerExecutable) XXXTestingPushToCallbackChan(cr capabilities.CapabilityResponse) {
	f.callback <- cr
}

func (f *mockTriggerExecutable) RegisterTrigger(ctx context.Context, request capabilities.CapabilityRequest) (<-chan capabilities.CapabilityResponse, error) {
	return f.callback, nil
}

func (f *mockTriggerExecutable) UnregisterTrigger(ctx context.Context, request capabilities.CapabilityRequest) error {
	f.callback = nil
	return nil
}

var _ capabilities.CallbackExecutable = (*mockCallbackExecutable)(nil)

type mockCallbackExecutable struct {
	registeredWorkflowRequest *capabilities.RegisterToWorkflowRequest
	callback                  chan capabilities.CapabilityResponse
}

func (f *mockCallbackExecutable) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	f.registeredWorkflowRequest = &request
	return nil
}

func (f *mockCallbackExecutable) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	f.registeredWorkflowRequest = nil
	return nil
}

func (f *mockCallbackExecutable) Execute(ctx context.Context, request capabilities.CapabilityRequest) (<-chan capabilities.CapabilityResponse, error) {
	f.callback <- capabilities.CapabilityResponse{
		Value: nil,
		Err:   errors.New("some-error"),
	}
	return f.callback, nil
}

var _ capabilities.TriggerCapability = (*mockTriggerCapability)(nil)

type mockTriggerCapability struct {
	*mockBaseCapability
	*mockTriggerExecutable
}

type mockActionCapability struct {
	*mockBaseCapability
	*mockCallbackExecutable
}

type mockConsensusCapability struct {
	*mockBaseCapability
	*mockCallbackExecutable
}

type mockTargetCapability struct {
	*mockBaseCapability
	*mockCallbackExecutable
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
		Err:   errors.New("some-error"),
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
	_, err = rc.Get(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")

	pid := p2ptypes.PeerID([32]byte{0: 1})
	expectedNode := capabilities.Node{
		PeerID: &pid,
		WorkflowDON: capabilities.DON{
			ID: 11,
			Members: []p2ptypes.PeerID{
				[32]byte{0: 2},
				[32]byte{0: 3},
			},
			F:             2,
			ConfigVersion: 1,
		},
		CapabilityDONs: []capabilities.DON{
			{
				ID:            22,
				Members:       []p2ptypes.PeerID{},
				F:             1,
				ConfigVersion: 2,
			},
			{
				ID: 33,
				Members: []p2ptypes.PeerID{
					[32]byte{0: 4},
					[32]byte{0: 5},
					[32]byte{0: 6},
				},
				F:             3,
				ConfigVersion: 3,
			},
		},
	}

	reg.On("GetLocalNode", mock.Anything).Once().Return(expectedNode, nil)

	actualNode, err := rc.GetLocalNode(tests.Context(t))
	require.NoError(t, err)
	// check local node struct
	require.Equal(t, expectedNode.PeerID, actualNode.PeerID)

	// check workflow DON
	require.Len(t, expectedNode.WorkflowDON.Members, len(actualNode.WorkflowDON.Members))
	require.ElementsMatch(t, expectedNode.WorkflowDON.Members, actualNode.WorkflowDON.Members)
	require.Equal(t, expectedNode.WorkflowDON.ID, actualNode.WorkflowDON.ID)
	require.Equal(t, expectedNode.WorkflowDON.F, actualNode.WorkflowDON.F)
	require.Equal(t, expectedNode.WorkflowDON.ConfigVersion, actualNode.WorkflowDON.ConfigVersion)

	// check capability DONs
	require.Len(t, expectedNode.CapabilityDONs, len(actualNode.CapabilityDONs))
	for i := range expectedNode.CapabilityDONs {
		require.Equal(t, expectedNode.CapabilityDONs[i].ID, actualNode.CapabilityDONs[i].ID)
		require.Len(t, expectedNode.CapabilityDONs[i].Members, len(actualNode.CapabilityDONs[i].Members))
		require.ElementsMatch(t, expectedNode.CapabilityDONs[i].Members, actualNode.CapabilityDONs[i].Members)
		require.Equal(t, expectedNode.CapabilityDONs[i].F, actualNode.CapabilityDONs[i].F)
		require.Equal(t, expectedNode.CapabilityDONs[i].ConfigVersion, actualNode.CapabilityDONs[i].ConfigVersion)
	}

	// Check zero values for empty node
	emptyNode := capabilities.Node{}
	reg.On("GetLocalNode", mock.Anything).Once().Return(emptyNode, nil)
	actualNode, err = rc.GetLocalNode(tests.Context(t))
	require.NoError(t, err)
	require.Nil(t, actualNode.PeerID)
	require.Equal(t, capabilities.DON{
		ID:            0,
		Members:       nil,
		F:             0,
		ConfigVersion: 0,
	}, actualNode.WorkflowDON)
	require.Empty(t, actualNode.CapabilityDONs)

	reg.On("GetAction", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.GetAction(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")

	reg.On("GetConsensus", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.GetConsensus(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")

	reg.On("GetTarget", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.GetTarget(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")

	reg.On("GetTrigger", mock.Anything, "some-id").Return(nil, errors.New("capability not found"))
	_, err = rc.GetTrigger(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")

	reg.On("List", mock.Anything).Return([]capabilities.BaseCapability{}, nil)
	list, err := rc.List(tests.Context(t))
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
		mockTriggerExecutable: &mockTriggerExecutable{callback: make(chan capabilities.CapabilityResponse, 10)},
	}

	// After adding the trigger, we'll expect something wrapped by the internal client type below.
	reg.On("Add", mock.Anything, mock.AnythingOfType("*capability.TriggerCapabilityClient")).Return(nil)
	err = rc.Add(tests.Context(t), testTrigger)
	require.NoError(t, err)

	reg.On("GetTrigger", mock.Anything, "trigger-1@1.0.0").Return(testTrigger, nil)
	triggerCap, err := rc.GetTrigger(tests.Context(t), "trigger-1@1.0.0")
	require.NoError(t, err)

	// Test trigger Info()
	testCapabilityInfo(t, triggerInfo, triggerCap)

	// Test TriggerExecutable
	callbackChan, err := triggerCap.RegisterTrigger(tests.Context(t), capabilities.CapabilityRequest{
		Inputs: &values.Map{},
		Config: &values.Map{},
	})
	require.NoError(t, err)

	testTrigger.XXXTestingPushToCallbackChan(capabilityResponse)
	require.Equal(t, capabilityResponse, <-callbackChan)

	err = triggerCap.UnregisterTrigger(tests.Context(t), capabilities.CapabilityRequest{
		Inputs: &values.Map{},
		Config: &values.Map{},
	})
	require.NoError(t, err)
	require.Nil(t, testTrigger.callback)

	// Add capability Trigger
	actionInfo := capabilities.CapabilityInfo{
		ID:             "action-1@2.0.0",
		CapabilityType: capabilities.CapabilityTypeAction,
		Description:    "action-1-description",
	}

	actionCallbackChan := make(chan capabilities.CapabilityResponse, 10)
	testAction := mockActionCapability{
		mockBaseCapability:     &mockBaseCapability{info: actionInfo},
		mockCallbackExecutable: &mockCallbackExecutable{callback: actionCallbackChan},
	}
	reg.On("GetAction", mock.Anything, "action-1@2.0.0").Return(testAction, nil)
	actionCap, err := rc.GetAction(tests.Context(t), "action-1@2.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, actionInfo, actionCap)

	// Test Executable
	workflowRequest := capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: "workflow-ID",
		},
	}
	err = actionCap.RegisterToWorkflow(tests.Context(t), workflowRequest)
	require.NoError(t, err)
	require.Equal(t, workflowRequest.Metadata.WorkflowID, testAction.registeredWorkflowRequest.Metadata.WorkflowID)

	actionCallbackChan <- capabilityResponse
	callbackChan, err = actionCap.Execute(tests.Context(t), capabilities.CapabilityRequest{})
	require.NoError(t, err)
	require.Equal(t, capabilityResponse, <-callbackChan)
	err = actionCap.UnregisterFromWorkflow(tests.Context(t), capabilities.UnregisterFromWorkflowRequest{})
	require.NoError(t, err)
	require.Nil(t, testAction.registeredWorkflowRequest)

	// Add capability Consensus
	consensusInfo := capabilities.CapabilityInfo{
		ID:             "consensus-1@3.0.0",
		CapabilityType: capabilities.CapabilityTypeConsensus,
		Description:    "consensus-1-description",
	}
	testConsensus := mockConsensusCapability{
		mockBaseCapability:     &mockBaseCapability{info: consensusInfo},
		mockCallbackExecutable: &mockCallbackExecutable{},
	}
	reg.On("GetConsensus", mock.Anything, "consensus-1@3.0.0").Return(testConsensus, nil)
	consensusCap, err := rc.GetConsensus(tests.Context(t), "consensus-1@3.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, consensusInfo, consensusCap)

	// Add capability Target
	targetInfo := capabilities.CapabilityInfo{
		ID:             "target-1@1.0.0",
		CapabilityType: capabilities.CapabilityTypeTarget,
		Description:    "target-1-description",
	}
	testTarget := mockTargetCapability{
		mockBaseCapability:     &mockBaseCapability{info: targetInfo},
		mockCallbackExecutable: &mockCallbackExecutable{},
	}
	reg.On("GetTarget", mock.Anything, "target-1@1.0.0").Return(testTarget, nil)
	targetCap, err := rc.GetTarget(tests.Context(t), "target-1@1.0.0")
	require.NoError(t, err)

	testCapabilityInfo(t, targetInfo, targetCap)
}

func testCapabilityInfo(t *testing.T, expectedInfo capabilities.CapabilityInfo, cap capabilities.BaseCapability) {
	gotInfo, err := cap.Info(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, expectedInfo.ID, gotInfo.ID)
	require.Equal(t, expectedInfo.CapabilityType, gotInfo.CapabilityType)
	require.Equal(t, expectedInfo.Description, gotInfo.Description)
	require.Equal(t, expectedInfo.Version(), gotInfo.Version())
}
func TestToDON(t *testing.T) {
	don := &pb.DON{
		Id: 0,
		Members: [][]byte{
			{0: 4, 31: 0},
			{0: 5, 31: 0},
		},
		F:             2,
		ConfigVersion: 1,
	}

	expected := capabilities.DON{
		ID: 0,
		Members: []p2ptypes.PeerID{
			[32]byte{0: 4},
			[32]byte{0: 5},
		},
		F:             2,
		ConfigVersion: 1,
	}

	actual := toDON(don)

	require.Equal(t, expected, actual)
}

func TestCapabilitiesRegistry_ConfigForCapabilities(t *testing.T) {
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
	expectedCapConfig := capabilities.CapabilityConfiguration{
		ExecuteConfig:       wm,
		RemoteTriggerConfig: rtc,
	}
	reg.On("ConfigForCapability", mock.Anything, capID, donID).Once().Return(expectedCapConfig, nil)

	capConf, err := rc.ConfigForCapability(tests.Context(t), capID, donID)
	require.NoError(t, err)
	assert.Equal(t, expectedCapConfig, capConf)
}
