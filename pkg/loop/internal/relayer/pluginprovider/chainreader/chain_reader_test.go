package chainreader_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader"
	chainreadertest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests" //nolint
)

func TestVersionedBytesFunctions(t *testing.T) {
	const unsupportedVer = 25913
	t.Run("chainreader.EncodeVersionedBytes unsupported type", func(t *testing.T) {
		invalidData := make(chan int)

		_, err := chainreader.EncodeVersionedBytes(invalidData, chainreader.JSONEncodingVersion2)

		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("chainreader.EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := fmt.Errorf("%w: unsupported encoding version %d for data map[key:value]", types.ErrInvalidEncoding, unsupportedVer)
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := chainreader.EncodeVersionedBytes(data, unsupportedVer)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("chainreader.DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := fmt.Errorf("unsupported encoding version %d for versionedData [97 98 99 100 102]", unsupportedVer)
		versionedBytes := &pb.VersionedBytes{
			Version: unsupportedVer, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := chainreader.DecodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestChainReaderInterfaceTests(t *testing.T) {
	t.Parallel()

	chainreadertest.TestAllEncodings(t, func(version chainreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			fake := &fakeChainReader{}
			RunChainComponentsInterfaceTests(
				t,
				chainreadertest.WrapContractReaderTesterForLoop(
					&fakeChainReaderInterfaceTester{impl: fake},
					chainreadertest.WithChainReaderLoopEncoding(version),
				),
				true,
			)
		}
	})
}

func TestBind(t *testing.T) {
	t.Parallel()

	chainreadertest.TestAllEncodings(t, func(version chainreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errChainReader{}
			errTester := chainreadertest.WrapContractReaderTesterForLoop(
				&fakeChainReaderInterfaceTester{impl: es},
				chainreadertest.WithChainReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			chainReader := errTester.GetChainReader(t)

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("Bind unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					err := chainReader.Bind(ctx, []types.BoundContract{{Name: "Name", Address: "address"}})
					assert.True(t, errors.Is(err, errorType))
				})
			}
		}
	})
}

func TestGetLatestValue(t *testing.T) {
	t.Parallel()

	chainreadertest.TestAllEncodings(t, func(version chainreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errChainReader{}
			errTester := chainreadertest.WrapContractReaderTesterForLoop(
				&fakeChainReaderInterfaceTester{impl: es},
				chainreadertest.WithChainReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			chainReader := errTester.GetChainReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				t.Parallel()

				ctx := tests.Context(t)

				nilTester := chainreadertest.WrapContractReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetChainReader(t)

				err := nilCr.GetLatestValue(ctx, "", "method", primitives.Unconfirmed, "anything", "anything")
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					err := chainReader.GetLatestValue(ctx, "", "method", primitives.Unconfirmed, nil, "anything")
					assert.True(t, errors.Is(err, errorType))
				})
			}

			// make sure that errors come from client directly
			es.err = nil
			t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
				ctx := tests.Context(t)
				err := chainReader.GetLatestValue(ctx, "", "method", primitives.Unconfirmed, &cannotEncode{}, &TestStruct{})
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			})
		}
	})
}

func TestBatchGetLatestValues(t *testing.T) {
	t.Parallel()

	chainreadertest.TestAllEncodings(t, func(version chainreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errChainReader{}
			errTester := chainreadertest.WrapContractReaderTesterForLoop(
				&fakeChainReaderInterfaceTester{impl: es},
				chainreadertest.WithChainReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			chainReader := errTester.GetChainReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				t.Parallel()

				ctx := tests.Context(t)

				nilTester := chainreadertest.WrapContractReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetChainReader(t)

				_, err := nilCr.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("BatchGetLatestValues unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					_, err := chainReader.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
					assert.True(t, errors.Is(err, errorType))
				})
			}

			// make sure that errors come from client directly
			es.err = nil
			t.Run("BatchGetLatestValues returns error if type cannot be encoded in the wire format", func(t *testing.T) {
				ctx := tests.Context(t)
				_, err := chainReader.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{"contract": {{ReadName: "method", Params: &cannotEncode{}, ReturnVal: &cannotEncode{}}}})
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			})
		}
	})
}

