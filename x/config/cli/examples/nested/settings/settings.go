package settings

//go:generate go run ./gen

type Logging struct {
	// LogLevel is the minimum level to log.
	LogLevel string
}

type ServerConfig struct {
	// Host is the address to listen on.
	Host string
}

type MetricsConfig struct {
	// Endpoint is where to push metrics.
	Endpoint string
}

type Config struct {
	Logging

	// Server is the server configuration.
	Server ServerConfig

	// Metrics is the optional metrics configuration.
	Metrics *MetricsConfig
}
