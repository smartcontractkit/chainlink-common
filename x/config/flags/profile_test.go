package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- profiles ---

type profileChain struct {
	ID  uint32 `toml:"id"`
	RPC string `toml:"rpc"`
}

type profileCfg struct {
	Chain profileChain `toml:"chain"`
}

// runWithProfile registers c plus a profile map keyed on Chain.ID, then executes with args.
func runWithProfile(t *testing.T, c *profileCfg, args ...string) error {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "Chain.ID", map[uint32]profileCfg{
		1:   {Chain: profileChain{RPC: "https://one"}},
		137: {Chain: profileChain{RPC: "https://one-thirty-seven"}},
	}, DefaultTOMLOptions("TEST")))
	root.SetArgs(args)
	return root.Execute()
}

func TestProfileFillsDefaultsForSelectedKey(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c))
	assert.Equal(t, "https://one", c.Chain.RPC)
}

func TestProfileFollowsSelector(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.id", "137"))
	assert.Equal(t, uint32(137), c.Chain.ID)
	assert.Equal(t, "https://one-thirty-seven", c.Chain.RPC)
}

func TestProfileDoesNotOverrideExplicitValue(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.rpc", "https://mine"))
	assert.Equal(t, "https://mine", c.Chain.RPC)
}

func TestProfileUnknownSelectorAppliesNothing(t *testing.T) {
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, runWithProfile(t, c, "--chain.id", "999"))
	assert.Empty(t, c.Chain.RPC, "no matching profile means no defaults, not an error")
}

func TestProfileRequiresARegisteredTargetOfItsType(t *testing.T) {
	root := newRoot(t)
	err := RegisterProfile(root, "Chain.ID", map[uint32]profileCfg{1: {}}, DefaultTOMLOptions())
	require.ErrorContains(t, err, "no registered target")
}

func TestProfileRejectsAmbiguousTargetType(t *testing.T) {
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &profileCfg{}, DefaultTOMLOptions("A")))
	require.NoError(t, RegisterCommandFlags(root, &profileCfg{}, Options{Namespace: "b", Prefixes: []string{"B"},
		DecoderConfig: DefaultTOMLOptions().DecoderConfig}))

	err := RegisterProfile(root, "Chain.ID", map[uint32]profileCfg{1: {}}, DefaultTOMLOptions())
	require.ErrorContains(t, err, "multiple registered targets")
}

func TestProfileRejectsUnknownSelectorField(t *testing.T) {
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &profileCfg{}, DefaultTOMLOptions("TEST")))

	err := RegisterProfile(root, "Chain.Missing", map[uint32]profileCfg{1: {}}, DefaultTOMLOptions())
	require.ErrorContains(t, err, "not found")
}

// The selector's Go type has to match the profile map's key type, or the map could never be
// looked up with the decoded value.
func TestProfileRejectsSelectorTypeMismatch(t *testing.T) {
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &profileCfg{}, DefaultTOMLOptions("TEST")))

	err := RegisterProfile(root, "Chain.ID", map[string]profileCfg{"1": {}}, DefaultTOMLOptions())
	require.ErrorContains(t, err, "type mismatch")
}

// The selector may be named by its config key rather than its Go field name, so a caller that
// only knows the TOML layout can still register a profile.
func TestProfileSelectorMayBeNamedByItsConfigKey(t *testing.T) {
	root := newRoot(t)
	c := &profileCfg{Chain: profileChain{ID: 1}}
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "chain.id", map[uint32]profileCfg{
		1: {Chain: profileChain{RPC: "https://one"}},
	}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	assert.Equal(t, "https://one", c.Chain.RPC)
}

// A section the profile fills in may itself be absent from the target, in which case applying
// the profile is what allocates it.
func TestProfileAllocatesAPointerSection(t *testing.T) {
	type limits struct {
		Max int `toml:"max"`
	}
	type cfg struct {
		Env    string  `toml:"env"`
		Limits *limits `toml:"limits"`
	}

	root := newRoot(t)
	c := &cfg{Env: "prod"}
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "Env", map[string]cfg{
		"prod": {Limits: &limits{Max: 10}},
	}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	require.NotNil(t, c.Limits)
	assert.Equal(t, 10, c.Limits.Max)
}

// A pointer section the profile itself leaves nil stays nil rather than being allocated empty.
func TestProfileLeavesAPointerSectionNilWhenItHasNone(t *testing.T) {
	type limits struct {
		Max int `toml:"max"`
	}
	type cfg struct {
		Env    string  `toml:"env"`
		Limits *limits `toml:"limits"`
	}

	root := newRoot(t)
	c := &cfg{Env: "dev"}
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "Env", map[string]cfg{"dev": {}}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	assert.Nil(t, c.Limits)
}

// A list of structs has no flag of its own, so whether the user "set" it has to be decided from
// the column flags underneath it.
func TestProfileFillsAListOfStructsOnlyWhenNoColumnWasSet(t *testing.T) {
	type node struct {
		Name string `toml:"name"`
	}
	type cfg struct {
		Env   string `toml:"env"`
		Nodes []node `toml:"nodes"`
	}

	profiles := map[string]cfg{"prod": {Nodes: []node{{Name: "from-profile"}}}}

	t.Run("nothing set", func(t *testing.T) {
		root := newRoot(t)
		c := &cfg{Env: "prod"}
		require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
		require.NoError(t, RegisterProfile(root, "Env", profiles, DefaultTOMLOptions("TEST")))

		root.SetArgs(nil)
		require.NoError(t, root.Execute())
		assert.Equal(t, []node{{Name: "from-profile"}}, c.Nodes)
	})

	t.Run("a column set", func(t *testing.T) {
		root := newRoot(t)
		c := &cfg{Env: "prod"}
		require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
		require.NoError(t, RegisterProfile(root, "Env", profiles, DefaultTOMLOptions("TEST")))

		root.SetArgs([]string{"--nodes.name", "mine"})
		require.NoError(t, root.Execute())
		assert.Equal(t, []node{{Name: "mine"}}, c.Nodes)
	})
}

// Two profile maps can select on different fields of the same struct, each scoped to the section
// owning its selector so neither clobbers the other's branch.
func TestTwoProfilesScopeToTheirOwnSection(t *testing.T) {
	type chain struct {
		ID  uint32 `toml:"id"`
		RPC string `toml:"rpc"`
	}
	type system struct {
		Env      string `toml:"env"`
		LogLevel string `toml:"log-level"`
	}
	type cfg struct {
		Chain  chain  `toml:"chain"`
		System system `toml:"system"`
	}

	root := newRoot(t)
	c := &cfg{Chain: chain{ID: 1}, System: system{Env: "dev"}}
	require.NoError(t, RegisterCommandFlags(root, c, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "Chain.ID", map[uint32]cfg{
		1: {Chain: chain{RPC: "https://one"}},
	}, DefaultTOMLOptions("TEST")))
	require.NoError(t, RegisterProfile(root, "System.Env", map[string]cfg{
		"dev": {System: system{LogLevel: "debug"}},
	}, DefaultTOMLOptions("TEST")))

	root.SetArgs(nil)
	require.NoError(t, root.Execute())
	assert.Equal(t, "https://one", c.Chain.RPC)
	assert.Equal(t, "debug", c.System.LogLevel)
	assert.Equal(t, uint32(1), c.Chain.ID, "the selector keeps the value that picked the profile")
	assert.Equal(t, "dev", c.System.Env)
}
