package internal_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestVersionedBytesFunctions(t *testing.T) {
	const unsupportedVer = 25913
	t.Run("internal.EncodeVersionedBytes unsupported type", func(t *testing.T) {
		invalidData := make(chan int)

		_, err := internal.EncodeVersionedBytes(invalidData, internal.JSONEncodingVersion2)

		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("internal.EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := fmt.Errorf("%w: unsupported encoding version %d for data map[key:value]", types.ErrInvalidEncoding, unsupportedVer)
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := internal.EncodeVersionedBytes(data, unsupportedVer)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("internal.DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := fmt.Errorf("unsupported encoding version %d for versionedData [97 98 99 100 102]", unsupportedVer)
		versionedBytes := &pb.VersionedBytes{
			Version: unsupportedVer, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := internal.DecodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestChainReaderClient(t *testing.T) {
	fake := &fakeChainReader{}
	RunChainReaderInterfaceTests(t, test.WrapChainReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: fake}))

	es := &errChainReader{}
	errTester := test.WrapChainReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: es})
	errTester.Setup(t)
	chainReader := errTester.GetChainReader(t)

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := tests.Context(t)
			err := chainReader.GetLatestValue(ctx, "", "method", nil, "anything")
			assert.True(t, errors.Is(err, errorType))
		})
	}

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("Bind unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := tests.Context(t)
			err := chainReader.Bind(ctx, []types.BoundContract{{Name: "Name", Address: "address"}})
			assert.True(t, errors.Is(err, errorType))
		})
	}

	// make sure that errors come from client directly
	es.err = nil
	t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		ctx := tests.Context(t)
		err := chainReader.GetLatestValue(ctx, "", "method", &cannotEncode{}, &TestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("nil reader should return unimplemented", func(t *testing.T) {
		ctx := tests.Context(t)

		nilTester := test.WrapChainReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: nil})
		nilTester.Setup(t)
		nilCr := nilTester.GetChainReader(t)

		err := nilCr.GetLatestValue(ctx, "", "method", "anything", "anything")
		assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
	})
}

func generateQueryFilterTestCases() []types.QueryFilter {
	confirmationsValues := []types.Confirmations{types.Finalized, types.Unconfirmed}
	operatorValues := []types.ComparisonOperator{types.Eq, types.Neq, types.Gt, types.Lt, types.Gte, types.Lte}

	singleFilters := []types.QueryFilter{&types.TxHashFilter{TxHash: "txhash"}}
	for _, op := range operatorValues {
		singleFilters = append(singleFilters, &types.BlockFilter{Block: 123, Operator: op})
		singleFilters = append(singleFilters, &types.TimestampFilter{Timestamp: 123, Operator: op})
	}
	for _, conf := range confirmationsValues {
		singleFilters = append(singleFilters, &types.ConfirmationsFilter{Confirmations: conf})
		singleFilters = append(singleFilters, &types.AddressFilter{Addresses: []string{"addr1", "addr2"}})
	}

	andFilter := &types.AndFilter{Filters: singleFilters}
	nestedAndFilter := &types.AndFilter{Filters: append(singleFilters, andFilter)}

	return append(
		singleFilters,
		append([]types.QueryFilter{andFilter}, nestedAndFilter)...,
	)
}

func TestChainReaderProtoRequestsConversions(t *testing.T) {
	impl := &protoConversionTestChainReader{}
	crTester := test.WrapChainReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: impl})
	crTester.Setup(t)
	cr := crTester.GetChainReader(t)

	queryFilterTestCases := generateQueryFilterTestCases()
	t.Run("test QueryKey proto conversion", func(t *testing.T) {
		for _, tc := range queryFilterTestCases {
			impl.expectedQueryFilter = tc
			_, err := cr.QueryKey(context.Background(), "", tc, types.LimitAndSort{})
			require.NoError(t, err)
		}
	})

	t.Run("test QueryKeys proto conversion", func(t *testing.T) {
		for _, tc := range queryFilterTestCases {
			impl.expectedQueryFilter = tc
			_, err := cr.QueryKeys(context.Background(), []string{"", ""}, tc, types.LimitAndSort{})
			require.NoError(t, err)
		}
	})

	t.Run("test QueryKeyByValues proto conversion", func(t *testing.T) {
		for _, tc := range queryFilterTestCases {
			impl.expectedQueryFilter = tc
			_, err := cr.QueryKeyByValues(context.Background(), "", []string{"", ""}, tc, types.LimitAndSort{})
			require.NoError(t, err)
		}
	})

	t.Run("test QueryKeysByValues proto conversion", func(t *testing.T) {
		for _, tc := range queryFilterTestCases {
			impl.expectedQueryFilter = tc
			_, err := cr.QueryKeysByValues(context.Background(), []string{"", ""}, [][]string{{""}}, tc, types.LimitAndSort{})
			require.NoError(t, err)
		}
	})
}

