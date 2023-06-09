package types

import (
	"fmt"
	"time"
)

type Config interface {
	BlockEmissionIdleWarningThreshold() time.Duration
	FinalityDepth() uint32
	HeadTrackerHistoryDepth() uint32
	HeadTrackerMaxBufferSize() uint32
	HeadTrackerSamplingInterval() time.Duration
}

func FriendlyInt64(n int64) string {
	return fmt.Sprintf("#%[1]v (0x%[1]x)", n)
}
