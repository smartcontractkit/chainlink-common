//go:build !unix

package beholder

import "context"

func (m *durableEmitterMetrics) recordProcessCPU(ctx context.Context) {
	_ = ctx
}
