package durableemitter

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	chipingressbatch "github.com/smartcontractkit/chainlink-common/pkg/chipingress/batch"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// globalEmitter holds the process-wide DurableEmitter instance, set by Setup.
var globalEmitter atomic.Pointer[DurableEmitter]

// SetGlobalEmitter sets the global DurableEmitter.
func SetGlobalEmitter(d *DurableEmitter) {
	globalEmitter.Store(d)
}

// GetGlobalEmitter returns the global DurableEmitter, or nil if Setup has not
// been called or DurableEmitterEnabled was false.
func GetGlobalEmitter() *DurableEmitter {
	return globalEmitter.Load()
}

// GlobalEmit emits an event via the global DurableEmitter.
// Returns a non-nil error when the global emitter has not been initialized.
func GlobalEmit(ctx context.Context, body []byte, attrKVs ...any) error {
	d := globalEmitter.Load()
	if d == nil {
		return errors.New("global DurableEmitter not initialized; call durableemitter.Setup first")
	}
	return d.Emit(ctx, body, attrKVs...)
}

// SetupConfig holds all configuration required to create and start a
// DurableEmitter including its chip ingress transport clients.
type SetupConfig struct {
	// DurableEmitterEnabled gates the entire setup. Setup returns (nil, nil)
	// when false — callers do not need an outer guard.
	DurableEmitterEnabled bool
	// Endpoint is the gRPC address for the Chip Ingress service.
	Endpoint string
	// InsecureConnection disables TLS when true.
	InsecureConnection bool
	// Auth is the per-RPC credential provider. Nil means no auth.
	Auth chipingress.HeaderProvider
	// RetransmitEnabled controls whether the retransmit and cleanup loops run.
	// Set to true for the host (chainlink node) process.
	// Set to false for LOOP plugin processes — the host's retransmit loop picks
	// up any rows inserted by plugin-side DurableEmitters from the shared DB.
	RetransmitEnabled bool

	// Batch client tuning — zero values use package defaults.
	BatchSize          int           // default: 50
	BatchInterval      time.Duration // default: 50ms
	MaxConcurrentSends int           // default: 4
	MaxPublishTimeout  time.Duration // default: 5s
	ShutdownTimeout    time.Duration // default: 30s

	// EmitterConfig overrides DefaultDurableEmitterConfig when non-nil.
	EmitterConfig *DurableEmitterConfig
	// Meter is the OpenTelemetry meter for instrumentation. Nil disables metrics.
	Meter metric.Meter
}

// Setup creates a DurableEmitter with dedicated batch and fallback chip ingress
// clients, registers it as the global emitter, and returns it unconfigured.
//
// The caller is responsible for starting and stopping the emitter:
//   - In chainlink application: append the returned emitter to srvcs so the
//     service runner manages Start/Close.
//   - In LOOP server: call emitter.Start(ctx) then emitter.Close() on shutdown.
//
// When cfg.DurableEmitterEnabled is false, Setup is a no-op and returns
// (nil, nil) — callers do not need to guard the call.
func Setup(
	store DurableEventStore,
	cfg SetupConfig,
	lggr logger.Logger,
) (*DurableEmitter, error) {
	if !cfg.DurableEmitterEnabled {
		return nil, nil
	}
	if cfg.Endpoint == "" {
		return nil, errors.New("chip ingress endpoint is required for DurableEmitter")
	}
	if store == nil {
		return nil, errors.New("durable event store is required for DurableEmitter")
	}
	if lggr == nil {
		return nil, errors.New("logger is required for DurableEmitter")
	}

	chipOpts := buildChipOpts(cfg)

	// Primary client — owned by batch.Client, closed on batch.Client.Stop().
	batchChipClient, err := chipingress.NewClient(cfg.Endpoint, chipOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch chip ingress client: %w", err)
	}

	batchClient, err := chipingressbatch.NewBatchClient(batchChipClient,
		chipingressbatch.WithBatchSize(defaultInt(cfg.BatchSize, 50)),
		chipingressbatch.WithBatchInterval(defaultDuration(cfg.BatchInterval, 50*time.Millisecond)),
		chipingressbatch.WithMaxConcurrentSends(defaultInt(cfg.MaxConcurrentSends, 4)),
		chipingressbatch.WithMaxPublishTimeout(defaultDuration(cfg.MaxPublishTimeout, 5*time.Second)),
		chipingressbatch.WithShutdownTimeout(defaultDuration(cfg.ShutdownTimeout, 30*time.Second)),
	)
	if err != nil {
		_ = batchChipClient.Close()
		return nil, fmt.Errorf("failed to create batch client: %w", err)
	}

	// Fallback client — owned by DurableEmitter, closed on DurableEmitter.Stop().
	// Used for single-event direct Publish retry when a batch delivery fails.
	fallbackClient, err := chipingress.NewClient(cfg.Endpoint, chipOpts...)
	if err != nil {
		batchClient.Stop()
		return nil, fmt.Errorf("failed to create fallback chip ingress client: %w", err)
	}

	emitterCfg := DefaultDurableEmitterConfig()
	if cfg.EmitterConfig != nil {
		emitterCfg = *cfg.EmitterConfig
	}

	emitter, err := NewDurableEmitter(store, batchClient, fallbackClient, cfg.RetransmitEnabled, emitterCfg, lggr, cfg.Meter)
	if err != nil {
		batchClient.Stop()
		_ = fallbackClient.Close()
		return nil, fmt.Errorf("failed to create durable emitter: %w", err)
	}

	SetGlobalEmitter(emitter)

	lggr.Infow("DurableEmitter created — call Start() or register with service lifecycle",
		"endpoint", cfg.Endpoint,
		"retransmitEnabled", cfg.RetransmitEnabled,
	)
	return emitter, nil
}

func buildChipOpts(cfg SetupConfig) []chipingress.Opt {
	var opts []chipingress.Opt
	if cfg.InsecureConnection {
		opts = append(opts, chipingress.WithInsecureConnection())
	} else {
		opts = append(opts, chipingress.WithTLS())
	}
	if cfg.Auth != nil {
		opts = append(opts, chipingress.WithTokenAuth(cfg.Auth))
	}
	return opts
}

func defaultInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func defaultDuration(v, def time.Duration) time.Duration {
	if v <= 0 {
		return def
	}
	return v
}
