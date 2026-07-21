package batch

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

var (
	ErrMessageBufferFull = errors.New("message buffer is full")
	ErrClientShutdown    = errors.New("client is shutdown")
)

type messageWithCallback struct {
	event    *chipingress.CloudEventPb
	callback func(error)
}

type seqnumKey struct {
	source    string
	eventType string
}

// PublishError is an error returned per-event when partial delivery is enabled
// and an individual event fails validation/production.
type PublishError struct {
	Code   chipingress.PublishErrorCode
	Reason string
}

func (e *PublishError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code.String(), e.Reason)
}

// Error codes returned by the server in PublishError.Code.
// Re-exported from the proto package for convenience.
const (
	ErrCodeUnknown                = chipingress.PublishErrorCode(0)  // PUBLISH_ERROR_CODE_UNKNOWN
	ErrCodeValidationFailed       = chipingress.PublishErrorCode(1)  // PUBLISH_ERROR_CODE_VALIDATION_FAILED
	ErrCodeSchemaMissing          = chipingress.PublishErrorCode(2)  // PUBLISH_ERROR_CODE_SCHEMA_MISSING
	ErrCodeEncodeError            = chipingress.PublishErrorCode(3)  // PUBLISH_ERROR_CODE_ENCODE_ERROR
	ErrCodeDomainMisconfiguration = chipingress.PublishErrorCode(4)  // PUBLISH_ERROR_CODE_DOMAIN_MISCONFIGURATION
	ErrCodeResultsMismatch        = chipingress.PublishErrorCode(-1) // client-side synthetic code
)

// Client is a batching client that accumulates messages and sends them in batches.
type Client struct {
	client                  chipingress.Client
	batchSize               int
	maxGRPCRequestSize      int // configured max, used for metrics/error reporting
	effectiveMaxRequestSize int // maxGRPCRequestSize minus grpcFramingOverhead, used for splitting
	cloneEvent              bool
	maxConcurrentSends      chan struct{}
	batchInterval           time.Duration
	maxPublishTimeout       time.Duration
	messageBuffer           chan *messageWithCallback
	stopCh                  stopCh
	log                     *zap.SugaredLogger
	callbackWg              sync.WaitGroup
	shutdownTimeout         time.Duration
	shutdownOnce            sync.Once
	batcherDone             chan struct{}
	started                 bool
	counters                sync.Map // map[seqnumKey]*atomic.Uint64 for per-(source,type) seqnum, cleared on Stop()
	clientName              string

	metrics batchClientMetrics

	transactionEnabled bool
}

type batchClientMetrics struct {
	sendRequestsTotal    otelmetric.Int64Counter
	requestSizeMessages  otelmetric.Int64Histogram
	requestSizeBytes     otelmetric.Int64Histogram
	requestLatencyMS     otelmetric.Float64Histogram
	configInfo           otelmetric.Int64Gauge
	batchSplitsTotal     otelmetric.Int64Counter
	resultsMismatchTotal otelmetric.Int64Counter
	batchSizeAttr        otelmetric.MeasurementOption
	maxGRPCReqSizeAttr   otelmetric.MeasurementOption
	successStatusAttr    otelmetric.MeasurementOption
	failureStatusAttr    otelmetric.MeasurementOption
	clientNameAttr       otelmetric.MeasurementOption
}

// Opt is a functional option for configuring the batch Client.
type Opt func(*Client)

