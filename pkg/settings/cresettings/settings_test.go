package cresettings

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
)

var update = flag.Bool("update", false, "update the golden files of this test")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

//go:generate go test . -update
var (
	//go:embed defaults.json
	defaultsJSON string
	//go:embed defaults.toml
	defaultsTOML string
	//go:embed defaults.yaml
	defaultsYAML string
)

func TestDefault(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		b, err := json.MarshalIndent(Default, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		if *update {
			require.NoError(t, os.WriteFile("defaults.json", b, 0644))
		} else {
			require.Equal(t, defaultsJSON, string(b))
		}
	})

	t.Run("toml", func(t *testing.T) {
		b, err := toml.Marshal(Default)
		if err != nil {
			log.Fatal(err)
		}
		if *update {
			require.NoError(t, os.WriteFile("defaults.toml", b, 0644))
		} else {
			require.Equal(t, defaultsTOML, string(b))
		}
	})

	t.Run("yaml", func(t *testing.T) {
		jb, err := json.MarshalIndent(Default, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		b, err := yaml.JSONToYAML(jb)
		if err != nil {
			log.Fatal(err)
		}
		if *update {
			require.NoError(t, os.WriteFile("defaults.yaml", b, 0644))
		} else {
			require.Equal(t, defaultsYAML, string(b))
		}
	})
}

func TestSchema_Unmarshal(t *testing.T) {
	cfg := Default
	require.NoError(t, json.Unmarshal([]byte(`{
	"WorkflowLimit": "500",
	"GatewayUnauthenticatedRequestRateLimit": "200rps:50",
	"GatewayUnauthenticatedRequestRateLimitPerIP": "1rps:100",
	"GatewayIncomingPayloadSizeLimit": "14kb",
    "GatewayVaultManagementEnabled": "true",
	"PerOrg": {
		"ZeroBalancePruningTimeout": "48h"
	},
	"PerOwner": {
		"WorkflowExecutionConcurrencyLimit": "99"
	},
	"PerWorkflow": {
		"WASMMemoryLimit": "250mb",
		"ChainAllowed": {
			"Default": "false",
			"Values": {
				"1": "true"
			}
		},
		"CRONTrigger": {
			"RateLimit": "every10s:5"
		},
		"HTTPTrigger": {
			"RateLimit": "every30s:3"
		},
		"LogTrigger": {
			"EventRateLimit": "every13s:6"
		},
		"HTTPAction": {
			"CallLimit": "5",
			"CacheAgeLimit": "5m"
		},
		"ConfidentialHTTP": {
			"CallLimit":         "5",
			"ConnectionTimeout": "10s",
			"RequestSizeLimit":  "10kb",
		    "ResponseSizeLimit": "100kb"
		},
		"Secrets": {
			"CallLimit": "5"
		},
		"ChainWrite": {
			"EVM": {
				"TransactionGasLimit": "500000"
			}
		},
		"ChainRead": {
			"CallLimit": "3"
		},
		"FeatureMultiTriggerExecutionIDsActiveAt": "2025-06-15 00:00:00 +0000 UTC"
	}
}`), &cfg))

	assert.Equal(t, 500, cfg.WorkflowLimit.DefaultValue)
	assert.Equal(t, 14*config.KByte, cfg.GatewayIncomingPayloadSizeLimit.DefaultValue)
	assert.Equal(t, true, cfg.GatewayVaultManagementEnabled.DefaultValue)
	assert.Equal(t, 48*time.Hour, cfg.PerOrg.ZeroBalancePruningTimeout.DefaultValue)
	assert.Equal(t, 99, cfg.PerOwner.WorkflowExecutionConcurrencyLimit.DefaultValue)
	assert.Equal(t, 250*config.MByte, cfg.PerWorkflow.WASMMemoryLimit.DefaultValue)
	assert.Equal(t, false, cfg.PerWorkflow.ChainAllowed.Default.DefaultValue)
	assert.Equal(t, "true", cfg.PerWorkflow.ChainAllowed.Values["1"])
	assert.NotNil(t, cfg.PerWorkflow.ChainAllowed.Default.Parse)
	assert.NotNil(t, cfg.PerWorkflow.ChainAllowed.KeyFromCtx)
	assert.Equal(t, config.Rate{Limit: rate.Every(30 * time.Second), Burst: 3}, cfg.PerWorkflow.HTTPTrigger.RateLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: rate.Every(13 * time.Second), Burst: 6}, cfg.PerWorkflow.LogTrigger.EventRateLimit.DefaultValue)
	assert.Equal(t, 5, cfg.PerWorkflow.HTTPAction.CallLimit.DefaultValue)
	assert.Equal(t, 5*time.Minute, cfg.PerWorkflow.HTTPAction.CacheAgeLimit.DefaultValue)
	assert.Equal(t, 5, cfg.PerWorkflow.ConfidentialHTTP.CallLimit.DefaultValue)
	assert.Equal(t, 10*config.KByte, cfg.PerWorkflow.ConfidentialHTTP.RequestSizeLimit.DefaultValue)
	assert.Equal(t, 5, cfg.PerWorkflow.Secrets.CallLimit.DefaultValue)
	assert.Equal(t, uint64(500000), cfg.PerWorkflow.ChainWrite.EVM.TransactionGasLimit.DefaultValue)
	assert.Equal(t, 3, cfg.PerWorkflow.ChainRead.CallLimit.DefaultValue)
	assert.Equal(t, time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), cfg.PerWorkflow.FeatureMultiTriggerExecutionIDsActiveAt.DefaultValue)
}

