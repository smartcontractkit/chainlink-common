package wasm

import (
	"errors"
	"io"
	"unsafe"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type runtimeBase struct {
	execId          string
	config          []byte
	maxResponseSize uint64
	modeErr         error
}

func (r *runtimeBase) CallCapability(request *sdkpb.CapabilityRequest) sdk.Promise[*sdkpb.CapabilityResponse] {
	if r.modeErr != nil {
		return sdk.PromiseFromResult[*sdkpb.CapabilityResponse](nil, r.modeErr)
	}

	request.ExecutionId = r.execId
	marshalled, err := proto.Marshal(request)
	if err != nil {
		return sdk.PromiseFromResult[*sdkpb.CapabilityResponse](nil, err)
	}

	id := make([]byte, sdk.IdLen)
	marshalledPtr, marshalledLen, err := bufferToPointerLen(marshalled)
	if err != nil {
		return sdk.PromiseFromResult[*sdkpb.CapabilityResponse](nil, err)
	}

	result := callCapability(marshalledPtr, marshalledLen, unsafe.Pointer(&id[0]))
	if result < 0 {
		return sdk.PromiseFromResult[*sdkpb.CapabilityResponse](nil, errors.New("cannot find capability "+request.Id))
	}

	return sdk.NewBasicPromise(func() (*sdkpb.CapabilityResponse, error) {
		awaitRequest := &pb.AwaitCapabilitiesRequest{
			ExecId: r.execId,
			Ids:    []string{string(id)},
		}
		m, err := proto.Marshal(awaitRequest)
		if err != nil {
			return nil, err
		}

		mptr, mlen, err := bufferToPointerLen(m)
		if err != nil {
			return nil, err
		}

		response := make([]byte, r.maxResponseSize)
		responsePtr, responseLen, err := bufferToPointerLen(response)
		if err != nil {
			return nil, err
		}

		bytes := awaitCapabilities(mptr, mlen, responsePtr, responseLen)
		if bytes < 0 {
			return nil, errors.New(string(response[:-bytes]))
		}

		awaitResponse := &pb.AwaitCapabilitiesResponse{}
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

func (r *runtimeBase) LogWriter() io.Writer {
	return &writer{}
}

type donRuntime struct {
	runtimeBase
}

func (d *donRuntime) RunInNodeMode(fn func(nodeRuntime sdk.NodeRuntime) *sdkpb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	nrt := &nodeRuntime{runtimeBase: d.runtimeBase}
	d.modeErr = sdk.DonModeCallInNodeMode()
	observation := fn(nrt)
	nrt.modeErr = sdk.NodeModeCallInDonMode()
	d.modeErr = nil
	wrapped, _ := anypb.New(observation)

	capabilityRequest := &sdkpb.CapabilityRequest{
		ExecutionId: d.execId,
		Id:          "consensus@1.0.0",
		Payload:     wrapped,
		Method:      "BuiltIn",
	}
	response := d.CallCapability(capabilityRequest)
	return sdk.Then(response, func(resp *sdkpb.CapabilityResponse) (values.Value, error) {
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

var _ sdk.DonRuntime = &donRuntime{}

type nodeRuntime struct {
	runtimeBase
}

var _ sdk.NodeRuntime = &nodeRuntime{}

func (n *nodeRuntime) IsNodeRuntime() {}
