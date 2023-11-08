package loop

import (
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

// RegisterStandAloneProvider register the servers needed for a plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed
func RegisterStandAloneProvider(s *grpc.Server, p types.PluginProvider, pType types.OCR2PluginType) error {
	switch pType {
	case types.Median:
		mp, ok := p.(types.MedianProvider)
		if !ok {
			return errors.New(fmt.Sprintf("expected median provider got %t", p))
		}
		internal.RegisterStandAloneMedianProvider(s, mp)
		return nil
	default:
		return errors.New(fmt.Sprintf("stand alone provider only supports median, got %q", pType))
	}
}
