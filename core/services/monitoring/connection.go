package monitoring

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"net/url"

	"github.com/smartcontractkit/wsrpc"
)

type WSRPCConnection interface {
	wsrpc.ClientInterface
	Close()
}

func NewConnection(
	serverURL *url.URL,
	clientPrivateKeyHex string,
	serverPublicKeyHex string,
) (WSRPCConnection, error) {
	clientPrivateKey := make([]byte, hex.DecodedLen(len(clientPrivateKeyHex)))
	if _, err := hex.Decode(clientPrivateKey, []byte(clientPrivateKeyHex)); err != nil {
		return nil, err
	}
	serverPublicKey := make([]byte, hex.DecodedLen(len(serverPublicKeyHex)))
	if _, err := hex.Decode(serverPublicKey, []byte(serverPublicKeyHex)); err != nil {
		return nil, err
	}
	return wsrpc.DialWithContext(
		context.TODO(),
		serverURL.String(),
		wsrpc.WithTransportCreds(
			ed25519.PrivateKey(clientPrivateKey),
			ed25519.PublicKey(serverPublicKey),
		),
	)
}
