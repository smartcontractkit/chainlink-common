package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

func TestFeatureFlags_InitSetting(t *testing.T) {
	type testSchema struct {
		FeatureFlags FeatureFlags
		PerWorkflow  struct {
			FeatureFlags FeatureFlags
		} `scope:"workflow"`
	}
	var s testSchema
	require.NoError(t, InitConfig(&s))

	assert.Equal(t, "FeatureFlags", s.FeatureFlags.GetKey())
	assert.Equal(t, ScopeGlobal, s.FeatureFlags.GetScope())

	assert.Equal(t, "PerWorkflow.FeatureFlags", s.PerWorkflow.FeatureFlags.GetKey())
	assert.Equal(t, ScopeWorkflow, s.PerWorkflow.FeatureFlags.GetScope())
}

func TestFeatureFlags_GetFlag_DefaultsOnly(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000, Metadata: map[string]string{"k": "v"}},
			{Name: "feat_b", ActivateAt: 2000},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeGlobal, nil)

	ctx := t.Context()
	flag, err := ff.GetFlag(ctx, "feat_a")
	require.NoError(t, err)
	require.NotNil(t, flag)
	assert.Equal(t, int64(1000), flag.ActivateAt)
	assert.Equal(t, "v", flag.Metadata["k"])

	flag, err = ff.GetFlag(ctx, "missing")
	require.NoError(t, err)
	assert.Nil(t, flag)
}

func TestFeatureFlags_GetFlag_GetterOverride(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000, Metadata: map[string]string{"k": "default"}},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeWorkflow, nil)

	configJSON := `{
		"workflow": {
			"wf1": {
				"FeatureFlags": {
					"feat_a": { "ActivateAt": 5000 }
				}
			}
		}
	}`
	g, err := NewJSONGetter([]byte(configJSON))
	require.NoError(t, err)

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{
		Owner:    "owner1",
		Workflow: "wf1",
	})
	ffWithGetter := ff.With(g)
	flag, err := ffWithGetter.GetFlag(ctx, "feat_a")
	require.NoError(t, err)
	require.NotNil(t, flag)
	assert.Equal(t, int64(5000), flag.ActivateAt)
	assert.Equal(t, "default", flag.Metadata["k"])
}

func TestFeatureFlags_GetFlag_ScopePrecedence(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeWorkflow, nil)

	configJSON := `{
		"global": {
			"FeatureFlags": { "feat_a": { "ActivateAt": 2000 } }
		},
		"owner": {
			"owner1": {
				"FeatureFlags": { "feat_a": { "ActivateAt": 3000 } }
			}
		},
		"workflow": {
			"wf1": {
				"FeatureFlags": { "feat_a": { "ActivateAt": 4000 } }
			}
		}
	}`
	g, err := NewJSONGetter([]byte(configJSON))
	require.NoError(t, err)

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{
		Owner:    "owner1",
		Workflow: "wf1",
	})

	ffWithGetter := ff.With(g)
	flag, err := ffWithGetter.GetFlag(ctx, "feat_a")
	require.NoError(t, err)
	require.NotNil(t, flag)
	assert.Equal(t, int64(4000), flag.ActivateAt, "workflow scope should win")

	// Remove workflow override, owner should win
	configJSON2 := `{
		"global": {
			"FeatureFlags": { "feat_a": { "ActivateAt": 2000 } }
		},
		"owner": {
			"owner1": {
				"FeatureFlags": { "feat_a": { "ActivateAt": 3000 } }
			}
		}
	}`
	g2, err := NewJSONGetter([]byte(configJSON2))
	require.NoError(t, err)

	ffWithGetter2 := ff.With(g2)
	flag, err = ffWithGetter2.GetFlag(ctx, "feat_a")
	require.NoError(t, err)
	require.NotNil(t, flag)
	assert.Equal(t, int64(3000), flag.ActivateAt, "owner scope should win when no workflow override")
}

func TestFeatureFlags_IsActive(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeGlobal, nil)

	ctx := t.Context()

	active, err := ff.IsActive(ctx, "feat_a", 999)
	require.NoError(t, err)
	assert.False(t, active)

	active, err = ff.IsActive(ctx, "feat_a", 1000)
	require.NoError(t, err)
	assert.True(t, active)

	active, err = ff.IsActive(ctx, "feat_a", 2000)
	require.NoError(t, err)
	assert.True(t, active)

	active, err = ff.IsActive(ctx, "missing", 9999)
	require.NoError(t, err)
	assert.False(t, active)
}

func TestFeatureFlags_GetMetadata(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000, Metadata: map[string]string{"k1": "default_v1"}},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeWorkflow, nil)

	configJSON := `{
		"workflow": {
			"wf1": {
				"FeatureFlags": {
					"feat_a": {
						"ActivateAt": 5000,
						"Metadata": { "k1": "override_v1", "k2": "new_v2" }
					}
				}
			}
		}
	}`
	g, err := NewJSONGetter([]byte(configJSON))
	require.NoError(t, err)

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{
		Owner:    "owner1",
		Workflow: "wf1",
	})

	ffWithGetter := ff.With(g)

	val, err := ffWithGetter.GetMetadata(ctx, "feat_a", "k1")
	require.NoError(t, err)
	assert.Equal(t, "override_v1", val, "getter should override default metadata")

	val, err = ffWithGetter.GetMetadata(ctx, "feat_a", "k2")
	require.NoError(t, err)
	assert.Equal(t, "new_v2", val, "getter should provide new metadata keys")

	// Without getter, fall back to defaults
	val, err = ff.GetMetadata(ctx, "feat_a", "k1")
	require.NoError(t, err)
	assert.Equal(t, "default_v1", val)

	val, err = ff.GetMetadata(ctx, "feat_a", "k2")
	require.NoError(t, err)
	assert.Equal(t, "", val, "missing metadata key returns empty string")
}

func TestFeatureFlags_GetFlag_ReturnsCopy(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000, Metadata: map[string]string{"k": "v"}},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeGlobal, nil)

	flag, err := ff.GetFlag(t.Context(), "feat_a")
	require.NoError(t, err)
	flag.ActivateAt = 9999

	original := ff.getDefault("feat_a")
	assert.Equal(t, int64(1000), original.ActivateAt, "mutation should not affect stored flag")
}

func TestFeatureFlags_With_DoesNotMutateOriginal(t *testing.T) {
	ff := FeatureFlags{
		Flags: []FeatureFlag{
			{Name: "feat_a", ActivateAt: 1000},
		},
	}
	_ = ff.initSetting("FeatureFlags", ScopeGlobal, nil)

	g, err := NewJSONGetter([]byte(`{}`))
	require.NoError(t, err)

	bound := ff.With(g)
	assert.Nil(t, ff.getter, "original should not have getter set")
	assert.NotNil(t, bound.getter, "bound copy should have getter set")
}
