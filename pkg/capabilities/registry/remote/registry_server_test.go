package remote

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type countingTransport struct {
	Transport
	open atomic.Int32
}

func (t *countingTransport) Publish(string, func(*grpc.Server)) (Locator, io.Closer, error) {
	t.open.Add(1)
	return Locator{Target: "t"}, closerFunc(func() error { t.open.Add(-1); return nil }), nil
}

type closerFunc func() error

func (f closerFunc) Close() error { return f() }

type stubCapability struct{ info capabilities.CapabilityInfo }

func (s stubCapability) Info(context.Context) (capabilities.CapabilityInfo, error) {
	return s.info, nil
}

// blockingRegistry pauses Get after the lookup so a Remove can interleave.
type blockingRegistry struct {
	registry.CapabilitiesRegistry
	mu        sync.Mutex
	caps      map[string]capabilities.BaseCapability
	looked    chan struct{}
	resumeGet chan struct{}
}

func (r *blockingRegistry) Get(_ context.Context, id string) (capabilities.BaseCapability, error) {
	r.mu.Lock()
	c := r.caps[id]
	r.mu.Unlock()
	close(r.looked)
	<-r.resumeGet
	return c, nil
}

func (r *blockingRegistry) Remove(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.caps, id)
	return nil
}

func TestServer_RemoveDuringGetDoesNotLeakPublishedCapability(t *testing.T) {
	ctx := t.Context()
	id := "cap@1.0.0"
	impl := &blockingRegistry{
		caps:      map[string]capabilities.BaseCapability{id: stubCapability{capabilities.MustNewCapabilityInfo(id, capabilities.CapabilityTypeAction, "")}},
		looked:    make(chan struct{}),
		resumeGet: make(chan struct{}),
	}
	tr := &countingTransport{}
	srv := NewCapabilitiesRegistryServer(logger.Test(t), impl, tr)

	getDone := make(chan struct{})
	go func() {
		defer close(getDone)
		_, _ = srv.Get(ctx, &registrypb.GetRequest{Id: id})
	}()
	<-impl.looked

	removeDone := make(chan struct{})
	go func() {
		defer close(removeDone)
		_, err := srv.Remove(ctx, &registrypb.RemoveRequest{Id: id})
		assert.NoError(t, err)
	}()

	close(impl.resumeGet)
	<-getDone
	<-removeDone

	require.Equal(t, int32(0), tr.open.Load(), "removed capability is still published")
}
