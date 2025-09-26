package ccipocr3

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

// ChainSpecificAddressCodec client
var _ ccipocr3.ChainSpecificAddressCodec = (*chainSpecificAddressCodecClient)(nil)

type chainSpecificAddressCodecClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.ChainSpecificAddressCodecClient
}

func NewChainSpecificAddressCodecClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.ChainSpecificAddressCodec {
	return &chainSpecificAddressCodecClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewChainSpecificAddressCodecClient(cc),
	}
}

func (c *chainSpecificAddressCodecClient) AddressBytesToString(address []byte) (string, error) {
	resp, err := c.grpc.AddressBytesToString(context.Background(), &ccipocr3pb.AddressBytesToStringRequest{
		Address: address,
	})
	if err != nil {
		return "", err
	}
	return resp.AddressString, nil
}

func (c *chainSpecificAddressCodecClient) AddressStringToBytes(addressString string) ([]byte, error) {
	resp, err := c.grpc.AddressStringToBytes(context.Background(), &ccipocr3pb.AddressStringToBytesRequest{
		AddressString: addressString,
	})
	if err != nil {
		return nil, err
	}
	return resp.Address, nil
}

func (c *chainSpecificAddressCodecClient) OracleIDAsAddressBytes(oracleID uint8) ([]byte, error) {
	resp, err := c.grpc.OracleIDAsAddressBytes(context.Background(), &ccipocr3pb.OracleIDAsAddressBytesRequest{
		OracleId: uint32(oracleID),
	})
	if err != nil {
		return nil, err
	}
	return resp.Address, nil
}

func (c *chainSpecificAddressCodecClient) TransmitterBytesToString(transmitter []byte) (string, error) {
	resp, err := c.grpc.TransmitterBytesToString(context.Background(), &ccipocr3pb.TransmitterBytesToStringRequest{
		Transmitter: transmitter,
	})
	if err != nil {
		return "", err
	}
	return resp.TransmitterString, nil
}

// ChainSpecificAddressCodec server
var _ ccipocr3pb.ChainSpecificAddressCodecServer = (*chainSpecificAddressCodecServer)(nil)

type chainSpecificAddressCodecServer struct {
	ccipocr3pb.UnimplementedChainSpecificAddressCodecServer
	impl ccipocr3.ChainSpecificAddressCodec
}

func NewChainSpecificAddressCodecServer(impl ccipocr3.ChainSpecificAddressCodec) *chainSpecificAddressCodecServer {
	return &chainSpecificAddressCodecServer{impl: impl}
}

func (s *chainSpecificAddressCodecServer) AddressBytesToString(ctx context.Context, req *ccipocr3pb.AddressBytesToStringRequest) (*ccipocr3pb.AddressBytesToStringResponse, error) {
	addressString, err := s.impl.AddressBytesToString(req.Address)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.AddressBytesToStringResponse{
		AddressString: addressString,
	}, nil
}

func (s *chainSpecificAddressCodecServer) AddressStringToBytes(ctx context.Context, req *ccipocr3pb.AddressStringToBytesRequest) (*ccipocr3pb.AddressStringToBytesResponse, error) {
	address, err := s.impl.AddressStringToBytes(req.AddressString)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.AddressStringToBytesResponse{
		Address: address,
	}, nil
}

func (s *chainSpecificAddressCodecServer) OracleIDAsAddressBytes(ctx context.Context, req *ccipocr3pb.OracleIDAsAddressBytesRequest) (*ccipocr3pb.OracleIDAsAddressBytesResponse, error) {
	address, err := s.impl.OracleIDAsAddressBytes(uint8(req.OracleId))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.OracleIDAsAddressBytesResponse{
		Address: address,
	}, nil
}

func (s *chainSpecificAddressCodecServer) TransmitterBytesToString(ctx context.Context, req *ccipocr3pb.TransmitterBytesToStringRequest) (*ccipocr3pb.TransmitterBytesToStringResponse, error) {
	transmitterString, err := s.impl.TransmitterBytesToString(req.Transmitter)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.TransmitterBytesToStringResponse{
		TransmitterString: transmitterString,
	}, nil
}

