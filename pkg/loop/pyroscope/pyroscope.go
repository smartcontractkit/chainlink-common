package pyroscope

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/otel"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type Profiler = pyroscope.Profiler

type Config struct {
	AuthToken                 string
	ServerAddress             string
	Environment               string
	LinkTracesToProfiles      bool
	PPROFBlockProfileRate     int
	PPROFMutexProfileFraction int
}

// Start returns a nil Profiler if no ServerAddress is configured.
func Start(lggr logger.Logger, cfg Config, tracingEnabled bool) (*Profiler, error) {
	if cfg.ServerAddress == "" {
		return nil, nil
	}

	runtime.SetBlockProfileRate(cfg.PPROFBlockProfileRate)
	runtime.SetMutexProfileFraction(cfg.PPROFMutexProfileFraction)

	hostname, _ := os.Hostname()
	var ver, sha, goVer, module string
	if bi, ok := debug.ReadBuildInfo(); ok {
		ver = bi.Main.Version
		sha = bi.Main.Sum
		if len(sha) > 7 {
			sha = sha[:7]
		}
		goVer = bi.GoVersion
		module = bi.Main.Path
	}

	appName, err := os.Executable()
	if err != nil {
		lggr.Warnf("Failed to get executable name: %v", err)
		appName = "unknown"
	} else {
		appName = filepath.Base(appName)
	}

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: appName,
		ServerAddress:   cfg.ServerAddress,
		AuthToken:       cfg.AuthToken,

		Tags: map[string]string{
			"module":      module,
			"SHA":         sha,
			"Version":     ver,
			"go":          goVer,
			"Environment": cfg.Environment,
			"hostname":    hostname,
		},
		ProfileTypes: []pyroscope.ProfileType{
			// these profile types are enabled by default:
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,

			// these profile types are optional:
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start pyroscope profiler: %w", err)
	}
	if tracingEnabled && cfg.LinkTracesToProfiles {
		otel.SetTracerProvider(otelpyroscope.NewTracerProvider(otel.GetTracerProvider()))
	}
	return profiler, nil
}
