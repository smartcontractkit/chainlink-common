//go:build unix

package durableemitter

import (
	"context"
	"runtime"
	"syscall"
)

func (m *durableEmitterMetrics) pollProcessGauges(ctx context.Context) {
	if m == nil {
		return
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	m.procHeapInuse.Record(ctx, int64(mem.HeapInuse))
	m.procHeapSys.Record(ctx, int64(mem.HeapSys))

	var rusage syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &rusage); err != nil {
		return
	}
	user := float64(rusage.Utime.Sec) + float64(rusage.Utime.Usec)/1e6
	sys := float64(rusage.Stime.Sec) + float64(rusage.Stime.Usec)/1e6
	m.procCPUUser.Record(ctx, user)
	m.procCPUSys.Record(ctx, sys)
}
