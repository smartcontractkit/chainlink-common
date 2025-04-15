package wasm

import (
	"errors"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	sdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	wpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type runtimeBase struct {
	execId string
	config []byte
}

//go:wasmimport env call_capability
func callCapability(req unsafe.Pointer, reqLen int32, id unsafe.Pointer) int64

//go:wasmimport env await_capabilities
func awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

func (r *runtimeBase) CallCapability(request *wpb.CapabilityRequest) sdk.Promise[*wpb.CapabilityResponse] {
	request.ExecutionId = r.execId
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return sdk.PromiseFromResult[*wpb.CapabilityResponse](nil, err)
	}

	id := make([]byte, sdk.IdLen)
	result := callCapability(unsafe.Pointer(&marshalled[0]), int32(len(marshalled)), unsafe.Pointer(&id[0]))
	if result < 0 {
		return sdk.PromiseFromResult[*wpb.CapabilityResponse](nil, errors.New("cannot find capability "+request.Id))
	}

	return sdk.NewBasicPromise(func() (*wpb.CapabilityResponse, error) {
		awaitRequest := &wpb.AwaitCapabilitiesRequest{
			ExecId: r.execId,
			Ids:    []string{string(id)},
		}
		m, err := proto.Marshal(awaitRequest)
		if err != nil {
			return nil, err
		}

		// TODO make this configurable?
		response := make([]byte, 2048)
		bytes := awaitCapabilities(unsafe.Pointer(&m[0]), int32(len(m)), unsafe.Pointer(&response[0]), int32(len(response)))
		if bytes < 0 {
			return nil, errors.New(string(response[:-bytes]))
		}

		awaitResponse := &wpb.AwaitCapabilitiesResponse{}
		err = proto.Unmarshal(response[:bytes], awaitResponse)
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

func (r *runtimeBase) Config() []byte {
	return r.config
}

type donRuntime struct {
	runtimeBase
}

func (d *donRuntime) RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime sdk.NodeRuntime) *wpb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	// TODO verify DON runtime isn't used inside node mode :)
	observation := fn(&nodeRuntime{})
	wrapped, _ := anypb.New(observation)

	// In real life, the payload can be different than
	capabilityRequest := &wpb.CapabilityRequest{
		ExecutionId: d.execId,
		Id:          "consensus@1.0.0",
		Payload:     wrapped,
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
