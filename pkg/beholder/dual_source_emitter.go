package beholder

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// dualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type dualSourceEmitter struct {
	chipIngressEmitter Emitter
	otelCollectorEmitter Emitter 
}

func NewDualSourceEmitter(chipIngressClient chipingress.ChipIngressClient, otelCollectorEmitter Emitter) Emitter {
	return &dualSourceEmitter{
		chipIngressEmitter: NewChipIngressEmitter(chipIngressClient),
		otelCollectorEmitter: otelCollectorEmitter,
	}
}

func (d *dualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	
	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}
	
	if err := d.chipIngressEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	return nil
}