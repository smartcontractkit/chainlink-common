package contractreader_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/chainreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	contractreadertest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader/test"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests" //nolint
)

func TestVersionedBytesFunctions(t *testing.T) {
	const unsupportedVer = 25913
	t.Run("contractreader.EncodeVersionedBytes unsupported type", func(t *testing.T) {
		invalidData := make(chan int)

		_, err := contractreader.EncodeVersionedBytes(invalidData, contractreader.JSONEncodingVersion2)

		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("contractreader.EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := fmt.Errorf("%w: unsupported encoding version %d for data map[key:value]", types.ErrInvalidEncoding, unsupportedVer)
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := contractreader.EncodeVersionedBytes(data, unsupportedVer)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("contractreader.DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := fmt.Errorf("unsupported encoding version %d for versionedData [97 98 99 100 102]", unsupportedVer)
		versionedBytes := &pb.VersionedBytes{
			Version: unsupportedVer, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := contractreader.DecodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestContractReaderInterfaceTests(t *testing.T) {
	t.Parallel()

	contractreadertest.TestAllEncodings(t, func(version contractreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			fake := &fakeContractReader{}
			RunContractReaderInterfaceTests(
				t,
				contractreadertest.WrapContractReaderTesterForLoop(
					&fakeContractReaderInterfaceTester{impl: fake},
					contractreadertest.WithContractReaderLoopEncoding(version),
				),
				true,
			)
		}
	})
}

func TestContractReaderByIDWrapper(t *testing.T) {
	t.Parallel()
	t.Run("Contract Reader by ID GetLatestValue", runContractReaderByIDGetLatestValue)
	t.Run("Contract Reader by ID BatchGetLatestValues", runContractReaderByIDBatchGetLatestValues)
	t.Run("Contract Reader by ID QueryKey", runContractReaderByIDQueryKey)
}

func TestBind(t *testing.T) {
	t.Parallel()

	contractreadertest.TestAllEncodings(t, func(version contractreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errContractReader{}
			errTester := contractreadertest.WrapContractReaderTesterForLoop(
				&fakeContractReaderInterfaceTester{impl: es},
				contractreadertest.WithContractReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			contractReader := errTester.GetContractReader(t)

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("Bind unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					err := contractReader.Bind(ctx, []types.BoundContract{{Name: "Contract", Address: "address"}})
					assert.True(t, errors.Is(err, errorType))
				})

				t.Run("Unbind unwraps errors from server"+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					err := contractReader.Unbind(ctx, []types.BoundContract{{Name: "Contract", Address: "address"}})
					assert.True(t, errors.Is(err, errorType))
				})
			}
		}
	})
}

func TestGetLatestValue(t *testing.T) {
	t.Parallel()

	contractreadertest.TestAllEncodings(t, func(version contractreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errContractReader{}
			errTester := contractreadertest.WrapContractReaderTesterForLoop(
				&fakeContractReaderInterfaceTester{impl: es},
				contractreadertest.WithContractReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			contractReader := errTester.GetContractReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				t.Parallel()

				ctx := tests.Context(t)

				nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetContractReader(t)

				err := nilCr.GetLatestValue(ctx, "method", primitives.Unconfirmed, "anything", "anything")
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					err := contractReader.GetLatestValue(ctx, "method", primitives.Unconfirmed, nil, "anything")
					assert.True(t, errors.Is(err, errorType))
				})
			}

			// make sure that errors come from client directly
			es.err = nil
			t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
				ctx := tests.Context(t)
				err := contractReader.GetLatestValue(ctx, "method", primitives.Unconfirmed, &cannotEncode{}, &TestStruct{})
				assert.True(t, errors.Is(err, types.ErrInvalidType))
			})
		}
	})
}

