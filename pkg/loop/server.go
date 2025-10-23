package loop

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/otelhealth"
	"github.com/smartcontractkit/chainlink-common/pkg/services/promhealth"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
)

// NewStartedServer returns a started Server.
// The caller is responsible for calling Server.Stop().
func NewStartedServer(loggerName string) (*Server, error) {
	return NewStartedServerWithOtelViews(loggerName, nil)
}

// MustNewStartedServer returns a new started Server like NewStartedServer, but logs and exits in the event of error.
// The caller is responsible for calling Server.Stop().
func MustNewStartedServer(loggerName string) *Server {
	return MustNewStartedServerWithOtelViews(loggerName, nil)
}

// NewStartedServerWithOtelViews returns a started Server with otel views registered on the beholder client.
// The caller is responsible for calling Server.Stop().
func NewStartedServerWithOtelViews(loggerName string, otelViews []sdkmetric.View) (*Server, error) {
	s, err := newServer(loggerName, otelViews)
	if err != nil {
		return nil, err
	}
	err = s.start()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MustNewStartedServerWithOtelViews returns a new started Server like NewStartedServerWithOtelViews, but logs and exits in the event of error.
// The caller is responsible for calling Server.Stop().
func MustNewStartedServerWithOtelViews(loggerName string, otelViews []sdkmetric.View) *Server {
	s, err := newServer(loggerName, otelViews)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err)
		os.Exit(1)
	}
	err = s.start()
	if err != nil {
		s.Logger.Fatalf("Failed to start server: %s", err)
	}

	return s
}

// Server holds common plugin server fields.
type Server struct {
	EnvConfig       EnvConfig
	GRPCOpts        GRPCOpts
	Logger          logger.SugaredLogger
	db              *sqlx.DB           // optional
	dbStatsReporter *pg.StatsReporter  // optional
	DataSource      sqlutil.DataSource // optional
	promServer      *PromServer
	checker         *services.HealthChecker
	LimitsFactory   limits.Factory
	otelViews       []sdkmetric.View
}

func newServer(loggerName string, otelViews []sdkmetric.View) (*Server, error) {
	lggr, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}
	lggr = logger.Named(lggr, loggerName)
	return &Server{
		GRPCOpts:  NewGRPCOpts(nil), // default prometheus.Registerer
		Logger:    logger.Sugared(lggr),
		otelViews: otelViews,
	}, nil
}