// NewBatchClient creates a new batching client with the given options.
func NewBatchClient(client chipingress.Client, opts ...Opt) (*Client, error) {
	c := &Client{
		client:                  client,
		log:                     zap.NewNop().Sugar(),
		batchSize:               10,
		maxGRPCRequestSize:      10 * 1024 * 1024,
		effectiveMaxRequestSize: 10*1024*1024 - grpcFramingOverhead,
		cloneEvent:              true,
		transactionEnabled:      false,
		maxConcurrentSends:      make(chan struct{}, 1),
		messageBuffer:           make(chan *messageWithCallback, 200),
		batchInterval:           100 * time.Millisecond,
		maxPublishTimeout:       5 * time.Second,
		stopCh:                  make(chan struct{}),
		callbackWg:              sync.WaitGroup{},
		shutdownTimeout:         5 * time.Second,
		batcherDone:             make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	var err error
	c.metrics, err = newBatchClientMetrics(c.clientName)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start begins processing messages from the queue and sending them in batches.
// The context is used only for the initial metrics recording call and is NOT
// retained after Start returns. The client manages its own internal lifecycle
// context that is cancelled when Stop is called.
func (b *Client) Start(ctx context.Context) {
	b.metrics.recordConfig(ctx, b)
	b.started = true

	// Detach from the caller's cancellation but keep its values (trace IDs, etc.).
	// This avoids retaining a startup context whose cancellation we don't control.
	batcherCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))

	go func() {
		defer close(b.batcherDone)

		go func() {
			select {
			case <-b.stopCh:
				cancel()
			}
		}()

		batchWithInterval(
			batcherCtx,
			b.messageBuffer,
			b.batchSize,
			b.batchInterval,
			func(batch []*messageWithCallback) {
				// Detach from cancellation so final flush can still publish during shutdown.
				// sendBatch still enforces maxPublishTimeout for each publish call.
				b.sendBatch(context.WithoutCancel(batcherCtx), batch)
			},
		)
	}()
}

// Stop ensures:
// - current batch is flushed
// - all current network calls are completed
// - all callbacks are completed
// Forcibly shutdowns down after timeout if not completed.
func (b *Client) Stop() {
	b.shutdownOnce.Do(func() {
		// Use a standalone timeout context so the shutdown wait isn't cancelled
		// by close(b.stopCh) below.
		ctx, cancel := context.WithTimeout(context.Background(), b.shutdownTimeout)
		defer cancel()

		close(b.stopCh)

		// Only wait for the batcher goroutine when Start() was called;
		// otherwise batcherDone is never closed and we'd block until timeout.
		if b.started {
			done := make(chan struct{})
			go func() {
				<-b.batcherDone
				for range cap(b.maxConcurrentSends) {
					b.maxConcurrentSends <- struct{}{}
				}
				// wait for all callbacks to complete
				b.callbackWg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// All successfully shutdown
			case <-ctx.Done(): // timeout or context cancelled
				b.log.Warnw("timed out waiting for shutdown to finish, force closing", "timeout", b.shutdownTimeout)
			}
		}

		// Release per-stream seqnum state to avoid unbounded growth from high-cardinality source/type values.
		b.clearCounters()

		if err := b.client.Close(); err != nil {
			b.log.Warnw("failed to close chip ingress client", "error", err)
		}
	})
}

func (b *Client) clearCounters() {
	b.counters.Range(func(key, _ any) bool {
		b.counters.Delete(key)
		return true
	})
}

// seqnumFor returns the next sequence number for the given source+type pair.
// Each unique (source, type) pair has its own independent counter starting at 1.
func (b *Client) seqnumFor(source, typ string) uint64 {
	key := seqnumKey{source: source, eventType: typ}
	v, _ := b.counters.LoadOrStore(key, &atomic.Uint64{})
	return v.(*atomic.Uint64).Add(1)
}

// QueueMessage queues a single message to the batch client with an optional callback.
// The callback will be invoked after the batch containing this message is sent.
// The callback receives an error parameter (nil on success).
// Callbacks are invoked from goroutines
// Returns immediately with no blocking - drops message if channel is full.
// Returns an error if the message was dropped.
// QueueMessage stamps/overwrites the "seqnum" extension on the event it buffers.
// By default, it clones the input event first (WithEventClone(true)) so caller-owned
// objects are not mutated and queued snapshots remain immutable under pointer reuse.
// If cloning is disabled via WithEventClone(false), the caller event is mutated in place.
func (b *Client) QueueMessage(event *chipingress.CloudEventPb, callback func(error)) error {
	if event == nil {
		return nil
	}

	// Check shutdown first to avoid race with buffer send
	select {
	case <-b.stopCh:
		return ErrClientShutdown
	default:
	}

	eventToQueue := event
	if b.cloneEvent {
		// Clone the caller-owned event so queued messages keep an immutable seqnum snapshot.
		eventCopy, ok := proto.Clone(event).(*chipingress.CloudEventPb)
		if !ok {
			return errors.New("failed to clone event")
		}
		eventToQueue = eventCopy
	}

	// Stamp seqnum extension attribute using the event snapshot being queued.
	seq := b.seqnumFor(eventToQueue.Source, eventToQueue.Type)
	if eventToQueue.Attributes == nil {
		eventToQueue.Attributes = make(map[string]*cepb.CloudEventAttributeValue)
	}
	eventToQueue.Attributes["seqnum"] = &cepb.CloudEventAttributeValue{
		Attr: &cepb.CloudEventAttributeValue_CeString{
			CeString: strconv.FormatUint(seq, 10),
		},
	}

	msg := &messageWithCallback{
		event:    eventToQueue,
		callback: callback,
	}

	select {
	case b.messageBuffer <- msg:
		return nil
	default:
		return ErrMessageBufferFull
	}
}

