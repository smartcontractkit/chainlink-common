package monitoring

import (
	"go.opentelemetry.io/otel/attribute"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// MonitoringContext carries per-capability monitoring inputs provided by the
// capability at runtime via MonitoringContext() on the generated ClientCapability
// interface. Action OTel metrics are owned by the generated server wrapper.
type MonitoringContext struct {
	Logger logger.Logger
	// MetricsAttributes returns capability-scoped low-cardinality OTel labels such as
	// chain_id, network_name, and capability_id. Generated server code adds method
	// and per-request labels from RequestMetadata.
	MetricsAttributes func() []attribute.KeyValue
}
