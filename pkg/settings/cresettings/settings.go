// Package cresettings contains configurable settings definitions for nodes in the CRE.
// Environment Variables:
//   - CL_CRE_SETTINGS_DEFAULT: defaults like in ./defaults.json - initializes Default
//   - CL_CRE_SETTINGS: scoped settings like in ../settings/testdata/config.json - initializes DefaultGetter
package cresettings

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	. "github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func init() {
	if v, ok := os.LookupEnv("CL_CRE_SETTINGS_DEFAULT"); ok {
		err := json.Unmarshal([]byte(v), &Default)
		if err != nil {
			log.Fatalf("failed to initialize defaults: %v", err)
		}
	}
	err := InitConfig(&Default)
	if err != nil {
		log.Fatalf("failed to initialize keys: %v", err)
	}
	Config = Default

	if v, ok := os.LookupEnv("CL_CRE_SETTINGS"); ok {
		DefaultGetter, err = NewJSONGetter([]byte(v))
		if err != nil {
			log.Fatalf("failed to initialize settings: %v", err)
		}
	}
}

// DefaultGetter is a default settings getter populated from the env var CL_CRE_SETTINGS if set, otherwise it is nil.
var DefaultGetter Getter

// Deprecated: use Default
var Config Schema

var Default = Schema{
	WorkflowLimit:                     Int(200),
	WorkflowExecutionConcurrencyLimit: Int(200),
	WorkflowTriggerRateLimit:          Rate(200, 200),
	GatewayIncomingPayloadSizeLimit:   Size(1 * config.MByte),

	// DANGER(cedric): Be extremely careful changing these vault limits as they act as a default value
	// used by the Vault OCR plugin -- changing these values could cause issues with the plugin during an image
	// upgrade as nodes apply the old and new values inconsistently. A safe upgrade path
	// must ensure that we are overriding the default in the onchain configuration for the contract.
	VaultCiphertextSizeLimit:          Size(2 * config.KByte),
	VaultIdentifierKeySizeLimit:       Size(64 * config.Byte),
	VaultIdentifierOwnerSizeLimit:     Size(64 * config.Byte),
	VaultIdentifierNamespaceSizeLimit: Size(64 * config.Byte),
	VaultPluginBatchSizeLimit:         Int(20),
	VaultRequestBatchSizeLimit:        Int(10),

	PerOrg: Orgs{
		WorkflowDeploymentRateLimit: Rate(rate.Every(time.Minute), 1),
		ZeroBalancePruningTimeout:   Duration(24 * time.Hour),
	},
	PerOwner: Owners{
		WorkflowExecutionConcurrencyLimit: Int(5),
		WorkflowTriggerRateLimit:          Rate(5, 5),

		// DANGER(cedric): Be extremely careful changing this vault limit as it acts as a default value
		// used by the Vault OCR plugin -- changing this value could cause issues with the plugin during an image
		// upgrade as nodes apply the old and new values inconsistently. A safe upgrade path
		// must ensure that we are overriding the default in the onchain configuration for the contract.
		VaultSecretsLimit: Int(100),
	},
	PerWorkflow: Workflows{
		TriggerRateLimit:              Rate(rate.Every(30*time.Second), 3),
		TriggerRegistrationsTimeout:   Duration(10 * time.Second),
		TriggerEventQueueLimit:        Int(1_000),
		TriggerEventQueueTimeout:      Duration(10 * time.Minute),
		TriggerSubscriptionTimeout:    Duration(15 * time.Second),
		TriggerSubscriptionLimit:      Int(10),
		CapabilityConcurrencyLimit:    Int(3),
		CapabilityCallTimeout:         Duration(3 * time.Minute),
		SecretsConcurrencyLimit:       Int(5),
		ExecutionConcurrencyLimit:     Int(5),
		ExecutionTimeout:              Duration(5 * time.Minute),
		ExecutionResponseLimit:        Size(100 * config.KByte),
		WASMMemoryLimit:               Size(100 * config.MByte),
		WASMBinarySizeLimit:           Size(100 * config.MByte),
		WASMCompressedBinarySizeLimit: Size(20 * config.MByte),
		WASMConfigSizeLimit:           Size(config.MByte),
		WASMSecretsSizeLimit:          Size(config.MByte),
		WASMResponseSizeLimit:         Size(100 * config.KByte),
		ConsensusObservationSizeLimit: Size(100 * config.KByte),
		ConsensusCallsLimit:           Int(2000),
		LogLineLimit:                  Size(config.KByte),
		LogEventLimit:                 Int(1_000),

		CRONTrigger: cronTrigger{
			RateLimit: Rate(rate.Every(30*time.Second), 1),
		},
		HTTPTrigger: httpTrigger{
			RateLimit: Rate(rate.Every(30*time.Second), 3),
		},
		LogTrigger: logTrigger{
			Limit:                    Int(5),
			EventRateLimit:           Rate(rate.Every(time.Minute/10), 10),
			FilterAddressLimit:       Int(5),
			FilterTopicsPerSlotLimit: Int(10),
			EventSizeLimit:           Size(5 * config.KByte),
		},

		ChainWrite: chainWrite{
			TargetsLimit:    Int(10),
			ReportSizeLimit: Size(5 * config.KByte),
			EVM: evmChainWrite{
				TransactionGasLimit: Uint64(5_000_000),
			},
		},
		ChainRead: chainRead{
			CallLimit:          Int(10),
			LogQueryBlockLimit: Uint64(100),
			PayloadSizeLimit:   Size(5 * config.KByte),
		},
		Consensus: consensus{
			ObservationSizeLimit: Size(100 * config.KByte),
			CallLimit:            Int(2000),
		},
		HTTPAction: httpAction{
			CallLimit:         Int(5),
			CacheAgeLimit:     Duration(10 * time.Minute),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(10 * config.KByte),
			ResponseSizeLimit: Size(100 * config.KByte),
		},
	},
}

