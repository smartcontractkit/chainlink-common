//go:build unix

package beholder

import (
	"context"
	"syscall"
)

func (m *durableEmitterMetrics) recordProcessCPU(ctx context.Context) {
	if m == nil {
		return
	}
	var r syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &r); err != nil {
		return
	}
	u := float64(r.Utime.Sec) + float64(r.Utime.Usec)/1e6
	s := float64(r.Stime.Sec) + float64(r.Stime.Usec)/1e6
	m.procCPUUser.Record(ctx, u)
	m.procCPUSys.Record(ctx, s)
}
