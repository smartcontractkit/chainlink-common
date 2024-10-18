package wasm

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unsafe"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

const (
	labelWorkflowOwner       = "workflow_owner_address"
	labelWorkflowID          = "workflow_id"
	labelWorkflowExecutionID = "workflow_execution_id"
	labelWorkflowName        = "workflow_name"
)

// Length of responses are encoded into 4 byte buffers in little endian.  uint32Size is the size
// of that buffer.
const uint32Size = int32(4)

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

// createEmitFn builds the runtime's emit function implementation, which is a function
// that handles marshalling and unmarshalling messages for the WASM to act on.
func createEmitFn(
	sdkConfig *RuntimeConfig,
	l logger.Logger,
	emit func(respptr unsafe.Pointer, resplenptr unsafe.Pointer, reqptr unsafe.Pointer, reqptrlen int32) int32,
) func(string, map[string]any) error {
	emitFn := func(msg string, labels map[string]any) error {
		// Prepare the labels to be emitted
		if sdkConfig.MetaData == nil {
			return NewEmissionError(fmt.Errorf("metadata is required to emit"))
		}

		labels, err := toEmitLabels(sdkConfig.MetaData, labels)
		if err != nil {
			return NewEmissionError(err)
		}

		vm, err := values.NewMap(labels)
		if err != nil {
			return NewEmissionError(fmt.Errorf("could not wrap labels to map: %w", err))
		}

		// Marshal the message and labels into a protobuf message
		b, err := proto.Marshal(&wasmpb.EmitMessageRequest{
			Message: msg,
			Labels:  values.ProtoMap(vm),
		})
		if err != nil {
			return err
		}

		// Prepare the request to be sent to the host memory by allocating space for the
		// response and response length buffers.
		respBuffer := make([]byte, sdkConfig.MaxFetchResponseSizeBytes)
		respptr, _ := bufferToPointerLen(respBuffer)

		resplenBuffer := make([]byte, uint32Size)
		resplenptr, _ := bufferToPointerLen(resplenBuffer)

		// The request buffer is the wasm memory, get a pointer to the first element and the length
		// of the protobuf message.
		reqptr, reqptrlen := bufferToPointerLen(b)

		// Emit the message via the method imported from the host
		errno := emit(respptr, resplenptr, reqptr, reqptrlen)
		if errno != 0 {
			return NewEmissionError(fmt.Errorf("emit failed with errno %d", errno))
		}

		// Attempt to read and handle the response from the host memory
		responseSize := binary.LittleEndian.Uint32(resplenBuffer)
		response := &wasmpb.EmitMessageResponse{}
		if err := proto.Unmarshal(respBuffer[:responseSize], response); err != nil {
			l.Errorw("failed to unmarshal emit response", "error", err.Error())
			return NewEmissionError(err)
		}

		if response.Error != nil && response.Error.Message != "" {
			return NewEmissionError(errors.New(response.Error.Message))
		}

		return nil
	}

	return emitFn
}

// bufferToPointerLen returns a pointer to the first element of the buffer and the length of the buffer.
func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32) {
	return unsafe.Pointer(&buf[0]), int32(len(buf))
}

// toEmitLabels ensures that the required metadata is present in the labels map
func toEmitLabels(md *capabilities.RequestMetadata, labels map[string]any) (map[string]any, error) {
	if md.WorkflowID == "" {
		return nil, fmt.Errorf("must provide workflow id to emit event")
	}

	if md.WorkflowName == "" {
		return nil, fmt.Errorf("must provide workflow name to emit event")
	}

	if md.WorkflowOwner == "" {
		return nil, fmt.Errorf("must provide workflow owner to emit event")
	}

	labels[labelWorkflowExecutionID] = md.WorkflowExecutionID
	labels[labelWorkflowOwner] = md.WorkflowOwner
	labels[labelWorkflowID] = md.WorkflowID
	labels[labelWorkflowName] = md.WorkflowName
	return labels, nil
}

// EmissionError wraps all errors that occur during the emission process for the runtime to handle.
type EmissionError struct {
	Wrapped error
}

func NewEmissionError(err error) *EmissionError {
	return &EmissionError{Wrapped: err}
}

func (e *EmissionError) Error() string {
	return fmt.Errorf("failed to create emission: %w", e.Wrapped).Error()
}
