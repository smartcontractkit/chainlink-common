package median

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr2"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var (
	_ types.MedianProvider    = (*ProviderClient)(nil)
	_ goplugin.GRPCClientConn = (*ProviderClient)(nil)
)

type ProviderClient struct {
	*ocr2.PluginProviderClient
	reportCodec        median.ReportCodec
	medianContract     median.MedianContract
	onchainConfigCodec median.OnchainConfigCodec
	contractReader     types.ContractReader
	codec              types.Codec
}

func NewProviderClient(b *net.BrokerExt, cc grpc.ClientConnInterface) *ProviderClient {
	m := &ProviderClient{PluginProviderClient: ocr2.NewPluginProviderClient(b.WithName("MedianProviderClient"), cc)}
	m.reportCodec = &reportCodecClient{b, pb.NewReportCodecClient(cc)}
	m.medianContract = &medianContractClient{pb.NewMedianContractClient(cc)}
	m.onchainConfigCodec = &onchainConfigCodecClient{b, pb.NewOnchainConfigCodecClient(cc)}
	m.contractReader = contractreader.NewClient(b, cc)
	m.codec = contractreader.NewCodecClient(b, cc)
	return m
}

// RmUnimplemented sets the ContractReader and Codec clients to nil, if they are not implemented.
func (p *ProviderClient) RmUnimplemented(ctx context.Context) {
	var anyRetVal int
	err := p.contractReader.GetLatestValue(ctx, "", primitives.Unconfirmed, nil, &anyRetVal)
	if status.Convert(err).Code() == codes.Unimplemented {
		p.contractReader = nil
	}

	err = p.codec.Decode(ctx, []byte{}, &anyRetVal, "")
	if status.Convert(err).Code() == codes.Unimplemented {
		p.codec = nil
	}
}

func (p *ProviderClient) ReportCodec() median.ReportCodec {
	return p.reportCodec
}

func (p *ProviderClient) MedianContract() median.MedianContract {
	return p.medianContract
}

func (p *ProviderClient) OnchainConfigCodec() median.OnchainConfigCodec {
	return p.onchainConfigCodec
}

func (p *ProviderClient) ContractReader() types.ContractReader {
	return p.contractReader
}

func (p *ProviderClient) Codec() types.Codec {
	return p.codec
}

var _ median.ReportCodec = (*reportCodecClient)(nil)

type reportCodecClient struct {
	*net.BrokerExt
	grpc pb.ReportCodecClient
}

func (r *reportCodecClient) BuildReport(ctx context.Context, observations []median.ParsedAttributedObservation) (report libocr.Report, err error) {
	var req pb.BuildReportRequest
	for _, o := range observations {
		req.Observations = append(req.Observations, &pb.ParsedAttributedObservation{
			Timestamp:        o.Timestamp,
			Value:            pb.NewBigIntFromInt(o.Value),
			JulesPerFeeCoin:  pb.NewBigIntFromInt(o.JuelsPerFeeCoin),
			GasPriceSubunits: pb.NewBigIntFromInt(o.GasPriceSubunits),
			Observer:         uint32(o.Observer),
		})
	}
	var reply *pb.BuildReportReply
	reply, err = r.grpc.BuildReport(ctx, &req)
	if err != nil {
		return
	}
	report = reply.Report
	return
}

func (r *reportCodecClient) MedianFromReport(ctx context.Context, report libocr.Report) (*big.Int, error) {
	reply, err := r.grpc.MedianFromReport(ctx, &pb.MedianFromReportRequest{Report: report})
	if err != nil {
		return nil, err
	}
	return reply.Median.Int(), nil
}

func (r *reportCodecClient) MaxReportLength(ctx context.Context, n int) (int, error) {
	reply, err := r.grpc.MaxReportLength(ctx, &pb.MaxReportLengthRequest{N: int64(n)})
	if err != nil {
		return -1, err
	}
	return int(reply.Max), nil
}

var _ pb.ReportCodecServer = (*reportCodecServer)(nil)

type reportCodecServer struct {
	pb.UnimplementedReportCodecServer
	impl median.ReportCodec
}

