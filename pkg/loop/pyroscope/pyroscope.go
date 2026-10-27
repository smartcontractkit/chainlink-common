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

// Start returns a nil Profiler if no PyroscopeServerAddress is configured. appName defaults to the
// executable's base name.
func Start(lggr logger.Logger, cfg PyroscopeConfig, appName string, tracingEnabled bool) (*Profiler, error) {
	if cfg.PyroscopeServerAddress == "" {
		return nil, nil
	}

	runtime.SetBlockProfileRate(cfg.PyroscopePPROFBlockProfileRate)
	runtime.SetMutexProfileFraction(cfg.PyroscopePPROFMutexProfileFraction)

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

	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: orExecutableName(lggr, appName),
		ServerAddress:   cfg.PyroscopeServerAddress,
		AuthToken:       cfg.PyroscopeAuthToken,

		Tags: map[string]string{
			"module":      module,
			"SHA":         sha,
			"Version":     ver,
			"go":          goVer,
			"Environment": cfg.PyroscopeEnvironment,
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
	if tracingEnabled && cfg.PyroscopeLinkTracesToProfiles {
		otel.SetTracerProvider(otelpyroscope.NewTracerProvider(otel.GetTracerProvider()))
	}
	return profiler, nil
}

func orExecutableName(lggr logger.Logger, appName string) string {
	if appName != "" {
		return appName
	}
	exe, err := os.Executable()
	if err != nil {
		lggr.Warnf("Failed to get executable name: %v", err)
		return "unknown"
	}
	return filepath.Base(exe)
}
