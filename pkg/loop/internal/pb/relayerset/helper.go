package relayerset

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

type HasChainServers interface {
	SolanaServer() solana.SolanaServer
	EVMServer() evm.EVMServer
	TONServer() ton.TONServer
	AptosServer() aptos.AptosServer
	ContractReaderServer() pb.ContractReaderServer
}

// RegisterRelayerSetServerWithDependants registers all the grpc services hidden injected into and hidden behind RelayerSet.
func RegisterRelayerSetServerWithDependants(s grpc.ServiceRegistrar, srv RelayerSetServer) {
	RegisterRelayerSetServer(s, srv)
	switch eSrv := srv.(type) {
	case pb.ContractReaderServer:
		pb.RegisterContractReaderServer(s, eSrv)
	}
	if h, ok := srv.(HasChainServers); ok {
		solana.RegisterSolanaServer(s, h.SolanaServer())
		evm.RegisterEVMServer(s, h.EVMServer())
		ton.RegisterTONServer(s, h.TONServer())
		aptos.RegisterAptosServer(s, h.AptosServer())
		pb.RegisterContractReaderServer(s, h.ContractReaderServer())
	}
}
