package wasm

import (
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type RunnerV2 struct {
	sendResponse   func(payload *wasmpb.Response)
	runtimeFactory func(sdkConfig *RuntimeConfig, refToResponse map[string]capabilities.CapabilityResponse, hostReqID string) *RuntimeV2
	args           []string
	req            *wasmpb.Request
}

var _ sdk.Runner = (*RunnerV2)(nil)

func (r *RunnerV2) Run(factory *sdk.WorkflowSpecFactory) {
	if r.req == nil {
		success := r.cacheRequest()
		if !success {
			return
		}
	}

	req := r.req

	// We set this up *after* parsing the request, so that we can guarantee
	// that we'll have access to the request object.
	defer func() {
		if err := recover(); err != nil {
			asErr, ok := err.(error)
			if ok {
				r.sendResponse(errorResponse(r.req.Id, asErr))
			} else {
				r.sendResponse(errorResponse(r.req.Id, fmt.Errorf("caught panic: %+v", err)))
			}
		}
	}()

	resp := &wasmpb.Response{
		Id: req.Id,
	}

	switch {
	case req.GetSpecRequest() != nil:
		rsp, innerErr := r.handleSpecRequest(factory, req.Id)
		if innerErr != nil {
			resp.ErrMsg = innerErr.Error()
		} else {
			resp = rsp
		}
	case req.GetRunRequest() != nil:
		rsp, innerErr := r.handleRunRequest(factory, req.Id, req.GetRunRequest())
		if innerErr != nil {
			resp.ErrMsg = innerErr.Error()
		} else {
			resp = rsp // should happen only when workflow is done processing (i.e. no more capability calls)
		}
	default:
		resp.ErrMsg = "invalid request: message must be SpecRequest or RunRequest"
	}

	r.sendResponse(resp)
}

func (r *RunnerV2) Config() []byte {
	if r.req == nil {
		success := r.cacheRequest()
		if !success {
			return nil
		}
	}

	return r.req.Config
}

func (r *RunnerV2) ExitWithError(err error) {
	if r.req == nil {
		success := r.cacheRequest()
		if !success {
			return
		}
	}

	r.sendResponse(errorResponse(r.req.Id, err))
	return
}

func (r *RunnerV2) cacheRequest() bool {
	if r.req == nil {
		req, err := r.parseRequest()
		if err != nil {
			r.sendResponse(errorResponse(unknownID, err))
			return false
		}

		r.req = req
	}
	return true
}

func (r *RunnerV2) parseRequest() (*wasmpb.Request, error) {
	// We expect exactly 2 args, i.e. `wasm <blob>`,
	// where <blob> is a base64 encoded protobuf message.
	if len(r.args) != 2 {
		return nil, errors.New("invalid request: request must contain a payload")
	}

	request := r.args[1]
	if request == "" {
		return nil, errors.New("invalid request: request cannot be empty")
	}

	b, err := base64.StdEncoding.DecodeString(request)
	if err != nil {
		return nil, fmt.Errorf("invalid request: could not decode request into bytes")
	}

	req := &wasmpb.Request{}
	err = proto.Unmarshal(b, req)
	if err != nil {
		return nil, fmt.Errorf("invalid request: could not unmarshal proto: %w", err)
	}
	return req, err
}

func (r *RunnerV2) handleSpecRequest(factory *sdk.WorkflowSpecFactory, id string) (*wasmpb.Response, error) {
	spec, err := factory.Spec()
	if err != nil {
		return nil, fmt.Errorf("error getting spec from factory: %w", err)
	}

	specpb, err := wasmpb.WorkflowSpecToProto(&spec)
	if err != nil {
		return nil, fmt.Errorf("failed to translate workflow spec to proto: %w", err)
	}

	return &wasmpb.Response{
		Id: id,
		Message: &wasmpb.Response_SpecResponse{
			SpecResponse: specpb,
		},
	}, nil
}

func (r *RunnerV2) handleRunRequest(factory *sdk.WorkflowSpecFactory, id string, runReq *wasmpb.RunRequest) (*wasmpb.Response, error) {
	// Extract config from the request
	drc := defaultRuntimeConfig(id, nil)

	refToResponse := map[string]capabilities.CapabilityResponse{}
	for ref, resp := range runReq.RefToResponse {
		unmarshalled, err := pb.CapabilityResponseFromProto(resp)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling capability response: %w", err)
		}
		refToResponse[ref] = unmarshalled
	}

	if runReq.TriggerEvent == nil {
		return nil, errors.New("missing trigger event")
	}

	var event capabilities.TriggerEvent
	event.TriggerType = runReq.TriggerEvent.TriggerType
	event.ID = runReq.TriggerEvent.Id
	outputs, err := values.FromMapValueProto(runReq.TriggerEvent.Outputs)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal event payload: %w", err)
	}
	event.Outputs = outputs

	// execute workflow
	runtime := r.runtimeFactory(drc, refToResponse, r.req.Id)
	runFn := factory.GetRunFn(runReq.TriggerRef)
	if runFn == nil {
		return nil, fmt.Errorf("could not find run function for ref %s", runReq.TriggerRef)
	}
	err = runFn(runtime, event)
	if err != nil {
		return nil, fmt.Errorf("error executing workflow: %w", err)
	}

	// successful execution termination
	return &wasmpb.Response{
		Id: id,
	}, nil
}
