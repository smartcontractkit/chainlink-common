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

var (
	ErrNotInitialized = errors.New("durable emitter not initialized")
	ErrEmitFailed     = errors.New("durable emitter emit failed")
)

// SetGlobalEmitter sets the global DurableEmitter.
func SetGlobalEmitter(d *DurableEmitter) {
	globalEmitter.Store(d)
}

// GetGlobalEmitter returns the global DurableEmitter, or nil if Setup has not been called.
func GetGlobalEmitter() *DurableEmitter {
	return globalEmitter.Load()
}

// GlobalEmit emits an event via the global DurableEmitter.
//
// This function is non-blocking: it enqueues the event to the emitter's
// async channel and returns immediately. A background goroutine persists
// and publishes the event. If the async channel is full (backpressure),
// the event is dropped (fail-open) — the beholder emitter already delivered
// it in real-time, and the durable emitter's value is persistence for
// retransmit, not inline delivery.
func GlobalEmit(ctx context.Context, body []byte, attrKVs ...any) error {
	d := globalEmitter.Load()
	if d == nil {
		return ErrNotInitialized
	}
	select {
	case d.asyncEmitCh <- &asyncEmitRequest{body: body, attrKVs: attrKVs}:
		return nil
	default:
		// Channel full — fail open. The event was already emitted via the
		// beholder emitter; dropping the durable copy is preferable to
		// blocking the workflow execution path.
		return nil
	}
}

// SetupConfig holds all configuration required to create and start a
// DurableEmitter including its chip ingress transport clients.
type SetupConfig struct {
	// Endpoint is the gRPC address for the Chip Ingress service.
	Endpoint string
	// InsecureConnection disables TLS when true.
	InsecureConnection bool
	// Auth configures chip ingress credentials. AuthKeySigner may be nil at init
	// for LOOP plugins; call SetGlobalSigner after the CSA keystore is available.
	Auth AuthConfig
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
	// MessageBufferSize is the capacity of the batch client's producer→batcher
	// channel. QueueMessage drops events (non-blocking send) when this is full,
	// which happens when emit throughput outpaces the batcher.
	MessageBufferSize int // default: 10000

	// EmitterConfig overrides DefaultConfig when non-nil.
	EmitterConfig *Config
	// Meter is the OpenTelemetry meter for instrumentation. Nil disables metrics.
	Meter metric.Meter
}

// Setup creates a DurableEmitter with a dedicated batch chip ingress client,
// registers it as the global emitter, and returns it unconfigured.
func Setup(
	store DurableEventStore,
	cfg SetupConfig,
	lggr logger.Logger,
) (*DurableEmitter, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("chip ingress endpoint is required for DurableEmitter")
	}
	if store == nil {
		return nil, errors.New("durable event store is required for DurableEmitter")
	}
	if lggr == nil {
		return nil, errors.New("logger is required for DurableEmitter")
	}

	auth, err := NewAuthHeaderProvider(cfg.Auth)
	if err != nil {
		return nil, fmt.Errorf("failed to build chip ingress auth: %w", err)
	}

	chipOpts := buildChipOpts(cfg, auth)

	// Primary client — owned by batch.Client, closed on batch.Client.Stop().
	batchChipClient, err := chipingress.NewClient(cfg.Endpoint, chipOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch chip ingress client: %w", err)
	}

	batchClient, err := chipingressbatch.NewBatchClient(batchChipClient,
		chipingressbatch.WithBatchSize(defaultInt(cfg.BatchSize, 50)),
		chipingressbatch.WithBatchInterval(defaultDuration(cfg.BatchInterval, 50*time.Millisecond)),
		chipingressbatch.WithMaxConcurrentSends(defaultInt(cfg.MaxConcurrentSends, 4)),
		chipingressbatch.WithMessageBuffer(defaultInt(cfg.MessageBufferSize, 10_000)),
		chipingressbatch.WithMaxPublishTimeout(defaultDuration(cfg.MaxPublishTimeout, 5*time.Second)),
		chipingressbatch.WithShutdownTimeout(defaultDuration(cfg.ShutdownTimeout, 30*time.Second)),
		chipingressbatch.WithClientName(chipingressbatch.ClientNameDurableEmitter),
	)
	if err != nil {
		_ = batchChipClient.Close()
		return nil, fmt.Errorf("failed to create batch client: %w", err)
	}

	emitterCfg := DefaultConfig()
	if cfg.EmitterConfig != nil {
		emitterCfg = *cfg.EmitterConfig
	}

	emitter, err := NewDurableEmitter(store, batchClient, cfg.RetransmitEnabled, emitterCfg, lggr, cfg.Meter)
	if err != nil {
		batchClient.Stop()
		return nil, fmt.Errorf("failed to create durable emitter: %w", err)
	}

	SetGlobalEmitter(emitter)

	lggr.Infow("DurableEmitter created — call Start() or register with service lifecycle",
		"endpoint", cfg.Endpoint,
		"retransmitEnabled", cfg.RetransmitEnabled,
	)
	return emitter, nil
}

func buildChipOpts(cfg SetupConfig, auth chipingress.HeaderProvider) []chipingress.Opt {
	var opts []chipingress.Opt
	if cfg.InsecureConnection {
		opts = append(opts, chipingress.WithInsecureConnection())
	} else {
		opts = append(opts, chipingress.WithTLS())
	}
	if auth != nil {
		opts = append(opts, chipingress.WithTokenAuth(auth))
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
