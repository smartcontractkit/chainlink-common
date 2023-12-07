package internal

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	. "github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

var errorTypes = []error{
	types.ErrInvalidEncoding,
	types.ErrInvalidType,
	types.ErrFieldNotFound,
	types.ErrWrongNumberOfElements,
	types.ErrNotASlice,
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

type interfaceTesterBase struct {
	lis       *bufconn.Listener
	server    *grpc.Server
	conn      *grpc.ClientConn
	setupHook func(server *grpc.Server)
}

var anyAccountBytes = []byte{1, 2, 3}

func (it *interfaceTesterBase) GetAccountBytes(_ int) []byte {
	return anyAccountBytes
}

func (it *interfaceTesterBase) Setup(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	it.lis = lis
	s := grpc.NewServer()
	it.setupHook(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			require.NoError(t, err)
		}
	}()

	t.Cleanup(func() {
		if it.server != nil {
			it.server.Stop()
		}

		if it.conn != nil {
			require.NoError(t, it.conn.Close())
		}

		it.lis = nil
		it.server = nil
		it.conn = nil
	})
}

func (it *interfaceTesterBase) Name() string {
	return "relay client"
}

type fakeTypeProvider struct{}

func (fakeTypeProvider) CreateType(itemType string, isEncode bool) (any, error) {
	switch itemType {
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
	case MethodReturningUint64:
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
	case EventName:
		if isEncode {
			return &struct{}{}, nil
		}
		return &TestStruct{}, nil
	}

	return nil, types.ErrInvalidType
}
