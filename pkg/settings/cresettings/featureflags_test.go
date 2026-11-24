package cresettings

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// notFeatureFlags lists every bool or *ActivePeriod setting that is deliberately not a feature flag.
// A new such setting must either be tagged `featureflag:"true"` in the Schema or be added here with a reason.
var notFeatureFlags = map[string]string{
	"CoordinatedEngineEnabled":                                   "deprecated in favor of PerWorkflow.CoordinatedEngineEnabled",
	"PerOrg.BaseTriggerRetransmitEnabled":                        "per-org tuning switch",
	"PerWorkflow.UserMetricEnabled":                              "per-workflow capability toggle",
	"PerWorkflow.ChainAllowed":                                   "per-chain allowlist",
	"PerWorkflow.ConfidentialWorkflows.Enabled":                  "capability toggle",
	"ConfidentialCompute.InsecureSkipTLSVerify":                  "dev/test security override",
	"ConfidentialCompute.SecretsCacheEnabled":                    "confidential compute tuning toggle",
	"ConfidentialCompute.PublicKeyCache.Enabled":                 "confidential compute tuning toggle",
	"ConfidentialCompute.PublicKeyCache.ProactiveRefreshEnabled": "confidential compute tuning toggle",
	"ConfidentialCompute.Session.PersistenceEnabled":             "confidential compute tuning toggle",
}

func init() {
	for _, c := range []string{"HTTP", "Workflows"} {
		p := "ConfidentialCompute." + c + "."
		for _, k := range []string{"SecretsCacheEnabled", "PublicKeyCache.Enabled", "PublicKeyCache.ProactiveRefreshEnabled", "Session.PersistenceEnabled"} {
			notFeatureFlags[p+k] = "confidential compute tuning toggle"
		}
	}
}

func TestFeatureFlags_Complete(t *testing.T) {
	infos, err := settings.Infos(&Default)
	require.NoError(t, err)

	timeRange := reflect.TypeFor[settings.Range[config.Timestamp]]()
	seen := map[string]bool{}
	for _, i := range infos {
		seen[i.Key] = true
		candidate := i.ValueType.Kind() == reflect.Bool ||
			(i.ValueType == timeRange && strings.HasSuffix(i.Key, "ActivePeriod"))
		if !candidate {
			require.False(t, i.FeatureFlag, "%s is marked as a feature flag but is not a bool or ActivePeriod", i.Key)
			continue
		}
		reason, allowed := notFeatureFlags[i.Key]
		switch {
		case i.FeatureFlag && allowed:
			t.Errorf("%s is both marked as a feature flag and allowlisted; remove one", i.Key)
		case !i.FeatureFlag && !allowed:
			t.Errorf("%s is a bool/ActivePeriod setting: tag it `featureflag:\"true\"` in Schema, or add it to notFeatureFlags with a reason", i.Key)
		case allowed && reason == "":
			t.Errorf("%s: notFeatureFlags entry needs a reason", i.Key)
		}
	}
	for k := range notFeatureFlags {
		if !seen[k] {
			t.Errorf("notFeatureFlags has stale entry %s", k)
		}
	}
}

func TestFeatureFlagKeys(t *testing.T) {
	flags := FeatureFlagKeys()
	require.NotEmpty(t, flags)
	byKey := map[string]settings.Scope{}
	for _, f := range flags {
		byKey[f.Key] = f.Scope
	}
	require.Equal(t, settings.ScopeGlobal, byKey["DonTimeSequencedTimestampsActivePeriod"])
	require.Equal(t, settings.ScopeGlobal, byKey["CRONV2ActivePeriod"])
	require.Equal(t, settings.ScopeWorkflow, byKey["PerWorkflow.FeatureShadowTriggerModeActivePeriod"])
	require.Equal(t, settings.ScopeWorkflow, byKey["PerWorkflow.FeatureTriggerConsensusEnabledActivePeriod"])
	require.Contains(t, byKey, "ConfidentialCompute.ConfidentialHTTPVaultGetSecretsDirectlyActivePeriod")
}
