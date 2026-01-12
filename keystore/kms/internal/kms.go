package kms

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	kmslib "github.com/aws/aws-sdk-go/service/kms"
)

// Client is an interface that defines the methods for interacting with AWS KMS. We only expose
// the methods that are needed for our use case, which is to get a public key and sign data.
//
// These methods are directly copied from the kms.Client interface in the AWS SDK for Go v1.
type Client interface {
	// Duck Typed from:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.55.7/service/kms#KMS.GetPublicKey
	GetPublicKey(input *kmslib.GetPublicKeyInput) (*kmslib.GetPublicKeyOutput, error)
	// Duck Typed from:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.55.7/service/kms#KMS.Sign
	Sign(input *kmslib.SignInput) (*kmslib.SignOutput, error)
	// Duck Typed from:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.55.7/service/kms#KMS.DescribeKey
	DescribeKey(input *kmslib.DescribeKeyInput) (*kmslib.DescribeKeyOutput, error)
}

// ClientConfig holds the configuration for the AWS KMS client.
type ClientConfig struct {
	// Required: KeyRegion is the AWS region where the KMS key is located.
	KeyRegion string
	// Optional: AWSProfile is the name of the AWS profile to use for authentication.
	// If not provided, environment variables will be used to determine the AWS profile.
	AWSProfile string
}

// validate checks if the ClientConfig has the required fields set.
func (c ClientConfig) validate() error {
	if c.KeyRegion == "" {
		return errors.New("KMS key region is required")
	}

	return nil
}

// NewClient constructs a new kmslib.KMS instance using the provided configuration. This adheres to
// the KMSClient interface, allowing for signing and public key retrieval using AWS KMS.
func NewClient(config ClientConfig) (Client, error) {
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid KMS config: %w", err)
	}

	// Create a new AWS session using the provided region and profile name if specified. Defaults
	// to using environment variables.
	session := sessionFromEnvVars(config.KeyRegion)
	if config.AWSProfile != "" {
		session = sessionFromProfile(config.KeyRegion, config.AWSProfile)
	}

	return kmslib.New(session), nil
}

// sessionFromEnvVars creates a new AWS session using environment variables to load the profile.
func sessionFromEnvVars(region string) *session.Session {
	return session.Must(
		session.NewSession(&aws.Config{
			Region:                        aws.String(region),
			CredentialsChainVerboseErrors: aws.Bool(true),
		}),
	)
}

// sessionFromProfile creates a new AWS session using a specific profile name and region.
func sessionFromProfile(region string, profile string) *session.Session {
	return session.Must(
		session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           profile,
			Config: aws.Config{
				Region:                        aws.String(region),
				CredentialsChainVerboseErrors: aws.Bool(true),
			},
		}),
	)
}