func TestQueryKey(t *testing.T) {
	t.Parallel()

	chainreadertest.TestAllEncodings(t, func(version chainreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			impl := &protoConversionTestChainReader{}
			crTester := chainreadertest.WrapContractReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: impl}, chainreadertest.WithChainReaderLoopEncoding(version))
			crTester.Setup(t)
			cr := crTester.GetChainReader(t)

			es := &errChainReader{}
			errTester := chainreadertest.WrapContractReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: es})
			errTester.Setup(t)
			chainReader := errTester.GetChainReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				ctx := tests.Context(t)

				nilTester := chainreadertest.WrapContractReaderTesterForLoop(&fakeChainReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetChainReader(t)

				_, err := nilCr.QueryKey(ctx, "", query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("QueryKey unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					_, err := chainReader.QueryKey(ctx, "", query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
					assert.True(t, errors.Is(err, errorType))
				})
			}

			t.Run("test QueryKey proto conversion", func(t *testing.T) {
				for _, tc := range generateQueryFilterTestCases(t) {
					impl.expectedQueryFilter = tc
					filter, err := query.Where(tc.Key, tc.Expressions...)
					require.NoError(t, err)
					_, err = cr.QueryKey(tests.Context(t), "", filter, query.LimitAndSort{}, &[]interface{}{nil})
					require.NoError(t, err)
				}
			})
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
	impl types.ContractReader
	cw   fakeChainWriter
}

func (it *fakeChainReaderInterfaceTester) Setup(_ *testing.T) {
	fake, ok := it.impl.(*fakeChainReader)
	if ok {
		fake.vals = []valConfidencePair{}
		fake.triggers = []eventConfidencePair{}
		fake.stored = []TestStruct{}
	}
}

func (it *fakeChainReaderInterfaceTester) GetChainReader(_ *testing.T) types.ContractReader {
	return it.impl
}

func (it *fakeChainReaderInterfaceTester) GetChainWriter(_ *testing.T) types.ChainWriter {
	it.cw.cr = it.impl.(*fakeChainReader)
	return &it.cw
}

func (it *fakeChainReaderInterfaceTester) DirtyContracts() {}

func (it *fakeChainReaderInterfaceTester) GetBindings(_ *testing.T) []types.BoundContract {
	return []types.BoundContract{
		{Name: AnyContractName, Address: AnyContractName},
		{Name: AnySecondContractName, Address: AnySecondContractName},
	}
}

func (it *fakeChainReaderInterfaceTester) GenerateBlocksTillConfidenceLevel(t *testing.T, contractName, readName string, confidenceLevel primitives.ConfidenceLevel) {
	fake, ok := it.impl.(*fakeChainReader)
	assert.True(t, ok)
	fake.GenerateBlocksTillConfidenceLevel(t, contractName, readName, confidenceLevel)
}

func (it *fakeChainReaderInterfaceTester) MaxWaitTimeForEvents() time.Duration {
	return time.Millisecond * 100
}

type valConfidencePair struct {
	val             uint64
	confidenceLevel primitives.ConfidenceLevel
}

type eventConfidencePair struct {
	testStruct      TestStruct
	confidenceLevel primitives.ConfidenceLevel
}

type fakeChainReader struct {
	fakeTypeProvider
	vals        []valConfidencePair
	triggers    []eventConfidencePair
	stored      []TestStruct
	batchStored BatchCallEntry
	lock        sync.Mutex
}

type fakeChainWriter struct {
	types.ChainWriter
	cr *fakeChainReader
}

func (f *fakeChainWriter) SubmitTransaction(ctx context.Context, contractName, method string, args any, transactionID string, toAddress string, meta *types.TxMeta, value *big.Int) error {
	switch method {
	case "addTestStruct":
		v, ok := args.(TestStruct)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetTestStructLatestValue(&v)
	case "setAlterablePrimitiveValue":
		v, ok := args.(PrimitiveArgs)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetUintLatestValue(v.Value, ExpectedGetLatestValueArgs{})
	case "triggerEvent":
		v, ok := args.(TestStruct)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetTrigger(&v)
	case "batchChainWrite":
		v, ok := args.(BatchCallEntry)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetBatchLatestValues(v)
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}

	return nil
}

func (f *fakeChainWriter) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	return types.Finalized, nil
}

func (f *fakeChainWriter) GetFeeComponents(ctx context.Context) (*types.ChainFeeComponents, error) {
	return &types.ChainFeeComponents{}, nil
}

func (f *fakeChainReader) Start(_ context.Context) error { return nil }

func (f *fakeChainReader) Close() error { return nil }

func (f *fakeChainReader) Ready() error { panic("unimplemented") }

func (f *fakeChainReader) Name() string { panic("unimplemented") }

func (f *fakeChainReader) HealthReport() map[string]error { panic("unimplemented") }