func (r *reportCodecServer) BuildReport(ctx context.Context, request *pb.BuildReportRequest) (*pb.BuildReportReply, error) {
	var obs []median.ParsedAttributedObservation
	for _, o := range request.Observations {
		val, jpfc, gpsu := o.Value.Int(), o.JulesPerFeeCoin.Int(), o.GasPriceSubunits.Int()
		if o.Observer > math.MaxUint8 {
			return nil, fmt.Errorf("expected uint8 Observer (max %d) but got %d", math.MaxUint8, o.Observer)
		}
		obs = append(obs, median.ParsedAttributedObservation{
			Timestamp:        o.Timestamp,
			Value:            val,
			JuelsPerFeeCoin:  jpfc,
			GasPriceSubunits: gpsu,
			Observer:         commontypes.OracleID(o.Observer),
		})
	}
	report, err := r.impl.BuildReport(ctx, obs)
	if err != nil {
		return nil, err
	}
	return &pb.BuildReportReply{Report: report}, nil
}

func (r *reportCodecServer) MedianFromReport(ctx context.Context, request *pb.MedianFromReportRequest) (*pb.MedianFromReportReply, error) {
	m, err := r.impl.MedianFromReport(ctx, request.Report)
	if err != nil {
		return nil, err
	}
	return &pb.MedianFromReportReply{Median: pb.NewBigIntFromInt(m)}, nil
}

func (r *reportCodecServer) MaxReportLength(ctx context.Context, request *pb.MaxReportLengthRequest) (*pb.MaxReportLengthReply, error) {
	l, err := r.impl.MaxReportLength(ctx, int(request.N))
	if err != nil {
		return nil, err
	}
	return &pb.MaxReportLengthReply{Max: int64(l)}, nil
}

var _ median.MedianContract = (*medianContractClient)(nil)

type medianContractClient struct {
	grpc pb.MedianContractClient
}

func (m *medianContractClient) LatestTransmissionDetails(ctx context.Context) (configDigest libocr.ConfigDigest, epoch uint32, round uint8, latestAnswer *big.Int, latestTimestamp time.Time, err error) {
	var reply *pb.LatestTransmissionDetailsReply
	reply, err = m.grpc.LatestTransmissionDetails(ctx, &pb.LatestTransmissionDetailsRequest{})
	if err != nil {
		return
	}
	if l := len(reply.ConfigDigest); l != 32 {
		err = fmt.Errorf("expected ConfigDigest length 32 but got %d", l)
		return
	}
	copy(configDigest[:], reply.ConfigDigest)
	epoch = reply.Epoch
	if reply.Round > math.MaxUint8 {
		err = fmt.Errorf("expected uint8 Round (max %d) but got %d", math.MaxUint8, reply.Round)
		return
	}
	round = uint8(reply.Round)
	latestAnswer = reply.LatestAnswer.Int()
	latestTimestamp = reply.LatestTimestamp.AsTime()
	return
}

func (m *medianContractClient) LatestRoundRequested(ctx context.Context, lookback time.Duration) (configDigest libocr.ConfigDigest, epoch uint32, round uint8, err error) {
	reply, err := m.grpc.LatestRoundRequested(ctx, &pb.LatestRoundRequestedRequest{Lookback: int64(lookback)})
	if err != nil {
		return
	}
	if l := len(reply.ConfigDigest); l != 32 {
		err = fmt.Errorf("expected ConfigDigest length 32 but got %d", l)
		return
	}
	copy(configDigest[:], reply.ConfigDigest)
	epoch = reply.Epoch
	if reply.Round > math.MaxUint8 {
		err = fmt.Errorf("expected uint8 Round (max %d) but got %d", math.MaxUint8, reply.Round)
		return
	}
	round = uint8(reply.Round)
	return
}

var _ pb.MedianContractServer = (*medianContractServer)(nil)

type medianContractServer struct {
	pb.UnimplementedMedianContractServer
	impl median.MedianContract
}

