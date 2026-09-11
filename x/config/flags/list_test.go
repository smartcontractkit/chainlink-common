package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- slices ---

func TestStringSliceFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		URLs []string `toml:"url"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--url", "a", "--url", "b"))
	assert.Equal(t, []string{"a", "b"}, c.URLs)
}

func TestStringSliceFromCommaSeparatedEnv(t *testing.T) {
	type cfg struct {
		URLs []string `toml:"url"`
	}

	t.Setenv("TEST_URL", "a,b")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, []string{"a", "b"}, c.URLs)
}

// A list of a non-string primitive splits on commas the way a []string does - mapstructure's
// own StringToSliceHookFunc only ever splits into []string, so this needs the wider hook that
// map-of-list values needed anyway.
func TestNonStringSliceFromCommaSeparatedFlag(t *testing.T) {
	type cfg struct {
		Ports []int `toml:"ports"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--ports", "1,2,3"))
	assert.Equal(t, []int{1, 2, 3}, c.Ports)
}

// A list of a non-string primitive is written the same way a []string is, and repeating the
// flag accumulates rather than replacing.
func TestNonStringSliceFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		Ports []int `toml:"ports"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--ports", "1,2", "--ports", "3"))
	assert.Equal(t, []int{1, 2, 3}, c.Ports)
}

// An element that contains the separator is written in quotes, since the flag's value is
// re-parsed on the way to the decoder.
func TestSliceElementKeepsAQuotedComma(t *testing.T) {
	type cfg struct {
		Tags []string `toml:"tags"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--tags", `"a,b",c`))
	assert.Equal(t, []string{"a,b", "c"}, c.Tags)
}

// A list of lists is built one inner list per occurrence: at depth two a comma separates the
// elements of an inner list, not the outer ones, so this is the repeated-flag form.
func TestListOfListsFromRepeatedFlag(t *testing.T) {
	type cfg struct {
		Groups [][]string `toml:"groups"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--groups", "a,b", "--groups", "c"))
	assert.Equal(t, [][]string{{"a", "b"}, {"c"}}, c.Groups)
}

// The same value written at once, which is the only form an env var can take.
func TestListOfListsFromOneFlagAndFromEnv(t *testing.T) {
	type cfg struct {
		Groups [][]int `toml:"groups"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--groups", "[1,2],[3]"))
	assert.Equal(t, [][]int{{1, 2}, {3}}, c.Groups)

	t.Setenv("TEST_GROUPS", "[1,2],[3]")
	var fromEnv cfg
	require.NoError(t, run(t, &fromEnv))
	assert.Equal(t, [][]int{{1, 2}, {3}}, fromEnv.Groups)
}

// A single bracketed occurrence is one inner list, not a whole value - "[a,b]" and "a,b" mean
// the same thing, which is what makes the repeated form and the all-at-once form agree.
func TestBracketedOccurrenceIsOneInnerList(t *testing.T) {
	type cfg struct {
		Groups [][]string `toml:"groups"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--groups", "[a,b]"))
	assert.Equal(t, [][]string{{"a", "b"}}, c.Groups)
}

// A list of lists round-trips through its own default: viper hands an unchanged flag's own
// string to the decoder, so the rendered form has to read back as the value it was rendered
// from.
func TestListOfListsDefaultRoundTrips(t *testing.T) {
	type cfg struct {
		Groups [][]string `toml:"groups"`
	}

	c := cfg{Groups: [][]string{{"a", "b"}, {"c"}}}
	require.NoError(t, run(t, &c, "--groups", "[x],[y]"))
	assert.Equal(t, [][]string{{"x"}, {"y"}}, c.Groups)

	root := newRoot(t)
	unchanged := cfg{Groups: [][]string{{"a", "b"}, {"c"}}}
	require.NoError(t, RegisterCommandFlags(root, &unchanged, DefaultTOMLOptions("TEST")))
	assert.Equal(t, "[[a,b],[c]]", root.PersistentFlags().Lookup("groups").DefValue)
}

// A byte slice is text - a JSON blob, a PEM key - so its default is the text itself and not the
// Go rendering of a list of numbers, which could not be read back.
func TestByteSliceIsBoundAsText(t *testing.T) {
	type cfg struct {
		Raw []byte `toml:"raw"`
	}

	c := cfg{Raw: []byte(`{"a":1}`)}
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))
	assert.Equal(t, `{"a":1}`, root.PersistentFlags().Lookup("raw").DefValue)

	var set cfg
	require.NoError(t, run(t, &set, "--raw", `{"b":2}`))
	assert.Equal(t, []byte(`{"b":2}`), set.Raw)
}

// A list with no text form at all gets no flag, the way an unsupported map doesn't: a flag whose
// default cannot be read back is worse than no flag.
func TestListWithNoTextFormGetsNoFlag(t *testing.T) {
	type cfg struct {
		Raw  []map[string]string `toml:"raw"`
		Name string              `toml:"name"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))
	assert.Nil(t, root.PersistentFlags().Lookup("raw"))
	assert.NotNil(t, root.PersistentFlags().Lookup("name"))
}

