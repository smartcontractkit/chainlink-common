// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc, DO NOT EDIT.

package evm

import (
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type Client struct {
	// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 allow defaults for capabilities
}

func (c *Client) CallContract(runtime sdk.DonRuntime, input *evm.CallContractRequest) sdk.Promise[*evm.CallContractReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.CallContractReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "CallContract",
	}), func(i *sdkpb.CapabilityResponse) (*evm.CallContractReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.CallContractReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) FilterLogs(runtime sdk.DonRuntime, input *evm.FilterLogsRequest) sdk.Promise[*evm.FilterLogsReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.FilterLogsReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "FilterLogs",
	}), func(i *sdkpb.CapabilityResponse) (*evm.FilterLogsReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.FilterLogsReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) BalanceAt(runtime sdk.DonRuntime, input *evm.BalanceAtRequest) sdk.Promise[*evm.BalanceAtReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.BalanceAtReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "BalanceAt",
	}), func(i *sdkpb.CapabilityResponse) (*evm.BalanceAtReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.BalanceAtReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) EstimateGas(runtime sdk.DonRuntime, input *evm.EstimateGasRequest) sdk.Promise[*evm.EstimateGasReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.EstimateGasReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "EstimateGas",
	}), func(i *sdkpb.CapabilityResponse) (*evm.EstimateGasReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.EstimateGasReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) GetTransactionByHash(runtime sdk.DonRuntime, input *evm.GetTransactionByHashRequest) sdk.Promise[*evm.GetTransactionByHashReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.GetTransactionByHashReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "GetTransactionByHash",
	}), func(i *sdkpb.CapabilityResponse) (*evm.GetTransactionByHashReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.GetTransactionByHashReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) GetTransactionReceipt(runtime sdk.DonRuntime, input *evm.GetTransactionReceiptRequest) sdk.Promise[*evm.GetTransactionReceiptReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.GetTransactionReceiptReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "GetTransactionReceipt",
	}), func(i *sdkpb.CapabilityResponse) (*evm.GetTransactionReceiptReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.GetTransactionReceiptReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) LatestAndFinalizedHead(runtime sdk.DonRuntime, input *emptypb.Empty) sdk.Promise[*evm.LatestAndFinalizedHeadReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.LatestAndFinalizedHeadReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "LatestAndFinalizedHead",
	}), func(i *sdkpb.CapabilityResponse) (*evm.LatestAndFinalizedHeadReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.LatestAndFinalizedHeadReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) QueryTrackedLogs(runtime sdk.DonRuntime, input *evm.QueryTrackedLogsRequest) sdk.Promise[*evm.QueryTrackedLogsReply] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*evm.QueryTrackedLogsReply](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "QueryTrackedLogs",
	}), func(i *sdkpb.CapabilityResponse) (*evm.QueryTrackedLogsReply, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &evm.QueryTrackedLogsReply{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) RegisterLogTracking(runtime sdk.DonRuntime, input *evm.RegisterLogTrackingRequest) sdk.Promise[*emptypb.Empty] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*emptypb.Empty](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "RegisterLogTracking",
	}), func(i *sdkpb.CapabilityResponse) (*emptypb.Empty, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &emptypb.Empty{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}

func (c *Client) UnregisterLogTracking(runtime sdk.DonRuntime, input *evm.UnregisterLogTrackingRequest) sdk.Promise[*emptypb.Empty] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*emptypb.Empty](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "evm@1.0.0",
		Payload: wrapped,
		Method:  "UnregisterLogTracking",
	}), func(i *sdkpb.CapabilityResponse) (*emptypb.Empty, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &emptypb.Empty{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}
