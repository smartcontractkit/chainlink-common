package loop

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
)

// NewStartedServer returns a started Server.
// The caller is responsible for calling Server.Stop().
func NewStartedServer(loggerName string) (*Server, error) {
	s, err := newServer(loggerName)
	if err != nil {
		return nil, err
	}
	err = s.start()
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MustNewStartedServer returns a new started Server like NewStartedServer, but logs and exits in the event of error.
// The caller is responsible for calling Server.Stop().
func MustNewStartedServer(loggerName string) *Server {
	s, err := newServer(loggerName)
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
	GRPCOpts        GRPCOpts
	Logger          logger.SugaredLogger
	db              *sqlx.DB           // optional
	dbStatsReporter *pg.StatsReporter  // optional
	DataSource      sqlutil.DataSource // optional
	promServer      *PromServer
	checker         *services.HealthChecker
}

func newServer(loggerName string) (*Server, error) {
	s := &Server{
		// default prometheus.Registerer
		GRPCOpts: NewGRPCOpts(nil),
	}

	lggr, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}
	lggr = logger.Named(lggr, loggerName)
	s.Logger = logger.Sugared(lggr)
	return s, nil
}

func (s *Server) start() error {
	ctx, stopSig := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSig()
	stopAfter := context.AfterFunc(ctx, stopSig)
	defer stopAfter()

	var envCfg EnvConfig
	if err := envCfg.parse(); err != nil {
		return fmt.Errorf("error getting environment configuration: %w", err)
	}

	tracingConfig := TracingConfig{
		Enabled:         envCfg.TracingEnabled,
		CollectorTarget: envCfg.TracingCollectorTarget,
		SamplingRatio:   envCfg.TracingSamplingRatio,
		TLSCertPath:     envCfg.TracingTLSCertPath,
		NodeAttributes:  envCfg.TracingAttributes,
		OnDialError:     func(err error) { s.Logger.Errorw("Failed to dial", "err", err) },
	}

	if envCfg.TelemetryEndpoint == "" {
		err := SetupTracing(tracingConfig)
		if err != nil {
			return fmt.Errorf("failed to setup tracing: %w", err)
		}
	} else {
		var attributes []attribute.KeyValue
		if tracingConfig.Enabled {
			attributes = tracingConfig.Attributes()
		}
		beholderCfg := beholder.Config{
			InsecureConnection:             envCfg.TelemetryInsecureConnection,
			CACertFile:                     envCfg.TelemetryCACertFile,
			OtelExporterGRPCEndpoint:       envCfg.TelemetryEndpoint,
			ResourceAttributes:             append(attributes, envCfg.TelemetryAttributes.AsStringAttributes()...),
			TraceSampleRatio:               envCfg.TelemetryTraceSampleRatio,
			AuthHeaders:                    envCfg.TelemetryAuthHeaders,
			AuthPublicKeyHex:               envCfg.TelemetryAuthPubKeyHex,
			EmitterBatchProcessor:          envCfg.TelemetryEmitterBatchProcessor,
			EmitterExportTimeout:           envCfg.TelemetryEmitterExportTimeout,
			EmitterExportInterval:          envCfg.TelemetryEmitterExportInterval,
			EmitterExportMaxBatchSize:      envCfg.TelemetryEmitterExportMaxBatchSize,
			EmitterMaxQueueSize:            envCfg.TelemetryEmitterMaxQueueSize,
			ChipIngressEmitterEnabled:      envCfg.ChipIngressEndpoint != "",
			ChipIngressEmitterGRPCEndpoint: envCfg.ChipIngressEndpoint,
		}

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
	}

	s.promServer = NewPromServer(envCfg.PrometheusPort, s.Logger)
	if err := s.promServer.Start(); err != nil {
		return fmt.Errorf("error starting prometheus server: %w", err)
	}

	s.checker = services.NewChecker("", "")
	if err := s.checker.Start(); err != nil {
		return fmt.Errorf("error starting health checker: %w", err)
	}

	if envCfg.DatabaseURL != nil {
		pg.SetApplicationName(envCfg.DatabaseURL.URL(), build.Program)
		dbURL := envCfg.DatabaseURL.URL().String()
		var err error
		s.db, err = pg.DBConfig{
			IdleInTxSessionTimeout: envCfg.DatabaseIdleInTxSessionTimeout,
			LockTimeout:            envCfg.DatabaseLockTimeout,
			MaxOpenConns:           envCfg.DatabaseMaxOpenConns,
			MaxIdleConns:           envCfg.DatabaseMaxIdleConns,
		}.New(ctx, dbURL, pg.DriverPostgres)
		if err != nil {
			return fmt.Errorf("error connecting to DataBase: %w", err)
		}
		s.DataSource = sqlutil.WrapDataSource(s.db, s.Logger,
			sqlutil.TimeoutHook(func() time.Duration { return envCfg.DatabaseQueryTimeout }),
			sqlutil.MonitorHook(func() bool { return envCfg.DatabaseLogSQL }))

		s.dbStatsReporter = pg.NewStatsReporter(s.db.Stats, s.Logger)
		s.dbStatsReporter.Start()
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
