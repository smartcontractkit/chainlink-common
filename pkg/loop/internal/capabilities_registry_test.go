package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type fakeRegistry struct {
	caps map[string]map[string]capabilities.BaseCapability
}

func (r *fakeRegistry) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	c, ok := r.caps["trigger"][ID]
	if ok {
		return c.(capabilities.TriggerCapability), nil
	}

	return nil, errors.New("capability not found")
}

func (r *fakeRegistry) GetAction(ctx context.Context, ID string) (capabilities.ActionCapability, error) {
	c, ok := r.caps["action"][ID]
	if ok {
		return c.(capabilities.ActionCapability), nil
	}

	return nil, errors.New("capability not found")
}

func (r *fakeRegistry) GetConsensus(ctx context.Context, ID string) (capabilities.ConsensusCapability, error) {
	c, ok := r.caps["consensus"][ID]
	if ok {
		return c.(capabilities.ConsensusCapability), nil
	}

	return nil, errors.New("capability not found")
}

func (r *fakeRegistry) GetTarget(ctx context.Context, ID string) (capabilities.TargetCapability, error) {
	c, ok := r.caps["target"][ID]
	if ok {
		return c.(capabilities.TargetCapability), nil
	}

	return nil, errors.New("capability not found")
}

func (r *fakeRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	var caps []capabilities.BaseCapability

	for _, capType := range r.caps {
		for _, c := range capType {
			caps = append(caps, c)
		}
	}

	return caps, nil
}

func (r *fakeRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}
	switch info.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		r.caps["trigger"][info.ID] = c
	case capabilities.CapabilityTypeAction:
		r.caps["action"][info.ID] = c
	case capabilities.CapabilityTypeConsensus:
		r.caps["consensus"][info.ID] = c
	case capabilities.CapabilityTypeTarget:
		r.caps["target"][info.ID] = c
	}
	return nil
}

func (r *fakeRegistry) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	for _, caps := range r.caps {
		c, ok := caps[id]
		if ok {
			return c, nil
		}
	}
	return nil, errors.New("capability not found")
}

var _ capabilities.BaseCapability = (*fakeBaseCapability)(nil)

type fakeBaseCapability struct {
	info capabilities.CapabilityInfo
}

func (f *fakeBaseCapability) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return f.info, nil
}

var _ capabilities.TriggerExecutable = (*fakeTriggerExecutable)(nil)

type fakeTriggerExecutable struct {
	callback chan<- capabilities.CapabilityResponse
}

func (f *fakeTriggerExecutable) XXXTestingPushToCallbackChan(cr capabilities.CapabilityResponse) {
	f.callback <- cr
}

func (f *fakeTriggerExecutable) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	f.callback = callback
	return nil
}

func (f *fakeTriggerExecutable) UnregisterTrigger(ctx context.Context, request capabilities.CapabilityRequest) error {
	f.callback = nil
	return nil
}

var _ capabilities.CallbackExecutable = (*fakeCallbackExecutable)(nil)

type fakeCallbackExecutable struct {
	registeredWorkflowRequest *capabilities.RegisterToWorkflowRequest
}

func (f *fakeCallbackExecutable) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	f.registeredWorkflowRequest = &request
	return nil
}

func (f *fakeCallbackExecutable) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	f.registeredWorkflowRequest = nil
	return nil
}

func (f *fakeCallbackExecutable) Execute(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	callback <- capabilities.CapabilityResponse{
		Value: nil,
		Err:   errors.New("some-error"),
	}
	return nil
}

var _ capabilities.TriggerCapability = (*fakeTriggerCapability)(nil)

type fakeTriggerCapability struct {
	*fakeBaseCapability
	*fakeTriggerExecutable
}

var _ capabilities.ActionCapability = (*fakeActionCapability)(nil)

type fakeActionCapability struct {
	*fakeBaseCapability
	*fakeCallbackExecutable
}

type fakeConsensusCapability struct {
	*fakeBaseCapability
	*fakeCallbackExecutable
}

type fakeTargetCapability struct {
	*fakeBaseCapability
	*fakeCallbackExecutable
}

type registryPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerExt *BrokerExt
	impl      *fakeRegistry
}

func (r *registryPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker //TODO: fix this mess
	return NewCapabilitiesRegistryClient(client, r.brokerExt), nil
}

func (r *registryPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker //TODO: fix this mess
	pb.RegisterCapabilitiesRegistryServer(server, NewCapabilitiesRegistryServer(r.brokerExt, r.impl))
	return nil
}