var _ ccipocr3.CommitPluginCodec = (*commitPluginCodecClient)(nil)

type commitPluginCodecClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.CommitPluginCodecClient
}

func NewCommitPluginCodecClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.CommitPluginCodec {
	return &commitPluginCodecClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewCommitPluginCodecClient(cc),
	}
}

func (c *commitPluginCodecClient) Encode(ctx context.Context, report ccipocr3.CommitPluginReport) ([]byte, error) {
	resp, err := c.grpc.Encode(ctx, &ccipocr3pb.EncodeCommitPluginReportInput{
		Report: commitPluginReportToPb(report),
	})
	if err != nil {
		return nil, err
	}
	return resp.EncodedReport, nil
}

func (c *commitPluginCodecClient) Decode(ctx context.Context, encodedReport []byte) (ccipocr3.CommitPluginReport, error) {
	resp, err := c.grpc.Decode(ctx, &ccipocr3pb.DecodeCommitPluginReportInput{
		EncodedReport: encodedReport,
	})
	if err != nil {
		return ccipocr3.CommitPluginReport{}, err
	}

	// Convert the protobuf response to Go struct
	return pbToCommitPluginReportDetailed(resp.Report), nil
}

// Helper conversion functions for CommitPluginReport
func commitPluginReportToPb(report ccipocr3.CommitPluginReport) *ccipocr3pb.CommitPluginReport {
	pbReport := &ccipocr3pb.CommitPluginReport{}

	// Convert PriceUpdates
	if len(report.PriceUpdates.TokenPriceUpdates) > 0 || len(report.PriceUpdates.GasPriceUpdates) > 0 {
		pbReport.PriceUpdates = &ccipocr3pb.PriceUpdates{}

		// Convert token price updates
		for _, tokenUpdate := range report.PriceUpdates.TokenPriceUpdates {
			pbReport.PriceUpdates.TokenPriceUpdates = append(pbReport.PriceUpdates.TokenPriceUpdates, &ccipocr3pb.TokenAmount{
				Token:  string(tokenUpdate.TokenID),
				Amount: intToPbBigInt(tokenUpdate.Price.Int),
			})
		}

		// Convert gas price updates
		for _, gasUpdate := range report.PriceUpdates.GasPriceUpdates {
			pbReport.PriceUpdates.GasPriceUpdates = append(pbReport.PriceUpdates.GasPriceUpdates, &ccipocr3pb.GasPriceChain{
				ChainSelector: uint64(gasUpdate.ChainSel),
				Price:         intToPbBigInt(gasUpdate.GasPrice.Int),
			})
		}
	}

	// Convert BlessedMerkleRoots
	for _, merkleRoot := range report.BlessedMerkleRoots {
		pbReport.BlessedMerkleRoots = append(pbReport.BlessedMerkleRoots, &ccipocr3pb.MerkleRootChain{
			ChainSelector: uint64(merkleRoot.ChainSel),
			MerkleRoot:    merkleRoot.MerkleRoot[:],
			SeqNumRange: &ccipocr3pb.SeqNumRange{
				Start: uint64(merkleRoot.SeqNumsRange.Start()),
				End:   uint64(merkleRoot.SeqNumsRange.End()),
			},
			OnRampAddress: merkleRoot.OnRampAddress,
		})
	}

	// Convert UnblessedMerkleRoots
	for _, merkleRoot := range report.UnblessedMerkleRoots {
		pbReport.UnblessedMerkleRoots = append(pbReport.UnblessedMerkleRoots, &ccipocr3pb.MerkleRootChain{
			ChainSelector: uint64(merkleRoot.ChainSel),
			MerkleRoot:    merkleRoot.MerkleRoot[:],
			SeqNumRange: &ccipocr3pb.SeqNumRange{
				Start: uint64(merkleRoot.SeqNumsRange.Start()),
				End:   uint64(merkleRoot.SeqNumsRange.End()),
			},
			OnRampAddress: merkleRoot.OnRampAddress,
		})
	}

	// Convert RMN signatures
	for _, rmnSig := range report.RMNSignatures {
		pbReport.RmnSignatures = append(pbReport.RmnSignatures, &ccipocr3pb.RMNECDSASignature{
			R: rmnSig.R[:],
			S: rmnSig.S[:],
		})
	}

	return pbReport
}

