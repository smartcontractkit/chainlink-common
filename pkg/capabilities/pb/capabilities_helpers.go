package pb

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	meter "github.com/smartcontractkit/chainlink-common/pkg/metering/pb"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

const (
	CapabilityTypeUnknown   = CapabilityType_CAPABILITY_TYPE_UNKNOWN
	CapabilityTypeTrigger   = CapabilityType_CAPABILITY_TYPE_TRIGGER
	CapabilityTypeAction    = CapabilityType_CAPABILITY_TYPE_ACTION
	CapabilityTypeConsensus = CapabilityType_CAPABILITY_TYPE_CONSENSUS
	CapabilityTypeTarget    = CapabilityType_CAPABILITY_TYPE_TARGET
	CapabilityTypeCombined  = CapabilityType_CAPABILITY_TYPE_COMBINED

	// OCR3ConfigDefaultKey is the default key used in the ocr3_configs map
	// for single-instance OCR capabilities. Multi-instance capabilities
	// (e.g., blue/green deployments) use custom keys.
	OCR3ConfigDefaultKey = "__default__"
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
			DecodedWorkflowName:      req.Metadata.DecodedWorkflowName,
			SpendLimits:              spendLimitsToProto(req.Metadata.SpendLimits),
			WorkflowTag:              req.Metadata.WorkflowTag,
		},
		Inputs:        values.ProtoMap(inputs),
		Config:        values.ProtoMap(config),
		Payload:       req.Payload,
		Method:        req.Method,
		CapabilityId:  req.CapabilityId,
		ConfigPayload: req.ConfigPayload,
	}
}

func CapabilityResponseToProto(resp capabilities.CapabilityResponse) *CapabilityResponse {
	metering := make([]*meter.MeteringReportNodeDetail, len(resp.Metadata.Metering))
	for idx, detail := range resp.Metadata.Metering {
		metering[idx] = &meter.MeteringReportNodeDetail{
			Peer_2PeerId: detail.Peer2PeerID,
			SpendUnit:    detail.SpendUnit,
			SpendValue:   detail.SpendValue,
		}
	}

	return &CapabilityResponse{
		Value: values.ProtoMap(resp.Value),
		Metadata: &ResponseMetadata{
			Metering: metering,
			CapdonN:  resp.Metadata.CapDON_N,
		},
		Payload: resp.Payload,
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
			DecodedWorkflowName:      md.DecodedWorkflowName,
			SpendLimits:              spendLimitsFromProto(md.SpendLimits),
			WorkflowTag:              md.WorkflowTag,
		},
		Config:        config,
		Inputs:        inputs,
		Payload:       pr.Payload,
		Method:        pr.Method,
		CapabilityId:  pr.CapabilityId,
		ConfigPayload: pr.ConfigPayload,
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

	var metering []capabilities.MeteringNodeDetail

	if pr.Metadata != nil {
		metering = make([]capabilities.MeteringNodeDetail, len(pr.Metadata.Metering))

		for idx, detail := range pr.Metadata.Metering {
			metering[idx] = capabilities.MeteringNodeDetail{
				Peer2PeerID: detail.Peer_2PeerId,
				SpendUnit:   detail.SpendUnit,
				SpendValue:  detail.SpendValue,
			}
		}
	}

	resp := capabilities.CapabilityResponse{
		Value: val,
		Metadata: capabilities.ResponseMetadata{
			Metering: metering,
			CapDON_N: pr.Metadata.GetCapdonN(),
		},
		Payload: pr.Payload,
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
			WorkflowId:    req.Metadata.WorkflowID,
			ReferenceId:   req.Metadata.ReferenceID,
			WorkflowOwner: req.Metadata.WorkflowOwner,
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
			WorkflowID:    req.Metadata.WorkflowId,
			ReferenceID:   req.Metadata.ReferenceId,
			WorkflowOwner: req.Metadata.WorkflowOwner,
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
			WorkflowId:    req.Metadata.WorkflowID,
			ReferenceId:   req.Metadata.ReferenceID,
			WorkflowOwner: req.Metadata.WorkflowOwner,
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
			WorkflowID:    req.Metadata.WorkflowId,
			ReferenceID:   req.Metadata.ReferenceId,
			WorkflowOwner: req.Metadata.WorkflowOwner,
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
			WorkflowId:                    md.WorkflowID,
			WorkflowExecutionId:           md.WorkflowExecutionID,
			WorkflowOwner:                 md.WorkflowOwner,
			WorkflowName:                  md.WorkflowName,
			WorkflowDonId:                 md.WorkflowDonID,
			WorkflowDonConfigVersion:      md.WorkflowDonConfigVersion,
			ReferenceId:                   md.ReferenceID,
			DecodedWorkflowName:           md.DecodedWorkflowName,
			SpendLimits:                   spendLimitsToProto(md.SpendLimits),
			WorkflowTag:                   md.WorkflowTag,
			WorkflowRegistryChainSelector: md.WorkflowRegistryChainSelector,
			WorkflowRegistryAddress:       md.WorkflowRegistryAddress,
			EngineVersion:                 md.EngineVersion,
		},
		Config:  values.ProtoMap(config),
		Payload: req.Payload,
		Method:  req.Method,
	}
}

func spendLimitsToProto(limits []capabilities.SpendLimit) []*SpendLimit {
	result := make([]*SpendLimit, len(limits))
	for i, limit := range limits {
		result[i] = &SpendLimit{
			SpendType: string(limit.SpendType),
			Limit:     limit.Limit,
		}
	}
	return result
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
			WorkflowID:                    md.WorkflowId,
			WorkflowOwner:                 md.WorkflowOwner,
			WorkflowExecutionID:           md.WorkflowExecutionId,
			WorkflowName:                  md.WorkflowName,
			WorkflowDonID:                 md.WorkflowDonId,
			WorkflowDonConfigVersion:      md.WorkflowDonConfigVersion,
			ReferenceID:                   md.ReferenceId,
			DecodedWorkflowName:           md.DecodedWorkflowName,
			SpendLimits:                   spendLimitsFromProto(md.SpendLimits),
			WorkflowTag:                   md.WorkflowTag,
			WorkflowRegistryChainSelector: md.WorkflowRegistryChainSelector,
			WorkflowRegistryAddress:       md.WorkflowRegistryAddress,
			EngineVersion:                 md.EngineVersion,
		},
		Config:  config,
		Payload: req.Payload,
		Method:  req.Method,
	}, nil
}

func spendLimitsFromProto(limits []*SpendLimit) []capabilities.SpendLimit {
	result := make([]capabilities.SpendLimit, len(limits))
	for i, limit := range limits {
		result[i] = capabilities.SpendLimit{
			SpendType: capabilities.CapabilitySpendType(limit.GetSpendType()),
			Limit:     limit.GetLimit(),
		}
	}
	return result
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
			Payload:     resp.Event.Payload,
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
		event.Payload = eventpb.Payload
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
