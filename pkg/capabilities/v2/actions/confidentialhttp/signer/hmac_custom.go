package signer

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// hmacTemplateBudget bounds how long a user-supplied canonical template may
// spend executing. Prevents pathological templates (huge repeated loops) from
// stalling signing.
const hmacTemplateBudget = 30 * time.Millisecond

type hmacCustomSigner struct {
	secret          *confhttppb.SecretIdentifier
	tpl             *template.Template
	hashFactory     func() hash.Hash
	signatureHeader string
	signaturePrefix string
	timestampHeader string
	nonceHeader     string
	encoding        string
	nowFn           func() time.Time
	nonceFn         func() (string, error)
}

func newHmacCustomSigner(cfg *confhttppb.HmacCustom) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("hmac_custom config is nil")
	}
	if cfg.GetSecret() == nil {
		return nil, errors.New("hmac_custom: secret is required")
	}
	if cfg.GetCanonicalTemplate() == "" {
		return nil, errors.New("hmac_custom: canonical_template is required")
	}
	if cfg.GetSignatureHeader() == "" {
		return nil, errors.New("hmac_custom: signature_header is required")
	}

	var factory func() hash.Hash
	switch cfg.GetHash() {
	case confhttppb.HmacCustom_HASH_SHA256:
		factory = sha256.New
	case confhttppb.HmacCustom_HASH_SHA512:
		factory = sha512.New
	default:
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedHashAlgorithm, cfg.GetHash())
	}

	tpl, err := template.New("hmac_canonical").
		Funcs(template.FuncMap{
			"header": func(_ string) string {
				// Placeholder — overridden per-request so the template can
				// observe mutable req.Header values. Parsing requires the
				// func be defined at parse time.
				return ""
			},
		}).
		Parse(cfg.GetCanonicalTemplate())
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrHmacTemplateParse, err)
	}

	return &hmacCustomSigner{
		secret:          cfg.GetSecret(),
		tpl:             tpl,
		hashFactory:     factory,
		signatureHeader: cfg.GetSignatureHeader(),
		signaturePrefix: cfg.GetSignaturePrefix(),
		timestampHeader: cfg.GetTimestampHeader(),
		nonceHeader:     cfg.GetNonceHeader(),
		encoding:        cfg.GetEncoding(),
		nowFn:           time.Now,
		nonceFn:         defaultNonce,
	}, nil
}

func (s *hmacCustomSigner) Sign(ctx context.Context, req *http.Request, secrets map[string]string) error {
	secret, err := resolveSecretID(secrets, s.secret)
	if err != nil {
		return err
	}

	body, err := readBodyForHashing(req)
	if err != nil {
		return err
	}

	ts := strconv.FormatInt(s.nowFn().Unix(), 10)
	var nonce string
	if s.nonceHeader != "" {
		nonce, err = s.nonceFn()
		if err != nil {
			return err
		}
	}

	data := map[string]any{
		"method":      req.Method,
		"url":         req.URL.String(),
		"path":        req.URL.Path,
		"query":       req.URL.RawQuery,
		"body":        string(body),
		"body_sha256": sha256Hex(body),
		"timestamp":   ts,
		"nonce":       nonce,
	}

	// Re-clone the template to bind a request-scoped `header` func that reads
	// actual req.Header values.
	tpl, err := s.tpl.Clone()
	if err != nil {
		return fmt.Errorf("%w: clone: %v", ErrHmacTemplateExec, err)
	}
	tpl = tpl.Funcs(template.FuncMap{
		"header": func(name string) string { return req.Header.Get(name) },
	})

	canonical, err := executeTemplateWithBudget(ctx, tpl, data)
	if err != nil {
		return err
	}

	mac := hmac.New(s.hashFactory, []byte(secret))
	mac.Write([]byte(canonical))
	sig, err := encodeMAC(mac.Sum(nil), s.encoding)
	if err != nil {
		return err
	}

	if s.timestampHeader != "" {
		req.Header.Set(s.timestampHeader, ts)
	}
	if s.nonceHeader != "" {
		req.Header.Set(s.nonceHeader, nonce)
	}
	req.Header.Set(s.signatureHeader, s.signaturePrefix+sig)
	return nil
}

// executeTemplateWithBudget runs tpl.Execute with a time budget. If the
// budget is exceeded we return ErrHmacTemplateTimeout. The template itself
// is not interruptible, but this bounds the *detectable* stall — a truly
// runaway template would still pin the goroutine; in practice this is
// extremely unlikely for text/template.
func executeTemplateWithBudget(ctx context.Context, tpl *template.Template, data any) (string, error) {
	var (
		out strings.Builder
		wg  sync.WaitGroup
		err error
	)
	ctx, cancel := context.WithTimeout(ctx, hmacTemplateBudget)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = tpl.Execute(&out, data)
	}()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return "", ErrHmacTemplateTimeout
		}
		return "", ctx.Err()
	case <-done:
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrHmacTemplateExec, err)
		}
		return out.String(), nil
	}
}

func defaultNonce() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
