package signer

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	confhttppb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
)

// unsignedPayload is the magic payload-hash value recognized by services
// that support streaming uploads without hashing the body (notably S3).
const unsignedPayload = "UNSIGNED-PAYLOAD"

type awsSigV4Signer struct {
	accessKeyID     *confhttppb.StringOrSecret
	secretAccessKey *confhttppb.SecretIdentifier
	sessionToken    *confhttppb.SecretIdentifier // optional
	region          string
	service         string
	unsignedPayload bool
	signer          *v4.Signer
	nowFn           func() time.Time
}

func newAwsSigV4Signer(cfg *confhttppb.AwsSigV4) (Signer, error) {
	if cfg == nil {
		return nil, errors.New("aws_sig_v4 config is nil")
	}
	if cfg.GetAccessKeyId() == nil || cfg.GetSecretAccessKey() == nil {
		return nil, errors.New("aws_sig_v4: access_key_id and secret_access_key are required")
	}
	if cfg.GetRegion() == "" {
		return nil, errors.New("aws_sig_v4: region is required")
	}
	if cfg.GetService() == "" {
		return nil, errors.New("aws_sig_v4: service is required")
	}
	return &awsSigV4Signer{
		accessKeyID:     cfg.GetAccessKeyId(),
		secretAccessKey: cfg.GetSecretAccessKey(),
		sessionToken:    cfg.GetSessionToken(),
		region:          cfg.GetRegion(),
		service:         cfg.GetService(),
		unsignedPayload: cfg.GetUnsignedPayload(),
		signer:          v4.NewSigner(),
		nowFn:           time.Now,
	}, nil
}

func (s *awsSigV4Signer) Sign(ctx context.Context, req *http.Request, secrets map[string]string) error {
	ak, err := resolveStringOrSecret(secrets, s.accessKeyID)
	if err != nil {
		return err
	}
	sk, err := resolveSecretID(secrets, s.secretAccessKey)
	if err != nil {
		return err
	}
	creds := aws.Credentials{
		AccessKeyID:     ak,
		SecretAccessKey: sk,
	}
	if s.sessionToken != nil {
		st, err := resolveSecretID(secrets, s.sessionToken)
		if err != nil {
			return err
		}
		creds.SessionToken = st
	}

	var payloadHash string
	if s.unsignedPayload {
		payloadHash = unsignedPayload
		// S3 and other services that accept UNSIGNED-PAYLOAD require the
		// caller to explicitly advertise that choice via
		// X-Amz-Content-Sha256 — the signer does not set it for us.
		req.Header.Set("X-Amz-Content-Sha256", unsignedPayload)
	} else {
		body, err := readBodyForHashing(req)
		if err != nil {
			return fmt.Errorf("%w: read body: %v", ErrSigV4Sign, err)
		}
		payloadHash = sha256Hex(body)
	}

	if err := s.signer.SignHTTP(ctx, creds, req, payloadHash, s.service, s.region, s.nowFn()); err != nil {
		return fmt.Errorf("%w: %v", ErrSigV4Sign, err)
	}
	return nil
}