func TestBatchGetLatestValues(t *testing.T) {
	t.Parallel()

	contractreadertest.TestAllEncodings(t, func(version contractreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			es := &errContractReader{}
			errTester := contractreadertest.WrapContractReaderTesterForLoop(
				&fakeContractReaderInterfaceTester{impl: es},
				contractreadertest.WithContractReaderLoopEncoding(version),
			)

			errTester.Setup(t)
			contractReader := errTester.GetContractReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				t.Parallel()

				ctx := tests.Context(t)

				nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetContractReader(t)

				_, err := nilCr.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("BatchGetLatestValues unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					_, err := contractReader.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
					assert.True(t, errors.Is(err, errorType))
				})
			}

			// make sure that errors come from client directly
			es.err = nil
			t.Run("BatchGetLatestValues returns error if type cannot be encoded in the wire format", func(t *testing.T) {
				ctx := tests.Context(t)
				_, err := contractReader.BatchGetLatestValues(
					ctx,
					types.BatchGetLatestValuesRequest{
						types.BoundContract{Name: "contract"}: {
							{ReadName: "method", Params: &cannotEncode{}, ReturnVal: &cannotEncode{}},
						},
					},
				)

				assert.True(t, errors.Is(err, types.ErrInvalidType))
			})
		}
	})
}

