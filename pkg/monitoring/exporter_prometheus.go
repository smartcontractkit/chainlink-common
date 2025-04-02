package monitoring

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func NewPrometheusExporterFactory(
	log Logger,
	metrics Metrics,
) ExporterFactory {
	return &prometheusExporterFactory{
		log,
		metrics,
	}
}

type prometheusExporterFactory struct {
	log     Logger
	metrics Metrics
}

func (p *prometheusExporterFactory) NewExporter(
	params ExporterParams,
) (Exporter, error) {
	chainConfig, feedConfig, nodes := params.ChainConfig, params.FeedConfig, params.Nodes
	p.metrics.SetFeedContractMetadata(
		chainConfig.GetChainID(),
		feedConfig.GetID(),
		feedConfig.GetID(),
		feedConfig.GetContractStatus(),
		feedConfig.GetContractType(),
		feedConfig.GetName(),
		feedConfig.GetPath(),
		chainConfig.GetNetworkID(),
		chainConfig.GetNetworkName(),
		feedConfig.GetSymbol(),
	)
	exporter := &prometheusExporter{
		chainConfig,
		feedConfig,
		nodes,
		p.log,
		p.metrics,
		prometheusLabels{},
		sync.Mutex{},
		new(big.Int),
		time.Time{},
		sync.Mutex{},
	}
	exporter.updateLabels(prometheusLabels{
		networkName:     chainConfig.GetNetworkName(),
		networkID:       chainConfig.GetNetworkID(),
		chainID:         chainConfig.GetChainID(),
		feedName:        feedConfig.GetName(),
		feedPath:        feedConfig.GetPath(),
		symbol:          feedConfig.GetSymbol(),
		contractType:    feedConfig.GetContractType(),
		contractStatus:  feedConfig.GetContractStatus(),
		contractAddress: feedConfig.GetID(),
		feedID:          feedConfig.GetID(),
	})
	return exporter, nil
}

type prometheusExporter struct {
	chainConfig ChainConfig
	feedConfig  FeedConfig
	nodes       []NodeConfig

	log     Logger
	metrics Metrics

	labels   prometheusLabels
	labelsMu sync.Mutex

	prevValue     *big.Int
	prevTimestamp time.Time
	prevMu        sync.Mutex
}

func (p *prometheusExporter) Export(_ context.Context, data interface{}) {
	switch typed := data.(type) {
	case Envelope:
		p.exportEnvelope(typed)
	case TxResults:
		p.exportTxResults(typed)
	}
}

