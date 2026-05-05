package eventstore

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

var _ pb.EventStoreServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedEventStoreServer
	impl capabilities.EventStore
}

func NewServer(impl capabilities.EventStore) *Server {
	return &Server{impl: impl}
}

func (s *Server) Insert(ctx context.Context, req *pb.InsertEventRequest) (*emptypb.Empty, error) {
	if s.impl == nil {
		return nil, errors.New("event store implementation is nil")
	}
	ev := req.GetEvent()
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
	if err := s.impl.Insert(ctx, rec); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateDelivery(ctx context.Context, req *pb.UpdateDeliveryRequest) (*emptypb.Empty, error) {
	if s.impl == nil {
		return nil, errors.New("event store implementation is nil")
	}
	var lastSentAt = req.GetLastSentAt().AsTime()
	if err := s.impl.UpdateDelivery(ctx, req.GetTriggerId(), req.GetEventId(), lastSentAt, int(req.GetAttempts())); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) List(ctx context.Context, _ *emptypb.Empty) (*pb.ListEventsResponse, error) {
	if s.impl == nil {
		return nil, errors.New("event store implementation is nil")
	}
	events, err := s.impl.List(ctx)
	if err != nil {
		return nil, err
	}
	resp := &pb.ListEventsResponse{
		Events: make([]*pb.PendingEventProto, 0, len(events)),
	}
	for _, ev := range events {
		pev := &pb.PendingEventProto{
			TriggerId: ev.TriggerId,
			EventId:   ev.EventId,
			Payload:   ev.Payload,
			FirstAt:   timestamppb.New(ev.FirstAt),
			Attempts:  int32(ev.Attempts),
		}
		if !ev.LastSentAt.IsZero() {
			pev.LastSentAt = timestamppb.New(ev.LastSentAt)
		}
		resp.Events = append(resp.Events, pev)
	}
	return resp, nil
}

func (s *Server) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*emptypb.Empty, error) {
	if s.impl == nil {
		return nil, errors.New("event store implementation is nil")
	}
	if err := s.impl.DeleteEvent(ctx, req.GetTriggerId(), req.GetEventId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteEventsForTrigger(ctx context.Context, req *pb.DeleteEventsForTriggerRequest) (*emptypb.Empty, error) {
	if s.impl == nil {
		return nil, errors.New("event store implementation is nil")
	}
	if err := s.impl.DeleteEventsForTrigger(ctx, req.GetTriggerId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
