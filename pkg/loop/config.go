package loop

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/go-plugin"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/pyroscope"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/telemetry"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

const (
	envAppID = "CL_APP_ID"

	envDatabaseURL                          = "CL_DATABASE_URL"
	envDatabaseIdleInTxSessionTimeout       = "CL_DATABASE_IDLE_IN_TX_SESSION_TIMEOUT"
	envDatabaseLockTimeout                  = "CL_DATABASE_LOCK_TIMEOUT"
	envDatabaseQueryTimeout                 = "CL_DATABASE_QUERY_TIMEOUT"
	envDatabaseListenerFallbackPollInterval = "CL_DATABASE_LISTNER_FALLBACK_POLL_INTERVAL"
	envDatabaseLogSQL                       = "CL_DATABASE_LOG_SQL"
	envDatabaseMaxOpenConns                 = "CL_DATABASE_MAX_OPEN_CONNS"
	envDatabaseMaxIdleConns                 = "CL_DATABASE_MAX_IDLE_CONNS"
	envDatabaseTracingEnabled               = "CL_DATABASE_TRACING_ENABLED"

	envFeatureLogPoller = "CL_FEATURE_LOG_POLLER"

	envGRPCServerMaxRecvMsgSize = "CL_GRPC_SERVER_MAX_RECV_MSG_SIZE"

	envMercuryCacheLatestReportDeadline = "CL_MERCURY_CACHE_LATEST_REPORT_DEADLINE"
	envMercuryCacheLatestReportTTL      = "CL_MERCURY_CACHE_LATEST_REPORT_TTL"
	envMercuryCacheMaxStaleAge          = "CL_MERCURY_CACHE_MAX_STALE_AGE"

	envMercuryTransmitterProtocol             = "CL_MERCURY_TRANSMITTER_PROTOCOL"
	envMercuryTransmitterTransmitQueueMaxSize = "CL_MERCURY_TRANSMITTER_TRANSMIT__QUEUE_MAX_SIZE"
	envMercuryTransmitterTransmitTimeout      = "CL_MERCURY_TRANSMITTER_TRANSMIT_TIMEOUT"
	envMercuryTransmitterTransmitConcurrency  = "CL_MERCURY_TRANSMITTER_TRANSMIT_CONCURRENCY"
	envMercuryTransmitterReaperFrequency      = "CL_MERCURY_TRANSMITTER_REAPER_FREQUENCY"
	envMercuryTransmitterReaperMaxAge         = "CL_MERCURY_TRANSMITTER_REAPER_MAX_AGE"
	envMercuryVerboseLogging                  = "CL_MERCURY_VERBOSE_LOGGING"

	envPromPort = "CL_PROMETHEUS_PORT"

	envMeterRecordsEnabled   = "CL_METER_RECORDS_ENABLED"
	envMeterSnapshotsEnabled = "CL_METER_SNAPSHOTS_ENABLED"
	envMeterProduct          = "CL_METER_PRODUCT"
	envMeterTenant           = "CL_METER_TENANT"
	envMeterNumericTenantID  = "CL_METER_NUMERIC_TENANT_ID"
	envMeterEnvironment      = "CL_METER_ENVIRONMENT"
	envMeterZone             = "CL_METER_ZONE"
	envMeterNodeID           = "CL_METER_NODE_ID"
	envMeterFaultMode        = "CL_METER_FAULT_MODE"
	envMeterFaultRate        = "CL_METER_FAULT_RATE"
	envMeterFaultDelta       = "CL_METER_FAULT_DELTA"
	envMeterFaultNodeRate    = "CL_METER_FAULT_NODE_RATE"

	envCRESettings        = cresettings.EnvNameSettings
	envCRESettingsDefault = cresettings.EnvNameSettingsDefault
)

