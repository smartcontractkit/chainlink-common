package beholder

import (
	"context"
	"fmt"
	"log"
)

// dualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type DualSourceEmitter struct {
	chipIngressEmitter   Emitter
	otelCollectorEmitter Emitter
}

func NewDualSourceEmitter(chipIngressEmitter Emitter, otelCollectorEmitter Emitter) (Emitter, error) {

	if chipIngressEmitter == nil {
		return nil, fmt.Errorf("chip ingress emitter is nil")
	}

	if otelCollectorEmitter == nil {
		return nil, fmt.Errorf("otel collector emitter is nil")
	}

	return &DualSourceEmitter{
		chipIngressEmitter:   chipIngressEmitter,
		otelCollectorEmitter: otelCollectorEmitter,
	}, nil
}

func (d *DualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	// Emit via chip ingress async
	go func() {
		if err := d.chipIngressEmitter.Emit(ctx, body, attrKVs...); err != nil {
			// If the chip ingress emitter fails, we log the error but do not return it
			// because we still want to send the data to the OTLP collector and not cause disruption
			log.Printf("Error emitting message to chip ingress: %v", err)
		}
	}()

	return nil
}