func pbToCommitPluginReportDetailed(pb *ccipocr3pb.CommitPluginReport) ccipocr3.CommitPluginReport {
	report := ccipocr3.CommitPluginReport{}

	// Convert PriceUpdates
	if pb.PriceUpdates != nil {
		// Convert token price updates
		for _, pbTokenUpdate := range pb.PriceUpdates.TokenPriceUpdates {
			report.PriceUpdates.TokenPriceUpdates = append(report.PriceUpdates.TokenPriceUpdates, ccipocr3.TokenPrice{
				TokenID: ccipocr3.UnknownEncodedAddress(pbTokenUpdate.Token),
				Price:   pbToBigInt(pbTokenUpdate.Amount),
			})
		}

		// Convert gas price updates
		for _, pbGasUpdate := range pb.PriceUpdates.GasPriceUpdates {
			report.PriceUpdates.GasPriceUpdates = append(report.PriceUpdates.GasPriceUpdates, ccipocr3.GasPriceChain{
				ChainSel: ccipocr3.ChainSelector(pbGasUpdate.ChainSelector),
				GasPrice: pbToBigInt(pbGasUpdate.Price),
			})
		}
	}

	// Convert BlessedMerkleRoots
	for _, pbMerkleRoot := range pb.BlessedMerkleRoots {
		var merkleRoot ccipocr3.Bytes32
		copy(merkleRoot[:], pbMerkleRoot.MerkleRoot)

		report.BlessedMerkleRoots = append(report.BlessedMerkleRoots, ccipocr3.MerkleRootChain{
			ChainSel:   ccipocr3.ChainSelector(pbMerkleRoot.ChainSelector),
			MerkleRoot: merkleRoot,
			SeqNumsRange: ccipocr3.NewSeqNumRange(
				ccipocr3.SeqNum(pbMerkleRoot.SeqNumRange.Start),
				ccipocr3.SeqNum(pbMerkleRoot.SeqNumRange.End),
			),
			OnRampAddress: pbMerkleRoot.OnRampAddress,
		})
	}

	// Convert UnblessedMerkleRoots
	for _, pbMerkleRoot := range pb.UnblessedMerkleRoots {
		var merkleRoot ccipocr3.Bytes32
		copy(merkleRoot[:], pbMerkleRoot.MerkleRoot)

		report.UnblessedMerkleRoots = append(report.UnblessedMerkleRoots, ccipocr3.MerkleRootChain{
			ChainSel:   ccipocr3.ChainSelector(pbMerkleRoot.ChainSelector),
			MerkleRoot: merkleRoot,
			SeqNumsRange: ccipocr3.NewSeqNumRange(
				ccipocr3.SeqNum(pbMerkleRoot.SeqNumRange.Start),
				ccipocr3.SeqNum(pbMerkleRoot.SeqNumRange.End),
			),
			OnRampAddress: pbMerkleRoot.OnRampAddress,
		})
	}

	// Convert RMN signatures
	for _, pbRmnSig := range pb.RmnSignatures {
		var r, s ccipocr3.Bytes32
		copy(r[:], pbRmnSig.R)
		copy(s[:], pbRmnSig.S)

		report.RMNSignatures = append(report.RMNSignatures, ccipocr3.RMNECDSASignature{
			R: r,
			S: s,
		})
	}

	return report
}

