package sdkimpl

import (
	"fmt"
	"math/rand"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/consensus"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type RuntimeHelpers interface {
	Call(request *pb.CapabilityRequest) error
	Await(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error)
	SwitchModes(mode pb.Mode)
	GetSource(mode pb.Mode) rand.Source
}

type RuntimeBase struct {
	MaxResponseSize uint64
	RuntimeHelpers

	source     rand.Source
	source64   rand.Source64
	modeErr    error
	Mode       pb.Mode
	nextCallId int32
}

var _ sdk.RuntimeBase = (*RuntimeBase)(nil)
var _ rand.Source = (*RuntimeBase)(nil)
var _ rand.Source64 = (*RuntimeBase)(nil)

func (r *RuntimeBase) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	if r.Mode == pb.Mode_MODE_DON {
		r.nextCallId++
	} else {
		r.nextCallId--
	}

	myId := r.nextCallId
	request.CallbackId = myId
	if r.modeErr != nil {
		return sdk.PromiseFromResult[*pb.CapabilityResponse](nil, r.modeErr)
	}

	err := r.RuntimeHelpers.Call(request)
	if err != nil {
		return sdk.PromiseFromResult[*pb.CapabilityResponse](nil, err)
	}

	return sdk.NewBasicPromise(func() (*pb.CapabilityResponse, error) {
		awaitRequest := &pb.AwaitCapabilitiesRequest{
			Ids: []int32{myId},
		}
		awaitResponse, err := r.Await(awaitRequest, r.MaxResponseSize)
		if err != nil {
			return nil, err
		}

		capResponse, ok := awaitResponse.Responses[myId]
		if !ok {
			return nil, fmt.Errorf("cannot find response for %d", myId)
		}

		return capResponse, err
	})
}

func (r *RuntimeBase) Rand() (*rand.Rand, error) {
	if r.modeErr != nil {
		return nil, r.modeErr
	}

	if r.source == nil {
		r.source = r.RuntimeHelpers.GetSource(r.Mode)
		r64, ok := r.source.(rand.Source64)
		if ok {
			r.source64 = r64
		}
	}

	return rand.New(r), nil
}

type Runtime struct {
	RuntimeBase
	nextNodeCallId int32
}

func (d *Runtime) RunInNodeMode(fn func(nodeRuntime sdk.NodeRuntime) *pb.SimpleConsensusInputs) sdk.Promise[values.Value] {
	nrt := &NodeRuntime{RuntimeBase: d.RuntimeBase}
	nrt.nextCallId = d.nextNodeCallId
	nrt.Mode = pb.Mode_MODE_NODE
	d.modeErr = sdk.DonModeCallInNodeMode()
	d.SwitchModes(pb.Mode_MODE_NODE)
	observation := fn(nrt)
	d.SwitchModes(pb.Mode_MODE_DON)
	nrt.modeErr = sdk.NodeModeCallInDonMode()
	d.modeErr = nil
	d.nextNodeCallId = nrt.nextCallId
	c := &consensus.Consensus{}
	return sdk.Then(c.Simple(d, observation), func(result *valuespb.Value) (values.Value, error) {
		return values.FromProto(result)
	})
}

var _ sdk.Runtime = &Runtime{}

func (r *RuntimeBase) Int63() int64 {
	if r.modeErr != nil {
		panic("random cannot be used outside the mode it was created in")
	}

	return r.source.Int63()
}

func (r *RuntimeBase) Uint64() uint64 {
	if r.modeErr != nil {
		panic("random cannot be used outside the mode it was created in")
	}

	// borrowed from math/rand
	if r.source64 != nil {
		return r.source64.Uint64()
	}

	return uint64(r.source.Int63())>>31 | uint64(r.source.Int63())<<32
}

func (r *RuntimeBase) Seed(seed int64) {
	if r.modeErr != nil {
		panic("random cannot be used outside the mode it was created in")
	}

	r.source.Seed(seed)
}

type NodeRuntime struct {
	RuntimeBase
}

var _ sdk.NodeRuntime = &NodeRuntime{}

func (n *NodeRuntime) IsNodeRuntime() {}
