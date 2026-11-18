package pb

import "cmp"

func (e *ConsensusQueueEvent) CompareDeadline(o *ConsensusQueueEvent) int {
	return cmp.Or(cmp.Compare(e.Deadline.Seconds, o.Deadline.Seconds),
		cmp.Compare(e.Deadline.Nanos, o.Deadline.Nanos),
		cmp.Compare(e.TriggerEventID, o.TriggerEventID))
}