// CommitPluginCodec server
var _ ccipocr3pb.CommitPluginCodecServer = (*commitPluginCodecServer)(nil)

type commitPluginCodecServer struct {
	ccipocr3pb.UnimplementedCommitPluginCodecServer
	impl ccipocr3.CommitPluginCodec
}

func NewCommitPluginCodecServer(impl ccipocr3.CommitPluginCodec) ccipocr3pb.CommitPluginCodecServer {
	return &commitPluginCodecServer{impl: impl}
}

func (s *commitPluginCodecServer) Encode(ctx context.Context, req *ccipocr3pb.EncodeCommitPluginReportInput) (*ccipocr3pb.EncodeCommitPluginReportOutput, error) {
	report := pbToCommitPluginReportDetailed(req.Report)
	encodedReport, err := s.impl.Encode(ctx, report)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.EncodeCommitPluginReportOutput{
		EncodedReport: encodedReport,
	}, nil
}

func (s *commitPluginCodecServer) Decode(ctx context.Context, req *ccipocr3pb.DecodeCommitPluginReportInput) (*ccipocr3pb.DecodeCommitPluginReportOutput, error) {
	report, err := s.impl.Decode(ctx, req.EncodedReport)
	if err != nil {
		return nil, err
	}

	return &ccipocr3pb.DecodeCommitPluginReportOutput{
		Report: commitPluginReportToPb(report),
	}, nil
}

// ExecutePluginCodec client
var _ ccipocr3.ExecutePluginCodec = (*executePluginCodecClient)(nil)

type executePluginCodecClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.ExecutePluginCodecClient
}

func NewExecutePluginCodecClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.ExecutePluginCodec {
	return &executePluginCodecClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewExecutePluginCodecClient(cc),
	}
}

func (c *executePluginCodecClient) Encode(ctx context.Context, report ccipocr3.ExecutePluginReport) ([]byte, error) {
	resp, err := c.grpc.Encode(ctx, &ccipocr3pb.EncodeExecutePluginReportRequest{
		Report: executePluginReportToPb(report),
	})
	if err != nil {
		return nil, err
	}
	return resp.EncodedReport, nil
}

func (c *executePluginCodecClient) Decode(ctx context.Context, encodedReport []byte) (ccipocr3.ExecutePluginReport, error) {
	resp, err := c.grpc.Decode(ctx, &ccipocr3pb.DecodeExecutePluginReportRequest{
		EncodedReport: encodedReport,
	})
	if err != nil {
		return ccipocr3.ExecutePluginReport{}, err
	}
	return pbToExecutePluginReport(resp.Report), nil
}

// ExecutePluginCodec server
var _ ccipocr3pb.ExecutePluginCodecServer = (*executePluginCodecServer)(nil)

type executePluginCodecServer struct {
	ccipocr3pb.UnimplementedExecutePluginCodecServer
	impl ccipocr3.ExecutePluginCodec
}

func NewExecutePluginCodecServer(impl ccipocr3.ExecutePluginCodec) ccipocr3pb.ExecutePluginCodecServer {
	return &executePluginCodecServer{impl: impl}
}

func (s *executePluginCodecServer) Encode(ctx context.Context, req *ccipocr3pb.EncodeExecutePluginReportRequest) (*ccipocr3pb.EncodeExecutePluginReportResponse, error) {
	report := pbToExecutePluginReport(req.Report)
	encodedReport, err := s.impl.Encode(ctx, report)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.EncodeExecutePluginReportResponse{
		EncodedReport: encodedReport,
	}, nil
}

func (s *executePluginCodecServer) Decode(ctx context.Context, req *ccipocr3pb.DecodeExecutePluginReportRequest) (*ccipocr3pb.DecodeExecutePluginReportResponse, error) {
	report, err := s.impl.Decode(ctx, req.EncodedReport)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.DecodeExecutePluginReportResponse{
		Report: executePluginReportToPb(report),
	}, nil
}

