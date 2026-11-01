package pyroscope

import (
	"fmt"
	"os"
	"strconv"
)

const (
	envPyroscopeAuthToken                 = "CL_PYROSCOPE_AUTH_TOKEN"
	envPyroscopeServerAddress             = "CL_PYROSCOPE_SERVER_ADDRESS"
	envPyroscopeEnvironment               = "CL_PYROSCOPE_ENVIRONMENT"
	envPyroscopeLinkTracesToProfiles      = "CL_PYROSCOPE_LINK_TRACES_TO_PROFILES"
	envPyroscopePPROFBlockProfileRate     = "CL_PYROSCOPE_PPROF_BLOCK_PROFILE_RATE"
	envPyroscopePPROFMutexProfileFraction = "CL_PYROSCOPE_PPROF_MUTEX_PROFILE_FRACTION"
)

// PyroscopeConfig is the profiling part of loop.EnvConfig.
type PyroscopeConfig struct {
	PyroscopeAuthToken                 string
	PyroscopeServerAddress             string
	PyroscopeEnvironment               string
	PyroscopeLinkTracesToProfiles      bool
	PyroscopePPROFBlockProfileRate     int
	PyroscopePPROFMutexProfileFraction int
}

// AsCmdEnv returns c as environment variable key/value pairs for an exec.Cmd.
func (c *PyroscopeConfig) AsCmdEnv() []string {
	return []string{
		envPyroscopeAuthToken + "=" + c.PyroscopeAuthToken,
		envPyroscopeServerAddress + "=" + c.PyroscopeServerAddress,
		envPyroscopeEnvironment + "=" + c.PyroscopeEnvironment,
		envPyroscopeLinkTracesToProfiles + "=" + strconv.FormatBool(c.PyroscopeLinkTracesToProfiles),
		envPyroscopePPROFBlockProfileRate + "=" + strconv.Itoa(c.PyroscopePPROFBlockProfileRate),
		envPyroscopePPROFMutexProfileFraction + "=" + strconv.Itoa(c.PyroscopePPROFMutexProfileFraction),
	}
}

// Parse reads c from the environment variables AsCmdEnv writes.
func (c *PyroscopeConfig) Parse() error {
	var err error
	c.PyroscopeAuthToken = os.Getenv(envPyroscopeAuthToken)
	c.PyroscopeServerAddress = os.Getenv(envPyroscopeServerAddress)
	c.PyroscopeEnvironment = os.Getenv(envPyroscopeEnvironment)
	c.PyroscopeLinkTracesToProfiles, err = getBool(envPyroscopeLinkTracesToProfiles)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envPyroscopeLinkTracesToProfiles, err)
	}
	c.PyroscopePPROFBlockProfileRate, err = getInt(envPyroscopePPROFBlockProfileRate)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envPyroscopePPROFBlockProfileRate, err)
	}
	c.PyroscopePPROFMutexProfileFraction, err = getInt(envPyroscopePPROFMutexProfileFraction)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", envPyroscopePPROFMutexProfileFraction, err)
	}
	return nil
}

func getBool(envKey string) (bool, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}

func getInt(envKey string) (int, error) {
	s := os.Getenv(envKey)
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
