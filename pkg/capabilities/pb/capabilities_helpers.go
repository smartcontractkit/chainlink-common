package pb

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	CapabilityTypeUnknown   = CapabilityType_CAPABILITY_TYPE_UNKNOWN
	CapabilityTypeTrigger   = CapabilityType_CAPABILITY_TYPE_TRIGGER
	CapabilityTypeAction    = CapabilityType_CAPABILITY_TYPE_ACTION
	CapabilityTypeConsensus = CapabilityType_CAPABILITY_TYPE_CONSENSUS
	CapabilityTypeTarget    = CapabilityType_CAPABILITY_TYPE_TARGET
)

func MarshalCapabilityRequest(req capabilities.CapabilityRequest) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(CapabilityRequestToProto(req))
}

func MarshalCapabilityResponse(resp capabilities.CapabilityResponse) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(CapabilityResponseToProto(resp))
}

func UnmarshalCapabilityRequest(raw []byte) (capabilities.CapabilityRequest, error) {
	var cr CapabilityRequest
	if err := proto.Unmarshal(raw, &cr); err != nil {
		return capabilities.CapabilityRequest{}, err
	}
	return CapabilityRequestFromProto(&cr)
}

func UnmarshalCapabilityResponse(raw []byte) (capabilities.CapabilityResponse, error) {
	var cr CapabilityResponse
	if err := proto.Unmarshal(raw, &cr); err != nil {
		return capabilities.CapabilityResponse{}, err
	}
	return CapabilityResponseFromProto(&cr)
}

func CapabilityRequestToProto(req capabilities.CapabilityRequest) *CapabilityRequest {
	inputs := values.EmptyMap()
	if req.Inputs != nil {
		inputs = req.Inputs
	}
	config := values.EmptyMap()
	if req.Config != nil {
		config = req.Config
	}
	return &CapabilityRequest{
		Metadata: &RequestMetadata{
			WorkflowId:               req.Metadata.WorkflowID,
			WorkflowExecutionId:      req.Metadata.WorkflowExecutionID,
			WorkflowOwner:            req.Metadata.WorkflowOwner,
			WorkflowName:             req.Metadata.WorkflowName,
			WorkflowDonId:            req.Metadata.WorkflowDonID,
			WorkflowDonConfigVersion: req.Metadata.WorkflowDonConfigVersion,
			ReferenceId:              req.Metadata.ReferenceID,
		},
		Inputs: values.ProtoMap(inputs),
		Config: values.ProtoMap(config),
	}
}

func CapabilityResponseToProto(resp capabilities.CapabilityResponse) *CapabilityResponse {
	return &CapabilityResponse{
		Value: values.ProtoMap(resp.Value),
	}
}

func CapabilityRequestFromProto(pr *CapabilityRequest) (capabilities.CapabilityRequest, error) {
	if pr == nil {
		return capabilities.CapabilityRequest{}, errors.New("could not convert nil proto to CapabilityRequest")
	}

	md := pr.Metadata
	if md == nil {
		return capabilities.CapabilityRequest{}, errors.New("could not convert nil metadata to RequestMetadata")
	}

	config, err := values.FromMapValueProto(pr.Config)
	if err != nil {
		return capabilities.CapabilityRequest{}, err
	}

	inputs, err := values.FromMapValueProto(pr.Inputs)
	if err != nil {
		return capabilities.CapabilityRequest{}, err
	}

	req := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:               md.WorkflowId,
			WorkflowExecutionID:      md.WorkflowExecutionId,
			WorkflowOwner:            md.WorkflowOwner,
			WorkflowName:             md.WorkflowName,
			WorkflowDonID:            md.WorkflowDonId,
			WorkflowDonConfigVersion: md.WorkflowDonConfigVersion,
			ReferenceID:              md.ReferenceId,
		},
		Config: config,
		Inputs: inputs,
	}
	return req, nil
}

