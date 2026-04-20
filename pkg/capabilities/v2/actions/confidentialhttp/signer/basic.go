package signer

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type basicSigner struct {
	usernameSecretName string
	passwordSecretName string
}

func newBasicSigner(cfg *confhttppb.BasicAuth) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("basic auth config is nil")
	}
	if cfg.GetUsernameSecretName() == "" || cfg.GetPasswordSecretName() == "" {
		return nil, errors.New("basic: username_secret_name and password_secret_name are required")
	}
	return &basicSigner{
		usernameSecretName: cfg.GetUsernameSecretName(),
		passwordSecretName: cfg.GetPasswordSecretName(),
	}, nil
}

func (s *basicSigner) Sign(_ context.Context, req *http.Request, secrets map[string]string) error {
	u, err := resolveSecret(secrets, s.usernameSecretName)
	if err != nil {
		return err
	}
	p, err := resolveSecret(secrets, s.passwordSecretName)
	if err != nil {
		return err
	}
	creds := base64.StdEncoding.EncodeToString([]byte(u + ":" + p))
	req.Header.Set("Authorization", "Basic "+creds)
	return nil
}