func TestQueryKey(t *testing.T) {
	t.Parallel()

	contractreadertest.TestAllEncodings(t, func(version contractreader.EncodingVersion) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			impl := &protoConversionTestContractReader{}
			crTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: impl}, contractreadertest.WithContractReaderLoopEncoding(version))
			crTester.Setup(t)
			cr := crTester.GetContractReader(t)

			es := &errContractReader{}
			errTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: es})
			errTester.Setup(t)
			contractReader := errTester.GetContractReader(t)

			t.Run("nil reader should return unimplemented", func(t *testing.T) {
				ctx := tests.Context(t)

				nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
				nilTester.Setup(t)
				nilCr := nilTester.GetContractReader(t)

				_, err := nilCr.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
				assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
			})

			for _, errorType := range errorTypes {
				es.err = errorType
				t.Run("QueryKey unwraps errors from server "+errorType.Error(), func(t *testing.T) {
					ctx := tests.Context(t)
					_, err := contractReader.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
					assert.True(t, errors.Is(err, errorType))
				})
			}

			t.Run("test QueryKey proto conversion", func(t *testing.T) {
				for _, tc := range generateQueryFilterTestCases(t) {
					impl.expectedQueryFilter = tc
					filter, err := query.Where(tc.Key, tc.Expressions...)
					require.NoError(t, err)
					_, err = cr.QueryKey(tests.Context(t), types.BoundContract{}, filter, query.LimitAndSort{}, &[]interface{}{nil})
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

type fakeContractReaderInterfaceTester struct {
	interfaceTesterBase
	impl types.ContractReader
	cw   fakeChainWriter
}

func (it *fakeContractReaderInterfaceTester) Setup(_ *testing.T) {
	fake, ok := it.impl.(*fakeContractReader)
	if ok {
		fake.vals = make(map[string][]valConfidencePair)
		fake.triggers = make(map[string][]eventConfidencePair)
		fake.stored = make(map[string][]TestStruct)
	}
}

func (it *fakeContractReaderInterfaceTester) GetContractReader(_ *testing.T) types.ContractReader {
	return it.impl
}

func (it *fakeContractReaderInterfaceTester) GetChainWriter(_ *testing.T) types.ChainWriter {
	it.cw.cr = it.impl.(*fakeContractReader)
	return &it.cw
}

func (it *fakeContractReaderInterfaceTester) DirtyContracts() {}

func (it *fakeContractReaderInterfaceTester) GetBindings(_ *testing.T) []types.BoundContract {
	return []types.BoundContract{
		{Name: AnyContractName, Address: AnyContractName},
		{Name: AnyContractName, Address: AnyContractName + "-2"},
		{Name: AnySecondContractName, Address: AnySecondContractName},
		{Name: AnySecondContractName, Address: AnySecondContractName + "-2"},
	}
}

func (it *fakeContractReaderInterfaceTester) GenerateBlocksTillConfidenceLevel(t *testing.T, contractID, readIdentifier string, confidenceLevel primitives.ConfidenceLevel) {
	fake, ok := it.impl.(*fakeContractReader)
	assert.True(t, ok)
	fake.GenerateBlocksTillConfidenceLevel(t, contractID, readIdentifier, confidenceLevel)
}

func (it *fakeContractReaderInterfaceTester) MaxWaitTimeForEvents() time.Duration {
	return time.Millisecond * 1000
}

type valConfidencePair struct {
	val             uint64
	confidenceLevel primitives.ConfidenceLevel
}

type eventConfidencePair struct {
	testStruct      TestStruct
	confidenceLevel primitives.ConfidenceLevel
}

type fakeContractReader struct {
	types.UnimplementedContractReader
	fakeTypeProvider
	vals        map[string][]valConfidencePair
	triggers    map[string][]eventConfidencePair
	stored      map[string][]TestStruct
	batchStored BatchCallEntry
	lock        sync.Mutex
}

type fakeChainWriter struct {
	types.ChainWriter
	cr *fakeContractReader
}

func (f *fakeChainWriter) SubmitTransaction(_ context.Context, contractName, method string, args any, transactionID string, toAddress string, meta *types.TxMeta, value *big.Int) error {
	contractID := toAddress + "-" + contractName
	switch method {
	case MethodSettingStruct:
		v, ok := args.(TestStruct)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetTestStructLatestValue(contractID, &v)
	case MethodSettingUint64:
		v, ok := args.(PrimitiveArgs)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetUintLatestValue(contractID, v.Value, ExpectedGetLatestValueArgs{})
	case MethodTriggeringEvent:
		v, ok := args.(TestStruct)
		if !ok {
			return fmt.Errorf("unexpected type %T", args)
		}
		f.cr.SetTrigger(contractID, &v)
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

func (f *fakeContractReader) Start(_ context.Context) error { return nil }

func (f *fakeContractReader) Close() error { return nil }

func (f *fakeContractReader) Ready() error { panic("unimplemented") }

func (f *fakeContractReader) Name() string { panic("unimplemented") }

func (f *fakeContractReader) HealthReport() map[string]error { panic("unimplemented") }

func (f *fakeContractReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (f *fakeContractReader) Unbind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (f *fakeContractReader) SetTestStructLatestValue(contractID string, ts *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, ok := f.stored[contractID]; !ok {
		f.stored[contractID] = []TestStruct{}
	}
	f.stored[contractID] = append(f.stored[contractID], *ts)
}

func (f *fakeContractReader) SetUintLatestValue(contractID string, val uint64, _ ExpectedGetLatestValueArgs) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, ok := f.vals[contractID]; !ok {
		f.vals[contractID] = []valConfidencePair{}
	}
	f.vals[contractID] = append(f.vals[contractID], valConfidencePair{val: val, confidenceLevel: primitives.Unconfirmed})
}

func (f *fakeContractReader) SetBatchLatestValues(batchCallEntry BatchCallEntry) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.batchStored = make(BatchCallEntry)
	for contractName, contractBatchEntry := range batchCallEntry {
		f.batchStored[contractName] = contractBatchEntry
	}
}

func (f *fakeContractReader) GetLatestValue(_ context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	split := strings.Split(readIdentifier, "-")
	contractName := strings.Join([]string{split[0], split[1]}, "-")
	if strings.HasSuffix(readIdentifier, MethodReturningAlterableUint64) {
		r := returnVal.(*uint64)
		vals := f.vals[contractName]
		for i := len(vals) - 1; i >= 0; i-- {
			if vals[i].confidenceLevel == confidenceLevel {
				*r = vals[i].val
				return nil
			}
		}
		return fmt.Errorf("%w: no val with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if strings.HasSuffix(readIdentifier, MethodReturningUint64) {
		r := returnVal.(*uint64)

		if strings.Contains(readIdentifier, "-"+AnyContractName+"-") {
			*r = AnyValueToReadWithoutAnArgument
		} else {
			*r = AnyDifferentValueToReadWithoutAnArgument
		}

		return nil
	} else if strings.HasSuffix(readIdentifier, MethodReturningUint64Slice) {
		r := returnVal.(*[]uint64)
		*r = AnySliceToReadWithoutAnArgument
		return nil
	} else if strings.HasSuffix(readIdentifier, MethodReturningSeenStruct) {
		pv := params.(*TestStruct)
		rv := returnVal.(*TestStructWithExtraField)
		rv.TestStruct = *pv
		rv.ExtraField = AnyExtraValue
		rv.Account = anyAccountBytes
		rv.BigField = big.NewInt(2)
		return nil
	} else if strings.HasSuffix(readIdentifier, EventName) {
		f.lock.Lock()
		defer f.lock.Unlock()

		triggers := f.triggers[contractName]
		if len(triggers) == 0 {
			return types.ErrNotFound
		}

		for i := len(triggers) - 1; i >= 0; i-- {
			if triggers[i].confidenceLevel == confidenceLevel {
				*returnVal.(*TestStruct) = triggers[i].testStruct
				return nil
			}
		}

		return fmt.Errorf("%w: no event with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if strings.HasSuffix(readIdentifier, EventWithFilterName) {
		f.lock.Lock()
		defer f.lock.Unlock()
		param := params.(*FilterEventParams)
		triggers := f.triggers[contractName]
		for i := len(triggers) - 1; i >= 0; i-- {
			if *triggers[i].testStruct.Field == param.Field {
				*returnVal.(*TestStruct) = triggers[i].testStruct
				return nil
			}
		}
		return types.ErrNotFound
	} else if !strings.HasSuffix(readIdentifier, MethodTakingLatestParamsReturningTestStruct) {
		return errors.New("unknown method " + readIdentifier)
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	stored := f.stored[contractName]
	lp := params.(*LatestParams)

	if lp.I-1 >= len(stored) {
		return errors.New("latest params index out of bounds for stored test structs")
	}

	_, isValue := returnVal.(*values.Value)
	if isValue {
		var err error
		ptrToVal := returnVal.(*values.Value)
		*ptrToVal, err = values.Wrap(stored[lp.I-1])
		if err != nil {
			return err
		}
	} else {
		rv := returnVal.(*TestStruct)
		*rv = stored[lp.I-1]
	}

	return nil
}

func (f *fakeContractReader) BatchGetLatestValues(_ context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	result := make(types.BatchGetLatestValuesResult)
	for requestContract, requestContractBatch := range request {
		storedContractBatch := f.batchStored[requestContract]

		contractBatchResults := types.ContractBatchResults{}
		for i := 0; i < len(requestContractBatch); i++ {
			var err error
			var returnVal any

			req := requestContractBatch[i]

			if req.ReadName == MethodReturningUint64 {
				returnVal = req.ReturnVal.(*uint64)
				if requestContract.Name == AnyContractName {
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

		result[requestContract] = contractBatchResults
	}

	return result, nil
}

func (f *fakeContractReader) QueryKey(_ context.Context, bc types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceType any) ([]types.Sequence, error) {
	_, isValueType := sequenceType.(*values.Value)
	if filter.Key == EventName {
		f.lock.Lock()
		defer f.lock.Unlock()
		if len(f.triggers) == 0 {
			return []types.Sequence{}, nil
		}

		var sequences []types.Sequence
		for idx, trigger := range f.triggers[bc.String()] {
			doAppend := true
			for _, expr := range filter.Expressions {
				if primitive, ok := expr.Primitive.(*primitives.Comparator); ok {
					if len(primitive.ValueComparators) == 0 {
						return nil, fmt.Errorf("value comparator for %s should not be empty", primitive.Name)
					}
					if primitive.Name == "Field" {
						for _, valComp := range primitive.ValueComparators {
							doAppend = doAppend && Compare(*trigger.testStruct.Field, *valComp.Value.(*int32), valComp.Operator)
						}
					}
				}
			}

			var skipAppend bool
			if limitAndSort.HasCursorLimit() {
				cursor, err := strconv.Atoi(limitAndSort.Limit.Cursor)
				if err != nil {
					return nil, err
				}

				// assume CursorFollowing order for now
				if cursor >= idx {
					skipAppend = true
				}
			}

			if (len(filter.Expressions) == 0 || doAppend) && !skipAppend {
				if isValueType {
					value, err := values.Wrap(trigger.testStruct)
					if err != nil {
						return nil, err
					}
					sequences = append(sequences, types.Sequence{Cursor: strconv.Itoa(idx), Data: &value})
				} else {
					sequences = append(sequences, types.Sequence{Cursor: fmt.Sprintf("%d", idx), Data: trigger.testStruct})
				}
			}

			if limitAndSort.Limit.Count > 0 && len(sequences) >= int(limitAndSort.Limit.Count) {
				break
			}
		}

		if isValueType {
			if !limitAndSort.HasSequenceSort() && !limitAndSort.HasCursorLimit() {
				sort.Slice(sequences, func(i, j int) bool {
					valI := *sequences[i].Data.(*values.Value)
					valJ := *sequences[j].Data.(*values.Value)

					mapI := valI.(*values.Map)
					mapJ := valJ.(*values.Map)

					if mapI.Underlying["Field"] == nil || mapJ.Underlying["Field"] == nil {
						return false
					}
					var iVal int32
					err := mapI.Underlying["Field"].UnwrapTo(&iVal)
					if err != nil {
						panic(err)
					}

					var jVal int32
					err = mapJ.Underlying["Field"].UnwrapTo(&jVal)
					if err != nil {
						panic(err)
					}

					return iVal > jVal
				})
			}
		} else {
			if !limitAndSort.HasSequenceSort() && !limitAndSort.HasCursorLimit() {
				sort.Slice(sequences, func(i, j int) bool {
					if sequences[i].Data.(TestStruct).Field == nil || sequences[j].Data.(TestStruct).Field == nil {
						return false
					}
					return *sequences[i].Data.(TestStruct).Field > *sequences[j].Data.(TestStruct).Field
				})
			}
		}

		return sequences, nil
	}
	return nil, nil
}

func (f *fakeContractReader) SetTrigger(contractID string, testStruct *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	if _, ok := f.triggers[contractID]; !ok {
		f.triggers[contractID] = []eventConfidencePair{}
	}

	f.triggers[contractID] = append(f.triggers[contractID], eventConfidencePair{testStruct: *testStruct, confidenceLevel: primitives.Unconfirmed})
}

func (f *fakeContractReader) GenerateBlocksTillConfidenceLevel(_ *testing.T, _, _ string, confidenceLevel primitives.ConfidenceLevel) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for contractID, vals := range f.vals {
		for i, val := range vals {
			f.vals[contractID][i] = valConfidencePair{val: val.val, confidenceLevel: confidenceLevel}
		}
	}

	for contractID, triggers := range f.triggers {
		for i, trigger := range triggers {
			f.triggers[contractID][i] = eventConfidencePair{testStruct: trigger.testStruct, confidenceLevel: confidenceLevel}
		}
	}
}

type errContractReader struct {
	types.UnimplementedContractReader
	err error
}

func (e *errContractReader) Start(_ context.Context) error { return nil }

func (e *errContractReader) Close() error { return nil }

func (e *errContractReader) Ready() error { panic("unimplemented") }

func (e *errContractReader) Name() string { panic("unimplemented") }

func (e *errContractReader) HealthReport() map[string]error { panic("unimplemented") }

func (e *errContractReader) GetLatestValue(_ context.Context, _ string, _ primitives.ConfidenceLevel, _, _ any) error {
	return e.err
}

func (e *errContractReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, e.err
}

func (e *errContractReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return e.err
}

func (e *errContractReader) Unbind(_ context.Context, _ []types.BoundContract) error {
	return e.err
}

func (e *errContractReader) QueryKey(_ context.Context, _ types.BoundContract, _ query.KeyFilter, _ query.LimitAndSort, _ any) ([]types.Sequence, error) {
	return nil, e.err
}

type protoConversionTestContractReader struct {
	types.UnimplementedContractReader
	testProtoConversionTypeProvider
	expectedBindings     types.BoundContract
	expectedQueryFilter  query.KeyFilter
	expectedLimitAndSort query.LimitAndSort
}

func (pc *protoConversionTestContractReader) Start(_ context.Context) error { return nil }

func (pc *protoConversionTestContractReader) Close() error { return nil }

func (pc *protoConversionTestContractReader) Ready() error { panic("unimplemented") }

func (pc *protoConversionTestContractReader) Name() string { panic("unimplemented") }

func (pc *protoConversionTestContractReader) HealthReport() map[string]error { panic("unimplemented") }

func (pc *protoConversionTestContractReader) GetLatestValue(_ context.Context, _ string, _ primitives.ConfidenceLevel, _, _ any) error {
	return nil
}

func (pc *protoConversionTestContractReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, nil
}

func (pc *protoConversionTestContractReader) Bind(_ context.Context, bc []types.BoundContract) error {
	if !reflect.DeepEqual(pc.expectedBindings, bc) {
		return fmt.Errorf("bound contract wasn't parsed properly")
	}
	return nil
}

func (pc *protoConversionTestContractReader) Unbind(_ context.Context, bc []types.BoundContract) error {
	if !reflect.DeepEqual(pc.expectedBindings, bc) {
		return fmt.Errorf("bound contract wasn't parsed properly")
	}

	return nil
}

func (pc *protoConversionTestContractReader) QueryKey(_ context.Context, _ types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, _ any) ([]types.Sequence, error) {
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

func runContractReaderByIDGetLatestValue(t *testing.T) {
	t.Parallel()
	fake := &fakeContractReader{}
	tester := &fakeContractReaderInterfaceTester{impl: fake}
	tester.Setup(t)
	t.Run(
		"Get latest value works with multiple custom contract IDs",
		func(t *testing.T) {
			t.Parallel()
			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContract := BindingsByName(tester.GetBindings(t), AnyContractName)[0]
			anySecondContract := BindingsByName(tester.GetBindings(t), AnySecondContractName)[0]

			anyContractID := "1-" + anyContract.String()
			anySecondContractID := "1-" + anySecondContract.String()

			toBind[anySecondContractID] = anySecondContract
			toBind[anyContractID] = anyContract
			require.NoError(t, cr.Bind(ctx, toBind))

			var primAnyContract, primAnySecondContract uint64
			require.NoError(t, cr.GetLatestValue(ctx, anyContractID, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnyContract))
			require.NoError(t, cr.GetLatestValue(ctx, anySecondContractID, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnySecondContract))

			assert.Equal(t, AnyValueToReadWithoutAnArgument, primAnyContract)
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, primAnySecondContract)
		})

	t.Run(
		"Get latest value works with multiple custom contract IDs and supports same contracts on different addresses",
		func(t *testing.T) {
			t.Parallel()
			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContracts := BindingsByName(tester.GetBindings(t), AnyContractName)
			anyContract1, anyContract2 := anyContracts[0], anyContracts[1]
			anyContractID1, anyContractID2 := "1-"+anyContract1.String(), "2-"+anyContract2.String()
			toBind[anyContractID1], toBind[anyContractID2] = anyContract1, anyContract2

			anySecondContracts := BindingsByName(tester.GetBindings(t), AnySecondContractName)
			anySecondContract1, anySecondContract2 := anySecondContracts[0], anySecondContracts[1]
			anySecondContractID1, anySecondContractID2 := "1-"+anySecondContract1.String(), "2-"+anySecondContract2.String()
			toBind[anySecondContractID1], toBind[anySecondContractID2] = anySecondContract1, anySecondContract2

			require.NoError(t, cr.Bind(ctx, toBind))

			var primAnyContract1, primAnyContract2, primAnySecondContract1, primAnySecondContract2 uint64
			require.NoError(t, cr.GetLatestValue(ctx, anyContractID1, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnyContract1))
			require.NoError(t, cr.GetLatestValue(ctx, anyContractID2, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnyContract2))
			require.NoError(t, cr.GetLatestValue(ctx, anySecondContractID1, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnySecondContract1))
			require.NoError(t, cr.GetLatestValue(ctx, anySecondContractID2, MethodReturningUint64, primitives.Unconfirmed, nil, &primAnySecondContract2))

			assert.Equal(t, AnyValueToReadWithoutAnArgument, primAnyContract1)
			assert.Equal(t, AnyValueToReadWithoutAnArgument, primAnyContract2)
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, primAnySecondContract1)
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, primAnySecondContract2)
		})
}

func runContractReaderByIDBatchGetLatestValues(t *testing.T) {
	t.Parallel()
	fake := &fakeContractReader{}
	tester := &fakeContractReaderInterfaceTester{impl: fake}
	tester.Setup(t)

	t.Run(
		"BatchGetLatestValueByIDs works with multiple custom contract IDs",
		func(t *testing.T) {
			t.Parallel()
			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContract := BindingsByName(tester.GetBindings(t), AnyContractName)[0]
			anyContractID := "1-" + anyContract.String()
			toBind[anyContractID] = anyContract

			anySecondContract := BindingsByName(tester.GetBindings(t), AnySecondContractName)[0]
			anySecondContractID := "1-" + anySecondContract.String()
			toBind[anySecondContractID] = anySecondContract
			require.NoError(t, cr.Bind(ctx, toBind))

			var primitiveReturnValueAnyContract, primitiveReturnValueAnySecondContract uint64
			batchGetLatestValuesRequest := make(chainreader.BatchGetLatestValuesRequestByCustomID)

			batchGetLatestValuesRequest[anyContractID] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnyContract}}
			batchGetLatestValuesRequest[anySecondContractID] = []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract}}

			result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
			require.NoError(t, err)

			anyContractBatch, anySecondContractBatch := result[anyContractID], result[anySecondContractID]
			returnValueAnyContract, errAnyContract := anyContractBatch[0].GetResult()
			returnValueAnySecondContract, errAnySecondContract := anySecondContractBatch[0].GetResult()
			require.NoError(t, errAnyContract)
			require.NoError(t, errAnySecondContract)
			assert.Contains(t, anyContractBatch[0].ReadName, MethodReturningUint64)
			assert.Contains(t, anySecondContractBatch[0].ReadName, MethodReturningUint64)
			assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValueAnyContract.(*uint64))
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, *returnValueAnySecondContract.(*uint64))
		})

	t.Run(
		"BatchGetLatestValueByIDs works with multiple custom contract IDs and supports same contracts on different addresses",
		func(t *testing.T) {
			t.Parallel()
			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContracts := BindingsByName(tester.GetBindings(t), AnyContractName)
			anyContract1, anyContract2 := anyContracts[0], anyContracts[1]
			anyContractID1, anyContractID2 := "1-"+anyContract1.String(), "2-"+anyContract2.String()
			toBind[anyContractID1], toBind[anyContractID2] = anyContract1, anyContract2

			anySecondContracts := BindingsByName(tester.GetBindings(t), AnySecondContractName)
			anySecondContract1, anySecondContract2 := anySecondContracts[0], anySecondContracts[1]
			anySecondContractID1, anySecondContractID2 := "1-"+anySecondContract1.String(), "2-"+anySecondContract2.String()
			toBind[anySecondContractID1], toBind[anySecondContractID2] = anySecondContract1, anySecondContract2

			require.NoError(t, cr.Bind(ctx, toBind))

			var primitiveReturnValueAnyContract1, primitiveReturnValueAnyContract2, primitiveReturnValueAnySecondContract1, primitiveReturnValueAnySecondContract2 uint64
			batchGetLatestValuesRequest := make(chainreader.BatchGetLatestValuesRequestByCustomID)

			anyContract1Req := []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnyContract1}}
			anyContract2Req := []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnyContract2}}
			anySecondContract1Req := []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract1}}
			anySecondContract2Req := []types.BatchRead{{ReadName: MethodReturningUint64, Params: nil, ReturnVal: &primitiveReturnValueAnySecondContract2}}
			batchGetLatestValuesRequest[anyContractID1], batchGetLatestValuesRequest[anyContractID2] = anyContract1Req, anyContract2Req
			batchGetLatestValuesRequest[anySecondContractID1], batchGetLatestValuesRequest[anySecondContractID2] = anySecondContract1Req, anySecondContract2Req

			result, err := cr.BatchGetLatestValues(ctx, batchGetLatestValuesRequest)
			require.NoError(t, err)

			anyContract1Batch, anyContract2Batch := result[anyContractID1], result[anyContractID2]
			anySecondContract1Batch, anySecondContract2Batch := result[anySecondContractID1], result[anySecondContractID2]

			returnValueAnyContract1, errAnyContract1 := anyContract1Batch[0].GetResult()
			returnValueAnyContract2, errAnyContract2 := anyContract2Batch[0].GetResult()
			returnValueAnySecondContract1, errAnySecondContract := anySecondContract1Batch[0].GetResult()
			returnValueAnySecondContract2, errAnySecondContract2 := anySecondContract2Batch[0].GetResult()

			require.NoError(t, errAnyContract1)
			require.NoError(t, errAnyContract2)
			require.NoError(t, errAnySecondContract)
			require.NoError(t, errAnySecondContract2)

			assert.Contains(t, anyContract1Batch[0].ReadName, MethodReturningUint64)
			assert.Contains(t, anyContract2Batch[0].ReadName, MethodReturningUint64)
			assert.Contains(t, anySecondContract1Batch[0].ReadName, MethodReturningUint64)
			assert.Contains(t, anySecondContract2Batch[0].ReadName, MethodReturningUint64)

			assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValueAnyContract1.(*uint64))
			assert.Equal(t, AnyValueToReadWithoutAnArgument, *returnValueAnyContract2.(*uint64))
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, *returnValueAnySecondContract1.(*uint64))
			assert.Equal(t, AnyDifferentValueToReadWithoutAnArgument, *returnValueAnySecondContract2.(*uint64))
		})
}

