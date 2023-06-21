package types

import (
	"fmt"
	"math/big"
	"time"

	"golang.org/x/exp/constraints"
)

type Config interface {
	BlockEmissionIdleWarningThreshold() time.Duration
	FinalityDepth() uint32
	HeadTrackerHistoryDepth() uint32
	HeadTrackerMaxBufferSize() uint32
	HeadTrackerSamplingInterval() time.Duration
}

// FriendlyNumber returns a string printing the integer or big.Int in both
// decimal and hexadecimal formats.
func FriendlyNumber[N constraints.Integer | *big.Int](n N) string {
	return fmt.Sprintf("#%[1]v (0x%[1]x)", n)
}
