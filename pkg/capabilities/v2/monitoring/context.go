package monitoring

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// MonitoringContext carries the per-capability infrastructure injected by the
// generated server (--with-monitoring) into every action invocation.
// It is provided by the capability at runtime via the MonitoringContext() method
// on the generated ClientCapability interface.
type MonitoringContext struct {
	Logger    logger.Logger
	Processor beholder.ProtoProcessor
	// ExecCtx builds an ExecutionContext from request metadata and the dispatch
	// timestamp. MetaCapabilityTimestampStart is set to tsStart;
	// MetaCapabilityTimestampEmit is set to time.Now() inside the function.
	ExecCtx func(metadata capabilities.RequestMetadata, tsStart time.Time) *ExecutionContext
}
