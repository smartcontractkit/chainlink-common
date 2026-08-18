package loop

import (
	"context"
	"errors"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// webServer serves web routes including
// - /metrics prometheus metrics
// - /debug/pprof/ debug profiling
type webServer struct {
	lggr    logger.Logger
	port    int
	handler http.Handler

	srvr        *http.Server
	tcpListener *net.TCPListener

	done chan struct{}
}

type WebServerOpts struct {
	Handler http.Handler
}

func (o WebServerOpts) New(lggr logger.Logger, port int) *webServer {
	s := &webServer{
		lggr:    logger.Named(lggr, "WebServer"),
		port:    port,
		handler: o.Handler,
		srvr: &http.Server{
			// reasonable default based on typical prom poll interval of 15s.
			ReadTimeout: 5 * time.Second,
		},
		done: make(chan struct{}),
	}
	if s.handler == nil {
		s.handler = promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		)
	}
	return s
}

// setupListener creates an explicit listener so that we can resolve `:0` port, which is needed for testing
// if we didn't need the resolved addr, or could pick a static port we could use p.srvr.ListenAndServer
func (w *webServer) setupListener() error {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: w.port,
	})
	if err != nil {
		return err
	}

	w.tcpListener = l
	return nil
}

func (w *webServer) Start(ctx context.Context) error {
	err := w.setupListener()
	if err != nil {
		return err
	}

	http.Handle("/metrics", w.handler)

	// pprof handler registered via import side effects

	go func() {
		defer close(w.done)
		err := w.srvr.Serve(w.tcpListener)
		if !errors.Is(err, http.ErrServerClosed) {
			w.lggr.Errorw("Unexpected server error", "err", err)
		}
	}()
	return nil
}

func (w *webServer) Close() error {
	err := w.srvr.Shutdown(context.Background())
	<-w.done
	return err
}

func (w *webServer) Ready() error { return nil }

func (w *webServer) HealthReport() map[string]error { return map[string]error{w.Name(): nil} }

func (w *webServer) Name() string { return w.lggr.Name() }
