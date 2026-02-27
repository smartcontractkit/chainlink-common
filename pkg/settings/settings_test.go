package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

func TestTime(t *testing.T) {
	s := Time(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))

	t.Run("parse RFC3339", func(t *testing.T) {
		got, err := s.Parse("2025-06-15T12:30:00Z")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2025, 6, 15, 12, 30, 0, 0, time.UTC), got)
	})

	t.Run("parse RFC3339 with nanoseconds", func(t *testing.T) {
		got, err := s.Parse("2025-06-15T12:30:00.123456789Z")
		require.NoError(t, err)
		assert.Equal(t, time.Date(2025, 6, 15, 12, 30, 0, 123456789, time.UTC), got)
	})

	t.Run("MarshalText", func(t *testing.T) {
		b, err := s.MarshalText()
		require.NoError(t, err)
		assert.Equal(t, "2100-01-01 00:00:00 +0000 UTC", string(b))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		var s2 Setting[time.Time]
		s2.Parse = s.Parse
		require.NoError(t, s2.UnmarshalText([]byte("2100-01-01T00:00:00Z")))
		assert.Equal(t, s.DefaultValue, s2.DefaultValue)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := s.Parse("not-a-date")
		assert.Error(t, err)
	})
}

func TestInitConfig(t *testing.T) {
	require.Error(t, InitConfig(&struct{ Field int }{Field: 10})) // fields must be type Setting

	type org struct {
		OrgField Setting[int]
	}
	type owner struct {
		OwnerField Setting[int64]
	}
	type workflow struct {
		WorkflowField Setting[float64]
	}
	var c = struct {
		GlobalField Setting[config.Rate]
		PerOrg      org      `scope:"org"`
		PerOwner    owner    `scope:"owner"`
		PerWorkflow workflow `scope:"workflow"`
	}{
		GlobalField: Rate(rate.Every(time.Second), 5),
		PerOrg:      org{OrgField: Int(42)},
		PerOwner:    owner{OwnerField: Int64(13)},
		PerWorkflow: workflow{WorkflowField: Float64(1.5)},
	}
	require.NoError(t, InitConfig(&c))

	assert.NotEmpty(t, c.GlobalField.Key)
	assert.Equal(t, ScopeGlobal, c.GlobalField.Scope)
	assert.NotNil(t, c.GlobalField.Parse)

	assert.NotEmpty(t, c.PerOrg.OrgField.Key)
	assert.Equal(t, ScopeOrg, c.PerOrg.OrgField.Scope)
	assert.NotNil(t, c.PerOrg.OrgField.Parse)

	assert.NotEmpty(t, c.PerOwner.OwnerField.Key)
	assert.Equal(t, ScopeOwner, c.PerOwner.OwnerField.Scope)
	assert.NotNil(t, c.PerOwner.OwnerField.Parse)

	assert.NotEmpty(t, c.PerWorkflow.WorkflowField.Key)
	assert.Equal(t, ScopeWorkflow, c.PerWorkflow.WorkflowField.Scope)
	assert.NotNil(t, c.PerWorkflow.WorkflowField.Parse)
}
