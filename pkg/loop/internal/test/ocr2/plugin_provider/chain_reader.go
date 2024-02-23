package pluginprovider_test

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/stretchr/testify/assert"
)

type ChainReaderEvaluator interface {
	types.ChainReader
	// Evaluate runs all the methods of the other chain reader and checks if they return
	// the values of the embedded chain reader.
	Evaluate(ctx context.Context, other types.ChainReader) error
}

var (
	contractName         = "anyContract"
	contractMethod       = "anyMethod"
	getLatestValueParams = map[string]string{"param1": "value1", "param2": "value2"}
	latestValue          = map[string]any{"ret1": "latestValue1", "ret2": "latestValue2"}

	// TestChainReader is a static implementation of ChainReaderEvaluator for testing
	TestChainReader = staticChainReader{
		contractName:   contractName,
		contractMethod: contractMethod,
		latestValue:    latestValue,
		params:         getLatestValueParams,
	}
)

// staticChainReader is a static implementation of ChainReaderEvaluator
type staticChainReader struct {
	contractName   string
	contractMethod string
	latestValue    map[string]any
	params         map[string]string
}

var _ ChainReaderEvaluator = staticChainReader{}

func (c staticChainReader) Bind(context.Context, []types.BoundContract) error {
	// lazy initialization
	c.contractName = contractName
	c.contractMethod = contractMethod
	c.latestValue = latestValue
	c.params = getLatestValueParams
	return nil
}

func (c staticChainReader) GetLatestValue(ctx context.Context, cn, method string, params, returnVal any) error {
	if !assert.ObjectsAreEqual(cn, c.contractName) {
		return fmt.Errorf("%w: expected report context %v but got %v", types.ErrInvalidType, c.contractName, cn)
	}
	if method != c.contractMethod {
		return fmt.Errorf("%w: expected generic contract method %v but got %v", types.ErrInvalidType, c.contractMethod, method)
	}
	gotParams, ok := params.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Invalid parameter type received in GetLatestValue. Expected %T but received %T", types.ErrInvalidEncoding, gotParams, params)
	}
	if (*gotParams)["param1"] != c.params["param1"] || (*gotParams)["param2"] != c.params["param2"] {
		return fmt.Errorf("%w: Wrong params value received in GetLatestValue. Expected %v but received %v", types.ErrInvalidEncoding, getLatestValueParams, *gotParams)
	}

	ret, ok := returnVal.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Wrong type passed for retVal param to GetLatestValue impl (expected %T instead of %T", types.ErrInvalidType, ret, returnVal)
	}

	(*ret)["ret1"] = c.latestValue["ret1"]
	(*ret)["ret2"] = c.latestValue["ret2"]

	return nil
}

func (c staticChainReader) Evaluate(ctx context.Context, cr types.ChainReader) error {
	var gotLatestValue map[string]any
	err := cr.GetLatestValue(ctx, c.contractName, c.contractMethod, c.params, &gotLatestValue)
	if err != nil {
		return fmt.Errorf("failed to call GetLatestValue(): %w", err)
	}

	if !assert.ObjectsAreEqual(gotLatestValue, c.latestValue) {
		return fmt.Errorf("GetLatestValue: expected %v but got %v", c.latestValue, gotLatestValue)
	}
	return nil
}
