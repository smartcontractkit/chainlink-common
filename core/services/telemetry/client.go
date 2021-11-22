package telemetry

import (
	"context"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink-relay/core/services/telemetry/generated"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/wsrpc"
)

type client struct {
	ctx     context.Context
	backend generated.TelemetryClient

	sendCh         chan *generated.TelemetryRequest
	bufferCapacity uint32

	dropMessageCount uint32

	log *logger.Logger
}

func NewClient(
	ctx context.Context,
	backend generated.TelemetryClient,
	bufferCapacity uint32,
	log *logger.Logger,
) Client {
	c := &client{
		ctx,
		backend,
		make(chan *generated.TelemetryRequest, bufferCapacity),
		bufferCapacity,
		0,
		log,
	}
	go c.run()
	return c
}

func (c *client) Send(req *generated.TelemetryRequest) {
	select {
	case c.sendCh <- req:
	default:
		c.logBufferFullWithExpBackoff(req)
	}
}

func (c *client) run() {
	for {
		select {
		case req := <-c.sendCh:
			res, err := c.backend.Telemetry(c.ctx, req)
			if err == wsrpc.ErrNotConnected {
				c.log.Warnf("failed to deliver telemetry because the client is no longer connected: %v", err)
				return
			}
			if err != nil {
				c.log.Errorf("failed to deliver telemetry: %v", err)
			} else {
				c.log.Debug("delivered telemetry payload: %v", res.Body)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *client) logBufferFullWithExpBackoff(req *generated.TelemetryRequest) {
	count := atomic.AddUint32(&c.dropMessageCount, 1)
	if count > 0 && (count%c.bufferCapacity == 0 || count&(count-1) == 0) {
		c.log.Warnf("client send buffer full, dropping telemetry payload=%s address=%s droppedCount=%d bufferCapacity=%d",
			string(req.Telemetry),
			req.Address,
			count,
			c.bufferCapacity,
		)
	}
}
