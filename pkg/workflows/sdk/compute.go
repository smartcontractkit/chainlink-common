package sdk

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

//go:generate go run ./gen

type ComputeOutput[T any] struct {
	Value T
}

type ComputeOutputCap[T any] interface {
	CapDefinition[ComputeOutput[T]]
	Value() CapDefinition[T]
}

type computeOutputCap[T any] struct {
	CapDefinition[ComputeOutput[T]]
}

func (c *computeOutputCap[T]) Value() CapDefinition[T] {
	return AccessField[ComputeOutput[T], T](c.CapDefinition, "Value")
}

var _ ComputeOutputCap[struct{}] = &computeOutputCap[struct{}]{}

type ComputeConfig[C any] struct {
	Config C
}

func (c *ComputeConfig[C]) ToMap() (map[string]any, error) {
	var m map[string]any
	switch cm := any(c.Config).(type) {
	case map[string]any:
		m = cm
	default:
		wc, err := values.WrapMap(c.Config)
		if err != nil {
			return nil, err
		}

		uc, err := wc.Unwrap()
		if err != nil {
			return nil, err
		}

		tm, ok := uc.(map[string]any)
		if !ok {
			return nil, errors.New("could not convert config into map")
		}

		m = tm
	}

	if _, ok := m["config"]; ok {
		return nil, errors.New("`config` is a reserved keyword inside Compute config")
	}
	m["config"] = "$(ENV.config)"

	if _, ok := m["binary"]; ok {
		return nil, errors.New("`binary` is a reserved keyword inside Compute config")
	}
	m["binary"] = "$(ENV.binary)"

	return m, nil
}

func EmptyComputeConfig() *ComputeConfig[struct{}] {
	return &ComputeConfig[struct{}]{Config: struct{}{}}
}