func (f *fakeChainReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (f *fakeChainReader) SetTestStructLatestValue(ts *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.stored = append(f.stored, *ts)
}

func (f *fakeChainReader) SetUintLatestValue(val uint64, _ ExpectedGetLatestValueArgs) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.vals = append(f.vals, valConfidencePair{val: val, confidenceLevel: primitives.Unconfirmed})
}

func (f *fakeChainReader) SetBatchLatestValues(batchCallEntry BatchCallEntry) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.batchStored = make(BatchCallEntry)
	for contractName, contractBatchEntry := range batchCallEntry {
		f.batchStored[contractName] = contractBatchEntry
	}
}

func (f *fakeChainReader) GetLatestValue(_ context.Context, contractName, method string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	if method == MethodReturningAlterableUint64 {
		r := returnVal.(*uint64)
		for i := len(f.vals) - 1; i >= 0; i-- {
			if f.vals[i].confidenceLevel == confidenceLevel {
				*r = f.vals[i].val
				return nil
			}
		}
		return fmt.Errorf("%w: no val with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if method == MethodReturningUint64 {
		r := returnVal.(*uint64)
		if contractName == AnyContractName {
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

		for i := len(f.triggers) - 1; i >= 0; i-- {
			if f.triggers[i].confidenceLevel == confidenceLevel {
				*returnVal.(*TestStruct) = f.triggers[i].testStruct
				return nil
			}
		}

		return fmt.Errorf("%w: no event with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if method == EventWithFilterName {
		f.lock.Lock()
		defer f.lock.Unlock()
		param := params.(*FilterEventParams)
		for i := len(f.triggers) - 1; i >= 0; i-- {
			if *f.triggers[i].testStruct.Field == param.Field {
				*returnVal.(*TestStruct) = f.triggers[i].testStruct
				return nil
			}
		}
		return types.ErrNotFound
	} else if method != MethodTakingLatestParamsReturningTestStruct {
		return errors.New("unknown method " + method)
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	lp := params.(*LatestParams)
	rv := returnVal.(*TestStruct)
	if lp.I-1 >= len(f.stored) {
		return errors.New("latest params index out of bounds for stored test structs")
	}
	*rv = f.stored[lp.I-1]
	return nil
}

func (f *fakeChainReader) BatchGetLatestValues(_ context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	result := make(types.BatchGetLatestValuesResult)
	for requestContractName, requestContractBatch := range request {
		storedContractBatch := f.batchStored[requestContractName]

		contractBatchResults := types.ContractBatchResults{}
		for i := 0; i < len(requestContractBatch); i++ {
			var err error
			var returnVal any
			req := requestContractBatch[i]
			if req.ReadName == MethodReturningUint64 {
				returnVal = req.ReturnVal.(*uint64)
				if requestContractName == AnyContractName {
					*returnVal.(*uint64) = AnyValueToReadWithoutAnArgument
				} else {
					*returnVal.(*uint64) = AnyDifferentValueToReadWithoutAnArgument
				}
			} else if req.ReadName == MethodReturningUint64Slice {
				returnVal = req.ReturnVal.(*[]uint64)
				*returnVal.(*[]uint64) = AnySliceToReadWithoutAnArgument
			} else if req.ReadName == MethodReturningSeenStruct {
				ts := *req.Params.(*TestStruct)
				ts.Account = anyAccountBytes
				ts.BigField = big.NewInt(2)
				returnVal = &TestStructWithExtraField{
					TestStruct: ts,
					ExtraField: AnyExtraValue,
				}
			} else if req.ReadName == MethodTakingLatestParamsReturningTestStruct {
				latestParams := requestContractBatch[i].Params.(*LatestParams)
				if latestParams.I <= 0 {
					returnVal = &LatestParams{}
					err = fmt.Errorf("invalid param %d", latestParams.I)
				} else {
					returnVal = storedContractBatch[latestParams.I-1].ReturnValue
				}
			} else {
				return nil, errors.New("unknown read " + req.ReadName)
			}
			brr := types.BatchReadResult{ReadName: req.ReadName}
			brr.SetResult(returnVal, err)
			contractBatchResults = append(contractBatchResults, brr)
		}
		result[requestContractName] = contractBatchResults
	}
	return result, nil
}

func (f *fakeChainReader) QueryKey(_ context.Context, _ string, filter query.KeyFilter, limitAndSort query.LimitAndSort, _ any) ([]types.Sequence, error) {
	if filter.Key == EventName {
		f.lock.Lock()
		defer f.lock.Unlock()
		if len(f.triggers) == 0 {
			return []types.Sequence{}, nil
		}

		var sequences []types.Sequence
		for _, trigger := range f.triggers {
			sequences = append(sequences, types.Sequence{Data: trigger.testStruct})
		}

		if !limitAndSort.HasSequenceSort() {
			sort.Slice(sequences, func(i, j int) bool {
				if sequences[i].Data.(TestStruct).Field == nil || sequences[j].Data.(TestStruct).Field == nil {
					return false
				}
				return *sequences[i].Data.(TestStruct).Field > *sequences[j].Data.(TestStruct).Field
			})
		}

		return sequences, nil
	}
	return nil, nil
}

func (f *fakeChainReader) SetTrigger(testStruct *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.triggers = append(f.triggers, eventConfidencePair{testStruct: *testStruct, confidenceLevel: primitives.Unconfirmed})
}

func (f *fakeChainReader) GenerateBlocksTillConfidenceLevel(_ *testing.T, _, _ string, confidenceLevel primitives.ConfidenceLevel) {
	f.lock.Lock()
	defer f.lock.Unlock()
	for i, val := range f.vals {
		f.vals[i] = valConfidencePair{val: val.val, confidenceLevel: confidenceLevel}
	}

	for i, trigger := range f.triggers {
		f.triggers[i] = eventConfidencePair{testStruct: trigger.testStruct, confidenceLevel: confidenceLevel}
	}
}

type errChainReader struct {
	err error
}

func (e *errChainReader) Start(_ context.Context) error { return nil }

func (e *errChainReader) Close() error { return nil }

func (e *errChainReader) Ready() error { panic("unimplemented") }

func (e *errChainReader) Name() string { panic("unimplemented") }

func (e *errChainReader) HealthReport() map[string]error { panic("unimplemented") }

func (e *errChainReader) GetLatestValue(_ context.Context, _, _ string, _ primitives.ConfidenceLevel, _, _ any) error {
	return e.err
}

func (e *errChainReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, e.err
}

func (e *errChainReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return e.err
}

func (e *errChainReader) QueryKey(_ context.Context, _ string, _ query.KeyFilter, _ query.LimitAndSort, _ any) ([]types.Sequence, error) {
	return nil, e.err
}

type protoConversionTestChainReader struct {
	expectedBindings     types.BoundContract
	expectedQueryFilter  query.KeyFilter
	expectedLimitAndSort query.LimitAndSort
}

func (pc *protoConversionTestChainReader) Start(_ context.Context) error { return nil }

func (pc *protoConversionTestChainReader) Close() error { return nil }

func (pc *protoConversionTestChainReader) Ready() error { panic("unimplemented") }

func (pc *protoConversionTestChainReader) Name() string { panic("unimplemented") }

func (pc *protoConversionTestChainReader) HealthReport() map[string]error { panic("unimplemented") }

func (pc *protoConversionTestChainReader) GetLatestValue(_ context.Context, _, _ string, _ primitives.ConfidenceLevel, _, _ any) error {
	return nil
}

func (pc *protoConversionTestChainReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, nil
}

func (pc *protoConversionTestChainReader) Bind(_ context.Context, bc []types.BoundContract) error {
	if !reflect.DeepEqual(pc.expectedBindings, bc) {
		return fmt.Errorf("bound contract wasn't parsed properly")
	}
	return nil
}

func (pc *protoConversionTestChainReader) QueryKey(_ context.Context, _ string, filter query.KeyFilter, limitAndSort query.LimitAndSort, _ any) ([]types.Sequence, error) {
	if !reflect.DeepEqual(pc.expectedQueryFilter, filter) {
		return nil, fmt.Errorf("filter wasn't parsed properly")
	}

	// using deep equal on a slice returns false when one slice is nil and another is empty
	// normalize to nil slices if empty or nil for comparison
	var (
		aSlice []query.SortBy
		bSlice []query.SortBy
	)

	if len(pc.expectedLimitAndSort.SortBy) > 0 {
		aSlice = pc.expectedLimitAndSort.SortBy
	}

	if len(limitAndSort.SortBy) > 0 {
		bSlice = limitAndSort.SortBy
	}

	if !reflect.DeepEqual(pc.expectedLimitAndSort.Limit, limitAndSort.Limit) || !reflect.DeepEqual(aSlice, bSlice) {
		return nil, fmt.Errorf("limitAndSort wasn't parsed properly")
	}

	return nil, nil
}
