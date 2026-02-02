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

const (
	EnvNameSettings        = "CL_CRE_SETTINGS"
	EnvNameSettingsDefault = "CL_CRE_SETTINGS_DEFAULT"
)

func init() { reinit() }
func reinit() {
	if v, ok := os.LookupEnv(EnvNameSettingsDefault); ok {
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

	if v, ok := os.LookupEnv(EnvNameSettings); ok {
		DefaultGetter, err = NewJSONGetter([]byte(v))
		if err != nil {
			log.Fatalf("failed to initialize settings: %v", err)
		}
	} else {
		DefaultGetter = nil
	}
}

// DefaultGetter is a default settings getter populated from the env var CL_CRE_SETTINGS if set, otherwise it is nil.
var DefaultGetter Getter

// Deprecated: use Default
var Config Schema

var Default = Schema{
	WorkflowLimit:                     Int(200),
	WorkflowExecutionConcurrencyLimit: Int(200),
	GatewayIncomingPayloadSizeLimit:   Size(1 * config.MByte),
	GatewayVaultManagementEnabled:     Bool(true),

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
		ZeroBalancePruningTimeout: Duration(24 * time.Hour),
	},
	PerOwner: Owners{
		WorkflowExecutionConcurrencyLimit: Int(5),

		// DANGER(cedric): Be extremely careful changing this vault limit as it acts as a default value
		// used by the Vault OCR plugin -- changing this value could cause issues with the plugin during an image
		// upgrade as nodes apply the old and new values inconsistently. A safe upgrade path
		// must ensure that we are overriding the default in the onchain configuration for the contract.
		VaultSecretsLimit: Int(100),
	},
	PerWorkflow: Workflows{
		TriggerRegistrationsTimeout:   Duration(10 * time.Second),
		TriggerEventQueueLimit:        Int(50),
		TriggerEventQueueTimeout:      Duration(10 * time.Minute),
		TriggerSubscriptionTimeout:    Duration(15 * time.Second),
		TriggerSubscriptionLimit:      Int(10),
		CapabilityConcurrencyLimit:    Int(30), // we should rely on per-capability execution limits instead of concurrency limit
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
		LogLineLimit:                  Size(config.KByte),
		LogEventLimit:                 Int(1_000),
		ChainAllowed: PerChainSelector(Bool(false), map[string]bool{
			// geth-testnet
			"3379446385462418246": true,
			// geth-devnet2
			"12922642891491394802": true,
		}),

		CRONTrigger: cronTrigger{
			FastestScheduleInterval: Duration(30 * time.Second),
		},
		HTTPTrigger: httpTrigger{
			RateLimit: Rate(rate.Every(30*time.Second), 3),
		},
		LogTrigger: logTrigger{
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
				GasLimit: PerChainSelector(Uint64(5_000_000), map[string]uint64{
					// geth-testnet
					"3379446385462418246": 10_000_000,
					// geth-devnet2
					"12922642891491394802": 50_000_000,
				}),
			},
		},
		ChainRead: chainRead{
			CallLimit:          Int(15),
			LogQueryBlockLimit: Uint64(100),
			PayloadSizeLimit:   Size(5 * config.KByte),
		},
		Consensus: consensus{
			ObservationSizeLimit: Size(100 * config.KByte),
			CallLimit:            Int(20),
		},
		HTTPAction: httpAction{
			CallLimit:         Int(5),
			CacheAgeLimit:     Duration(10 * time.Minute),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(10 * config.KByte),
			ResponseSizeLimit: Size(100 * config.KByte),
		},
		Secrets: secrets{
			CallLimit: Int(5),
		},
	},
}

type Schema struct {
	WorkflowLimit                     Setting[int] `unit:"{workflow}"` // Deprecated
	WorkflowExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	GatewayIncomingPayloadSizeLimit   Setting[config.Size]
	GatewayVaultManagementEnabled     Setting[bool]

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
	ZeroBalancePruningTimeout Setting[time.Duration]
}

type Owners struct {
	WorkflowExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	VaultSecretsLimit                 Setting[int] `unit:"{secret}"`
}

type Workflows struct {
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

	LogLineLimit  Setting[config.Size]
	LogEventLimit Setting[int] `unit:"{log}"`

	ChainAllowed SettingMap[bool]

	CRONTrigger cronTrigger
	HTTPTrigger httpTrigger
	LogTrigger  logTrigger

	ChainWrite chainWrite
	ChainRead  chainRead
	Consensus  consensus
	HTTPAction httpAction
	Secrets    secrets
}

type cronTrigger struct {
	FastestScheduleInterval Setting[time.Duration]
}
type httpTrigger struct {
	RateLimit Setting[config.Rate]
}
type logTrigger struct {
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
	TransactionGasLimit Setting[uint64]    `unit:"{gas}"` // Deprecated
	GasLimit            SettingMap[uint64] `unit:"{gas}"`
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
type secrets struct {
	CallLimit Setting[int] `unit:"{call}"`
}
type consensus struct {
	ObservationSizeLimit Setting[config.Size]
	CallLimit            Setting[int] `unit:"{call}"`
}
