package relayerset

import "google.golang.org/grpc"

// RegisterRelayerSetServerWithDependants registers all the grpc services hidden injected into and hidden behind RelayerSet.
func RegisterRelayerSetServerWithDependants(s grpc.ServiceRegistrar, srv RelayerSetServer) {
	RegisterRelayerSetServer(s, srv)
	switch eSrv := srv.(type) {
	case EVMRelayerSetServer:
		RegisterEVMRelayerSetServer(s, eSrv)
	}
}
