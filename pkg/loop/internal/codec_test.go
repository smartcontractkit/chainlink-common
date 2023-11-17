package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/mitchellh/mapstructure"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
	. "github.com/smartcontractkit/chainlink-relay/pkg/types/interfacetests"
	"github.com/smartcontractkit/chainlink-relay/pkg/utils/tests"
)

func TestCodecClient(t *testing.T) {
	RunCodecInterfaceTests(t, &codecInterfaceTester{codec: &fakeCodec{}})

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	es := &codecErrServer{}
	pb.RegisterCodecServer(s, es)
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
	defer s.Stop()
	conn := connFromLis(t, lis)
	client := &codecClient{grpc: pb.NewCodecClient(conn)}
	ctx := tests.Context(t)

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
	var invalidTypeErr types.InvalidTypeError

	t.Run("Encode returns error if type cannot be encoded in the wire format", func(t *testing.T) {
		_, err := client.Encode(ctx, &cannotEncode{}, "doesnotmatter")

		assert.NotNil(t, errors.As(err, &invalidTypeErr))
	})

	t.Run("Decode returns error if type cannot be decoded in the wire format", func(t *testing.T) {
		err := client.Decode(ctx, []byte("does not matter"), &cannotEncode{}, "")
		assert.NotNil(t, errors.As(err, &invalidTypeErr))
	})
}

type codecInterfaceTester struct {
	interfaceTesterBase
	codec *fakeCodec
}

func (it *codecInterfaceTester) Setup(t *testing.T) {
	it.setupHook = func(s *grpc.Server) {
		pb.RegisterCodecServer(s, &codecServer{impl: it.codec})
	}
	it.interfaceTesterBase.Setup(t)
}

func (it *codecInterfaceTester) GetCodec(t *testing.T) types.Codec {
	if it.conn == nil {
		it.conn = connFromLis(t, it.lis)
	}

	return &codecClient{grpc: pb.NewCodecClient(it.conn)}
}

type fakeCodec struct {
	fakeTypeProvider
	lastItem any
}

func (f *fakeCodec) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return f.GetMaxEncodingSize(ctx, n, itemType)
}

func (f *fakeCodec) GetMaxEncodingSize(_ context.Context, _ int, itemType string) (int, error) {
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return 1, nil
	}
	return 0, types.InvalidTypeError{}
}

func (it *codecInterfaceTester) EncodeFields(t *testing.T, request *EncodeRequest) ocrtypes.Report {
	if request.TestOn == TestItemType {
		bytes, err := encoder.Marshal(request.TestStructs[0])
		require.NoError(t, err)
		return bytes
	}

	bytes, err := encoder.Marshal(request.TestStructs)
	require.NoError(t, err)
	return bytes
}

func (it *codecInterfaceTester) IncludeArrayEncodingSizeEnforcement() bool {
	return false
}

func (f *fakeCodec) Encode(_ context.Context, item any, itemType string) (ocrtypes.Report, error) {
	f.lastItem = item
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return encoder.Marshal(item)
	}
	return nil, types.InvalidTypeError{}
}

func (f *fakeCodec) Decode(_ context.Context, _ []byte, into any, itemType string) error {
	switch itemType {
	case TestItemType, TestItemSliceType, TestItemArray2Type, TestItemArray1Type:
		return mapstructure.Decode(f.lastItem, into)
	}
	return types.InvalidTypeError{}
}

type codecErrServer struct {
	err error
	pb.UnimplementedCodecServer
}

func (e *codecErrServer) GetEncoding(context.Context, *pb.GetEncodingRequest) (*pb.GetEncodingResponse, error) {
	return nil, e.err
}

func (e *codecErrServer) GetDecoding(context.Context, *pb.GetDecodingRequest) (*pb.GetDecodingResponse, error) {
	return nil, e.err
}

func (e *codecErrServer) GetMaxSize(context.Context, *pb.GetMaxSizeRequest) (*pb.GetMaxSizeResponse, error) {
	return nil, e.err
}
