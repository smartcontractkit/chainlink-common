package remote

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/remote/pb"
)

// NewSettingsClient returns a [settings.Broadcaster] backed by a remote Settings service. Each subscription is one
// server stream, forwarded until it ends or ctx is done.
func NewSettingsClient(lggr logger.Logger, cc grpc.ClientConnInterface) settings.Broadcaster {
	return &client{lggr: logger.Named(lggr, "SettingsClient"), grpc: pb.NewSettingsClient(cc)}
}

type client struct {
	lggr logger.Logger
	grpc pb.SettingsClient
}

func (c *client) Subscribe(ctx context.Context) (<-chan settings.Snapshot, error) {
	stream, err := c.grpc.Subscribe(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	ch := make(chan settings.Snapshot)
	go func() {
		defer close(ch)
		for {
			update, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) && stream.Context().Err() == nil {
					c.lggr.Errorw("Failed to receive settings update", "err", err)
				}
				return
			}
			select {
			case <-stream.Context().Done():
				return
			case ch <- settings.Snapshot{Settings: update.Settings, Hash: update.Hash}:
			}
		}
	}()
	return ch, nil
}

func RegisterSettingsServer(s grpc.ServiceRegistrar, lggr logger.Logger, impl settings.Broadcaster) {
	pb.RegisterSettingsServer(s, NewSettingsServer(lggr, impl))
}

func NewSettingsServer(lggr logger.Logger, impl settings.Broadcaster) pb.SettingsServer {
	return &server{lggr: logger.Named(lggr, "SettingsServer"), impl: impl}
}

type server struct {
	pb.UnimplementedSettingsServer
	lggr logger.Logger
	impl settings.Broadcaster
}

func (s *server) Subscribe(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.SettingsUpdate]) error {
	ch, err := s.impl.Subscribe(stream.Context())
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	if local, ok := s.impl.(settings.LocalBroadcaster); ok {
		defer local.Unsubscribe(ch)
	}
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if err = stream.Send(&pb.SettingsUpdate{Settings: update.Settings, Hash: update.Hash}); err != nil {
				s.lggr.Debugw("Failed to send settings update", "hash", update.Hash, "err", err)
				return fmt.Errorf("error sending settings update: %w", err)
			}
		}
	}
}
