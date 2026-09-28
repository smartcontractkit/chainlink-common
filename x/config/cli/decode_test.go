package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExplicitDefaultValueOverridesANonDefault(t *testing.T) {
	flag := []string{"--bool=false", "--uint", "0", "--string", ""}
	for _, tc := range everySource(flag, "", "Bool = false\nUint = 0\nString = ''") {
		t.Run(tc.name, func(t *testing.T) {
			c := hasEveryKind{Bool: true, Uint: 5, String: "set"}
			require.NoError(t, run(t, &c, testOptions, supply(t, "", "", tc.file, tc.args...)...))
			assert.Equal(t, hasEveryKind{}, c)
		})
	}
}
