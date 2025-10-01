package beholder

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/timeutil"
)

// Heartbeat represents a periodic heartbeat service that emits metrics and logs
type Heartbeat struct {
	services.Service
	eng *services.Engine

	Beat        time.Duration
	Emitter     Emitter
	Meter       metric.Meter
	Logger      logger.Logger
	Tracer      trace.Tracer
	AppID       string
	ServiceName string
	Version     string
	Commit      string
	Labels      map[string]string
}

// NewHeartbeat creates a new heartbeat service with custom configuration
func NewHeartbeat(beat time.Duration, lggr logger.Logger, opts ...HeartbeatOpt) *Heartbeat {
	// Setup default emitter, meter, and tracer
	noopClient := NewNoopClient()

	// Create heartbeat with defaults
	h := &Heartbeat{
		Beat:        beat,
		Logger:      lggr,
		Emitter:     noopClient.Emitter,
		Meter:       noopClient.Meter,
		Tracer:      noopClient.Tracer,
		AppID:       "chainlink",          // Default app ID
		ServiceName: build.Program,        // Default service name
		Version:     build.Version,        // Use build version
		Commit:      build.ChecksumPrefix, // Use build commit
		Labels:      make(map[string]string),
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	// Build labels from current values
	h.Labels = map[string]string{
		"service": h.ServiceName,
		"version": h.Version,
		"commit":  h.Commit,
	}
	if h.AppID != "" {
		h.Labels["app_id"] = h.AppID
	}

	// Create service engine
	h.Service, h.eng = services.Config{
		Name:  "BeholderHeartbeat",
		Start: h.start,
	}.NewServiceEngine(lggr)

	return h
}

// HeartbeatOpt is a functional option for configuring the heartbeat
type HeartbeatOpt func(*Heartbeat)

// WithEmitter sets a custom message emitter for the heartbeat
func WithEmitter(emitter Emitter) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Emitter = emitter
	}
}

// WithMeter sets a custom meter for the heartbeat
func WithMeter(meter metric.Meter) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Meter = meter
	}
}

// WithTracer sets a custom tracer for the heartbeat
func WithTracer(tracer trace.Tracer) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Tracer = tracer
	}
}

// WithAppID sets a custom app ID for the heartbeat
func WithAppID(appID string) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.AppID = appID
		if appID != "" {
			h.Labels["app_id"] = appID
		} else {
			delete(h.Labels, "app_id")
		}
	}
}

// WithServiceName sets a custom service name for the heartbeat
func WithServiceName(serviceName string) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.ServiceName = serviceName
		h.Labels["service"] = serviceName
	}
}

// WithVersion sets a custom version for the heartbeat
func WithVersion(version string) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Version = version
		h.Labels["version"] = version
	}
}

// WithCommit sets a custom commit for the heartbeat
func WithCommit(commit string) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Commit = commit
		h.Labels["commit"] = commit
	}
}

// WithBeatInterval sets a custom beat interval for the heartbeat
func WithBeatInterval(beat time.Duration) HeartbeatOpt {
	return func(h *Heartbeat) {
		h.Beat = beat
	}
}

// start initializes and starts the heartbeat service
func (h *Heartbeat) start(ctx context.Context) error {
	// Create heartbeat metrics
	heartbeatGauge, err := h.Meter.Int64Gauge("beholder_heartbeat")
	if err != nil {
		return fmt.Errorf("failed to create heartbeat status gauge: %w", err)
	}

	heartbeatCount, err := h.Meter.Int64Counter("beholder_heartbeat_count")
	if err != nil {
		return fmt.Errorf("failed to create heartbeat counter: %w", err)
	}

	// Define the heartbeat function
	beatFn := func(ctx context.Context) {
		start := time.Now()

		// Create a trace span for the heartbeat
		ctx, span := h.Tracer.Start(ctx, "beholder_heartbeat", trace.WithAttributes(
			attribute.String("service", h.ServiceName),
			attribute.String("app_id", h.AppID),
			attribute.String("version", h.Version),
			attribute.String("commit", h.Commit),
		))
		defer span.End()

		// Record heartbeat metrics
		heartbeatGauge.Record(ctx, 1)
		heartbeatCount.Add(ctx, 1)

		// Emit heartbeat message

		payload := &pb.BaseMessage{
			Msg:    "beholder heartbeat",
			Labels: h.Labels,
		}
		payloadBytes, err := proto.Marshal(payload)
		if err != nil {
			// log error
			h.Logger.Errorw("heartbeat marshal protobuf failed", "err", err)
		}

		err = h.Emitter.Emit(ctx, payloadBytes,
			AttrKeyDataSchema, "/beholder-base-message/versions/1", // required
			AttrKeyDomain, "platform", // required
			AttrKeyEntity, "BaseMessage", // required
			"service", h.ServiceName,
			"app_id", h.AppID,
			"version", h.Version,
			"commit", h.Commit,
			"timestamp", start.Unix(),
		)

		if err != nil {
			h.Logger.Errorw("heartbeat emit failed", "err", err)
		}

		// Log heartbeat
		h.Logger.Debugw("beholder heartbeat emitted",
			"service", h.ServiceName,
			"app_id", h.AppID,
			"version", h.Version,
			"commit", h.Commit,
			"timestamp", start.Unix(),
		)
	}

	// Start the heartbeat ticker
	// Execute immediately first, then continue with regular intervals
	h.eng.Go(func(ctx context.Context) {
		beatFn(ctx)
	})
	h.eng.GoTick(timeutil.NewTicker(func() time.Duration { return h.Beat }), beatFn)

	h.Logger.Infow("beholder heartbeat service started",
		"service", h.ServiceName,
		"beat_interval", h.Beat,
	)

	return nil
}
