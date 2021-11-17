package monitoring

import (
	"context"
	"sync/atomic"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/wsrpc"
)

type Client interface {
	commontypes.MonitoringEndpoint

	Start()
	Close()
}

type client struct {
	ctx       context.Context
	cancelCtx context.CancelFunc

	address string

	otiConn WSRPCConnection

	sendLogCh      chan []byte
	bufferCapacity uint32

	logger           commontypes.Logger
	dropMessageCount uint32
}

func NewClient(
	otiConn WSRPCConnection,
	address string,
	bufferCapacity uint32,
	logger commontypes.Logger,
) Client {
	ctx, cancelCtx := context.WithCancel(context.Background())

	return &client{
		ctx,
		cancelCtx,
		address,
		otiConn,
		make(chan []byte, bufferCapacity),
		bufferCapacity,
		logger,
		0,
	}
}

// Start must be executed as a goroutine
func (c *client) Start() {
	telemetryClient := NewTelemetryClient(c.otiConn)
	for {
		select {
		case log := <-c.sendLogCh:
			res, err := telemetryClient.Telemetry(c.ctx, &TelemetryRequest{
				Telemetry: log,
				Address:   c.address,
			})
			if err == wsrpc.ErrNotConnected {
				c.logger.Warn("failed to deliver telemetry because the client is no longer connected", commontypes.LogFields{"error": err})
				return
			}
			if err != nil {
				c.logger.Error("failed to deliver telemetry:", commontypes.LogFields{"error": err})
			}
			c.logger.Debug("delivered telemetry payload", commontypes.LogFields{"response": res.Body})
		case <-c.ctx.Done():
			return
		}
	}
}

// Close must be called in a deferred statement.
func (c *client) Close() {
	c.cancelCtx()
	c.otiConn.Close()
}

// SendLog implements the commontypes.MonitoringEndpoint
func (c *client) SendLog(log []byte) {
	select {
	case c.sendLogCh <- log:
	default: // The channels is at capacity so we ditch the new message.
		c.logBufferFullWithExpBackoff(log)
	}
}

func (c *client) logBufferFullWithExpBackoff(payload []byte) {
	count := atomic.AddUint32(&c.dropMessageCount, 1)
	if count > 0 && (count%c.bufferCapacity == 0 || count&(count-1) == 0) {
		c.logger.Warn("client send buffer full, dropping telemetry", commontypes.LogFields{
			"payload":        string(payload),
			"droppedCount":   count,
			"bufferCapacity": c.bufferCapacity,
		})
	}
}
