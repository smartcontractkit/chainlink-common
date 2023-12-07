package internal

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

func TestVersionedBytesFunctions(t *testing.T) {
	const unsupportedVer = 25913
	t.Run("EncodeVersionedBytes unsupported type", func(t *testing.T) {
		invalidData := make(chan int)

		_, err := encodeVersionedBytes(invalidData, JSONEncodingVersion2)

		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := fmt.Errorf("%w: unsupported encoding version %d for data map[key:value]", types.ErrInvalidEncoding, unsupportedVer)
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := encodeVersionedBytes(data, unsupportedVer)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := fmt.Errorf("unsupported encoding version %d for versionedData [97 98 99 100 102]", unsupportedVer)
		versionedBytes := &pb.VersionedBytes{
			Version: unsupportedVer, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := decodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestChainReaderClient(t *testing.T) {
	RunChainReaderInterfaceTests(t, &chainReaderInterfaceTester{chainReader: &fakeChainReader{lock: &sync.Mutex{}}})

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	es := &chainReaderErrServer{}
	pb.RegisterChainReaderServer(s, es)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()
	conn := connFromLis(t, lis)
	client := &chainReaderClient{grpc: pb.NewChainReaderClient(conn)}
	ctx := tests.Context(t)

	for _, errorType := range errorTypes {
		es.err = errorType

		t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			err := client.GetLatestValue(ctx, types.BoundContract{}, "method", "anything", "anything")
			assert.True(t, errors.Is(err, errorType))
		})
	}

	// make sure that errors come from client directly
	es.err = nil
	t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		err := client.GetLatestValue(ctx, types.BoundContract{}, "method", &cannotEncode{}, &TestStruct{})
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})
}

var encoder = makeEncoder()

func makeEncoder() cbor.EncMode {
	opts := cbor.CoreDetEncOptions()
	opts.Sort = cbor.SortCanonical
	e, _ := opts.EncMode()
	return e
}

type chainReaderInterfaceTester struct {
	interfaceTesterBase
	chainReader *fakeChainReader
}

func (it *chainReaderInterfaceTester) Setup(_ context.Context, t *testing.T) {
	it.setupHook = func(s *grpc.Server) {
		pb.RegisterChainReaderServer(s, &chainReaderServer{impl: it.chainReader})
	}
	it.interfaceTesterBase.Setup(t)
}

func (it *chainReaderInterfaceTester) SetLatestValue(_ context.Context, _ *testing.T, testStruct *TestStruct) types.BoundContract {
	it.chainReader.SetLatestValue(testStruct)
	return types.BoundContract{}
}

func (it *chainReaderInterfaceTester) GetPrimitiveContract(_ context.Context, _ *testing.T) types.BoundContract {
	return types.BoundContract{}
}

func (it *chainReaderInterfaceTester) GetSliceContract(_ context.Context, _ *testing.T) types.BoundContract {
	return types.BoundContract{}
}

func (it *chainReaderInterfaceTester) GetReturnSeenContract(ctx context.Context, t *testing.T) types.BoundContract {
	return types.BoundContract{}
}

func (it *chainReaderInterfaceTester) GetChainReader(ctx context.Context, t *testing.T) types.ChainReader {
	if it.conn == nil {
		it.conn = connFromLis(t, it.lis)
	}

	return &chainReaderClient{grpc: pb.NewChainReaderClient(it.conn)}
}

func (it *chainReaderInterfaceTester) TriggerEvent(ctx context.Context, t *testing.T, testStruct *TestStruct) types.BoundContract {
	it.chainReader.SetTrigger(testStruct)
	return types.BoundContract{}
}

type fakeChainReader struct {
	fakeTypeProvider
	stored      []TestStruct
	lock        *sync.Mutex
	lastTrigger TestStruct
}

func (f *fakeChainReader) SetLatestValue(ts *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.stored = append(f.stored, *ts)
}

func (f *fakeChainReader) GetLatestValue(_ context.Context, _ types.BoundContract, method string, params, returnVal any) error {
	if method == MethodReturningUint64 {
		r := returnVal.(*uint64)
		*r = AnyValueToReadWithoutAnArgument
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
		*returnVal.(*TestStruct) = f.lastTrigger
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

func (f *fakeChainReader) SetTrigger(testStruct *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.lastTrigger = *testStruct
}

type chainReaderErrServer struct {
	err error
	pb.UnimplementedChainReaderServer
}

func (e *chainReaderErrServer) GetLatestValue(context.Context, *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	return nil, e.err
}
