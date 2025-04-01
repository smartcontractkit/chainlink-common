package beholder

import (
	"context"
)

// dualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type DualSourceEmitter struct {
	chipIngressEmitter   Emitter
	otelCollectorEmitter Emitter
}

func NewDualSourceEmitter(chipIngressEmitter Emitter, otelCollectorEmitter Emitter) Emitter {
	return &DualSourceEmitter{
		chipIngressEmitter:   chipIngressEmitter,
		otelCollectorEmitter: otelCollectorEmitter,
	}
}

func (d *DualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	if err := d.chipIngressEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	return nil
}
