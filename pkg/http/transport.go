package http

import (
	"net/http"
	"time"
)

// TransportConfig configures HTTP connection pooling for pipeline adapters.
// Zero values are replaced with DefaultTransportConfig before being applied.
type TransportConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
}

// DefaultTransportConfig returns transport pool settings tuned for high-throughput
// bridge and pipeline HTTP adapters.
func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}
}

func (c TransportConfig) withDefaults() TransportConfig {
	defaults := DefaultTransportConfig()
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = defaults.MaxIdleConns
	}
	if c.MaxIdleConnsPerHost == 0 {
		c.MaxIdleConnsPerHost = defaults.MaxIdleConnsPerHost
	}
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = defaults.IdleConnTimeout
	}
	return c
}

func (c TransportConfig) apply(t *http.Transport) {
	cfg := c.withDefaults()
	t.MaxIdleConns = cfg.MaxIdleConns
	t.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	t.IdleConnTimeout = cfg.IdleConnTimeout
}

func newDefaultTransport(cfg TransportConfig) *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	// There are certain classes of vulnerabilities that open up when
	// compression is enabled. For simplicity, we disable compression
	// to cut off this class of attacks.
	// https://www.cyberis.co.uk/2013/08/vulnerabilities-that-just-wont-die.html
	t.DisableCompression = true
	cfg.apply(t)
	return t
}
