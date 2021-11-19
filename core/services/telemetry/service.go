package telemetry

import (
	context "context"
	"crypto/ed25519"
	"net/url"

	"github.com/smartcontractkit/chainlink/core/logger"
	wsrpc "github.com/smartcontractkit/wsrpc"
)

type Service struct {
	ctx              context.Context
	cancelCtx        context.CancelFunc
	serverURL        *url.URL
	clientPrivateKey ed25519.PrivateKey
	serverPublicKey  ed25519.PublicKey
	log              *logger.Logger
}

func NewService(
	serverURL *url.URL,
	clientPrivateKey ed25519.PrivateKey,
	serverPublicKey ed25519.PublicKey,
	log *logger.Logger,
) Service {
	ctx, cancelFunc := context.WithCancel(context.TODO())
	return Service{
		ctx,
		cancelFunc,
		serverURL,
		clientPrivateKey,
		serverPublicKey,
		log,
	}
}

func (s Service) Start() (Client, error) {
	conn, err := wsrpc.DialWithContext(
		s.ctx,
		s.serverURL.String(),
		wsrpc.WithTransportCreds(
			s.clientPrivateKey,
			s.serverPublicKey,
		),
	)
	if err != nil {
		return Client{}, err
	}
	client := NewClient(s.ctx, NewTelemetryClient(conn), 100, s.log)
	return client, nil
}

func (s Service) Stop() {
	s.cancelCtx()
}
