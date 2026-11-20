package pb

import (
	"bytes"
	"cmp"
	"crypto/sha256"

	"google.golang.org/protobuf/proto"
)

func (e *ConsensusQueueEvent) CompareDeadline(o *ConsensusQueueEvent) int {
	return cmp.Or(cmp.Compare(e.Deadline.Seconds, o.Deadline.Seconds),
		cmp.Compare(e.Deadline.Nanos, o.Deadline.Nanos),
		bytes.Compare(e.TimelessHash, o.TimelessHash))
}

func (e *CoordinatedEvent) TimelessHash() ([]byte, error) {
	clone := proto.CloneOf(e)
	clone.ObservedAt = nil
	clone.Deadline = nil
	pbBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(clone)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(pbBytes)
	return hash[:], nil
}
