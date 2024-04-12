package chainreader_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
)

var errorTypes = []error{
	types.ErrInvalidEncoding,
	types.ErrInvalidType,
	types.ErrFieldNotFound,
	types.ErrSliceWrongLen,
	types.ErrNotASlice,
	types.ErrNotFound,
}

type cannotEncode struct{}

func (*cannotEncode) MarshalCBOR() ([]byte, error) {
	return nil, errors.New("nope")
}

func (*cannotEncode) UnmarshalCBOR([]byte) error {
	return errors.New("nope")
}

func (*cannotEncode) MarshalText() ([]byte, error) {
	return nil, errors.New("nope")
}

func (*cannotEncode) UnmarshalText() error {
	return errors.New("nope")
}

type interfaceTesterBase struct{}

var anyAccountBytes = []byte{1, 2, 3}

func (it *interfaceTesterBase) GetAccountBytes(_ int) []byte {
	return anyAccountBytes
}

func (it *interfaceTesterBase) Name() string {
	return "relay client"
}

type fakeTypeProvider struct{}

func (f fakeTypeProvider) CreateType(key string, isEncode bool) (any, error) {
	return f.CreateContractType(key, isEncode)
}

var _ types.ContractTypeProvider = (*fakeTypeProvider)(nil)

func (fakeTypeProvider) CreateContractType(key string, isEncode bool) (any, error) {
	tokens := strings.Split(key, ".")
	if len(tokens) < 2 {
		return nil, fmt.Errorf("key should be in form of contractName.type, got %s instead", key)
	}

	switch tokens[1] {
	case NilType:
		return &struct{}{}, nil
	case TestItemType:
		return &TestStruct{}, nil
	case TestItemSliceType:
		return &[]TestStruct{}, nil
	case TestItemArray2Type:
		return &[2]TestStruct{}, nil
	case TestItemArray1Type:
		return &[1]TestStruct{}, nil
	case MethodTakingLatestParamsReturningTestStruct:
		if isEncode {
			return &LatestParams{}, nil
		}
		return &TestStruct{}, nil
	case MethodReturningUint64, DifferentMethodReturningUint64:
		tmp := uint64(0)
		return &tmp, nil
	case MethodReturningUint64Slice:
		var tmp []uint64
		return &tmp, nil
	case MethodReturningSeenStruct, TestItemWithConfigExtra:
		if isEncode {
			return &TestStruct{}, nil
		}
		return &TestStructWithExtraField{}, nil
	case EventName, EventWithFilterName:
		if isEncode {
			return &FilterEventParams{}, nil
		}
		return &TestStruct{}, nil
	}

	return nil, types.ErrInvalidType
}

func generateQueryFilterTestCases(t *testing.T) []query.KeyFilter {
	var queryFilters []query.KeyFilter
	confirmationsValues := []query.ConfirmationLevel{query.Finalized, query.Unconfirmed}
	operatorValues := []query.ComparisonOperator{query.Eq, query.Neq, query.Gt, query.Lt, query.Gte, query.Lte}
	comparableValues := []string{"", " ", "number", "123"}

	primitives := []query.Expression{query.TxHash("txHash")}
	for _, op := range operatorValues {
		primitives = append(primitives, query.Block(123, op))
		primitives = append(primitives, query.Timestamp(123, op))

		var valueComparers []query.ValueComparer
		for _, comparableValue := range comparableValues {
			valueComparers = append(valueComparers, query.ValueComparer{
				Value:    comparableValue,
				Operator: op,
			})
		}
		primitives = append(primitives, query.Comparer("someName", valueComparers...))
	}

	for _, conf := range confirmationsValues {
		primitives = append(primitives, query.Confirmation(conf))
	}

	qf, err := query.Where("primitives", primitives...)
	require.NoError(t, err)
	queryFilters = append(queryFilters, qf)

	andOverPrimitivesBoolExpr := query.And(primitives...)
	orOverPrimitivesBoolExpr := query.Or(primitives...)

	nestedBoolExpr := query.And(
		query.TxHash("txHash"),
		andOverPrimitivesBoolExpr,
		orOverPrimitivesBoolExpr,
		query.TxHash("txHash"),
	)
	require.NoError(t, err)

	qf, err = query.Where("andOverPrimitivesBoolExpr", andOverPrimitivesBoolExpr)
	require.NoError(t, err)
	queryFilters = append(queryFilters, qf)

	qf, err = query.Where("orOverPrimitivesBoolExpr", orOverPrimitivesBoolExpr)
	require.NoError(t, err)
	queryFilters = append(queryFilters, qf)

	qf, err = query.Where("nestedBoolExpr", nestedBoolExpr)
	require.NoError(t, err)
	queryFilters = append(queryFilters, qf)

	return queryFilters
}
