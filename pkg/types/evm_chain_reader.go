package types

import (
	"context"
	"fmt"
)

type UnimplementedEVMChainReader struct{}

var _ EVMChainReader = UnimplementedEVMChainReader{}

func (u UnimplementedEVMChainReader) ReadContract(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte{}, fmt.Errorf("not implemented")
}
