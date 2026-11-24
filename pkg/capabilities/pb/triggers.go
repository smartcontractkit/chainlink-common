package pb

import (
	"bytes"
	"cmp"

	"google.golang.org/protobuf/proto"
)

func (e *ConsensusQueueEvent) CompareDeadline(o *ConsensusQueueEvent) int {
	return cmp.Or(cmp.Compare(e.Deadline.Seconds, o.Deadline.Seconds),
		cmp.Compare(e.Deadline.Nanos, o.Deadline.Nanos),
		bytes.Compare(e.TimelessHash, o.TimelessHash))
}

func (e *CoordinatedEvent) MarshalTimeless() ([]byte, error) {
	clone := proto.CloneOf(e)
	clone.ObservedAt = nil
	clone.Deadline = nil
	return proto.MarshalOptions{Deterministic: true}.Marshal(clone)
}
