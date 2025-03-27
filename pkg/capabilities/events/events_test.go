package events

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

type testEmitter struct {
	payload []byte
	attrs   []any
}

func (t *testEmitter) Emit(ctx context.Context, payload []byte, attrKVs ...any) error {
	t.payload = payload
	t.attrs = attrKVs
	return nil
}

func TestEmitter(t *testing.T) {
	client := &testEmitter{}
	emitter := &Emitter{client: client}
	msg := "a message"

	message := Message{
		Msg: msg,
	}
	err := emitter.Emit(t.Context(), message)
	assert.ErrorContains(t, err, "must provide workflow owner")

	message.Metadata.WorkflowOwner = "owner"
	err = emitter.Emit(t.Context(), message)
	assert.ErrorContains(t, err, "must provide workflow id")

	message.Metadata.WorkflowID = "id"
	err = emitter.Emit(t.Context(), message)
	assert.ErrorContains(t, err, "must provide workflow name")

	message.Metadata.WorkflowName = "name"
	err = emitter.Emit(t.Context(), message)
	require.NoError(t, err)

	event := &pb.BaseMessage{}
	err = proto.Unmarshal(client.payload, event)
	require.NoError(t, err)

	assert.Equal(t, event.Msg, msg)
}

func assertHasKey(t *testing.T, attrs []any, keyName, keyValue string) {
	for i, a := range attrs {
		if a.(string) == keyName {
			assert.Equal(t, attrs[i+1].(string), keyValue)
			return
		}
	}

	assert.FailNow(t, fmt.Sprintf("could not find keyName %s in attrs", keyName))
}

func TestEmitter_WithMetadata(t *testing.T) {
	client := &testEmitter{}
	emitter := &Emitter{client: client}
	emitter = emitter.With(EmitMetadata{
		WorkflowOwner: "owner",
		WorkflowID:    "id",
		WorkflowName:  "name",
	})
	msg := "a message"

	message := Message{
		Msg: msg,
	}
	err := emitter.Emit(t.Context(), message)
	require.NoError(t, err)

	fmt.Printf("%+v", client.attrs)
	assertHasKey(t, client.attrs, "workflow_owner_address", "owner")
	assertHasKey(t, client.attrs, "workflow_id", "id")
	assertHasKey(t, client.attrs, "workflow_name", "name")
}