func runContractReaderByIDQueryKey(t *testing.T) {
	t.Parallel()
	t.Run(
		"QueryKey works with multiple custom contract IDs",
		func(t *testing.T) {
			t.Parallel()
			fake := &fakeContractReader{}
			tester := &fakeContractReaderInterfaceTester{impl: fake}
			tester.Setup(t)

			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContract := BindingsByName(tester.GetBindings(t), AnyContractName)[0]
			anyContractID := "1-" + anyContract.String()
			toBind[anyContractID] = anyContract

			anySecondContract := BindingsByName(tester.GetBindings(t), AnySecondContractName)[0]
			anySecondContractID := "1-" + anySecondContract.String()
			toBind[anySecondContractID] = anySecondContract
			require.NoError(t, cr.Bind(ctx, toBind))

			ts1AnyContract := CreateTestStruct(0, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnyContract, anyContract, types.Unconfirmed)
			ts2AnyContract := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnyContract, anyContract, types.Unconfirmed)

			ts1AnySecondContract := CreateTestStruct(0, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnySecondContract, anySecondContract, types.Unconfirmed)
			ts2AnySecondContract := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnySecondContract, anySecondContract, types.Unconfirmed)

			tsAnyContractType := &TestStruct{}
			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
		})

	t.Run(
		"QueryKey works with multiple custom contract IDs and supports same contracts on different addresses",
		func(t *testing.T) {
			t.Parallel()
			fake := &fakeContractReader{}
			tester := &fakeContractReaderInterfaceTester{impl: fake}
			tester.Setup(t)

			toBind := make(map[string]types.BoundContract)
			ctx := tests.Context(t)
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContract1 := BindingsByName(tester.GetBindings(t), AnyContractName)[0]
			anyContract2 := types.BoundContract{Address: "new-" + anyContract1.Address, Name: anyContract1.Name}
			anyContractID1, anyContractID2 := "1-"+anyContract1.String(), "2-"+anyContract2.String()
			toBind[anyContractID1], toBind[anyContractID2] = anyContract1, anyContract2

			anySecondContract1 := BindingsByName(tester.GetBindings(t), AnySecondContractName)[0]
			anySecondContract2 := types.BoundContract{Address: "new-" + anySecondContract1.Address, Name: anySecondContract1.Name}
			anySecondContractID1, anySecondContractID2 := "1"+"-"+anySecondContract1.String(), "2"+"-"+anySecondContract2.String()
			toBind[anySecondContractID1], toBind[anySecondContractID2] = anySecondContract1, anySecondContract2

			require.NoError(t, cr.Bind(ctx, toBind))

			ts1AnyContract1 := CreateTestStruct(0, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnyContract1, anyContract1, types.Unconfirmed)
			ts2AnyContract1 := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnyContract1, anyContract1, types.Unconfirmed)
			ts1AnyContract2 := CreateTestStruct(2, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnyContract2, anyContract2, types.Unconfirmed)
			ts2AnyContract2 := CreateTestStruct(3, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnyContract2, anyContract2, types.Unconfirmed)

			ts1AnySecondContract1 := CreateTestStruct(4, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnySecondContract1, anySecondContract1, types.Unconfirmed)
			ts2AnySecondContract1 := CreateTestStruct(5, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnySecondContract1, anySecondContract1, types.Unconfirmed)
			ts1AnySecondContract2 := CreateTestStruct(6, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts1AnySecondContract2, anySecondContract2, types.Unconfirmed)
			ts2AnySecondContract2 := CreateTestStruct(7, tester)
			_ = SubmitTransactionToCW(t, tester, MethodTriggeringEvent, ts2AnySecondContract2, anySecondContract2, types.Unconfirmed)

			tsAnyContractType := &TestStruct{}
			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID1, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract1, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract1, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID2, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract2, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract2, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anySecondContractID1, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract1, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract1, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			assert.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anySecondContractID2, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract2, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract2, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
		})
}
