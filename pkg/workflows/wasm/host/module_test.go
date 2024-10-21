package host

import (
	"encoding/binary"
	"testing"

	"github.com/bytecodealliance/wasmtime-go/v23"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type mockEmitLabeler struct {
	e      func(string, map[string]string) error
	labels map[string]string
}

func (m *mockEmitLabeler) Emit(msg string) error {
	return m.e(msg, m.labels)
}

func (m *mockEmitLabeler) WithMapLabels(labels map[string]string) EmitLabeler {
	m.labels = labels
	return m
}

func newMockEmitLabeler(e func(string, map[string]string) error) EmitLabeler {
	return &mockEmitLabeler{e: e}
}

// Test_createEmitFn tests that the emit function used by the module is created correctly.  Memory
// access functions are injected as mocks.
func Test_createEmitFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		emitFn := createEmitFn(
			logger.Test(t),
			newMockEmitLabeler(func(_ string, _ map[string]string) error {
				return nil
			}),
			unsafeReaderFunc(func(_ *wasmtime.Caller, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{
					Message: "hello, world",
					Labels: &pb.Map{
						Fields: map[string]*pb.Value{
							"foo": {
								Value: &pb.Value_StringValue{
									StringValue: "bar",
								},
							},
						},
					},
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c *wasmtime.Caller, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c *wasmtime.Caller, ptr int32, val uint32) int64 {
				return 0
			}),
		)
		gotCode := emitFn(new(wasmtime.Caller), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("success without labels", func(t *testing.T) {
		emitFn := createEmitFn(
			logger.Test(t),
			newMockEmitLabeler(func(_ string, _ map[string]string) error {
				return nil
			}),
			unsafeReaderFunc(func(_ *wasmtime.Caller, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c *wasmtime.Caller, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c *wasmtime.Caller, ptr int32, val uint32) int64 {
				return 0
			}),
		)
		gotCode := emitFn(new(wasmtime.Caller), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("successfully write error to memory on failure to read", func(t *testing.T) {
		respBytes, err := proto.Marshal(&wasmpb.EmitMessageResponse{
			Error: &wasmpb.Error{
				Message: assert.AnError.Error(),
			},
		})
		assert.NoError(t, err)

		emitFn := createEmitFn(
			logger.Test(t),
			nil,
			unsafeReaderFunc(func(_ *wasmtime.Caller, _, _ int32) ([]byte, error) {
				return nil, assert.AnError
			}),
			unsafeWriterFunc(func(c *wasmtime.Caller, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c *wasmtime.Caller, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(new(wasmtime.Caller), 0, int32(len(respBytes)), 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode, "code mismatch")
	})

	t.Run("failure to emit writes error to memory", func(t *testing.T) {
		respBytes, err := proto.Marshal(&wasmpb.EmitMessageResponse{
			Error: &wasmpb.Error{
				Message: assert.AnError.Error(),
			},
		})
		assert.NoError(t, err)

		emitFn := createEmitFn(
			logger.Test(t),
			newMockEmitLabeler(func(_ string, _ map[string]string) error {
				return assert.AnError
			}),
			unsafeReaderFunc(func(_ *wasmtime.Caller, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c *wasmtime.Caller, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c *wasmtime.Caller, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(new(wasmtime.Caller), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("bad read failure to unmarshal protos", func(t *testing.T) {
		badData := []byte("not proto bufs")
		msg := &wasmpb.EmitMessageRequest{}
		marshallErr := proto.Unmarshal(badData, msg)
		assert.Error(t, marshallErr)

		respBytes, err := proto.Marshal(&wasmpb.EmitMessageResponse{
			Error: &wasmpb.Error{
				Message: marshallErr.Error(),
			},
		})
		assert.NoError(t, err)

		emitFn := createEmitFn(
			logger.Test(t),
			nil,
			unsafeReaderFunc(func(_ *wasmtime.Caller, _, _ int32) ([]byte, error) {
				return badData, nil
			}),
			unsafeWriterFunc(func(c *wasmtime.Caller, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c *wasmtime.Caller, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(new(wasmtime.Caller), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})
}

func Test_read(t *testing.T) {
	t.Run("successfully read from slice", func(t *testing.T) {
		memory := []byte("hello, world")
		got, err := read(memory, 0, int32(len(memory)))
		assert.NoError(t, err)
		assert.Equal(t, []byte("hello, world"), got)
	})

	t.Run("fail to read because out of bounds request", func(t *testing.T) {
		memory := []byte("hello, world")
		_, err := read(memory, 0, int32(len(memory)+1))
		assert.Error(t, err)
	})

	t.Run("fails to read because of invalid pointer or length", func(t *testing.T) {
		memory := []byte("hello, world")
		_, err := read(memory, 0, -1)
		assert.Error(t, err)

		_, err = read(memory, -1, 1)
		assert.Error(t, err)
	})

	t.Run("validate that memory is read only once copied", func(t *testing.T) {
		memory := []byte("hello, world")
		copied, err := read(memory, 0, int32(len(memory)))
		assert.NoError(t, err)

		// mutate copy
		copied[0] = 'H'
		assert.Equal(t, []byte("Hello, world"), copied)

		// original memory is unchanged
		assert.Equal(t, []byte("hello, world"), memory)
	})
}

func Test_write(t *testing.T) {
	t.Run("successfully write to slice", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 12)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)))
		assert.Equal(t, n, int64(len(giveSrc)))
		assert.Equal(t, []byte("hello, world"), memory[:len(giveSrc)])
	})

	t.Run("cannot write to slice because memory too small", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, len(giveSrc)-1)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)))
		assert.Equal(t, n, int64(-1))
	})

	t.Run("fails to write to invalid access", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, len(giveSrc))
		n := write(memory, giveSrc, 0, -1)
		assert.Equal(t, n, int64(-1))

		n = write(memory, giveSrc, -1, 1)
		assert.Equal(t, n, int64(-1))
	})
}

// Test_writeUInt32 tests that a uint32 is written to memory correctly.
func Test_writeUInt32(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		memory := make([]byte, 4)
		n := writeUInt32(memory, 0, 42)
		wantBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(wantBuf, 42)
		assert.Equal(t, n, int64(4))
		assert.Equal(t, wantBuf, memory)
	})
}

func Test_toValidatedLabels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		msg := &wasmpb.EmitMessageRequest{
			Labels: &pb.Map{
				Fields: map[string]*pb.Value{
					"test": {
						Value: &pb.Value_StringValue{
							StringValue: "value",
						},
					},
				},
			},
		}
		wantLabels := map[string]string{
			"test": "value",
		}
		gotLabels, err := toValidatedLabels(msg)
		assert.NoError(t, err)
		assert.Equal(t, wantLabels, gotLabels)
	})

	t.Run("success with empty labels", func(t *testing.T) {
		msg := &wasmpb.EmitMessageRequest{}
		wantLabels := map[string]string{}
		gotLabels, err := toValidatedLabels(msg)
		assert.NoError(t, err)
		assert.Equal(t, wantLabels, gotLabels)
	})

	t.Run("fails with non string", func(t *testing.T) {
		msg := &wasmpb.EmitMessageRequest{
			Labels: &pb.Map{
				Fields: map[string]*pb.Value{
					"test": {
						Value: &pb.Value_Int64Value{
							Int64Value: *proto.Int64(42),
						},
					},
				},
			},
		}
		_, err := toValidatedLabels(msg)
		assert.Error(t, err)
	})
}

func Test_toEmissible(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		msg := &wasmpb.EmitMessageRequest{
			Message: "hello, world",
			Labels: &pb.Map{
				Fields: map[string]*pb.Value{
					"test": {
						Value: &pb.Value_StringValue{
							StringValue: "value",
						},
					},
				},
			},
		}

		b, err := proto.Marshal(msg)
		assert.NoError(t, err)

		gotMsg, gotLabels, err := toEmissible(b)
		assert.NoError(t, err)
		assert.Equal(t, "hello, world", gotMsg)
		assert.Equal(t, map[string]string{"test": "value"}, gotLabels)
	})

	t.Run("fails with bad message", func(t *testing.T) {
		_, _, err := toEmissible([]byte("not proto bufs"))
		assert.Error(t, err)
	})
}