type Schema struct {
	WorkflowLimit                     Setting[int] `unit:"{workflow}"`
	WorkflowExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	// Deprecated
	WorkflowTriggerRateLimit        Setting[config.Rate]
	GatewayIncomingPayloadSizeLimit Setting[config.Size]

	VaultCiphertextSizeLimit          Setting[config.Size]
	VaultIdentifierKeySizeLimit       Setting[config.Size]
	VaultIdentifierOwnerSizeLimit     Setting[config.Size]
	VaultIdentifierNamespaceSizeLimit Setting[config.Size]
	VaultPluginBatchSizeLimit         Setting[int] `unit:"{request}"`
	VaultRequestBatchSizeLimit        Setting[int] `unit:"{request}"`

	PerOrg      Orgs      `scope:"org"`
	PerOwner    Owners    `scope:"owner"`
	PerWorkflow Workflows `scope:"workflow"`
}
type Orgs struct {
	// Deprecated
	WorkflowDeploymentRateLimit Setting[config.Rate]
	ZeroBalancePruningTimeout   Setting[time.Duration]
}

type Owners struct {
	WorkflowExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	// Deprecated
	WorkflowTriggerRateLimit Setting[config.Rate]
	VaultSecretsLimit        Setting[int] `unit:"{secret}"`
}

type Workflows struct {
	// Deprecated
	TriggerRateLimit            Setting[config.Rate]
	TriggerRegistrationsTimeout Setting[time.Duration]
	TriggerSubscriptionTimeout  Setting[time.Duration]
	TriggerSubscriptionLimit    Setting[int] `unit:"{subscription}"`
	TriggerEventQueueLimit      Setting[int] `unit:"{trigger}"`
	TriggerEventQueueTimeout    Setting[time.Duration]

	CapabilityConcurrencyLimit Setting[int] `unit:"{capability}"`
	CapabilityCallTimeout      Setting[time.Duration]

	SecretsConcurrencyLimit Setting[int] `unit:"{secret}"`

	ExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	ExecutionTimeout          Setting[time.Duration]
	ExecutionResponseLimit    Setting[config.Size]

	WASMMemoryLimit               Setting[config.Size]
	WASMBinarySizeLimit           Setting[config.Size]
	WASMCompressedBinarySizeLimit Setting[config.Size]
	WASMConfigSizeLimit           Setting[config.Size]
	WASMSecretsSizeLimit          Setting[config.Size]
	// Deprecated: use ExecutionResponseLimit
	WASMResponseSizeLimit Setting[config.Size]

	// Deprecated: use Consensus.ObservationSizeLimit
	ConsensusObservationSizeLimit Setting[config.Size]
	// Deprecated: use Consensus.CallLimit
	ConsensusCallsLimit Setting[int] `unit:"{call}"`

	LogLineLimit  Setting[config.Size]
	LogEventLimit Setting[int] `unit:"{log}"`

	CRONTrigger cronTrigger
	HTTPTrigger httpTrigger
	LogTrigger  logTrigger

	ChainWrite chainWrite
	ChainRead  chainRead
	Consensus  consensus
	HTTPAction httpAction
}

type cronTrigger struct {
	// Deprecated: to be removed
	RateLimit Setting[config.Rate]
}
type httpTrigger struct {
	RateLimit Setting[config.Rate]
}
type logTrigger struct {
	// Deprecated
	Limit                    Setting[int] `unit:"{trigger}"`
	EventRateLimit           Setting[config.Rate]
	EventSizeLimit           Setting[config.Size]
	FilterAddressLimit       Setting[int] `unit:"{address}"`
	FilterTopicsPerSlotLimit Setting[int] `unit:"{topic}"`
}
type chainWrite struct {
	TargetsLimit    Setting[int] `unit:"{target}"`
	ReportSizeLimit Setting[config.Size]

	EVM evmChainWrite
}
type evmChainWrite struct {
	TransactionGasLimit Setting[uint64] `unit:"{gas}"`
}
type chainRead struct {
	CallLimit          Setting[int]    `unit:"{call}"`
	LogQueryBlockLimit Setting[uint64] `unit:"{block}"`
	PayloadSizeLimit   Setting[config.Size]
}
type httpAction struct {
	CallLimit         Setting[int] `unit:"{call}"`
	CacheAgeLimit     Setting[time.Duration]
	ConnectionTimeout Setting[time.Duration]
	RequestSizeLimit  Setting[config.Size]
	ResponseSizeLimit Setting[config.Size]
}
type consensus struct {
	ObservationSizeLimit Setting[config.Size]
	CallLimit            Setting[int] `unit:"{call}"`
}