// EnvConfig is the configuration between the application and the LOOP executable. The values
// are fully resolved and static and passed via the environment.
type EnvConfig struct {
	AppID string

	telemetry.TelemetryConfig

	CRESettings        string
	CRESettingsDefault string

	DatabaseURL                          *config.SecretURL
	DatabaseIdleInTxSessionTimeout       time.Duration
	DatabaseLockTimeout                  time.Duration
	DatabaseQueryTimeout                 time.Duration
	DatabaseListenerFallbackPollInterval time.Duration
	DatabaseLogSQL                       bool
	DatabaseMaxOpenConns                 int
	DatabaseMaxIdleConns                 int
	DatabaseTracingEnabled               bool

	FeatureLogPoller bool

	GRPCServerMaxRecvMsgSize int

	MercuryCacheLatestReportDeadline time.Duration
	MercuryCacheLatestReportTTL      time.Duration
	MercuryCacheMaxStaleAge          time.Duration

	MercuryTransmitterProtocol             string
	MercuryTransmitterTransmitQueueMaxSize uint32
	MercuryTransmitterTransmitTimeout      time.Duration
	MercuryTransmitterTransmitConcurrency  uint32
	MercuryTransmitterReaperFrequency      time.Duration
	MercuryTransmitterReaperMaxAge         time.Duration
	MercuryVerboseLogging                  bool

	PrometheusPort int // also serves pprof routes

	pyroscope.PyroscopeConfig

	MeterRecordsEnabled   bool
	MeterSnapshotsEnabled bool

	// MeterProduct / MeterTenant / MeterNumericTenantID / MeterEnvironment /
	// MeterZone / MeterNodeID are
	// the static deployment+node identity dimensions used as coarse
	// metering/billing rollup dimensions. They are resolved once from node
	// config by the host and delivered to every LOOP plugin over the env, the
	// same channel as the meter-record toggles above (rather than the
	// standard-capabilities boundary). Any may be empty if the host did not
	// provide it.
	//
	// MeterNodeID is the node's logical name (e.g. "clp-cre-wf-zone-a-1"),
	// not the CSA public key; the CSA key rides emitted events separately as the
	// node_csa_key attribute.
	MeterProduct         string
	MeterTenant          string
	MeterNumericTenantID string
	MeterEnvironment     string
	MeterZone            string
	MeterNodeID          string
	// MeterFault* configure usage fault injection (resourcemanager.UsageFaultInjection).
	MeterFaultMode     string
	MeterFaultRate     float64
	MeterFaultDelta    int64
	MeterFaultNodeRate float64
}

// AsCmdEnv returns a slice of environment variable key/value pairs for an exec.Cmd.
func (e *EnvConfig) AsCmdEnv() (env []string) {
	add := func(k, v string) {
		env = append(env, k+"="+v)
	}

	add(envAppID, e.AppID)

	if e.DatabaseURL != nil { // optional
		add(envDatabaseURL, e.DatabaseURL.URL().String())
		add(envDatabaseIdleInTxSessionTimeout, e.DatabaseIdleInTxSessionTimeout.String())
		add(envDatabaseLockTimeout, e.DatabaseLockTimeout.String())
		add(envDatabaseQueryTimeout, e.DatabaseQueryTimeout.String())
		add(envDatabaseListenerFallbackPollInterval, e.DatabaseListenerFallbackPollInterval.String())
		add(envDatabaseLogSQL, strconv.FormatBool(e.DatabaseLogSQL))
		add(envDatabaseMaxOpenConns, strconv.Itoa(e.DatabaseMaxOpenConns))
		add(envDatabaseMaxIdleConns, strconv.Itoa(e.DatabaseMaxIdleConns))
		add(envDatabaseTracingEnabled, strconv.FormatBool(e.DatabaseTracingEnabled))
	}

	add(envFeatureLogPoller, strconv.FormatBool(e.FeatureLogPoller))

	add(envGRPCServerMaxRecvMsgSize, strconv.Itoa(e.GRPCServerMaxRecvMsgSize))

	add(envMercuryCacheLatestReportDeadline, e.MercuryCacheLatestReportDeadline.String())
	add(envMercuryCacheLatestReportTTL, e.MercuryCacheLatestReportTTL.String())
	add(envMercuryCacheMaxStaleAge, e.MercuryCacheMaxStaleAge.String())

	add(envMercuryTransmitterProtocol, e.MercuryTransmitterProtocol)
	add(envMercuryTransmitterTransmitQueueMaxSize, strconv.FormatUint(uint64(e.MercuryTransmitterTransmitQueueMaxSize), 10))
	add(envMercuryTransmitterTransmitTimeout, e.MercuryTransmitterTransmitTimeout.String())
	add(envMercuryTransmitterTransmitConcurrency, strconv.FormatUint(uint64(e.MercuryTransmitterTransmitConcurrency), 10))
	add(envMercuryTransmitterReaperFrequency, e.MercuryTransmitterReaperFrequency.String())
	add(envMercuryTransmitterReaperMaxAge, e.MercuryTransmitterReaperMaxAge.String())
	add(envMercuryVerboseLogging, strconv.FormatBool(e.MercuryVerboseLogging))

	add(envPromPort, strconv.Itoa(e.PrometheusPort))

	env = append(env, e.PyroscopeConfig.AsCmdEnv()...)
	env = append(env, e.TelemetryConfig.AsCmdEnv()...)

	add(envMeterRecordsEnabled, strconv.FormatBool(e.MeterRecordsEnabled))
	add(envMeterSnapshotsEnabled, strconv.FormatBool(e.MeterSnapshotsEnabled))
	add(envMeterProduct, e.MeterProduct)
	add(envMeterTenant, e.MeterTenant)
	add(envMeterNumericTenantID, e.MeterNumericTenantID)
	add(envMeterEnvironment, e.MeterEnvironment)
	add(envMeterZone, e.MeterZone)
	add(envMeterNodeID, e.MeterNodeID)
	add(envMeterFaultMode, e.MeterFaultMode)
	add(envMeterFaultRate, strconv.FormatFloat(e.MeterFaultRate, 'g', -1, 64))
	add(envMeterFaultDelta, strconv.FormatInt(e.MeterFaultDelta, 10))
	add(envMeterFaultNodeRate, strconv.FormatFloat(e.MeterFaultNodeRate, 'g', -1, 64))

	if e.CRESettings != "" {
		add(envCRESettings, e.CRESettings)
	}
	if e.CRESettingsDefault != "" {
		add(envCRESettingsDefault, e.CRESettingsDefault)
	}

	return
}

