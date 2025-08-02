package contractreader_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
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
	codecpb "github.com/smartcontractkit/chainlink-common/pkg/internal/codec"
	contractreadertest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader/test"
	loopjson "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/json"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/values"

	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
	
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"net"
)

func extractParamValue[T any](params any, fieldName string, target T) error {
	if typed, ok := params.(T); ok {
		reflect.ValueOf(target).Elem().Set(reflect.ValueOf(typed).Elem())
		return nil
	}

	if m, ok := params.(map[string]interface{}); ok {
		if fieldName == "" {
			data, err := json.Marshal(m)
			if err != nil {
				return fmt.Errorf("failed to marshal map: %w", err)
			}
			return loopjson.UnmarshalJson(data, target)
		}

		if val, exists := m[fieldName]; exists {
			data, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal field %s: %w", fieldName, err)
			}
			return loopjson.UnmarshalJson(data, target)
		}
		return fmt.Errorf("field %s not found in params", fieldName)
	}

	return fmt.Errorf("unexpected params type: %T", params)
}

func TestVersionedBytesFunctions(t *testing.T) {
	const unsupportedVer = 25913
	t.Run("EncodeVersionedBytes unsupported type", func(t *testing.T) {
		invalidData := make(chan int)

		_, err := codecpb.EncodeVersionedBytes(invalidData, codecpb.JSONEncodingVersion2)

		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := fmt.Errorf("%w: unsupported encoding version %d for data map[key:value]", types.ErrInvalidEncoding, unsupportedVer)
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := codecpb.EncodeVersionedBytes(data, unsupportedVer)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := fmt.Errorf("unsupported encoding version %d for versionedData [97 98 99 100 102]", unsupportedVer)
		versionedBytes := &codecpb.VersionedBytes{
			Version: unsupportedVer, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := codecpb.DecodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("DecodeVersionedBytes if nil returns error", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := errors.New("cannot decode nil versioned bytes")

		err := codecpb.DecodeVersionedBytes(&decodedData, nil)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestContractReaderInterfaceTests(t *testing.T) {
	t.Parallel()

	fake := &fakeContractReader{}
	RunContractReaderInterfaceTests(
		t,
		contractreadertest.WrapContractReaderTesterForLoop(
			&fakeContractReaderInterfaceTester{impl: fake},
		),
		true,
		false,
	)
}

func TestContractReaderByIDWrapper(t *testing.T) {
	t.Parallel()
	t.Run("Contract Reader by ID GetLatestValue", runContractReaderByIDGetLatestValue)
	t.Run("Contract Reader by ID BatchGetLatestValues", runContractReaderByIDBatchGetLatestValues)
	t.Run("Contract Reader by ID QueryKey", runContractReaderByIDQueryKey)
}

func TestBind(t *testing.T) {
	t.Parallel()

	es := &errContractReader{}
	errTester := contractreadertest.WrapContractReaderTesterForLoop(
		&fakeContractReaderInterfaceTester{impl: es},
	)

	errTester.Setup(t)
	contractReader := errTester.GetContractReader(t)

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("Bind unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := t.Context()
			err := contractReader.Bind(ctx, []types.BoundContract{{Name: "Contract", Address: "address"}})
			assert.True(t, errors.Is(err, errorType))
		})

		t.Run("Unbind unwraps errors from server"+errorType.Error(), func(t *testing.T) {
			ctx := t.Context()
			err := contractReader.Unbind(ctx, []types.BoundContract{{Name: "Contract", Address: "address"}})
			assert.True(t, errors.Is(err, errorType))
		})
	}
}

func TestGetLatestValue(t *testing.T) {
	t.Parallel()

	es := &errContractReader{}
	errTester := contractreadertest.WrapContractReaderTesterForLoop(
		&fakeContractReaderInterfaceTester{impl: es},
	)

	errTester.Setup(t)
	contractReader := errTester.GetContractReader(t)

	t.Run("nil reader should return unimplemented", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
		nilTester.Setup(t)
		nilCr := nilTester.GetContractReader(t)

		err := nilCr.GetLatestValue(ctx, "method", primitives.Unconfirmed, "anything", "anything")
		assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
	})

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := t.Context()
			err := contractReader.GetLatestValue(ctx, "method", primitives.Unconfirmed, nil, "anything")
			assert.True(t, errors.Is(err, errorType))
		})
	}

	// make sure that errors come from client directly
	es.err = nil
	t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		ctx := t.Context()
		err := contractReader.GetLatestValue(ctx, "method", primitives.Unconfirmed, &cannotEncode{}, &TestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

func TestBatchGetLatestValues(t *testing.T) {
	t.Parallel()

	es := &errContractReader{}
	errTester := contractreadertest.WrapContractReaderTesterForLoop(
		&fakeContractReaderInterfaceTester{impl: es},
	)

	errTester.Setup(t)
	contractReader := errTester.GetContractReader(t)

	t.Run("nil reader should return unimplemented", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
		nilTester.Setup(t)
		nilCr := nilTester.GetContractReader(t)

		_, err := nilCr.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
		assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
	})

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("BatchGetLatestValues unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := t.Context()
			_, err := contractReader.BatchGetLatestValues(ctx, types.BatchGetLatestValuesRequest{})
			assert.True(t, errors.Is(err, errorType))
		})
	}

	// make sure that errors come from client directly
	es.err = nil
	t.Run("BatchGetLatestValues returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		ctx := t.Context()
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

func TestQueryKey(t *testing.T) {
	t.Parallel()

	impl := &protoConversionTestContractReader{}
	crTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: impl})
	crTester.Setup(t)
	cr := crTester.GetContractReader(t)

	es := &errContractReader{}
	errTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: es})
	errTester.Setup(t)
	contractReader := errTester.GetContractReader(t)

	t.Run("nil reader should return unimplemented", func(t *testing.T) {
		ctx := t.Context()

		nilTester := contractreadertest.WrapContractReaderTesterForLoop(&fakeContractReaderInterfaceTester{impl: nil})
		nilTester.Setup(t)
		nilCr := nilTester.GetContractReader(t)

		_, err := nilCr.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
		assert.Equal(t, codes.Unimplemented, status.Convert(err).Code())
	})

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("QueryKey unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			ctx := t.Context()
			_, err := contractReader.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &[]interface{}{nil})
			assert.True(t, errors.Is(err, errorType))
		})
	}

	t.Run("test QueryKey proto conversion", func(t *testing.T) {
		for _, tc := range generateQueryFilterTestCases(t) {
			impl.expectedQueryFilter = tc
			filter, err := query.Where(tc.Key, tc.Expressions...)
			require.NoError(t, err)
			_, err = cr.QueryKey(t.Context(), types.BoundContract{}, filter, query.LimitAndSort{}, &[]interface{}{nil})
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

type fakeContractReaderInterfaceTester struct {
	interfaceTesterBase
	TestSelectionSupport
	impl types.ContractReader
	cw   fakeContractWriter
}

func (it *fakeContractReaderInterfaceTester) Setup(_ *testing.T) {
	fake, ok := it.impl.(*fakeContractReader)
	if ok {
		fake.vals = make(map[string][]valConfidencePair)
		fake.triggers = newEventsRecorder()
		fake.stored = make(map[string][]TestStruct)
	}
}

func (it *fakeContractReaderInterfaceTester) GetContractReader(_ *testing.T) types.ContractReader {
	return it.impl
}

func (it *fakeContractReaderInterfaceTester) GetContractWriter(_ *testing.T) types.ContractWriter {
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

type dynamicTopicEventConfidencePair struct {
	someDynamicTopicEvent SomeDynamicTopicEvent
	confidenceLevel       primitives.ConfidenceLevel
}
type event struct {
	contractID      string
	event           any
	confidenceLevel primitives.ConfidenceLevel
	eventType       string
}

type eventsRecorder struct {
	mux    sync.Mutex
	events []event
}

func newEventsRecorder() *eventsRecorder {
	return &eventsRecorder{}
}

func (e *eventsRecorder) RecordEvent(contractID string, evt any, confidenceLevel primitives.ConfidenceLevel, eventType string) error {
	e.mux.Lock()
	defer e.mux.Unlock()

	switch eventType {
	case EventName:
		_, ok := evt.(TestStruct)
		if !ok {
			return fmt.Errorf("unexpected event type %T", evt)
		}
	case DynamicTopicEventName:
		_, ok := evt.(SomeDynamicTopicEvent)
		if !ok {
			return fmt.Errorf("unexpected event type %T", evt)
		}

	}

	e.events = append(e.events, event{contractID: contractID, event: evt, confidenceLevel: confidenceLevel, eventType: eventType})

	return nil
}

func (e *eventsRecorder) setConfidenceLevelOnAllEvents(confidenceLevel primitives.ConfidenceLevel) {
	e.mux.Lock()
	defer e.mux.Unlock()

	for i := range e.events {
		e.events[i].confidenceLevel = confidenceLevel
	}
}

func (e *eventsRecorder) getEvents(filters ...func(event) bool) []event {
	e.mux.Lock()
	defer e.mux.Unlock()

	events := make([]event, 0)
	for _, event := range e.events {
		match := true
		for _, filter := range filters {
			if !filter(event) {
				match = false
				break
			}
		}
		if match {
			events = append(events, event)
		}
	}

	return events
}

type fakeContractReader struct {
	types.UnimplementedContractReader
	vals        map[string][]valConfidencePair
	triggers    *eventsRecorder
	stored      map[string][]TestStruct
	batchStored BatchCallEntry
	lock        sync.Mutex
}

type fakeContractWriter struct {
	types.ContractWriter
	cr *fakeContractReader
}

func (f *fakeContractWriter) SubmitTransaction(_ context.Context, contractName, method string, args any, transactionID string, toAddress string, meta *types.TxMeta, value *big.Int) error {
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
		if err := f.cr.triggers.RecordEvent(contractID, args, primitives.Unconfirmed, EventName); err != nil {
			return fmt.Errorf("failed to record event: %w", err)
		}
	case MethodTriggeringEventWithDynamicTopic:
		if err := f.cr.triggers.RecordEvent(contractID, args, primitives.Unconfirmed, DynamicTopicEventName); err != nil {
			return fmt.Errorf("failed to record event: %w", err)
		}
	case "batchContractWrite":
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

func (f *fakeContractWriter) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	return types.Finalized, nil
}

func (f *fakeContractWriter) GetFeeComponents(ctx context.Context) (*types.ChainFeeComponents, error) {
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
		vals := f.vals[contractName]
		for i := len(vals) - 1; i >= 0; i-- {
			if vals[i].confidenceLevel == confidenceLevel {
				return setReturnValue(returnVal, vals[i].val)
			}
		}
		return fmt.Errorf("%w: no val with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if strings.HasSuffix(readIdentifier, MethodReturningUint64) {
		var value uint64
		if strings.Contains(readIdentifier, "-"+AnyContractName+"-") {
			value = AnyValueToReadWithoutAnArgument
		} else {
			value = AnyDifferentValueToReadWithoutAnArgument
		}
		return setReturnValue(returnVal, value)
	} else if strings.HasSuffix(readIdentifier, MethodReturningUint64Slice) {
		return setReturnValue(returnVal, AnySliceToReadWithoutAnArgument)
	} else if strings.HasSuffix(readIdentifier, MethodReturningSeenStruct) {
		var pv TestStruct
		if err := extractParamValue(params, "", &pv); err != nil {
			return fmt.Errorf("failed to extract TestStruct params: %w", err)
		}

		value := TestStructWithExtraField{
			TestStruct: pv,
			ExtraField: AnyExtraValue,
		}
		value.AccountStruct = AccountStruct{
			Account:    anyAccountBytes,
			AccountStr: anyAccountString,
		}
		value.BigField = big.NewInt(2)
		return setReturnValue(returnVal, value)
	} else if strings.HasSuffix(readIdentifier, EventName) {
		f.lock.Lock()
		defer f.lock.Unlock()

		events := f.triggers.getEvents(func(e event) bool {
			return e.contractID == contractName && e.eventType == EventName
		})

		if len(events) == 0 {
			return types.ErrNotFound
		}

		for i := len(events) - 1; i >= 0; i-- {
			if events[i].confidenceLevel == confidenceLevel {
				return setReturnValue(returnVal, events[i].event.(TestStruct))
			}
		}

		return fmt.Errorf("%w: no event with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if strings.HasSuffix(readIdentifier, EventWithFilterName) {
		f.lock.Lock()
		defer f.lock.Unlock()

		var param FilterEventParams
		if err := extractParamValue(params, "", &param); err != nil {
			return fmt.Errorf("failed to extract FilterEventParams: %w", err)
		}

		triggers := f.triggers.getEvents(func(e event) bool { return e.contractID == contractName && e.eventType == EventName })
		for i := len(triggers) - 1; i >= 0; i-- {
			testStruct := triggers[i].event.(TestStruct)
			if *testStruct.Field == param.Field {
				return setReturnValue(returnVal, testStruct)
			}
		}
		return types.ErrNotFound
	} else if strings.HasSuffix(readIdentifier, DynamicTopicEventName) {
		f.lock.Lock()
		defer f.lock.Unlock()

		triggers := f.triggers.getEvents(func(e event) bool { return e.contractID == contractName && e.eventType == DynamicTopicEventName })

		if len(triggers) == 0 {
			return types.ErrNotFound
		}

		for i := len(triggers) - 1; i >= 0; i-- {
			if triggers[i].confidenceLevel == confidenceLevel {
				return setReturnValue(returnVal, triggers[i].event.(SomeDynamicTopicEvent))
			}
		}

		return fmt.Errorf("%w: no event with %s confidence was found ", types.ErrNotFound, confidenceLevel)
	} else if !strings.HasSuffix(readIdentifier, MethodTakingLatestParamsReturningTestStruct) {
		return errors.New("unknown method " + readIdentifier)
	}

	f.lock.Lock()
	defer f.lock.Unlock()
	stored := f.stored[contractName]

	var lp LatestParams
	if err := extractParamValue(params, "", &lp); err != nil {
		return fmt.Errorf("failed to extract LatestParams: %w", err)
	}

	if lp.I-1 >= len(stored) {
		return errors.New("latest params index out of bounds for stored test structs")
	}

	_, isValue := returnVal.(*values.Value)
	if isValue {
		wrapped, err := values.Wrap(stored[lp.I-1])
		if err != nil {
			return err
		}
		return setReturnValue(returnVal, wrapped)
	}

	return setReturnValue(returnVal, stored[lp.I-1])
}

func (f *fakeContractReader) GetLatestValueWithHeadData(_ context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) (*types.Head, error) {
	err := f.GetLatestValue(context.Background(), readIdentifier, confidenceLevel, params, returnVal)
	if err != nil {
		return nil, err
	}

	return &types.Head{}, nil
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
				var value uint64
				if requestContract.Name == AnyContractName {
					value = AnyValueToReadWithoutAnArgument
				} else {
					value = AnyDifferentValueToReadWithoutAnArgument
				}
				if err := setReturnValue(req.ReturnVal, value); err != nil {
					return nil, err
				}
				returnVal = req.ReturnVal
			} else if req.ReadName == MethodReturningUint64Slice {
				value := AnySliceToReadWithoutAnArgument
				if err := setReturnValue(req.ReturnVal, value); err != nil {
					return nil, err
				}
				returnVal = req.ReturnVal
			} else if req.ReadName == MethodReturningSeenStruct {
				var ts TestStruct
				if err := extractParamValue(req.Params, "", &ts); err != nil {
					return nil, fmt.Errorf("failed to extract TestStruct params: %w", err)
				}
				ts.AccountStruct = AccountStruct{
					Account:    anyAccountBytes,
					AccountStr: anyAccountString,
				}
				ts.BigField = big.NewInt(2)
				value := TestStructWithExtraField{
					TestStruct: ts,
					ExtraField: AnyExtraValue,
				}
				if err := setReturnValue(req.ReturnVal, value); err != nil {
					return nil, err
				}
				returnVal = req.ReturnVal
			} else if req.ReadName == MethodTakingLatestParamsReturningTestStruct {
				var latestParams LatestParams
				if err := extractParamValue(req.Params, "", &latestParams); err != nil {
					return nil, fmt.Errorf("failed to extract LatestParams: %w", err)
				}
				if latestParams.I <= 0 {
					err = fmt.Errorf("invalid param %d", latestParams.I)
					if setErr := setReturnValue(req.ReturnVal, &LatestParams{}); setErr != nil {
						return nil, setErr
					}
					returnVal = req.ReturnVal
				} else {
					value := storedContractBatch[latestParams.I-1].ReturnValue
					if setErr := setReturnValue(req.ReturnVal, value); setErr != nil {
						return nil, setErr
					}
					returnVal = req.ReturnVal
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

func (f *fakeContractReader) QueryKey(ctx context.Context, bc types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceType any) ([]types.Sequence, error) {
	seqsIter, err := f.QueryKeys(ctx, []types.ContractKeyFilter{{
		KeyFilter:        filter,
		Contract:         bc,
		SequenceDataType: sequenceType,
	}}, limitAndSort)

	if err != nil {
		return nil, err
	}

	if seqsIter != nil {
		var seqs []types.Sequence
		for _, seq := range seqsIter {
			seqs = append(seqs, seq)
		}

		return seqs, nil
	}

	return nil, nil
}

type sequenceWithEventType struct {
	eventType string
	sequence  types.Sequence
}

func (f *fakeContractReader) QueryKeys(_ context.Context, filters []types.ContractKeyFilter, limitAndSort query.LimitAndSort) (iter.Seq2[string, types.Sequence], error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	supportedEventTypes := map[string]struct{}{EventName: {}, DynamicTopicEventName: {}}

	for _, filter := range filters {
		if _, ok := supportedEventTypes[filter.Key]; !ok {
			return nil, fmt.Errorf("unsupported event type %s", filter.Key)
		}
	}

	if len(filters) > 1 {
		fmt.Printf("filters: %v\n", filters)
	}

	isValueType := false
	eventTypeToFilter := map[string]types.ContractKeyFilter{}
	for _, filter := range filters {
		eventTypeToFilter[filter.Key] = filter
		_, isValueType = filter.SequenceDataType.(*values.Value)
	}

	events := f.triggers.getEvents(func(e event) bool {
		filter := eventTypeToFilter[e.eventType]

		if e.contractID != filter.Contract.String() {
			return false
		}
		_, filterExistsForType := eventTypeToFilter[e.eventType]

		return filterExistsForType
	})

	var sequences []sequenceWithEventType
	for idx, trigger := range events {
		filter := eventTypeToFilter[trigger.eventType]

		doAppend := true
		for _, expr := range filter.Expressions {
			if primitive, ok := expr.Primitive.(*primitives.Comparator); ok {
				if len(primitive.ValueComparators) == 0 {
					return nil, fmt.Errorf("value comparator for %s should not be empty", primitive.Name)
				}
				if primitive.Name == "Field" {
					for _, valComp := range primitive.ValueComparators {
						// Handle both direct int32 pointers and map[string]interface{} from JSON
						var fieldValue int32
						if valPtr, ok := valComp.Value.(*int32); ok {
							fieldValue = *valPtr
						} else if val, ok := valComp.Value.(int32); ok {
							// Handle direct int32 values (from type hints)
							fieldValue = val
						} else if val, ok := valComp.Value.(int); ok {
							// Handle direct int values (from test)
							fieldValue = int32(val)
						} else if m, ok := valComp.Value.(map[string]interface{}); ok {
							// For JSON deserialized values
							if err := extractParamValue(m, "", &fieldValue); err != nil {
								return nil, fmt.Errorf("failed to extract int32 from comparator: %w", err)
							}
						} else if num, ok := valComp.Value.(json.Number); ok {
							// Handle json.Number from UseNumber decoding
							val, err := num.Int64()
							if err != nil {
								return nil, fmt.Errorf("failed to convert json.Number to int: %w", err)
							}
							fieldValue = int32(val)
						} else {
							return nil, fmt.Errorf("unexpected comparator value type: %T", valComp.Value)
						}
						doAppend = doAppend && Compare(*trigger.event.(TestStruct).Field, fieldValue, valComp.Operator)
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

		if (len(eventTypeToFilter[trigger.eventType].Expressions) == 0 || doAppend) && !skipAppend {
			if isValueType {
				value, err := values.Wrap(trigger.event)
				if err != nil {
					return nil, err
				}
				sequences = append(sequences, sequenceWithEventType{eventType: trigger.eventType, sequence: types.Sequence{Cursor: strconv.Itoa(idx), TxHash: []byte("0xtest"), Data: &value}})
			} else {
				sequences = append(sequences, sequenceWithEventType{eventType: trigger.eventType, sequence: types.Sequence{Cursor: fmt.Sprintf("%d", idx), TxHash: []byte("0xtest"), Data: trigger.event}})
			}
		}

		if limitAndSort.Limit.Count > 0 && len(sequences) >= int(limitAndSort.Limit.Count) {
			break
		}
	}

	if isValueType {
		if !limitAndSort.HasSequenceSort() && !limitAndSort.HasCursorLimit() {
			sort.Slice(sequences, func(i, j int) bool {
				valI := *sequences[i].sequence.Data.(*values.Value)
				valJ := *sequences[j].sequence.Data.(*values.Value)

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
			if len(eventTypeToFilter) == 1 {
				if _, ok := eventTypeToFilter[EventName]; ok {
					sort.Slice(sequences, func(i, j int) bool {
						if sequences[i].sequence.Data.(TestStruct).Field == nil || sequences[j].sequence.Data.(TestStruct).Field == nil {
							return false
						}
						return *sequences[i].sequence.Data.(TestStruct).Field > *sequences[j].sequence.Data.(TestStruct).Field
					})
				}
			}
		}
	}

	return func(yield func(string, types.Sequence) bool) {
		for _, s := range sequences {
			if !yield(s.eventType, s.sequence) {
				return
			}
		}
	}, nil

}

func (f *fakeContractReader) GenerateBlocksTillConfidenceLevel(_ *testing.T, _, _ string, confidenceLevel primitives.ConfidenceLevel) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for contractID, vals := range f.vals {
		for i, val := range vals {
			f.vals[contractID][i] = valConfidencePair{val: val.val, confidenceLevel: confidenceLevel}
		}
	}

	f.triggers.setConfidenceLevelOnAllEvents(confidenceLevel)
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
	if !equalKeyFilters(pc.expectedQueryFilter, filter) {
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

// equalKeyFilters compares two KeyFilter instances for semantic equality,
// handling pointer values correctly (comparing dereferenced values rather than addresses)
func equalKeyFilters(a, b query.KeyFilter) bool {
	if a.Key != b.Key {
		return false
	}

	if len(a.Expressions) != len(b.Expressions) {
		return false
	}

	for i := range a.Expressions {
		if !equalExpressions(a.Expressions[i], b.Expressions[i]) {
			return false
		}
	}

	return true
}

func equalExpressions(a, b query.Expression) bool {
	// Check if both are primitive or both are boolean expressions
	if a.IsPrimitive() != b.IsPrimitive() {
		return false
	}

	if a.IsPrimitive() {
		return equalPrimitives(a.Primitive, b.Primitive)
	}

	// Both are boolean expressions
	if a.BoolExpression.BoolOperator != b.BoolExpression.BoolOperator {
		return false
	}

	if len(a.BoolExpression.Expressions) != len(b.BoolExpression.Expressions) {
		return false
	}

	for i := range a.BoolExpression.Expressions {
		if !equalExpressions(a.BoolExpression.Expressions[i], b.BoolExpression.Expressions[i]) {
			return false
		}
	}

	return true
}

func equalPrimitives(a, b primitives.Primitive) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Check if they're the same type
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch ap := a.(type) {
	case *primitives.Comparator:
		bp := b.(*primitives.Comparator)
		if ap.Name != bp.Name {
			return false
		}
		if len(ap.ValueComparators) != len(bp.ValueComparators) {
			return false
		}
		for i := range ap.ValueComparators {
			if ap.ValueComparators[i].Operator != bp.ValueComparators[i].Operator {
				return false
			}
			// Compare values, handling pointers correctly
			if !equalValues(ap.ValueComparators[i].Value, bp.ValueComparators[i].Value) {
				return false
			}
		}
		return true
	case *primitives.Confidence:
		bp := b.(*primitives.Confidence)
		return ap.ConfidenceLevel == bp.ConfidenceLevel
	case *primitives.Block:
		bp := b.(*primitives.Block)
		return ap.Block == bp.Block && ap.Operator == bp.Operator
	case *primitives.Timestamp:
		bp := b.(*primitives.Timestamp)
		return ap.Timestamp == bp.Timestamp && ap.Operator == bp.Operator
	case *primitives.TxHash:
		bp := b.(*primitives.TxHash)
		return ap.TxHash == bp.TxHash
	default:
		// Fall back to DeepEqual for unknown types
		return reflect.DeepEqual(a, b)
	}
}

func equalValues(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle pointers by dereferencing
	aVal := reflect.ValueOf(a)
	bVal := reflect.ValueOf(b)

	// If both are pointers, dereference and compare values
	if aVal.Kind() == reflect.Ptr && bVal.Kind() == reflect.Ptr {
		if aVal.IsNil() && bVal.IsNil() {
			return true
		}
		if aVal.IsNil() || bVal.IsNil() {
			return false
		}
		// Compare dereferenced values
		return reflect.DeepEqual(aVal.Elem().Interface(), bVal.Elem().Interface())
	}

	// If one is a pointer and the other isn't, compare the dereferenced value
	if aVal.Kind() == reflect.Ptr && bVal.Kind() != reflect.Ptr {
		if aVal.IsNil() {
			return false
		}
		return reflect.DeepEqual(aVal.Elem().Interface(), b)
	}
	if aVal.Kind() != reflect.Ptr && bVal.Kind() == reflect.Ptr {
		if bVal.IsNil() {
			return false
		}
		return reflect.DeepEqual(a, bVal.Elem().Interface())
	}

	// If types don't match exactly, it's not equal
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	// For non-pointer types, use DeepEqual
	return reflect.DeepEqual(a, b)
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
			ctx := t.Context()
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
			ctx := t.Context()
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
			ctx := t.Context()
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
			ctx := t.Context()
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
			fakeCW := &fakeContractWriter{cr: fake}
			tester := &fakeContractReaderInterfaceTester{impl: fake}
			tester.Setup(t)

			toBind := make(map[string]types.BoundContract)
			ctx := t.Context()
			cr := chainreader.WrapContractReaderByIDs(tester.GetContractReader(t))

			anyContract := BindingsByName(tester.GetBindings(t), AnyContractName)[0]
			anyContractID := "1-" + anyContract.String()
			toBind[anyContractID] = anyContract

			anySecondContract := BindingsByName(tester.GetBindings(t), AnySecondContractName)[0]
			anySecondContractID := "1-" + anySecondContract.String()
			toBind[anySecondContractID] = anySecondContract
			require.NoError(t, cr.Bind(ctx, toBind))

			ts1AnyContract := CreateTestStruct(0, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnyContract, anyContract, types.Unconfirmed)
			ts2AnyContract := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnyContract, anyContract, types.Unconfirmed)

			ts1AnySecondContract := CreateTestStruct(0, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnySecondContract, anySecondContract, types.Unconfirmed)
			ts2AnySecondContract := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnySecondContract, anySecondContract, types.Unconfirmed)

			tsAnyContractType := &TestStruct{}
			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract, sequences[0].Data) && assert.Equal(t, []byte("0xtest"), sequences[0].TxHash) && assert.Equal(t, []byte("0xtest"), sequences[1].TxHash)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
		})

	t.Run(
		"QueryKey works with multiple custom contract IDs and supports same contracts on different addresses",
		func(t *testing.T) {
			t.Parallel()
			fake := &fakeContractReader{}
			fakeCW := &fakeContractWriter{cr: fake}

			tester := &fakeContractReaderInterfaceTester{impl: fake}
			tester.Setup(t)

			toBind := make(map[string]types.BoundContract)
			ctx := t.Context()
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
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnyContract1, anyContract1, types.Unconfirmed)
			ts2AnyContract1 := CreateTestStruct(1, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnyContract1, anyContract1, types.Unconfirmed)
			ts1AnyContract2 := CreateTestStruct(2, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnyContract2, anyContract2, types.Unconfirmed)
			ts2AnyContract2 := CreateTestStruct(3, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnyContract2, anyContract2, types.Unconfirmed)

			ts1AnySecondContract1 := CreateTestStruct(4, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnySecondContract1, anySecondContract1, types.Unconfirmed)
			ts2AnySecondContract1 := CreateTestStruct(5, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnySecondContract1, anySecondContract1, types.Unconfirmed)
			ts1AnySecondContract2 := CreateTestStruct(6, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts1AnySecondContract2, anySecondContract2, types.Unconfirmed)
			ts2AnySecondContract2 := CreateTestStruct(7, tester)
			_ = SubmitTransactionToCW(t, tester, fakeCW, MethodTriggeringEvent, ts2AnySecondContract2, anySecondContract2, types.Unconfirmed)

			tsAnyContractType := &TestStruct{}
			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID1, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract1, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract1, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anyContractID2, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnyContract2, sequences[1].Data) && reflect.DeepEqual(ts2AnyContract2, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)

			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anySecondContractID1, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract1, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract1, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
			require.Eventually(t, func() bool {
				sequences, err := cr.QueryKey(ctx, anySecondContractID2, query.KeyFilter{Key: EventName}, query.LimitAndSort{}, tsAnyContractType)
				return err == nil && len(sequences) == 2 && reflect.DeepEqual(ts1AnySecondContract2, sequences[1].Data) && reflect.DeepEqual(ts2AnySecondContract2, sequences[0].Data)
			}, tester.MaxWaitTimeForEvents(), time.Millisecond*10)
		})
}

func setReturnValue(returnVal any, value any) error {
	if anyRetVal, ok := returnVal.(*any); ok {
		*anyRetVal = value
		return nil
	}

	switch rv := returnVal.(type) {
	case *TestStruct:
		if v, ok := value.(TestStruct); ok {
			*rv = v
			return nil
		}
	case *TestStructWithExtraField:
		if v, ok := value.(TestStructWithExtraField); ok {
			*rv = v
			return nil
		}
	case *SomeDynamicTopicEvent:
		if v, ok := value.(SomeDynamicTopicEvent); ok {
			*rv = v
			return nil
		}
	case *uint64:
		if v, ok := value.(uint64); ok {
			*rv = v
			return nil
		}
	case *[]uint64:
		if v, ok := value.([]uint64); ok {
			*rv = v
			return nil
		}
	}

	return fmt.Errorf("unexpected return value type: %T", returnVal)
}

// Test case for debugging *[]byte parameter handling
func TestContractReader_PointerToByteSliceParams(t *testing.T) {
	t.Run("GetLatestValue with *[]byte params", func(t *testing.T) {
		// Create test contract reader
		testCR := &testPointerParamsContractReader{t: t}
		
		// Create client/server through LOOP
		server := contractreader.NewServer(testCR)
		grpcServer := grpc.NewServer()
		pb.RegisterContractReaderServer(grpcServer, server)
		
		// Create in-memory connection
		lis := bufconn.Listen(1024 * 1024)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				t.Logf("Server exited with error: %v", err)
			}
		}()
		t.Cleanup(grpcServer.Stop)
		
		// Create client connection
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() })
		
		// Create contract reader client
		grpcClient := pb.NewContractReaderClient(conn)
		client := contractreader.NewClient(brokerExt{}, grpcClient)
		
		// Test: Pass *[]byte as params (simulating chainlink-aptos behavior)
		jsonData := []byte(`{"key":"value","number":123}`)
		params := &jsonData // This is *[]byte
		t.Logf("Client sending params of type: %T", params)
		
		var result []byte
		err = client.GetLatestValue(ctx, "test-identifier", primitives.Finalized, params, &result)
		require.NoError(t, err)
		
		// Verify the server received the correct type
		require.True(t, testCR.gotCorrectType, "Server should have received *[]byte")
	})
	
	t.Run("BatchGetLatestValues with *[]byte params", func(t *testing.T) {
		// Create test contract reader
		testCR := &testPointerParamsContractReader{t: t, batchMode: true}
		
		// Create client/server through LOOP
		server := contractreader.NewServer(testCR)
		grpcServer := grpc.NewServer()
		pb.RegisterContractReaderServer(grpcServer, server)
		
		// Create in-memory connection
		lis := bufconn.Listen(1024 * 1024)
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				t.Logf("Server exited with error: %v", err)
			}
		}()
		t.Cleanup(grpcServer.Stop)
		
		// Create client connection
		ctx := context.Background()
		conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		t.Cleanup(func() { conn.Close() })
		
		// Create contract reader client
		grpcClient := pb.NewContractReaderClient(conn)
		client := contractreader.NewClient(brokerExt{}, grpcClient)
		
		// Test: Pass *[]byte as params in batch request
		jsonData := []byte(`{"key":"value","number":456}`)
		params := &jsonData // This is *[]byte
		
		var result []byte
		request := types.BatchGetLatestValuesRequest{
			types.BoundContract{Address: "0x123", Name: "test"}: []types.BatchRead{
				{
					ReadName:  "testRead",
					Params:    params,
					ReturnVal: &result,
				},
			},
		}
		
		_, err = client.BatchGetLatestValues(ctx, request)
		require.NoError(t, err)
		
		// Verify the server received the correct type
		require.True(t, testCR.gotCorrectType, "Server should have received *[]byte in batch")
	})
}

