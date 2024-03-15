package ccip

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
)

// CommitStoreGRPCClient implements [cciptypes.CommitStoreReader] by wrapping a
// [ccippb.CommitStoreReaderGRPCClient] grpc client.
// It is used by a ReportingPlugin to call the CommitStoreReader service, which
// is hosted by the relayer
type CommitStoreGRPCClient struct {
	client ccippb.CommitStoreReaderClient
}

func NewCommitStoreGRPCClient(cc grpc.ClientConnInterface) *CommitStoreGRPCClient {
	return &CommitStoreGRPCClient{client: ccippb.NewCommitStoreReaderClient(cc)}
}

// CommitStoreGRPCServer implements [ccippb.CommitStoreReaderServer] by wrapping a
// [cciptypes.CommitStoreReader] implementation.
// This server is hosted by the relayer and is called ReportingPlugin via
// the [CommitStoreGRPCClient]
type CommitStoreGRPCServer struct {
	ccippb.UnimplementedCommitStoreReaderServer

	impl    ccip.CommitStoreReader
	onClose func() error
}

func NewCommitStoreGRPCServer(impl ccip.CommitStoreReader) *CommitStoreGRPCServer {
	return &CommitStoreGRPCServer{impl: impl}
}

// ensure the types are satisfied
var _ ccippb.CommitStoreReaderServer = (*CommitStoreGRPCServer)(nil)
var _ ccip.CommitStoreReader = (*CommitStoreGRPCClient)(nil)

// ChangeConfig implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) ChangeConfig(ctx context.Context, onchainConfig []byte, offchainConfig []byte) (ccip.Address, error) {
	resp, err := c.client.ChangeConfig(ctx, &ccippb.CommitStoreChangeConfigRequest{
		OnchainConfig:  onchainConfig,
		OffchainConfig: offchainConfig,
	})
	if err != nil {
		return ccip.Address(""), err
	}
	return ccip.Address(resp.Address), nil
}

// DecodeCommitReport implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) DecodeCommitReport(ctx context.Context, report []byte) (ccip.CommitStoreReport, error) {
	resp, err := c.client.DecodeCommitReport(ctx, &ccippb.DecodeCommitReportRequest{EncodedReport: report})
	if err != nil {
		return ccip.CommitStoreReport{}, err
	}
	return commitStoreReport(resp.Report)

}

// EncodeCommitReport implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) EncodeCommitReport(ctx context.Context, report ccip.CommitStoreReport) ([]byte, error) {
	pb, err := commitStoreReportPB(report)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.EncodeCommitReport(ctx, &ccippb.EncodeCommitReportRequest{Report: pb})
	if err != nil {
		return nil, err
	}
	return resp.EncodedReport, nil
}

// GasPriceEstimator implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GasPriceEstimator(ctx context.Context) (ccip.GasPriceEstimatorCommit, error) {
	panic("unimplemented: GasPriceEstimator BCF-2991")
}

// GetAcceptedCommitReportsGteTimestamp implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GetAcceptedCommitReportsGteTimestamp(ctx context.Context, ts time.Time, confirmations int) ([]ccip.CommitStoreReportWithTxMeta, error) {
	resp, err := c.client.GetAcceptedCommitReportsGteTimestamp(ctx, &ccippb.GetAcceptedCommitReportsGteTimestampRequest{
		Timestamp:     timestamppb.New(ts),
		Confirmations: uint64(confirmations),
	})
	if err != nil {
		return nil, err
	}
	return commitStoreReportWithTxMetaSlice(resp.Reports)
}

// GetCommitReportMatchingSeqNum implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GetCommitReportMatchingSeqNum(ctx context.Context, seqNum uint64, confirmations int) ([]ccip.CommitStoreReportWithTxMeta, error) {
	resp, err := c.client.GeteCommitReportMatchingSequenceNumber(ctx, &ccippb.GetCommitReportMatchingSequenceNumberRequest{
		SequenceNumber: seqNum,
		Confirmations:  uint64(confirmations),
	})
	if err != nil {
		return nil, err
	}
	return commitStoreReportWithTxMetaSlice(resp.Reports)
}

