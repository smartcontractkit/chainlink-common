package mercury_v3

import (
	"github.com/pkg/errors"
	"math"
)

func ValidateValidFromTimestamp(observationTimestamp uint32, validFromTimestamp uint32) error {
	if observationTimestamp < validFromTimestamp {
		return errors.Errorf("observationTimestamp (%d) must be >= validFromTimestamp (%d)", observationTimestamp, validFromTimestamp)
	}

	return nil
}

func ValidateExpiresAt(observationTimestamp uint32, expirationWindow uint32) error {
	if int64(observationTimestamp)+int64(expirationWindow) > math.MaxUint32 {
		return errors.Errorf("timestamp %d + expiration window %d overflows uint32", observationTimestamp, expirationWindow)
	}

	return nil
}