func (b *Client) sendBatch(ctx context.Context, messages []*messageWithCallback) {
	if len(messages) == 0 {
		return
	}

	// acquire semaphore, limiting concurrent sends
	b.maxConcurrentSends <- struct{}{}

	go func() {
		defer func() { <-b.maxConcurrentSends }()

		splitBatches := splitMessagesByRequestSize(messages, b.effectiveMaxRequestSize, b.transactionEnabled)
		if len(splitBatches) > 1 {
			b.metrics.batchSplitsTotal.Add(ctx, 1, b.metrics.clientNameAddOpts()...)
		}
		for _, batchMessages := range splitBatches {
			batchReq, batchBytes := newBatchRequest(batchMessages, b.transactionEnabled)
			if b.maxGRPCRequestSize > 0 && batchBytes > b.maxGRPCRequestSize {
				err := fmt.Errorf("publish batch serialized size %d exceeds max gRPC request size %d", batchBytes, b.maxGRPCRequestSize)
				b.metrics.recordSend(ctx, len(batchMessages), batchBytes, 0, false)
				b.log.Errorw("failed to publish batch", "error", err)
				b.completeBatchCallbacks(batchMessages, err)
				continue
			}

			// this is specifically to prevent long running network calls
			ctxTimeout, cancel := context.WithTimeout(ctx, b.maxPublishTimeout)
			startedAt := time.Now()
			resp, err := b.client.PublishBatch(ctxTimeout, batchReq)
			cancel()

			b.metrics.recordSend(ctx, len(batchMessages), batchBytes, time.Since(startedAt), err == nil)
			if err != nil {
				b.log.Errorw("failed to publish batch", "error", err)
				b.completeBatchCallbacks(batchMessages, err)
			} else if !b.transactionEnabled && resp != nil && len(resp.Results) > 0 {
				b.completeBatchCallbacksFromResults(batchMessages, resp.Results)
			} else {
				b.completeBatchCallbacks(batchMessages, nil)
			}
		}
	}()
}

func (b *Client) completeBatchCallbacks(messages []*messageWithCallback, err error) {
	callbackMessages, callbackErr := messages, err
	// the callbacks are placed in their own goroutine to not block releasing the semaphore
	// we use a wait group, to ensure all callbacks are completed if  .Stop() is called.
	b.callbackWg.Go(func() {
		for _, msg := range callbackMessages {
			if msg.callback != nil {
				msg.callback(callbackErr)
			}
		}
	})
}

