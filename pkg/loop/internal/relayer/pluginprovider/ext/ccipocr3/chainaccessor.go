package ccipocr3

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var _ ccipocr3.ChainAccessor = (*chainAccessorClient)(nil)

type chainAccessorClient struct {
	*net.BrokerExt
	grpc ccipocr3pb.ChainAccessorClient

	// Local persistence for refresh functionality
	mu              sync.RWMutex
	syncedContracts map[string]ccipocr3.UnknownAddress // contractName -> contractAddress
}

func NewChainAccessorClient(broker *net.BrokerExt, cc grpc.ClientConnInterface) ccipocr3.ChainAccessor {
	return &chainAccessorClient{
		BrokerExt:       broker,
		grpc:            ccipocr3pb.NewChainAccessorClient(cc),
		syncedContracts: make(map[string]ccipocr3.UnknownAddress),
	}
}

// AllAccessors methods
func (c *chainAccessorClient) GetContractAddress(contractName string) ([]byte, error) {
	resp, err := c.grpc.GetContractAddress(context.Background(), &ccipocr3pb.GetContractAddressRequest{
		ContractName: contractName,
	})
	if err != nil {
		return nil, err
	}
	return resp.Address, nil
}

func (c *chainAccessorClient) GetAllConfigsLegacy(
	ctx context.Context,
	destChainSelector ccipocr3.ChainSelector,
	sourceChainSelectors []ccipocr3.ChainSelector,
) (ccipocr3.ChainConfigSnapshot, map[ccipocr3.ChainSelector]ccipocr3.SourceChainConfig, error) {
	var sourceSels []uint64
	for _, sel := range sourceChainSelectors {
		sourceSels = append(sourceSels, uint64(sel))
	}

	resp, err := c.grpc.GetAllConfigsLegacy(ctx, &ccipocr3pb.GetAllConfigsLegacyRequest{
		DestChainSelector:    uint64(destChainSelector),
		SourceChainSelectors: sourceSels,
	})
	if err != nil {
		return ccipocr3.ChainConfigSnapshot{}, nil, err
	}

	// Convert response source chain configs
	sourceConfigs := make(map[ccipocr3.ChainSelector]ccipocr3.SourceChainConfig)
	for chainSel, pbConfig := range resp.SourceChainConfigs {
		sourceConfigs[ccipocr3.ChainSelector(chainSel)] = pbToSourceChainConfig(pbConfig)
	}

	return pbToChainConfigSnapshotDetailed(resp.Snapshot), sourceConfigs, nil
}

func (c *chainAccessorClient) GetChainFeeComponents(ctx context.Context) (ccipocr3.ChainFeeComponents, error) {
	resp, err := c.grpc.GetChainFeeComponents(ctx, &emptypb.Empty{})
	if err != nil {
		return ccipocr3.ChainFeeComponents{}, err
	}
	return ccipocr3.ChainFeeComponents{
		ExecutionFee:        pbBigIntToInt(resp.FeeComponents.ExecutionFee),
		DataAvailabilityFee: pbBigIntToInt(resp.FeeComponents.DataAvailabilityFee),
	}, nil
}

func (c *chainAccessorClient) Sync(ctx context.Context, contractName string, contractAddress ccipocr3.UnknownAddress) error {
	_, err := c.grpc.Sync(ctx, &ccipocr3pb.SyncRequest{
		ContractName:    contractName,
		ContractAddress: contractAddress,
	})

	if err == nil {
		// Persist the synced contract locally for client refresh
		c.mu.Lock()
		c.syncedContracts[contractName] = contractAddress
		c.mu.Unlock()
		c.Logger.Debugw("Persisted synced contract", "contractName", contractName, "contractAddress", contractAddress)
	}

	return err
}

// Refresh re-syncs all previously synced contracts after a connection refresh.
func (c *chainAccessorClient) Refresh(ctx context.Context) error {
	c.mu.RLock()
	contracts := make(map[string]ccipocr3.UnknownAddress)
	for name, addr := range c.syncedContracts {
		contracts[name] = addr
	}
	c.mu.RUnlock()

	if len(contracts) == 0 {
		c.Logger.Debug("No previously synced contracts to refresh")
		return nil
	}

	c.Logger.Infow("Refreshing synced contracts", "count", len(contracts))

	for contractName, contractAddress := range contracts {
		if err := c.Sync(ctx, contractName, contractAddress); err != nil {
			// If sync fails, log, but continue with the rest of the syncing
			c.Logger.Errorw("Failed to refresh contract sync",
				"contractName", contractName,
				"contractAddress", contractAddress,
				"err", err)
		} else {
			c.Logger.Debugw("Successfully refreshed contract sync",
				"contractName", contractName,
				"contractAddress", contractAddress)
		}
	}

	return nil
}

