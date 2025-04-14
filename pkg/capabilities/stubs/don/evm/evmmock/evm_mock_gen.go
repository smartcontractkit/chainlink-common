// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package evmmock

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/crosschain"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

// avoid unused imports
var _ = capabilities.CapabilityInfo{}
var _ = testutils.Registry{}

type ClientCapability struct {
	// TODO teardown with unrgister if register is needed, or allow setup and teardown
	// TODO register if needed...
	GetTxResult func(ctx context.Context, input *evm.TxID /* TODO config? */) (*crosschain.TxResult, error)
	// TODO register if needed...
	ReadMethod func(ctx context.Context, input *evm.ReadMethodRequest /* TODO config? */) (*crosschain.ByteArray, error)
	// TODO register if needed...
	QueryLogs func(ctx context.Context, input *evm.QueryLogsRequest /* TODO config? */) (*evm.LogList, error)
	// TODO register if needed...
	SubmitTransaction func(ctx context.Context, input *evm.SubmitTransactionRequest /* TODO config? */) (*evm.TxID, error)

	OnFinalityViolation func(ctx context.Context, input *emptypb.Empty) (capabilities.TriggerAndId[*crosschain.BlockRange], error)
}

func (cap *ClientCapability) Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	capResp := &pb.CapabilityResponse{}
	switch request.Method {
	case "GetTxResult":
		input := &evm.TxID{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.GetTxResult == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for GetTxResult"}
			break
		}
		resp, err := cap.GetTxResult(ctx, input)
		if err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &pb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "ReadMethod":
		input := &evm.ReadMethodRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.ReadMethod == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for ReadMethod"}
			break
		}
		resp, err := cap.ReadMethod(ctx, input)
		if err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &pb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "QueryLogs":
		input := &evm.QueryLogsRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.QueryLogs == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for QueryLogs"}
			break
		}
		resp, err := cap.QueryLogs(ctx, input)
		if err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &pb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	case "SubmitTransaction":
		input := &evm.SubmitTransactionRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.SubmitTransaction == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for SubmitTransaction"}
			break
		}
		resp, err := cap.SubmitTransaction(ctx, input)
		if err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &pb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	default:
		capResp.Response = &pb.CapabilityResponse_Error{Error: fmt.Sprintf("method %s not found", request.Method)}
	}
	return capResp
}

func (cap *ClientCapability) InvokeTrigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error) {
	trigger := &pb.Trigger{}
	switch request.Method {
	case "OnFinalityViolation":
		input := &emptypb.Empty{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			return nil, err
		}

		if cap.OnFinalityViolation == nil {
			return nil, testutils.NoTriggerStub("OnFinalityViolation")
		}

		resp, err := cap.OnFinalityViolation(ctx, input)
		if err != nil {
			return nil, err
		} else {
			payload, err := anypb.New(resp.Trigger)
			if err != nil {
				return nil, err
			}
			trigger.Payload = payload
			trigger.Id = resp.Id
		}
	default:
		return nil, fmt.Errorf("method %s not found", request.Method)
	}
	return trigger, nil
}

func (cap *ClientCapability) ID() string {
	return "evm@1.0.0"
}
