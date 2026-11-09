package remote

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/remote/pb"
)

// LOOPs built against older versions dial this path, so it must not change.
func TestSettingsWirePath(t *testing.T) {
	require.Equal(t, "/capabilities.Settings/Subscribe", pb.Settings_Subscribe_FullMethodName)
}

func TestSettings_RoundTrip(t *testing.T) {
	lggr := logger.Test(t)
	host := &recordingBroadcaster{Atomic: &settings.Atomic{}, unsubscribed: make(chan struct{})}
	require.NoError(t, host.Store(settings.Snapshot{Settings: `Foo = "initial"`, Hash: "initial"}))

	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	RegisterSettingsServer(srv, lggr, host)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	ctx, cancel := context.WithCancel(t.Context())
	ch, err := NewSettingsClient(lggr, conn).Subscribe(ctx)
	require.NoError(t, err)

	recv := func() settings.Snapshot {
		select {
		case got := <-ch:
			return got
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for settings")
			return settings.Snapshot{}
		}
	}

	require.Equal(t, settings.Snapshot{Settings: `Foo = "initial"`, Hash: "initial"}, recv())

	require.NoError(t, host.Store(settings.Snapshot{Settings: `Foo = "next"`, Hash: "next"}))
	require.Equal(t, settings.Snapshot{Settings: `Foo = "next"`, Hash: "next"}, recv())

	cancel()
	select {
	case <-host.unsubscribed:
	case <-time.After(5 * time.Second):
		t.Fatal("the server should unsubscribe when the client goes away")
	}
}

type recordingBroadcaster struct {
	*settings.Atomic
	unsubscribed chan struct{}
}

func (r *recordingBroadcaster) Unsubscribe(ch <-chan settings.Snapshot) {
	r.Atomic.Unsubscribe(ch)
	close(r.unsubscribed)
}
