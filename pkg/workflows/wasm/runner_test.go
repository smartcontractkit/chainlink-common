package wasm

import (
	"encoding/base64"
	"encoding/binary"
	"testing"
	"unsafe"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func Test_Runner_Config_InvalidRequest(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", "bla"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "could not decode request")
}

func Test_Runner_Config_InvalidRequest_NotEnoughArgs(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm"},
	}
	c := runner.Config()
	assert.Nil(t, c)
	assert.Equal(t, unknownID, gotResponse.Id)
	assert.Contains(t, gotResponse.ErrMsg, "request must contain a payload")
}

func marshalRequest(req *wasmpb.Request) (string, error) {
	rpb, err := proto.Marshal(req)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(rpb), nil
}

func Test_Runner_Config(t *testing.T) {
	var gotResponse *wasmpb.Response
	responseFn := func(resp *wasmpb.Response) {
		gotResponse = resp
	}

	cfg := []byte(`{"hello": "world"}`)
	request := &wasmpb.Request{
		Id:     uuid.New().String(),
		Config: cfg,
	}
	str, err := marshalRequest(request)
	require.NoError(t, err)
	runner := &Runner{
		sendResponse: responseFn,
		args:         []string{"wasm", str},
	}
	c := runner.Config()
	assert.Equal(t, cfg, c)
	assert.Nil(t, gotResponse)
}

// Test_createEmitFn validates the runtime's emit function implementation.  Uses mocks of the
// imported wasip1 emit function.
func Test_createEmitFn(t *testing.T) {
	var (
		l         = logger.Test(t)
		reqId     = "random-id"
		sdkConfig = &RuntimeConfig{
			MaxResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
			RequestID: &reqId,
		}
		giveMsg    = "testing guest"
		giveLabels = map[string]string{
			"some-key": "some-value",
		}
	)

	t.Run("success", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.NoError(t, err)
	})

	t.Run("success if no labels are given", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, nil)
		assert.NoError(t, err)
	})

	t.Run("successfully read error message when emit fails", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// marshall the protobufs
			b, err := proto.Marshal(&wasmpb.EmitMessageResponse{
				Error: &wasmpb.Error{
					Message: assert.AnError.Error(),
				},
			})
			assert.NoError(t, err)

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, assert.AnError.Error())
	})

	t.Run("fail to deserialize response from memory", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			// b is a non-protobuf byte slice
			b := []byte(assert.AnError.Error())

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(b))
			copy(resp, b)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(b)))

			return 0
		}

		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid wire-format data")
	})

	t.Run("fail with nonzero code from emit", func(t *testing.T) {
		hostEmit := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 42
		}
		runtimeEmit := createEmitFn(sdkConfig, l, hostEmit)
		err := runtimeEmit(giveMsg, giveLabels)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "emit failed with errno 42")
	})
}

func Test_createFetchFn(t *testing.T) {
	var (
		l         = logger.Test(t)
		requestID = uuid.New().String()
		sdkConfig = &RuntimeConfig{
			RequestID:            &requestID,
			MaxResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
		}
	)

	t.Run("OK-success", func(t *testing.T) {
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeFetch := createFetchFn(sdkConfig, l, hostFetch)
		response, err := runtimeFetch(sdk.FetchRequest{})
		assert.NoError(t, err)
		assert.Equal(t, sdk.FetchResponse{
			Headers: map[string]string{},
		}, response)
	})

	t.Run("NOK-config_missing_request_id", func(t *testing.T) {
		invalidConfig := &RuntimeConfig{
			RequestID:            nil,
			MaxResponseSizeBytes: 1_000,
			Metadata: &capabilities.RequestMetadata{
				WorkflowID:          "workflow_id",
				WorkflowExecutionID: "workflow_execution_id",
				WorkflowName:        "workflow_name",
				WorkflowOwner:       "workflow_owner_address",
			},
		}
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			return 0
		}
		runtimeFetch := createFetchFn(invalidConfig, l, hostFetch)
		_, err := runtimeFetch(sdk.FetchRequest{})
		assert.ErrorContains(t, err, "request ID is required to fetch")
	})

	t.Run("NOK-fetch_returns_handled_error", func(t *testing.T) {
		hostFetch := func(respptr, resplenptr, reqptr unsafe.Pointer, reqptrlen int32) int32 {
			fetchResponse := &wasmpb.FetchResponse{
				ExecutionError: true,
				ErrorMessage:   assert.AnError.Error(),
			}
			respBytes, perr := proto.Marshal(fetchResponse)
			if perr != nil {
				return 0
			}

			// write the marshalled response message to memory
			resp := unsafe.Slice((*byte)(respptr), len(respBytes))
			copy(resp, respBytes)

			// write the length of the response to memory in little endian
			respLen := unsafe.Slice((*byte)(resplenptr), uint32Size)
			binary.LittleEndian.PutUint32(respLen, uint32(len(respBytes)))

			return 0
		}
		runtimeFetch := createFetchFn(sdkConfig, l, hostFetch)
		_, err := runtimeFetch(sdk.FetchRequest{})
		assert.ErrorContains(t, err, assert.AnError.Error())
	})
}
