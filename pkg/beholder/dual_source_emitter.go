package beholder

import (
	"context"
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// DualSourceEmitter emits both to chip ingress and to the otel collector
// this is to help transition from sending custom messages via OTLP to instead use chip-ingress
// we want to send to both during the transition period, then cutover to using
// chipIngressEmitter only
type DualSourceEmitter struct {
	chipIngressEmitter   Emitter
	otelCollectorEmitter Emitter
	lggr                 logger.Logger
}

func NewDualSourceEmitter(chipIngressEmitter Emitter, otelCollectorEmitter Emitter) (Emitter, error) {
	return DualSourceEmitterConfig{}.New(chipIngressEmitter, otelCollectorEmitter)
}

// DualSourceEmitterConfig holds configuration for creating a DualSourceEmitter.
type DualSourceEmitterConfig struct {
	Lggr logger.Logger
}

// New creates a DualSourceEmitter from the config.
func (c DualSourceEmitterConfig) New(chipIngressEmitter Emitter, otelCollectorEmitter Emitter) (Emitter, error) {
	if chipIngressEmitter == nil {
		return nil, errors.New("chip ingress emitter is nil")
	}

	if otelCollectorEmitter == nil {
		return nil, errors.New("otel collector emitter is nil")
	}

	lggr := c.Lggr
	if lggr == nil {
		lggr = logger.Nop()
	}

	return &DualSourceEmitter{
		chipIngressEmitter:   chipIngressEmitter,
		otelCollectorEmitter: otelCollectorEmitter,
		lggr:                 lggr,
	}, nil
}

func (d *DualSourceEmitter) Close() error {
	return errors.Join(d.chipIngressEmitter.Close(), d.otelCollectorEmitter.Close())
}

func (d *DualSourceEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	// Emit via OTLP first
	if err := d.otelCollectorEmitter.Emit(ctx, body, attrKVs...); err != nil {
		return err
	}

	if err := d.chipIngressEmitter.Emit(ctx, body, attrKVs...); err != nil {
		d.lggr.Infof("failed to emit to chip ingress: %v", err)
	}

	return nil
}
