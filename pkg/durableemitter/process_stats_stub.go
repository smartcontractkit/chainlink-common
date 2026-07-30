//go:build !unix

package durableemitter

import (
	"context"
	"runtime"
)

func (m *durableEmitterMetrics) pollProcessGauges(ctx context.Context) {
	if m == nil {
		return
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.procHeapInuse.Record(ctx, int64(mem.HeapInuse))
	m.procHeapSys.Record(ctx, int64(mem.HeapSys))
}
