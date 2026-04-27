package batch

import (
	"context"
	"errors"
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

type messageWithCallback struct {
	event    *chipingress.CloudEventPb
	callback func(error)
}

type seqnumKey struct {
	source    string
	eventType string
}

// Client is a batching client that accumulates messages and sends them in batches.
type Client struct {
	client             chipingress.Client
	batchSize          int
	maxGRPCRequestSize int
	cloneEvent         bool
	maxConcurrentSends chan struct{}
	batchInterval      time.Duration
	maxPublishTimeout  time.Duration
	messageBuffer      chan *messageWithCallback
	stopCh             stopCh
	log                *zap.SugaredLogger
	callbackWg         sync.WaitGroup
	shutdownTimeout    time.Duration
	shutdownOnce       sync.Once
	batcherDone        chan struct{}
	cancelBatcher      context.CancelFunc
	counters           sync.Map // map[seqnumKey]*atomic.Uint64 for per-(source,type) seqnum, cleared on Stop()

	metrics batchClientMetrics
}

type batchClientMetrics struct {
	sendRequestsTotal   otelmetric.Int64Counter
	sendFailuresTotal   otelmetric.Int64Counter
	requestSizeMessages otelmetric.Int64Histogram
	requestSizeBytes    otelmetric.Int64Histogram
	requestLatencyMS    otelmetric.Float64Histogram
	configInfo          otelmetric.Int64Gauge
	batchSizeAttr       otelmetric.MeasurementOption
	maxGRPCReqSizeAttr  otelmetric.MeasurementOption
	successStatusAttr   otelmetric.MeasurementOption
	failureStatusAttr   otelmetric.MeasurementOption
}

// Opt is a functional option for configuring the batch Client.
type Opt func(*Client)

// NewBatchClient creates a new batching client with the given options.
func NewBatchClient(client chipingress.Client, opts ...Opt) (*Client, error) {
	c := &Client{
		client:             client,
		log:                zap.NewNop().Sugar(),
		batchSize:          10,
		maxGRPCRequestSize: 16 * 1024 * 1024, // Match chipingress maxMessageSize default.
		cloneEvent:         true,
		maxConcurrentSends: make(chan struct{}, 1),
		messageBuffer:      make(chan *messageWithCallback, 200),
		batchInterval:      100 * time.Millisecond,
		maxPublishTimeout:  5 * time.Second,
		stopCh:             make(chan struct{}),
		callbackWg:         sync.WaitGroup{},
		shutdownTimeout:    5 * time.Second,
		batcherDone:        make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	var err error
	c.metrics, err = newBatchClientMetrics()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Start begins processing messages from the queue and sending them in batches
func (b *Client) Start(ctx context.Context) {
	b.metrics.recordConfig(context.Background(), b)

	// Create a cancellable context for the batcher
	batcherCtx, cancel := context.WithCancel(ctx)
	b.cancelBatcher = cancel

	go func() {
		defer close(b.batcherDone)

		go func() {
			select {
			case <-ctx.Done():
				b.Stop()
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

		if b.cancelBatcher != nil {
			b.cancelBatcher()
		}
		close(b.stopCh)

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

		// Release per-stream seqnum state to avoid unbounded growth from high-cardinality source/type values.
		b.clearCounters()
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
		return errors.New("client is shutdown")
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
		return errors.New("message buffer is full")
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
		// this is specifically to prevent long running network calls
		ctxTimeout, cancel := context.WithTimeout(ctx, b.maxPublishTimeout)
		defer cancel()

		events := make([]*chipingress.CloudEventPb, len(messages))
		for i, msg := range messages {
			events[i] = msg.event
		}
		batchReq := &chipingress.CloudEventBatch{Events: events}
		batchBytes := proto.Size(batchReq)
		startedAt := time.Now()
		_, err := b.client.PublishBatch(ctxTimeout, batchReq)
		b.metrics.recordSend(context.Background(), len(messages), batchBytes, time.Since(startedAt), err == nil)
		if err != nil {
			b.log.Errorw("failed to publish batch", "error", err)
		}
		// the callbacks are placed in their own goroutine to not block releasing the semaphore
		// we use a wait group, to ensure all callbacks are completed if  .Stop() is called.
		b.callbackWg.Go(func() {
			for _, msg := range messages {
				if msg.callback != nil {
					msg.callback(err)
				}
			}
		})
	}()
}

// WithBatchSize sets the number of messages to accumulate before sending a batch
func WithBatchSize(batchSize int) Opt {
	return func(c *Client) {
		c.batchSize = batchSize
	}
}

// WithMaxGRPCRequestSize sets the max gRPC request size in bytes used for metric comparison attributes.
func WithMaxGRPCRequestSize(maxReqSize int) Opt {
	return func(c *Client) {
		c.maxGRPCRequestSize = maxReqSize
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

func newBatchClientMetrics() (batchClientMetrics, error) {
	meter := otel.Meter("chipingress/batch_client")
	sendRequestsTotal, err := meter.Int64Counter(
		"chip_ingress.batch.send_requests_total",
		otelmetric.WithDescription("Total PublishBatch requests sent by batch client"),
		otelmetric.WithUnit("{request}"),
	)
	if err != nil {
		return batchClientMetrics{}, err
	}
	sendFailuresTotal, err := meter.Int64Counter(
		"chip_ingress.batch.send_failures_total",
		otelmetric.WithDescription("Total failed PublishBatch requests sent by batch client"),
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

	return batchClientMetrics{
		sendRequestsTotal:   sendRequestsTotal,
		sendFailuresTotal:   sendFailuresTotal,
		requestSizeMessages: requestSizeMessages,
		requestSizeBytes:    requestSizeBytes,
		requestLatencyMS:    requestLatencyMS,
		configInfo:          configInfo,
		successStatusAttr: otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("status", "success"),
		)),
		failureStatusAttr: otelmetric.WithAttributeSet(attribute.NewSet(
			attribute.String("status", "failure"),
		)),
	}, nil
}

func (m *batchClientMetrics) recordConfig(ctx context.Context, c *Client) {
	m.batchSizeAttr = otelmetric.WithAttributeSet(attribute.NewSet(
		attribute.Int("max_batch_size", c.batchSize),
	))
	m.maxGRPCReqSizeAttr = otelmetric.WithAttributeSet(attribute.NewSet(
		attribute.Int("max_grpc_request_size_bytes", c.maxGRPCRequestSize),
	))
	m.configInfo.Record(ctx, 1, otelmetric.WithAttributes(
		attribute.Int("max_batch_size", c.batchSize),
		attribute.Int("message_buffer_size", cap(c.messageBuffer)),
		attribute.Int("max_concurrent_sends", cap(c.maxConcurrentSends)),
		attribute.Int64("batch_interval_ms", c.batchInterval.Milliseconds()),
		attribute.Int64("max_publish_timeout_ms", c.maxPublishTimeout.Milliseconds()),
		attribute.Int64("shutdown_timeout_ms", c.shutdownTimeout.Milliseconds()),
		attribute.Bool("clone_event", c.cloneEvent),
		attribute.Int("max_grpc_request_size_bytes", c.maxGRPCRequestSize),
	))
}

func (m *batchClientMetrics) recordSend(ctx context.Context, messageCount int, requestBytes int, latency time.Duration, success bool) {
	statusAttr := m.successStatusAttr
	if !success {
		statusAttr = m.failureStatusAttr
	}
	m.sendRequestsTotal.Add(ctx, 1, statusAttr)
	if !success {
		m.sendFailuresTotal.Add(ctx, 1)
	}

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
