package cresettings

import (
	_ "embed"
	"encoding/json"
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
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
}

func TestSchema_Unmarshal(t *testing.T) {
	cfg := Default
	require.NoError(t, json.Unmarshal([]byte(`{
	"WorkflowLimit": "500",
	"GatewayUnauthenticatedRequestRateLimit": "200rps:50",
	"GatewayUnauthenticatedRequestRateLimitPerIP": "1rps:100",
	"GatewayIncomingPayloadSizeLimit": "14kb",
	"PerOrg": {
		"ZeroBalancePruningTimeout": "48h"
	},
	"PerOwner": {
		"WorkflowExecutionConcurrencyLimit": "99"
	},
	"PerWorkflow": {
		"WASMMemoryLimit": "250mb",
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
			"RateLimit": "every3s:5",
			"CacheAgeLimit": "5m"
		},
		"ChainWrite": {
			"EVM": {
				"TransactionGasLimit": "500000"
			}
		},
		"ChainRead": {
			"CallLimit": "3"
		}
	}
}`), &cfg))

	assert.Equal(t, 500, cfg.WorkflowLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: 200, Burst: 50}, cfg.GatewayUnauthenticatedRequestRateLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: 1, Burst: 100}, cfg.GatewayUnauthenticatedRequestRateLimitPerIP.DefaultValue)
	assert.Equal(t, 14*config.KByte, cfg.GatewayIncomingPayloadSizeLimit.DefaultValue)
	assert.Equal(t, 48*time.Hour, cfg.PerOrg.ZeroBalancePruningTimeout.DefaultValue)
	assert.Equal(t, 99, cfg.PerOwner.WorkflowExecutionConcurrencyLimit.DefaultValue)
	assert.Equal(t, 250*config.MByte, cfg.PerWorkflow.WASMMemoryLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: rate.Every(10 * time.Second), Burst: 5}, cfg.PerWorkflow.CRONTrigger.RateLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: rate.Every(30 * time.Second), Burst: 3}, cfg.PerWorkflow.HTTPTrigger.RateLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: rate.Every(13 * time.Second), Burst: 6}, cfg.PerWorkflow.LogTrigger.EventRateLimit.DefaultValue)
	assert.Equal(t, config.Rate{Limit: rate.Every(3 * time.Second), Burst: 5}, cfg.PerWorkflow.HTTPAction.RateLimit.DefaultValue)
	assert.Equal(t, 5*time.Minute, cfg.PerWorkflow.HTTPAction.CacheAgeLimit.DefaultValue)
	assert.Equal(t, uint64(500000), cfg.PerWorkflow.ChainWrite.EVM.TransactionGasLimit.DefaultValue)
	assert.Equal(t, 3, cfg.PerWorkflow.ChainRead.CallLimit.DefaultValue)
}
