package signer

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type oauth2ClientCredsSigner struct {
	tokenURL         string
	clientID         *confhttppb.StringOrSecret
	clientSecret     *confhttppb.SecretIdentifier
	scopes           []string
	audience         string
	clientAuthMethod string
	extraParams      map[string]string
	httpClient       *http.Client
	cache            *oauth2Cache
}

func newOAuth2ClientCredsSigner(cfg *confhttppb.OAuth2ClientCredentials, httpClient *http.Client, cache *oauth2Cache) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("oauth2 client_credentials config is nil")
	}
	if cfg.GetTokenUrl() == "" {
		return nil, errors.New("oauth2 client_credentials: token_url is required")
	}
	if cfg.GetClientId() == nil || cfg.GetClientSecret() == nil {
		return nil, errors.New("oauth2 client_credentials: client_id and client_secret are required")
	}
	method := cfg.GetClientAuthMethod()
	if method == "" {
		method = clientAuthMethodBasic
	}
	if method != clientAuthMethodBasic && method != clientAuthMethodBody {
		return nil, errors.New("oauth2 client_credentials: client_auth_method must be 'basic_auth' or 'request_body'")
	}
	return &oauth2ClientCredsSigner{
		tokenURL:         cfg.GetTokenUrl(),
		clientID:         cfg.GetClientId(),
		clientSecret:     cfg.GetClientSecret(),
		scopes:           cfg.GetScopes(),
		audience:         cfg.GetAudience(),
		clientAuthMethod: method,
		extraParams:      cfg.GetExtraParams(),
		httpClient:       httpClient,
		cache:            cache,
	}, nil
}

func (s *oauth2ClientCredsSigner) Sign(ctx context.Context, req *http.Request, secrets map[string]string) error {
	cid, err := resolveStringOrSecret(secrets, s.clientID)
	if err != nil {
		return err
	}
	csec, err := resolveSecretID(secrets, s.clientSecret)
	if err != nil {
		return err
	}

	key := cacheKey(
		"oauth2.client_credentials",
		s.tokenURL,
		cid,
		sortedJoin(s.scopes, " "),
		s.audience,
		s.clientAuthMethod,
	)

	tok, err := s.cache.fetchOrFetch(key, func() (string, int64, error) {
		form := url.Values{}
		form.Set("grant_type", "client_credentials")
		if len(s.scopes) > 0 {
			form.Set("scope", strings.Join(s.scopes, " "))
		}
		if s.audience != "" {
			form.Set("audience", s.audience)
		}
		for k, v := range s.extraParams {
			form.Set(k, v)
		}

		var basic *struct{ user, pass string }
		if s.clientAuthMethod == clientAuthMethodBasic {
			basic = &struct{ user, pass string }{cid, csec}
		} else {
			form.Set("client_id", cid)
			form.Set("client_secret", csec)
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