// DestinationAccessor methods
func (c *chainAccessorClient) CommitReportsGTETimestamp(
	ctx context.Context,
	ts time.Time,
	confidence primitives.ConfidenceLevel,
	limit int,
) ([]ccipocr3.CommitPluginReportWithMeta, error) {
	resp, err := c.grpc.CommitReportsGTETimestamp(ctx, &ccipocr3pb.CommitReportsGTETimestampRequest{
		Timestamp:       timestamppb.New(ts),
		ConfidenceLevel: confidenceLevelToPb(confidence),
		Limit:           int32(limit),
	})
	if err != nil {
		return nil, err
	}

	var reports []ccipocr3.CommitPluginReportWithMeta
	for _, r := range resp.Reports {
		reports = append(reports, ccipocr3.CommitPluginReportWithMeta{
			Report:    pbToCommitPluginReportDetailed(r.Report),
			Timestamp: r.Timestamp.AsTime(),
			BlockNum:  r.BlockNum,
		})
	}
	return reports, nil
}

func (c *chainAccessorClient) ExecutedMessages(
	ctx context.Context,
	ranges map[ccipocr3.ChainSelector][]ccipocr3.SeqNumRange,
	confidence primitives.ConfidenceLevel,
) (map[ccipocr3.ChainSelector][]ccipocr3.SeqNum, error) {
	req := &ccipocr3pb.ExecutedMessagesRequest{
		Ranges:          make(map[uint64]*ccipocr3pb.SequenceNumberRangeList),
		ConfidenceLevel: confidenceLevelToPb(confidence),
	}

	for chainSel, rangeList := range ranges {
		pbRanges := &ccipocr3pb.SequenceNumberRangeList{}
		for _, r := range rangeList {
			pbRanges.Ranges = append(pbRanges.Ranges, &ccipocr3pb.SeqNumRange{
				Start: uint64(r.Start()),
				End:   uint64(r.End()),
			})
		}
		req.Ranges[uint64(chainSel)] = pbRanges
	}

	resp, err := c.grpc.ExecutedMessages(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make(map[ccipocr3.ChainSelector][]ccipocr3.SeqNum)
	for chainSel, seqNums := range resp.ExecutedMessages {
		var seqs []ccipocr3.SeqNum
		for _, seqNum := range seqNums.SeqNums {
			seqs = append(seqs, ccipocr3.SeqNum(seqNum))
		}
		result[ccipocr3.ChainSelector(chainSel)] = seqs
	}
	return result, nil
}

func (c *chainAccessorClient) NextSeqNum(ctx context.Context, sources []ccipocr3.ChainSelector) (map[ccipocr3.ChainSelector]ccipocr3.SeqNum, error) {
	var chainSelectors []uint64
	for _, source := range sources {
		chainSelectors = append(chainSelectors, uint64(source))
	}

	resp, err := c.grpc.NextSeqNum(ctx, &ccipocr3pb.NextSeqNumRequest{
		SourceChainSelectors: chainSelectors,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[ccipocr3.ChainSelector]ccipocr3.SeqNum)
	for chainSel, seqNum := range resp.NextSeqNums {
		result[ccipocr3.ChainSelector(chainSel)] = ccipocr3.SeqNum(seqNum)
	}
	return result, nil
}

func (c *chainAccessorClient) Nonces(ctx context.Context, addresses map[ccipocr3.ChainSelector][]ccipocr3.UnknownEncodedAddress) (map[ccipocr3.ChainSelector]map[string]uint64, error) {
	req := &ccipocr3pb.NoncesRequest{
		Addresses: make(map[uint64]*ccipocr3pb.UnknownEncodedAddressList),
	}

	for chainSel, addrs := range addresses {
		addrList := &ccipocr3pb.UnknownEncodedAddressList{}
		for _, addr := range addrs {
			addrList.Addresses = append(addrList.Addresses, string(addr))
		}
		req.Addresses[uint64(chainSel)] = addrList
	}

	resp, err := c.grpc.Nonces(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make(map[ccipocr3.ChainSelector]map[string]uint64)
	for chainSel, nonceMap := range resp.Nonces {
		result[ccipocr3.ChainSelector(chainSel)] = nonceMap.Nonces
	}
	return result, nil
}

func (c *chainAccessorClient) GetChainFeePriceUpdate(ctx context.Context, selectors []ccipocr3.ChainSelector) (map[ccipocr3.ChainSelector]ccipocr3.TimestampedUnixBig, error) {
	var chainSelectors []uint64
	for _, sel := range selectors {
		chainSelectors = append(chainSelectors, uint64(sel))
	}

	resp, err := c.grpc.GetChainFeePriceUpdate(ctx, &ccipocr3pb.GetChainFeePriceUpdateRequest{
		ChainSelectors: chainSelectors,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[ccipocr3.ChainSelector]ccipocr3.TimestampedUnixBig)
	for chainSel, timestampedUnixBig := range resp.FeePriceUpdates {
		result[ccipocr3.ChainSelector(chainSel)] = ccipocr3.TimestampedUnixBig{
			Value:     pbToBigInt(timestampedUnixBig.Value).Int,
			Timestamp: timestampedUnixBig.Timestamp,
		}
	}
	return result, nil
}

func (c *chainAccessorClient) GetLatestPriceSeqNr(ctx context.Context) (ccipocr3.SeqNum, error) {
	resp, err := c.grpc.GetLatestPriceSeqNr(ctx, &emptypb.Empty{})
	if err != nil {
		return 0, err
	}
	return ccipocr3.SeqNum(resp.SeqNr), nil
}

// SourceAccessor methods
func (c *chainAccessorClient) MsgsBetweenSeqNums(ctx context.Context, dest ccipocr3.ChainSelector, seqNumRange ccipocr3.SeqNumRange) ([]ccipocr3.Message, error) {
	resp, err := c.grpc.MsgsBetweenSeqNums(ctx, &ccipocr3pb.MsgsBetweenSeqNumsRequest{
		DestChainSelector: uint64(dest),
		SeqNumRange: &ccipocr3pb.SeqNumRange{
			Start: uint64(seqNumRange.Start()),
			End:   uint64(seqNumRange.End()),
		},
	})
	if err != nil {
		return nil, err
	}

	var messages []ccipocr3.Message
	for _, msg := range resp.Messages {
		messages = append(messages, pbToMessage(msg))
	}
	return messages, nil
}

func (c *chainAccessorClient) LatestMessageTo(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.SeqNum, error) {
	resp, err := c.grpc.LatestMessageTo(ctx, &ccipocr3pb.LatestMessageToRequest{
		DestChainSelector: uint64(dest),
	})
	if err != nil {
		return 0, err
	}
	return ccipocr3.SeqNum(resp.SeqNum), nil
}

func (c *chainAccessorClient) GetExpectedNextSequenceNumber(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.SeqNum, error) {
	resp, err := c.grpc.GetExpectedNextSequenceNumber(ctx, &ccipocr3pb.GetExpectedNextSequenceNumberRequest{
		DestChainSelector: uint64(dest),
	})
	if err != nil {
		return 0, err
	}
	return ccipocr3.SeqNum(resp.SeqNum), nil
}

func (c *chainAccessorClient) GetTokenPriceUSD(ctx context.Context, address ccipocr3.UnknownAddress) (ccipocr3.TimestampedUnixBig, error) {
	resp, err := c.grpc.GetTokenPriceUSD(ctx, &ccipocr3pb.GetTokenPriceUSDRequest{
		Address: address,
	})
	if err != nil {
		return ccipocr3.TimestampedUnixBig{}, err
	}
	return ccipocr3.TimestampedUnixBig{
		Value:     pbBigIntToInt(resp.Price.Value),
		Timestamp: resp.Price.Timestamp,
	}, nil
}

func (c *chainAccessorClient) GetFeeQuoterDestChainConfig(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.FeeQuoterDestChainConfig, error) {
	resp, err := c.grpc.GetFeeQuoterDestChainConfig(ctx, &ccipocr3pb.GetFeeQuoterDestChainConfigRequest{
		DestChainSelector: uint64(dest),
	})
	if err != nil {
		return ccipocr3.FeeQuoterDestChainConfig{}, err
	}
	return pbToFeeQuoterDestChainConfigDetailed(resp.Config), nil
}

// USDCMessageReader methods
func (c *chainAccessorClient) MessagesByTokenID(ctx context.Context, source, dest ccipocr3.ChainSelector, tokens map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount) (map[ccipocr3.MessageTokenID]ccipocr3.Bytes, error) {
	resp, err := c.grpc.MessagesByTokenID(ctx, &ccipocr3pb.MessagesByTokenIDRequest{
		SourceChainSelector: uint64(source),
		DestChainSelector:   uint64(dest),
		Tokens:              messageTokenIDMapToPb(tokens),
	})
	if err != nil {
		return nil, err
	}
	return pbToMessagesByTokenID(resp.Messages)
}

// PriceReader methods
func (c *chainAccessorClient) GetFeedPricesUSD(ctx context.Context, tokens []ccipocr3.UnknownEncodedAddress, tokenInfo map[ccipocr3.UnknownEncodedAddress]ccipocr3.TokenInfo) (ccipocr3.TokenPriceMap, error) {
	var tokenStrs []string
	for _, token := range tokens {
		tokenStrs = append(tokenStrs, string(token))
	}

	resp, err := c.grpc.GetFeedPricesUSD(ctx, &ccipocr3pb.GetFeedPricesUSDRequest{
		Tokens:    tokenStrs,
		TokenInfo: tokenInfoMapToPb(tokenInfo),
	})
	if err != nil {
		return nil, err
	}
	return pbToTokenPriceMap(resp.Prices), nil
}

func (c *chainAccessorClient) GetFeeQuoterTokenUpdates(ctx context.Context, tokens []ccipocr3.UnknownEncodedAddress, chain ccipocr3.ChainSelector) (map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig, error) {
	var tokenStrs []string
	for _, token := range tokens {
		tokenStrs = append(tokenStrs, string(token))
	}

	resp, err := c.grpc.GetFeeQuoterTokenUpdates(ctx, &ccipocr3pb.GetFeeQuoterTokenUpdatesRequest{
		Tokens:        tokenStrs,
		ChainSelector: uint64(chain),
	})
	if err != nil {
		return nil, err
	}
	return pbToTokenUpdatesUnix(resp.TokenUpdates), nil
}

// Server implementation
var _ ccipocr3pb.ChainAccessorServer = (*chainAccessorServer)(nil)

type chainAccessorServer struct {
	ccipocr3pb.UnimplementedChainAccessorServer
	impl ccipocr3.ChainAccessor
}

func NewChainAccessorServer(impl ccipocr3.ChainAccessor) *chainAccessorServer {
	return &chainAccessorServer{impl: impl}
}

// AllAccessors methods
func (s *chainAccessorServer) GetContractAddress(ctx context.Context, req *ccipocr3pb.GetContractAddressRequest) (*ccipocr3pb.GetContractAddressResponse, error) {
	addr, err := s.impl.GetContractAddress(req.ContractName)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetContractAddressResponse{Address: addr}, nil
}

func (s *chainAccessorServer) GetAllConfigsLegacy(ctx context.Context, req *ccipocr3pb.GetAllConfigsLegacyRequest) (*ccipocr3pb.GetAllConfigsLegacyResponse, error) {
	// Convert request parameters
	var sourceChainSelectors []ccipocr3.ChainSelector
	for _, sel := range req.SourceChainSelectors {
		sourceChainSelectors = append(sourceChainSelectors, ccipocr3.ChainSelector(sel))
	}

	snapshot, sourceConfigs, err := s.impl.GetAllConfigsLegacy(
		ctx,
		ccipocr3.ChainSelector(req.DestChainSelector),
		sourceChainSelectors,
	)
	if err != nil {
		return nil, err
	}

	// Convert response source chain configs
	pbSourceConfigs := make(map[uint64]*ccipocr3pb.SourceChainConfig)
	for chainSel, config := range sourceConfigs {
		pbSourceConfigs[uint64(chainSel)] = sourceChainConfigToPb(config)
	}

	return &ccipocr3pb.GetAllConfigsLegacyResponse{
		Snapshot:           chainConfigSnapshotToPbDetailed(snapshot),
		SourceChainConfigs: pbSourceConfigs,
	}, nil
}

func (s *chainAccessorServer) GetChainFeeComponents(ctx context.Context, req *emptypb.Empty) (*ccipocr3pb.GetChainFeeComponentsResponse, error) {
	feeComponents, err := s.impl.GetChainFeeComponents(ctx)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetChainFeeComponentsResponse{
		FeeComponents: &ccipocr3pb.ChainFeeComponents{
			ExecutionFee:        intToPbBigInt(feeComponents.ExecutionFee),
			DataAvailabilityFee: intToPbBigInt(feeComponents.DataAvailabilityFee),
		},
	}, nil
}

func (s *chainAccessorServer) Sync(ctx context.Context, req *ccipocr3pb.SyncRequest) (*emptypb.Empty, error) {
	err := s.impl.Sync(ctx, req.ContractName, ccipocr3.UnknownAddress(req.ContractAddress))
	return &emptypb.Empty{}, err
}

// DestinationAccessor server methods
func (s *chainAccessorServer) CommitReportsGTETimestamp(ctx context.Context, req *ccipocr3pb.CommitReportsGTETimestampRequest) (*ccipocr3pb.CommitReportsGTETimestampResponse, error) {
	reports, err := s.impl.CommitReportsGTETimestamp(
		ctx,
		req.Timestamp.AsTime(),
		pbToConfidenceLevel(req.ConfidenceLevel),
		int(req.Limit),
	)
	if err != nil {
		return nil, err
	}

	var pbReports []*ccipocr3pb.CommitPluginReportWithMeta
	for _, report := range reports {
		pbReports = append(pbReports, &ccipocr3pb.CommitPluginReportWithMeta{
			Report:    commitPluginReportToPb(report.Report),
			Timestamp: timestamppb.New(report.Timestamp),
			BlockNum:  report.BlockNum,
		})
	}

	return &ccipocr3pb.CommitReportsGTETimestampResponse{Reports: pbReports}, nil
}

func (s *chainAccessorServer) ExecutedMessages(ctx context.Context, req *ccipocr3pb.ExecutedMessagesRequest) (*ccipocr3pb.ExecutedMessagesResponse, error) {
	ranges := make(map[ccipocr3.ChainSelector][]ccipocr3.SeqNumRange)
	for chainSel, rangeList := range req.Ranges {
		var seqRanges []ccipocr3.SeqNumRange
		for _, pbRange := range rangeList.Ranges {
			seqRanges = append(seqRanges, ccipocr3.NewSeqNumRange(
				ccipocr3.SeqNum(pbRange.Start),
				ccipocr3.SeqNum(pbRange.End),
			))
		}
		ranges[ccipocr3.ChainSelector(chainSel)] = seqRanges
	}

	executedMessages, err := s.impl.ExecutedMessages(
		ctx,
		ranges,
		pbToConfidenceLevel(req.ConfidenceLevel),
	)
	if err != nil {
		return nil, err
	}

	pbExecutedMessages := make(map[uint64]*ccipocr3pb.SequenceNumberList)
	for chainSel, seqNums := range executedMessages {
		seqNumList := &ccipocr3pb.SequenceNumberList{}
		for _, seqNum := range seqNums {
			seqNumList.SeqNums = append(seqNumList.SeqNums, uint64(seqNum))
		}
		pbExecutedMessages[uint64(chainSel)] = seqNumList
	}

	return &ccipocr3pb.ExecutedMessagesResponse{ExecutedMessages: pbExecutedMessages}, nil
}

func (s *chainAccessorServer) NextSeqNum(ctx context.Context, req *ccipocr3pb.NextSeqNumRequest) (*ccipocr3pb.NextSeqNumResponse, error) {
	// Convert request: []uint64 -> []ChainSelector
	var sources []ccipocr3.ChainSelector
	for _, sel := range req.SourceChainSelectors {
		sources = append(sources, ccipocr3.ChainSelector(sel))
	}

	seqNumMap, err := s.impl.NextSeqNum(ctx, sources)
	if err != nil {
		return nil, err
	}

	// Convert response: map[ChainSelector]SeqNum -> map[uint64]uint64
	nextSeqNums := make(map[uint64]uint64)
	for chainSel, seqNum := range seqNumMap {
		nextSeqNums[uint64(chainSel)] = uint64(seqNum)
	}

	return &ccipocr3pb.NextSeqNumResponse{NextSeqNums: nextSeqNums}, nil
}

func (s *chainAccessorServer) Nonces(ctx context.Context, req *ccipocr3pb.NoncesRequest) (*ccipocr3pb.NoncesResponse, error) {
	// Convert request: map[uint64]UnknownEncodedAddressList -> map[ChainSelector][]UnknownEncodedAddress
	addresses := make(map[ccipocr3.ChainSelector][]ccipocr3.UnknownEncodedAddress)
	for chainSel, addrList := range req.Addresses {
		var addrs []ccipocr3.UnknownEncodedAddress
		for _, addr := range addrList.Addresses {
			addrs = append(addrs, ccipocr3.UnknownEncodedAddress(addr))
		}
		addresses[ccipocr3.ChainSelector(chainSel)] = addrs
	}

	nonces, err := s.impl.Nonces(ctx, addresses)
	if err != nil {
		return nil, err
	}

	// Convert response: map[ChainSelector]map[string]uint64 -> map[uint64]NonceMap
	pbNonces := make(map[uint64]*ccipocr3pb.NonceMap)
	for chainSel, nonceMap := range nonces {
		pbNonces[uint64(chainSel)] = &ccipocr3pb.NonceMap{
			Nonces: nonceMap,
		}
	}

	return &ccipocr3pb.NoncesResponse{Nonces: pbNonces}, nil
}

func (s *chainAccessorServer) GetChainFeePriceUpdate(ctx context.Context, req *ccipocr3pb.GetChainFeePriceUpdateRequest) (*ccipocr3pb.GetChainFeePriceUpdateResponse, error) {
	// Convert request chain selectors
	var chainSelectors []ccipocr3.ChainSelector
	for _, sel := range req.ChainSelectors {
		chainSelectors = append(chainSelectors, ccipocr3.ChainSelector(sel))
	}

	priceUpdates, err := s.impl.GetChainFeePriceUpdate(ctx, chainSelectors)
	if err != nil {
		return nil, err
	}

	pbUpdates := make(map[uint64]*ccipocr3pb.TimestampedUnixBig)
	for chainSel, update := range priceUpdates {
		pbUpdates[uint64(chainSel)] = &ccipocr3pb.TimestampedUnixBig{
			Value:     intToPbBigInt(update.Value),
			Timestamp: update.Timestamp,
		}
	}

	return &ccipocr3pb.GetChainFeePriceUpdateResponse{
		FeePriceUpdates: pbUpdates,
	}, nil
}

func (s *chainAccessorServer) GetLatestPriceSeqNr(ctx context.Context, req *emptypb.Empty) (*ccipocr3pb.GetLatestPriceSeqNrResponse, error) {
	seqNr, err := s.impl.GetLatestPriceSeqNr(ctx)
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetLatestPriceSeqNrResponse{SeqNr: uint64(seqNr)}, nil
}

// SourceAccessor server methods
func (s *chainAccessorServer) MsgsBetweenSeqNums(ctx context.Context, req *ccipocr3pb.MsgsBetweenSeqNumsRequest) (*ccipocr3pb.MsgsBetweenSeqNumsResponse, error) {
	seqNumRange := ccipocr3.NewSeqNumRange(
		ccipocr3.SeqNum(req.SeqNumRange.Start),
		ccipocr3.SeqNum(req.SeqNumRange.End),
	)

	messages, err := s.impl.MsgsBetweenSeqNums(
		ctx,
		ccipocr3.ChainSelector(req.DestChainSelector),
		seqNumRange,
	)
	if err != nil {
		return nil, err
	}

	var pbMessages []*ccipocr3pb.Message
	for _, msg := range messages {
		pbMessages = append(pbMessages, messageToPb(msg))
	}

	return &ccipocr3pb.MsgsBetweenSeqNumsResponse{Messages: pbMessages}, nil
}

func (s *chainAccessorServer) LatestMessageTo(ctx context.Context, req *ccipocr3pb.LatestMessageToRequest) (*ccipocr3pb.LatestMessageToResponse, error) {
	seqNum, err := s.impl.LatestMessageTo(ctx, ccipocr3.ChainSelector(req.DestChainSelector))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.LatestMessageToResponse{SeqNum: uint64(seqNum)}, nil
}

func (s *chainAccessorServer) GetExpectedNextSequenceNumber(ctx context.Context, req *ccipocr3pb.GetExpectedNextSequenceNumberRequest) (*ccipocr3pb.GetExpectedNextSequenceNumberResponse, error) {
	seqNum, err := s.impl.GetExpectedNextSequenceNumber(ctx, ccipocr3.ChainSelector(req.DestChainSelector))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetExpectedNextSequenceNumberResponse{SeqNum: uint64(seqNum)}, nil
}

func (s *chainAccessorServer) GetTokenPriceUSD(ctx context.Context, req *ccipocr3pb.GetTokenPriceUSDRequest) (*ccipocr3pb.GetTokenPriceUSDResponse, error) {
	price, err := s.impl.GetTokenPriceUSD(ctx, ccipocr3.UnknownAddress(req.Address))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetTokenPriceUSDResponse{
		Price: &ccipocr3pb.TimestampedUnixBig{
			Value:     intToPbBigInt(price.Value),
			Timestamp: price.Timestamp,
		},
	}, nil
}

func (s *chainAccessorServer) GetFeeQuoterDestChainConfig(ctx context.Context, req *ccipocr3pb.GetFeeQuoterDestChainConfigRequest) (*ccipocr3pb.GetFeeQuoterDestChainConfigResponse, error) {
	config, err := s.impl.GetFeeQuoterDestChainConfig(ctx, ccipocr3.ChainSelector(req.DestChainSelector))
	if err != nil {
		return nil, err
	}
	return &ccipocr3pb.GetFeeQuoterDestChainConfigResponse{
		Config: feeQuoterDestChainConfigToPb(config),
	}, nil
}

// USDCMessageReader server methods
func (s *chainAccessorServer) MessagesByTokenID(ctx context.Context, req *ccipocr3pb.MessagesByTokenIDRequest) (*ccipocr3pb.MessagesByTokenIDResponse, error) {
	tokens, err := pbToMessageTokenIDMap(req.Tokens)
	if err != nil {
		return nil, err
	}

	messages, err := s.impl.MessagesByTokenID(
		ctx,
		ccipocr3.ChainSelector(req.SourceChainSelector),
		ccipocr3.ChainSelector(req.DestChainSelector),
		tokens,
	)
	if err != nil {
		return nil, err
	}

	return &ccipocr3pb.MessagesByTokenIDResponse{
		Messages: messagesByTokenIDToPb(messages),
	}, nil
}

// PriceReader server methods
func (s *chainAccessorServer) GetFeedPricesUSD(ctx context.Context, req *ccipocr3pb.GetFeedPricesUSDRequest) (*ccipocr3pb.GetFeedPricesUSDResponse, error) {
	var tokens []ccipocr3.UnknownEncodedAddress
	for _, tokenStr := range req.Tokens {
		tokens = append(tokens, ccipocr3.UnknownEncodedAddress(tokenStr))
	}
	tokenInfo := pbToTokenInfoMap(req.TokenInfo)

	prices, err := s.impl.GetFeedPricesUSD(ctx, tokens, tokenInfo)
	if err != nil {
		return nil, err
	}

	return &ccipocr3pb.GetFeedPricesUSDResponse{
		Prices: tokenPriceMapToPb(prices),
	}, nil
}

func (s *chainAccessorServer) GetFeeQuoterTokenUpdates(ctx context.Context, req *ccipocr3pb.GetFeeQuoterTokenUpdatesRequest) (*ccipocr3pb.GetFeeQuoterTokenUpdatesResponse, error) {
	var tokens []ccipocr3.UnknownEncodedAddress
	for _, tokenStr := range req.Tokens {
		tokens = append(tokens, ccipocr3.UnknownEncodedAddress(tokenStr))
	}

	updates, err := s.impl.GetFeeQuoterTokenUpdates(
		ctx,
		tokens,
		ccipocr3.ChainSelector(req.ChainSelector),
	)
	if err != nil {
		return nil, err
	}

	return &ccipocr3pb.GetFeeQuoterTokenUpdatesResponse{
		TokenUpdates: tokenUpdatesUnixToPb(updates),
	}, nil
}