// TokenDataEncoder client
var _ ccipocr3.TokenDataEncoder = (*tokenDataEncoderClient)(nil)

type tokenDataEncoderClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.TokenDataEncoderClient
}

func NewTokenDataEncoderClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.TokenDataEncoder {
	return &tokenDataEncoderClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewTokenDataEncoderClient(cc),
	}
}

func (c *tokenDataEncoderClient) EncodeUSDC(ctx context.Context, message ccipocr3.Bytes, attestation ccipocr3.Bytes) (ccipocr3.Bytes, error) {
	resp, err := c.grpc.EncodeUSDC(ctx, &ccipocr3pb.EncodeUSDCRequest{
		Message:     message,
		Attestation: attestation,
	})
	if err != nil {
		return nil, err
	}
	return ccipocr3.Bytes(resp.EncodedData), nil
}

// TokenDataEncoder server
var _ ccipocr3pb.TokenDataEncoderServer = (*tokenDataEncoderServer)(nil)

type tokenDataEncoderServer struct {
	ccipocr3pb.UnimplementedTokenDataEncoderServer
	impl ccipocr3.TokenDataEncoder
}

func NewTokenDataEncoderServer(impl ccipocr3.TokenDataEncoder) ccipocr3pb.TokenDataEncoderServer {
	return &tokenDataEncoderServer{impl: impl}
}

func (s *tokenDataEncoderServer) EncodeUSDC(ctx context.Context, req *ccipocr3pb.EncodeUSDCRequest) (*ccipocr3pb.EncodeUSDCResponse, error) {
	encodedData, err := s.impl.EncodeUSDC(ctx, ccipocr3.Bytes(req.Message), ccipocr3.Bytes(req.Attestation))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.EncodeUSDCResponse{
		EncodedData: encodedData,
	}, nil
}

// SourceChainExtraDataCodec client
var _ ccipocr3.SourceChainExtraDataCodec = (*sourceChainExtraDataCodecClient)(nil)

type sourceChainExtraDataCodecClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.SourceChainExtraDataCodecClient
}

func NewSourceChainExtraDataCodecClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.SourceChainExtraDataCodec {
	return &sourceChainExtraDataCodecClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewSourceChainExtraDataCodecClient(cc),
	}
}

func (c *sourceChainExtraDataCodecClient) DecodeExtraArgsToMap(extraArgs ccipocr3.Bytes) (map[string]any, error) {
	resp, err := c.grpc.DecodeExtraArgsToMap(context.Background(), &ccipocr3pb.DecodeExtraArgsToMapRequest{
		ExtraArgs: extraArgs,
	})
	if err != nil {
		return nil, err
	}
	return pbMapToGoMap(resp.DecodedMap)
}

func (c *sourceChainExtraDataCodecClient) DecodeDestExecDataToMap(destExecData ccipocr3.Bytes) (map[string]any, error) {
	resp, err := c.grpc.DecodeDestExecDataToMap(context.Background(), &ccipocr3pb.DecodeDestExecDataToMapRequest{
		DestExecData: destExecData,
	})
	if err != nil {
		return nil, err
	}
	return pbMapToGoMap(resp.DecodedMap)
}

// SourceChainExtraDataCodec server
var _ ccipocr3pb.SourceChainExtraDataCodecServer = (*sourceChainExtraDataCodecServer)(nil)

type sourceChainExtraDataCodecServer struct {
	ccipocr3pb.UnimplementedSourceChainExtraDataCodecServer
	impl ccipocr3.SourceChainExtraDataCodec
}

func NewSourceChainExtraDataCodecServer(impl ccipocr3.SourceChainExtraDataCodec) ccipocr3pb.SourceChainExtraDataCodecServer {
	return &sourceChainExtraDataCodecServer{impl: impl}
}