func TestDefaultGetter(t *testing.T) {
	limit := Default.PerWorkflow.HTTPAction.CallLimit

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "owner-id", Workflow: "foo"})
	overrideCtx := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "owner-id", Workflow: "test-wf-id"})

	// Default 5
	got, err := limit.GetOrDefault(ctx, DefaultGetter)
	require.NoError(t, err)
	require.Equal(t, 5, got)

	// No override
	got, err = limit.GetOrDefault(overrideCtx, DefaultGetter)
	require.NoError(t, err)
	require.Equal(t, 5, got)

	t.Cleanup(reinit) // restore default vars
	t.Setenv(EnvNameSettings, `{
	"workflow": {
		"test-wf-id": {
			"PerWorkflow": {
				"HTTPAction": {
					"CallLimit": "20"
				}
			}
		}
	}
}`)
	reinit() // set default vars

	// Default unchanged
	got, err = limit.GetOrDefault(ctx, DefaultGetter)
	require.NoError(t, err)
	require.Equal(t, 5, got)

	// Override applied
	got, err = limit.GetOrDefault(overrideCtx, DefaultGetter)
	require.NoError(t, err)
	require.Equal(t, 20, got)

}

func TestDefaultGetter_SettingMap(t *testing.T) {
	limit := Default.PerWorkflow.ChainAllowed

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "owner-id", Workflow: "foo"})
	ctx = contexts.WithChainSelector(ctx, 1234)
	overrideCtx := contexts.WithCRE(t.Context(), contexts.CRE{Owner: "owner-id", Workflow: "test-wf-id"})
	overrideCtx = contexts.WithChainSelector(overrideCtx, 1234)

	// None allowed by default
	got, err := limit.GetOrDefault(ctx, DefaultGetter)
	require.NoError(t, err)
	require.False(t, got)
	got, err = limit.GetOrDefault(overrideCtx, DefaultGetter)
	require.NoError(t, err)
	require.False(t, got)

	t.Cleanup(reinit) // restore default vars

	// Org override to allow
	t.Setenv(EnvNameSettings, `{
	"workflow": {
		"test-wf-id": {
			"PerWorkflow": {
				"ChainAllowed": {
					"Values": {
						"1234": "true"
					}
				}
			}
		}
	}
}`)
	reinit() // set default vars

	// ensure merged values; defaults must remain
	require.Equal(t, "true", Default.PerWorkflow.ChainAllowed.Values["3379446385462418246"])
	// confirm
	got, err = limit.GetOrDefault(ctx, DefaultGetter)
	require.NoError(t, err)
	require.False(t, got)
	got, err = limit.GetOrDefault(overrideCtx, DefaultGetter)
	require.NoError(t, err)
	require.True(t, got)

	// Org override to allow by default, but disallow some
	t.Setenv(EnvNameSettings, `{
	"workflow": {
		"test-wf-id": {
			"PerWorkflow": {
				"ChainAllowed": {
					"Default": true,
					"Values": {
						"1234": "false"
					}
				}
			}
		}
	}
}`)
	reinit() // set default vars
	got, err = limit.GetOrDefault(ctx, DefaultGetter)
	require.NoError(t, err)
	require.False(t, got)
	got, err = limit.GetOrDefault(overrideCtx, DefaultGetter)
	require.NoError(t, err)
	require.False(t, got)
	got, err = limit.GetOrDefault(contexts.WithChainSelector(overrideCtx, 42), DefaultGetter)
	require.NoError(t, err)
	require.True(t, got)
}