var encoder = makeEncoder()

func makeEncoder() cbor.EncMode {
	opts := cbor.CoreDetEncOptions()
	opts.Sort = cbor.SortCanonical
	e, _ := opts.EncMode()
	return e
}

type fakeChainReaderInterfaceTester struct {
	interfaceTesterBase
	impl types.ChainReader
}

func (it *fakeChainReaderInterfaceTester) Setup(_ *testing.T) {
	fake, ok := it.impl.(*fakeChainReader)
	if ok {
		fake.stored = []TestStruct{}
		fake.triggers = []TestStruct{}
	}
}

func (it *fakeChainReaderInterfaceTester) GetChainReader(_ *testing.T) types.ChainReader {
	return it.impl
}

func (it *fakeChainReaderInterfaceTester) GetBindings(_ *testing.T) []types.BoundContract {
	return []types.BoundContract{
		{Name: AnyContractName, Address: AnyContractName},
		{Name: AnySecondContractName, Address: AnySecondContractName},
	}
}

func (it *fakeChainReaderInterfaceTester) SetLatestValue(t *testing.T, testStruct *TestStruct) {
	fake, ok := it.impl.(*fakeChainReader)
	assert.True(t, ok)
	fake.SetLatestValue(testStruct)
}

func (it *fakeChainReaderInterfaceTester) TriggerEvent(t *testing.T, testStruct *TestStruct) {
	fake, ok := it.impl.(*fakeChainReader)
	assert.True(t, ok)
	fake.SetTrigger(testStruct)
}

func (it *fakeChainReaderInterfaceTester) MaxWaitTimeForEvents() time.Duration {
	return time.Millisecond * 100
}

type fakeChainReader struct {
	fakeTypeProvider
	stored   []TestStruct
	lock     sync.Mutex
	triggers []TestStruct
}

func (f *fakeChainReader) Bind(context.Context, []types.BoundContract) error {
	return nil
}

func (f *fakeChainReader) SetLatestValue(ts *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.stored = append(f.stored, *ts)
}

