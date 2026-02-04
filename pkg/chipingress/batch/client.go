package batch

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// Client is a batching client that accumulates messages and sends them in batches.
type Client struct {
	client             chipingress.Client
	batchSize          int
	maxConcurrentSends chan struct{}
	batchInterval      time.Duration
	maxPublishTimeout  time.Duration
	messageBuffer      chan *messageWithCallback
	stopCh             stopCh
	log                *zap.SugaredLogger
	callbackWg         sync.WaitGroup
	shutdownTimeout    time.Duration
	shutdownOnce       sync.Once
	batch              *buffer
}

// Opt is a functional option for configuring the batch Client.
type Opt func(*Client)

// NewBatchClient creates a new batching client with the given options.
func NewBatchClient(client chipingress.Client, opts ...Opt) (*Client, error) {
	c := &Client{
		client:             client,
		log:                zap.NewNop().Sugar(),
		batchSize:          10,
		maxConcurrentSends: make(chan struct{}, 1),
		messageBuffer:      make(chan *messageWithCallback, 200),
		batchInterval:      100 * time.Millisecond,
		maxPublishTimeout:  5 * time.Second,
		stopCh:             make(chan struct{}),
		callbackWg:         sync.WaitGroup{},
		shutdownTimeout:    5 * time.Second,
		batch:              newBuffer(10),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Start begins processing messages from the queue and sending them in batches
func (b *Client) Start(ctx context.Context) {
	go func() {
		timer := time.NewTimer(b.batchInterval)
		timer.Stop()

		for {
			select {
			case <-ctx.Done():
				// ensure:
				// - current batch is flushed
				// - all current network calls are completed
				// - all callbacks are completed
				b.Stop()
				return
			case <-b.stopCh: // this can only happen if .Stop()
				// since this was called from Stop, the remaining batch will be flushed alredy
				return
			case msg := <-b.messageBuffer:
				if b.batch.Len() == 0 {
					timer.Reset(b.batchInterval)
				}

				b.batch.Add(msg)

				if b.batch.Len() >= b.batchSize {
					batchToSend := b.batch.Clear()
					timer.Stop()
					b.sendBatch(ctx, batchToSend)
				}
			case <-timer.C:
				if b.batch.Len() > 0 {
					batchToSend := b.batch.Clear()
					b.sendBatch(ctx, batchToSend)
				}
			}
		}
	}()
}

// Stop ensures:
// - current batch is flushed
// - all current network calls are completed
// - all callbacks are completed
// Forcibly shutdowns down after timeout if not completed.
func (b *Client) Stop() {
	b.shutdownOnce.Do(func() {
		ctx, cancel := b.stopCh.CtxWithTimeout(b.shutdownTimeout)
		defer cancel()

		close(b.stopCh)

		done := make(chan struct{})
		go func() {
			// flush remaining batch
			b.flush(b.batch.Clear())
			// wait for pending sends by getting all semaphore slots
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
	})
}

// QueueMessage queues a single message to the batch client with an optional callback.
// The callback will be invoked after the batch containing this message is sent.
// The callback receives an error parameter (nil on success).
// Callbacks are invoked from goroutines
// Returns immediately with no blocking - drops message if channel is full.
// Returns an error if the message was dropped.
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

	msg := &messageWithCallback{
		event:    event,
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

		_, err := b.client.PublishBatch(ctxTimeout, &chipingress.CloudEventBatch{Events: events})
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

func (b *Client) flush(batch []*messageWithCallback) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := b.stopCh.CtxWithTimeout(b.maxPublishTimeout)
	defer cancel()

	b.sendBatch(ctx, batch)
}

// WithBatchSize sets the number of messages to accumulate before sending a batch
func WithBatchSize(batchSize int) Opt {
	return func(c *Client) {
		c.batchSize = batchSize
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
