package models

import (
	"fmt"
	"time"
)

// Duration is a non-negative time duration.
type Duration struct{ d time.Duration }

func MakeDuration(d time.Duration) (Duration, error) {
	if d < time.Duration(0) {
		return Duration{}, fmt.Errorf("cannot make negative time duration: %s", d)
	}
	return Duration{d: d}, nil
}

func ParseDuration(s string) (Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return Duration{}, err
	}

	return MakeDuration(d)
}

func MustMakeDuration(d time.Duration) Duration {
	rv, err := MakeDuration(d)
	if err != nil {
		panic(err)
	}
	return rv
}

func MustNewDuration(d time.Duration) *Duration {
	rv := MustMakeDuration(d)
	return &rv
}

// Duration returns the value as the standard time.Duration value.
func (d Duration) Duration() time.Duration {
	return d.d
}
