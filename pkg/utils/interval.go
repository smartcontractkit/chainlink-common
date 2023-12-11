package utils

import (
	"database/sql/driver"
	"time"

	"github.com/pkg/errors"
)

// Interval represents a time.Duration stored as a Postgres interval type
type Interval time.Duration

// NewInterval creates Interval for specified duration
func NewInterval(d time.Duration) *Interval {
	i := new(Interval)
	*i = Interval(d)
	return i
}

func (i Interval) Duration() time.Duration {
	return time.Duration(i)
}

// MarshalText implements the text.Marshaler interface.
func (i Interval) MarshalText() ([]byte, error) {
	return []byte(time.Duration(i).String()), nil
}

// UnmarshalText implements the text.Unmarshaler interface.
func (i *Interval) UnmarshalText(input []byte) error {
	v, err := time.ParseDuration(string(input))
	if err != nil {
		return err
	}
	*i = Interval(v)
	return nil
}

func (i *Interval) Scan(v interface{}) error {
	if v == nil {
		*i = Interval(time.Duration(0))
		return nil
	}
	asInt64, is := v.(int64)
	if !is {
		return errors.Errorf("models.Interval#Scan() wanted int64, got %T", v)
	}
	*i = Interval(time.Duration(asInt64) * time.Nanosecond)
	return nil
}

func (i Interval) Value() (driver.Value, error) {
	return time.Duration(i).Nanoseconds(), nil
}

func (i Interval) IsZero() bool {
	return time.Duration(i) == time.Duration(0)
}
