package beholder

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// dualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type DualSourceEmitter struct {
	chipIngressEmitter   Emitter
	otelCollectorEmitter Emitter
	log                  logger.Logger
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
	}, nil
}

func (d *DualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	// Emit via chip ingress async
	go func() {
		ctx2 := context.Background()
		ctx2, cancel := context.WithTimeout(ctx2, 1*time.Minute)
		defer cancel()
		d.log.Debugw("emitting to chip ingress", "body", string(body), "attrs", attrKVs)
		d.log.Debugw("chip ingress overriding timeout to 1 minute, original ctx", "ctx", ctx)
		if err := d.chipIngressEmitter.Emit(ctx2, body, attrKVs...); err != nil {
			// If the chip ingress emitter fails, we ONLY log the error
			// because we still want to send the data to the OTLP collector and not cause disruption
			// TODO prom metric so we can alert on this
			// TODO maybe we should have a retry mechanism here or instead the chip ingress emitter should have one?

			d.log.Errorw("failed to emit to chip ingress", "error", err)
		}
	}()

	return nil
}
