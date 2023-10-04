package servicetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

func AssertHealthReportNames(t *testing.T, hp map[string]error, names ...string) {
	t.Helper()
	keys := maps.Keys(hp)
	slices.Sort(keys)
	slices.Sort(names)
	assert.EqualValues(t, names, keys)
}