func (s *sourceChainExtraDataCodecServer) DecodeExtraArgsToMap(ctx context.Context, req *ccipocr3pb.DecodeExtraArgsToMapRequest) (*ccipocr3pb.DecodeExtraArgsToMapResponse, error) {
	decodedMap, err := s.impl.DecodeExtraArgsToMap(ccipocr3.Bytes(req.ExtraArgs))
	if err != nil {
		return nil, err
	}
	pbMap, err := goMapToPbMap(decodedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decoded map to protobuf: %w", err)
	}
	return &ccipocr3pb.DecodeExtraArgsToMapResponse{
		DecodedMap: pbMap,
	}, nil
}

func (s *sourceChainExtraDataCodecServer) DecodeDestExecDataToMap(ctx context.Context, req *ccipocr3pb.DecodeDestExecDataToMapRequest) (*ccipocr3pb.DecodeDestExecDataToMapResponse, error) {
	decodedMap, err := s.impl.DecodeDestExecDataToMap(ccipocr3.Bytes(req.DestExecData))
	if err != nil {
		return nil, err
	}
	pbMap, err := goMapToPbMap(decodedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decoded map to protobuf: %w", err)
	}
	return &ccipocr3pb.DecodeDestExecDataToMapResponse{
		DecodedMap: pbMap,
	}, nil
}

// Helper conversion functions
func executePluginReportToPb(report ccipocr3.ExecutePluginReport) *ccipocr3pb.ExecutePluginReport {
	var chainReports []*ccipocr3pb.ChainReport
	for _, cr := range report.ChainReports {
		pbChainReport := &ccipocr3pb.ChainReport{
			SourceChainSelector: uint64(cr.SourceChainSelector),
			ProofFlagBits:       intToPbBigInt(cr.ProofFlagBits.Int),
		}

		// Convert messages
		for _, msg := range cr.Messages {
			pbChainReport.Messages = append(pbChainReport.Messages, messageToPb(msg))
		}

		// Convert offchain token data - maintain 3D structure with MessageOffchainTokenData
		for _, messageTokenData := range cr.OffchainTokenData {
			pbMessageTokenData := &ccipocr3pb.MessageOffchainTokenData{
				TokenData: messageTokenData,
			}
			pbChainReport.OffchainTokenData = append(pbChainReport.OffchainTokenData, pbMessageTokenData)
		}

		// Convert proofs
		for _, proof := range cr.Proofs {
			pbChainReport.Proofs = append(pbChainReport.Proofs, proof[:])
		}

		chainReports = append(chainReports, pbChainReport)
	}

	return &ccipocr3pb.ExecutePluginReport{
		ChainReports: chainReports,
	}
}

func pbToExecutePluginReport(pb *ccipocr3pb.ExecutePluginReport) ccipocr3.ExecutePluginReport {
	var chainReports []ccipocr3.ExecutePluginReportSingleChain

	// Initialize as empty slice to match test expectations (not nil)
	if len(pb.ChainReports) == 0 {
		chainReports = []ccipocr3.ExecutePluginReportSingleChain{}
	}

	for _, pbCr := range pb.ChainReports {
		chainReport := ccipocr3.ExecutePluginReportSingleChain{
			SourceChainSelector: ccipocr3.ChainSelector(pbCr.SourceChainSelector),
			ProofFlagBits:       pbToBigInt(pbCr.ProofFlagBits),
		}

		// Convert messages
		if len(pbCr.Messages) > 0 {
			for _, pbMsg := range pbCr.Messages {
				chainReport.Messages = append(chainReport.Messages, pbToMessage(pbMsg))
			}
		} else {
			chainReport.Messages = []ccipocr3.Message{}
		}

		// Convert offchain token data - restore 3D structure from MessageOffchainTokenData
		if len(pbCr.OffchainTokenData) > 0 {
			for _, pbMessageTokenData := range pbCr.OffchainTokenData {
				chainReport.OffchainTokenData = append(chainReport.OffchainTokenData, pbMessageTokenData.TokenData)
			}
		} else {
			chainReport.OffchainTokenData = [][][]byte{}
		}

		// Convert proofs
		for _, proof := range pbCr.Proofs {
			chainReport.Proofs = append(chainReport.Proofs, ccipocr3.Bytes32(proof))
		}

		chainReports = append(chainReports, chainReport)
	}

	return ccipocr3.ExecutePluginReport{
		ChainReports: chainReports,
	}
}

