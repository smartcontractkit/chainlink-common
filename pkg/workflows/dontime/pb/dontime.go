package pb

import "math"

// MaxSeqNum returns the max sequence number from TimestampsBySequence, or -1 if none exist.
func (t *ObservedDonTimes) MaxSeqNum() int64 {
	var maxSeqNum int64 = -1
	for seqNum := range t.TimestampsBySequence {
		if seqNum > maxSeqNum {
			maxSeqNum = seqNum
		}
	}
	return maxSeqNum
}

// EarliestTS returns the earliest timestamp value from TimestampsBySequence, or math.MaxInt64 if none exist.
func (t *ObservedDonTimes) EarliestTS() (earliestTS int64) {
	earliestTS = math.MaxInt64
	for _, ts := range t.TimestampsBySequence {
		if ts < earliestTS {
			earliestTS = ts
		}
	}
	return
}