// GetCommitStoreStaticConfig implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GetCommitStoreStaticConfig(ctx context.Context) (ccip.CommitStoreStaticConfig, error) {
	resp, err := c.client.GetCommitStoreStaticConfig(ctx, &emptypb.Empty{})
	if err != nil {
		return ccip.CommitStoreStaticConfig{}, err
	}
	return ccip.CommitStoreStaticConfig{
		ChainSelector:       resp.StaticConfig.ChainSelector,
		SourceChainSelector: resp.StaticConfig.SourceChainSelector,
		OnRamp:              ccip.Address(resp.StaticConfig.OnRamp),
		ArmProxy:            ccip.Address(resp.StaticConfig.ArmProxy),
	}, nil
}

// GetExpectedNextSequenceNumber implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GetExpectedNextSequenceNumber(ctx context.Context) (uint64, error) {
	resp, err := c.client.GetExpectedNextSequenceNumber(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.SequenceNumber, nil
}

// GetLatestPriceEpochAndRound implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) GetLatestPriceEpochAndRound(ctx context.Context) (uint64, error) {
	resp, err := c.client.GetLatestPriceEpochAndRound(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return resp.EpochAndRound, nil
}

// IsBlessed implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) IsBlessed(ctx context.Context, root [32]byte) (bool, error) {
	resp, err := c.client.IsBlessed(ctx, &ccippb.IsBlessedRequest{Root: root[:]})
	if err != nil {
		return false, err
	}
	return resp.IsBlessed, nil
}

// IsDown implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) IsDown(ctx context.Context) (bool, error) {
	resp, err := c.client.IsDown(ctx, &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	return resp.IsDown, nil
}

// OffchainConfig implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) OffchainConfig(ctx context.Context) (ccip.CommitOffchainConfig, error) {
	resp, err := c.client.GetOffchainConfig(ctx, &emptypb.Empty{})
	if err != nil {
		return ccip.CommitOffchainConfig{}, err
	}
	return ccip.CommitOffchainConfig{
		GasPriceDeviationPPB:   resp.OffchainConfig.GasPriceDeviationPpb,
		GasPriceHeartBeat:      resp.OffchainConfig.GasPriceHeartbeat.AsDuration(),
		TokenPriceDeviationPPB: resp.OffchainConfig.TokenPriceDeviationPpb,
		TokenPriceHeartBeat:    resp.OffchainConfig.TokenPriceHeartbeat.AsDuration(),
		InflightCacheExpiry:    resp.OffchainConfig.InflightCacheExpiry.AsDuration(),
	}, nil
}

// VerifyExecutionReport implements ccip.CommitStoreReader.
func (c *CommitStoreGRPCClient) VerifyExecutionReport(ctx context.Context, report ccip.ExecReport) (bool, error) {
	resp, err := c.client.VerifyExecutionReport(ctx, &ccippb.VerifyExecutionReportRequest{Report: executionReportPB(report)})
	if err != nil {
		return false, err
	}
	return resp.IsValid, nil
}

// Server implementation

// ChangeConfig implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) ChangeConfig(ctx context.Context, req *ccippb.CommitStoreChangeConfigRequest) (*ccippb.CommitStoreChangeConfigResponse, error) {
	addr, err := c.impl.ChangeConfig(ctx, req.OnchainConfig, req.OffchainConfig)
	if err != nil {
		return nil, err
	}
	return &ccippb.CommitStoreChangeConfigResponse{Address: string(addr)}, nil
}

// DecodeCommitReport implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) DecodeCommitReport(ctx context.Context, req *ccippb.DecodeCommitReportRequest) (*ccippb.DecodeCommitReportResponse, error) {
	r, err := c.impl.DecodeCommitReport(ctx, req.EncodedReport)
	if err != nil {
		return nil, err
	}
	pb, err := commitStoreReportPB(r)
	if err != nil {
		return nil, err
	}
	return &ccippb.DecodeCommitReportResponse{Report: pb}, nil
}