// Test contract reader that checks if it receives *[]byte params
type testPointerParamsContractReader struct {
	types.UnimplementedContractReader
	t              *testing.T
	gotCorrectType bool
	batchMode      bool
}

func (t *testPointerParamsContractReader) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, retVal any) error {
	t.t.Logf("GetLatestValue called with params type: %T, value: %v", params, params)
	
	// Log detailed type information
	if params != nil {
		t.t.Logf("Params is nil: %v", params == nil)
		t.t.Logf("Type of params: %T", params)
		
		// Try different type assertions
		if p, ok := params.(*[]byte); ok {
			t.gotCorrectType = true
			t.t.Logf("Successfully asserted as *[]byte, value: %v", *p)
		} else if p, ok := params.([]byte); ok {
			t.t.Logf("ERROR: Received []byte instead of *[]byte, value: %v", p)
		} else if p, ok := params.(*[]uint8); ok {
			t.gotCorrectType = true
			t.t.Logf("Successfully asserted as *[]uint8 (same as *[]byte), value: %v", *p)
		} else if p, ok := params.([]uint8); ok {
			t.t.Logf("ERROR: Received []uint8 instead of *[]uint8, value: %v", p)
		} else {
			t.t.Logf("ERROR: Could not assert to any byte slice type")
		}
	}
	
	// Return some dummy data
	if retValPtr, ok := retVal.(*[]byte); ok {
		*retValPtr = []byte(`{"result":"success"}`)
	}
	
	return nil
}

