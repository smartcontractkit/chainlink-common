package loop

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

func TestPromServer(t *testing.T) {

	testReg := prometheus.NewRegistry()
	testHandler := promhttp.HandlerFor(testReg, promhttp.HandlerOpts{})
	testMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_metric",
	})
	testReg.MustRegister(testMetric)
	testMetric.Inc()

	s := &PromServer{Handler: testHandler, Logger: logger.Test(t)}
	// check that port is not resolved yet
	require.Equal(t, -1, s.ResolvedPort())
	require.NoError(t, s.Start())

	url := fmt.Sprintf("http://localhost:%d/metrics", s.ResolvedPort())
	resp, err := http.Get(url) //nolint
	require.NoError(t, err)
	require.NoError(t, err, "endpoint %s", url)
	require.NotNil(t, resp.Body)
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(b), "test_metric")
	defer resp.Body.Close()

	require.NoError(t, s.Close())
}

// ResolvedPort is the resolved port and is only known after Start().
// returns -1 before it is resolved or if there was an error during resolution.
func (p *PromServer) ResolvedPort() int {
	if p.tcpListener == nil {
		return -1
	}
	// always safe to cast because we explicitly have a tcp listener
	// there is direct access to Port without the addr casting
	// Note: addr `:0` is not resolved to non-zero port until ListenTCP is called
	// net.ResolveTCPAddr sounds promising, but doesn't work in practice
	return p.tcpListener.Addr().(*net.TCPAddr).Port
}
