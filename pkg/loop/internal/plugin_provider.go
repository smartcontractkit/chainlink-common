package internal

import (
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type pluginProviderClient struct {
	*configProviderClient
	chainReader         types.ChainReader
	contractTransmitter libocr.ContractTransmitter
}

func (p *pluginProviderClient) ChainReader() types.ChainReader {
	return p.chainReader
}

func (p *pluginProviderClient) ClientConn() grpc.ClientConnInterface { return p.cc }

func newPluginProviderClient(b *brokerExt, cc grpc.ClientConnInterface) *pluginProviderClient {
	p := &pluginProviderClient{configProviderClient: newConfigProviderClient(b.withName("PluginProviderClient"), cc)}
	p.contractTransmitter = &contractTransmitterClient{b, pb.NewContractTransmitterClient(p.cc)}
	return p
}

func (p *pluginProviderClient) ContractTransmitter() libocr.ContractTransmitter {
	return p.contractTransmitter
}
