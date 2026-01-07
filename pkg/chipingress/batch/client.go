package batch

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type Client struct {
	client             chipingress.Client
	batchSize          int
	maxConcurrentSends chan struct{}
	batchTimeout       time.Duration
	compressionType    string
	messageBuffer      chan *chipingress.CloudEventPb
	shutdownChan       chan struct{}
	log                *zap.SugaredLogger
}

type Opt func(*Client)

func NewBatchClient(client chipingress.Client, opts ...Opt) (*Client, error) {
	c := &Client{
		client:             client,
		batchSize:          1,
		maxConcurrentSends: make(chan struct{}, 1),
		messageBuffer:      make(chan *chipingress.CloudEventPb, 1000),
		batchTimeout:       100 * time.Millisecond,
		compressionType:    "gzip",
		shutdownChan:       make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func (b *Client) Start(ctx context.Context) {
	go func() {
		batch := make([]*chipingress.CloudEventPb, 0, b.batchSize)
		timer := time.NewTimer(b.batchTimeout)
		timer.Stop()

		for {
			select {
			case <-ctx.Done():
				b.flush(batch)
				close(b.shutdownChan)
				return
			case <-b.shutdownChan:
				b.flush(batch)
				return
			case event := <-b.messageBuffer:
				if len(batch) == 0 {
					timer.Reset(b.batchTimeout)
				}

				batch = append(batch, event)

				if len(batch) >= b.batchSize {
					batchToSend := batch
					batch = make([]*chipingress.CloudEventPb, 0, b.batchSize)
					timer.Stop()
					b.sendBatch(ctx, batchToSend)
				}
			case <-timer.C:
				if len(batch) > 0 {
					batchToSend := batch
					batch = make([]*chipingress.CloudEventPb, 0, b.batchSize)
					b.sendBatch(ctx, batchToSend)
				}
			}
		}
	}()
}

func (b *Client) Stop() {
	close(b.shutdownChan)
	// wait for pending sends by getting all semaphore slots
	for range cap(b.maxConcurrentSends) {
		b.maxConcurrentSends <- struct{}{}
	}
}

// QueueMessage queues a single message to the batch client.
// Returns immediately with no blocking - drops message if channel is full.
// Returns true if message was queued, false if it was dropped.
func (b *Client) QueueMessage(event *chipingress.CloudEventPb) bool {
	if event == nil {
		return false
	}

	select {
	case b.messageBuffer <- event:
		return true
	default:
		return false
	}
}

func (b *Client) sendBatch(ctx context.Context, events []*chipingress.CloudEventPb) {
	if len(events) == 0 {
		return
	}

	b.maxConcurrentSends <- struct{}{}

	go func() {
		defer func() { <-b.maxConcurrentSends }()
		_, err := b.client.PublishBatch(ctx, &chipingress.CloudEventBatch{Events: events})
		if err != nil {
			b.log.Errorw("failed to publish batch", "error", err)
		}
	}()
}

func (b *Client) flush(batch []*chipingress.CloudEventPb) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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

func WithBatchTimeout(batchTimeout time.Duration) Opt {
	return func(c *Client) {
		c.batchTimeout = batchTimeout
	}
}

func WithCompressionType(compressionType string) Opt {
	return func(c *Client) {
		c.compressionType = compressionType
	}
}

func WithMessageBuffer(messageBufferSize int) Opt {
	return func(c *Client) {
		c.messageBuffer = make(chan *chipingress.CloudEventPb, messageBufferSize)
	}
}
