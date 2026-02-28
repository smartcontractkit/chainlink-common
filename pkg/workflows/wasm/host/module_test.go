package host

import (
	"context"
	"encoding/binary"
	"strings"
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

type fakeMemoryAccessor struct{ data []byte }

func (f *fakeMemoryAccessor) Memory() []byte { return f.data }

func newFakeMemoryAccessor() engine.MemoryAccessor {
	return &fakeMemoryAccessor{data: make([]byte, 65536)}
}

type mockMessageEmitter struct {
	e      func(context.Context, string, map[string]string) error
	labels map[string]string
}

func (m *mockMessageEmitter) Emit(ctx context.Context, msg string) error {
	return m.e(ctx, msg, m.labels)
}

func (m *mockMessageEmitter) WithMapLabels(labels map[string]string) custmsg.MessageEmitter {
	m.labels = labels
	return m
}

func (m *mockMessageEmitter) With(keyValues ...string) custmsg.MessageEmitter {
	// do nothing
	return m
}

func (m *mockMessageEmitter) Labels() map[string]string {
	return m.labels
}

func newMockMessageEmitter(e func(context.Context, string, map[string]string) error) custmsg.MessageEmitter {
	return &mockMessageEmitter{e: e}
}

// Test_createEmitFn tests that the emit function used by the module is created correctly.  Memory
// access functions are injected as mocks.
func Test_createEmitFn(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctxKey := "key"
		ctxValue := "test-value"
		ctx := t.Context()
		ctx = context.WithValue(ctx, ctxKey, "test-value")
		exec := &execution[*wasmpb.Response]{ctx: ctx}
		reqId := "random-id"
		emitFn := createEmitFn(
			logger.Test(t),
			exec,
			newMockMessageEmitter(func(ctx context.Context, _ string, _ map[string]string) error {
				v := ctx.Value(ctxKey)
				assert.Equal(t, ctxValue, v)
				return nil
			}),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{
					RequestId: reqId,
					Message:   "hello, world",
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
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
		)
		gotCode := emitFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("success without labels", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}
		emitFn := createEmitFn(
			logger.Test(t),
			exec,
			newMockMessageEmitter(func(_ context.Context, _ string, _ map[string]string) error {
				return nil
			}),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
		)
		gotCode := emitFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("successfully write error to memory on failure to read", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}
		respBytes, err := proto.Marshal(&wasmpb.EmitMessageResponse{
			Error: &wasmpb.Error{
				Message: assert.AnError.Error(),
			},
		})
		assert.NoError(t, err)

		emitFn := createEmitFn(
			logger.Test(t),
			exec,
			nil,
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				return nil, assert.AnError
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(newFakeMemoryAccessor(), 0, int32(len(respBytes)), 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode, "code mismatch")
	})

	t.Run("failure to emit writes error to memory", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}
		reqId := "random-id"
		respBytes, err := proto.Marshal(&wasmpb.EmitMessageResponse{
			Error: &wasmpb.Error{
				Message: assert.AnError.Error(),
			},
		})
		assert.NoError(t, err)

		emitFn := createEmitFn(
			logger.Test(t),
			exec,
			newMockMessageEmitter(func(_ context.Context, _ string, _ map[string]string) error {
				return assert.AnError
			}),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.EmitMessageRequest{
					RequestId: reqId,
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("bad read failure to unmarshal protos", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}
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
			exec,
			nil,
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				return badData, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				assert.Equal(t, respBytes, src, "marshalled response not equal to bytes to write")
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				assert.Equal(t, uint32(len(respBytes)), val, "did not write length of response")
				return 0
			}),
		)
		gotCode := emitFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})
}