func CapabilityResponseFromProto(pr *CapabilityResponse) (capabilities.CapabilityResponse, error) {
	if pr == nil {
		return capabilities.CapabilityResponse{}, errors.New("could not convert nil proto to CapabilityResponse")
	}

	val, err := values.FromMapValueProto(pr.Value)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}

	resp := capabilities.CapabilityResponse{
		Value: val,
	}

	return resp, err
}

func MarshalTriggerRegistrationRequest(req capabilities.TriggerRegistrationRequest) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(TriggerRegistrationRequestToProto(req))
}

func MarshalTriggerResponse(resp capabilities.TriggerResponse) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(TriggerResponseToProto(resp))
}

func UnmarshalTriggerRegistrationRequest(raw []byte) (capabilities.TriggerRegistrationRequest, error) {
	var tr TriggerRegistrationRequest
	if err := proto.Unmarshal(raw, &tr); err != nil {
		return capabilities.TriggerRegistrationRequest{}, err
	}
	return TriggerRegistrationRequestFromProto(&tr)
}

func UnmarshalTriggerResponse(raw []byte) (capabilities.TriggerResponse, error) {
	var tr TriggerResponse
	if err := proto.Unmarshal(raw, &tr); err != nil {
		return capabilities.TriggerResponse{}, err
	}
	return TriggerResponseFromProto(&tr)
}

func RegisterToWorkflowRequestToProto(req capabilities.RegisterToWorkflowRequest) *RegisterToWorkflowRequest {
	config := values.EmptyMap()
	if req.Config != nil {
		config = req.Config
	}

	return &RegisterToWorkflowRequest{
		Metadata: &RegistrationMetadata{
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
		},
		Config: values.ProtoMap(config),
	}
}

func RegisterToWorkflowRequestFromProto(req *RegisterToWorkflowRequest) (capabilities.RegisterToWorkflowRequest, error) {
	if req == nil {
		return capabilities.RegisterToWorkflowRequest{}, errors.New("received nil register to workflow request")
	}

	if req.Metadata == nil {
		return capabilities.RegisterToWorkflowRequest{}, errors.New("received nil metadata in register to workflow request")
	}

	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return capabilities.RegisterToWorkflowRequest{}, err
	}

	return capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:  req.Metadata.WorkflowId,
			ReferenceID: req.Metadata.ReferenceId,
		},
		Config: config,
	}, nil
}

func UnregisterFromWorkflowRequestToProto(req capabilities.UnregisterFromWorkflowRequest) *UnregisterFromWorkflowRequest {
	config := values.EmptyMap()
	if req.Config != nil {
		config = req.Config
	}

	return &UnregisterFromWorkflowRequest{
		Metadata: &RegistrationMetadata{
			WorkflowId:  req.Metadata.WorkflowID,
			ReferenceId: req.Metadata.ReferenceID,
		},
		Config: values.ProtoMap(config),
	}
}

func UnregisterFromWorkflowRequestFromProto(req *UnregisterFromWorkflowRequest) (capabilities.UnregisterFromWorkflowRequest, error) {
	if req == nil {
		return capabilities.UnregisterFromWorkflowRequest{}, errors.New("received nil unregister from workflow request")
	}

	if req.Metadata == nil {
		return capabilities.UnregisterFromWorkflowRequest{}, errors.New("received nil metadata in unregister from workflow request")
	}

	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return capabilities.UnregisterFromWorkflowRequest{}, err
	}

	return capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID:  req.Metadata.WorkflowId,
			ReferenceID: req.Metadata.ReferenceId,
		},
		Config: config,
	}, nil
}

func UnmarshalUnregisterFromWorkflowRequest(raw []byte) (capabilities.UnregisterFromWorkflowRequest, error) {
	var r UnregisterFromWorkflowRequest
	if err := proto.Unmarshal(raw, &r); err != nil {
		return capabilities.UnregisterFromWorkflowRequest{}, err
	}
	return UnregisterFromWorkflowRequestFromProto(&r)
}

