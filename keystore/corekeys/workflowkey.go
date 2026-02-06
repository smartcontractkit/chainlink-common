// `corekeys` provides utilities to generate keys that are compatible with the core node
// and can be imported by it.
package corekeys

import (
	"context"
)

const TypeWorkflowKey = "workflow"

func (ks *Store) GenerateEncryptedWorkflowKey(ctx context.Context, password string) ([]byte, error) {
	return ks.generateEncryptedKey(ctx, TypeWorkflowKey, password)
}

func FromEncryptedWorkflowKey(data []byte, password string) ([]byte, error) {
	return fromEncryptedKey(data, TypeWorkflowKey, password)
}
