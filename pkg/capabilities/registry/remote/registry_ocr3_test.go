package remote

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
)

// TestConfigForCapability_OCR3ConfigsCrossAsTheNodeHasThem checks a process gets the node's OCR3 configs as they are,
// digest and transmitter format included, so its oracles agree with the node's.
func TestConfigForCapability_OCR3ConfigsCrossAsTheNodeHasThem(t *testing.T) {
	want := ocrtypes.ContractConfig{
		ConfigDigest:          ocrtypes.ConfigDigest{0, 14, 1},
		ConfigCount:           2,
		Signers:               []ocrtypes.OnchainPublicKey{{1, 2}},
		Transmitters:          []ocrtypes.Account{"0xf98665742002D7212B9a74c37Bb17dcd0f96FF9D"},
		F:                     1,
		OnchainConfig:         []byte("onchain"),
		OffchainConfigVersion: 30,
		OffchainConfig:        []byte("offchain"),
	}

	server := &capabilitiesRegistryServer{impl: &ocr3Registry{config: want}}
	client := &capabilitiesRegistryClient{grpc: &serverAsClient{server: server}}

	got, err := client.ConfigForCapability(t.Context(), "consensus@1.0.0-alpha", 1)
	require.NoError(t, err)
	assert.Equal(t, want, got.Ocr3Configs["__default__"])
}

type ocr3Registry struct {
	registry.CapabilitiesRegistry
	config ocrtypes.ContractConfig
}

func (r *ocr3Registry) ConfigForCapability(context.Context, string, uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{Ocr3Configs: map[string]ocrtypes.ContractConfig{"__default__": r.config}}, nil
}

// serverAsClient calls server directly for the one method the test uses.
type serverAsClient struct {
	registrypb.CapabilitiesRegistryClient
	server *capabilitiesRegistryServer
}

func (c *serverAsClient) ConfigForCapability(ctx context.Context, req *registrypb.ConfigForCapabilityRequest,
	_ ...grpc.CallOption) (*registrypb.ConfigForCapabilityReply, error) {
	return c.server.ConfigForCapability(ctx, req)
}
