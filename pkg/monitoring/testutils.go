// This file contains data generators and utilities to simplify tests.
// The data generated here shouldn't be used to run OCR instances
package monitoring

import (
	"context"
	cryptoRand "crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"math/rand"
	"net/http"
	"time"

	"github.com/linkedin/goavro/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/monitoring/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// Sources

func NewFakeRDDSource(minFeeds, maxFeeds uint8) Source {
	return &fakeRddSource{minFeeds, maxFeeds}
}

type fakeRddSource struct {
	minFeeds, maxFeeds uint8
}

func (f *fakeRddSource) Fetch(_ context.Context) (interface{}, error) {
	numFeeds := int(f.minFeeds) + rand.Intn(int(f.maxFeeds-f.minFeeds))
	feeds := make([]FeedConfig, numFeeds)
	for i := 0; i < numFeeds; i++ {
		feeds[i] = generateFeedConfig()
	}
	return feeds, nil
}

type fakeEnvelopeSourceFactory struct{}
type fakeTxResultsSourceFactory struct{}

var _ SourceFactory = (*fakeEnvelopeSourceFactory)(nil)
var _ SourceFactory = (*fakeTxResultsSourceFactory)(nil)

func (f *fakeEnvelopeSourceFactory) GetType() string {
	return "fake-envelope"
}
func (f *fakeTxResultsSourceFactory) GetType() string {
	return "fake-txresults"
}

func (f *fakeEnvelopeSourceFactory) NewSource(_ ChainConfig, _ FeedConfig) (Source, error) {
	return &fakeEnvelopeSource{}, nil
}
func (f *fakeTxResultsSourceFactory) NewSource(_ ChainConfig, _ FeedConfig) (Source, error) {
	return &fakeTxResultsSource{}, nil
}

type fakeEnvelopeSource struct{}
type fakeTxResultsSource struct{}

func (f *fakeEnvelopeSource) Fetch(ctx context.Context) (interface{}, error) {
	return generateEnvelope(ctx)
}
func (f *fakeTxResultsSource) Fetch(ctx context.Context) (interface{}, error) {
	return generateTxResults(), nil
}

type fakeRandomDataSourceFactory struct {
	updates chan interface{}
}

var _ SourceFactory = (*fakeRandomDataSourceFactory)(nil)

func (f *fakeRandomDataSourceFactory) NewSource(_ ChainConfig, _ FeedConfig) (Source, error) {
	return &fakeSource{f}, nil
}

func (f *fakeRandomDataSourceFactory) GetType() string {
	return "fake"
}

type fakeSource struct {
	factory *fakeRandomDataSourceFactory
}

func (f *fakeSource) Fetch(ctx context.Context) (interface{}, error) {
	select {
	case update := <-f.factory.updates:
		return update, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("source closed")
	}
}

type fakeSourceWithWait struct {
	waitOnRead time.Duration
}

func (f *fakeSourceWithWait) Fetch(ctx context.Context) (interface{}, error) {
	select {
	case <-time.After(f.waitOnRead):
		return 1, nil
	case <-ctx.Done():
		return 0, nil
	}
}

type fakeSourceFactoryWithError struct {
	updates     chan interface{}
	errors      chan error
	returnError bool
}

func (f *fakeSourceFactoryWithError) NewSource(_ ChainConfig, _ FeedConfig) (Source, error) {
	if f.returnError {
		return nil, fmt.Errorf("fake source factory error")
	}
	return &fakeSourceWithError{
		f.updates,
		f.errors,
	}, nil
}

func (f *fakeSourceFactoryWithError) GetType() string {
	return "fake-with-error"
}

type fakeSourceWithError struct {
	updates chan interface{}
	errors  chan error
}

func (f *fakeSourceWithError) Fetch(ctx context.Context) (interface{}, error) {
	select {
	case update := <-f.updates:
		return update, nil
	case err := <-f.errors:
		return nil, err
	case <-ctx.Done():
		return nil, nil
	}
}

type fakeSourceWithPanic struct {
	updates chan interface{}
	panics  chan error
}

func (f *fakeSourceWithPanic) Fetch(ctx context.Context) (interface{}, error) {
	select {
	case update := <-f.updates:
		return update, nil
	case err := <-f.panics:
		panic(err)
	case <-ctx.Done():
		return nil, nil
	}
}

// Exporters

type fakeExporterFactory struct {
	data        chan interface{}
	returnError bool
}

func (f *fakeExporterFactory) NewExporter(_ ExporterParams) (Exporter, error) {
	if f.returnError {
		return nil, fmt.Errorf("fake exporter factory error")
	}
	return &fakeExporter{
		f.data,
	}, nil
}

type fakeExporter struct {
	data chan interface{}
}

func (f *fakeExporter) Export(ctx context.Context, data interface{}) {
	select {
	case f.data <- data:
	case <-ctx.Done():
	}
}

func (f *fakeExporter) Cleanup(_ context.Context) {
}

// Generators

func generateBigInt(bitSize uint8) *big.Int {
	maxBigInt := new(big.Int)
	maxBigInt.Exp(big.NewInt(2), big.NewInt(int64(bitSize)), nil).Sub(maxBigInt, big.NewInt(1))

	//Generate cryptographically strong pseudo-random between 0 - max
	num, err := cryptoRand.Int(cryptoRand.Reader, maxBigInt)
	if err != nil {
		panic(fmt.Sprintf("failed to generate a really big number: %v", err))
	}
	return num
}

func generate32ByteArr() [32]byte {
	buf := make([]byte, 32)
	_, err := cryptoRand.Read(buf)
	if err != nil {
		panic("unable to generate [32]byte from rand")
	}
	var out [32]byte
	copy(out[:], buf[:32])
	return out
}

type fakeFeedConfig struct {
	Name           string `json:"name,omitempty"`
	Path           string `json:"path,omitempty"`
	Symbol         string `json:"symbol,omitempty"`
	HeartbeatSec   int64  `json:"heartbeat,omitempty"`
	ContractType   string `json:"contract_type,omitempty"`
	ContractStatus string `json:"status,omitempty"`
	// This functions as a feed identifier.
	ContractAddress        []byte   `json:"-"`
	ContractAddressEncoded string   `json:"contract_address_encoded,omitempty"`
	Multiply               *big.Int `json:"-"`
	MultiplyRaw            string   `json:"multiply,omitempty"`
}

func (f fakeFeedConfig) GetID() string             { return f.ContractAddressEncoded }
func (f fakeFeedConfig) GetName() string           { return f.Name }
func (f fakeFeedConfig) GetPath() string           { return f.Path }
func (f fakeFeedConfig) GetSymbol() string         { return f.Symbol }
func (f fakeFeedConfig) GetHeartbeatSec() int64    { return f.HeartbeatSec }
func (f fakeFeedConfig) GetContractType() string   { return f.ContractType }
func (f fakeFeedConfig) GetContractStatus() string { return f.ContractStatus }
func (f fakeFeedConfig) GetContractAddress() string {
	return base64.StdEncoding.EncodeToString(f.ContractAddress)
}
func (f fakeFeedConfig) GetContractAddressBytes() []byte { return f.ContractAddress }
func (f fakeFeedConfig) GetMultiply() *big.Int           { return f.Multiply }
func (f fakeFeedConfig) ToMapping() map[string]interface{} {
	return map[string]interface{}{
		"feed_name":               f.Name,
		"feed_path":               f.Path,
		"symbol":                  f.Symbol,
		"heartbeat_sec":           f.HeartbeatSec,
		"contract_type":           f.ContractType,
		"contract_status":         f.ContractStatus,
		"contract_address":        f.ContractAddress,
		"contract_address_string": map[string]interface{}{"string": f.ContractAddressEncoded},
		// These are solana specific but are kept here for backwards compatibility in Avro.
		"transmissions_account": []byte{},
		"state_account":         []byte{},
	}
}

func generateFeedConfig() FeedConfig {
	coins := []string{"btc", "eth", "matic", "link", "avax", "ftt", "srm", "usdc", "sol", "ray"}
	coin := coins[rand.Intn(len(coins))]
	contractAddress := generate32ByteArr()
	return fakeFeedConfig{
		Name:                   fmt.Sprintf("%s / usd", coin),
		Path:                   fmt.Sprintf("%s-usd", coin),
		Symbol:                 "$",
		HeartbeatSec:           1,
		ContractType:           "ocr2",
		ContractStatus:         "status",
		ContractAddress:        contractAddress[:],
		ContractAddressEncoded: hex.EncodeToString(contractAddress[:]),
		Multiply:               big.NewInt(10000),
		MultiplyRaw:            "10000",
	}
}

func fakeFeedsParser(buf io.ReadCloser) ([]FeedConfig, error) {
	rawFeeds := []fakeFeedConfig{}
	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(&rawFeeds); err != nil {
		return nil, fmt.Errorf("unable to unmarshal feeds config data: %w", err)
	}
	feeds := make([]FeedConfig, len(rawFeeds))
	for i, rawFeed := range rawFeeds {
		multiply, ok := new(big.Int).SetString(rawFeed.MultiplyRaw, 10)
		if !ok {
			return nil, fmt.Errorf("failed to parse multiply from '%s'", rawFeed.MultiplyRaw)
		}
		rawFeed.Multiply = multiply
		rawFeed.ContractAddress = []byte(rawFeed.ContractAddressEncoded)[:32]
		feeds[i] = FeedConfig(rawFeed)
	}
	return feeds, nil
}

type fakeNodeConfig struct {
	Name    string        `json:"name,omitempty"`
	Account types.Account `json:"account,omitempty"`
}

func (f fakeNodeConfig) GetName() string           { return f.Name }
func (f fakeNodeConfig) GetAccount() types.Account { return f.Account }

// hexEncode encodes b as a hex string with 0x prefix.
// Copied from github.com/ethereum/go-ethereum/common/hexutil.Encode
func hexEncode(b []byte) string {
	enc := make([]byte, len(b)*2+2)
	copy(enc, "0x")
	hex.Encode(enc[2:], b)
	return string(enc)
}

func generateNodeConfig() NodeConfig {
	id := uint8(rand.Intn(32))
	return fakeNodeConfig{
		fmt.Sprintf("noop-#%d", id),
		types.Account(hexEncode([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, id})),
	}
}

func fakeNodesParser(buf io.ReadCloser) ([]NodeConfig, error) {
	rawNodes := []fakeNodeConfig{}
	decoder := json.NewDecoder(buf)
	if err := decoder.Decode(&rawNodes); err != nil {
		return nil, fmt.Errorf("unable to unmarshal nodes config data: %w", err)
	}
	nodes := []NodeConfig{}
	for _, node := range rawNodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func generateNumericalMedianOffchainConfig() (*pb.NumericalMedianConfigProto, []byte, error) {
	out := &pb.NumericalMedianConfigProto{
		AlphaReportInfinite: ([]bool{true, false})[rand.Intn(2)],
		AlphaReportPpb:      rand.Uint64(),
		AlphaAcceptInfinite: ([]bool{true, false})[rand.Intn(2)],
		AlphaAcceptPpb:      rand.Uint64(),
		DeltaCNanoseconds:   rand.Uint64(),
	}
	buf, err := proto.Marshal(out)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal median plugin config: %w", err)
	}
	return out, buf, nil
}

func generateOffchainConfig(numOracles int) (
	*pb.OffchainConfigProto,
	*pb.NumericalMedianConfigProto,
	[]byte,
	error,
) {
	numericalMedianOffchainConfig, encodedNumericalMedianOffchainConfig, err := generateNumericalMedianOffchainConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	schedule := []uint32{}
	for i := 0; i < 10; i++ {
		schedule = append(schedule, 1)
	}
	offchainPublicKeys := [][]byte{}
	for i := 0; i < numOracles; i++ {
		randArr := generate32ByteArr()
		offchainPublicKeys = append(offchainPublicKeys, randArr[:])
	}
	peerIDs := []string{}
	for i := 0; i < numOracles; i++ {
		peerIDs = append(peerIDs, fmt.Sprintf("peer#%d", i))
	}
	config := &pb.OffchainConfigProto{
		DeltaProgressNanoseconds: rand.Uint64(),
		DeltaResendNanoseconds:   rand.Uint64(),
		DeltaRoundNanoseconds:    rand.Uint64(),
		DeltaGraceNanoseconds:    rand.Uint64(),
		DeltaStageNanoseconds:    rand.Uint64(),

		RMax:                  rand.Uint32(),
		S:                     schedule,
		OffchainPublicKeys:    offchainPublicKeys,
		PeerIds:               peerIDs,
		ReportingPluginConfig: encodedNumericalMedianOffchainConfig,

		MaxDurationQueryNanoseconds:       rand.Uint64(),
		MaxDurationObservationNanoseconds: rand.Uint64(),
		MaxDurationReportNanoseconds:      rand.Uint64(),

		MaxDurationShouldAcceptFinalizedReportNanoseconds:  rand.Uint64(),
		MaxDurationShouldTransmitAcceptedReportNanoseconds: rand.Uint64(),

		SharedSecretEncryptions: &pb.SharedSecretEncryptionsProto{
			DiffieHellmanPoint: []byte{'p', 'o', 'i', 'n', 't'},
			SharedSecretHash:   []byte{'h', 'a', 's', 'h'},
			Encryptions:        [][]byte{[]byte("encryption")},
		},
	}
	encodedConfig, err := proto.Marshal(config)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to marshal offchain config: %w", err)
	}
	return config, numericalMedianOffchainConfig, encodedConfig, nil
}