func TestDefaultEnvVars(t *testing.T) {
	// confirm defaults
	require.Equal(t, "", Default.PerWorkflow.ChainAllowed.Values["1234"])
	require.Equal(t, "true", Default.PerWorkflow.ChainAllowed.Values["3379446385462418246"])

	t.Cleanup(reinit) // restore after

	// update defaults
	t.Setenv(EnvNameSettingsDefault, `{
	"PerWorkflow": {
		"ChainAllowed": {
			"Values": {
				"1234": "true"
			}
		}
	}
}`)
	reinit() // set default vars

	// confirm through Default
	require.Equal(t, "true", Default.PerWorkflow.ChainAllowed.Values["1234"])
	// without affecting others (they must merge)
	require.Equal(t, "true", Default.PerWorkflow.ChainAllowed.Values["3379446385462418246"])

	// confirm through DefaultGetter
	gl, err := limits.MakeGateLimiter(limits.Factory{Logger: logger.Test(t), Settings: DefaultGetter}, Default.PerWorkflow.ChainAllowed)
	require.NoError(t, err)

	ctx := contexts.WithCRE(t.Context(), contexts.CRE{Org: "foo", Owner: "owner-id", Workflow: "foo"})
	// defaults and global override allowed
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 3379446385462418246)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 12922642891491394802)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 1234)))

	// update overrides
	t.Setenv(EnvNameSettingsDefault, "{}")
	t.Setenv(EnvNameSettings, `{
	"global": {
		"PerWorkflow": {
			"ChainAllowed": {
				"Values": {
					"1234": "true"
				}
			}
		}
	}
}`)

	reinit() // set default vars

	// confirm through DefaultGetter
	gl, err = limits.MakeGateLimiter(limits.Factory{Logger: logger.Test(t), Settings: DefaultGetter}, Default.PerWorkflow.ChainAllowed)
	require.NoError(t, err)

	// defaults and global override allowed
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 3379446385462418246)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 12922642891491394802)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 1234)))

	// confirm through an empty, but non-nil getter
	getter, err := settings.NewJSONGetter([]byte(`{}`))
	require.NoError(t, err)
	gl, err = limits.MakeGateLimiter(limits.Factory{Logger: logger.Test(t), Settings: getter}, Default.PerWorkflow.ChainAllowed)
	require.NoError(t, err)

	// defaults and global override allowed
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 3379446385462418246)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 12922642891491394802)))
	assert.NoError(t, gl.AllowErr(contexts.WithChainSelector(ctx, 1234)))
}

//go:embed README.md
var readme string

// TestFlowchartComplete ensures that every field is included in the flowchart.
func TestFlowchartComplete(t *testing.T) {
	var keys []string
	var addKeys func(a any)
	addKeys = func(a any) {
		if v := reflect.ValueOf(a).Elem(); v.Type().Kind() == reflect.Struct {
			for i := range v.NumField() {
				f := v.Field(i)
				if gk, ok := f.Addr().Interface().(interface{ GetKey() string }); ok {
					keys = append(keys, gk.GetKey())
					continue
				}
				addKeys(f.Addr().Interface())
			}
		}
	}
	addKeys(&Default)
	require.NotEmpty(t, keys)
	require.Greater(t, len(keys), 20) // sanity check
	for _, k := range keys {
		assert.Contains(t, readme, k, "missing key %q in README.md", k)
	}
}