func MarshalUnregisterFromWorkflowRequest(req capabilities.UnregisterFromWorkflowRequest) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(UnregisterFromWorkflowRequestToProto(req))
}

func UnmarshalRegisterToWorkflowRequest(raw []byte) (capabilities.RegisterToWorkflowRequest, error) {
	var r RegisterToWorkflowRequest
	if err := proto.Unmarshal(raw, &r); err != nil {
		return capabilities.RegisterToWorkflowRequest{}, err
	}
	return RegisterToWorkflowRequestFromProto(&r)
}

func MarshalRegisterToWorkflowRequest(req capabilities.RegisterToWorkflowRequest) ([]byte, error) {
	return proto.MarshalOptions{Deterministic: true}.Marshal(RegisterToWorkflowRequestToProto(req))
}

func TriggerRegistrationRequestToProto(req capabilities.TriggerRegistrationRequest) *TriggerRegistrationRequest {
	md := req.Metadata

	config := values.EmptyMap()
	if req.Config != nil {
		config = req.Config
	}

	return &TriggerRegistrationRequest{
		TriggerId: req.TriggerID,
		Metadata: &RequestMetadata{
			WorkflowId:               md.WorkflowID,
			WorkflowExecutionId:      md.WorkflowExecutionID,
			WorkflowOwner:            md.WorkflowOwner,
			WorkflowName:             md.WorkflowName,
			WorkflowDonId:            md.WorkflowDonID,
			WorkflowDonConfigVersion: md.WorkflowDonConfigVersion,
		},
		Config: values.ProtoMap(config),
	}
}

func TriggerRegistrationRequestFromProto(req *TriggerRegistrationRequest) (capabilities.TriggerRegistrationRequest, error) {
	if req == nil {
		return capabilities.TriggerRegistrationRequest{}, errors.New("received nil trigger registration request")
	}

	if req.Metadata == nil {
		return capabilities.TriggerRegistrationRequest{}, errors.New("received nil metadata in trigger registration request")
	}

	md := req.Metadata

	config, err := values.FromMapValueProto(req.Config)
	if err != nil {
		return capabilities.TriggerRegistrationRequest{}, err
	}

	return capabilities.TriggerRegistrationRequest{
		TriggerID: req.TriggerId,
		Metadata: capabilities.RequestMetadata{
			WorkflowID:               md.WorkflowId,
			WorkflowExecutionID:      md.WorkflowExecutionId,
			WorkflowOwner:            md.WorkflowOwner,
			WorkflowName:             md.WorkflowName,
			WorkflowDonID:            md.WorkflowDonId,
			WorkflowDonConfigVersion: md.WorkflowDonConfigVersion,
		},
		Config: config,
	}, nil
}

func TriggerResponseToProto(resp capabilities.TriggerResponse) *TriggerResponse {
	var errs string
	if resp.Err != nil {
		errs = resp.Err.Error()
	}
	return &TriggerResponse{
		Error: errs,
		Event: &TriggerEvent{
			TriggerType: resp.Event.TriggerType,
			Id:          resp.Event.ID,
			Outputs:     values.ProtoMap(resp.Event.Outputs),
		},
	}
}

func TriggerResponseFromProto(resp *TriggerResponse) (capabilities.TriggerResponse, error) {
	if resp == nil {
		return capabilities.TriggerResponse{}, errors.New("could not unmarshal nil trigger registration response")
	}

	var event capabilities.TriggerEvent
	eventpb := resp.Event
	if eventpb != nil {
		event.TriggerType = eventpb.TriggerType
		event.ID = eventpb.Id

		outputs, err := values.FromMapValueProto(eventpb.Outputs)
		if err != nil {
			return capabilities.TriggerResponse{}, fmt.Errorf("could not unmarshal event payload: %w", err)
		}
		event.Outputs = outputs
	}

	var err error
	if resp.Error != "" {
		err = errors.New(resp.Error)
	}

	return capabilities.TriggerResponse{
		Event: event,
		Err:   err,
	}, nil
}
