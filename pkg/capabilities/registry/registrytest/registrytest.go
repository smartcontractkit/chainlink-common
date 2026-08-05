// Package registrytest is the in-memory plumbing for testing against a real
// CapabilitiesRegistry.
//
// It holds no registry of its own: tests run the production
// registry/server.Registry over an in-memory listener, so what they exercise is
// the same code that runs in the proxy rather than a stub that can drift from it.
// What is left here is only what a test cannot get from production code — a
// listener that needs no port, and a book of fake capability addresses that can be
// dialed back.
package registrytest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/server"
)

// Serve runs the real registry server for reg on an in-memory listener and
// returns a connection to it, cleaning both up when the test ends.
//
// An in-memory pipe rather than a loopback port: nothing here depends on real
// addressing, and binding a port needs a permission some sandboxes withhold,
// which turns into silently skipped tests.
func Serve(t testing.TB, reg *server.Registry) *grpc.ClientConn {
	t.Helper()

	s := grpc.NewServer()
	server.Register(s, reg)

	lis := bufconn.Listen(1 << 20)
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	conn, err := grpc.NewClient("passthrough:///registry",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("failed to dial the test registry: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

// AddrBook hands out in-memory capability addresses and dials them back, so a
// client's dial-by-address path runs for real without needing ports.
//
// Addresses use the passthrough scheme: gRPC would otherwise parse a bare
// "host:port" as a URI scheme and try to resolve it via DNS.
type AddrBook struct {
	mu        sync.Mutex
	n         int
	listeners map[string]*bufconn.Listener
	dials     map[string]int
}

func NewAddrBook() *AddrBook {
	return &AddrBook{
		listeners: map[string]*bufconn.Listener{},
		dials:     map[string]int{},
	}
}

// Listen registers a new listener and returns it. Its Addr is unique per call,
// so a caller that opens one listener per capability announces distinct
// addresses, exactly as it would with port 0.
func (b *AddrBook) Listen() net.Listener {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.n++
	name := fmt.Sprintf("passthrough:///cap-%d", b.n)
	lis := bufconn.Listen(1 << 20)
	b.listeners[name] = lis
	return &namedListener{Listener: lis, addr: memAddr(name)}
}

// Target returns the address for an explicitly named listener, registering it.
func (b *AddrBook) Target(name string) string { return "passthrough:///" + name }

// Serve registers srv under name and starts serving it, returning the address a
// registry should hand out for it.
func (b *AddrBook) Serve(t testing.TB, name string, srv *grpc.Server) string {
	t.Helper()

	lis := bufconn.Listen(1 << 20)
	target := b.Target(name)

	b.mu.Lock()
	b.listeners[target] = lis
	b.mu.Unlock()

	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	return target
}

// DialOption resolves the addresses this book handed out. Unknown addresses
// fail, so "the registry returned an address nobody serves" is observable rather
// than a hang.
func (b *AddrBook) DialOption() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, target string) (net.Conn, error) {
		b.mu.Lock()
		// The passthrough resolver strips the scheme before dialing, so accept
		// either form.
		lis, ok := b.listeners[target]
		if !ok {
			lis, ok = b.listeners["passthrough:///"+target]
		}
		b.dials[target]++
		b.mu.Unlock()

		if !ok {
			return nil, fmt.Errorf("nothing serving %s", target)
		}
		return lis.DialContext(ctx)
	})
}

// DialCount reports how many transport dials were made to name, which is how a
// test asserts that connections are reused rather than reopened per call.
func (b *AddrBook) DialCount(name string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dials[name] + b.dials["passthrough:///"+name]
}

type memAddr string

func (a memAddr) Network() string { return "bufconn" }
func (a memAddr) String() string  { return string(a) }

// namedListener overrides bufconn's fixed Addr so each listener has its own.
type namedListener struct {
	net.Listener
	addr memAddr
}

func (l *namedListener) Addr() net.Addr { return l.addr }

// Metadata is a server.MetadataSource backed by plain fields.
//
// This is the one part of the registry a test cannot run for real: in production
// the source is a poller over an on-chain contract. Everything above it — the
// registry, its wire adapter, the client — is the production code.
//
// A zero Metadata reports nothing, which the server surfaces as NotFound or
// FailedPrecondition, matching a registry that has not synced yet.
type Metadata struct {
	Node       *capabilities.Node
	Nodes      map[ragetypes.PeerID]capabilities.Node
	Configs    map[string][]byte
	DONsForCap map[string][]capabilities.DONWithNodes
	DONsByID   map[uint32]capabilities.DON
}

var _ server.MetadataSource = (*Metadata)(nil)

func (m *Metadata) LocalNode(context.Context) (capabilities.Node, error) {
	if m.Node == nil {
		return capabilities.Node{}, errors.New("registry not synced yet")
	}
	return *m.Node, nil
}

func (m *Metadata) NodeByPeerID(_ context.Context, peerID ragetypes.PeerID) (capabilities.Node, error) {
	n, ok := m.Nodes[peerID]
	if !ok {
		return capabilities.Node{}, fmt.Errorf("no such peer %s", peerID)
	}
	return n, nil
}

func (m *Metadata) DONsForCapability(_ context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	dons, ok := m.DONsForCap[capabilityID]
	if !ok {
		return nil, fmt.Errorf("no DONs for capability %s", capabilityID)
	}
	return dons, nil
}

func (m *Metadata) DONByID(_ context.Context, donID uint32) (capabilities.DON, error) {
	d, ok := m.DONsByID[donID]
	if !ok {
		return capabilities.DON{}, fmt.Errorf("no DON %d", donID)
	}
	return d, nil
}

func (m *Metadata) RawConfigForCapability(_ context.Context, capabilityID string, donID uint32) ([]byte, error) {
	cfg, ok := m.Configs[capabilityID]
	if !ok {
		return nil, fmt.Errorf("no config for capability %s on DON %d", capabilityID, donID)
	}
	return cfg, nil
}
