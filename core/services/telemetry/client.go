package telemetry

import (
	"context"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/wsrpc"
)

type Client struct {
	ctx     context.Context
	backend TelemetryClient

	sendCh         chan *TelemetryRequest
	bufferCapacity uint32

	dropMessageCount uint32

	log *logger.Logger
}

func NewClient(
	ctx context.Context,
	backend TelemetryClient,
	bufferCapacity uint32,
	log *logger.Logger,
) Client {
	c := Client{
		ctx,
		backend,
		make(chan *TelemetryRequest, bufferCapacity),
		bufferCapacity,
		0,
		log,
	}
	go c.run()
	return c
}

func (c Client) Send(req *TelemetryRequest) {
	select {
	case c.sendCh <- req:
	default:
		c.logBufferFullWithExpBackoff(req)
	}
}

func (c Client) run() {
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
			}
			c.log.Debug("delivered telemetry payload: %v", res.Body)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c Client) logBufferFullWithExpBackoff(req *TelemetryRequest) {
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