// EncodeCommitReport implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) EncodeCommitReport(ctx context.Context, req *ccippb.EncodeCommitReportRequest) (*ccippb.EncodeCommitReportResponse, error) {
	r, err := commitStoreReport(req.Report)
	if err != nil {
		return nil, err
	}
	encoded, err := c.impl.EncodeCommitReport(ctx, r)
	if err != nil {
		return nil, err
	}
	return &ccippb.EncodeCommitReportResponse{EncodedReport: encoded}, nil
}

// GetAcceptedCommitReportsGteTimestamp implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetAcceptedCommitReportsGteTimestamp(ctx context.Context, req *ccippb.GetAcceptedCommitReportsGteTimestampRequest) (*ccippb.GetAcceptedCommitReportsGteTimestampResponse, error) {
	reports, err := c.impl.GetAcceptedCommitReportsGteTimestamp(ctx, req.Timestamp.AsTime(), int(req.Confirmations))
	if err != nil {
		return nil, err
	}
	pbReports, err := commitStoreReportWithTxMetaPBSlice(reports)
	if err != nil {
		return nil, err
	}
	return &ccippb.GetAcceptedCommitReportsGteTimestampResponse{Reports: pbReports}, nil
}

// GetCommitGasPriceEstimator implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetCommitGasPriceEstimator(ctx context.Context, req *emptypb.Empty) (*ccippb.GetCommitGasPriceEstimatorResponse, error) {
	panic("unimplemented")
}

// GetCommitStoreStaticConfig implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetCommitStoreStaticConfig(ctx context.Context, req *emptypb.Empty) (*ccippb.GetCommitStoreStaticConfigResponse, error) {
	panic("unimplemented")
}

// GetExpectedNextSequenceNumber implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetExpectedNextSequenceNumber(ctx context.Context, req *emptypb.Empty) (*ccippb.GetExpectedNextSequenceNumberResponse, error) {
	panic("unimplemented")
}

// GetLatestPriceEpochAndRound implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetLatestPriceEpochAndRound(ctx context.Context, req *emptypb.Empty) (*ccippb.GetLatestPriceEpochAndRoundResponse, error) {
	panic("unimplemented")
}

// GetOffchainConfig implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GetOffchainConfig(ctx context.Context, req *emptypb.Empty) (*ccippb.GetOffchainConfigResponse, error) {
	panic("unimplemented")
}

// GeteCommitReportMatchingSequenceNumber implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) GeteCommitReportMatchingSequenceNumber(ctx context.Context, req *ccippb.GetCommitReportMatchingSequenceNumberRequest) (*ccippb.GetCommitReportMatchingSequenceNumberResponse, error) {
	panic("unimplemented")
}

// IsBlessed implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) IsBlessed(ctx context.Context, req *ccippb.IsBlessedRequest) (*ccippb.IsBlessedResponse, error) {
	panic("unimplemented")
}

// IsDown implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) IsDown(ctx context.Context, req *emptypb.Empty) (*ccippb.IsDownResponse, error) {
	panic("unimplemented")
}

// VerifyExecutionReport implements ccippb.CommitStoreReaderServer.
func (c *CommitStoreGRPCServer) VerifyExecutionReport(ctx context.Context, req *ccippb.VerifyExecutionReportRequest) (*ccippb.VerifyExecutionReportResponse, error) {
	panic("unimplemented")
}

func commitStoreReport(pb *ccippb.CommitStoreReport) (ccip.CommitStoreReport, error) {
	root, err := merkleRoot(pb.MerkleRoot)
	if err != nil {
		return ccip.CommitStoreReport{}, fmt.Errorf("cannot convert merkle root: %w", err)
	}
	out := ccip.CommitStoreReport{
		TokenPrices: tokenPrices(pb.TokenPrices),
		GasPrices:   gasPrices(pb.GasPrices),
		Interval:    commitStoreInterval(pb.Interval),
		MerkleRoot:  root,
	}
	return out, nil
}

func tokenPrices(pb []*ccippb.TokenPrice) []ccip.TokenPrice {
	out := make([]ccip.TokenPrice, len(pb))
	for i, p := range pb {
		out[i] = tokenPrice(p)
	}
	return out
}

func gasPrices(pb []*ccippb.GasPrice) []ccip.GasPrice {
	out := make([]ccip.GasPrice, len(pb))
	for i, p := range pb {
		out[i] = gasPrice(p)
	}
	return out
}

