package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
)

// Broker is a subset of the methods exported by *plugin.GRPCBroker.
type Broker interface {
	Accept(id uint32) (net.Listener, error)
	DialWithOptions(id uint32, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error)
	NextId() uint32
}

var _ Broker = (*atomicBroker)(nil)

// An atomicBroker implements [Broker] and is backed by a swappable [*plugin.GRPCBroker]
type atomicBroker struct {
	broker atomic.Pointer[Broker]
}

func (a *atomicBroker) store(b Broker) { a.broker.Store(&b) }
func (a *atomicBroker) load() Broker   { return *a.broker.Load() }

func (a *atomicBroker) Accept(id uint32) (net.Listener, error) {
	return a.load().Accept(id)
}

func (a *atomicBroker) DialWithOptions(id uint32, opts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	return a.load().DialWithOptions(id, opts...)
}

func (a *atomicBroker) NextId() uint32 { //nolint:revive
	return a.load().NextId()
}

// GRPCOpts has GRPC client and server options.
type GRPCOpts struct {
	// Optionally include additional options when dialing a client.
	// Normally aligned with [plugin.ClientConfig.GRPCDialOptions].
	DialOpts []grpc.DialOption
	// Optionally override the default *grpc.Server constructor.
	// Normally aligned with [plugin.ServeConfig.GRPCServer].
	NewServer func([]grpc.ServerOption) *grpc.Server
}

// BrokerConfig holds Broker configuration fields.
type BrokerConfig struct {
	StopCh <-chan struct{}
	Logger logger.Logger

	GRPCOpts // optional
}

type BrokerOrchestrator interface {
	Broker
	LoggerStopper

	CloneOrchestratorTo(name string) BrokerOrchestrator

	Dial(id uint32) (conn *grpc.ClientConn, err error)
	Config() BrokerConfig
	stopCh() <-chan struct{}
}

type LoggerStopper interface {
	StopCtx() (context.Context, context.CancelFunc)
	Logger() logger.Logger
	// create a new loggerstopper with the given name
	Named(name string) LoggerStopper
}

type loggerStopper struct {
	cfg BrokerConfig
}

func (l *loggerStopper) StopCtx() (context.Context, context.CancelFunc) {
	return utils.ContextFromChan(l.cfg.StopCh)
}

func (l *loggerStopper) Logger() logger.Logger {
	return l.cfg.Logger
}

func (l *loggerStopper) Named(name string) LoggerStopper {
	l2 := *l
	l2.cfg.Logger = logger.Named(l.cfg.Logger, name)
	return &l2
}

type ResourceManager interface {
	BrokerOrchestrator
	Serve(name string, server *grpc.Server, deps ...Resource) (uint32, Resource, error)
	ServeNew(name string, register func(*grpc.Server), deps ...Resource) (uint32, Resource, error)
	Dial(id uint32) (conn *grpc.ClientConn, err error)
	CloseAll(resources ...Resource)
	New(name string) ResourceManager
}

type ConfiguredBroker struct {
	Broker
	cfg BrokerConfig
}

var _ BrokerOrchestrator = (*ConfiguredBroker)(nil)

func (b *ConfiguredBroker) Config() BrokerConfig {
	return b.cfg
}

func (b *ConfiguredBroker) LoggerStopper() LoggerStopper {
	return &loggerStopper{b.cfg}
}

func (b *ConfiguredBroker) StopCtx() (context.Context, context.CancelFunc) {
	return utils.ContextFromChan(b.cfg.StopCh)
}

func (b *ConfiguredBroker) Named(name string) LoggerStopper {
	return b.CloneOrchestratorTo(name)
}

func (b *ConfiguredBroker) CloneOrchestratorTo(name string) BrokerOrchestrator {
	bn := *b
	bn.cfg.Logger = logger.Named(b.cfg.Logger, name)
	return &bn
}

func (b *ConfiguredBroker) Logger() logger.Logger {
	return b.cfg.Logger
}

func (b *ConfiguredBroker) Opts() GRPCOpts {
	return b.cfg.GRPCOpts
}

func (b *ConfiguredBroker) Dial(id uint32) (conn *grpc.ClientConn, err error) {
	return b.DialWithOptions(id, b.cfg.GRPCOpts.DialOpts...)
}

func (b *ConfiguredBroker) stopCh() <-chan struct{} {
	return b.cfg.StopCh
}

type BrokerManager struct {
	BrokerOrchestrator
}

func NewBrokerManager(b Broker, cfg BrokerConfig) *BrokerManager {
	return &BrokerManager{&ConfiguredBroker{b, cfg}}
}

var _ ResourceManager = (*BrokerManager)(nil)

func (b *BrokerManager) New(name string) ResourceManager {
	bn := *b
	bn.BrokerOrchestrator = b.BrokerOrchestrator.CloneOrchestratorTo(name)
	return &bn
}

