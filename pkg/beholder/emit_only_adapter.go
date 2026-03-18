package beholder

import "context"

// emitOnlyAdapter wraps a ChipIngressBatchEmitterService as an Emitter but with a
// no-op Close. This decouples emission (used by DualSourceEmitter) from
// lifecycle management (owned by the application's service list).
type emitOnlyAdapter struct {
	e *ChipIngressBatchEmitterService
}

func (a *emitOnlyAdapter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {
	return a.e.Emit(ctx, body, attrKVs...)
}

func (a *emitOnlyAdapter) Close() error {
	return nil // lifecycle is managed externally
}
