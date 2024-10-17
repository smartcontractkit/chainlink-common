package wasm

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	"google.golang.org/protobuf/proto"
)

type Runtime struct {
	fetchFn func(req sdk.FetchRequest) (sdk.FetchResponse, error)
	emitFn  func(msg string, labels map[string]any) error
	logger  logger.Logger
}

type RuntimeConfig struct {
	MaxFetchResponseSizeBytes int64
	RequestID                 *string
	MetaData                  *capabilities.RequestMetadata
}

func WithRequestMetaData(md *capabilities.RequestMetadata) func(*RuntimeConfig) {
	return func(rc *RuntimeConfig) {
		rc.MetaData = md
	}
}

func WithRequestID(id string) func(*RuntimeConfig) {
	return func(rc *RuntimeConfig) {
		rc.RequestID = &id
	}
}

const (
	defaultMaxFetchResponseSizeBytes = 5 * 1024
)

func defaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		MaxFetchResponseSizeBytes: defaultMaxFetchResponseSizeBytes,
	}
}

var _ sdk.Runtime = (*Runtime)(nil)

func (r *Runtime) Fetch(req sdk.FetchRequest) (sdk.FetchResponse, error) {
	return r.fetchFn(req)
}

func (r *Runtime) Logger() logger.Logger {
	return r.logger
}

func (r *Runtime) Emit(msg string, labels map[string]any) error {
	return r.emitFn(msg, labels)
}

// createEmitFn injects dependencies to implement an adapter for the wasm imported emit method
// that handles marshalling and unmarshalling messages for the WASM to act on.
func createEmitFn(
	sdkConfig *RuntimeConfig,
	l logger.Logger,
	emit func(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32,
) func(string, map[string]any) error {
	emitFn := func(msg string, labels map[string]any) error {
		vm, err := values.NewMap(labels)
		if err != nil {
			return err
		}

		b, err := proto.Marshal(&wasmpb.EmitMessageRequest{
			Message: msg,
			Labels:  values.ProtoMap(vm),
		})
		if err != nil {
			return err
		}

		respBuffer := make([]byte, sdkConfig.MaxFetchResponseSizeBytes)
		respptr, _ := bufferToPointerLen(respBuffer)

		resplenBuffer := make([]byte, uint32Size)
		resplenptr, _ := bufferToPointerLen(resplenBuffer)

		reqptr, reqptrlen := bufferToPointerLen(b)
		errno := emit(respptr, resplenptr, reqptr, reqptrlen)
		if errno != 0 {
			return fmt.Errorf("failed to emit with: %d", errno)
		}

		responseSize := binary.LittleEndian.Uint32(resplenBuffer)
		response := &wasmpb.EmitMessageResponse{}

		if err := proto.Unmarshal(respBuffer[:responseSize], response); err != nil {
			l.Errorw("failed to unmarshal emit response", "error", err.Error())
			return fmt.Errorf("failed to unmarshal emit response: %w", err)
		}

		if response.Error != nil && response.Error.Message != "" {
			return fmt.Errorf("failed to emit with: %s", response.Error.Message)
		}

		return nil
	}

	return emitFn
}

const uint32Size = int32(4)

// bufferToPointerLen returns a pointer to the first element of the buffer and the length of the buffer.
func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}