// completeBatchCallbacksFromResults dispatches per-event callbacks using the server's
// PublishResult slice. Results are matched to messages by index (server contract guarantees
// positional correspondence). If a result has a non-nil Error, the callback receives a
// PublishError; otherwise it receives nil.
//
// Defensive behaviour:
//   - If len(results) < len(messages): remaining callbacks get a synthetic RESULTS_MISMATCH error.
//   - If len(results) > len(messages): extras are ignored.
//   - If results[i].EventId != messages[i].event.Id: a warning is logged.
func (b *Client) completeBatchCallbacksFromResults(messages []*messageWithCallback, results []*chipingress.PublishResult) {
	if len(results) != len(messages) {
		b.log.Warnw("publish results length mismatch",
			"results", len(results),
			"messages", len(messages),
		)
		b.metrics.resultsMismatchTotal.Add(context.Background(), 1, b.metrics.clientNameAddOpts()...)
	}

	b.callbackWg.Go(func() {
		for i, msg := range messages {
			if msg.callback == nil {
				continue
			}
			if i >= len(results) {
				msg.callback(&PublishError{
					Code:   ErrCodeResultsMismatch,
					Reason: fmt.Sprintf("server returned %d results for %d events", len(results), len(messages)),
				})
				continue
			}
			result := results[i]
			if result == nil {
				msg.callback(nil)
				continue
			}
			// Defensive: warn on ID mismatch but still dispatch by index.
			if result.EventId != "" && msg.event.Id != "" && result.EventId != msg.event.Id {
				b.log.Warnw("publish result event_id mismatch at index",
					"index", i,
					"expected", msg.event.Id,
					"got", result.EventId,
				)
				b.metrics.resultsMismatchTotal.Add(context.Background(), 1, b.metrics.clientNameAddOpts()...)
			}
			if result.Error != nil {
				msg.callback(&PublishError{
					Code:   result.Error.ErrorCode,
					Reason: result.Error.Reason,
				})
			} else {
				msg.callback(nil)
			}
		}
	})
}

// grpcFramingOverhead accounts for gRPC framing, HTTP/2 headers, auth tokens,
// tracing metadata, and other per-request overhead not captured by proto.Size.
const grpcFramingOverhead = 10 * 1024 // 10 KiB

// minMaxGRPCRequestSize is the minimum allowed value for maxGRPCRequestSize.
// Values below this threshold are clamped to ensure the framing overhead
// reservation remains meaningful.
const minMaxGRPCRequestSize = 1024 * 1024 // 1 MiB

