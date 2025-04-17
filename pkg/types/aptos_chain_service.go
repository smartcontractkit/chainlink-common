package types

import (
	"context"
	"fmt"
)

type UnimplementedAptosChainService struct{}

var _ AptosChainService = UnimplementedAptosChainService{}

func (u UnimplementedAptosChainService) Start(_ context.Context) error {
	return fmt.Errorf("start not implemented")
}

func (u UnimplementedAptosChainService) Close() error {
	return fmt.Errorf("close not implemented")
}

func (u UnimplementedAptosChainService) Ready() error {
	return fmt.Errorf("ready not implemented")
}

func (u UnimplementedAptosChainService) HealthReport() map[string]error {
	return map[string]error{}
}

func (u UnimplementedAptosChainService) Name() string {
	return "not implemented"
}

func (u UnimplementedAptosChainService) ReadContract(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return []byte{}, fmt.Errorf("not implemented")
}
