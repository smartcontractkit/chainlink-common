package cli

import (
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListFromEverySource(t *testing.T) {
	type hasStringAndIntLists struct {
		Strings []string
		Ints    []int
	}

	for _, tc := range []struct {
		name, env string
		args      []string
		want      hasStringAndIntLists
	}{
		{"repeated flag", "", []string{"--strings", "a", "--strings", "b"}, hasStringAndIntLists{Strings: []string{"a", "b"}}},
		{"comma-separated env", "a,b", nil, hasStringAndIntLists{Strings: []string{"a", "b"}}},
		{"comma-separated flag, non-string elements", "", []string{"--ints", "1,2,3"}, hasStringAndIntLists{Ints: []int{1, 2, 3}}},
		{"an element holding the separator is quoted", "", []string{"--strings", `"a,b",c`}, hasStringAndIntLists{Strings: []string{"a,b", "c"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var c hasStringAndIntLists
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_STRINGS", tc.env, "", tc.args...)...))
			assert.Equal(t, tc.want, c)
		})
	}
}

// A byte slice is text (JSON, PEM), so neither a flag nor an env var splits it on its commas.
func TestByteSliceIsText(t *testing.T) {
	type hasByteSlice struct {
		Bytes []byte
	}

	assert.Equal(t, `{"a":1}`, flagsOf(t, &hasByteSlice{Bytes: []byte(`{"a":1}`)}, testOptions).Lookup("bytes").DefValue)
	for _, tc := range everySource([]string{"--bytes", `{"a":1,"b":2}`}, `{"a":1,"b":2}`, "") {
		t.Run(tc.name, func(t *testing.T) {
			var c hasByteSlice
			require.NoError(t, run(t, &c, testOptions, supply(t, "TEST_BYTES", tc.env, tc.file, tc.args...)...))
			assert.Equal(t, []byte(`{"a":1,"b":2}`), c.Bytes)
		})
	}
}

func TestSelfDecodingSliceOwnsItsForm(t *testing.T) {
	type hasSelfSplittingList struct {
		List selfSplittingList
	}
	t.Setenv("TEST_LIST", "80+443")

	var c hasSelfSplittingList
	require.NoError(t, run(t, &c, testOptions))
	assert.Equal(t, selfSplittingList{80, 443}, c.List)
}

type selfSplittingList []int

func (p *selfSplittingList) UnmarshalText(b []byte) error {
	*p = nil
	for part := range strings.SplitSeq(string(b), "+") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		*p = append(*p, n)
	}
	return nil
}
