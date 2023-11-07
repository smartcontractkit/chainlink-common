package internal

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/test/bufconn"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"

	"github.com/fxamacker/cbor/v2"
	"github.com/mitchellh/mapstructure"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	. "github.com/smartcontractkit/chainlink-relay/pkg/types/interfacetests"
)

func TestVersionedBytesFunctionsBadPaths(t *testing.T) {
	t.Run("EncodeVersionedBytes unsupported type", func(t *testing.T) {
		expected := types.InvalidTypeError{}
		invalidData := make(chan int)

		_, err := encodeVersionedBytes(invalidData, SimpleJsonEncodingVersion)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := errors.New("unsupported encoding version 2 for data map[key:value]")
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := encodeVersionedBytes(data, 2)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := errors.New("unsupported encoding version 2 for versionedData [97 98 99 100 102]")
		versionedBytes := &pb.VersionedBytes{
			Version: 2, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := decodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}

func TestChainReaderClient(t *testing.T) {
	RunChainReaderInterfaceTests(t, &interfaceTester{})

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	es := &errorServer{}
	pb.RegisterChainReaderServer(s, es)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()
	conn := connFromLis(t, lis)
	client := &chainReaderClient{grpc: pb.NewChainReaderClient(conn)}
	ctx := context.Background()

	errorTypes := []error{
		types.InvalidEncodingError{},
		types.InvalidTypeError{},
		types.FieldNotFoundError{},
		types.WrongNumberOfElements{},
		types.NotASliceError{},
	}

	for _, errorType := range errorTypes {
		es.err = errorType
		t.Run("Encode unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			_, err := client.Encode(ctx, "any", "doesnotmatter")
			assert.IsType(t, errorType, err)
		})

		t.Run("Decode unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			_, err := client.Encode(ctx, "any", "doesnotmatter")
			assert.IsType(t, errorType, err)
		})

		t.Run("GetLatestValue unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			err := client.GetLatestValue(ctx, types.BoundContract{}, "method", "anything", "anything")
			assert.IsType(t, errorType, err)
		})

		t.Run("GetMaxEncodingSize unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			_, err := client.GetMaxEncodingSize(ctx, 1, "anything")
			assert.IsType(t, errorType, err)
		})

		t.Run("GetMaxEncodingSize unwraps errors from server "+errorType.Error(), func(t *testing.T) {
			_, err := client.GetMaxDecodingSize(ctx, 1, "anything")
			assert.IsType(t, errorType, err)
		})
	}

	// make sure that errors come from client directly
	es.err = nil

	t.Run("Encode returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		_, err := client.Encode(ctx, &cannotEncode{}, "doesnotmatter")
		assert.IsType(t, types.InvalidTypeError{}, err)
	})

	t.Run("Decode returns error if type cannot be decoded in the wire format", func(t *testing.T) {
		_, err := client.Encode(ctx, &cannotEncode{}, "doesnotmatter")
		assert.IsType(t, types.InvalidTypeError{}, err)
	})

	t.Run("GetLatestValue returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		err := client.GetLatestValue(ctx, types.BoundContract{}, "method", &cannotEncode{}, &TestStruct{})
		assert.IsType(t, types.InvalidTypeError{}, err)
	})
}

type interfaceTester struct {
	lis    *bufconn.Listener
	server *grpc.Server
	conn   *grpc.ClientConn
	fs     *fakeCodecServer
}

const methodName = "method"

var encoder = makeEncoder()

func makeEncoder() cbor.EncMode {
	opts := cbor.CoreDetEncOptions()
	opts.Sort = cbor.SortCanonical
	e, _ := opts.EncMode()
	return e
}

func (it *interfaceTester) SetLatestValue(_ *testing.T, testStruct *TestStruct) (types.BoundContract, string) {
	it.fs.SetLatestValue(testStruct)
	return types.BoundContract{}, methodName
}

func (it *interfaceTester) EncodeFields(t *testing.T, request *EncodeRequest) ocrtypes.Report {
	if request.TestOn == TestItemType {
		bytes, err := encoder.Marshal(request.TestStructs[0])
		require.NoError(t, err)
		return bytes
	}

	bytes, err := encoder.Marshal(request.TestStructs)
	require.NoError(t, err)
	return bytes
}

