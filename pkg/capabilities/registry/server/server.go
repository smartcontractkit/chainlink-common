package server

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/pb"
)

// Server adapts Registry to the plain-gRPC CapabilitiesRegistry service.
//
// It is a thin translation layer: no broker, no per-capability sub-servers, no
// resource lifetimes. Handle replies carry a URL the caller dials itself.
type Server struct {
	registrypb.UnimplementedCapabilitiesRegistryServer

	impl *Registry
}

var _ registrypb.CapabilitiesRegistryServer = (*Server)(nil)

func NewServer(impl *Registry) *Server { return &Server{impl: impl} }

// Register registers this server on a plain grpc.Server.
func Register(s *grpc.Server, impl *Registry) {
	registrypb.RegisterCapabilitiesRegistryServer(s, NewServer(impl))
}

func (s *Server) Add(ctx context.Context, req *registrypb.AddRequest) (*emptypb.Empty, error) {
	capType, err := capabilitiespb.CapabilityTypeFromProto(req.GetType())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if capType == capabilities.CapabilityTypeUnknown {
		return nil, status.Errorf(codes.InvalidArgument,
			"capability %s has no capability type, so the registry cannot know which services it serves", req.GetCapabilityId())
	}
	h := Handle{ID: req.GetCapabilityId(), Type: capType, URL: req.GetCallbackUrl()}
	if err := s.impl.Add(ctx, h); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) Remove(ctx context.Context, req *registrypb.RemoveRequest) (*emptypb.Empty, error) {
	if err := s.impl.Remove(ctx, req.GetCapabilityId()); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// handle resolves a capability through lookup and converts it for the wire. The
// three Get* RPCs differ only in which lookup they use.
func handle(ctx context.Context, id string, lookup func(context.Context, string) (Handle, error)) (*registrypb.CapabilityHandle, error) {
	h, err := lookup(ctx, id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return handleToProto(h)
}

func (s *Server) Get(ctx context.Context, req *registrypb.GetRequest) (*registrypb.CapabilityHandle, error) {
	return handle(ctx, req.GetCapabilityId(), s.impl.Get)
}

func (s *Server) GetTrigger(ctx context.Context, req *registrypb.GetRequest) (*registrypb.CapabilityHandle, error) {
	return handle(ctx, req.GetCapabilityId(), s.impl.GetTrigger)
}

func (s *Server) GetExecutable(ctx context.Context, req *registrypb.GetRequest) (*registrypb.CapabilityHandle, error) {
	return handle(ctx, req.GetCapabilityId(), s.impl.GetExecutable)
}

func (s *Server) List(ctx context.Context, _ *emptypb.Empty) (*registrypb.ListReply, error) {
	handles := s.impl.List(ctx)
	out := make([]*registrypb.CapabilityHandle, 0, len(handles))
	for _, h := range handles {
		ph, err := handleToProto(h)
		if err != nil {
			return nil, err
		}
		out = append(out, ph)
	}
	return &registrypb.ListReply{Handles: out}, nil
}

func (s *Server) LocalNode(ctx context.Context, _ *emptypb.Empty) (*registrypb.NodeReply, error) {
	n, err := s.impl.LocalNode(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	return nodeToProto(n), nil
}

func (s *Server) NodeByPeerID(ctx context.Context, req *registrypb.NodeRequest) (*registrypb.NodeReply, error) {
	peerID, err := toPeerID(req.GetPeerId())
	if err != nil {
		return nil, err
	}
	n, err := s.impl.NodeByPeerID(ctx, peerID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return nodeToProto(n), nil
}

func (s *Server) ConfigForCapability(ctx context.Context, req *registrypb.ConfigForCapabilityRequest) (*registrypb.ConfigForCapabilityReply, error) {
	raw, err := s.impl.RawConfigForCapability(ctx, req.GetCapabilityId(), req.GetDonId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// The contract stores a wire-encoded CapabilityConfig, so this only parses the
	// bytes into the shared message; none of the config is interpreted here.
	cfg := &capabilitiespb.CapabilityConfig{}
	if err := proto.Unmarshal(raw, cfg); err != nil {
		return nil, status.Errorf(codes.Internal,
			"capability %s on DON %d has an unparseable on-chain config: %s",
			req.GetCapabilityId(), req.GetDonId(), err)
	}
	return &registrypb.ConfigForCapabilityReply{CapabilityConfig: cfg}, nil
}

func (s *Server) DONsForCapability(ctx context.Context, req *registrypb.DONsForCapabilityRequest) (*registrypb.DONsForCapabilityReply, error) {
	dons, err := s.impl.DONsForCapability(ctx, req.GetCapabilityId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	out := make([]*registrypb.DONWithNodes, 0, len(dons))
	for _, d := range dons {
		nodes := make([]*registrypb.NodeReply, 0, len(d.Nodes))
		for _, n := range d.Nodes {
			nodes = append(nodes, nodeToProto(n))
		}
		out = append(out, &registrypb.DONWithNodes{Don: donToProto(d.DON), Nodes: nodes})
	}
	return &registrypb.DONsForCapabilityReply{Dons: out}, nil
}

func (s *Server) DONByID(ctx context.Context, req *registrypb.DONByIDRequest) (*registrypb.DONByIDReply, error) {
	d, err := s.impl.DONByID(ctx, req.GetDonId())
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &registrypb.DONByIDReply{Don: donToProto(d)}, nil
}

// --- conversions ---

func handleToProto(h Handle) (*registrypb.CapabilityHandle, error) {
	pbType, err := capabilitiespb.CapabilityTypeToProto(h.Type)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &registrypb.CapabilityHandle{
		CapabilityId: h.ID,
		Type:         pbType,
		CallbackUrl:  h.URL,
	}, nil
}

func donToProto(d capabilities.DON) *registrypb.DON {
	members := make([][]byte, len(d.Members))
	for i, m := range d.Members {
		members[i] = m[:]
	}
	return &registrypb.DON{
		Id:               d.ID,
		Name:             d.Name,
		Members:          members,
		F:                uint32(d.F),
		ConfigVersion:    d.ConfigVersion,
		Families:         d.Families,
		Config:           d.Config,
		IsPublic:         d.IsPublic,
		AcceptsWorkflows: d.AcceptsWorkflows,
	}
}

func nodeToProto(n capabilities.Node) *registrypb.NodeReply {
	reply := &registrypb.NodeReply{
		WorkflowDon:         donToProto(n.WorkflowDON),
		NodeOperatorId:      n.NodeOperatorID,
		Signer:              n.Signer[:],
		EncryptionPublicKey: n.EncryptionPublicKey[:],
	}
	if n.PeerID != nil {
		reply.PeerId = n.PeerID[:]
	}
	for _, d := range n.CapabilityDONs {
		reply.CapabilityDons = append(reply.CapabilityDons, donToProto(d))
	}
	return reply
}

func toPeerID(b []byte) (ragetypes.PeerID, error) {
	var peerID ragetypes.PeerID
	if len(b) != len(peerID) {
		return peerID, status.Errorf(codes.InvalidArgument,
			"invalid peer ID length: got %d, want %d", len(b), len(peerID))
	}
	copy(peerID[:], b)
	return peerID, nil
}