// ExtraDataCodecBundle client
var _ ccipocr3.ExtraDataCodecBundle = (*extraDataCodecBundleClient)(nil)

type extraDataCodecBundleClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.ExtraDataCodecBundleClient
}

func NewExtraDataCodecBundleClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.ExtraDataCodecBundle {
	return &extraDataCodecBundleClient{
		BrokerExt: broker,
		grpc:      ccipocr3pb.NewExtraDataCodecBundleClient(cc),
	}
}

func (c *extraDataCodecBundleClient) DecodeExtraArgs(extraArgs ccipocr3.Bytes, sourceChainSelector ccipocr3.ChainSelector) (map[string]any, error) {
	resp, err := c.grpc.DecodeExtraArgs(context.Background(), &ccipocr3pb.DecodeExtraArgsWithChainSelectorRequest{
		ExtraArgs:           extraArgs,
		SourceChainSelector: uint64(sourceChainSelector),
	})
	if err != nil {
		return nil, err
	}
	return pbMapToGoMap(resp.DecodedMap)
}

func (c *extraDataCodecBundleClient) DecodeTokenAmountDestExecData(destExecData ccipocr3.Bytes, sourceChainSelector ccipocr3.ChainSelector) (map[string]any, error) {
	resp, err := c.grpc.DecodeTokenAmountDestExecData(context.Background(), &ccipocr3pb.DecodeTokenAmountDestExecDataRequest{
		DestExecData:        destExecData,
		SourceChainSelector: uint64(sourceChainSelector),
	})
	if err != nil {
		return nil, err
	}
	return pbMapToGoMap(resp.DecodedMap)
}

// ExtraDataCodecBundle server
var _ ccipocr3pb.ExtraDataCodecBundleServer = (*extraDataCodecBundleServer)(nil)

type extraDataCodecBundleServer struct {
	ccipocr3pb.UnimplementedExtraDataCodecBundleServer
	impl ccipocr3.ExtraDataCodecBundle
}

func NewExtraDataCodecBundleServer(impl ccipocr3.ExtraDataCodecBundle) ccipocr3pb.ExtraDataCodecBundleServer {
	return &extraDataCodecBundleServer{impl: impl}
}

func (s *extraDataCodecBundleServer) DecodeExtraArgs(ctx context.Context, req *ccipocr3pb.DecodeExtraArgsWithChainSelectorRequest) (*ccipocr3pb.DecodeExtraArgsWithChainSelectorResponse, error) {
	decodedMap, err := s.impl.DecodeExtraArgs(req.ExtraArgs, ccipocr3.ChainSelector(req.SourceChainSelector))
	if err != nil {
		return nil, err
	}
	pbMap, err := goMapToPbMap(decodedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decoded map to protobuf: %w", err)
	}
	return &ccipocr3pb.DecodeExtraArgsWithChainSelectorResponse{
		DecodedMap: pbMap,
	}, nil
}

func (s *extraDataCodecBundleServer) DecodeTokenAmountDestExecData(ctx context.Context, req *ccipocr3pb.DecodeTokenAmountDestExecDataRequest) (*ccipocr3pb.DecodeTokenAmountDestExecDataResponse, error) {
	decodedMap, err := s.impl.DecodeTokenAmountDestExecData(req.DestExecData, ccipocr3.ChainSelector(req.SourceChainSelector))
	if err != nil {
		return nil, err
	}
	pbMap, err := goMapToPbMap(decodedMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert decoded map to protobuf: %w", err)
	}
	return &ccipocr3pb.DecodeTokenAmountDestExecDataResponse{
		DecodedMap: pbMap,
	}, nil
}
