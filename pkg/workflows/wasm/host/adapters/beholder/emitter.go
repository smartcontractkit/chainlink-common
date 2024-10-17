package beholder

import (
	"context"

	b "github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

type emitter struct{}

func NewEmitter() *emitter {
	return &emitter{}
}

func (e *emitter) Emit(msg string, labels map[string]any) error {
	return b.GetClient().Emitter.Emit(context.Background(), []byte(msg), labels)
}