func (f *fakeChainReader) GetLatestValue(_ context.Context, name, method string, params, returnVal any) error {
	if method == MethodReturningUint64 {
		r := returnVal.(*uint64)
		if name == AnyContractName {
			*r = AnyValueToReadWithoutAnArgument
		} else {
			*r = AnyDifferentValueToReadWithoutAnArgument
		}

		return nil
	} else if method == MethodReturningUint64Slice {
		r := returnVal.(*[]uint64)
		*r = AnySliceToReadWithoutAnArgument
		return nil
	} else if method == MethodReturningSeenStruct {
		pv := params.(*TestStruct)
		rv := returnVal.(*TestStructWithExtraField)
		rv.TestStruct = *pv
		rv.ExtraField = AnyExtraValue
		rv.Account = anyAccountBytes
		rv.BigField = big.NewInt(2)
		return nil
	} else if method == EventName {
		f.lock.Lock()
		defer f.lock.Unlock()
		if len(f.triggers) == 0 {
			return types.ErrNotFound
		}
		*returnVal.(*TestStruct) = f.triggers[len(f.triggers)-1]
		return nil
	} else if method == EventWithFilterName {
		f.lock.Lock()
		defer f.lock.Unlock()
		param := params.(*FilterEventParams)
		for i := len(f.triggers) - 1; i >= 0; i-- {
			if *f.triggers[i].Field == param.Field {
				*returnVal.(*TestStruct) = f.triggers[i]
				return nil
			}
		}
		return types.ErrNotFound
	} else if method == DifferentMethodReturningUint64 {
		r := returnVal.(*uint64)
		*r = AnyDifferentValueToReadWithoutAnArgument
		return nil
	} else if method != MethodTakingLatestParamsReturningTestStruct {
		return errors.New("unknown method " + method)
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	lp := params.(*LatestParams)
	rv := returnVal.(*TestStruct)
	*rv = f.stored[lp.I-1]
	return nil
}

func (f *fakeChainReader) QueryKey(_ context.Context, _ string, _ types.QueryFilter, _ types.LimitAndSort) ([]types.Sequence, error) {
	return nil, nil
}

func (f *fakeChainReader) QueryKeys(_ context.Context, _ []string, _ types.QueryFilter, _ types.LimitAndSort) ([][]types.Sequence, error) {
	return nil, nil
}

func (f *fakeChainReader) QueryKeyByValues(_ context.Context, _ string, _ []string, _ types.QueryFilter, _ types.LimitAndSort) ([]types.Sequence, error) {
	return nil, nil
}

func (f *fakeChainReader) QueryKeysByValues(_ context.Context, _ []string, _ [][]string, _ types.QueryFilter, _ types.LimitAndSort) ([][]types.Sequence, error) {
	return nil, nil
}

func (f *fakeChainReader) SetTrigger(testStruct *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.triggers = append(f.triggers, *testStruct)
}

type errChainReader struct {
	err error
}

func (e *errChainReader) GetLatestValue(_ context.Context, _, _ string, _, _ any) error {
	return e.err
}

func (e *errChainReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return e.err
}

func (e *errChainReader) QueryKey(_ context.Context, _ string, _ types.QueryFilter, _ types.LimitAndSort) ([]types.Sequence, error) {
	return nil, nil
}

func (e *errChainReader) QueryKeys(_ context.Context, _ []string, _ types.QueryFilter, _ types.LimitAndSort) ([][]types.Sequence, error) {
	return nil, nil
}

func (e *errChainReader) QueryKeyByValues(_ context.Context, _ string, _ []string, _ types.QueryFilter, _ types.LimitAndSort) ([]types.Sequence, error) {
	return nil, nil
}

func (e *errChainReader) QueryKeysByValues(_ context.Context, _ []string, _ [][]string, _ types.QueryFilter, _ types.LimitAndSort) ([][]types.Sequence, error) {
	return nil, nil
}

type protoConversionTestChainReader struct {
	expectedBindings     types.BoundContract
	expectedQueryFilter  types.QueryFilter
	expectedLimitAndSort types.LimitAndSort
}

func (e *protoConversionTestChainReader) GetLatestValue(_ context.Context, _, _ string, _, _ any) error {
	return nil
}

func (e *protoConversionTestChainReader) Bind(_ context.Context, bc []types.BoundContract) error {
	if !reflect.DeepEqual(e.expectedBindings, bc) {
		return fmt.Errorf("bound contract wasn't parsed properly")
	}
	return nil
}

func (e *protoConversionTestChainReader) QueryKey(_ context.Context, _ string, filter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Sequence, error) {
	if !reflect.DeepEqual(e.expectedQueryFilter, filter) {
		return nil, fmt.Errorf("filter wasn't parsed properly")
	}

	if !reflect.DeepEqual(e.expectedLimitAndSort, limitAndSort) {
		return nil, fmt.Errorf("limitAndSort wasn't parsed properly")
	}
	return nil, nil
}

func (e *protoConversionTestChainReader) QueryKeys(_ context.Context, _ []string, filter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	if !reflect.DeepEqual(e.expectedQueryFilter, filter) {
		return nil, fmt.Errorf("filter wasn't parsed properly")
	}

	if !reflect.DeepEqual(e.expectedLimitAndSort, limitAndSort) {
		return nil, fmt.Errorf("limitAndSort wasn't parsed properly")
	}
	return nil, nil
}

func (e *protoConversionTestChainReader) QueryKeyByValues(_ context.Context, _ string, _ []string, filter types.QueryFilter, limitAndSort types.LimitAndSort) ([]types.Sequence, error) {
	if !reflect.DeepEqual(e.expectedQueryFilter, filter) {
		return nil, fmt.Errorf("filter wasn't parsed properly")
	}

	if !reflect.DeepEqual(e.expectedLimitAndSort, limitAndSort) {
		return nil, fmt.Errorf("limitAndSort wasn't parsed properly")
	}
	return nil, nil
}

func (e *protoConversionTestChainReader) QueryKeysByValues(_ context.Context, _ []string, _ [][]string, filter types.QueryFilter, limitAndSort types.LimitAndSort) ([][]types.Sequence, error) {
	if !reflect.DeepEqual(e.expectedQueryFilter, filter) {
		return nil, fmt.Errorf("filter wasn't parsed properly")
	}

	if !reflect.DeepEqual(e.expectedLimitAndSort, limitAndSort) {
		return nil, fmt.Errorf("limitAndSort wasn't parsed properly")
	}
	return nil, nil
}