func generateContractConfig(ctx context.Context, n int) (
	types.ContractConfig,
	median.OnchainConfig,
	*pb.OffchainConfigProto,
	*pb.NumericalMedianConfigProto,
	error,
) {
	signers := make([]types.OnchainPublicKey, n)
	transmitters := make([]types.Account, n)
	for i := 0; i < n; i++ {
		randArr := generate32ByteArr()
		signers[i] = types.OnchainPublicKey(randArr[:])
		transmitters[i] = types.Account(hexEncode([]byte{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, uint8(i),
		}))
	}
	onchainConfig := median.OnchainConfig{
		Min: generateBigInt(128),
		Max: generateBigInt(128),
	}
	onchainConfigEncoded, err := median.StandardOnchainConfigCodec{}.Encode(ctx, onchainConfig)
	if err != nil {
		return types.ContractConfig{}, median.OnchainConfig{}, nil, nil, err
	}
	offchainConfig, pluginOffchainConfig, offchainConfigEncoded, err := generateOffchainConfig(n)
	if err != nil {
		return types.ContractConfig{}, median.OnchainConfig{}, nil, nil, err
	}
	contractConfig := types.ContractConfig{
		ConfigDigest:          generate32ByteArr(),
		ConfigCount:           rand.Uint64(),
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint8(10),
		OnchainConfig:         onchainConfigEncoded,
		OffchainConfigVersion: rand.Uint64(),
		OffchainConfig:        offchainConfigEncoded,
	}
	return contractConfig, onchainConfig, offchainConfig, pluginOffchainConfig, nil
}