func (t *testPointerParamsContractReader) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	t.t.Log("BatchGetLatestValues called")
	
	result := make(types.BatchGetLatestValuesResult)
	
	for contract, reads := range request {
		batchResults := make([]types.BatchReadResult, len(reads))
		
		for i, read := range reads {
			t.t.Logf("Batch read %d params type: %T", i, read.Params)
			
			// Check if we received *[]byte
			if _, ok := read.Params.(*[]byte); ok {
				t.gotCorrectType = true
				t.t.Log("Batch: Received correct type: *[]byte")
			} else {
				t.t.Logf("Batch ERROR: Expected *[]byte but got %T", read.Params)
			}
			
			// Set dummy result
			if retValPtr, ok := read.ReturnVal.(*[]byte); ok {
				*retValPtr = []byte(`{"batch_result":"success"}`)
			}
			
			batchResults[i] = types.BatchReadResult{
				ReadName: read.ReadName,
			}
			batchResults[i].SetResult(read.ReturnVal, nil)
		}
		
		result[contract] = batchResults
	}
	
	return result, nil
}

// Minimal broker extension for testing
type brokerExt struct {
	types.UnimplementedContractReader
}

func (b brokerExt) ClientConn() grpc.ClientConnInterface { return nil }
func (b brokerExt) Close() error { return nil }
func (b brokerExt) HealthReport() map[string]error { return nil }
func (b brokerExt) Name() string { return "test" }
func (b brokerExt) Ready() error { return nil }
func (b brokerExt) Start(ctx context.Context) error { return nil }
