package batch

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type messageWithCallback struct {
	event    *chipingress.CloudEventPb
	callback func(error)
}

type Client struct {
	client             chipingress.Client
	batchSize          int
	maxConcurrentSends chan struct{}
	batchInterval      time.Duration
	maxPublishTimeout  time.Duration
	messageBuffer      chan *messageWithCallback
	shutdownChan       chan struct{}
	log                *zap.SugaredLogger
	callbackWg         sync.WaitGroup
	shutdownTimeout    time.Duration
	shutdownOnce       sync.Once
}

type Opt func(*Client)

func NewBatchClient(client chipingress.Client, opts ...Opt) (*Client, error) {
	c := &Client{
		client:             client,
		log:                zap.NewNop().Sugar(),
		batchSize:          10,
		maxConcurrentSends: make(chan struct{}, 1),
		messageBuffer:      make(chan *messageWithCallback, 200),
		batchInterval:      100 * time.Millisecond,
		maxPublishTimeout:  5 * time.Second,
		shutdownChan:       make(chan struct{}),
		callbackWg:         sync.WaitGroup{},
		shutdownTimeout:    5 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (b *Client) Start(ctx context.Context) {
	go func() {
		batch := make([]*messageWithCallback, 0, b.batchSize)
		timer := time.NewTimer(b.batchInterval)
		timer.Stop()

		for {
			select {
			case <-ctx.Done():
				b.flush(batch)
				return
			case <-b.shutdownChan:
				b.flush(batch)
				return
			case msg := <-b.messageBuffer:
				if len(batch) == 0 {
					timer.Reset(b.batchInterval)
				}

				batch = append(batch, msg)

				if len(batch) >= b.batchSize {
					batchToSend := batch
					batch = make([]*messageWithCallback, 0, b.batchSize)
					timer.Stop()
					b.sendBatch(ctx, batchToSend)
				}
			case <-timer.C:
				if len(batch) > 0 {
					batchToSend := batch
					batch = make([]*messageWithCallback, 0, b.batchSize)
					b.sendBatch(ctx, batchToSend)
				}
			}
		}
	}()
}

func (b *Client) Stop() {
	b.shutdownOnce.Do(func() {
		close(b.shutdownChan)
		// wait for pending sends by getting all semaphore slots
		for range cap(b.maxConcurrentSends) {
			b.maxConcurrentSends <- struct{}{}
		}
		// wait for all callbacks to complete with timeout
		done := make(chan struct{})
		go func() {
			b.callbackWg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// All callbacks completed successfully
		case <-time.After(b.shutdownTimeout):
			b.log.Warnw("timed out waiting for callbacks to complete", "timeout", b.shutdownTimeout)
		}
	})
}

// QueueMessage queues a single message to the batch client with an optional callback.
// The callback will be invoked after the batch containing this message is sent.
// The callback receives an error parameter (nil on success).
// Callbacks are invoked from goroutines
// Returns immediately with no blocking - drops message if channel is full.
// Returns true if message was queued, false if it was dropped.
func (b *Client) QueueMessage(event *chipingress.CloudEventPb, callback func(error)) bool {
	if event == nil {
		return false
	}

	msg := &messageWithCallback{
		event:    event,
		callback: callback,
	}

	select {
	case b.messageBuffer <- msg:
		return true
	default:
		return false
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

	ctx, cancel := context.WithTimeout(context.Background(), b.maxPublishTimeout)
	defer cancel()

	b.sendBatch(ctx, batch)
}

func WithBatchSize(batchSize int) Opt {
	return func(c *Client) {
		c.batchSize = batchSize
	}
}

func WithMaxConcurrentSends(maxConcurrentSends int) Opt {
	return func(c *Client) {
		c.maxConcurrentSends = make(chan struct{}, maxConcurrentSends)
	}
}

func WithBatchInterval(batchTimeout time.Duration) Opt {
	return func(c *Client) {
		c.batchInterval = batchTimeout
	}
}

func WithShutdownTimeout(shutdownTimeout time.Duration) Opt {
	return func(c *Client) {
		c.shutdownTimeout = shutdownTimeout
	}
}

func WithMessageBuffer(messageBufferSize int) Opt {
	return func(c *Client) {
		c.messageBuffer = make(chan *messageWithCallback, messageBufferSize)
	}
}

func WithMaxPublishTimeout(maxPublishTimeout time.Duration) Opt {
	return func(c *Client) {
		c.maxPublishTimeout = maxPublishTimeout
	}
}

func WithLogger(log *zap.SugaredLogger) Opt {
	return func(c *Client) {
		c.log = log
	}
}