func TestCreateFetchFn(t *testing.T) {
	const testID = "test-id"
	t.Run("OK-success", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.FetchRequest{
					Id: testID,
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return &FetchResponse{}, nil
				},
				MaxFetchRequests: 5,
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("NOK-fetch_fails_to_read_from_store", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				return nil, assert.AnError
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				// the error is handled and written to the buffer
				resp := &wasmpb.FetchResponse{}
				err := proto.Unmarshal(src, resp)
				require.NoError(t, err)
				assert.Equal(t, assert.AnError.Error(), resp.ErrorMessage)
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return &FetchResponse{}, nil
				},
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("NOK-fetch_fails_to_unmarshal_request", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				return []byte("bad-request-payload"), nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				// the error is handled and written to the buffer
				resp := &wasmpb.FetchResponse{}
				err := proto.Unmarshal(src, resp)
				require.NoError(t, err)
				expectedErr := "cannot parse invalid wire-format data"
				assert.Contains(t, resp.ErrorMessage, expectedErr)
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return &FetchResponse{}, nil
				},
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("NOK-fetch_returns_an_error", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.FetchRequest{
					Id: testID,
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				// the error is handled and written to the buffer
				resp := &wasmpb.FetchResponse{}
				err := proto.Unmarshal(src, resp)
				require.NoError(t, err)
				expectedErr := assert.AnError.Error()
				assert.Equal(t, expectedErr, resp.ErrorMessage)
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return nil, assert.AnError
				},
				MaxFetchRequests: 1,
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoSuccess, gotCode)
	})

	t.Run("NOK-fetch_fails_to_write_response", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.FetchRequest{
					Id: testID,
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				return -1
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return 0
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return &FetchResponse{}, nil
				},
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoFault, gotCode)
	})

	t.Run("NOK-fetch_fails_to_write_response_size", func(t *testing.T) {
		exec := &execution[*wasmpb.Response]{ctx: t.Context()}

		fetchFn := createFetchFn(
			logger.Test(t),
			unsafeReaderFunc(func(_ engine.MemoryAccessor, _, _ int32) ([]byte, error) {
				b, err := proto.Marshal(&wasmpb.FetchRequest{
					Id: testID,
				})
				assert.NoError(t, err)
				return b, nil
			}),
			unsafeWriterFunc(func(c engine.MemoryAccessor, src []byte, ptr, len int32) int64 {
				return 0
			}),
			unsafeFixedLengthWriterFunc(func(c engine.MemoryAccessor, ptr int32, val uint32) int64 {
				return -1
			}),
			&ModuleConfig{
				Logger: logger.Test(t),
				Fetch: func(ctx context.Context, req *FetchRequest) (*FetchResponse, error) {
					return &FetchResponse{}, nil
				},
			},
			exec,
		)

		gotCode := fetchFn(newFakeMemoryAccessor(), 0, 0, 0, 0)
		assert.Equal(t, ErrnoFault, gotCode)
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
		assert.Equal(t, int64(-1), n)
	})

	t.Run("fails to write to invalid access", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, len(giveSrc))
		n := write(memory, giveSrc, 0, -1)
		assert.Equal(t, int64(-1), n)

		n = write(memory, giveSrc, -1, 1)
		assert.Equal(t, int64(-1), n)
	})

	t.Run("truncated write due to size being smaller than len", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 12)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)-2))
		assert.Equal(t, int64(-1), n)
	})

	t.Run("unwanted data when size exceeds written data only writes the data", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 20)
		n := write(memory, giveSrc, 0, 20)
		// TODO verify this won't break anything...
		assert.Equal(t, int64(12), n)
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
		reqID := "random-id"
		msg := &wasmpb.EmitMessageRequest{
			RequestId: reqID,
			Message:   "hello, world",
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

		rid, gotMsg, gotLabels, err := toEmissible(b)
		assert.NoError(t, err)
		assert.Equal(t, "hello, world", gotMsg)
		assert.Equal(t, map[string]string{"test": "value"}, gotLabels)
		assert.Equal(t, reqID, rid)
	})

	t.Run("fails with bad message", func(t *testing.T) {
		_, _, _, err := toEmissible([]byte("not proto bufs"))
		assert.Error(t, err)
	})
}

func Test_SdkLabeler(t *testing.T) {
	t.Run("defaults to no-op when nil", func(t *testing.T) {
		// ModuleConfig with nil SdkLabeler should not panic when creating a module
		binary := createTestBinary(successBinaryCmd, successBinaryLocation, true, t)
		mc := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
			Fetch:          func(context.Context, *FetchRequest) (*FetchResponse, error) { return &FetchResponse{}, nil },
		}
		_, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)
		require.NotNil(t, mc.SdkLabeler, "SdkLabeler should be set to no-op")
	})

	t.Run("is called with v2ImportName after discovery", func(t *testing.T) {
		binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)
		var capturedName string
		mc := defaultNoDAGModCfg(t)
		mc.SdkLabeler = func(name string) {
			capturedName = name
		}
		m, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)
		require.False(t, m.IsLegacyDAG(), "expected NoDAG module")
		require.NotEmpty(t, capturedName, "SdkLabeler should have been called with v2 import name")
		require.True(t, strings.HasPrefix(capturedName, "version_v2"), "captured name should have v2 prefix")
	})
}

// CallAwaitRace validates that every call can be awaited.
func Test_CallAwaitRace(t *testing.T) {
	ctx := t.Context()
	mockExecHelper := NewMockExecutionHelper(t)
	mockExecHelper.EXPECT().
		CallCapability(matches.AnyContext, mock.Anything).
		Return(&sdkpb.CapabilityResponse{}, nil)

	m := &module{}

	var wg sync.WaitGroup
	var wantAttempts = 100

	exec := &execution[*wasmpb.ExecutionResult]{
		module:              m,
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	wg.Add(wantAttempts)
	for on := range wantAttempts {
		go func() {
			defer wg.Done()
			// call
			err := exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{
				Id:         "test-cap-request",
				CallbackId: int32(on),
			})
			require.NoError(t, err)

			// await with id
			_, err = exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{
				Ids: []int32{int32(on)},
			})
			require.NoError(t, err)
		}()
	}

	wg.Wait()
}