func (b *BrokerManager) Serve(name string, server *grpc.Server, deps ...Resource) (uint32, Resource, error) {
	id := b.NextId()
	b.Logger().Debugf("Serving %s on connection %d", name, id)
	lis, err := b.Accept(id)
	if err != nil {
		b.CloseAll(deps...)
		return 0, Resource{}, ErrConnAccept{Name: name, ID: id, Err: err}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer b.CloseAll(deps...)
		if err := server.Serve(lis); err != nil {
			b.Logger().Errorw(fmt.Sprintf("Failed to serve %s on connection %d", name, id), "err", err)
		}
	}()

	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-b.stopCh():
			server.Stop()
		case <-done:
		}
	}()

	return id, Resource{fnCloser(func() {
		server.Stop()
		close(done)
		wg.Wait()
	}), name}, nil
}

func (b *BrokerManager) ServeNew(name string, register func(*grpc.Server), deps ...Resource) (uint32, Resource, error) {
	var server *grpc.Server
	if b.Config().GRPCOpts.NewServer == nil {
		server = grpc.NewServer()
	} else {
		server = b.Config().GRPCOpts.NewServer(nil)
	}
	register(server)
	return b.Serve(name, server, deps...)
}

func (b *BrokerManager) Dial(id uint32) (conn *grpc.ClientConn, err error) {
	return b.DialWithOptions(id, b.Config().DialOpts...)
}

func (b *BrokerManager) CloseAll(resources ...Resource) {
	for _, d := range resources {
		if err := d.Close(); err != nil {
			b.Logger().Error(fmt.Sprintf("Error closing %s", d.Name()), "err", err)
		}
	}
}

func CloseAll(resources ...Resource) error {
	var err error
	for _, r := range resources {
		if err2 := r.Close(); err != nil {
			err = errors.Join(err, fmt.Errorf("error closing %s: %w", r.Name(), err2))
		}
	}
	return err
}

func CloseAndLog(logger logger.Logger, resources ...Resource) {
	if err := CloseAll(resources...); err != nil {
		logger.Errorw("Error closing resources", "err", err)
	}
}

// newClientConn return a new *clientConn backed by this *brokerExt.
func NewClientConn(name string, b Broker, cfg BrokerConfig, newClient newClientFn) *clientConn {
	cfg2 := cfg
	cfg2.Logger = logger.Named(cfg.Logger, name)
	return &clientConn{
		Broker:       b,
		BrokerConfig: cfg2,
		newClient:    newClient,
		name:         name,
	}
}

/*
// brokerExt extends a Broker with various helper methods.
type brokerExt struct {
	broker Broker
	BrokerConfig
}

// withName returns a new [*brokerExt] with name added to the logger.
func (b *brokerExt) withName(name string) *brokerExt {
	bn := *b
	bn.Logger = logger.Named(b.Logger, name)
	return &bn
}

// newClientConn return a new *clientConn backed by this *brokerExt.
func (b *brokerExt) newClientConn(name string, newClient newClientFn) *clientConn {
	return &clientConn{
		brokerExt: b.withName(name),
		newClient: newClient,
		name:      name,
	}
}

func (b *brokerExt) stopCtx() (context.Context, context.CancelFunc) {
	return utils.ContextFromChan(b.StopCh)
}

func (b *brokerExt) dial(id uint32) (conn *grpc.ClientConn, err error) {
	return b.broker.DialWithOptions(id, b.DialOpts...)
}

func (b *brokerExt) serveNew(name string, register func(*grpc.Server), deps ...resource) (uint32, resource, error) {
	var server *grpc.Server
	if b.NewServer == nil {
		server = grpc.NewServer()
	} else {
		server = b.NewServer(nil)
	}
	register(server)
	return b.serve(name, server, deps...)
}

func (b *brokerExt) serve(name string, server *grpc.Server, deps ...resource) (uint32, resource, error) {
	id := b.broker.NextId()
	b.Logger.Debugf("Serving %s on connection %d", name, id)
	lis, err := b.broker.Accept(id)
	if err != nil {
		b.closeAll(deps...)
		return 0, resource{}, ErrConnAccept{Name: name, ID: id, Err: err}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer b.closeAll(deps...)
		if err := server.Serve(lis); err != nil {
			b.Logger.Errorw(fmt.Sprintf("Failed to serve %s on connection %d", name, id), "err", err)
		}
	}()

	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-b.StopCh:
			server.Stop()
		case <-done:
		}
	}()

	return id, resource{fnCloser(func() {
		server.Stop()
		close(done)
		wg.Wait()
	}), name}, nil
}

func (b *brokerExt) closeAll(deps ...resource) {
	for _, d := range deps {
		if err := d.Close(); err != nil {
			b.Logger.Error(fmt.Sprintf("Error closing %s", d.name), "err", err)
		}
	}
}
*/

/*
type Resource interface {
	io.Closer
	Name() string
}
*/

func NewResource(c io.Closer, name string) Resource {
	return Resource{c, name}
}

type Resource struct {
	io.Closer
	name string
}

func (r Resource) Name() string {
	return r.name
}

type Resources []Resource

func (rs *Resources) Add(r Resource) {
	*rs = append(*rs, r)
}

func (rs *Resources) Stop(s interface{ Stop() }, name string) {
	rs.Add(Resource{fnCloser(s.Stop), name})
}

func (rs *Resources) Close(c io.Closer, name string) {
	rs.Add(Resource{c, name})
}

// fnCloser implements io.Closer with a func().
type fnCloser func()

func (s fnCloser) Close() error {
	s()
	return nil
}
