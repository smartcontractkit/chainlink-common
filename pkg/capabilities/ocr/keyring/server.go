package keyring

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/keyring/internal/pb"
)

// RegisterServer serves keyrings on s.
func RegisterServer(s grpc.ServiceRegistrar, keyrings Keyrings) {
	pb.RegisterKeyringServer(s, &server{keyrings: keyrings})
}

type server struct {
	pb.UnimplementedKeyringServer
	keyrings Keyrings
}

func (s *server) PublicKeys(_ context.Context, _ *emptypb.Empty) (*pb.PublicKeysReply, error) {
	offchainPublicKey := s.keyrings.OffchainKeyring.OffchainPublicKey()
	configPublicKey := s.keyrings.OffchainKeyring.ConfigEncryptionPublicKey()
	return &pb.PublicKeysReply{
		OffchainPublicKey:         offchainPublicKey[:],
		ConfigEncryptionPublicKey: configPublicKey[:],
		OnchainPublicKey:          s.keyrings.OnchainKeyring.PublicKey(),
		MaxSignatureLength:        uint32(s.keyrings.OnchainKeyring.MaxSignatureLength()), //nolint:gosec // signature lengths are small
	}, nil
}

func (s *server) OffchainSign(_ context.Context, req *pb.OffchainSignRequest) (*pb.SignatureReply, error) {
	sig, err := s.keyrings.OffchainKeyring.OffchainSign(req.Message)
	if err != nil {
		return nil, err
	}

	return &pb.SignatureReply{Signature: sig}, nil
}

func (s *server) ConfigDiffieHellman(_ context.Context, req *pb.ConfigDiffieHellmanRequest) (*pb.ConfigDiffieHellmanReply, error) {
	var point [32]byte
	if len(req.Point) != len(point) {
		return nil, status.Errorf(codes.InvalidArgument, "point is %d bytes, want %d", len(req.Point), len(point))
	}

	copy(point[:], req.Point)
	secret, err := s.keyrings.OffchainKeyring.ConfigDiffieHellman(point)
	if err != nil {
		return nil, err
	}

	return &pb.ConfigDiffieHellmanReply{SharedSecret: secret[:]}, nil
}

func (s *server) Sign(_ context.Context, req *pb.SignRequest) (*pb.SignatureReply, error) {
	digest, err := ocrtypes.BytesToConfigDigest(req.ConfigDigest)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config digest: %v", err)
	}

	sig, err := s.keyrings.OnchainKeyring.Sign(digest, req.SeqNr, reportFromPB(req.Report))
	if err != nil {
		return nil, err
	}

	return &pb.SignatureReply{Signature: sig}, nil
}

func reportFromPB(r *pb.Report) ocr3types.ReportWithInfo[string] {
	return ocr3types.ReportWithInfo[string]{Report: r.GetReport(), Info: r.GetKeyBundle()}
}
