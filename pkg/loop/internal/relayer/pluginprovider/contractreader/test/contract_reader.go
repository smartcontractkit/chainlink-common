package chaincomponentstest

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/assert"

	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var (
	// ContractReader is a static implementation of [types.ContractReader], [testtypes.Evaluator] and [types.PluginProvider
	// it is used for testing the [types.PluginProvider] interface
	ContractReader = staticContractReader{
		address:        "0x24",
		contractName:   "anyContract",
		contractMethod: "anyMethod",
		latestValue:    map[string]any{"ret1": "latestValue1", "ret2": "latestValue2"},
		params:         map[string]any{"param1": "value1", "param2": "value2"},
	}
)

// staticContractReader is a static implementation of ContractReaderTester
type staticContractReader struct {
	address        string
	contractName   string
	contractMethod string
	latestValue    map[string]any
	params         map[string]any
}

var _ testtypes.Evaluator[types.ContractReader] = staticContractReader{}
var _ types.ContractReader = staticContractReader{}

func (c staticContractReader) Start(_ context.Context) error { return nil }

func (c staticContractReader) Close() error { return nil }

func (c staticContractReader) Ready() error { panic("unimplemented") }

func (c staticContractReader) Name() string { panic("unimplemented") }

func (c staticContractReader) HealthReport() map[string]error { panic("unimplemented") }

func (c staticContractReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (c staticContractReader) Unbind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (c staticContractReader) GetLatestValue(_ context.Context, readName string, _ primitives.ConfidenceLevel, params, returnVal any) error {
	comp := types.BoundContract{
		Address: c.address,
		Name:    c.contractName,
	}.ReadIdentifier(c.contractMethod)

	if !assert.ObjectsAreEqual(readName, comp) {
		return fmt.Errorf("%w: expected report context %v but got %v", types.ErrInvalidType, comp, readName)
	}

	//gotParams, ok := params.(*map[string]string)
	gotParams, ok := params.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Invalid parameter type received in GetLatestValue. Expected %T but received %T", types.ErrInvalidEncoding, c.params, params)
	}
	if (*gotParams)["param1"] != c.params["param1"] || (*gotParams)["param2"] != c.params["param2"] {
		return fmt.Errorf("%w: Wrong params value received in GetLatestValue. Expected %v but received %v", types.ErrInvalidEncoding, c.params, *gotParams)
	}

	ret, ok := returnVal.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Wrong type passed for retVal param to GetLatestValue impl (expected %T instead of %T", types.ErrInvalidType, c.latestValue, returnVal)
	}

	(*ret)["ret1"] = c.latestValue["ret1"]
	(*ret)["ret2"] = c.latestValue["ret2"]

	return nil
}

func (c staticContractReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, nil
}

func (c staticContractReader) QueryKey(_ context.Context, _ types.BoundContract, _ query.KeyFilter, _ query.LimitAndSort, _ any) ([]types.Sequence, error) {
	return nil, nil
}

func (c staticContractReader) Evaluate(ctx context.Context, cr types.ContractReader) error {
	gotLatestValue := make(map[string]any)

	if err := cr.GetLatestValue(
		ctx,
		types.BoundContract{
			Address: c.address,
			Name:    c.contractName,
		}.ReadIdentifier(c.contractMethod),
		primitives.Unconfirmed,
		&c.params,
		&gotLatestValue,
	); err != nil {
		return fmt.Errorf("failed to call GetLatestValue(): %w", err)
	}

	if !assert.ObjectsAreEqual(gotLatestValue, c.latestValue) {
		return fmt.Errorf("GetLatestValue: expected %v but got %v", c.latestValue, gotLatestValue)
	}

	return nil
}
