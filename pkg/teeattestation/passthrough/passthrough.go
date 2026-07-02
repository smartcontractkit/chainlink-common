// Package passthrough provides a TEE attestation validator that accepts any
// attestation without verification, recording each use via a metric so the
// insecure path stays observable. INSECURE; intended only for local and test
// environments where fake (non-Nitro) enclaves are trusted.
package passthrough

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// Validator accepts any attestation without verification. It implements the same
// validate methods as the Nitro validator so it can be swapped in wherever a
// validator is expected, and counts every call so the insecure passthrough path
// can be alerted on if ever enabled outside test environments.
type Validator struct {
	validations metric.Int64Counter
}

// New returns a passthrough Validator.
func New() (*Validator, error) {
	validations, err := beholder.GetMeter().Int64Counter("teeattestation_passthrough_validation_count")
	if err != nil {
		return nil, fmt.Errorf("failed to register passthrough validation counter: %w", err)
	}
	return &Validator{validations: validations}, nil
}

// ValidateAttestation accepts any attestation, recording the use.
func (v *Validator) ValidateAttestation(_, _, _ []byte) error {
	v.validations.Add(context.Background(), 1)
	return nil
}

// ValidateAttestationWithRoots accepts any attestation, recording the use.
func (v *Validator) ValidateAttestationWithRoots(_, _, _ []byte, _ string) error {
	v.validations.Add(context.Background(), 1)
	return nil
}