// parse deserializes environment variables
func (e *EnvConfig) parse() error {
	e.AppID = os.Getenv(envAppID)
	var err error
	e.DatabaseURL, err = getEnv(envDatabaseURL, func(s string) (*config.SecretURL, error) {
		if s == "" { // DatabaseURL is optional
			return nil, nil
		}
		u, err2 := url.Parse(s)
		if err2 != nil {
			return nil, err2
		}
		return (*config.SecretURL)(u), nil
	})
	if err != nil {
		return err
	}
	if e.DatabaseURL != nil {
		e.DatabaseIdleInTxSessionTimeout, err = getDuration(envDatabaseIdleInTxSessionTimeout)
		if err != nil {
			return err
		}
		e.DatabaseLockTimeout, err = getDuration(envDatabaseLockTimeout)
		if err != nil {
			return err
		}
		e.DatabaseQueryTimeout, err = getDuration(envDatabaseQueryTimeout)
		if err != nil {
			return err
		}
		e.DatabaseListenerFallbackPollInterval, err = getDuration(envDatabaseListenerFallbackPollInterval)
		if err != nil {
			return err
		}
		e.DatabaseLogSQL, err = getBool(envDatabaseLogSQL)
		if err != nil {
			return err
		}
		e.DatabaseMaxOpenConns, err = getInt(envDatabaseMaxOpenConns)
		if err != nil {
			return err
		}
		e.DatabaseMaxIdleConns, err = getInt(envDatabaseMaxIdleConns)
		if err != nil {
			return err
		}
		e.DatabaseTracingEnabled, err = getBool(envDatabaseTracingEnabled)
		if err != nil {
			return err
		}
	}

	e.FeatureLogPoller, err = getBool(envFeatureLogPoller)
	if err != nil {
		return err
	}

	e.GRPCServerMaxRecvMsgSize, err = getInt(envGRPCServerMaxRecvMsgSize)
	if err != nil {
		return err
	}

	e.MercuryCacheLatestReportDeadline, err = getDuration(envMercuryCacheLatestReportDeadline)
	if err != nil {
		return err
	}
	e.MercuryCacheLatestReportTTL, err = getDuration(envMercuryCacheLatestReportTTL)
	if err != nil {
		return err
	}
	e.MercuryCacheMaxStaleAge, err = getDuration(envMercuryCacheMaxStaleAge)
	if err != nil {
		return err
	}

	e.MercuryTransmitterProtocol = os.Getenv(envMercuryTransmitterProtocol)
	e.MercuryTransmitterTransmitQueueMaxSize, err = getUint32(envMercuryTransmitterTransmitQueueMaxSize)
	if err != nil {
		return err
	}
	e.MercuryTransmitterTransmitTimeout, err = getDuration(envMercuryTransmitterTransmitTimeout)
	if err != nil {
		return err
	}
	e.MercuryTransmitterTransmitConcurrency, err = getUint32(envMercuryTransmitterTransmitConcurrency)
	if err != nil {
		return err
	}
	e.MercuryTransmitterReaperFrequency, err = getDuration(envMercuryTransmitterReaperFrequency)
	if err != nil {
		return err
	}
	e.MercuryTransmitterReaperMaxAge, err = getDuration(envMercuryTransmitterReaperMaxAge)
	if err != nil {
		return err
	}
	e.MercuryVerboseLogging, err = getBool(envMercuryVerboseLogging)
	if err != nil {
		return err
	}

	promPortStr := os.Getenv(envPromPort)
	e.PrometheusPort, err = strconv.Atoi(promPortStr)
	if err != nil {
		return fmt.Errorf("failed to parse %s = %q: %w", envPromPort, promPortStr, err)
	}

	if err = e.PyroscopeConfig.Parse(); err != nil {
		return err
	}

	if err = e.TelemetryConfig.Parse(); err != nil {
		return err
	}

	e.CRESettings = os.Getenv(envCRESettings)
	e.CRESettingsDefault = os.Getenv(envCRESettingsDefault)

	e.MeterRecordsEnabled, err = getBool(envMeterRecordsEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envMeterRecordsEnabled, err)
	}
	e.MeterSnapshotsEnabled, err = getBool(envMeterSnapshotsEnabled)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envMeterSnapshotsEnabled, err)
	}

	e.MeterProduct = os.Getenv(envMeterProduct)
	e.MeterTenant = os.Getenv(envMeterTenant)
	e.MeterNumericTenantID = os.Getenv(envMeterNumericTenantID)
	e.MeterEnvironment = os.Getenv(envMeterEnvironment)
	e.MeterZone = os.Getenv(envMeterZone)
	e.MeterNodeID = os.Getenv(envMeterNodeID)
	e.MeterFaultMode = os.Getenv(envMeterFaultMode)
	if e.MeterFaultRate, err = getFloat64(envMeterFaultRate); err != nil {
		return fmt.Errorf("failed to parse %s: %w", envMeterFaultRate, err)
	}
	if v := os.Getenv(envMeterFaultDelta); v != "" {
		e.MeterFaultDelta, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", envMeterFaultDelta, err)
		}
	}
	if e.MeterFaultNodeRate, err = getFloat64(envMeterFaultNodeRate); err != nil {
		return fmt.Errorf("failed to parse %s: %w", envMeterFaultNodeRate, err)
	}

	return nil
}

