package gcpkms

import (
	"context"
	"errors"
	"fmt"

	apiv1 "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
	"github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Client is an interface that defines the operations needed by the keystore. It keeps the keystore
// independent of the generated Google Cloud KMS client.
//
// Every method here can be authorized per CryptoKey, so a deployment can bind exactly the keys it
// configures and nothing else. Listing a whole key ring is deliberately not part of this interface —
// see [KeyRingLister].
//
// These methods are based on the Google Cloud KMS Go client interface.
// https://pkg.go.dev/cloud.google.com/go/kms/apiv1
type Client interface {
	GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error)
	GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error)
	AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error)
	ListCryptoKeyVersions(ctx context.Context, cryptoKeyName string) ([]*kmspb.CryptoKeyVersion, error)
}

// KeyRingLister is an optional capability: a Client that can also enumerate a key ring. GetKeys only
// needs it when called with no key allowlist.
//
// It is kept out of [Client] on purpose. ListCryptoKeys takes the key ring as its parent, so
// cloudkms.cryptoKeys.list can only be bound at the key ring or above — strictly broader than every
// other permission the keystore needs, all of which bind to individual CryptoKeys. A least-privilege
// deployment that names its keys explicitly should never be forced to implement, or be granted, a
// ring-wide list.
type KeyRingLister interface {
	ListCryptoKeys(ctx context.Context, keyRingName string) ([]*kmspb.CryptoKey, error)
}

// ClientWithClose is the client returned by NewClient: the full Cloud KMS surface, including key
// ring listing, plus the underlying transport lifecycle. Whether listing actually succeeds is a
// matter of the credentials' IAM bindings, not of the Go type.
type ClientWithClose interface {
	Client
	KeyRingLister
	Close() error
}

// NewClient constructs a new Google Cloud KMS client using the Go SDK.
//
// Credentials always come from Application Default Credentials, which covers both production (GKE
// Workload Identity, GCE/Cloud Run service accounts) and local development (`gcloud auth
// application-default login`, or GOOGLE_APPLICATION_CREDENTIALS pointing at a service-account key
// file). There is deliberately no credentials-file option: ADC already reads that env var, so a
// dedicated field would only duplicate it while hard-coding a long-lived key file into config.
//
// opts is passed through to the SDK for the cases ADC does not cover — a custom endpoint or
// emulator, a quota project, a non-default token source.
// https://cloud.google.com/docs/authentication/application-default-credentials
func NewClient(ctx context.Context, opts ...option.ClientOption) (ClientWithClose, error) {
	client, err := apiv1.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Cloud KMS client: %w", err)
	}
	return &clientAdapter{client: client}, nil
}

type clientAdapter struct {
	client *apiv1.KeyManagementClient
}

func (c *clientAdapter) GetCryptoKeyVersion(ctx context.Context, req *kmspb.GetCryptoKeyVersionRequest, opts ...gax.CallOption) (*kmspb.CryptoKeyVersion, error) {
	return c.client.GetCryptoKeyVersion(ctx, req, opts...)
}

func (c *clientAdapter) GetPublicKey(ctx context.Context, req *kmspb.GetPublicKeyRequest, opts ...gax.CallOption) (*kmspb.PublicKey, error) {
	return c.client.GetPublicKey(ctx, req, opts...)
}

func (c *clientAdapter) AsymmetricSign(ctx context.Context, req *kmspb.AsymmetricSignRequest, opts ...gax.CallOption) (*kmspb.AsymmetricSignResponse, error) {
	return c.client.AsymmetricSign(ctx, req, opts...)
}

func (c *clientAdapter) ListCryptoKeys(ctx context.Context, keyRingName string) ([]*kmspb.CryptoKey, error) {
	iter := c.client.ListCryptoKeys(ctx, &kmspb.ListCryptoKeysRequest{Parent: keyRingName})
	return drain(iter.Next, "crypto keys in "+keyRingName)
}

func (c *clientAdapter) ListCryptoKeyVersions(ctx context.Context, cryptoKeyName string) ([]*kmspb.CryptoKeyVersion, error) {
	iter := c.client.ListCryptoKeyVersions(ctx, &kmspb.ListCryptoKeyVersionsRequest{Parent: cryptoKeyName})
	return drain(iter.Next, "crypto key versions of "+cryptoKeyName)
}

// drain reads a Cloud KMS iterator to completion. what describes the listed resources and is only
// used to build the error message.
func drain[T any](next func() (T, error), what string) ([]T, error) {
	items := make([]T, 0)
	for {
		item, err := next()
		if errors.Is(err, iterator.Done) {
			return items, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list %s: %w", what, err)
		}
		items = append(items, item)
	}
}

func (c *clientAdapter) Close() error {
	return c.client.Close()
}
