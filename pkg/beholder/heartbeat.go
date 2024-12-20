package beholder

import (
	"context"
	"errors"
	"time"

	"github.com/jonboulle/clockwork"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type Runnable interface {
	Start(context.Context) error
	Close() error
}

// Heartbeat implements Runable interface
type Heartbeat struct {
	interval   time.Duration
	counter    metric.Int64Counter
	attributes []attribute.KeyValue
	log        logger.Logger
	done       chan struct{}
	clock      clockwork.Clock
	tickCh     chan struct{}
}

func NewHeartbeat(client *Client, interval time.Duration, log logger.Logger, clock clockwork.Clock) (*Heartbeat, error) {
	if client == nil || interval == 0 {
		return nil, nil
	}
	meter := client.MeterProvider.Meter("beholder_heartbeat")
	counter, err := meter.Int64Counter("beholder_heartbeat_counter")
	if err != nil {
		return nil, err
	}
	return &Heartbeat{
		interval:   interval,
		counter:    counter,
		log:        log,
		attributes: getBuildAttributes(),
		clock:      clock,
	}, nil
}

func (h *Heartbeat) Start(ctx context.Context) error {
	if h == nil || h.interval == 0 {
		return nil
	}
	if h.done != nil {
		// Already started
		return errors.New("heartbeat already started")
	}
	h.done = make(chan struct{}, 1)
	h.tickCh = make(chan struct{}, 1)
	go func() {
		h.log.Info("Beholder heartbeat started")
		ticker := h.clock.NewTicker(h.interval)
		defer ticker.Stop()
		for {
			println("loop")
			select {
			case <-ctx.Done():
				h.log.Debug("Beholder heartbeat stopped")
				h.Close()
				return
			case <-h.done:
				h.log.Debug("Beholder heartbeat stopped")
				return
			case <-ticker.Chan():
				h.log.Info("Beholder heartbeat sent")
				h.counter.Add(ctx, 1, metric.WithAttributes(h.attributes...))
			case <-h.tickCh:
				h.log.Info("Beholder heartbeat sent")
				h.counter.Add(ctx, 1, metric.WithAttributes(h.attributes...))
			}
		}
	}()
	return nil
}

// Safe to call multiple times
func (h *Heartbeat) Close() error {
	if h == nil || h.interval == 0 || h.done == nil {
		return nil
	}
	select {
	case <-h.done:
		return errors.New("heartbeat already closed")
	default:
		close(h.done)
	}
	return nil
}

func (h *Heartbeat) Send() {
	if h == nil || h.interval == 0 || h.tickCh == nil {
		return
	}
	go func() {
		h.tickCh <- struct{}{}
	}()
}

func getBuildAttributes() []attribute.KeyValue {
	buildInfo := getBuildInfoOnce()
	// TODO: add these to beholder resource attributes
	return []attribute.KeyValue{
		attribute.String("build_beholder_sdk_version", nonEmpty(buildInfo.sdkVersion)),
		attribute.String("build_version", nonEmpty(buildInfo.mainVersion)),
		attribute.String("build_path", nonEmpty(buildInfo.mainPath)),
		attribute.String("build_commit", nonEmpty(buildInfo.mainCommit)),
	}
}

func nonEmpty(str string) string {
	if str == "" {
		return unknown
	}
	return str
}
