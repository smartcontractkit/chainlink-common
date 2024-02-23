package median_test

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/stretchr/testify/assert"
)

type StaticChainReader struct{}

func (c StaticChainReader) Bind(context.Context, []types.BoundContract) error {
	return nil
}

func (c StaticChainReader) GetLatestValue(ctx context.Context, cn, method string, params, returnVal any) error {
	if !assert.ObjectsAreEqual(cn, contractName) {
		return fmt.Errorf("%w: expected report context %v but got %v", types.ErrInvalidType, contractName, cn)
	}
	if method != medianContractGenericMethod {
		return fmt.Errorf("%w: expected generic contract method %v but got %v", types.ErrInvalidType, medianContractGenericMethod, method)
	}
	gotParams, ok := params.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Invalid parameter type received in GetLatestValue. Expected %T but received %T", types.ErrInvalidEncoding, gotParams, params)
	}
	if (*gotParams)["param1"] != getLatestValueParams["param1"] || (*gotParams)["param2"] != getLatestValueParams["param2"] {
		return fmt.Errorf("%w: Wrong params value received in GetLatestValue. Expected %v but received %v", types.ErrInvalidEncoding, getLatestValueParams, *gotParams)
	}

	ret, ok := returnVal.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Wrong type passed for retVal param to GetLatestValue impl (expected %T instead of %T", types.ErrInvalidType, ret, returnVal)
	}

	(*ret)["ret1"] = latestValue["ret1"]
	(*ret)["ret2"] = latestValue["ret2"]

	return nil
}