func (it *interfaceTester) GetAccountBytes(_ int) []byte {
	return []byte{1, 2, 3}
}

func (it *interfaceTester) IncludeArrayEncodingSizeEnforcement() bool {
	return false
}

func (it *interfaceTester) Setup(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	it.lis = lis
	it.fs = &fakeCodecServer{lock: &sync.Mutex{}}
	s := grpc.NewServer()
	pb.RegisterChainReaderServer(s, &chainReaderServer{impl: it.fs})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func (it *interfaceTester) Teardown(t *testing.T) {
	if it.server != nil {
		it.server.Stop()
	}

	if it.conn != nil {
		require.NoError(t, it.conn.Close())
	}

	it.lis = nil
	it.server = nil
	it.conn = nil
}

func (it *interfaceTester) Name() string {
	return "relay client"
}

func (it *interfaceTester) GetChainReader(t *testing.T) types.ChainReader {
	if it.conn == nil {
		it.conn = connFromLis(t, it.lis)
	}

	return &chainReaderClient{grpc: pb.NewChainReaderClient(it.conn)}
}

type fakeCodecServer struct {
	lastItem any
	latest   []TestStruct
	lock     *sync.Mutex
}

func (f *fakeCodecServer) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return f.GetMaxEncodingSize(ctx, n, itemType)
}

func (f *fakeCodecServer) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return 1, nil
	}
	return 0, types.InvalidTypeError{}
}

func (f *fakeCodecServer) SetLatestValue(ts *TestStruct) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.latest = append(f.latest, *ts)
}

func (f *fakeCodecServer) GetLatestValue(ctx context.Context, _ types.BoundContract, _ string, params, returnVal any) error {
	f.lock.Lock()
	defer f.lock.Unlock()
	lp := params.(*LatestParams)
	rv := returnVal.(*TestStruct)
	*rv = f.latest[lp.I-1]
	return nil
}

func (f *fakeCodecServer) Encode(_ context.Context, item any, itemType string) (ocrtypes.Report, error) {
	f.lastItem = item
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return encoder.Marshal(item)
	}
	return nil, types.InvalidTypeError{}
}

func (f *fakeCodecServer) Decode(_ context.Context, raw []byte, into any, itemType string) error {
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return mapstructure.Decode(f.lastItem, into)
	}
	return types.InvalidTypeError{}
}

func (f *fakeCodecServer) CreateType(itemType string, _, isEncode bool) (any, error) {
	switch itemType {
	case TestItemType:
		return &TestStruct{}, nil
	case TestItemSliceType:
		return &[]TestStruct{}, nil
	case TestItemArray2Type:
		return &[2]TestStruct{}, nil
	case TestItemArray1Type:
		return &[1]TestStruct{}, nil
	case methodName:
		if isEncode {
			return &LatestParams{}, nil
		}
		return &TestStruct{}, nil
	}

	return nil, types.InvalidTypeError{}
}

var _ types.RemoteCodec = &fakeCodecServer{}

type errorServer struct {
	err error
	pb.UnimplementedChainReaderServer
}

func (e *errorServer) GetLatestValue(context.Context, *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	return nil, e.err
}

func (e *errorServer) GetEncoding(context.Context, *pb.GetEncodingRequest) (*pb.GetEncodingResponse, error) {
	return nil, e.err
}

func (e *errorServer) GetDecoding(context.Context, *pb.GetDecodingRequest) (*pb.GetDecodingResponse, error) {
	return nil, e.err
}

func (e *errorServer) GetMaxSize(context.Context, *pb.GetMaxSizeRequest) (*pb.GetMaxSizeResponse, error) {
	return nil, e.err
}

func connFromLis(t *testing.T, lis *bufconn.Listener) *grpc.ClientConn {
	conn, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	require.NoError(t, err)
	return conn
}

type cannotEncode struct{}

func (*cannotEncode) MarshalBinary() ([]byte, error) {
	return nil, errors.New("nope")
}

func (*cannotEncode) UnmarshalBinary() error {
	return errors.New("nope")
}

func (*cannotEncode) MarshalText() ([]byte, error) {
	return nil, errors.New("nope")
}

func (*cannotEncode) UnmarshalText() error {
	return errors.New("nope")
}