// --- lists of structs ---

type listNode struct {
	Name string   `toml:"name"`
	URL  string   `toml:"url"`
	Tags []string `toml:"tags"`
}

type listChain struct {
	ID    string     `toml:"id"`
	Nodes []listNode `toml:"nodes"`
}

// A list of structs is written one column per field, the columns lining up by position.
func TestListOfStructsFromFlags(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--nodes.name", "a,b", "--nodes.url", "ua,ub"))
	assert.Equal(t, []listNode{{Name: "a", URL: "ua"}, {Name: "b", URL: "ub"}}, c.Nodes)
}

// A column shorter than the longest one leaves that field at its zero value rather than
// truncating the list or shifting the values that follow it.
func TestListOfStructsWithAShortColumn(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--nodes.name", "a,b,c", "--nodes.url", "ua"))
	assert.Equal(t, []listNode{{Name: "a", URL: "ua"}, {Name: "b"}, {Name: "c"}}, c.Nodes)
}

// A list-valued field inside a list of structs is a list of lists, so it takes the same
// repeated-flag form as any other.
func TestListOfStructsWithAListField(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	var c cfg
	require.NoError(t, run(t, &c, "--nodes.name", "a,b", "--nodes.tags", "t1,t2", "--nodes.tags", "t3"))
	assert.Equal(t, []listNode{
		{Name: "a", Tags: []string{"t1", "t2"}},
		{Name: "b", Tags: []string{"t3"}},
	}, c.Nodes)
}

func TestListOfStructsFromEnv(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	t.Setenv("TEST_NODES_NAME", "a,b")
	t.Setenv("TEST_NODES_URL", "ua,ub")

	var c cfg
	require.NoError(t, run(t, &c))
	assert.Equal(t, []listNode{{Name: "a", URL: "ua"}, {Name: "b", URL: "ub"}}, c.Nodes)
}

