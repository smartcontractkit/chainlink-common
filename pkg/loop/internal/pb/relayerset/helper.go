package relayerset

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

// RegisterRelayerSetServerWithDependants registers all the grpc services hidden injected into and hidden behind RelayerSet.
func RegisterRelayerSetServerWithDependants(s grpc.ServiceRegistrar, srv RelayerSetServer) {
	RegisterRelayerSetServer(s, srv)
	switch eSrv := srv.(type) {
	case evm.EVMServer:
		evm.RegisterEVMServer(s, eSrv)
	}
	switch eSrv := srv.(type) {
	case ton.TONServer:
		ton.RegisterTONServer(s, eSrv)
	}
	switch eSrv := srv.(type) {
	case solana.SolanaServer:
		solana.RegisterSolanaServer(s, eSrv)
	}
	switch eSrv := srv.(type) {
	case pb.ContractReaderServer:
		pb.RegisterContractReaderServer(s, eSrv)
	}
}
