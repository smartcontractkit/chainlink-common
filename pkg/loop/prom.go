package loop

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type PromServer struct {
	Port    int
	Logger  logger.Logger // required
	Handler http.Handler  // optional

	srvrDone    chan struct{} // closed when the http server is done
	srvr        *http.Server
	tcpListener *net.TCPListener
}

func (p *PromServer) init() *PromServer {
	p.srvrDone = make(chan struct{})
	p.srvr = &http.Server{
		// reasonable default based on typical prom poll interval of 15s.
		ReadTimeout: 5 * time.Second,
	}
	if p.Handler == nil {
		p.Handler = promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{
				EnableOpenMetrics: true,
			},
		)
	}
	return p
}

// Start starts HTTP server on specified port to handle metrics requests
func (p *PromServer) Start() error {
	if p.Logger == nil {
		return errors.New("nil Logger")
	}
	p.init()
	p.Logger.Debugf("Starting prom server on port %d", p.Port)
	err := p.setupListener()
	if err != nil {
		return err
	}

	http.Handle("/metrics", p.Handler)

	go func() {
		defer close(p.srvrDone)
		err := p.srvr.Serve(p.tcpListener)
		if errors.Is(err, net.ErrClosed) {
			// ErrClose is expected on gracefully shutdown
			p.Logger.Warnf("%s closed", p.Name())
		} else {
			p.Logger.Errorf("%s: %s", p.Name(), err)
		}

	}()
	return nil
}

// Close shuts down the underlying HTTP server. See [http.Server.Close] for details
func (p *PromServer) Close() error {
	err := p.srvr.Shutdown(context.Background())
	<-p.srvrDone
	return err
}

// Name of the server
func (p *PromServer) Name() string {
	return fmt.Sprintf("%s-prom-server", p.Logger.Name())
}

// setupListener creates explicit listener so that we can resolve `:0` port, which is needed for testing
// if we didn't need the resolved addr, or could pick a static port we could use p.srvr.ListenAndServer
func (p *PromServer) setupListener() error {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: p.Port,
	})
	if err != nil {
		return err
	}

	p.tcpListener = l
	return nil
}
