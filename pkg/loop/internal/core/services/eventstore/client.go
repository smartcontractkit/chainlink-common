package eventstore

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var _ capabilities.EventStore = (*Client)(nil)

type Client struct {
	grpc pb.EventStoreClient
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{grpc: pb.NewEventStoreClient(cc)}
}

func (c *Client) Insert(ctx context.Context, rec capabilities.PendingEvent) error {
	ev := &pb.PendingEventProto{
		TriggerId: rec.TriggerId,
		EventId:   rec.EventId,
		Payload:   rec.Payload,
		FirstAt:   timestamppb.New(rec.FirstAt),
		Attempts:  int32(rec.Attempts),
	}
	if !rec.LastSentAt.IsZero() {
		ev.LastSentAt = timestamppb.New(rec.LastSentAt)
	}
	_, err := c.grpc.Insert(ctx, &pb.InsertEventRequest{Event: ev})
	return err
}

func (c *Client) UpdateDelivery(ctx context.Context, triggerId string, eventId string, lastSentAt time.Time, attempts int) error {
	_, err := c.grpc.UpdateDelivery(ctx, &pb.UpdateDeliveryRequest{
		TriggerId:  triggerId,
		EventId:    eventId,
		LastSentAt: timestamppb.New(lastSentAt),
		Attempts:   int32(attempts),
	})
	return err
}

func (c *Client) List(ctx context.Context) ([]capabilities.PendingEvent, error) {
	resp, err := c.grpc.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	events := make([]capabilities.PendingEvent, 0, len(resp.GetEvents()))
	for _, ev := range resp.GetEvents() {
		rec := capabilities.PendingEvent{
			TriggerId: ev.GetTriggerId(),
			EventId:   ev.GetEventId(),
			Payload:   ev.GetPayload(),
			Attempts:  int(ev.GetAttempts()),
		}
		if t := ev.GetFirstAt(); t != nil {
			rec.FirstAt = t.AsTime()
		}
		if t := ev.GetLastSentAt(); t != nil {
			rec.LastSentAt = t.AsTime()
		}
		events = append(events, rec)
	}
	return events, nil
}

func (c *Client) DeleteEvent(ctx context.Context, triggerId string, eventId string) error {
	_, err := c.grpc.DeleteEvent(ctx, &pb.DeleteEventRequest{
		TriggerId: triggerId,
		EventId:   eventId,
	})
	return err
}

func (c *Client) DeleteEventsForTrigger(ctx context.Context, triggerID string) error {
	_, err := c.grpc.DeleteEventsForTrigger(ctx, &pb.DeleteEventsForTriggerRequest{
		TriggerId: triggerID,
	})
	return err
}
