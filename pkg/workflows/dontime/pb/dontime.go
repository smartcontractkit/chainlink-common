package pb

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

// EarliestTS returns the earliest timestamp value from TimestampsBySequence or nil if none exist.
func (t *ObservedDonTimes) EarliestTS() (earliestTS *int64) {
	for _, ts := range t.TimestampsBySequence {
		if earliestTS == nil || ts < *earliestTS {
			earliestTS = &ts
		}
	}
	return
}