func splitMessagesByRequestSize(messages []*messageWithCallback, maxRequestSize int, transactionEnabled bool) [][]*messageWithCallback {
	if len(messages) == 0 {
		return nil
	}
	if maxRequestSize <= 0 {
		return [][]*messageWithCallback{messages}
	}

	var batches [][]*messageWithCallback
	current := make([]*messageWithCallback, 0, len(messages))
	for _, msg := range messages {
		candidate := append(current, msg)
		_, candidateBytes := newBatchRequest(candidate, transactionEnabled)
		if len(current) > 0 && candidateBytes > maxRequestSize {
			batches = append(batches, current)
			current = []*messageWithCallback{msg}
			continue
		}
		current = candidate
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches
}

func newBatchRequest(messages []*messageWithCallback, transactionEnabled bool) (*chipingress.CloudEventBatch, int) {
	events := make([]*chipingress.CloudEventPb, len(messages))
	for i, msg := range messages {
		events[i] = msg.event
	}
	// Always emit PublishOptions so the wire form unambiguously reflects
	// client intent. The server treats unset and explicit false identically,
	// but explicit-false is defensive against any future server-default drift
	// and makes traces/logs self-describing.
	te := transactionEnabled
	batchReq := &chipingress.CloudEventBatch{
		Events:  events,
		Options: &chipingress.PublishOptions{TransactionEnabled: &te},
	}
	return batchReq, proto.Size(batchReq)
}

// WithBatchSize sets the number of messages to accumulate before sending a batch
func WithBatchSize(batchSize int) Opt {
	return func(c *Client) {
		c.batchSize = batchSize
	}
}

// WithMaxGRPCRequestSize sets the max gRPC request size in bytes used for splitting batches.
// Values below minMaxGRPCRequestSize (1 MiB) are clamped up to ensure the framing
// overhead reservation remains meaningful.
func WithMaxGRPCRequestSize(maxReqSize int) Opt {
	return func(c *Client) {
		clamped := max(maxReqSize, minMaxGRPCRequestSize)
		c.maxGRPCRequestSize = clamped
		c.effectiveMaxRequestSize = clamped - grpcFramingOverhead
	}
}

// WithEventClone controls whether QueueMessage clones events before stamping seqnum and buffering.
// Defaults to true for safety when caller reuses event pointers.
func WithEventClone(clone bool) Opt {
	return func(c *Client) {
		c.cloneEvent = clone
	}
}

// WithMaxConcurrentSends sets the maximum number of concurrent batch send operations
func WithMaxConcurrentSends(maxConcurrentSends int) Opt {
	return func(c *Client) {
		c.maxConcurrentSends = make(chan struct{}, maxConcurrentSends)
	}
}

// WithBatchInterval sets the maximum time to wait before sending an incomplete batch
func WithBatchInterval(batchTimeout time.Duration) Opt {
	return func(c *Client) {
		c.batchInterval = batchTimeout
	}
}

// WithShutdownTimeout sets the maximum time to wait for shutdown to complete
func WithShutdownTimeout(shutdownTimeout time.Duration) Opt {
	return func(c *Client) {
		c.shutdownTimeout = shutdownTimeout
	}
}

// WithMessageBuffer sets the size of the message queue buffer
func WithMessageBuffer(messageBufferSize int) Opt {
	return func(c *Client) {
		c.messageBuffer = make(chan *messageWithCallback, messageBufferSize)
	}
}

// WithMaxPublishTimeout sets the maximum time to wait for a batch publish operation
func WithMaxPublishTimeout(maxPublishTimeout time.Duration) Opt {
	return func(c *Client) {
		c.maxPublishTimeout = maxPublishTimeout
	}
}

// WithLogger sets the logger for the batch client
func WithLogger(log *zap.SugaredLogger) Opt {
	return func(c *Client) {
		c.log = log
	}
}

// Well-known client_name metric label values for batch client metrics.
// Pass to WithClientName when wiring batch.NewBatchClient.
const (
	ClientNameBeholder       = "beholder"
	ClientNameDurableEmitter = "durable_emitter"
)

// WithClientName sets client_name on batch client metrics. Omitted when unset.
func WithClientName(name string) Opt {
	return func(c *Client) {
		c.clientName = name
	}
}

// WithTransactionEnabled sets PublishOptions.transaction_enabled on every
// batch request. The option is always emitted on the wire so client intent
// is explicit in traces/logs; the server treats unset and explicit false
// identically (partial delivery).
//   - false (the default for NewBatchClient): partial delivery. Valid events
//     are produced and per-event errors are returned for invalid ones rather
//     than failing the entire batch.
//   - true: all-or-nothing. Any per-event failure fails the entire batch.
func WithTransactionEnabled(transactionEnabled bool) Opt {
	return func(c *Client) {
		c.transactionEnabled = transactionEnabled
	}
}

func batchMetricAttributeSet(clientName string, kvs ...attribute.KeyValue) attribute.Set {
	if clientName != "" {
		kvs = append(kvs, attribute.String("client_name", clientName))
	}
	return attribute.NewSet(kvs...)
}

func newBatchClientMetrics(clientName string) (batchClientMetrics, error) {
	meter := otel.Meter("chipingress/batch_client")
	sendRequestsTotal, err := meter.Int64Counter(
		"chip_ingress.batch.send_requests_total",
		otelmetric.WithDescription("Total PublishBatch requests sent by batch client"),
		otelmetric.WithUnit("{request}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	requestSizeMessages, err := meter.Int64Histogram(
		"chip_ingress.batch.request_size_messages",
		otelmetric.WithDescription("PublishBatch request size measured in number of events"),
		otelmetric.WithUnit("{event}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	requestSizeBytes, err := meter.Int64Histogram(
		"chip_ingress.batch.request_size_bytes",
		otelmetric.WithDescription("PublishBatch request size measured in bytes"),
		otelmetric.WithUnit("By"),
		otelmetric.WithExplicitBucketBoundaries(
			// Buckets from 1 KiB to 10 MiB (default maxGRPCRequestSize).
			1*1024, 4*1024, 16*1024, 64*1024, 256*1024,
			512*1024, 1*1024*1024, 2*1024*1024, 4*1024*1024,
			8*1024*1024, 10*1024*1024,
		),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	requestLatencyMS, err := meter.Float64Histogram(
		"chip_ingress.batch.request_latency_ms",
		otelmetric.WithDescription("PublishBatch end-to-end latency in milliseconds"),
		otelmetric.WithUnit("ms"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	configInfo, err := meter.Int64Gauge(
		"chip_ingress.batch.config.info",
		otelmetric.WithDescription("Batch client configuration info metric"),
		otelmetric.WithUnit("{info}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	batchSplitsTotal, err := meter.Int64Counter(
		"chip_ingress.batch.batch_splits_total",
		otelmetric.WithDescription("Total number of times a batch was split due to exceeding the effective gRPC request size limit (max request size minus reserved framing overhead)"),
		otelmetric.WithUnit("{split}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	resultsMismatchTotal, err := meter.Int64Counter(
		"chip_ingress.batch.results_mismatch_total",
		otelmetric.WithDescription("Total publish responses where result count or event IDs did not match the request"),
		otelmetric.WithUnit("{mismatch}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}

	var clientNameAttr otelmetric.MeasurementOption
	if clientName != "" {
		clientNameAttr = otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("client_name", clientName),
		))
	}

	return batchClientMetrics{
		sendRequestsTotal:    sendRequestsTotal,
		requestSizeMessages:  requestSizeMessages,
		requestSizeBytes:     requestSizeBytes,
		requestLatencyMS:     requestLatencyMS,
		configInfo:           configInfo,
		batchSplitsTotal:     batchSplitsTotal,
		resultsMismatchTotal: resultsMismatchTotal,
		clientNameAttr:       clientNameAttr,
		successStatusAttr: otelmetric.WithAttributeSet(batchMetricAttributeSet(clientName,
			attribute.String("status", "success"),
		)),
		failureStatusAttr: otelmetric.WithAttributeSet(batchMetricAttributeSet(clientName,
			attribute.String("status", "failure"),
		)),
	}, nil
}

func (m *batchClientMetrics) clientNameAddOpts() []otelmetric.AddOption {
	if m.clientNameAttr != nil {
		return []otelmetric.AddOption{m.clientNameAttr}
	}
	return nil
}

func (m *batchClientMetrics) recordConfig(ctx context.Context, c *Client) {
	clientName := c.clientName
	m.batchSizeAttr = otelmetric.WithAttributeSet(batchMetricAttributeSet(clientName,
		attribute.Int("max_batch_size", c.batchSize),
	))
	m.maxGRPCReqSizeAttr = otelmetric.WithAttributeSet(batchMetricAttributeSet(clientName,
		attribute.Int("max_grpc_request_size_bytes", c.maxGRPCRequestSize),
	))
	m.configInfo.Record(ctx, 1, otelmetric.WithAttributeSet(batchMetricAttributeSet(clientName,
		attribute.Int("max_batch_size", c.batchSize),
		attribute.Int("message_buffer_size", cap(c.messageBuffer)),
		attribute.Int("max_concurrent_sends", cap(c.maxConcurrentSends)),
		attribute.Int64("batch_interval_ms", c.batchInterval.Milliseconds()),
		attribute.Int64("max_publish_timeout_ms", c.maxPublishTimeout.Milliseconds()),
		attribute.Int64("shutdown_timeout_ms", c.shutdownTimeout.Milliseconds()),
		attribute.Bool("clone_event", c.cloneEvent),
		attribute.Bool("transaction_enabled", c.transactionEnabled),
		attribute.Int("max_grpc_request_size_bytes", c.maxGRPCRequestSize),
	)))
}

func (m *batchClientMetrics) recordSend(ctx context.Context, messageCount int, requestBytes int, latency time.Duration, success bool) {
	statusAttr := m.successStatusAttr
	if !success {
		statusAttr = m.failureStatusAttr
	}
	m.sendRequestsTotal.Add(ctx, 1, statusAttr)

	messageSizeOpts := []otelmetric.RecordOption{}
	if m.batchSizeAttr != nil {
		messageSizeOpts = append(messageSizeOpts, m.batchSizeAttr)
	}
	requestSizeOpts := []otelmetric.RecordOption{}
	if m.maxGRPCReqSizeAttr != nil {
		requestSizeOpts = append(requestSizeOpts, m.maxGRPCReqSizeAttr)
	}

	m.requestSizeMessages.Record(ctx, int64(messageCount), messageSizeOpts...)
	m.requestSizeBytes.Record(ctx, int64(requestBytes), requestSizeOpts...)
	m.requestLatencyMS.Record(ctx, float64(latency)/float64(time.Millisecond), statusAttr)
}
