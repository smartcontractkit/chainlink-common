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
// // https://pkg.go.dev/github.com/aws/aws-sdk-go@v1.55.7/service/kms
type Client interface {
	GetPublicKey(input *kmslib.GetPublicKeyInput) (*kmslib.GetPublicKeyOutput, error)
	Sign(input *kmslib.SignInput) (*kmslib.SignOutput, error)
	DescribeKey(input *kmslib.DescribeKeyInput) (*kmslib.DescribeKeyOutput, error)
	ListKeys(input *kmslib.ListKeysInput) (*kmslib.ListKeysOutput, error)
}

// NewClient constructs a new kmslib.KMS instance using the provided AWS profile. This adheres to
// the KMSClient interface, allowing for signing and public key retrieval using AWS KMS.
// The region is automatically read from the AWS profile configuration.
func NewClient(awsProfile string) (Client, error) {
	if awsProfile == "" {
		return nil, errors.New("AWSProfile is required")
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           awsProfile,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}
	return kmslib.New(sess), nil
}

// NewClientWithDefaultCredentials constructs a new kmslib.KMS instance using the default AWS
// credential chain. This is suitable for use in Kubernetes with IRSA (IAM Roles for Service Accounts),
// EC2 instance profiles, or environment variables.
func NewClientWithDefaultCredentials(region string) (Client, error) {
	if region == "" {
		return nil, errors.New("region is required")
	}
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}
	return kmslib.New(sess), nil
}
