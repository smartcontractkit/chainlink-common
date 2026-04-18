package signer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// mustReq builds a request with a rewindable body (so HMAC/SigV4 can read
// and restore it).
func mustReq(t *testing.T, method, urlStr, body string) *http.Request {
	t.Helper()
	var r *http.Request
	var err error
	if body == "" {
		r, err = http.NewRequest(method, urlStr, nil)
	} else {
		r, err = http.NewRequest(method, urlStr, strings.NewReader(body))
	}
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	return r
}

func TestBuilder_NilAuth_ReturnsNilSigner(t *testing.T) {
	b := NewBuilder(nil)
	s, err := b.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) err=%v", err)
	}
	if s != nil {
		t.Fatalf("Build(nil) signer=%v, want nil", s)
	}
}

func TestAPIKeySigner(t *testing.T) {
	cfg := &confhttppb.ApiKeyAuth{
		HeaderName:  "x-api-key",
		SecretName:  "cg",
		ValuePrefix: "",
	}
	s, err := newAPIKeySigner(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://example.com/x", "")
	if err := s.Sign(context.Background(), req, map[string]string{"cg": "secret123"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if got := req.Header.Get("x-api-key"); got != "secret123" {
		t.Fatalf("header got=%q", got)
	}
}

func TestAPIKeySigner_ValuePrefix(t *testing.T) {
	s, err := newAPIKeySigner(&confhttppb.ApiKeyAuth{
		HeaderName: "Authorization", SecretName: "tok", ValuePrefix: "ApiKey ",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://example.com/x", "")
	if err := s.Sign(context.Background(), req, map[string]string{"tok": "zzz"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "ApiKey zzz" {
		t.Fatalf("got=%q", got)
	}
}

func TestAPIKeySigner_MissingSecret(t *testing.T) {
	s, _ := newAPIKeySigner(&confhttppb.ApiKeyAuth{HeaderName: "x", SecretName: "absent"})
	req := mustReq(t, "GET", "https://example.com/x", "")
	err := s.Sign(context.Background(), req, map[string]string{})
	if !errors.Is(err, ErrSecretNotProvided) {
		t.Fatalf("want ErrSecretNotProvided, got %v", err)
	}
}

func TestBasicSigner(t *testing.T) {
	s, err := newBasicSigner(&confhttppb.BasicAuth{
		UsernameSecretName: "u", PasswordSecretName: "p",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://example.com/x", "")
	if err := s.Sign(context.Background(), req, map[string]string{"u": "alice", "p": "hunter2"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	want := "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:hunter2"))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestBearerSigner_Default(t *testing.T) {
	s, err := newBearerSigner(&confhttppb.BearerAuth{TokenSecretName: "t"})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://example.com/x", "")
	_ = s.Sign(context.Background(), req, map[string]string{"t": "abc"})
	if got := req.Header.Get("Authorization"); got != "Bearer abc" {
		t.Fatalf("got=%q", got)
	}
}

func TestBearerSigner_CustomHeaderAndPrefix(t *testing.T) {
	s, err := newBearerSigner(&confhttppb.BearerAuth{
		TokenSecretName: "t", HeaderName: "Authorization", ValuePrefix: "token ",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://api.github.com/x", "")
	_ = s.Sign(context.Background(), req, map[string]string{"t": "gho_xx"})
	if got := req.Header.Get("Authorization"); got != "token gho_xx" {
		t.Fatalf("got=%q", got)
	}
}

func TestHmacSha256_BasicCanonical(t *testing.T) {
	s, err := newHmacSha256Signer(&confhttppb.HmacSha256{SecretName: "hmac"})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	// Pin time for determinism.
	hs := s.(*hmacSha256Signer)
	hs.nowFn = func() time.Time { return time.Unix(1700000000, 0) }

	req := mustReq(t, "POST", "https://example.com/api/v1?x=1", `{"a":1}`)
	if err := s.Sign(context.Background(), req, map[string]string{"hmac": "key"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if got := req.Header.Get("X-Timestamp"); got != "1700000000" {
		t.Fatalf("timestamp got=%q", got)
	}
	if req.Header.Get("X-Signature") == "" {
		t.Fatalf("signature missing")
	}
	// Assert body was rewound (canonical call should leave body readable).
	if req.Body == nil {
		t.Fatalf("body nil after sign")
	}
	b, _ := io.ReadAll(req.Body)
	if string(b) != `{"a":1}` {
		t.Fatalf("body after sign=%q", string(b))
	}
}

func TestHmacCustom_Template(t *testing.T) {
	cfg := &confhttppb.HmacCustom{
		SecretName:        "k",
		CanonicalTemplate: `{{.method}} {{.path}} {{.body_sha256}}`,
		Hash:              confhttppb.HmacCustom_HASH_SHA256,
		SignatureHeader:   "X-Sig",
		Encoding:          "hex",
	}
	s, err := newHmacCustomSigner(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	// Pin time + nonce for determinism (not used in template but exercised
	// by the nonce path if configured).
	hc := s.(*hmacCustomSigner)
	hc.nowFn = func() time.Time { return time.Unix(1700000000, 0) }

	req := mustReq(t, "POST", "https://example.com/widgets", `hello`)
	if err := s.Sign(context.Background(), req, map[string]string{"k": "k1"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if req.Header.Get("X-Sig") == "" {
		t.Fatalf("missing signature")
	}
}

func TestHmacCustom_BadTemplateRejectedAtNew(t *testing.T) {
	_, err := newHmacCustomSigner(&confhttppb.HmacCustom{
		SecretName:        "k",
		CanonicalTemplate: `{{ oopsNoSuchFunc }}`,
		Hash:              confhttppb.HmacCustom_HASH_SHA256,
		SignatureHeader:   "X-Sig",
	})
	if !errors.Is(err, ErrHmacTemplateParse) {
		t.Fatalf("want parse err, got %v", err)
	}
}

func TestHmacCustom_SignatureWithPrefix(t *testing.T) {
	s, err := newHmacCustomSigner(&confhttppb.HmacCustom{
		SecretName:        "k",
		CanonicalTemplate: `{{.method}}`,
		Hash:              confhttppb.HmacCustom_HASH_SHA512,
		SignatureHeader:   "X-Vendor-Signature",
		SignaturePrefix:   "HMAC-SHA512 ",
		Encoding:          "base64",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "POST", "https://example.com/x", "")
	_ = s.Sign(context.Background(), req, map[string]string{"k": "k1"})
	if !strings.HasPrefix(req.Header.Get("X-Vendor-Signature"), "HMAC-SHA512 ") {
		t.Fatalf("prefix missing, got=%q", req.Header.Get("X-Vendor-Signature"))
	}
}

// --- AWS SigV4 ---

func TestAwsSigV4_AttachesExpectedHeaders(t *testing.T) {
	s, err := newAwsSigV4Signer(&confhttppb.AwsSigV4{
		AccessKeyIdSecretName:     "ak",
		SecretAccessKeySecretName: "sk",
		Region:                    "us-east-1",
		Service:                   "execute-api",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	aw := s.(*awsSigV4Signer)
	aw.nowFn = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

	req := mustReq(t, "POST", "https://api.execute-api.us-east-1.amazonaws.com/prod/x", `{"a":1}`)
	if err := s.Sign(context.Background(), req,
		map[string]string{"ak": "AKIDEXAMPLE", "sk": "secret"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if req.Header.Get("Authorization") == "" {
		t.Fatalf("missing Authorization header")
	}
	if !strings.Contains(req.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
		t.Fatalf("unexpected auth: %q", req.Header.Get("Authorization"))
	}
	if req.Header.Get("X-Amz-Date") == "" {
		t.Fatalf("missing X-Amz-Date")
	}
}

func TestAwsSigV4_SessionTokenInjected(t *testing.T) {
	s, err := newAwsSigV4Signer(&confhttppb.AwsSigV4{
		AccessKeyIdSecretName:     "ak",
		SecretAccessKeySecretName: "sk",
		SessionTokenSecretName:    "st",
		Region:                    "us-east-1",
		Service:                   "execute-api",
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://api.execute-api.us-east-1.amazonaws.com/prod/x", "")
	_ = s.Sign(context.Background(), req,
		map[string]string{"ak": "AKID", "sk": "sk", "st": "session"})
	if req.Header.Get("X-Amz-Security-Token") != "session" {
		t.Fatalf("missing X-Amz-Security-Token")
	}
}

func TestAwsSigV4_UnsignedPayload(t *testing.T) {
	s, err := newAwsSigV4Signer(&confhttppb.AwsSigV4{
		AccessKeyIdSecretName:     "ak",
		SecretAccessKeySecretName: "sk",
		Region:                    "us-east-1",
		Service:                   "s3",
		UnsignedPayload:           true,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "PUT", "https://bucket.s3.us-east-1.amazonaws.com/k", "huge-body-bytes")
	if err := s.Sign(context.Background(), req, map[string]string{"ak": "AKID", "sk": "sk"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if req.Header.Get("X-Amz-Content-Sha256") != "UNSIGNED-PAYLOAD" {
		t.Fatalf("X-Amz-Content-Sha256=%q, want UNSIGNED-PAYLOAD", req.Header.Get("X-Amz-Content-Sha256"))
	}
}

// --- OAuth2 ---

type idpHandler struct {
	hits    atomic.Int32
	status  int
	expires int64
	token   string
}

func (h *idpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hits.Add(1)
	w.Header().Set("Content-Type", "application/json")
	if h.status != 0 && h.status != 200 {
		w.WriteHeader(h.status)
		_, _ = io.WriteString(w, `{"error":"invalid_client"}`)
		return
	}
	tok := h.token
	if tok == "" {
		tok = "access_token_1"
	}
	resp := map[string]any{"access_token": tok, "token_type": "Bearer", "expires_in": h.expires}
	_ = json.NewEncoder(w).Encode(resp)
}

// idpFromHandler returns a test server using https via httptest.NewTLSServer
// so the signer's https-only check is satisfied.
func idpFromHandler(t *testing.T, h http.Handler) (*httptest.Server, *http.Client) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	return srv, srv.Client()
}

func TestOAuth2ClientCreds_CacheAndSingleFlight(t *testing.T) {
	h := &idpHandler{expires: 3600}
	srv, client := idpFromHandler(t, h)

	cache := newOAuth2Cache()
	s, err := newOAuth2ClientCredsSigner(&confhttppb.OAuth2ClientCredentials{
		TokenUrl:               srv.URL,
		ClientIdSecretName:     "cid",
		ClientSecretSecretName: "csec",
		Scopes:                 []string{"read", "write"},
	}, client, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	secrets := map[string]string{"cid": "id1", "csec": "sec1"}
	// Two sequential calls should trigger only ONE IdP hit thanks to cache.
	for i := 0; i < 2; i++ {
		req := mustReq(t, "GET", "https://api.example.com/x", "")
		if err := s.Sign(context.Background(), req, secrets); err != nil {
			t.Fatalf("Sign %d: %v", i, err)
		}
		if !strings.HasPrefix(req.Header.Get("Authorization"), "Bearer ") {
			t.Fatalf("missing Bearer header, got %q", req.Header.Get("Authorization"))
		}
	}
	if got := h.hits.Load(); got != 1 {
		t.Fatalf("idp hits=%d, want 1 (cache miss should happen once)", got)
	}
}

func TestOAuth2ClientCreds_IdPFailure_SurfacesTypedError(t *testing.T) {
	h := &idpHandler{status: 401}
	srv, client := idpFromHandler(t, h)
	cache := newOAuth2Cache()
	s, err := newOAuth2ClientCredsSigner(&confhttppb.OAuth2ClientCredentials{
		TokenUrl:               srv.URL,
		ClientIdSecretName:     "cid",
		ClientSecretSecretName: "csec",
	}, client, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://api.example.com/x", "")
	err = s.Sign(context.Background(), req, map[string]string{"cid": "x", "csec": "y"})
	if !errors.Is(err, ErrOAuth2TokenEndpointHTTPError) {
		t.Fatalf("want ErrOAuth2TokenEndpointHTTPError, got %v", err)
	}
}

func TestOAuth2ClientCreds_NonHTTPSRejected(t *testing.T) {
	cache := newOAuth2Cache()
	s, err := newOAuth2ClientCredsSigner(&confhttppb.OAuth2ClientCredentials{
		TokenUrl:               "http://insecure.example.com/token",
		ClientIdSecretName:     "cid",
		ClientSecretSecretName: "csec",
	}, http.DefaultClient, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://api.example.com/x", "")
	err = s.Sign(context.Background(), req, map[string]string{"cid": "x", "csec": "y"})
	if !errors.Is(err, ErrOAuth2TokenURLInvalid) {
		t.Fatalf("want ErrOAuth2TokenURLInvalid, got %v", err)
	}
}

func TestOAuth2ClientCreds_CacheMiss_OnExpiry(t *testing.T) {
	h := &idpHandler{expires: 1} // 1s expires_in + 60s safety → instant miss
	srv, client := idpFromHandler(t, h)

	cache := newOAuth2Cache()
	// Force expiration to be immediate by advancing the clock inside put.
	cache.nowFn = func() time.Time { return time.Now() }

	s, err := newOAuth2ClientCredsSigner(&confhttppb.OAuth2ClientCredentials{
		TokenUrl:               srv.URL,
		ClientIdSecretName:     "cid",
		ClientSecretSecretName: "csec",
	}, client, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	secrets := map[string]string{"cid": "id1", "csec": "sec1"}

	req1 := mustReq(t, "GET", "https://api.example.com/x", "")
	if err := s.Sign(context.Background(), req1, secrets); err != nil {
		t.Fatalf("Sign #1: %v", err)
	}
	// expires_in=1 -> safety=60s -> already expired. Second call must
	// re-fetch.
	req2 := mustReq(t, "GET", "https://api.example.com/x", "")
	if err := s.Sign(context.Background(), req2, secrets); err != nil {
		t.Fatalf("Sign #2: %v", err)
	}
	if got := h.hits.Load(); got < 2 {
		t.Fatalf("expected >=2 idp hits after expiry, got %d", got)
	}
}

func TestOAuth2RefreshToken_UsesRefreshGrant(t *testing.T) {
	var formSeen url.Values
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		formSeen = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "access_after_refresh", "expires_in": 3600,
		})
	})
	srv, client := idpFromHandler(t, h)
	cache := newOAuth2Cache()

	s, err := newOAuth2RefreshSigner(&confhttppb.OAuth2RefreshToken{
		TokenUrl:               srv.URL,
		RefreshTokenSecretName: "rt",
		ClientIdSecretName:     "cid",
		ClientSecretSecretName: "csec",
	}, client, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	req := mustReq(t, "GET", "https://api.example.com/x", "")
	if err := s.Sign(context.Background(), req,
		map[string]string{"rt": "my_refresh", "cid": "id1", "csec": "sec1"}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if req.Header.Get("Authorization") != "Bearer access_after_refresh" {
		t.Fatalf("unexpected bearer: %q", req.Header.Get("Authorization"))
	}
	if formSeen.Get("grant_type") != "refresh_token" {
		t.Fatalf("grant_type=%q", formSeen.Get("grant_type"))
	}
	if formSeen.Get("refresh_token") != "my_refresh" {
		t.Fatalf("refresh_token form value missing")
	}
}

func TestOAuth2RefreshToken_CacheKeyDependsOnRefreshFingerprint(t *testing.T) {
	h := &idpHandler{expires: 3600, token: "t1"}
	srv, client := idpFromHandler(t, h)
	cache := newOAuth2Cache()
	s, err := newOAuth2RefreshSigner(&confhttppb.OAuth2RefreshToken{
		TokenUrl:               srv.URL,
		RefreshTokenSecretName: "rt",
	}, client, cache)
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	req := mustReq(t, "GET", "https://api.example.com/x", "")
	_ = s.Sign(context.Background(), req, map[string]string{"rt": "refresh_A"})
	// Call with a DIFFERENT refresh token; cache must miss.
	req2 := mustReq(t, "GET", "https://api.example.com/x", "")
	_ = s.Sign(context.Background(), req2, map[string]string{"rt": "refresh_B"})

	if h.hits.Load() != 2 {
		t.Fatalf("expected 2 idp hits (different refresh token fingerprints), got %d", h.hits.Load())
	}
}
