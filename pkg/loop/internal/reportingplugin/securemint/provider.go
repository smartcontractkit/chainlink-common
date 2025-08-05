package securemint

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// RegisterProviderServices registers all SecureMint provider services on the given gRPC server
func RegisterProviderServices(s *grpc.Server, provider types.SecureMintProvider, brokerExt *net.BrokerExt) {
	// TODO(gg): Register specific provider services when they are defined
	// This would typically include:
	// - ExternalAdapter service
	// - ContractReader service
	// - ReportMarshaler service
	// - Any other provider-specific services
	
	// For now, we have placeholder implementations in the adapters
	// The actual services will be registered when the external plugin is integrated
} 