func (s *Server) start() error {
	ctx, stopSig := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSig()
	stopAfter := context.AfterFunc(ctx, stopSig)
	defer stopAfter()

	if err := s.EnvConfig.parse(); err != nil {
		return fmt.Errorf("error getting environment configuration: %w", err)
	}

	tracingAttrs := s.EnvConfig.TracingAttributes
	if tracingAttrs == nil {
		tracingAttrs = make(map[string]string, 1)
	}
	tracingAttrs[string(semconv.ServiceInstanceIDKey)] = s.EnvConfig.AppID
	tracingConfig := TracingConfig{
		Enabled:         s.EnvConfig.TracingEnabled,
		CollectorTarget: s.EnvConfig.TracingCollectorTarget,
		SamplingRatio:   s.EnvConfig.TracingSamplingRatio,
		TLSCertPath:     s.EnvConfig.TracingTLSCertPath,
		NodeAttributes:  tracingAttrs,
		OnDialError:     func(err error) { s.Logger.Errorw("Failed to dial", "err", err) },
	}

	if s.EnvConfig.TelemetryEndpoint == "" {
		err := SetupTracing(tracingConfig)
		if err != nil {
			return fmt.Errorf("failed to setup tracing: %w", err)
		}
	} else {
		var attributes []attribute.KeyValue
		if tracingConfig.Enabled {
			attributes = tracingConfig.Attributes()
		}
		if s.EnvConfig.TelemetryLogStreamingEnabled {
			attributes = append(attributes, attribute.String(beholder.AttrKeyLogSource, "loop"))
		}
		beholderCfg := beholder.Config{
			InsecureConnection:             s.EnvConfig.TelemetryInsecureConnection,
			CACertFile:                     s.EnvConfig.TelemetryCACertFile,
			OtelExporterGRPCEndpoint:       s.EnvConfig.TelemetryEndpoint,
			ResourceAttributes:             append(attributes, s.EnvConfig.TelemetryAttributes.AsStringAttributes()...),
			TraceSampleRatio:               s.EnvConfig.TelemetryTraceSampleRatio,
			EmitterBatchProcessor:          s.EnvConfig.TelemetryEmitterBatchProcessor,
			EmitterExportTimeout:           s.EnvConfig.TelemetryEmitterExportTimeout,
			EmitterExportInterval:          s.EnvConfig.TelemetryEmitterExportInterval,
			EmitterExportMaxBatchSize:      s.EnvConfig.TelemetryEmitterExportMaxBatchSize,
			EmitterMaxQueueSize:            s.EnvConfig.TelemetryEmitterMaxQueueSize,
			LogStreamingEnabled:            s.EnvConfig.TelemetryLogStreamingEnabled,
			LogLevel:                       s.EnvConfig.TelemetryLogLevel,
			ChipIngressEmitterEnabled:      s.EnvConfig.ChipIngressEndpoint != "",
			ChipIngressEmitterGRPCEndpoint: s.EnvConfig.ChipIngressEndpoint,
			ChipIngressInsecureConnection:  s.EnvConfig.ChipIngressInsecureConnection,
		}

		// Configure beholder auth - the client will determine rotating vs static mode
		// Rotating mode: when AuthHeadersTTL is set, client creates internal lazySigner
		// Static mode: no TTL is provided it is assumed that the headers are static
		if s.EnvConfig.TelemetryAuthHeadersTTL > 0 {
			// Rotating auth mode: client will create lazySigner internally and allow keystore injection after startup
			beholderCfg.AuthPublicKeyHex = s.EnvConfig.TelemetryAuthPubKeyHex
			beholderCfg.AuthHeadersTTL = s.EnvConfig.TelemetryAuthHeadersTTL
			beholderCfg.AuthHeaders = s.EnvConfig.TelemetryAuthHeaders // initial headers
		} else {
			// Static auth mode: headers and/or public key without rotation
			beholderCfg.AuthHeaders = s.EnvConfig.TelemetryAuthHeaders
			beholderCfg.AuthPublicKeyHex = s.EnvConfig.TelemetryAuthPubKeyHex
		}

		// note: due to the OTEL specification, all histogram buckets
		// must be defined when the beholder client is created
		beholderCfg.MetricViews = append(beholderCfg.MetricViews, s.otelViews...)

		if tracingConfig.Enabled {
			if beholderCfg.AuthHeaders != nil {
				tracingConfig.AuthHeaders = beholderCfg.AuthHeaders
			}
			exporter, err := tracingConfig.NewSpanExporter()
			if err != nil {
				return fmt.Errorf("failed to setup tracing exporter: %w", err)
			}
			beholderCfg.TraceSpanExporter = exporter
		}

		beholderClient, err := beholder.NewClient(beholderCfg)
		if err != nil {
			return fmt.Errorf("failed to create beholder client: %w", err)
		}
		beholder.SetClient(beholderClient)
		beholder.SetGlobalOtelProviders()

		if beholderCfg.LogStreamingEnabled {
			otelLogger, err := NewOtelLogger(beholderClient.Logger, beholderCfg.LogLevel)
			if err != nil {
				return fmt.Errorf("failed to enable log streaming: %w", err)
			}
			s.Logger = logger.Sugared(logger.Named(otelLogger, s.Logger.Name()))
		}
	}

	s.promServer = NewPromServer(s.EnvConfig.PrometheusPort, s.Logger)
	if err := s.promServer.Start(); err != nil {
		return fmt.Errorf("error starting prometheus server: %w", err)
	}

	var healthCfg services.HealthCheckerConfig
	healthCfg = promhealth.ConfigureHooks(healthCfg)
	if bc := beholder.GetClient(); bc != nil {
		var err error
		healthCfg, err = otelhealth.ConfigureHooks(healthCfg, bc.Meter)
		if err != nil {
			return fmt.Errorf("failed to configure health checker otel hooks: %w", err)
		}
	}
	s.checker = healthCfg.New()

	if err := s.checker.Start(); err != nil {
		return fmt.Errorf("error starting health checker: %w", err)
	}

	if s.EnvConfig.DatabaseURL != nil {
		pg.SetApplicationName(s.EnvConfig.DatabaseURL.URL(), build.Program)
		dbURL := s.EnvConfig.DatabaseURL.URL().String()
		var err error
		s.db, err = pg.DBConfig{
			IdleInTxSessionTimeout: s.EnvConfig.DatabaseIdleInTxSessionTimeout,
			LockTimeout:            s.EnvConfig.DatabaseLockTimeout,
			MaxOpenConns:           s.EnvConfig.DatabaseMaxOpenConns,
			MaxIdleConns:           s.EnvConfig.DatabaseMaxIdleConns,
			EnableTracing:          s.EnvConfig.DatabaseTracingEnabled,
		}.New(ctx, dbURL, pg.DriverPostgres)
		if err != nil {
			return fmt.Errorf("error connecting to DataBase: %w", err)
		}
		s.DataSource = sqlutil.WrapDataSource(s.db, s.Logger,
			sqlutil.TimeoutHook(func() time.Duration { return s.EnvConfig.DatabaseQueryTimeout }),
			sqlutil.MonitorHook(func() bool { return s.EnvConfig.DatabaseLogSQL }))

		s.dbStatsReporter = pg.NewStatsReporter(s.db.Stats, s.Logger)
		s.dbStatsReporter.Start()
	}

	s.LimitsFactory.Logger = s.Logger.Named("LimitsFactory")
	if bc := beholder.GetClient(); bc != nil {
		s.LimitsFactory.Meter = bc.Meter
		s.LimitsFactory.Settings = cresettings.DefaultGetter
	}

	return nil
}

// MustRegister registers the HealthReporter with services.HealthChecker, or exits upon failure.
func (s *Server) MustRegister(c services.HealthReporter) {
	if err := s.Register(c); err != nil {
		s.Logger.Fatalf("Failed to register %s with health checker: %v", c.Name(), err)
	}
}

func (s *Server) Register(c services.HealthReporter) error { return s.checker.Register(c) }

// Stop closes resources and flushes logs.
func (s *Server) Stop() {
	if s.dbStatsReporter != nil {
		s.dbStatsReporter.Stop()
	}
	if s.db != nil {
		s.Logger.ErrorIfFn(s.db.Close, "Failed to close database connection")
	}
	s.Logger.ErrorIfFn(s.checker.Close, "Failed to close health checker")
	s.Logger.ErrorIfFn(s.promServer.Close, "Failed to close prometheus server")
	if err := s.Logger.Sync(); err != nil {
		fmt.Println("Failed to sync logger:", err)
	}
}
