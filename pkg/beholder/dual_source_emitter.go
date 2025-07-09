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
		d.log.Debugw("emitting to chip ingress", "body", string(body), "attributes", attrKVs)
		d.log.Debugw("overriding context for chip ingress emission", "timeout", 5*time.Second)
		ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := d.chipIngressEmitter.Emit(ctx2, body, attrKVs...); err != nil {
			// If the chip ingress emitter fails, we ONLY log the error
			// because we still want to send the data to the OTLP collector and not cause disruption
			d.log.Errorw("failed to emit to chip ingress", "error", err)
		}
	}()

	return nil
}
