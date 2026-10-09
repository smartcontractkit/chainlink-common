package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type profiledSection struct {
	Selector uint32
	Profiled string
	Count    int
}

func selector(s *profiledSection) *uint32 { return &s.Selector }

var sectionProfiles = map[uint32]func(*profiledSection){
	1:   func(s *profiledSection) { s.Profiled = "one" },
	137: func(s *profiledSection) { s.Profiled = "137"; s.Count = 0 },
}

func withProfile[T any, K comparable](t *testing.T, c *T, key func(*T) *K, profiles map[K]func(*T)) *cobra.Command {
	t.Helper()

	root := newRoot(t)
	require.NoError(t, newBinder(t, testOptions).RegisterInNamespace(root, "Section", c, Profile(key, profiles)))
	return root
}

func TestProfileSelection(t *testing.T) {
	for _, tc := range []struct {
		name, file string
		section    profiledSection
		args       []string
		want       profiledSection
	}{
		{name: "a default selector picks its profile", section: profiledSection{Selector: 1},
			want: profiledSection{Selector: 1, Profiled: "one"}},
		{name: "a changed selector picks the new one", section: profiledSection{Selector: 1}, args: []string{"--section.selector", "137"},
			want: profiledSection{Selector: 137, Profiled: "137"}},
		{name: "it replaces a compiled-in default", section: profiledSection{Selector: 1, Profiled: "compiled"},
			want: profiledSection{Selector: 1, Profiled: "one"}},
		{name: "it can set a zero", section: profiledSection{Selector: 137, Count: 5},
			want: profiledSection{Selector: 137, Profiled: "137"}},
		{name: "it builds on the defaults", section: profiledSection{Selector: 1, Count: 5},
			want: profiledSection{Selector: 1, Profiled: "one", Count: 5}},
		{name: "a source beats it", section: profiledSection{Selector: 1}, file: "[Section]\nProfiled = 'file'",
			want: profiledSection{Selector: 1, Profiled: "file"}},
		// Such as a chain that isn't built in.
		{name: "a selector with no profile leaves the section to the sources", section: profiledSection{Selector: 1, Profiled: "compiled"},
			args: []string{"--section.selector", "999"}, want: profiledSection{Selector: 999, Profiled: "compiled"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := tc.section
			root := withProfile(t, &c, selector, sectionProfiles)
			root.SetArgs(supplyConfig(t, "", "", tc.file, tc.args...))
			require.NoError(t, root.Execute())
			assert.Equal(t, tc.want, c)
		})
	}
}

func TestProfileCannotSwitchItsSelector(t *testing.T) {
	c := profiledSection{Selector: 1}
	profiles := map[uint32]func(*profiledSection){1: func(s *profiledSection) { s.Selector = 137 }}
	require.NoError(t, withProfile(t, &c, selector, profiles).Execute())
	assert.Equal(t, uint32(1), c.Selector)
}

func TestProfileIsRejected(t *testing.T) {
	type nestsProfiledSection struct {
		Section profiledSection
		List    []profiledSection
	}
	var elsewhere uint32
	none := map[uint32]func(*nestsProfiledSection){}
	want := "profile key must return a top-level field of its section bound as a flag"

	for _, key := range []func(*nestsProfiledSection) *uint32{
		func(*nestsProfiledSection) *uint32 { return &elsewhere },
		func(*nestsProfiledSection) *uint32 { return nil },
		func(c *nestsProfiledSection) *uint32 { return &c.Section.Selector },
	} {
		require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &nestsProfiledSection{}, Profile(key, none)), want)
	}

	require.ErrorContains(t, newBinder(t, testOptions).Register(newRoot(t), &nestsProfiledSection{}, ProfileWithSelector(
		func(c *nestsProfiledSection) *profiledSection { return &c.List[0] }, selector, sectionProfiles)),
		"profile selector panicked: runtime error: index out of range")
}

func TestProfileHelpListsWhatEachProfileSets(t *testing.T) {
	type hasRequired struct {
		Selector uint32
		Profiled string
		Count    int
		Needed   string `validate:"required"`
	}
	profiles := map[uint32]func(*hasRequired){
		10: func(s *hasRequired) { s.Profiled = "ten"; s.Needed = "given" },
		1:  func(s *hasRequired) { s.Profiled = "" },
	}

	c := hasRequired{Profiled: "default", Count: 5}
	root := withProfile(t, &c, func(s *hasRequired) *uint32 { return &s.Selector }, profiles)
	var out strings.Builder
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})
	require.NoError(t, root.Execute())
	want := "AVAILABLE PROFILES FOR --section.selector:\n" +
		"  1:\n    section.profiled = \n    section.count = 5\n" +
		"  10:\n    section.profiled = ten\n    section.count = 5\n    section.needed = given\n"
	assert.True(t, strings.HasSuffix(out.String(), want), out.String())
}

func TestProfileOwnsAPointerSection(t *testing.T) {
	type hasValue struct {
		Value int
	}
	type hasPointerSection struct {
		Selector string
		Section  *hasValue
	}
	profiles := map[string]func(*hasPointerSection){
		"allocates": func(c *hasPointerSection) { c.Section = &hasValue{Value: 10} },
		"leaves":    func(*hasPointerSection) {},
	}
	selector := func(c *hasPointerSection) *string { return &c.Selector }

	allocated := hasPointerSection{Selector: "allocates"}
	require.NoError(t, withProfile(t, &allocated, selector, profiles).Execute())
	assert.Equal(t, hasPointerSection{"allocates", &hasValue{10}}, allocated)

	left := hasPointerSection{Selector: "leaves"}
	require.NoError(t, withProfile(t, &left, selector, profiles).Execute())
	assert.Equal(t, hasPointerSection{Selector: "leaves"}, left)
}

