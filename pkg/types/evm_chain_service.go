package types

import (
	"context"
	"fmt"
)

type UnimplementedEVMChainService struct{}

var _ EVMChainService = UnimplementedEVMChainService{}

func (u UnimplementedEVMChainService) Start(_ context.Context) error {
	return fmt.Errorf("start not implemented")
}

func (u UnimplementedEVMChainService) Close() error {
	return fmt.Errorf("close not implemented")
}

func (u UnimplementedEVMChainService) Ready() error {
	return fmt.Errorf("ready not implemented")
}

func (u UnimplementedEVMChainService) HealthReport() map[string]error {
	return map[string]error{}
}

func (u UnimplementedEVMChainService) Name() string {
	return "not implemented"
}

func (u UnimplementedEVMChainService) ReadContract(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte{}, fmt.Errorf("not implemented")
}
