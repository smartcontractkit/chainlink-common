package sdkimpl

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/consensus"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type CallCapabilityFn func(request *pb.CapabilityRequest) error
type AwaitCapabilitiesFn func(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error)

type RuntimeBase struct {
	ConfigBytes     []byte
	MaxResponseSize uint64
	Call            CallCapabilityFn
	Await           AwaitCapabilitiesFn
	Writer          io.Writer

	modeErr    error
	Mode       pb.Mode
	nextCallId int32
}

func (r *RuntimeBase) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	if r.Mode == pb.Mode_DON {
		r.nextCallId++
	} else {
		r.nextCallId--
	}

	myId := r.nextCallId
	request.CallbackId = myId
	if r.modeErr != nil {
		return sdk.PromiseFromResult[*pb.CapabilityResponse](nil, r.modeErr)
	}

	err := r.Call(request)
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

func (r *RuntimeBase) Config() []byte {
	return r.ConfigBytes
}

func (r *RuntimeBase) LogWriter() io.Writer {
	return r.Writer
}

func (r *RuntimeBase) Logger() *slog.Logger {
	return slog.New(slog.NewTextHandler(r.LogWriter(), nil))
}

type DonRuntime struct {
	RuntimeBase
	nextNodeCallId int32
}

func (d *DonRuntime) RunInNodeMode(fn func(nodeRuntime sdk.NodeRuntime) *pb.SimpleConsensusInputs) sdk.Promise[values.Value] {
	nrt := &NodeRuntime{RuntimeBase: d.RuntimeBase}
	nrt.nextCallId = d.nextNodeCallId
	nrt.Mode = pb.Mode_Node
	d.modeErr = sdk.DonModeCallInNodeMode()
	observation := fn(nrt)
	nrt.modeErr = sdk.NodeModeCallInDonMode()
	d.modeErr = nil
	d.nextNodeCallId = nrt.nextCallId
	c := &consensus.Consensus{}
	return sdk.Then(c.Simple(d, observation), func(result *valuespb.Value) (values.Value, error) {
		return values.FromProto(result)
	})
}

var _ sdk.DonRuntime = &DonRuntime{}

type NodeRuntime struct {
	RuntimeBase
}

var _ sdk.NodeRuntime = &NodeRuntime{}

func (n *NodeRuntime) IsNodeRuntime() {}