func TestProfileHelpOmitsANilSection(t *testing.T) {
	type hasValue struct {
		Value int
	}
	type hasPointerSection struct {
		Selector string
		Section  *hasValue
	}
	profiles := map[string]func(*hasPointerSection){"nils": func(c *hasPointerSection) { c.Section = nil }}

	c := hasPointerSection{Section: &hasValue{Value: 1}}
	root := withProfile(t, &c, func(c *hasPointerSection) *string { return &c.Selector }, profiles)
	var out strings.Builder
	root.SetOut(&out)
	root.SetArgs([]string{"--help"})
	require.NoError(t, root.Execute())
	assert.True(t, strings.HasSuffix(out.String(), "AVAILABLE PROFILES FOR --section.selector:\n  nils:\n"), out.String())
}

func TestTwoProfilesOnOneCommand(t *testing.T) {
	type profiledByString struct {
		Selector string
		Profiled string
	}

	root := newRoot(t)
	b := newBinder(t, testOptions)
	section, other := profiledSection{Selector: 1}, profiledByString{Selector: "s"}
	require.NoError(t, b.RegisterInNamespace(root, "Section", &section, Profile(selector, sectionProfiles)))
	require.NoError(t, b.RegisterInNamespace(root, "Other", &other, Profile(func(o *profiledByString) *string { return &o.Selector },
		map[string]func(*profiledByString){"s": func(o *profiledByString) { o.Profiled = "other" }})))
	require.NoError(t, root.Execute())
	assert.Equal(t, profiledSection{1, "one", 0}, section)
	assert.Equal(t, profiledByString{"s", "other"}, other)
}

func TestProfileLeavesAMapEntryASourceSet(t *testing.T) {
	type hasMap struct {
		Selector uint32
		Labels   map[string]string
	}
	profiles := map[uint32]func(*hasMap){1: func(c *hasMap) { c.Labels = map[string]string{"a": "profile"} }}

	c := hasMap{Selector: 1}
	root := withProfile(t, &c, func(c *hasMap) *uint32 { return &c.Selector }, profiles)
	root.SetArgs([]string{"--section.labels.a", "flag"})
	require.NoError(t, root.Execute())
	assert.Equal(t, map[string]string{"a": "flag"}, c.Labels)
}

func TestTwoNestedProfilesScopeToTheirOwnSection(t *testing.T) {
	type profiledByString struct {
		Selector string
		Profiled string
	}
	type nestsTwoProfiled struct {
		Section profiledSection
		Other   profiledByString
	}

	root := newRoot(t)
	c := nestsTwoProfiled{Section: profiledSection{Selector: 1}, Other: profiledByString{Selector: "s"}}
	require.NoError(t, newBinder(t, testOptions).Register(root, &c,
		ProfileWithSelector(func(c *nestsTwoProfiled) *profiledSection { return &c.Section }, selector, sectionProfiles),
		ProfileWithSelector(func(c *nestsTwoProfiled) *profiledByString { return &c.Other },
			func(o *profiledByString) *string { return &o.Selector },
			map[string]func(*profiledByString){"s": func(o *profiledByString) { o.Profiled = "other" }}),
	))
	root.SetArgs([]string{"--section.selector", "137"})
	require.NoError(t, root.Execute())
	assert.Equal(t, nestsTwoProfiled{profiledSection{137, "137", 0}, profiledByString{"s", "other"}}, c)
}

func TestNestedProfileMayReachThroughAPointerSection(t *testing.T) {
	type nestsProfiledSectionPointer struct {
		Section *profiledSection
	}
	profiles := map[uint32]func(**profiledSection){1: func(s **profiledSection) { (*s).Profiled = "one" }}
	register := func(c *nestsProfiledSectionPointer) *cobra.Command {
		root := newRoot(t)
		require.NoError(t, newBinder(t, testOptions).Register(root, c, ProfileWithSelector(
			func(c *nestsProfiledSectionPointer) **profiledSection { return &c.Section },
			func(s **profiledSection) *uint32 { return &(*s).Selector }, profiles)))
		return root
	}

	var c nestsProfiledSectionPointer
	root := register(&c)
	root.SetArgs([]string{"--section.selector", "1"})
	require.NoError(t, root.Execute())
	assert.Equal(t, nestsProfiledSectionPointer{&profiledSection{Selector: 1, Profiled: "one"}}, c)

	var unset nestsProfiledSectionPointer
	require.NoError(t, register(&unset).Execute())
	assert.Nil(t, unset.Section)
}

func TestNestedProfileUnderANilPointerDoesNothing(t *testing.T) {
	type nestsProfiledSection struct {
		Section profiledSection
	}
	type nestsUnderAPointer struct {
		Outer *nestsProfiledSection
	}

	var c nestsUnderAPointer
	root := newRoot(t)
	require.NoError(t, newBinder(t, testOptions).Register(root, &c, ProfileWithSelector(
		func(c *nestsUnderAPointer) *profiledSection { return &c.Outer.Section }, selector, sectionProfiles)))
	require.NoError(t, root.Execute())
	assert.Nil(t, c.Outer)
}