func generateEnvelope(ctx context.Context) (Envelope, error) {
	generated, _, _, _, err := generateContractConfig(ctx, 31)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		ConfigDigest:    generated.ConfigDigest,
		Round:           uint8(rand.Intn(256)),
		Epoch:           rand.Uint32(),
		LatestAnswer:    generateBigInt(128),
		LatestTimestamp: time.Now(),

		ContractConfig: generated,

		BlockNumber: rand.Uint64(),
		Transmitter: types.Account(hexEncode([]byte{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, uint8(rand.Intn(32)),
		})),
		LinkBalance:             generateBigInt(150),
		LinkAvailableForPayment: generateBigInt(150),

		JuelsPerFeeCoin:   generateBigInt(150),
		AggregatorRoundID: rand.Uint32(),
	}, nil
}

func generateTxResults() TxResults {
	return TxResults{
		NumSucceeded: rand.Uint64(),
		NumFailed:    rand.Uint64(),
	}
}

type fakeChainConfig struct {
	RPCEndpoint  string
	NetworkName  string
	NetworkID    string
	ChainID      string
	ReadTimeout  time.Duration
	PollInterval time.Duration
}

func generateChainConfig() ChainConfig {
	return fakeChainConfig{
		RPCEndpoint:  "http://some-chain-host:6666",
		NetworkName:  "mainnet-beta",
		NetworkID:    "1",
		ChainID:      "mainnet-beta",
		ReadTimeout:  100 * time.Millisecond,
		PollInterval: time.Duration(1+rand.Intn(5)) * time.Second,
	}
}