// A config file writes the same field row-wise, as an array of tables, which is what a config
// file is good at.
func TestListOfStructsFromConfigFile(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	path := writeConfig(t, "[[nodes]]\nname = 'a'\nurl = 'ua'\n\n[[nodes]]\nname = 'b'\nurl = 'ub'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path))
	assert.Equal(t, []listNode{{Name: "a", URL: "ua"}, {Name: "b", URL: "ub"}}, c.Nodes)
}

// The two forms merge per field, the way every other key does: a flag overrides the one field it
// names in the row it lands on, and leaves the rest of that row alone.
func TestListOfStructsFlagOverridesOneFieldOfAConfigFileRow(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	path := writeConfig(t, "[[nodes]]\nname = 'a'\nurl = 'ua'\n\n[[nodes]]\nname = 'b'\nurl = 'ub'\n")

	var c cfg
	require.NoError(t, run(t, &c, "--config", path, "--nodes.url", "over"))
	assert.Equal(t, []listNode{{Name: "a", URL: "over"}, {Name: "b", URL: "ub"}}, c.Nodes)
}

// A list of structs inside a list of structs adds a list level rather than a new kind of key.
func TestNestedListsOfStructs(t *testing.T) {
	type cfg struct {
		Chains []listChain `toml:"chains"`
	}

	var c cfg
	require.NoError(t, run(t, &c,
		"--chains.id", "1,2",
		"--chains.nodes.name", "[a,b],[c]",
	))
	assert.Equal(t, []listChain{
		{ID: "1", Nodes: []listNode{{Name: "a"}, {Name: "b"}}},
		{ID: "2", Nodes: []listNode{{Name: "c"}}},
	}, c.Chains)
}

// Two list levels plus a list-valued leaf is three, and reads the same way.
func TestNestedListsOfStructsWithAListField(t *testing.T) {
	type cfg struct {
		Chains []listChain `toml:"chains"`
	}

	var c cfg
	require.NoError(t, run(t, &c,
		"--chains.id", "1",
		"--chains.nodes.name", "[a,b]",
		"--chains.nodes.tags", "[[t1,t2],[t3]]",
	))
	assert.Equal(t, []listChain{{ID: "1", Nodes: []listNode{
		{Name: "a", Tags: []string{"t1", "t2"}},
		{Name: "b", Tags: []string{"t3"}},
	}}}, c.Chains)
}

// Nothing set means nothing decoded, so the struct keeps whatever it was constructed with -
// the same rule every other key follows.
func TestListOfStructsKeepsItsCompiledInDefault(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	c := cfg{Nodes: []listNode{{Name: "keep", URL: "u"}}}
	require.NoError(t, run(t, &c))
	assert.Equal(t, []listNode{{Name: "keep", URL: "u"}}, c.Nodes)
}

// That default is what each column's flag shows, pivoted the same way the flags are.
func TestListOfStructsDefaultsAreShownColumnWise(t *testing.T) {
	type cfg struct {
		Nodes []listNode `toml:"nodes"`
	}

	c := cfg{Nodes: []listNode{{Name: "a", Tags: []string{"x", "y"}}, {Name: "b"}}}
	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &c, DefaultTOMLOptions("TEST")))

	assert.Equal(t, "[a,b]", root.PersistentFlags().Lookup("nodes.name").DefValue)
	assert.Equal(t, "[[x,y],[]]", root.PersistentFlags().Lookup("nodes.tags").DefValue)
}

// A field inside a list of structs that has no text form to repeat stays config-file-only, the
// same way it would outside one.
func TestFieldWithNoTextFormInsideAListGetsNoFlag(t *testing.T) {
	type node struct {
		Name   string              `toml:"name"`
		Labels map[string]struct{} `toml:"labels"`
	}
	type cfg struct {
		Nodes []node `toml:"nodes"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))
	assert.NotNil(t, root.PersistentFlags().Lookup("nodes.name"))
	assert.Nil(t, root.PersistentFlags().Lookup("nodes.labels"))
}

// A struct that contains itself describes an unbounded set of keys; the walk stops at the
// second occurrence of the type instead of running out of stack.
func TestSelfReferentialStructTerminates(t *testing.T) {
	type node struct {
		Name     string  `toml:"name"`
		Children []*node `toml:"children"`
	}
	type cfg struct {
		Root node `toml:"root"`
	}

	root := newRoot(t)
	require.NoError(t, RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST")))
	assert.NotNil(t, root.PersistentFlags().Lookup("root.name"))
	assert.Nil(t, root.PersistentFlags().Lookup("root.children.name"))
}

// The squash option means "flatten this struct into its parent", which only a struct can do.
// TOML's own `,inline` says something else entirely and config structs do carry it on plain
// scalars; honouring it there would drop those fields' key segments and collapse every one of
// them onto their parent's flag name, and mapstructure rejects it too, so it's reported at
// registration rather than as a duplicate flag or a decode failure later.
func TestSquashTagOnANonStructIsRejected(t *testing.T) {
	type inner struct {
		A *uint32 `toml:",inline"`
		B *uint32 `toml:",inline"`
	}
	type cfg struct {
		Limits inner `toml:"limits"`
	}

	root := newRoot(t)
	err := RegisterCommandFlags(root, &cfg{}, DefaultTOMLOptions("TEST"))
	require.ErrorContains(t, err, `limits.a: the "inline" tag option flattens a struct into its parent`)
}
