package signer

import (
	"context"
	"errors"
	"net/http"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type bearerSigner struct {
	token       *confhttppb.SecretIdentifier
	headerName  string
	valuePrefix string
}

func newBearerSigner(cfg *confhttppb.BearerAuth) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("bearer auth config is nil")
	}
	if cfg.GetToken() == nil {
		return nil, errors.New("bearer: token is required")
	}
	h := cfg.GetHeaderName()
	if h == "" {
		h = "Authorization"
	}
	p := cfg.GetValuePrefix()
	if p == "" {
		p = "Bearer "
	}
	return &bearerSigner{
		token:       cfg.GetToken(),
		headerName:  h,
		valuePrefix: p,
	}, nil
}

func (s *bearerSigner) Sign(_ context.Context, req *http.Request, secrets map[string]string) error {
	tok, err := resolveSecretID(secrets, s.token)
	if err != nil {
		return err
	}
	req.Header.Set(s.headerName, s.valuePrefix+tok)
	return nil
}
