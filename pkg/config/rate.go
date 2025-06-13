package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Rate includes a per second limit and a burst.
// Rates >= 1 are printed as "<rate>rps". Rates < 1 are printed as "every<duration>".
type Rate struct {
	Limit rate.Limit
	Burst int
}

// ParseRate parses a two part rate limit & burst.
// They must be separated by a ','. Limit comes first, either in the form every<int duration>, or <float count>rps. Burst comes
// second, as an integer.
func ParseRate(s string) (Rate, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return Rate{}, fmt.Errorf("invalid rate limit: %s: must contain two parts", s)
	}
	var rateLimit Rate
	if strings.HasPrefix(parts[0], "every") { // optional format every<duration>
		limit := strings.TrimPrefix(parts[0], "every")
		d, err := time.ParseDuration(limit)
		if err != nil {
			return Rate{}, err
		}
		rateLimit.Limit = rate.Every(d)
	} else {
		s = strings.TrimSuffix(s, "rps") // allowed but not required
		f, err := strconv.ParseFloat(s, 64)
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
		return fmt.Sprintf("every%s,%d", time.Duration(1/r.Limit)*time.Second, r.Burst)
	}
	return fmt.Sprintf("%grps,%d", r.Limit, r.Burst)
}
