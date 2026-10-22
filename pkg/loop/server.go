package loop

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jmoiron/sqlx"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/config/build"
	"github.com/smartcontractkit/chainlink-common/pkg/durableemitter"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/pyroscope"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/telemetry"
	"github.com/smartcontractkit/chainlink-common/pkg/resourcemanager"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/services/otelhealth"
	"github.com/smartcontractkit/chainlink-common/pkg/services/promhealth"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil/pg"
)

// NewStartedServer returns a started Server.
// The caller is responsible for calling Server.Stop().
func NewStartedServer(loggerName string, opts ...ServerOpt) (*Server, error) {
	s, err := newServer(loggerName)
	if err != nil {
		return nil, err
	}
	err = s.start(opts...)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// MustNewStartedServer returns a new started Server like NewStartedServer, but logs and exits in the event of error.
// The caller is responsible for calling Server.Stop().
func MustNewStartedServer(loggerName string, opts ...ServerOpt) *Server {
	s, err := newServer(loggerName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err)
		os.Exit(1)
	}
	err = s.start(opts...)
	if err != nil {
		s.Logger.Fatalf("Failed to start server: %s", err)
	}

	return s
}

// Deprecated: use NewStartedServer(loggerName, WithOtelViews(otelViews))
//
//go:fix inline
func NewStartedServerWithOtelViews(loggerName string, otelViews []sdkmetric.View) (*Server, error) {
	return NewStartedServer(loggerName, WithOtelViews(otelViews))
}

// Deprecated: use MustNewStartedServer(loggerName, WithOtelViews(otelViews))
//
//go:fix inline
func MustNewStartedServerWithOtelViews(loggerName string, otelViews []sdkmetric.View) *Server {
	return MustNewStartedServer(loggerName, WithOtelViews(otelViews))
}

type ServerOpt func(*ServerConfig)

// ServerConfig holds additional, optional configuration.
type ServerConfig struct {
	otelViews      []sdkmetric.View
	settingsGetter settings.Getter
}

func WithOtelViews(otelViews []sdkmetric.View) ServerOpt {
	return func(cfg *ServerConfig) { cfg.otelViews = otelViews }
}

func WithSettingsGetter(settingsGetter settings.Getter) ServerOpt {
	return func(cfg *ServerConfig) { cfg.settingsGetter = settingsGetter }
}

// Server holds common plugin server fields.
type Server struct {
	Logger          logger.SugaredLogger
	EnvConfig       EnvConfig
	cfg             ServerConfig
	GRPCOpts        GRPCOpts
	db              *sqlx.DB           // optional
	dbStatsReporter *pg.StatsReporter  // optional
	DataSource      sqlutil.DataSource // optional
	webServer       *webServer
	checker         *services.HealthChecker
	LimitsFactory   limits.Factory
	profiler        *pyroscope.Profiler
	telemetry       *telemetry.Telemetry
	durableEmitter  *durableemitter.DurableEmitter
}

func newServer(loggerName string) (*Server, error) {
	lggr, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("error creating logger: %w", err)
	}
	return &Server{Logger: logger.Sugared(logger.Named(lggr, loggerName))}, nil
}

