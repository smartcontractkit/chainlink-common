package execution

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

// InterpolateKey takes a multi-part, dot-separated key and attempts to replace
// it with its corresponding value in `state`.
//
// A key is valid if it contains at least two parts, with:
//   - the first part being the workflow step's `ref` variable
//   - the second part being one of `inputs` or `outputs`
//
// If a key has more than two parts, then we traverse the parts
// to find the value we want to replace.
// We support traversing both nested maps and lists and any combination of the two.
func InterpolateKey(key string, state Results) (any, error) {
	parts := strings.Split(key, ".")

	if len(parts) < 2 {
		return "", fmt.Errorf("cannot interpolate %s: must have at least two parts", key)
	}

	// lookup the step we want to get either input or output state from
	sc, ok := state.GetResultForStep(parts[0])
	if !ok {
		return "", fmt.Errorf("could not find ref `%s`", parts[0])
	}

	var value values.Value
	switch parts[1] {
	case "inputs":
		value = sc.Inputs
	case "outputs":
		if sc.Error != nil {
			return "", fmt.Errorf("cannot interpolate ref part `%s` in `%+v`: step has errored", parts[1], sc)
		}

		value = sc.Outputs
	default:
		return "", fmt.Errorf("cannot interpolate ref part `%s` in `%+v`: second part must be `inputs` or `outputs`", parts[1], sc)
	}

	val, err := values.Unwrap(value)
	if err != nil {
		return "", err
	}

	remainingParts := parts[2:]
	for _, r := range remainingParts {
		switch v := val.(type) {
		case map[string]any:
			inner, ok := v[r]
			if !ok {
				return "", fmt.Errorf("could not find ref part `%s` (ref: `%s`) in `%+v`", r, key, v)
			}

			val = inner
		case []any:
			i, err := strconv.Atoi(r)
			if err != nil {
				return "", fmt.Errorf("could not interpolate ref part `%s` (ref: `%s`) in `%+v`: `%s` is not convertible to an int", r, key, v, r)
			}

			if (i > len(v)-1) || (i < 0) {
				return "", fmt.Errorf("could not interpolate ref part `%s` (ref: `%s`) in `%+v`: index out of bounds %d", r, key, v, i)
			}

			val = v[i]
		default:
			return "", fmt.Errorf("could not interpolate ref part `%s` (ref: `%s`) in `%+v`", r, key, val)
		}
	}

	return val, nil
}

// FindAndInterpolateAllKeys takes an `input` any value, and recursively
// identifies any values that should be replaced from `state`.
//
// A value `v` should be replaced if it is wrapped as follows: `$(v)`.
func FindAndInterpolateAllKeys(input any, state Results) (any, error) {
	return workflows.DeepMap(
		input,
		func(el string) (any, error) {
			matches := workflows.InterpolationTokenRe.FindStringSubmatch(el)
			if len(matches) < 2 {
				return el, nil
			}

			interpolatedVar := matches[1]
			return InterpolateKey(interpolatedVar, state)
		},
	)
}
