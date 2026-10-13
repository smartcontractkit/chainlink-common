package beholdertest

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const (
	packageNameBeholder = "test_beholder"
)

// Observer is a test helper that provides assertion methods on received messages within the beholder client.
type Observer struct {
	*Recorder
}

// NewObserver sets the global beholder client as a message collector and returns an Observer that provides helper assertion
// functions on received messages.
//
// NewObserver affects the whole process, it cannot be used in parallel tests or tests with parallel ancestors.
// Prefer NewRecorder and injecting it; NewObserver mutates process-global state.
func NewObserver(t *testing.T) Observer {
	t.Helper()

	// Fail fast if this test runs in parallel: swapping the process-global client is unsafe there.
	t.Setenv(packageNameBeholder, packageNameBeholder)

	rec := NewRecorder()

	client := beholder.NoopClientConfig{Lggr: logger.Test(t)}.New()
	client.Emitter = rec
	require.NoError(t, client.Start(t.Context()))

	// reset NewObserver state after the test
	prevClient := beholder.GetClient()
	t.Cleanup(func() {
		defer beholder.SetClient(prevClient)
		require.NoError(t, client.Close())
	})
	beholder.SetClient(client)

	return Observer{Recorder: rec}
}

// Recorder is a beholder.Emitter that records every emitted message for test assertions.
// It is safe for concurrent use, so it may be shared across parallel tests.
type Recorder struct {
	mu   sync.RWMutex
	msgs []beholder.Message
}

// NewRecorder returns an empty Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

var _ beholder.Emitter = (*Recorder)(nil)

func (r *Recorder) Close() error { return nil }

func (r *Recorder) Emit(_ context.Context, msg []byte, attrKVs ...any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.msgs = append(r.msgs, beholder.NewMessage(slices.Clone(msg), attrKVs...))

	return nil
}

func (r *Recorder) EmitMessage(_ context.Context, msg beholder.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.msgs = append(r.msgs, msg.Copy())

	return nil
}

// Len returns the total number of messages received that match the provided attribute key/value pairs.
func (r *Recorder) Len(t *testing.T, attrKVs ...any) int {
	t.Helper()

	if len(attrKVs) == 0 {
		r.mu.RLock()
		defer r.mu.RUnlock()

		return len(r.msgs)
	}

	found := r.msgsForKVs(t, attrKVs...)

	return len(found)
}

// Messages returns messages matching the provided keys and values.
// A message matching multiple pairs is returned once.
func (r *Recorder) Messages(t *testing.T, attrKVs ...any) []beholder.Message {
	t.Helper()

	if len(attrKVs) == 0 {
		r.mu.RLock()
		defer r.mu.RUnlock()

		return slices.Clone(r.msgs)
	}

	return r.msgsForKVs(t, attrKVs...)
}

// msgsForKVs returns messages matching any of the provided attribute key/value pairs.
func (r *Recorder) msgsForKVs(t *testing.T, attrKVs ...any) []beholder.Message {
	t.Helper()
	require.True(t, len(attrKVs)%2 == 0, "attribute key-value pairs must have an even length")

	for i := 0; i < len(attrKVs); i += 2 {
		_, ok := attrKVs[i].(string)
		require.True(t, ok, "attribute key at index %d must be a string", i)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	found := []beholder.Message{}

	for _, eMsg := range r.msgs {
		for i := 0; i+1 < len(attrKVs); i += 2 {
			key := attrKVs[i].(string)
			value := attrKVs[i+1]
			if val, ok := eMsg.Attrs[key]; ok && reflect.DeepEqual(value, val) {
				found = append(found, eMsg)
				break
			}
		}
	}

	return found
}

// BaseMessagesForLabels returns the decoded BaseMessage payloads received with all the provided labels.
func (r *Recorder) BaseMessagesForLabels(t *testing.T, labels map[string]string) ([]*pb.BaseMessage, error) {
	t.Helper()

	r.mu.RLock()
	defer r.mu.RUnlock()

	var found []*pb.BaseMessage

messageLoop:
	for _, eMsg := range r.msgs {
		entity, ok := eMsg.Attrs[beholder.AttrKeyEntity].(string)
		if !ok || entity != "BaseMessage" {
			continue
		}

		payload := pb.BaseMessage{}
		err := proto.Unmarshal(eMsg.Body, &payload)
		if err != nil {
			return nil, fmt.Errorf("error unmarshalling base message: %w", err)
		}

		for k, v := range labels {
			if payload.Labels[k] != v {
				continue messageLoop
			}
		}

		found = append(found, &payload)
	}

	return found, nil
}
