package batch

import (
	"strconv"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"google.golang.org/protobuf/proto"
)

// splitMessagesByRequestSizeQuadratic is the previous implementation, kept here
// only as a benchmark baseline. It re-serializes the whole growing batch on
// every message (O(n^2) bytes walked), whereas splitMessagesByRequestSize
// accumulates per-event sizes (O(n)).
func splitMessagesByRequestSizeQuadratic(messages []*messageWithCallback, maxRequestSize int) [][]*messageWithCallback {
	if len(messages) == 0 {
		return nil
	}
	if maxRequestSize <= 0 {
		return [][]*messageWithCallback{messages}
	}

	var batches [][]*messageWithCallback
	current := make([]*messageWithCallback, 0, len(messages))
	for _, msg := range messages {
		candidate := append(current, msg)
		_, candidateBytes := newBatchRequest(candidate)
		if len(current) > 0 && candidateBytes > maxRequestSize {
			batches = append(batches, current)
			current = []*messageWithCallback{msg}
			continue
		}
		current = candidate
	}
	if len(current) > 0 {
		batches = append(batches, current)
	}
	return batches
}

func benchSplitMessages(n int) []*messageWithCallback {
	msgs := make([]*messageWithCallback, n)
	for i := range msgs {
		msgs[i] = &messageWithCallback{event: largeTestEvent(strconv.Itoa(i))}
	}
	return msgs
}

// benchSplitMaxRequestSize returns a limit that holds ~perBatch events, so the
// input is split into many sub-batches (the case where the quadratic cost bites).
func benchSplitMaxRequestSize(msgs []*messageWithCallback, perBatch int) int {
	if perBatch > len(msgs) {
		perBatch = len(msgs)
	}
	events := make([]*chipingress.CloudEventPb, perBatch)
	for i := 0; i < perBatch; i++ {
		events[i] = msgs[i].event
	}
	return proto.Size(&chipingress.CloudEventBatch{Events: events})
}

// BenchmarkSplitMessagesByRequestSize compares the linear (current) and
// quadratic (old) splitters across batch sizes.
//
//	go test ./pkg/chipingress/batch -bench BenchmarkSplitMessagesByRequestSize -benchmem -run '^$'
func BenchmarkSplitMessagesByRequestSize(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		msgs := benchSplitMessages(n)
		maxRequestSize := benchSplitMaxRequestSize(msgs, 10)

		b.Run("linear/n="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = splitMessagesByRequestSize(msgs, maxRequestSize)
			}
		})

		b.Run("quadratic/n="+strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = splitMessagesByRequestSizeQuadratic(msgs, maxRequestSize)
			}
		})
	}
}