func (f fakeChainConfig) GetRPCEndpoint() string         { return f.RPCEndpoint }
func (f fakeChainConfig) GetNetworkName() string         { return f.NetworkName }
func (f fakeChainConfig) GetNetworkID() string           { return f.NetworkID }
func (f fakeChainConfig) GetChainID() string             { return f.ChainID }
func (f fakeChainConfig) GetReadTimeout() time.Duration  { return f.ReadTimeout }
func (f fakeChainConfig) GetPollInterval() time.Duration { return f.PollInterval }

func (f fakeChainConfig) ToMapping() map[string]interface{} {
	return map[string]interface{}{
		"network_name": f.NetworkName,
		"network_id":   f.NetworkID,
		"chain_id":     f.ChainID,
	}
}

// Metrics

type devnullMetrics struct{}

var _ Metrics = (*devnullMetrics)(nil)

func (d *devnullMetrics) SetHeadTrackerCurrentHead(blockNumber float64, networkName, chainID, networkID string) {
}
func (d *devnullMetrics) SetFeedContractMetadata(chainID, contractAddress, feedID, contractStatus, contractType, feedName, feedPath, networkID, networkName, symbol string) {
}
func (d *devnullMetrics) SetFeedContractLinkBalance(balance float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetLinkAvailableForPayment(amount float64, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetFeedContractTransactionsSucceeded(numSucceeded float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetFeedContractTransactionsFailed(numFailed float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetNodeMetadata(chainID, networkID, networkName, oracleName, sender string) {
}
func (d *devnullMetrics) SetOffchainAggregatorAnswersRaw(answer float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorAnswers(answer float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) IncOffchainAggregatorAnswersTotal(contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorAnswersLatestTimestamp(latestTimestampSeconds float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorJuelsPerFeeCoinRaw(juelsPerFeeCoin float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorJuelsPerFeeCoin(juelsPerFeeCoin float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorSubmissionReceivedValues(value float64, contractAddress, feedID, sender, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorJuelsPerFeeCoinReceivedValues(value float64, contractAddress, feedID, sender, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorAnswerStalled(isSet bool, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) SetOffchainAggregatorRoundID(aggregatorRoundID float64, contractAddress, feedID, chainID, contractStatus, contractType, feedName, feedPath, networkID, networkName string) {
}
func (d *devnullMetrics) Cleanup(networkName, networkID, chainID, oracleName, sender, feedName, feedPath, symbol, contractType, contractStatus, contractAddress, feedID string) {
}

func (d *devnullMetrics) HTTPHandler() http.Handler {
	return promhttp.Handler()
}

// Producer

type producerMessage struct {
	key, value []byte
	topic      string
}

type fakeProducer struct {
	sendCh chan producerMessage
	stopCh services.StopChan
}

func (f fakeProducer) Close() error { close(f.stopCh); return nil }

func (f fakeProducer) Produce(key, value []byte, topic string) error {
	select {
	case f.sendCh <- producerMessage{key, value, topic}:
	case <-f.stopCh:
	}
	return nil
}

// Schema

type fakeSchema struct {
	codec   *goavro.Codec
	subject string
}

func (f fakeSchema) ID() int {
	return 1
}

func (f fakeSchema) Version() int {
	return 1
}

func (f fakeSchema) Subject() string {
	return f.subject
}

func (f fakeSchema) Encode(value interface{}) ([]byte, error) {
	return f.codec.BinaryFromNative(nil, value)
}

func (f fakeSchema) Decode(buf []byte) (interface{}, error) {
	value, _, err := f.codec.NativeFromBinary(buf)
	return value, err
}

// Poller

type fakePoller struct {
	numUpdates int
	ch         chan interface{}
}

func (f *fakePoller) Run(ctx context.Context) {
	source := &fakeRddSource{1, 2}
	for i := 0; i < f.numUpdates; i++ {
		updates, _ := source.Fetch(ctx)
		select {
		case f.ch <- updates:
		case <-ctx.Done():
			return
		}
	}
}

func (f *fakePoller) Updates() <-chan interface{} {
	return f.ch
}

func newNullLogger() Logger {
	return logger.Nop()
}

// This utilities are used primarely in tests but are present in the monitoring package because they are not inside a file ending in _test.go.
// This is done in order to expose NewRandomDataReader for use in cmd/monitoring.
// The following code is added to comply with the "unused" linter:
var (
	_ = generateChainConfig()
	_ = generateFeedConfig
	_ = fakeProducer{}
	_ = fakeSchema{}
	_ = fakePoller{}
	_ = newNullLogger()
	_ = fakeExporterFactory{}
	_ = fakeSourceWithWait{}
	_ = fakeSourceFactoryWithError{}
	_ = fakeSourceWithPanic{}
	_ = fakeFeedsParser
	_ = generateNodeConfig()
	_ = fakeNodesParser
	_ = fakeNodeConfig{}
)
