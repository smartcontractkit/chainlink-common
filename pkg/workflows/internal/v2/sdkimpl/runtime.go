package sdkimpl

import (
	"errors"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type CallCapabilityFn func(request *pb.CapabilityRequest) (id []byte, err error)
type AwaitCapabilitiesFn func(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error)

type RuntimeBase struct {
	ExecId          string
	ConfigBytes     []byte
	MaxResponseSize uint64
	Call            CallCapabilityFn
	Await           AwaitCapabilitiesFn
	Writer          io.Writer

	modeErr error
}

func (r *RuntimeBase) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	if r.modeErr != nil {
		return sdk.PromiseFromResult[*pb.CapabilityResponse](nil, r.modeErr)
	}

	request.ExecutionId = r.ExecId
	id, err := r.Call(request)
	if err != nil {
		return sdk.PromiseFromResult[*pb.CapabilityResponse](nil, err)
	}

	return sdk.NewBasicPromise(func() (*pb.CapabilityResponse, error) {
		awaitRequest := &pb.AwaitCapabilitiesRequest{
			ExecId: r.ExecId,
			Ids:    []string{string(id)},
		}
		awaitResponse, err := r.Await(awaitRequest, r.MaxResponseSize)
		if err != nil {
			return nil, err
		}

		capResponse, ok := awaitResponse.Responses[string(id)]
		if !ok {
			return nil, errors.New("cannot find response for " + string(id))
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

type DonRuntime struct {
	RuntimeBase
}

func (d *DonRuntime) RunInNodeMode(fn func(nodeRuntime sdk.NodeRuntime) *pb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	nrt := &NodeRuntime{RuntimeBase: d.RuntimeBase}
	d.modeErr = sdk.DonModeCallInNodeMode()
	observation := fn(nrt)
	nrt.modeErr = sdk.NodeModeCallInDonMode()
	d.modeErr = nil
	wrapped, _ := anypb.New(observation)

	// TODO https: //smartcontract-it.atlassian.net/browse/CAPPL-816 use the generated consensus code
	capabilityRequest := &pb.CapabilityRequest{
		ExecutionId: d.ExecId,
		Id:          "consensus@1.0.0",
		Payload:     wrapped,
		Method:      "BuiltIn",
	}
	response := d.CallCapability(capabilityRequest)
	return sdk.Then(response, func(resp *pb.CapabilityResponse) (values.Value, error) {
		if p := resp.GetPayload(); p != nil {
			pbVal := valuespb.Value{}
			if err := proto.Unmarshal(p.Value, &pbVal); err != nil {
				return nil, err
			}

			return values.FromProto(&pbVal)
		}

		return nil, errors.New(resp.GetError())
	})
}

var _ sdk.DonRuntime = &DonRuntime{}

type NodeRuntime struct {
	RuntimeBase
}

var _ sdk.NodeRuntime = &NodeRuntime{}

func (n *NodeRuntime) IsNodeRuntime() {}