func (p *prometheusExporter) exportEnvelope(envelope Envelope) {
	p.updateLabels(prometheusLabels{
		sender: string(envelope.Transmitter),
	})
	multiply := toFloat64(p.feedConfig.GetMultiply())
	if multiply == 0.0 {
		multiply = 1.0
	}
	linkBalance := toFloat64(envelope.LinkBalance)
	p.metrics.SetFeedContractLinkBalance(
		linkBalance,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	linkAvailableForPayment := toFloat64(envelope.LinkAvailableForPayment)
	p.metrics.SetLinkAvailableForPayment(
		linkAvailableForPayment,
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	oracleName, found := getOracleName(envelope.Transmitter, p.nodes)
	if !found {
		oracleName = string(envelope.Transmitter)
	}
	p.metrics.SetNodeMetadata(
		p.chainConfig.GetChainID(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
		oracleName,                   // oracleName
		string(envelope.Transmitter), // sender
	)
	p.metrics.SetHeadTrackerCurrentHead(
		float64(envelope.BlockNumber),
		p.chainConfig.GetNetworkName(),
		p.chainConfig.GetChainID(),
		p.chainConfig.GetNetworkID(),
	)
	if p.feedConfig.GetHeartbeatSec() != 0 {
		isLateAnswer := time.Since(envelope.LatestTimestamp).Seconds() > float64(p.feedConfig.GetHeartbeatSec())
		p.metrics.SetOffchainAggregatorAnswerStalled(
			isLateAnswer,
			p.feedConfig.GetID(),
			p.feedConfig.GetID(),
			p.chainConfig.GetChainID(),
			p.feedConfig.GetContractStatus(),
			p.feedConfig.GetContractType(),
			p.feedConfig.GetName(),
			p.feedConfig.GetPath(),
			p.chainConfig.GetNetworkID(),
			p.chainConfig.GetNetworkName(),
		)
	}
	if !p.isNewTransmission(envelope.LatestAnswer, envelope.LatestTimestamp) {
		return
	}
	// All the metrics below are only updated if there was a fresh
	// transmission since the last chain read.
	latestAnswer := toFloat64(envelope.LatestAnswer)
	p.metrics.SetOffchainAggregatorAnswers(
		latestAnswer/multiply,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorAnswersRaw(
		latestAnswer,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.IncOffchainAggregatorAnswersTotal(
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorAnswersLatestTimestamp(
		float64(envelope.LatestTimestamp.Unix()),
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	juelsPerFeeCoin := toFloat64(envelope.JuelsPerFeeCoin)
	p.metrics.SetOffchainAggregatorJuelsPerFeeCoinRaw(
		juelsPerFeeCoin,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorJuelsPerFeeCoin(
		juelsPerFeeCoin/multiply,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorSubmissionReceivedValues(
		latestAnswer/multiply,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		string(envelope.Transmitter),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorJuelsPerFeeCoinReceivedValues(
		juelsPerFeeCoin/multiply,
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		string(envelope.Transmitter),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetOffchainAggregatorRoundID(
		float64(envelope.AggregatorRoundID),
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
}

func (p *prometheusExporter) exportTxResults(res TxResults) {
	p.metrics.SetFeedContractTransactionsSucceeded(
		float64(res.NumSucceeded),
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
	p.metrics.SetFeedContractTransactionsFailed(
		float64(res.NumFailed),
		p.feedConfig.GetID(),
		p.feedConfig.GetID(),
		p.chainConfig.GetChainID(),
		p.feedConfig.GetContractStatus(),
		p.feedConfig.GetContractType(),
		p.feedConfig.GetName(),
		p.feedConfig.GetPath(),
		p.chainConfig.GetNetworkID(),
		p.chainConfig.GetNetworkName(),
	)
}

func (p *prometheusExporter) Cleanup(_ context.Context) {
	p.labelsMu.Lock()
	defer p.labelsMu.Unlock()
	for sender := range p.labels.senders {
		oracleName, found := getOracleName(types.Account(sender), p.nodes)
		if !found {
			oracleName = sender
		}
		p.metrics.Cleanup(
			p.labels.networkName,
			p.labels.networkID,
			p.labels.chainID,
			oracleName,
			sender,
			p.labels.feedName,
			p.labels.feedPath,
			p.labels.symbol,
			p.labels.contractType,
			p.labels.contractStatus,
			p.labels.contractAddress,
			p.labels.feedID,
		)
	}
}

// isNewTransmission considers four cases:
//   - old value == new value && old timestamp == new timestamp => return false
//   - old value != new value && old timestamp == new timestamp => This is probably and error since
//     any new transmission updates the timestamp as well, but, to record the observation, we return true.
//   - old value != new value && old timestamp != new timestamp => return true
//   - old value == new value && old timestamp != new timestamp => An unlikely case given the
//     high precision of observations but still a valid update. Return true
func (p *prometheusExporter) isNewTransmission(value *big.Int, timestamp time.Time) bool {
	p.prevMu.Lock()
	defer p.prevMu.Unlock()
	if value.Cmp(p.prevValue) == 0 && timestamp.Equal(p.prevTimestamp) {
		return false
	}
	p.prevValue = value
	p.prevTimestamp = timestamp
	return true
}

// Labels

// prometheusLabels is a helper which stores the labels used an instance of this exporter.
// They are useful at Cleanup time, when this exporter needs to delete all the labels it created.
type prometheusLabels struct {
	networkName     string
	networkID       string
	chainID         string
	sender          string
	senders         map[string]struct{} // A set of unique senders!
	feedName        string
	feedPath        string
	symbol          string
	contractType    string
	contractStatus  string
	contractAddress string
	feedID          string
}

func (p *prometheusExporter) updateLabels(newLabels prometheusLabels) {
	p.labelsMu.Lock()
	defer p.labelsMu.Unlock()
	if newLabels.networkName != "" {
		p.labels.networkName = newLabels.networkName
	}
	if newLabels.networkID != "" {
		p.labels.networkID = newLabels.networkID
	}
	if newLabels.chainID != "" {
		p.labels.chainID = newLabels.chainID
	}
	if p.labels.senders == nil {
		p.labels.senders = map[string]struct{}{}
	}
	if newLabels.sender != "" {
		p.labels.senders[newLabels.sender] = struct{}{}
	}
	if newLabels.feedName != "" {
		p.labels.feedName = newLabels.feedName
	}
	if newLabels.feedPath != "" {
		p.labels.feedPath = newLabels.feedPath
	}
	if newLabels.symbol != "" {
		p.labels.symbol = newLabels.symbol
	}
	if newLabels.contractType != "" {
		p.labels.contractType = newLabels.contractType
	}
	if newLabels.contractStatus != "" {
		p.labels.contractStatus = newLabels.contractStatus
	}
	if newLabels.contractAddress != "" {
		p.labels.contractAddress = newLabels.contractAddress
	}
	if newLabels.feedID != "" {
		p.labels.feedID = newLabels.feedID
	}
}

// Helpers

func getOracleName(account types.Account, nodes []NodeConfig) (string, bool) {
	for _, node := range nodes {
		if node.GetAccount() == account {
			return node.GetName(), true
		}
	}
	return "", false
}

func toFloat64(bignum *big.Int) float64 {
	val, _ := new(big.Float).SetInt(bignum).Float64()
	return val
}
