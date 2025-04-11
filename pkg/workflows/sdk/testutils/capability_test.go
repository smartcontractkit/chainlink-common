package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type M struct {
	mock.Mock
}

func (m *M) Foo(ctx context.Context, msg proto.Message) (proto.Message, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(proto.Message), args.Error(1)
}

func TestThis(t *testing.T) {
	m := &M{}
	tmp := &pb.CapabilityRequest{}
	call := m.On("Foo", context.Background(), tmp)
	ccall := &MockCapabilityCall[proto.Message, proto.Message]{Call: call}
	ccall.Run(func(ctx context.Context, i proto.Message) (proto.Message, error) {
		return nil, errors.New("here we go")
	})
	f, err := m.Foo(context.Background(), tmp)
	assert.Nil(t, f)
	assert.Error(t, err)
}
