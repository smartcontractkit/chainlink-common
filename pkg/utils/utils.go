package utils

import (
	"context"
	"encoding/base64"
	"math"
	"math/big"
	mrand "math/rand"
	"reflect"
	"time"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

// WithJitter adds +/- 10% to a duration
func WithJitter(d time.Duration) time.Duration {
	// #nosec
	if d == 0 {
		return 0
	}
	// ensure non-zero arg to Intn to avoid panic
	max := math.Max(float64(d.Abs())/5.0, 1.)
	// #nosec - non critical randomness
	jitter := mrand.Intn(int(max))
	jitter = jitter - (jitter / 2)
	return time.Duration(int(d) + jitter)
}

// ContextFromChan creates a context that finishes when the provided channel
// receives or is closed.
// When channel closes, the ctx.Err() will always be context.Canceled
// NOTE: Spins up a goroutine that exits on cancellation.
// REMEMBER TO CALL CANCEL OTHERWISE IT CAN LEAD TO MEMORY LEAKS
func ContextFromChan(chStop <-chan struct{}) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		select {
		case <-chStop:
			cancel()
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}

// ContextWithDeadlineFn returns a copy of the parent context with the deadline modified by deadlineFn.
// deadlineFn will only be called if the parent has a deadline.
// The new deadline must be sooner than the old to have an effect.
func ContextWithDeadlineFn(ctx context.Context, deadlineFn func(orig time.Time) time.Time) (context.Context, context.CancelFunc) {
	cancel := func() {}
	if d, ok := ctx.Deadline(); ok {
		if m := deadlineFn(d); m.Before(d) {
			ctx, cancel = context.WithDeadline(ctx, m)
		}
	}
	return ctx, cancel
}

func FitsInNBitsSigned(n int, bi *big.Int) bool {
	if bi.Sign() < 0 {
		bi = new(big.Int).Neg(bi)
		bi.Sub(bi, big.NewInt(1))
	}
	return bi.BitLen() <= n-1
}

func MergeValueFields(valueFields []map[string]any) (map[string]any, error) {
	numItems := len(valueFields)

	switch numItems {
	case 0:
		return map[string]any{}, nil
	default:
		mergedReflect := map[string]reflect.Value{}
		for k, v := range valueFields[0] {
			rv := reflect.ValueOf(v)
			slice := reflect.MakeSlice(reflect.SliceOf(rv.Type()), numItems, numItems)
			slice.Index(0).Set(rv)
			mergedReflect[k] = slice
		}

		for i, valueField := range valueFields[1:] {
			if len(valueField) != len(mergedReflect) {
				return nil, types.InvalidTypeError{}
			}

			for k, slice := range mergedReflect {
				if value, ok := valueField[k]; ok {
					sliceElm := slice.Index(i + 1)
					rv := reflect.ValueOf(value)
					if !rv.Type().AssignableTo(sliceElm.Type()) {
						return nil, types.InvalidTypeError{}
					}
					sliceElm.Set(rv)
				} else {
					return nil, types.InvalidTypeError{}
				}
			}
		}

		merged := map[string]any{}

		for k, v := range mergedReflect {
			merged[k] = v.Interface()
		}

		return merged, nil
	}
}

func SplitValueFields(decoded map[string]any) ([]map[string]any, error) {
	var result []map[string]any

	for k, v := range decoded {
		iv := reflect.ValueOf(v)
		kind := iv.Kind()
		if kind != reflect.Slice && kind != reflect.Array {
			if kind != reflect.String {
				return nil, types.NotASliceError{}
			}
			rawBytes, err := base64.StdEncoding.DecodeString(v.(string))
			if err != nil {
				return nil, types.InvalidTypeError{}
			}
			iv = reflect.ValueOf(rawBytes)
		}

		length := iv.Len()
		if result == nil {
			result = make([]map[string]any, length)
			for i := 0; i < length; i++ {
				result[i] = map[string]any{}
			}
		}

		if len(result) != length {
			return nil, types.InvalidTypeError{}
		}

		for i := 0; i < length; i++ {
			result[i][k] = iv.Index(i).Interface()
		}
	}

	return result, nil
}
