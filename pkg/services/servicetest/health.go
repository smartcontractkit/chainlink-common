package servicetest

import (
	"strings"
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
	join := func(s []string) string { return strings.Join(s, "\n") }
	assert.Equal(t, join(names), join(keys))
}