// ManagedGRPCClientConfig return a Managed plugin and set grpc config values from the BrokerConfig.
// The innermost relevant BrokerConfig should be used, to include any relevant services in the logger name.
// Note: managed plugins shutdown when the parent process exits. We may want to change this behavior in the future
// to enable host process restarts without restarting the plugin. To do that we would also need
// supply the appropriate ReattachConfig to the plugin.ClientConfig.
func ManagedGRPCClientConfig(clientConfig *plugin.ClientConfig, c BrokerConfig) *plugin.ClientConfig {
	clientConfig.AllowedProtocols = []plugin.Protocol{plugin.ProtocolGRPC}
	clientConfig.GRPCDialOptions = c.DialOpts
	clientConfig.Logger = HCLogLogger(c.Logger)
	clientConfig.Managed = true
	return clientConfig
}

func getBool(envKey string) (bool, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}

func getFloat64(envKey string) (float64, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func getUint32(envKey string) (uint32, error) {
	s := os.Getenv(envKey)
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(u), nil
}

func getDuration(envKey string) (time.Duration, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

func getEnv[T any](key string, parse func(string) (T, error)) (t T, err error) {
	v := os.Getenv(key)
	t, err = parse(v)
	if err != nil {
		err = fmt.Errorf("failed to parse %s=%s: %w", key, v, err)
	}
	return
}

func getInt(envKey string) (int, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return i, nil
}