func gasPrice(pb *ccippb.GasPrice) ccip.GasPrice {
	return ccip.GasPrice{
		DestChainSelector: pb.DestChainSelector,
		Value:             pb.Value.Int(),
	}
}

func commitStoreInterval(pb *ccippb.CommitStoreInterval) ccip.CommitStoreInterval {
	return ccip.CommitStoreInterval{
		Min: pb.Min,
		Max: pb.Max,
	}
}

func merkleRoot(pb []byte) ([32]byte, error) {
	if len(pb) != 32 {
		return [32]byte{}, fmt.Errorf("expected 32 bytes, got %d", len(pb))
	}
	var out [32]byte
	copy(out[:], pb)
	return out, nil
}

func commitStoreReportPB(r ccip.CommitStoreReport) (*ccippb.CommitStoreReport, error) {
	if len(r.MerkleRoot) != 32 {
		return nil, fmt.Errorf("invalid merkle root: expected 32 bytes, got %d", len(r.MerkleRoot))
	}
	pb := &ccippb.CommitStoreReport{
		TokenPrices: tokenPricesPB(r.TokenPrices),
		GasPrices:   gasPricesPB(r.GasPrices),
		Interval:    commitStoreIntervalPB(r.Interval),
		MerkleRoot:  r.MerkleRoot[:],
	}
	return pb, nil
}

func tokenPricesPB(r []ccip.TokenPrice) []*ccippb.TokenPrice {
	out := make([]*ccippb.TokenPrice, len(r))
	for i, p := range r {
		out[i] = tokenPricePB(p)
	}
	return out
}

func gasPricesPB(r []ccip.GasPrice) []*ccippb.GasPrice {
	out := make([]*ccippb.GasPrice, len(r))
	for i, p := range r {
		out[i] = gasPricePB(p)
	}
	return out
}

func commitStoreIntervalPB(r ccip.CommitStoreInterval) *ccippb.CommitStoreInterval {
	return &ccippb.CommitStoreInterval{
		Min: r.Min,
		Max: r.Max,
	}
}

func commitStoreReportWithTxMetaSlice(pb []*ccippb.CommitStoreReportWithTxMeta) ([]ccip.CommitStoreReportWithTxMeta, error) {
	out := make([]ccip.CommitStoreReportWithTxMeta, len(pb))
	var err error
	for i, p := range pb {
		out[i], err = commitStoreReportWithTxMeta(p)
		if err != nil {
			return nil, fmt.Errorf("cannot convert commit store report with tx meta: %w", err)
		}
	}
	return out, nil
}

func commitStoreReportWithTxMeta(pb *ccippb.CommitStoreReportWithTxMeta) (ccip.CommitStoreReportWithTxMeta, error) {
	r, err := commitStoreReport(pb.Report)
	if err != nil {
		return ccip.CommitStoreReportWithTxMeta{}, fmt.Errorf("cannot convert commit store report: %w", err)
	}
	return ccip.CommitStoreReportWithTxMeta{
		TxMeta:            txMeta(pb.TxMeta),
		CommitStoreReport: r,
	}, nil
}

func commitStoreReportWithTxMetaPBSlice(r []ccip.CommitStoreReportWithTxMeta) ([]*ccippb.CommitStoreReportWithTxMeta, error) {
	out := make([]*ccippb.CommitStoreReportWithTxMeta, len(r))
	var err error
	for i, p := range r {
		out[i], err = commitStoreReportWithTxMetaPB(p)
		if err != nil {
			return nil, fmt.Errorf("cannot convert commit store report %v at %d with tx meta: %w", p, i, err)
		}
	}
	return out, nil
}

func commitStoreReportWithTxMetaPB(r ccip.CommitStoreReportWithTxMeta) (*ccippb.CommitStoreReportWithTxMeta, error) {
	report, err := commitStoreReportPB(r.CommitStoreReport)
	if err != nil {
		return nil, fmt.Errorf("cannot convert commit store report: %w", err)
	}
	return &ccippb.CommitStoreReportWithTxMeta{
		TxMeta: txMetaPB(r.TxMeta),
		Report: report,
	}, nil
}
