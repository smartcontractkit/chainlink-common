package beholdertest

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
)

func TestRecorder_Emit(t *testing.T) {
	t.Parallel()

	rec := NewRecorder()
	require.NoError(t, rec.Emit(t.Context(), []byte("m1"), beholder.AttrKeyEntity, "BaseMessage"))
	require.NoError(t, rec.EmitMessage(t.Context(), beholder.NewMessage([]byte("m2"))))

	msgs := rec.Messages(t)
	require.Len(t, msgs, 2)
	require.Equal(t, []byte("m1"), msgs[0].Body)
	require.Equal(t, []byte("m2"), msgs[1].Body)

	// Messages with no filter returns a copy of the recorder's slice.
	msgs[0] = beholder.Message{}
	require.Equal(t, []byte("m1"), rec.Messages(t)[0].Body)

	require.NoError(t, rec.Close())
}

func TestRecorder_MessagesFiltersWithOR(t *testing.T) {
	t.Parallel()

	rec := NewRecorder()
	require.NoError(t, rec.Emit(t.Context(), []byte("a"), "k1", "v1", "k2", "v2")) // matches both pairs
	require.NoError(t, rec.Emit(t.Context(), []byte("b"), "k1", "v1"))
	require.NoError(t, rec.Emit(t.Context(), []byte("c"), "k2", "v2"))
	require.NoError(t, rec.Emit(t.Context(), []byte("d"), "k3", "v3"))

	// A message matching several pairs is returned once.
	msgs := rec.Messages(t, "k1", "v1", "k2", "v2")
	require.Len(t, msgs, 3)
	bodies := make([]string, len(msgs))
	for i, m := range msgs {
		bodies[i] = string(m.Body)
	}
	require.ElementsMatch(t, []string{"a", "b", "c"}, bodies)

	require.Equal(t, 3, rec.Len(t, "k1", "v1", "k2", "v2"))
	require.Equal(t, 0, rec.Len(t, "k4", "v4"))
}

func TestRecorder_BaseMessagesForLabels(t *testing.T) {
	t.Parallel()

	rec := NewRecorder()
	payload := &pb.BaseMessage{Msg: "hello", Labels: map[string]string{"app": "test"}}
	body, err := proto.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, rec.Emit(t.Context(), body, beholder.AttrKeyEntity, "BaseMessage"))
	require.NoError(t, rec.Emit(t.Context(), []byte("other"), beholder.AttrKeyEntity, "SomethingElse"))

	found, err := rec.BaseMessagesForLabels(t, map[string]string{"app": "test"})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.Equal(t, "hello", found[0].Msg)
	require.Equal(t, map[string]string{"app": "test"}, found[0].Labels)

	found, err = rec.BaseMessagesForLabels(t, map[string]string{"app": "no-match"})
	require.NoError(t, err)
	require.Empty(t, found)
}

func TestRecorder_ConcurrentEmit(t *testing.T) {
	t.Parallel()

	rec := NewRecorder()

	const goroutines, perG = 8, 50
	var wg sync.WaitGroup
	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range perG {
				assert.NoError(t, rec.Emit(context.Background(), []byte("m")))
			}
		}()
	}
	wg.Wait()

	require.Len(t, rec.Messages(t), goroutines*perG)
}

func TestNewObserver_PanicsInParallelTests(t *testing.T) {
	t.Parallel()

	require.Panics(t, func() {
		NewObserver(t)
	}, "NewObserver must refuse to run in parallel tests")
}
