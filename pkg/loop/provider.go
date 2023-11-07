package loop

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

// RegisterStandAloneProvider register the servers needed for a plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed
func RegisterStandAloneProvider(s *grpc.Server, p types.PluginProvider) {
	internal.RegisterStandAloneProvider(s, p)
}