func (m *medianContractServer) LatestTransmissionDetails(ctx context.Context, _ *pb.LatestTransmissionDetailsRequest) (*pb.LatestTransmissionDetailsReply, error) {
	digest, epoch, round, latestAnswer, latestTimestamp, err := m.impl.LatestTransmissionDetails(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.LatestTransmissionDetailsReply{
		ConfigDigest:    digest[:],
		Epoch:           epoch,
		Round:           uint32(round),
		LatestAnswer:    pb.NewBigIntFromInt(latestAnswer),
		LatestTimestamp: timestamppb.New(latestTimestamp),
	}, nil
}

func (m *medianContractServer) LatestRoundRequested(ctx context.Context, request *pb.LatestRoundRequestedRequest) (*pb.LatestRoundRequestedReply, error) {
	digest, epoch, round, err := m.impl.LatestRoundRequested(ctx, time.Duration(request.Lookback))
	if err != nil {
		return nil, err
	}

	return &pb.LatestRoundRequestedReply{
		ConfigDigest: digest[:],
		Epoch:        epoch,
		Round:        uint32(round),
	}, nil
}

var _ median.OnchainConfigCodec = (*onchainConfigCodecClient)(nil)

type onchainConfigCodecClient struct {
	*net.BrokerExt
	grpc pb.OnchainConfigCodecClient
}

func (o *onchainConfigCodecClient) Encode(ctx context.Context, config median.OnchainConfig) ([]byte, error) {
	req := &pb.EncodeRequest{OnchainConfig: &pb.OnchainConfig{
		Min: pb.NewBigIntFromInt(config.Min),
		Max: pb.NewBigIntFromInt(config.Max),
	}}
	reply, err := o.grpc.Encode(ctx, req)
	if err != nil {
		return nil, err
	}
	return reply.Encoded, nil
}

func (o *onchainConfigCodecClient) Decode(ctx context.Context, bytes []byte) (oc median.OnchainConfig, err error) {
	var reply *pb.DecodeReply
	reply, err = o.grpc.Decode(ctx, &pb.DecodeRequest{Encoded: bytes})
	if err != nil {
		return
	}
	oc.Min, oc.Max = reply.OnchainConfig.Min.Int(), reply.OnchainConfig.Max.Int()
	return
}

var _ pb.OnchainConfigCodecServer = (*onchainConfigCodecServer)(nil)

type onchainConfigCodecServer struct {
	pb.UnimplementedOnchainConfigCodecServer
	impl median.OnchainConfigCodec
}

func (o *onchainConfigCodecServer) Encode(ctx context.Context, request *pb.EncodeRequest) (*pb.EncodeReply, error) {
	min, max := request.OnchainConfig.Min.Int(), request.OnchainConfig.Max.Int()
	b, err := o.impl.Encode(ctx, median.OnchainConfig{Max: max, Min: min})
	if err != nil {
		return nil, err
	}
	return &pb.EncodeReply{Encoded: b}, nil
}

func (o *onchainConfigCodecServer) Decode(ctx context.Context, request *pb.DecodeRequest) (*pb.DecodeReply, error) {
	oc, err := o.impl.Decode(ctx, request.Encoded)
	if err != nil {
		return nil, err
	}
	return &pb.DecodeReply{OnchainConfig: &pb.OnchainConfig{
		Min: pb.NewBigIntFromInt(oc.Min),
		Max: pb.NewBigIntFromInt(oc.Max),
	}}, nil
}

type ProviderServer struct{}

func (m ProviderServer) ConnToProvider(conn grpc.ClientConnInterface, broker net.Broker, brokerCfg net.BrokerConfig) types.MedianProvider {
	be := &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}
	pc := NewProviderClient(be, conn)
	pc.RmUnimplemented(context.Background())
	return pc
}

func RegisterProviderServices(s *grpc.Server, provider types.MedianProvider) {
	ocr2.RegisterPluginProviderServices(s, provider)
	pb.RegisterReportCodecServer(s, &reportCodecServer{impl: provider.ReportCodec()})
	pb.RegisterMedianContractServer(s, &medianContractServer{impl: provider.MedianContract()})
	pb.RegisterOnchainConfigCodecServer(s, &onchainConfigCodecServer{impl: provider.OnchainConfigCodec()})
}
