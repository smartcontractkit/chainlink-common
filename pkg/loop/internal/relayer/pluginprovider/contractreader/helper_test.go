package contractreader_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests" //nolint
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

func (f fakeTypeProvider) CreateType(readName string, isEncode bool) (any, error) {
	return f.CreateContractType(readName, isEncode)
}

var _ types.ContractTypeProvider = (*fakeTypeProvider)(nil)

func (fakeTypeProvider) CreateContractType(readName string, isEncode bool) (any, error) {
	switch true {
	case strings.HasSuffix(readName, NilType):
		return &struct{}{}, nil
	case strings.HasSuffix(readName, TestItemType):
		return &TestStruct{}, nil
	case strings.HasSuffix(readName, TestItemSliceType):
		return &[]TestStruct{}, nil
	case strings.HasSuffix(readName, TestItemArray2Type):
		return &[2]TestStruct{}, nil
	case strings.HasSuffix(readName, TestItemArray1Type):
		return &[1]TestStruct{}, nil
	case strings.HasSuffix(readName, MethodTakingLatestParamsReturningTestStruct):
		if isEncode {
			return &LatestParams{}, nil
		}
		return &TestStruct{}, nil
	case strings.HasSuffix(readName, MethodReturningUint64), strings.HasSuffix(readName, MethodReturningAlterableUint64):
		tmp := uint64(0)
		return &tmp, nil
	case strings.HasSuffix(readName, MethodReturningUint64Slice):
		var tmp []uint64
		return &tmp, nil
	case strings.HasSuffix(readName, MethodReturningSeenStruct), strings.HasSuffix(readName, TestItemWithConfigExtra):
		if isEncode {
			return &TestStruct{}, nil
		}
		return &TestStructWithExtraField{}, nil
	case strings.HasSuffix(readName, EventName), strings.HasSuffix(readName, EventWithFilterName):
		if isEncode {
			return &FilterEventParams{}, nil
		}
		return &TestStruct{}, nil
	}

	return nil, types.ErrInvalidType
}

func generateQueryFilterTestCases(t *testing.T) []query.KeyFilter {
	var queryFilters []query.KeyFilter
	confirmationsValues := []primitives.ConfidenceLevel{primitives.Finalized, primitives.Unconfirmed}
	operatorValues := []primitives.ComparisonOperator{primitives.Eq, primitives.Neq, primitives.Gt, primitives.Lt, primitives.Gte, primitives.Lte}
	comparableValues := []string{"", " ", "number", "123"}

	primitiveExpressions := []query.Expression{query.TxHash("txHash")}
	for _, op := range operatorValues {
		primitiveExpressions = append(primitiveExpressions, query.Block("123", op))
		primitiveExpressions = append(primitiveExpressions, query.Timestamp(123, op))

		var valueComparators []primitives.ValueComparator
		for _, comparableValue := range comparableValues {
			valueComparators = append(valueComparators, primitives.ValueComparator{
				Value:    comparableValue,
				Operator: op,
			})
		}
		primitiveExpressions = append(primitiveExpressions, query.Comparator("someName", valueComparators...))
	}

	for _, conf := range confirmationsValues {
		primitiveExpressions = append(primitiveExpressions, query.Confidence(conf))
	}

	qf, err := query.Where("primitives", primitiveExpressions...)
	require.NoError(t, err)
	queryFilters = append(queryFilters, qf)

	andOverPrimitivesBoolExpr := query.And(primitiveExpressions...)
	orOverPrimitivesBoolExpr := query.Or(primitiveExpressions...)

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
