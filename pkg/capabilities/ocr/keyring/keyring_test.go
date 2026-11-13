package keyring

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestConnect(t *testing.T) {
	t.Parallel()
	host := newFakeHost()
	conn := serve(t, host)

	keyrings, err := Connect(t.Context(), conn, Config{Version: VersionLatest}, noVerify)
	require.NoError(t, err)
	assert.Equal(t, host.offchainPublicKey, keyrings.OffchainKeyring.OffchainPublicKey())
	assert.Equal(t, host.configPublicKey, keyrings.OffchainKeyring.ConfigEncryptionPublicKey())
	assert.Equal(t, host.onchainPublicKey, keyrings.OnchainKeyring.PublicKey())
	assert.Equal(t, host.maxSignatureLength, keyrings.OnchainKeyring.MaxSignatureLength())

	_, err = Connect(t.Context(), conn, Config{Version: latestVersion + 1}, noVerify)
	require.ErrorContains(t, err, "unknown OCR keyring version")
}

func TestKeyring(t *testing.T) {
	t.Parallel()
	var verified ocr3types.ReportWithInfo[string]
	verify := func(_ ocrtypes.OnchainPublicKey, _ ocrtypes.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[string], _ []byte) bool {
		verified = r
		return true
	}

	keyrings, err := Connect(t.Context(), serve(t, newFakeHost()), Config{}, verify)
	require.NoError(t, err)

	digest := ocrtypes.ConfigDigest{1}
	report := ocr3types.ReportWithInfo[string]{Report: []byte("report"), Info: "evm"}

	sig, err := keyrings.OnchainKeyring.Sign(digest, 7, report)
	require.NoError(t, err)
	assert.Equal(t, []byte("onchain:report:evm"), sig)

	assert.True(t, keyrings.OnchainKeyring.Verify(nil, digest, 7, report, sig))
	assert.Equal(t, report, verified, "signatures are verified in the process")

	sig, err = keyrings.OffchainKeyring.OffchainSign([]byte("msg"))
	require.NoError(t, err)
	assert.Equal(t, []byte("offchain:msg"), sig)

	secret, err := keyrings.OffchainKeyring.ConfigDiffieHellman([32]byte{9})
	require.NoError(t, err)
	assert.Equal(t, [32]byte{10}, secret)
}

func noVerify(ocrtypes.OnchainPublicKey, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[string], []byte) bool {
	return false
}

func serve(t *testing.T, host *fakeHost) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer(grpc.MaxRecvMsgSize(MaxMessageSize))
	RegisterServer(s, Keyrings{OffchainKeyring: host, OnchainKeyring: host})
	go func() { _ = s.Serve(lis) }()
	t.Cleanup(s.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

// fakeHost is a host whose keys sign by labelling what they sign.
type fakeHost struct {
	offchainPublicKey  ocrtypes.OffchainPublicKey
	configPublicKey    ocrtypes.ConfigEncryptionPublicKey
	onchainPublicKey   ocrtypes.OnchainPublicKey
	maxSignatureLength int
}

func newFakeHost() *fakeHost {
	return &fakeHost{
		offchainPublicKey:  ocrtypes.OffchainPublicKey{1},
		configPublicKey:    ocrtypes.ConfigEncryptionPublicKey{2},
		onchainPublicKey:   ocrtypes.OnchainPublicKey("onchain-key"),
		maxSignatureLength: 65,
	}
}

func (h *fakeHost) OffchainSign(msg []byte) ([]byte, error) {
	return append([]byte("offchain:"), msg...), nil
}

func (h *fakeHost) ConfigDiffieHellman(point [32]byte) ([32]byte, error) {
	point[0]++
	return point, nil
}

func (h *fakeHost) OffchainPublicKey() ocrtypes.OffchainPublicKey { return h.offchainPublicKey }

func (h *fakeHost) ConfigEncryptionPublicKey() ocrtypes.ConfigEncryptionPublicKey {
	return h.configPublicKey
}

func (h *fakeHost) PublicKey() ocrtypes.OnchainPublicKey { return h.onchainPublicKey }

func (h *fakeHost) Sign(_ ocrtypes.ConfigDigest, _ uint64, r ocr3types.ReportWithInfo[string]) ([]byte, error) {
	return []byte("onchain:" + string(r.Report) + ":" + r.Info), nil
}

func (h *fakeHost) Verify(ocrtypes.OnchainPublicKey, ocrtypes.ConfigDigest, uint64, ocr3types.ReportWithInfo[string], []byte) bool {
	return false
}

func (h *fakeHost) MaxSignatureLength() int { return h.maxSignatureLength }
