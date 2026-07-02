package http

import (
	netHttp "net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTransportConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultTransportConfig()
	assert.Equal(t, 100, cfg.MaxIdleConns)
	assert.Equal(t, 100, cfg.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, cfg.IdleConnTimeout)
}

func TestUnrestrictedClientTransportPooling(t *testing.T) {
	t.Parallel()

	client := NewUnrestrictedClient()
	tr := client.Transport.(*netHttp.Transport)
	assert.True(t, tr.DisableCompression)
	assert.Equal(t, 100, tr.MaxIdleConns)
	assert.Equal(t, 100, tr.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, tr.IdleConnTimeout)
}

func TestUnrestrictedClientWithTransportConfig(t *testing.T) {
	t.Parallel()

	client := NewUnrestrictedClientWithTransportConfig(TransportConfig{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 25,
		IdleConnTimeout:     30 * time.Second,
	})
	tr := client.Transport.(*netHttp.Transport)
	assert.Equal(t, 50, tr.MaxIdleConns)
	assert.Equal(t, 25, tr.MaxIdleConnsPerHost)
	assert.Equal(t, 30*time.Second, tr.IdleConnTimeout)
}
