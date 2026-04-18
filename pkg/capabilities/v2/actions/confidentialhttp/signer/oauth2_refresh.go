package signer

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type oauth2RefreshSigner struct {
	tokenURL               string
	refreshTokenSecretName string
	clientIDSecretName     string // optional
	clientSecretSecretName string // optional
	scopes                 []string
	extraParams            map[string]string
	httpClient             *http.Client
	cache                  *oauth2Cache
}

func newOAuth2RefreshSigner(cfg *confhttppb.OAuth2RefreshToken, httpClient *http.Client, cache *oauth2Cache) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("oauth2 refresh_token config is nil")
	}
	if cfg.GetTokenUrl() == "" {
		return nil, errors.New("oauth2 refresh_token: token_url is required")
	}
	if cfg.GetRefreshTokenSecretName() == "" {
		return nil, errors.New("oauth2 refresh_token: refresh_token_secret_name is required")
	}
	return &oauth2RefreshSigner{
		tokenURL:               cfg.GetTokenUrl(),
		refreshTokenSecretName: cfg.GetRefreshTokenSecretName(),
		clientIDSecretName:     cfg.GetClientIdSecretName(),
		clientSecretSecretName: cfg.GetClientSecretSecretName(),
		scopes:                 cfg.GetScopes(),
		extraParams:            cfg.GetExtraParams(),
		httpClient:             httpClient,
		cache:                  cache,
	}, nil
}

func (s *oauth2RefreshSigner) Sign(ctx context.Context, req *http.Request, secrets map[string]string) error {
	rt, err := resolveSecret(secrets, s.refreshTokenSecretName)
	if err != nil {
		return err
	}

	var cid, csec string
	if s.clientIDSecretName != "" {
		cid, err = resolveSecret(secrets, s.clientIDSecretName)
		if err != nil {
			return err
		}
	}
	if s.clientSecretSecretName != "" {
		csec, err = resolveSecret(secrets, s.clientSecretSecretName)
		if err != nil {
			return err
		}
	}

	key := cacheKey(
		"oauth2.refresh_token",
		s.tokenURL,
		cid,
		sortedJoin(s.scopes, " "),
		fingerprint(rt),
	)

	tok, err := s.cache.fetchOrFetch(key, func() (string, int64, error) {
		form := url.Values{}
		form.Set("grant_type", "refresh_token")
		form.Set("refresh_token", rt)
		if len(s.scopes) > 0 {
			form.Set("scope", strings.Join(s.scopes, " "))
		}
		for k, v := range s.extraParams {
			form.Set(k, v)
		}

		// When both client_id AND client_secret are present, we send them
		// via HTTP Basic Auth (standard). When only client_id is set (PKCE-
		// style refresh), include it in the form body.
		var basic *struct{ user, pass string }
		switch {
		case cid != "" && csec != "":
			basic = &struct{ user, pass string }{cid, csec}
		case cid != "":
			form.Set("client_id", cid)
		}

		tr, err := postTokenRequest(ctx, s.httpClient, s.tokenURL, form, basic)
		if err != nil {
			return "", 0, err
		}
		return tr.AccessToken, tr.ExpiresIn, nil
	})
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+tok)
	return nil
}
