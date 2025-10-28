package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Rate includes a per second limit and a burst.
// Rates >= 1 are printed as "<rate>rps:<burst>". Rates < 1 are printed as "every<duration>:<burst>".
type Rate struct {
	Limit rate.Limit
	Burst int
}

// ParseRate parses a two part rate limit & burst.
// They must be separated by a ':'. Limit comes first, either in the form every<int duration>, or <float count>rps. Burst comes
// second, as an integer.
func ParseRate(s string) (Rate, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return Rate{}, fmt.Errorf("invalid rate limit: %s: must contain two parts", s)
	}
	var rateLimit Rate
	if after, ok := strings.CutPrefix(parts[0], "every"); ok { // optional format every<duration>
		limit := after
		d, err := time.ParseDuration(limit)
		if err != nil {
			return Rate{}, err
		}
		rateLimit.Limit = rate.Every(d)
	} else {
		parts[0] = strings.TrimSuffix(parts[0], "rps") // allowed but not required
		f, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			return Rate{}, err
		}
		rateLimit.Limit = rate.Limit(f)
	}
	burst, err := strconv.Atoi(parts[1])
	if err != nil {
		return Rate{}, err
	}
	rateLimit.Burst = burst
	return rateLimit, nil
}

func (r Rate) String() string {
	if r.Limit < 1 {
		return fmt.Sprintf("every%s:%d", time.Duration(1/r.Limit)*time.Second, r.Burst)
	}
	return fmt.Sprintf("%grps:%d", r.Limit, r.Burst)
}
