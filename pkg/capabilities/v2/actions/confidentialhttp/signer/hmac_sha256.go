package signer

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net/http"
	"strconv"
	"time"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

type hmacSha256Signer struct {
	secret          *confhttppb.SecretIdentifier
	signatureHeader string
	timestampHeader string
	includeQuery    bool
	encoding        string
	nowFn           func() time.Time
}

func newHmacSha256Signer(cfg *confhttppb.HmacSha256) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("hmac_sha256 config is nil")
	}
	if cfg.GetSecret() == nil {
		return nil, errors.New("hmac_sha256: secret is required")
	}
	sh := cfg.GetSignatureHeader()
	if sh == "" {
		sh = "X-Signature"
	}
	th := cfg.GetTimestampHeader()
	if th == "" {
		th = "X-Timestamp"
	}
	return &hmacSha256Signer{
		secret:          cfg.GetSecret(),
		signatureHeader: sh,
		timestampHeader: th,
		includeQuery:    cfg.GetIncludeQuery(),
		encoding:        cfg.GetEncoding(),
		nowFn:           time.Now,
	}, nil
}

func (s *hmacSha256Signer) Sign(_ context.Context, req *http.Request, secrets map[string]string) error {
	secret, err := resolveSecretID(secrets, s.secret)
	if err != nil {
		return err
	}
	body, err := readBodyForHashing(req)
	if err != nil {
		return err
	}
	url := req.URL.Path
	if s.includeQuery && req.URL.RawQuery != "" {
		url = url + "?" + req.URL.RawQuery
	}
	ts := strconv.FormatInt(s.nowFn().Unix(), 10)
	canonical := req.Method + "\n" + url + "\n" + sha256Hex(body) + "\n" + ts

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical))
	sig, err := encodeMAC(mac.Sum(nil), s.encoding)
	if err != nil {
		return err
	}
	req.Header.Set(s.timestampHeader, ts)
	req.Header.Set(s.signatureHeader, sig)
	return nil
}
