package beholder

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// DualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type DualSourceEmitter struct {
	chipIngressEmitter   Emitter
	otelCollectorEmitter Emitter
	log                  logger.Logger
	stopCh               services.StopChan
	wg                   services.WaitGroup
	closed               atomic.Bool
}

func NewDualSourceEmitter(chipIngressEmitter Emitter, otelCollectorEmitter Emitter) (Emitter, error) {

	if chipIngressEmitter == nil {
		return nil, fmt.Errorf("chip ingress emitter is nil")
	}

	if otelCollectorEmitter == nil {
		return nil, fmt.Errorf("otel collector emitter is nil")
	}

	logger, err := logger.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &DualSourceEmitter{
		chipIngressEmitter:   chipIngressEmitter,
		otelCollectorEmitter: otelCollectorEmitter,
		log:                  logger,
		stopCh:               make(services.StopChan),
	}, nil
}

func (d *DualSourceEmitter) Close() error {
	if wasClosed := d.closed.Swap(true); wasClosed {
		return errors.New("already closed")
	}
	close(d.stopCh)
	d.wg.Wait()
	return errors.Join(d.chipIngressEmitter.Close(), d.otelCollectorEmitter.Close())
}

func (d *DualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	// Emit via chip ingress async
	if err := d.wg.TryAdd(1); err != nil {
		return err
	}
	go func(ctx context.Context) {
		defer d.wg.Done()
		var cancel context.CancelFunc
		ctx, cancel = d.stopCh.Ctx(ctx)
		defer cancel()

		if err := d.chipIngressEmitter.Emit(ctx, body, attrKVs...); err != nil {
			// If the chip ingress emitter fails, we ONLY log the error
			// because we still want to send the data to the OTLP collector and not cause disruption
			d.log.Infof("failed to emit to chip ingress: %v", err)
		}
	}(context.WithoutCancel(ctx))

	return nil
}
