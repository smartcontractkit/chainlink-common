package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	nested "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/nested/settings"
	simple "github.com/smartcontractkit/chainlink-common/x/config/cli/examples/simple/settings"
	"github.com/smartcontractkit/chainlink-common/x/config/markup/tomlmarkup"
)

func TestHelpText(t *testing.T) {
	type hasEveryRule struct {
		Required    string   `validate:"required"`
		Set         int      `validate:"set"` //nolint:revive // set is the Binder's own rule
		Dive        []string `validate:"dive,required"`
		Conditional string   `validate:"required_without=Required"`
		Optional    *RequiredField
	}

	for _, tc := range []struct {
		name     string
		target   any
		prefixes []string
		flag     string
		want     string
	}{
		{"a multi-line comment is one line", &simple.Config{}, nil, "timeout", "Timeout bounds a single request. A negative value is rejected."},
		{"an embedded struct is documented by its own type", &nested.Config{}, nil, "log-level", "LogLevel is the minimum level to log."},
		{"env vars in the order tried", &simple.Config{}, []string{"CRE", "CL"}, "host", "Host is the host to dial. [env CRE_HOST, CL_HOST]"},
		{"required", &hasEveryRule{}, nil, "required", "(required)"},
		{"set", &hasEveryRule{}, nil, "set", "(required)"},
		{"dive applies to the elements", &hasEveryRule{}, nil, "dive", ""},
		{"a conditional rule", &hasEveryRule{}, nil, "conditional", ""},
		// An optional section's rules apply only once it is configured.
		{"under an optional section", &hasEveryRule{}, nil, "optional.value", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, flagsOf(t, tc.target, Options{Markup: tomlmarkup.New(), Prefixes: tc.prefixes}).Lookup(tc.flag).Usage)
		})
	}
}

// pflag would print a string kind's raw default; its own MarshalText is what redacts it.
func TestSecretDefaultIsRedacted(t *testing.T) {
	type hasSetAndUnsetSecrets struct {
		Set   config.SecretString
		Unset config.SecretString
	}

	flags := flagsOf(t, &hasSetAndUnsetSecrets{Set: "hunter2"}, testOptions)
	assert.Equal(t, "xxxxx", flags.Lookup("set").DefValue)
	assert.Empty(t, flags.Lookup("unset").DefValue)

	c := hasSetAndUnsetSecrets{Set: "hunter2"}
	require.NoError(t, run(t, &c, testOptions))
	assert.Equal(t, config.SecretString("hunter2"), c.Set)
}
