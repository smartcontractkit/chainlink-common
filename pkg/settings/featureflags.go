package settings

import (
	"context"
	"time"
)

// FeatureFlag wraps a Setting[time.Time] with a bound Getter, providing
// an Active method that compares an execution timestamp against the setting's
// activation time.
type FeatureFlag struct {
	setting Setting[time.Time]
	getter  Getter
}

func NewFeatureFlag(s Setting[time.Time], g Getter) FeatureFlag {
	return FeatureFlag{setting: s, getter: g}
}

// IsActive returns true when timestampMs (unix millis) is at or past the
// flag's activation time. Returns false when timestampMs is zero (no
// execution timestamp assigned).
func (f *FeatureFlag) IsActive(ctx context.Context, timestampMs int64) bool {
	if timestampMs <= 0 || f.setting.Parse == nil {
		return false
	}
	activateAt, _ := f.setting.GetOrDefault(ctx, f.getter)
	return timestampMs >= activateAt.UnixMilli()
}
