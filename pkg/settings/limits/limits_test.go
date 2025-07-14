package limits

import "time"

func init() {
	pollPeriod = time.Second // speed up tests
}
