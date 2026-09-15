package pb

import "math"

// MaxSeqNum returns the max sequence number from TimestampsBySequence, or 0 if none exist.
func (t *ObservedDonTimes) MaxSeqNum() (maxSeqNum int64) {
	for seqNum := range t.TimestampsBySequence {
		if seqNum > maxSeqNum {
			maxSeqNum = seqNum
		}
	}
	return
}

// EarliestTS returns the easliest timestamp value from TimestampsBySequence, or math.MaxInt64 if none exist.
func (t *ObservedDonTimes) EarliestTS() (earliestTS int64) {
	earliestTS = math.MaxInt64
	for _, ts := range t.TimestampsBySequence {
		if ts < earliestTS {
			earliestTS = ts
		}
	}
	return
}
