package settings

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

func NewClient(lggr logger.Logger, conn grpc.ClientConnInterface) core.SettingsBroadcaster {
	return &client{lggr: logger.Named(lggr, "CRESettingsClient"), grpc: pb.NewSettingsClient(conn)}
}

type client struct {
	lggr logger.Logger
	grpc pb.SettingsClient
}

func (c *client) Subscribe(ctx context.Context) (<-chan core.SettingsUpdate, error) {
	stream, err := c.grpc.Subscribe(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	ch := make(chan core.SettingsUpdate)
	go func(ctx context.Context) {
		defer close(ch)
		for {
			update, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				c.lggr.Errorw("Error receiving update", "err", err)
				return
			}
			select {
			case <-ctx.Done():
				return
			case ch <- core.SettingsUpdate{Settings: update.Settings, Hash: update.Hash}:
				c.lggr.Debug("Forwarded updated settings")
			}
		}
	}(stream.Context())

	return ch, nil
}

func NewServer(impl core.SettingsBroadcaster) *server {
	return &server{impl: impl}
}

type server struct {
	pb.UnimplementedSettingsServer
	impl core.SettingsBroadcaster
}

func (c *server) Subscribe(empty *emptypb.Empty, g grpc.ServerStreamingServer[pb.SettingsUpdate]) error {
	ch, err := c.impl.Subscribe(g.Context())
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}
	if local, ok := c.impl.(core.SettingsBroadcasterLocal); ok {
		defer local.Unsubscribe(ch)
	}
	for {
		select {
		case <-g.Context().Done():
			return nil
		case update, ok := <-ch:
			if !ok {
				return nil
			}
			if err = g.Send(&pb.SettingsUpdate{Settings: update.Settings, Hash: update.Hash}); err != nil {
				return fmt.Errorf("error sending update for CRE settings: %w", err)
			}
		}
	}
}
