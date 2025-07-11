package settings

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

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
