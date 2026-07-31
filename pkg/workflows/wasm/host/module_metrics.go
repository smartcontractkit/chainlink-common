package host

import (
	"context"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// moduleMetrics records host-side metrics for wasm module execution.
type moduleMetrics interface {
	IncHostFnPanicRecovered()
}

type moduleMetricsImpl struct {
	hostFnPanicRecoveredCount metric.Int64Counter
}

var _ moduleMetrics = &moduleMetricsImpl{}

func newModuleMetrics() (moduleMetrics, error) {
	hostFnPanicRecoveredCount, err := beholder.GetMeter().Int64Counter("platform_wasm_host_panic_recovered")
	if err != nil {
		return nil, err
	}

	return &moduleMetricsImpl{
		hostFnPanicRecoveredCount: hostFnPanicRecoveredCount,
	}, nil
}

// IncHostFnPanicRecovered records a panic that occurred in a wasm host
// function (e.g. call_capability, get_secrets, log, ...) and was recovered
// before it could propagate out of module.Execute.
func (m *moduleMetricsImpl) IncHostFnPanicRecovered() {
	m.hostFnPanicRecoveredCount.Add(context.Background(), 1)
}
