package execution_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/execution"

	"github.com/stretchr/testify/require"
)

func TestInterpolateKey(t *testing.T) {
	t.Parallel()
	val, err := values.NewMap(
		map[string]any{
			"reports": map[string]any{
				"inner": "key",
			},
			"reportsList": []any{
				"listElement",
			},
		},
	)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		key      string
		state    fakeResults
		expected any
		errMsg   string
	}{
		{
			name: "digging into a string",
			key:  "evm_median.outputs.reports",
			state: fakeResults{
				"evm_median": &execution.Result{
					Outputs: values.NewString("<a report>"),
				},
			},
			errMsg: "could not interpolate ref part `reports` (ref: `evm_median.outputs.reports`) in `<a report>`",
		},
		{
			name:   "ref doesn't exist",
			key:    "evm_median.outputs.reports",
			state:  fakeResults{},
			errMsg: "could not find ref `evm_median`",
		},
		{
			name:   "less than 2 parts",
			key:    "evm_median",
			state:  fakeResults{},
			errMsg: "must have at least two parts",
		},
		{
			name: "second part isn't `inputs` or `outputs`",
			key:  "evm_median.foo",
			state: fakeResults{
				"evm_median": {
					Outputs: values.NewString("<a report>"),
				},
			},
			errMsg: "second part must be `inputs` or `outputs`",
		},
		{
			name: "outputs has errored",
			key:  "evm_median.outputs",
			state: fakeResults{
				"evm_median": {
					Error: errors.New("catastrophic error"),
				},
			},
			errMsg: "step has errored",
		},
		{
			name: "digging into a recursive map",
			key:  "evm_median.outputs.reports.inner",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			expected: "key",
		},
		{
			name: "missing key in map",
			key:  "evm_median.outputs.reports.missing",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			errMsg: "could not find ref part `missing` (ref: `evm_median.outputs.reports.missing`) in",
		},
		{
			name: "digging into an array",
			key:  "evm_median.outputs.reportsList.0",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			expected: "listElement",
		},
		{
			name: "digging into an array that's too small",
			key:  "evm_median.outputs.reportsList.2",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			errMsg: "index out of bounds 2",
		},
		{
			name: "digging into an array with a string key",
			key:  "evm_median.outputs.reportsList.notAString",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			errMsg: "could not interpolate ref part `notAString` (ref: `evm_median.outputs.reportsList.notAString`) in `[listElement]`: `notAString` is not convertible to an int",
		},
		{
			name: "digging into an array with a negative index",
			key:  "evm_median.outputs.reportsList.-1",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			errMsg: "could not interpolate ref part `-1` (ref: `evm_median.outputs.reportsList.-1`) in `[listElement]`: index out of bounds -1",
		},
		{
			name: "empty element",
			key:  "evm_median.outputs..notAString",
			state: fakeResults{
				"evm_median": {
					Outputs: val,
				},
			},
			errMsg: "could not find ref part `` (ref: `evm_median.outputs..notAString`) in",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			got, err := execution.InterpolateKey(tc.key, tc.state)
			if tc.errMsg != "" {
				require.ErrorContains(st, err, tc.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestInterpolateInputsFromState(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		inputs   map[string]any
		state    fakeResults
		expected any
		errMsg   string
	}{
		{
			name: "substituting with a variable that exists",
			inputs: map[string]any{
				"shouldnotinterpolate": map[string]any{
					"shouldinterpolate": "$(evm_median.outputs)",
				},
			},
			state: fakeResults{
				"evm_median": {
					Outputs: values.NewString("<a report>"),
				},
			},
			expected: map[string]any{
				"shouldnotinterpolate": map[string]any{
					"shouldinterpolate": "<a report>",
				},
			},
		},
		{
			name: "no substitution required",
			inputs: map[string]any{
				"foo": "bar",
			},
			state: fakeResults{
				"evm_median": {
					Outputs: values.NewString("<a report>"),
				},
			},
			expected: map[string]any{
				"foo": "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(st *testing.T) {
			got, err := execution.FindAndInterpolateAllKeys(tc.inputs, tc.state)
			if tc.errMsg != "" {
				require.ErrorContains(st, err, tc.errMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

type fakeResults map[string]*execution.Result

func (f fakeResults) GetResultForStep(s string) (*execution.Result, bool) {
	r, ok := f[s]
	return r, ok
}

var _ execution.Results = fakeResults{}
