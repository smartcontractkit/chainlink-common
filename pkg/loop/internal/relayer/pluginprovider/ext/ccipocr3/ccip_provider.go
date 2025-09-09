package ccipocr3

import (
	"google.golang.org/grpc"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	ocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

var (
	_ types.CCIPProvider      = (*CCIPProviderClient)(nil)
	_ goplugin.GRPCClientConn = (*CCIPProviderClient)(nil)
)

type CCIPProviderClient struct {
	*goplugin.ServiceClient
	chainAccessor             *ChainAccessorClient
	contractTransmitter       ocr3types.ContractTransmitter[[]byte]
	chainSpecificAddressCodec ccipocr3.ChainSpecificAddressCodec
	commitPluginCodec         ccipocr3.CommitPluginCodec
	executePluginCodec        ccipocr3.ExecutePluginCodec
	tokenDataEncoder          ccipocr3.TokenDataEncoder
	sourceChainExtraDataCodec ccipocr3.SourceChainExtraDataCodec
}

func NewCCIPProviderClient(b *net.BrokerExt, cc grpc.ClientConnInterface) *CCIPProviderClient {
	c := &CCIPProviderClient{
		ServiceClient: goplugin.NewServiceClient(b.WithName("CCIPProviderClient"), cc),
	}

	// Initialize ChainAccessor
	c.chainAccessor = NewChainAccessorClient(b.WithName("ChainAccessor"), cc)

	// Initialize ContractTransmitter (reuse existing OCR3 implementation)
	c.contractTransmitter = ocr3.NewContractTransmitterClient(b.WithName("ContractTransmitter"), cc)

	// Initialize Codec components
	c.chainSpecificAddressCodec = NewChainSpecificAddressCodecClient(b.WithName("ChainSpecificAddressCodec"), cc)
	c.commitPluginCodec = NewCommitPluginCodecClient(b.WithName("CommitPluginCodec"), cc)
	c.executePluginCodec = NewExecutePluginCodecClient(b.WithName("ExecutePluginCodec"), cc)
	c.tokenDataEncoder = NewTokenDataEncoderClient(b.WithName("TokenDataEncoder"), cc)
	c.sourceChainExtraDataCodec = NewSourceChainExtraDataCodecClient(b.WithName("SourceChainExtraDataCodec"), cc)

	return c
}

func (p *CCIPProviderClient) GetSyncRequests() []*ccipocr3pb.SyncRequest {
	return p.chainAccessor.GetSyncRequests()
}

func (p *CCIPProviderClient) ChainAccessor() ccipocr3.ChainAccessor {
	return p.chainAccessor
}

func (p *CCIPProviderClient) ContractTransmitter() ocr3types.ContractTransmitter[[]byte] {
	return p.contractTransmitter
}

func (p *CCIPProviderClient) Codec() ccipocr3.Codec {
	return ccipocr3.Codec{
		ChainSpecificAddressCodec: p.chainSpecificAddressCodec,
		CommitPluginCodec:         p.commitPluginCodec,
		ExecutePluginCodec:        p.executePluginCodec,
		TokenDataEncoder:          p.tokenDataEncoder,
		SourceChainExtraDataCodec: p.sourceChainExtraDataCodec,
	}
}

// Server implementation
type CCIPProviderServer struct{}

func (s CCIPProviderServer) ConnToProvider(conn grpc.ClientConnInterface, broker net.Broker, brokerCfg net.BrokerConfig) types.CCIPProvider {
	be := &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}
	return NewCCIPProviderClient(be, conn)
}

func RegisterProviderServices(s *grpc.Server, provider types.CCIPProvider) {
	// Register base Service for goplugin framework
	pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: provider})

	// Register ChainAccessor service
	ccipocr3pb.RegisterChainAccessorServer(s, NewChainAccessorServer(provider.ChainAccessor()))

	// Register ContractTransmitter service (reuse existing OCR3 implementation)
	ocr3pb.RegisterContractTransmitterServer(s, ocr3.NewContractTransmitterServer(provider.ContractTransmitter()))

	// Register Codec services
	codec := provider.Codec()
	ccipocr3pb.RegisterChainSpecificAddressCodecServer(s, NewChainSpecificAddressCodecServer(codec.ChainSpecificAddressCodec))
	ccipocr3pb.RegisterCommitPluginCodecServer(s, NewCommitPluginCodecServer(codec.CommitPluginCodec))
	ccipocr3pb.RegisterExecutePluginCodecServer(s, NewExecutePluginCodecServer(codec.ExecutePluginCodec))
	ccipocr3pb.RegisterTokenDataEncoderServer(s, NewTokenDataEncoderServer(codec.TokenDataEncoder))
	ccipocr3pb.RegisterSourceChainExtraDataCodecServer(s, NewSourceChainExtraDataCodecServer(codec.SourceChainExtraDataCodec))
}
