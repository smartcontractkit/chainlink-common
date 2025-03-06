package legacy

import (
	"errors"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/legacy"
	wpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runtimeBase struct {
}

//go:wasmimport env call_capability
func callCapability(reqptr unsafe.Pointer, reqptrlen int32, id unsafe.Pointer) int32

//go:wasmimport env await_capabilities
func awaitCapabilities(id unsafe.Pointer, respptr unsafe.Pointer, resplen int32) int32

func (r *runtimeBase) CallCapability(request *wpb.CapabilityRequest) sdk.Promise[*wpb.CapabilityResponse] {
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return sdk.PromiseFromResult[*wpb.CapabilityResponse](nil, err)
	}

	id := make([]byte, sdk.IdLen)
	result := callCapability(unsafe.Pointer(&marshalled[0]), int32(len(marshalled)), unsafe.Pointer(&id[0]))
	if result == -1 {
		return sdk.PromiseFromResult[*wpb.CapabilityResponse](nil, errors.New("cannot find capability "+request.Id))
	}

	return sdk.NewBasicPromise(func() (*wpb.CapabilityResponse, error) {
		// TODO make this configurable?
		response := make([]byte, 2048)
		bytes := awaitCapabilities(unsafe.Pointer(&id[0]), unsafe.Pointer(&response[0]), int32(len(response)))
		capResponse := &wpb.CapabilityResponse{}
		err := proto.Unmarshal(response[:bytes], capResponse)
		return capResponse, err
	})
}

type donRuntime struct {
	runtimeBase
}

func (d donRuntime) RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime sdk.NodeRuntime) *wpb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	changeMode(int32(wpb.Mode_NODE))
	observation := fn(&nodeRuntime{})
	wrapped, _ := anypb.New(observation)

	changeMode(int32(wpb.Mode_DON))
	// In real life, the payload can be different than
	capabilityRequest := &wpb.CapabilityRequest{
		Id:      "consensus@1.0.0",
		Payload: wrapped,
	}
	response := d.CallCapability(capabilityRequest)
	return sdk.Then(response, func(resp *wpb.CapabilityResponse) (values.Value, error) {
		if p := resp.GetPayload(); p != nil {
			pbVal := pb.Value{}
			if err := proto.Unmarshal(p.Value, &pbVal); err != nil {
				return nil, err
			}

			return values.FromProto(&pbVal)
		}

		return nil, errors.New(resp.GetError())
	})
}

var _ sdk.DonRuntime = &donRuntime{}

type nodeRuntime struct {
	runtimeBase
}

var _ sdk.NodeRuntime = &nodeRuntime{}

func (n *nodeRuntime) IsNodeRuntime() {}
