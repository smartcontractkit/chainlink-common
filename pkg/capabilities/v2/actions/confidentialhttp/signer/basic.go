package signer

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type basicSigner struct {
	username *confhttppb.StringOrSecret
	password *confhttppb.SecretIdentifier
}

func newBasicSigner(cfg *confhttppb.BasicAuth) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("basic auth config is nil")
	}
	if cfg.GetUsername() == nil || cfg.GetPassword() == nil {
		return nil, errors.New("basic: username and password are required")
	}
	return &basicSigner{
		username: cfg.GetUsername(),
		password: cfg.GetPassword(),
	}, nil
}

func (s *basicSigner) Sign(_ context.Context, req *http.Request, secrets map[string]string) error {
	u, err := resolveStringOrSecret(secrets, s.username)
	if err != nil {
		return err
	}
	p, err := resolveSecretID(secrets, s.password)
	if err != nil {
		return err
	}
	creds := base64.StdEncoding.EncodeToString([]byte(u + ":" + p))
	req.Header.Set("Authorization", "Basic "+creds)
	return nil
}
