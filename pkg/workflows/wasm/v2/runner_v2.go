package v2

import (
	"encoding/base64"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type RunnerV2 struct {
	sendResponse   func(payload *wasmpb.Response)
	runtimeFactory func(sdkConfig *wasm.RuntimeConfig, refToResponse map[string]capabilities.CapabilityResponse, hostReqID string) *RuntimeV2
	args           []string
	req            *wasmpb.Request
	triggers       map[string]triggerInfo
}

type triggerInfo struct {
	id        string
	config    *values.Map
	handlerFn func(runtime sdk.RuntimeV2, triggerEvent capabilities.TriggerEvent) error
}

func SubscribeToTrigger[TriggerConfig any, TriggerOutputs any](runner *RunnerV2, id string, triggerCfg TriggerConfig, handler func(runtime sdk.RuntimeV2, triggerOutputs TriggerOutputs) error) error {
	ref := fmt.Sprintf("trigger-%v", len(runner.triggers))

	wrappedConfig, err := values.WrapMap(triggerCfg)
	if err != nil {
		return fmt.Errorf("could not wrap config into map: %w", err)
	}

	runner.triggers[ref] = triggerInfo{
		id:     id,
		config: wrappedConfig,
		handlerFn: func(runtime sdk.RuntimeV2, triggerEvent capabilities.TriggerEvent) error {
			var triggerOutputs TriggerOutputs
			err := triggerEvent.Outputs.UnwrapTo(&triggerOutputs)
			if err != nil {
				return err
			}
			return handler(runtime, triggerOutputs)
		},
	}
	return nil
}

func (r *RunnerV2) Run() {
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
				r.sendResponse(wasm.ErrorResponse(r.req.Id, asErr))
			} else {
				r.sendResponse(wasm.ErrorResponse(r.req.Id, fmt.Errorf("caught panic: %+v", err)))
			}
		}
	}()

	resp := &wasmpb.Response{
		Id: req.Id,
	}

	switch {
	case req.GetSpecRequest() != nil:
		rsp, innerErr := r.handleSpecRequest(req.Id)
		if innerErr != nil {
			resp.ErrMsg = innerErr.Error()
		} else {
			resp = rsp
		}
	case req.GetRunRequest() != nil:
		rsp, innerErr := r.handleRunRequest(req.Id, req.GetRunRequest())
		if innerErr != nil {
			resp.ErrMsg = innerErr.Error()
		} else {
			resp = rsp
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

	r.sendResponse(wasm.ErrorResponse(r.req.Id, err))
}

func (r *RunnerV2) cacheRequest() bool {
	if r.req == nil {
		req, err := r.parseRequest()
		if err != nil {
			r.sendResponse(wasm.ErrorResponse(wasm.UnknownID, err))
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

func (r *RunnerV2) handleSpecRequest(id string) (*wasmpb.Response, error) {
	specpb := &wasmpb.WorkflowSpec{
		Name:      "name_TODO",
		Owner:     "owner_TODO",
		IsDynamic: true,
	}

	for ref, info := range r.triggers {
		specpb.Triggers = append(specpb.Triggers, &wasmpb.StepDefinition{
			Id:             info.id,
			Ref:            ref,
			Inputs:         &wasmpb.StepInputs{},
			Config:         values.ProtoMap(info.config),
			CapabilityType: string(capabilities.CapabilityTypeTrigger),
		})
	}

	return &wasmpb.Response{
		Id: id,
		Message: &wasmpb.Response_SpecResponse{
			SpecResponse: specpb,
		},
	}, nil
}

func (r *RunnerV2) handleRunRequest(id string, runReq *wasmpb.RunRequest) (*wasmpb.Response, error) {
	// Extract config from the request
	drc := wasm.DefaultRuntimeConfig(id, nil)

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
	triggerInfo := r.triggers[runReq.TriggerRef]
	if triggerInfo.handlerFn == nil {
		return nil, fmt.Errorf("could not find run function for ref %s", runReq.TriggerRef)
	}
	err = triggerInfo.handlerFn(runtime, event)
	if err != nil {
		return nil, fmt.Errorf("error executing workflow: %w", err)
	}

	// successful execution termination
	return &wasmpb.Response{
		Id: id,
	}, nil
}
