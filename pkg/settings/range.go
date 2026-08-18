package settings

import (
	"cmp"
	"errors"
	"fmt"
	"strings"
)

// Range represents a closed (inclusive) interval of type N.
type Range[N cmp.Ordered] struct {
	Lower, Upper N
}

func (r Range[N]) Contains(n N) bool {
	return r.Lower <= n && n <= r.Upper
}

func (r Range[N]) String() string {
	return fmt.Sprintf("[%v,%v]", r.Lower, r.Upper)
}

// ParseRangeFn return a func for parsing a Range of type N.
func ParseRangeFn[N cmp.Ordered](parseFn func(string) (N, error)) func(s string) (Range[N], error) {
	return func(s string) (Range[N], error) {
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		parts := strings.Split(s, ",")
		if len(parts) != 2 {
			return Range[N]{}, fmt.Errorf("invalid range: must have two comma separated values: %q", s)
		}
		lower, lerr := parseFn(parts[0])
		upper, uerr := parseFn(parts[1])
		err := errors.Join(lerr, uerr)
		if err != nil {
			return Range[N]{}, err
		}
		return Range[N]{Lower: lower, Upper: upper}, nil
	}
}
