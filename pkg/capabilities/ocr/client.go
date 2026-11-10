package ocr

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/ocr/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Connect returns the Node served by [RegisterServer] on conn, for the node whose peer is peerID. Its onchain keyring
// signs on the node, but checks signatures with verify: verifying needs no secret, so the process does it itself.
func Connect(ctx context.Context, lggr logger.Logger, conn grpc.ClientConnInterface, peerID string, cfg Config,
	verify func(ocrtypes.OnchainPublicKey, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[string], []byte) bool) (Node, error) {
	version, err := resolveVersion(cfg.Version)
	if err != nil {
		return Node{}, err
	}

	switch version {
	case VersionV1:
		return connectV1(ctx, logger.Named(lggr, "OCRClient"), pb.NewOCRClient(conn), peerID, verify)
	default:
		return Node{}, fmt.Errorf("unknown OCR version %d", version)
	}
}

func resolveVersion(v Version) (Version, error) {
	if v == VersionLatest {
		return latestVersion, nil
	}

	if v < VersionLatest || v > latestVersion {
		return 0, fmt.Errorf("unknown OCR version %d", v)
	}

	return v, nil
}

func connectV1(ctx context.Context, lggr logger.Logger, client pb.OCRClient, peerID string,
	verify func(ocrtypes.OnchainPublicKey, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[string], []byte) bool) (Node, error) {
	keys, err := client.PublicKeys(ctx, &emptypb.Empty{})
	if err != nil {
		return Node{}, fmt.Errorf("failed to read the node's OCR public keys: %w", err)
	}

	k := &keyring{
		client:             client,
		verify:             verify,
		onchainPublicKey:   keys.OnchainPublicKey,
		maxSignatureLength: int(keys.MaxSignatureLength),
	}
	copy(k.offchainPublicKey[:], keys.OffchainPublicKey)
	copy(k.configEncryptionPublicKey[:], keys.ConfigEncryptionPublicKey)

	return Node{
		BinaryNetworkEndpointFactory: &endpointFactory{lggr: lggr, client: client, peerID: peerID},
		OffchainKeyring:              k,
		OnchainKeyring:               k,
	}, nil
}

var (
	_ ocrtypes.OffchainKeyring         = (*keyring)(nil)
	_ ocr3types.OnchainKeyring[string] = (*keyring)(nil)
)

// keyring signs with the node's keys. Its public keys cannot change, so they are read once.
type keyring struct {
	client                    pb.OCRClient
	verify                    func(ocrtypes.OnchainPublicKey, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[string], []byte) bool
	offchainPublicKey         ocrtypes.OffchainPublicKey
	configEncryptionPublicKey ocrtypes.ConfigEncryptionPublicKey
	onchainPublicKey          ocrtypes.OnchainPublicKey
	maxSignatureLength        int
}

func (k *keyring) OffchainSign(msg []byte) ([]byte, error) {
	reply, err := k.client.OffchainSign(context.Background(), &pb.OffchainSignRequest{Message: msg})
	if err != nil {
		return nil, fmt.Errorf("failed to sign offchain: %w", err)
	}

	return reply.Signature, nil
}

func (k *keyring) ConfigDiffieHellman(point [32]byte) ([32]byte, error) {
	var secret [32]byte
	reply, err := k.client.ConfigDiffieHellman(context.Background(), &pb.ConfigDiffieHellmanRequest{Point: point[:]})
	if err != nil {
		return secret, fmt.Errorf("failed to compute the config shared secret: %w", err)
	}

	copy(secret[:], reply.SharedSecret)
	return secret, nil
}

func (k *keyring) OffchainPublicKey() ocrtypes.OffchainPublicKey { return k.offchainPublicKey }

func (k *keyring) ConfigEncryptionPublicKey() ocrtypes.ConfigEncryptionPublicKey {
	return k.configEncryptionPublicKey
}

func (k *keyring) PublicKey() ocrtypes.OnchainPublicKey { return k.onchainPublicKey }

func (k *keyring) Sign(digest ocrtypes.ConfigDigest, seqNr uint64, report ocr3types.ReportWithInfo[string]) ([]byte, error) {
	reply, err := k.client.Sign(context.Background(), &pb.SignRequest{
		ConfigDigest: digest[:],
		SeqNr:        seqNr,
		Report:       &pb.Report{Report: report.Report, KeyBundle: report.Info},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign report: %w", err)
	}

	return reply.Signature, nil
}

func (k *keyring) Verify(publicKey ocrtypes.OnchainPublicKey, digest ocrtypes.ConfigDigest, seqNr uint64,
	report ocr3types.ReportWithInfo[string], signature []byte) bool {
	return k.verify(publicKey, digest, seqNr, report, signature)
}

func (k *keyring) MaxSignatureLength() int { return k.maxSignatureLength }
