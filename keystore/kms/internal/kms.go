package kms

import (
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
	// Duck Typed from:
	// https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.55.7/service/kms#KMS.ListKeys
	ListKeys(input *kmslib.ListKeysInput) (*kmslib.ListKeysOutput, error)
}

// ClientConfig holds the configuration for the AWS KMS client.
type ClientConfig struct {
	// Optional: AWSProfile is the name of the AWS profile to use for authentication.
	// If not provided, environment variables will be used to determine the AWS profile.
	AWSProfile string
}

// NewClient constructs a new kmslib.KMS instance using the provided configuration. This adheres to
// the KMSClient interface, allowing for signing and public key retrieval using AWS KMS.
func NewClient(config ClientConfig) (Client, error) {
	var sess *session.Session
	if config.AWSProfile != "" {
		// Use profile - region will come from profile config or can be overridden
		opts := session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Profile:           config.AWSProfile,
			Config: aws.Config{
				CredentialsChainVerboseErrors: aws.Bool(true),
			},
		}
		sess = session.Must(session.NewSessionWithOptions(opts))
	} else {
		// Use environment variables
		cfg := &aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
		}
		sess = session.Must(session.NewSession(cfg))
	}

	return kmslib.New(sess), nil
}
