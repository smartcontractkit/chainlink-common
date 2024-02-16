package ccip

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	cciptypes "github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
)

var _ cciptypes.OffRampReader = (*OffRampReaderClient)(nil)

type OffRampReaderClient struct {
	grpc ccippb.OffRampReaderClient
}

// Address i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) Address(ctx context.Context) (cciptypes.Address, error) {
	resp, err := o.grpc.Address(context.TODO(), &emptypb.Empty{})
	if err != nil {
		return cciptypes.Address(""), err
	}
	return cciptypes.Address(resp.Address), nil
}

// ChangeConfig implements [github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) ChangeConfig(ctx context.Context, onchainConfig []byte, offchainConfig []byte) (cciptypes.Address, cciptypes.Address, error) {
	resp, err := o.grpc.ChangeConfig(ctx, &ccippb.ChangeConfigRequest{
		OnchainConfig:  onchainConfig,
		OffchainConfig: offchainConfig,
	})
	if err != nil {
		return cciptypes.Address(""), cciptypes.Address(""), err
	}

	return cciptypes.Address(resp.OnchainConfigAddress), cciptypes.Address(resp.OffchainConfigAddress), nil
}

// CurrentRateLimiterState i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) CurrentRateLimiterState(ctx context.Context) (cciptypes.TokenBucketRateLimit, error) {
	resp, err := o.grpc.CurrentRateLimiterState(ctx, &emptypb.Empty{})
	if err != nil {
		return cciptypes.TokenBucketRateLimit{}, err
	}
	return tokenBucketRateLimit(resp.RateLimiter), nil
}

// DecodeExecutionReport i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) DecodeExecutionReport(ctx context.Context, report []byte) (cciptypes.ExecReport, error) {
	resp, err := o.grpc.DecodeExecutionReport(ctx, &ccippb.DecodeExecutionReportRequest{
		Report: report,
	})
	if err != nil {
		return cciptypes.ExecReport{}, err
	}

	return execReport(resp.Report)
}

// EncodeExecutionReport i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) EncodeExecutionReport(ctx context.Context, report cciptypes.ExecReport) ([]byte, error) {
	reportPB := executionReportPB(report)

	resp, err := o.grpc.EncodeExecutionReport(ctx, &ccippb.EncodeExecutionReportRequest{
		Report: reportPB,
	})
	if err != nil {
		return nil, err
	}
	return resp.Report, nil

}

// GasPriceEstimator i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GasPriceEstimator(ctx context.Context) (cciptypes.GasPriceEstimatorExec, error) {
	panic("BCF-2991 implement gas price estimator grpc service")
}

// GetExecutionState i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetExecutionState(ctx context.Context, sequenceNumber uint64) (uint8, error) {
	resp, err := o.grpc.GetExecutionState(ctx, &ccippb.GetExecutionStateRequest{
		SeqNum: sequenceNumber,
	})
	if err != nil {
		return 0, err
	}
	return uint8(resp.ExecutionState), nil
}

// GetExecutionStateChangesBetweenSeqNums i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetExecutionStateChangesBetweenSeqNums(ctx context.Context, seqNumMin uint64, seqNumMax uint64, confirmations int) ([]cciptypes.ExecutionStateChangedWithTxMeta, error) {
	resp, err := o.grpc.GetExecutionStateChangesBetweenSeqNums(ctx, &ccippb.GetExecutionStateChangesBetweenSeqNumsRequest{
		SeqNumMin:     seqNumMin,
		SeqNumMax:     seqNumMax,
		Confirmations: int32(confirmations),
	})
	if err != nil {
		return nil, err
	}
	return executionStateChangedWithTxMetaSlice(resp.ExecutionStateChanges), nil
}

// GetSenderNonce i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetSenderNonce(ctx context.Context, sender cciptypes.Address) (uint64, error) {
	panic("unimplemented")
}

// GetSourceToDestTokensMapping i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetSourceToDestTokensMapping(ctx context.Context) (map[cciptypes.Address]cciptypes.Address, error) {
	panic("unimplemented")
}

// GetStaticConfig i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetStaticConfig(ctx context.Context) (cciptypes.OffRampStaticConfig, error) {
	panic("unimplemented")
}

// GetTokens i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) GetTokens(ctx context.Context) (cciptypes.OffRampTokens, error) {
	panic("unimplemented")
}

// OffchainConfig i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) OffchainConfig(ctx context.Context) (cciptypes.ExecOffchainConfig, error) {
	panic("unimplemented")
}

// OnchainConfig i[github.com/smartcontractkit/chainlink-common/pkg/types/ccip.OffRampReader]
func (o *OffRampReaderClient) OnchainConfig(ctx context.Context) (cciptypes.ExecOnchainConfig, error) {
	panic("unimplemented")
}

func NewOffRampReaderClient(grpc ccippb.OffRampReaderClient) *OffRampReaderClient {
	return &OffRampReaderClient{grpc: grpc}
}

func tokenBucketRateLimit(pb *ccippb.TokenPoolRateLimit) cciptypes.TokenBucketRateLimit {
	return cciptypes.TokenBucketRateLimit{
		Tokens:      pb.Tokens.Int(),
		LastUpdated: pb.LastUpdated,
		IsEnabled:   pb.IsEnabled,
		Capacity:    pb.Capacity.Int(),
		Rate:        pb.Rate.Int(),
	}
}

