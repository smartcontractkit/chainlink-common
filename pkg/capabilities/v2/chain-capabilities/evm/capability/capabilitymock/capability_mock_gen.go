// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc, DO NOT EDIT.

package evmcappbmock

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/protobuf/types/known/emptypb"

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"

	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"
)

// avoid unused imports
var _ = registry.Registry{}

func NewEVMCapability(t testing.TB) (*EVMCapability, error) {
	c := &EVMCapability{}
	reg := registry.GetRegistry(t)
	err := reg.RegisterCapability(c)
	return c, err
}

type EVMCapability struct {
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	CallContract func(ctx context.Context, input *evmpb.CallContractRequest) (*evmpb.CallContractReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	FilterLogs func(ctx context.Context, input *evmpb.FilterLogsRequest) (*evmpb.FilterLogsReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	BalanceAt func(ctx context.Context, input *evmpb.BalanceAtRequest) (*evmpb.BalanceAtReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	EstimateGas func(ctx context.Context, input *evmpb.EstimateGasRequest) (*evmpb.EstimateGasReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	GetTransactionByHash func(ctx context.Context, input *evmpb.GetTransactionByHashRequest) (*evmpb.GetTransactionByHashReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	GetTransactionReceipt func(ctx context.Context, input *evmpb.GetTransactionReceiptRequest) (*evmpb.GetTransactionReceiptReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	LatestAndFinalizedHead func(ctx context.Context, input *emptypb.Empty) (*evmpb.LatestAndFinalizedHeadReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	QueryTrackedLogs func(ctx context.Context, input *evmpb.QueryTrackedLogsRequest) (*evmpb.QueryTrackedLogsReply, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	RegisterLogTracking func(ctx context.Context, input *evmpb.RegisterLogTrackingRequest) (*emptypb.Empty, error)
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
	UnregisterLogTracking func(ctx context.Context, input *evmpb.UnregisterLogTrackingRequest) (*emptypb.Empty, error)
}

func (cap *EVMCapability) Invoke(ctx context.Context, request *sdkpb.CapabilityRequest) *sdkpb.CapabilityResponse {
	capResp := &sdkpb.CapabilityResponse{}
	switch request.Method {
	case "CallContract":
		input := &evmpb.CallContractRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.CallContract == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for CallContract"}
			break
		}
		resp, err := cap.CallContract(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "FilterLogs":
		input := &evmpb.FilterLogsRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.FilterLogs == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for FilterLogs"}
			break
		}
		resp, err := cap.FilterLogs(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "BalanceAt":
		input := &evmpb.BalanceAtRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.BalanceAt == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for BalanceAt"}
			break
		}
		resp, err := cap.BalanceAt(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "EstimateGas":
		input := &evmpb.EstimateGasRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.EstimateGas == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for EstimateGas"}
			break
		}
		resp, err := cap.EstimateGas(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "GetTransactionByHash":
		input := &evmpb.GetTransactionByHashRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.GetTransactionByHash == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for GetTransactionByHash"}
			break
		}
		resp, err := cap.GetTransactionByHash(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "GetTransactionReceipt":
		input := &evmpb.GetTransactionReceiptRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.GetTransactionReceipt == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for GetTransactionReceipt"}
			break
		}
		resp, err := cap.GetTransactionReceipt(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "LatestAndFinalizedHead":
		input := &emptypb.Empty{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.LatestAndFinalizedHead == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for LatestAndFinalizedHead"}
			break
		}
		resp, err := cap.LatestAndFinalizedHead(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "QueryTrackedLogs":
		input := &evmpb.QueryTrackedLogsRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.QueryTrackedLogs == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for QueryTrackedLogs"}
			break
		}
		resp, err := cap.QueryTrackedLogs(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "RegisterLogTracking":
		input := &evmpb.RegisterLogTrackingRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.RegisterLogTracking == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for RegisterLogTracking"}
			break
		}
		resp, err := cap.RegisterLogTracking(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "UnregisterLogTracking":
		input := &evmpb.UnregisterLogTrackingRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.UnregisterLogTracking == nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: "no stub provided for UnregisterLogTracking"}
			break
		}
		resp, err := cap.UnregisterLogTracking(ctx, input)
		if err != nil {
			capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &sdkpb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &sdkpb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	default:
		capResp.Response = &sdkpb.CapabilityResponse_Error{Error: fmt.Sprintf("method %s not found", request.Method)}
	}
	return capResp
}

func (cap *EVMCapability) InvokeTrigger(ctx context.Context, request *sdkpb.TriggerSubscription) (*sdkpb.Trigger, error) {
	return nil, fmt.Errorf("method %s not found", request.Method)
}

func (cap *EVMCapability) ID() string {
	return "mainnet-evm@1.0.0"
}
