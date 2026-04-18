package signer

import (
	"context"
	"errors"
	"net/http"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type apiKeySigner struct {
	headerName  string
	secretName  string
	valuePrefix string
}

func newAPIKeySigner(cfg *confhttppb.ApiKeyAuth) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("api_key auth config is nil")
	}
	if cfg.GetHeaderName() == "" {
		return nil, errors.New("api_key: header_name is required")
	}
	if cfg.GetSecretName() == "" {
		return nil, errors.New("api_key: secret_name is required")
	}
	return &apiKeySigner{
		headerName:  cfg.GetHeaderName(),
		secretName:  cfg.GetSecretName(),
		valuePrefix: cfg.GetValuePrefix(),
	}, nil
}

func (s *apiKeySigner) Sign(_ context.Context, req *http.Request, secrets map[string]string) error {
	val, err := resolveSecret(secrets, s.secretName)
	if err != nil {
		return err
	}
	req.Header.Set(s.headerName, s.valuePrefix+val)
	return nil
}
