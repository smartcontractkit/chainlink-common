package pb

/*
import (
	"google.golang.org/protobuf/types/known/anypb"

	evmcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	solcappb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/solana"
	vault "github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/vault"
)

func PopulateTypedPayload(req *CapabilityRequest) {
	if req == nil || req.Payload == nil {
		return
	}

	if req.TypedPayload != nil {
		return
	}

	switch req.Payload.GetTypeUrl() {
	case "type.googleapis.com/capabilities.blockchain.evm.v1alpha.WriteReportRequest":
		var msg evmcappb.WriteReportRequest
		if err := req.Payload.UnmarshalTo(&msg); err == nil {
			req.TypedPayload = &CapabilityRequest_EvmWriteReportRequest{
				EvmWriteReportRequest: &msg,
			}
		}
	case "type.googleapis.com/capabilities.blockchain.solana.v1alpha.WriteReportRequest":
		var msg solcappb.WriteReportRequest
		if err := req.Payload.UnmarshalTo(&msg); err == nil {
			req.TypedPayload = &CapabilityRequest_SolanaWriteReportRequest{
				SolanaWriteReportRequest: &msg,
			}
		}
	case "type.googleapis.com/vault.GetSecretsRequest":
		var msg vault.GetSecretsRequest
		if err := req.Payload.UnmarshalTo(&msg); err == nil {
			req.TypedPayload = &CapabilityRequest_VaultGetSecretsRequest{
				VaultGetSecretsRequest: &msg,
			}
		}
	}
}

func ExtractTypedPayload(req *CapabilityRequest) {
	if req == nil {
		return
	}

	switch tp := req.TypedPayload.(type) {
	case *CapabilityRequest_EvmWriteReportRequest:
		if tp.EvmWriteReportRequest != nil && req.Payload == nil {
			req.Payload, _ = anypb.New(tp.EvmWriteReportRequest)
		}
	case *CapabilityRequest_SolanaWriteReportRequest:
		if tp.SolanaWriteReportRequest != nil && req.Payload == nil {
			req.Payload, _ = anypb.New(tp.SolanaWriteReportRequest)
		}
	case *CapabilityRequest_VaultGetSecretsRequest:
		if tp.VaultGetSecretsRequest != nil && req.Payload == nil {
			req.Payload, _ = anypb.New(tp.VaultGetSecretsRequest)
		}
	}
}

func PopulateTypedResponsePayload(resp *CapabilityResponse) {
	if resp == nil || resp.Payload == nil {
		return
	}

	if resp.TypedPayload != nil {
		return
	}

	switch resp.Payload.GetTypeUrl() {
	case "type.googleapis.com/vault.GetSecretsResponse":
		var msg vault.GetSecretsResponse
		if err := resp.Payload.UnmarshalTo(&msg); err == nil {
			resp.TypedPayload = &CapabilityResponse_VaultGetSecretsResponse{
				VaultGetSecretsResponse: &msg,
			}
		}
	}
}

func ExtractTypedResponsePayload(resp *CapabilityResponse) {
	if resp == nil {
		return
	}

	switch tp := resp.TypedPayload.(type) {
	case *CapabilityResponse_VaultGetSecretsResponse:
		if tp.VaultGetSecretsResponse != nil && resp.Payload == nil {
			resp.Payload, _ = anypb.New(tp.VaultGetSecretsResponse)
		}
	}
}
*/
