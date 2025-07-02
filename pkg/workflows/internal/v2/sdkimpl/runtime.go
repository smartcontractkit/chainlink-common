package sdkimpl

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/consensus"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"google.golang.org/protobuf/proto"
)

type RuntimeHelpers interface {
	Call(request *pb.CapabilityRequest) error
	Await(request *pb.AwaitCapabilitiesRequest, maxResponseSize uint64) (*pb.AwaitCapabilitiesResponse, error)

	GetSecrets(request *pb.GetSecretsRequest, maxResponseSize uint64) error
	AwaitSecrets(request *pb.AwaitSecretsRequest, maxResponseSize uint64) (*pb.AwaitSecretsResponse, error)

	SwitchModes(mode pb.Mode)
	GetSource(mode pb.Mode) rand.Source
}

type RuntimeBase struct {
	MaxResponseSize uint64
	RuntimeHelpers

	source   rand.Source
	source64 rand.Source64
	modeErr  error
	Mode     pb.Mode

	// nextCallId tracks the unique id for a request to the WASM host.
	// to avoid collisions of the ID in different modes, it is
	// incremented in DON mode and decremented in Node mode.
	// eg. - first call don mode: nextCallId = 1
	//     - second call: nextCallId = 2
	//     - first call node mode: nextCallId = -1
	//     - second call node mode: nextCallId = -2
	//     - etc...
	nextCallId int32
}

var (
	_ sdk.RuntimeBase = (*RuntimeBase)(nil)
	_ rand.Source     = (*RuntimeBase)(nil)
	_ rand.Source64   = (*RuntimeBase)(nil)
)

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

func (d *Runtime) GetSecret(req *pb.SecretRequest) sdk.Promise[*pb.Secret] {
	d.nextCallId++

	sr := &pb.GetSecretsRequest{
		Requests:   []*pb.SecretRequest{req},
		CallbackId: d.nextCallId,
	}

	err := d.RuntimeHelpers.GetSecrets(sr, d.MaxResponseSize)
	if err != nil {
		return sdk.PromiseFromResult[*pb.Secret](nil, err)
	}

	return sdk.NewBasicPromise(func() (*pb.Secret, error) {
		awaitRequest := &pb.AwaitSecretsRequest{
			Ids: []int32{d.nextCallId},
		}
		awaitResponse, err := d.AwaitSecrets(awaitRequest, d.MaxResponseSize)
		if err != nil {
			return nil, err
		}

		resp, ok := awaitResponse.Responses[d.nextCallId]
		if !ok {
			return nil, fmt.Errorf("cannot find response for %d", d.nextCallId)
		}

		if len(resp.Responses) != 1 {
			return nil, fmt.Errorf("expected 1 response, got %d", len(resp.Responses))
		}

		if resp.Responses[0].GetError() != nil {
			return nil, fmt.Errorf("error getting secret %s: %s", req.Id, resp.Responses[0].GetError().Error)
		}

		return resp.Responses[0].GetSecret(), nil
	})
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
	return sdk.Then(c.Simple(d, observation), func(result *pb.ConsensusOutputs) (values.Value, error) {
		var mapProto valuespb.Map
		err := proto.Unmarshal(result.RawReport, &mapProto)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw report: %w", err)
		}
		payload, ok := mapProto.Fields[sdk.ConsensusResponseMapKeyPayload]
		if !ok || payload == nil {
			return nil, fmt.Errorf("missing payload in consensus response")
		}
		return values.FromProto(payload)
	})
}

func (d *Runtime) GenerateReport(
	encodedPayload []byte,
	encoderName, signingAlgo, hashingAlgo string,
) sdk.Promise[*pb.ConsensusOutputs] {
	d.nextCallId++ // Question: do I increment here?

	observation := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Value{
			Value: &valuespb.Value{
				Value: &valuespb.Value_BytesValue{
					BytesValue: encodedPayload,
				},
			},
		},
		Descriptors: &pb.ConsensusDescriptor{
			EncoderName: encoderName,
			SigningAlgo: signingAlgo,
			HashingAlgo: hashingAlgo,
		},
	}
	c := &consensus.Consensus{}

	return sdk.Then(c.Simple(d, observation), func(result *pb.ConsensusOutputs) (*pb.ConsensusOutputs, error) {
		return result, errors.New("not implemented")
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