func Test(t *testing.T) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	reg := &fakeRegistry{
		caps: map[string]map[string]capabilities.BaseCapability{
			"trigger":   make(map[string]capabilities.BaseCapability),
			"action":    make(map[string]capabilities.BaseCapability),
			"consensus": make(map[string]capabilities.BaseCapability),
			"target":    make(map[string]capabilities.BaseCapability),
		},
	}

	capabilityResponse := capabilities.CapabilityResponse{
		Value: nil,
		Err:   errors.New("some-error"),
	}
	callbackChan := make(chan capabilities.CapabilityResponse)

	pluginName := "registry-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		map[string]plugin.Plugin{
			pluginName: &registryPlugin{
				impl: reg,
				brokerExt: &BrokerExt{
					BrokerConfig: BrokerConfig{
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

	//No capabilities in register
	_, err = rc.Get(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")
	_, err = rc.GetAction(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")
	_, err = rc.GetConsensus(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")
	_, err = rc.GetTarget(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")
	_, err = rc.GetTrigger(tests.Context(t), "some-id")
	require.ErrorContains(t, err, "capability not found")
	list, err := rc.List(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, list, []capabilities.BaseCapability(nil))

	//Add capability Trigger
	triggerInfo := capabilities.CapabilityInfo{
		ID:             "trigger-1",
		CapabilityType: capabilities.CapabilityTypeTrigger,
		Description:    "trigger-1-description",
		Version:        "trigger-1-version",
	}
	fakeTrigger := fakeTriggerCapability{
		fakeBaseCapability:    &fakeBaseCapability{info: triggerInfo},
		fakeTriggerExecutable: &fakeTriggerExecutable{},
	}
	err = rc.Add(tests.Context(t), fakeTrigger)
	require.NoError(t, err)
	triggerCap, err := rc.GetTrigger(tests.Context(t), "trigger-1")
	require.NoError(t, err)

	//Test trigger Info()
	testCapabilityInfo(t, triggerInfo, triggerCap)

	//Test TriggerExecutable
	err = triggerCap.RegisterTrigger(tests.Context(t), callbackChan, capabilities.CapabilityRequest{
		Inputs: &values.Map{},
		Config: &values.Map{},
	})
	require.NoError(t, err)

	fakeTrigger.XXXTestingPushToCallbackChan(capabilityResponse)
	require.Equal(t, capabilityResponse, <-callbackChan)

	err = triggerCap.UnregisterTrigger(tests.Context(t), capabilities.CapabilityRequest{
		Inputs: &values.Map{},
		Config: &values.Map{},
	})
	require.NoError(t, err)
	require.Nil(t, fakeTrigger.callback)

	//Add capability Trigger
	actionInfo := capabilities.CapabilityInfo{
		ID:             "action-1",
		CapabilityType: capabilities.CapabilityTypeAction,
		Description:    "action-1-description",
		Version:        "action-1-version",
	}
	fakeAction := fakeActionCapability{
		fakeBaseCapability:     &fakeBaseCapability{info: actionInfo},
		fakeCallbackExecutable: &fakeCallbackExecutable{},
	}
	err = rc.Add(tests.Context(t), fakeAction)
	require.NoError(t, err)
	actionCap, err := rc.GetAction(tests.Context(t), "action-1")
	require.NoError(t, err)

	testCapabilityInfo(t, actionInfo, actionCap)

	//Test Executable
	workflowRequest := capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: "workflow-ID",
		},
	}
	err = actionCap.RegisterToWorkflow(tests.Context(t), workflowRequest)
	require.NoError(t, err)
	require.Equal(t, workflowRequest.Metadata.WorkflowID, fakeAction.registeredWorkflowRequest.Metadata.WorkflowID)
	actionCap.Execute(tests.Context(t), callbackChan, capabilities.CapabilityRequest{})
	require.Equal(t, capabilityResponse, <-callbackChan)
	err = actionCap.UnregisterFromWorkflow(tests.Context(t), capabilities.UnregisterFromWorkflowRequest{})
	require.NoError(t, err)
	require.Nil(t, fakeAction.registeredWorkflowRequest)

	//Add capability Consensus
	consensusInfo := capabilities.CapabilityInfo{
		ID:             "consensus-1",
		CapabilityType: capabilities.CapabilityTypeConsensus,
		Description:    "consensus-1-description",
		Version:        "consensus-1-version",
	}
	fakeConsensus := fakeConsensusCapability{
		fakeBaseCapability:     &fakeBaseCapability{info: consensusInfo},
		fakeCallbackExecutable: &fakeCallbackExecutable{},
	}
	err = rc.Add(tests.Context(t), fakeConsensus)
	require.NoError(t, err)
	consensusCap, err := rc.GetConsensus(tests.Context(t), "consensus-1")
	require.NoError(t, err)

	testCapabilityInfo(t, consensusInfo, consensusCap)

	//Add capability Target
	targetInfo := capabilities.CapabilityInfo{
		ID:             "target-1",
		CapabilityType: capabilities.CapabilityTypeTarget,
		Description:    "target-1-description",
		Version:        "target-1-version",
	}
	fakeTarget := fakeTargetCapability{
		fakeBaseCapability:     &fakeBaseCapability{info: targetInfo},
		fakeCallbackExecutable: &fakeCallbackExecutable{},
	}
	err = rc.Add(tests.Context(t), fakeTarget)
	require.NoError(t, err)
	targetCap, err := rc.GetTarget(tests.Context(t), "target-1")
	require.NoError(t, err)

	testCapabilityInfo(t, targetInfo, targetCap)

	list, err = rc.List(tests.Context(t))
	require.NoError(t, err)
	require.Len(t, list, 4)

}

func testCapabilityInfo(t *testing.T, expectedInfo capabilities.CapabilityInfo, cap capabilities.BaseCapability) {
	gotInfo, err := cap.Info(tests.Context(t))
	require.NoError(t, err)
	require.Equal(t, expectedInfo.ID, gotInfo.ID)
	require.Equal(t, expectedInfo.CapabilityType, gotInfo.CapabilityType)
	require.Equal(t, expectedInfo.Description, gotInfo.Description)
	require.Equal(t, expectedInfo.Version, gotInfo.Version)
}