func (s *Server) start(opts ...ServerOpt) error {
	ctx, stopSig := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stopSig()
	stopAfter := context.AfterFunc(ctx, stopSig)
	defer stopAfter()

	for _, opt := range opts {
		opt(&s.cfg)
	}
	if s.cfg.settingsGetter == nil {
		s.cfg.settingsGetter = cresettings.DefaultGetter
	}

	if err := s.EnvConfig.parse(); err != nil {
		return fmt.Errorf("error getting environment configuration: %w", err)
	}

	s.GRPCOpts = GRPCOptsConfig{
		Registerer:           nil, // default prometheus.Registerer
		ServerMaxRecvMsgSize: s.EnvConfig.GRPCServerMaxRecvMsgSize,
	}.New(s.Logger)

	if err := s.startTelemetry(ctx); err != nil {
		return err
	}

	var err error
	s.profiler, err = pyroscope.Start(s.Logger, pyroscope.Config{
		AuthToken:                 s.EnvConfig.PyroscopeAuthToken,
		ServerAddress:             s.EnvConfig.PyroscopeServerAddress,
		Environment:               s.EnvConfig.PyroscopeEnvironment,
		LinkTracesToProfiles:      s.EnvConfig.PyroscopeLinkTracesToProfiles,
		PPROFBlockProfileRate:     s.EnvConfig.PyroscopePPROFBlockProfileRate,
		PPROFMutexProfileFraction: s.EnvConfig.PyroscopePPROFMutexProfileFraction,
	}, s.EnvConfig.TracingEnabled)
	if err != nil {
		return err
	}

	s.webServer = WebServerOpts{}.New(s.Logger, s.EnvConfig.PrometheusPort)
	if err := s.webServer.Start(ctx); err != nil {
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
	var meter otelmetric.Meter
	var tracerProvider oteltrace.TracerProvider
	if bc := beholder.GetClient(); bc != nil {
		meter = bc.Meter
		tracerProvider = bc.TracerProvider
		s.LimitsFactory.Meter = bc.Meter
		s.LimitsFactory.Settings = s.cfg.settingsGetter
	}

	if s.EnvConfig.ChipIngressDurableEmitterEnabled && s.EnvConfig.ChipIngressEndpoint != "" {
		if s.DataSource == nil {
			return errors.New("data source required when durable emitter is enabled")
		}

		// Rotating auth: signer is injected later via durableemitter.SetGlobalSigner when the host
		// provides the CSA keystore (see relayer and standard capabilities startup).
		emitterCfg := durableemitter.DefaultConfig()
		emitterCfg.Metrics = &durableemitter.DurableEmitterMetricsConfig{}
		durableCfg := durableemitter.SetupConfig{
			Endpoint:           s.EnvConfig.ChipIngressEndpoint,
			InsecureConnection: s.EnvConfig.ChipIngressInsecureConnection,
			Auth: durableemitter.AuthConfig{
				AuthHeaders:      s.EnvConfig.TelemetryAuthHeaders,
				AuthHeadersTTL:   s.EnvConfig.TelemetryAuthHeadersTTL,
				AuthPublicKeyHex: s.EnvConfig.TelemetryAuthPubKeyHex,
			},
			RetransmitEnabled: false, // LOOP plugins do not run the retransmit loop; the host process handles it.
			EmitterConfig:     &emitterCfg,
			Meter:             meter,
			TracerProvider:    tracerProvider,
		}
		store := durableemitter.NewPgDurableEventStore(s.DataSource)
		var err error
		s.durableEmitter, err = durableemitter.Setup(store, durableCfg, s.Logger)
		if err != nil {
			return fmt.Errorf("failed to set up durable emitter: %w", err)
		}
		if err = s.durableEmitter.Start(ctx); err != nil {
			return fmt.Errorf("failed to start durable emitter: %w", err)
		}
	}

	return nil
}

// MeteringConfig returns the resourcemanager.Config for this server's EnvConfig,
// injecting the server's own durable emitter (if durable emission is configured)
// rather than reaching for durableemitter's process-global. Opt-in: only LOOPs
// that want metering (e.g. capability producers) need call this; LOOPs that
// never will (e.g. csa_keystore, relayer) can ignore it entirely.
func (s *Server) MeteringConfig() resourcemanager.Config {
	var emitter resourcemanager.Emitter
	if s.durableEmitter != nil {
		emitter = s.durableEmitter
	}
	return s.EnvConfig.MeteringConfig(emitter)
}

// MustRegister registers the HealthReporter with services.HealthChecker, or exits upon failure.
func (s *Server) MustRegister(c services.HealthReporter) {
	if err := s.Register(c); err != nil {
		s.Logger.Fatalf("Failed to register %s with health checker: %v", c.Name(), err)
	}
}

func (s *Server) Register(c services.HealthReporter) error { return s.checker.Register(c) }

func (s *Server) startTelemetry(ctx context.Context) error {
	t, err := telemetry.NewTelemetry(s.Logger, s.EnvConfig.TelemetryConfig, s.EnvConfig.AppID, s.cfg.otelViews)
	if err != nil {
		return err
	}
	if err := t.Start(ctx); err != nil {
		return fmt.Errorf("failed to start telemetry: %w", err)
	}
	s.telemetry = t

	if level, ok := t.LogStreamingLevel(); ok {
		otelLogger, err := NewOtelLogger(t.Client().Logger, level)
		if err != nil {
			return fmt.Errorf("failed to enable log streaming: %w", err)
		}
		s.Logger = logger.Sugared(logger.Named(otelLogger, s.Logger.Name()))
	}
	return nil
}

// Stop closes resources and flushes logs.
func (s *Server) Stop() {
	if s.telemetry != nil {
		s.Logger.ErrorIfFn(s.telemetry.Close, "Failed to close telemetry")
	}
	if s.durableEmitter != nil {
		s.Logger.ErrorIfFn(s.durableEmitter.Close, "Failed to close durable emitter")
	}
	if s.dbStatsReporter != nil {
		s.dbStatsReporter.Stop()
	}
	if s.db != nil {
		s.Logger.ErrorIfFn(s.db.Close, "Failed to close database connection")
	}
	s.Logger.ErrorIfFn(s.checker.Close, "Failed to close health checker")
	s.Logger.ErrorIfFn(s.webServer.Close, "Failed to close web server")
	if s.profiler != nil {
		s.Logger.ErrorIfFn(s.profiler.Stop, "Failed to stop pyroscope profiler")
	}
	if err := s.Logger.Sync(); err != nil {
		fmt.Println("Failed to sync logger:", err)
	}
}
