package gcpkms

import (
	"context"
	"fmt"

	apiv1 "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/option"
)

// Client is an interface that defines the operations needed by the keystore. It keeps the keystore
// independent of the generated Google Cloud KMS client.
//
// Every method operates on a single CryptoKeyVersion, so each can be authorized with per-key IAM
// bindings; a deployment binds exactly the keys it configures and nothing else.
//
// These methods are based on the Google Cloud KMS Go client interface.
// https://pkg.go.dev/cloud.google.com/go/kms/apiv1
type Client interface {
	GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error)
	GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error)
	AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error)
}

// NewClient constructs a new Google Cloud KMS client using the Go SDK.
//
// Credentials always come from Application Default Credentials, which covers both production (GKE
// Workload Identity, GCE/Cloud Run service accounts) and local development (`gcloud auth
// application-default login`, or GOOGLE_APPLICATION_CREDENTIALS pointing at a service-account key file).
//
// opts is passed through to the SDK for the cases ADC does not cover a custom endpoint or
// emulator, a quota project, a non-default token source.
// https://cloud.google.com/docs/authentication/application-default-credentials
func NewClient(ctx context.Context, opts ...option.ClientOption) (*SDKClient, error) {
	client, err := apiv1.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud KMS client: %w", err)
	}
	return &SDKClient{client: client}, nil
}

// SDKClient adapts the generated Cloud KMS client to this package's [Client] interface and owns
// the underlying transport, so callers must Close it.
type SDKClient struct {
	client *apiv1.KeyManagementClient
}

var _ Client = (*SDKClient)(nil)

func (c *SDKClient) GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	return c.client.GetCryptoKeyVersion(ctx, req, opts...)
}

func (c *SDKClient) GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error) {
	return c.client.GetPublicKey(ctx, req, opts...)
}

func (c *SDKClient) AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error) {
	return c.client.AsymmetricSign(ctx, req, opts...)
}

func (c *SDKClient) Close() error {
	return c.client.Close()
}
