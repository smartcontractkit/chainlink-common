package mercury_v1

import (
	"sort"
)

func GetConsensusMaxFinalizedTimestamp(paos []IParsedAttributedObservation) uint32 {
	sort.Slice(paos, func(i, j int) bool {
		return paos[i].GetMaxFinalizedTimestamp() < paos[j].GetMaxFinalizedTimestamp()
	})
	return paos[len(paos)/2].GetMaxFinalizedTimestamp()
}
