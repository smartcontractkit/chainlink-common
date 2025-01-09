package wasm

import (
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	unknownID = "__UNKNOWN__"

	CodeInvalidResponse = 110
	CodeInvalidRequest  = 111
	CodeRunnerErr       = 112
	CodeHostErr         = 113
	CodeSuccess         = 0
)

var _ sdk.Runner = (*Runner)(nil)

type Runner struct {
	sendResponse func(payload *wasmpb.Response)
	sdkFactory   func(cfg *RuntimeConfig, opts ...func(*RuntimeConfig)) *Runtime
	args         []string
	req          *wasmpb.Request
}

func (r *Runner) Run(factory *sdk.WorkflowSpecFactory) {
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
	case req.GetComputeRequest() != nil:
		rsp, innerErr := r.handleComputeRequest(factory, req.Id, req.GetComputeRequest())
		if innerErr != nil {
			resp.ErrMsg = innerErr.Error()
		} else {
			resp = rsp
		}
	default:
		resp.ErrMsg = "invalid request: message must be SpecRequest or ComputeRequest"
	}

	r.sendResponse(resp)
}

func (r *Runner) Config() []byte {
	if r.req == nil {
		success := r.cacheRequest()
		if !success {
			return nil
		}
	}

	return r.req.Config
}

func (r *Runner) ExitWithError(err error) {
	if r.req == nil {
		success := r.cacheRequest()
		if !success {
			return
		}
	}

	r.sendResponse(errorResponse(r.req.Id, err))
	return
}

func errorResponse(id string, err error) *wasmpb.Response {
	return &wasmpb.Response{
		Id:     id,
		ErrMsg: err.Error(),
	}
}

func (r *Runner) cacheRequest() bool {
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

func (r *Runner) parseRequest() (*wasmpb.Request, error) {
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

func (r *Runner) handleSpecRequest(factory *sdk.WorkflowSpecFactory, id string) (*wasmpb.Response, error) {
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

func (r *Runner) handleComputeRequest(factory *sdk.WorkflowSpecFactory, id string, computeReq *wasmpb.ComputeRequest) (*wasmpb.Response, error) {
	req := computeReq.Request
	if req == nil {
		return nil, errors.New("invalid compute request: nil request")
	}

	if req.Metadata == nil {
		return nil, errors.New("invalid compute request: nil request metadata")
	}

	fn := factory.GetFn(req.Metadata.ReferenceId)
	if fn == nil {
		return nil, fmt.Errorf("invalid compute request: could not find compute function for id %s", req.Metadata.ReferenceId)
	}

	creq, err := capabilitiespb.CapabilityRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("invalid compute request: could not translate proto into capability request")
	}

	// Extract the config from the request
	drc := defaultRuntimeConfig(id, &creq.Metadata)
	if rc := computeReq.GetRuntimeConfig(); rc != nil {
		if rc.MaxFetchResponseSizeBytes != 0 {
			drc.MaxFetchResponseSizeBytes = rc.MaxFetchResponseSizeBytes
		}
	}

	sdk := r.sdkFactory(drc)

	resp, err := fn(sdk, creq)
	if err != nil {
		return nil, fmt.Errorf("error executing custom compute: %w", err)
	}

	resppb := capabilitiespb.CapabilityResponseToProto(resp)

	return &wasmpb.Response{
		Id: id,
		Message: &wasmpb.Response_ComputeResponse{
			ComputeResponse: &wasmpb.ComputeResponse{
				Response: resppb,
			},
		},
	}, nil
}