func execReport(pb *ccippb.ExecutionReport) (cciptypes.ExecReport, error) {
	proofs, err := byte32Slice(pb.Proofs)
	if err != nil {
		return cciptypes.ExecReport{}, fmt.Errorf("execReport: invalid proofs: %w", err)
	}
	msgs, err := evm2EVMMessageSlice(pb.EvmToEvmMessages)
	if err != nil {
		return cciptypes.ExecReport{}, fmt.Errorf("execReport: invalid messages: %w", err)
	}

	return cciptypes.ExecReport{
		Messages:          msgs,
		OffchainTokenData: offchainTokenData(pb.OffchainTokenData),
		Proofs:            proofs,
		ProofFlagBits:     pb.ProofFlagBits.Int(),
	}, nil
}

func evm2EVMMessageSlice(in []*ccippb.EVM2EVMMessage) ([]cciptypes.EVM2EVMMessage, error) {
	out := make([]cciptypes.EVM2EVMMessage, len(in))
	for i, m := range in {
		decodedMsg, err := evm2EVMMessage(m)
		if err != nil {
			return nil, err
		}
		out[i] = decodedMsg
	}
	return out, nil
}

func offchainTokenData(in []*ccippb.TokenData) [][][]byte {
	out := make([][][]byte, len(in))
	for i, b := range in {
		out[i] = b.Data
	}
	return out
}

func byte32Slice(in [][]byte) ([][32]byte, error) {
	out := make([][32]byte, len(in))
	for i, b := range in {
		if len(b) != 32 {
			return nil, fmt.Errorf("byte32Slice: invalid length %d", len(b))
		}
		copy(out[i][:], b)
	}
	return out, nil
}

func executionReportPB(report cciptypes.ExecReport) *ccippb.ExecutionReport {
	return &ccippb.ExecutionReport{
		EvmToEvmMessages:  evm2EVMMessageSliceToPB(report.Messages),
		OffchainTokenData: offchainTokenDataToPB(report.OffchainTokenData),
		Proofs:            byte32SliceToPB(report.Proofs),
		ProofFlagBits:     pb.NewBigIntFromInt(report.ProofFlagBits),
	}
}

func evm2EVMMessageSliceToPB(in []cciptypes.EVM2EVMMessage) []*ccippb.EVM2EVMMessage {
	out := make([]*ccippb.EVM2EVMMessage, len(in))
	for i, m := range in {
		out[i] = evm2EVMMessageToPB(m)
	}
	return out
}

func offchainTokenDataToPB(in [][][]byte) []*ccippb.TokenData {
	out := make([]*ccippb.TokenData, len(in))
	for i, b := range in {
		out[i] = &ccippb.TokenData{Data: b}
	}
	return out
}

func byte32SliceToPB(in [][32]byte) [][]byte {
	out := make([][]byte, len(in))
	for i, b := range in {
		out[i] = b[:]
	}
	return out
}

func evm2EVMMessageToPB(m cciptypes.EVM2EVMMessage) *ccippb.EVM2EVMMessage {
	return &ccippb.EVM2EVMMessage{
		SequenceNumber:      m.SequenceNumber,
		GasLimit:            pb.NewBigIntFromInt(m.GasLimit),
		Nonce:               m.Nonce,
		MessageId:           m.MessageID[:],
		SourceChainSelector: m.SourceChainSelector,
		Sender:              string(m.Sender),
		Receiver:            string(m.Receiver),
		Strict:              m.Strict,
		FeeToken:            string(m.FeeToken),
		FeeTokenAmount:      pb.NewBigIntFromInt(m.FeeTokenAmount),
		Data:                m.Data,
		TokenAmounts:        tokenAmountSliceToPB(m.TokenAmounts),
		SourceTokenData:     m.SourceTokenData,
	}
}

func tokenAmountSliceToPB(tokenAmounts []cciptypes.TokenAmount) []*ccippb.TokenAmount {
	res := make([]*ccippb.TokenAmount, len(tokenAmounts))
	for i, t := range tokenAmounts {
		res[i] = &ccippb.TokenAmount{
			Token:  string(t.Token),
			Amount: pb.NewBigIntFromInt(t.Amount),
		}
	}
	return res
}

func executionStateChangedWithTxMetaSlice(in []*ccippb.ExecutionStateChangeWithTxMeta) []cciptypes.ExecutionStateChangedWithTxMeta {
	out := make([]cciptypes.ExecutionStateChangedWithTxMeta, len(in))
	for i, m := range in {
		out[i] = executionStateChangedWithTxMeta(m)
	}
	return out
}

func executionStateChangedWithTxMeta(in *ccippb.ExecutionStateChangeWithTxMeta) cciptypes.ExecutionStateChangedWithTxMeta {
	return cciptypes.ExecutionStateChangedWithTxMeta{
		TxMeta: txMeta(in.TxMeta),
		ExecutionStateChanged: cciptypes.ExecutionStateChanged{
			SequenceNumber: in.ExecutionStateChange.SeqNum,
			Finalized:      in.ExecutionStateChange.Finalized,
		},
	}
}
