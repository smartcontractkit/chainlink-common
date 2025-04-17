package types

import (
	"context"
	"fmt"
)

type UnimplementedEVMChain struct{}

var _ EVMChain = UnimplementedEVMChain{}

func (u UnimplementedEVMChain) Start(_ context.Context) error {
	return fmt.Errorf("start not implemented")
}

func (u UnimplementedEVMChain) Close() error {
	return fmt.Errorf("close not implemented")
}

func (u UnimplementedEVMChain) Ready() error {
	return fmt.Errorf("ready not implemented")
}

func (u UnimplementedEVMChain) HealthReport() map[string]error {
	return map[string]error{}
}

func (u UnimplementedEVMChain) Name() string {
	return "not implemented"
}

func (u UnimplementedEVMChain) ReadContract(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte{}, fmt.Errorf("not implemented")
}
