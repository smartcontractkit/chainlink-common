package mercury_v1

import (
	"sort"
	"fmt"
)

func GetConsensusMaxFinalizedTimestamp(paos []ParsedAttributedObservation, f int) (uint32, error) {
	var validTimestampCount int
	timestampFrequencyMap := map[uint32]int{}
	for _, pao := range paos {
	ts, valid := pao.GetMaxFinalizedTimestamp()
		if valid {
			validTimestampCount++
			timestampFrequencyMap[ts]++
		}
	}

	// check if we have enough valid timestamps
	if validTimestampCount < f+1 {
		return 0, fmt.Errorf("fewer than f+1 observations have a valid maxFinalizedTimestamp (got: %d/%d)", validTimestampCount, len(paos))
	}

	var timestampFrequencyMaxCnt int
	for _, cnt := range timestampFrequencyMap {
		if cnt > timestampFrequencyMaxCnt {
			timestampFrequencyMaxCnt = cnt
		}
	}

	// check if we have enough valid timestamps with the max frequency
	if timestampFrequencyMaxCnt < f+1 {
		return 0, fmt.Errorf("no valid maxFinalizedTimestamp with at least f+1 votes (got counts: %v)", timestampFrequencyMap)
	}

	// select timestamps with the max frequency (in case there are more than one)
	// sort them deterministically
	var validTimestampsWithMaxFrequency []uint32
	for ts, cnt := range timestampFrequencyMap {
		if cnt == timestampFrequencyMaxCnt {
			validTimestampsWithMaxFrequency = append(validTimestampsWithMaxFrequency, ts)
		}
	}
	sort.Slice(validTimestampsWithMaxFrequency, func(i, j int) bool {
		return validTimestampsWithMaxFrequency[i] < validTimestampsWithMaxFrequency[j]
	})

	return validTimestampsWithMaxFrequency[0], nil
}
