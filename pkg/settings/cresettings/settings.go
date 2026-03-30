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
	WorkflowLimit:                          Int(200),
	WorkflowExecutionConcurrencyLimit:      Int(200),
	GatewayIncomingPayloadSizeLimit:        Size(1 * config.MByte),
	GatewayVaultManagementEnabled:          Bool(true),
	VaultJWTAuthEnabled:                    Bool(false),
	VaultOrgIdAsSecretOwnerEnabled:         Bool(false),
	GatewayHTTPGlobalRate:                  Rate(rate.Limit(500), 500),
	GatewayHTTPPerNodeRate:                 Rate(rate.Limit(100), 100),
	TriggerRegistrationStatusUpdateTimeout: Duration(0 * time.Second),
	// DANGER(cedric): Be extremely careful changing these vault limits as they act as a default value
	// used by the Vault OCR plugin -- changing these values could cause issues with the plugin during an image
	// upgrade as nodes apply the old and new values inconsistently. A safe upgrade path
	// must ensure that we are overriding the default in the onchain configuration for the contract.
	VaultCiphertextSizeLimit:          Size(2 * config.KByte),
	VaultIdentifierKeySizeLimit:       Size(64 * config.Byte),
	VaultIdentifierOwnerSizeLimit:     Size(64 * config.Byte),
	VaultIdentifierNamespaceSizeLimit: Size(64 * config.Byte),
	VaultPluginBatchSizeLimit:         Int(10),
	VaultRequestBatchSizeLimit:        Int(10),
	VaultShareSizeLimit:               Size(600 * config.Byte),

	VaultMaxQuerySizeLimit:       Size(102400 * config.Byte),
	VaultMaxObservationSizeLimit: Size(512 * config.KByte),
	// Back of the envelope calculation:
	// - An item can contain 2KB of ciphertext, 192 bytes of metadata (key, owner, namespace),
	// a UUID (16 bytes) plus some overhead = ~2.5KB per item
	// There can be 10 such items in a request, and 20 per batch, so 2.5KB * 10 * 20 = 500KB
	// However as a buffer for reports, which have additional data, setting the next 2 fields to 2 mb.
	VaultMaxReportsPlusPrecursorSizeLimit: Size(2 * config.MByte),
	VaultMaxReportSizeLimit:               Size(2 * config.MByte),
	VaultMaxReportCount:                   Int(10),
	// assumption for largest item:
	// create request with the maximum ciphertext length:
	// - 192 bytes (sum of MaxIdentifierKeyLengthBytes + MaxIdentifierOwnerLengthBytes + MaxIdentifierNamespaceLengthBytes)
	// - 2048 bytes (MaxCiphertextLengthBytes)
	// = ~2240 bytes for an item
	// There are 10 items per request (separate vault setting), 10 request per batch (BatchSize)
	// i.e. ~224 KB per batch
	// For a batch we will write:
	// - a secret + metadata record per item
	//   - the secrets are 224 KB total
	//   - the metadata is a list of secret identifiers,
	//     there are a maximum of 100 secrets per owner (MaxSecretsPerOwner)
	//     i.e. 192 bytes * 100 = ~19.2 KB
	// - the pending queue
	//   - 10 requests in the pending queue, each request is ~22.4Kb = ~22.4 KB
	//   - an index record =  8bytes
	// - total = ~224 KB + ~19.2 KB + ~224 KB + 8 bytes = ~467.2 KB
	// Setting to 1.4MB to allow for some buffer.
	VaultMaxKeyValueModifiedKeysPlusValuesSizeLimit: Size(1468006 * config.Byte),
	// 10 batch size * 10 items per batch * 2 records modified per item (secret + metadata record)
	// plus 10 batchsize items in the pending queue + 1 index record
	// = 211 total.
	// plus some buffer.
	VaultMaxKeyValueModifiedKeys: Int(300),
	// Assuming a request is max 25KB, we add a bit of buffer to allow some room.
	VaultMaxBlobPayloadSizeLimit: Size(25600 * config.Byte),
	// Per docs, this should allow some additional buffer to allow for reaping time.
	VaultMaxPerOracleUnexpiredBlobCumulativePayloadSizeLimit: Size(31457280 * config.Byte),
	VaultMaxPerOracleUnexpiredBlobCount:                      Int(1000),

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
		ExecutionTimestampsEnabled:    Bool(false),
		WASMMemoryLimit:               Size(100 * config.MByte),
		WASMBinarySizeLimit:           Size(100 * config.MByte),
		WASMCompressedBinarySizeLimit: Size(20 * config.MByte),
		WASMConfigSizeLimit:           Size(config.MByte),
		WASMSecretsSizeLimit:          Size(config.MByte),
		LogLineLimit:                  Size(config.KByte),
		LogEventLimit:                 Int(1_000),
		ChainAllowed: PerChainSelector(Bool(false), map[string]bool{
			// geth-devnet2
			"12922642891491394802": true,
			// geth-testnet
			"3379446385462418246": true,
			// solana-testnet
			"12463857294658392847": true,
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
				ReportSizeLimit: Size(5 * config.KByte),
			},
			Solana: solanaChainWrite{
				ReportSizeLimit: Size(265 * config.Byte),
				GasLimit:        PerChainSelector(Uint32(300_000), map[string]uint32{}),
			},
			Aptos: aptosChainWrite{
				ReportSizeLimit: Size(5 * config.KByte),
				GasLimit:        PerChainSelector(Uint64(2_000_000), map[string]uint64{}),
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
		ConfidentialHTTP: confidentialHTTP{
			CallLimit:         Int(5),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(10 * config.KByte),
			ResponseSizeLimit: Size(100 * config.KByte),
		},
		Secrets: secrets{
			CallLimit: Int(5),
		},

		FeatureMultiTriggerExecutionIDsActiveAt: Time(time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)),
		FeatureMultiTriggerExecutionIDsActivePeriod: TimeRange(
			time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2101, 1, 1, 0, 0, 0, 0, time.UTC)),
		FeatureChainCapabilityHashBasedOCRActivePeriod: TimeRange(
			time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2101, 1, 1, 0, 0, 0, 0, time.UTC)),
	},
}

type Schema struct {
	WorkflowLimit                          Setting[int] `unit:"{workflow}"` // Deprecated
	WorkflowExecutionConcurrencyLimit      Setting[int] `unit:"{workflow}"`
	GatewayIncomingPayloadSizeLimit        Setting[config.Size]
	GatewayVaultManagementEnabled          Setting[bool]
	VaultJWTAuthEnabled                    Setting[bool]
	VaultOrgIdAsSecretOwnerEnabled         Setting[bool]
	GatewayHTTPGlobalRate                  Setting[config.Rate]
	GatewayHTTPPerNodeRate                 Setting[config.Rate]
	TriggerRegistrationStatusUpdateTimeout Setting[time.Duration]

	VaultCiphertextSizeLimit          Setting[config.Size]
	VaultShareSizeLimit               Setting[config.Size]
	VaultIdentifierKeySizeLimit       Setting[config.Size]
	VaultIdentifierOwnerSizeLimit     Setting[config.Size]
	VaultIdentifierNamespaceSizeLimit Setting[config.Size]
	VaultPluginBatchSizeLimit         Setting[int] `unit:"{request}"`
	VaultRequestBatchSizeLimit        Setting[int] `unit:"{request}"`

	VaultMaxQuerySizeLimit                                   Setting[config.Size]
	VaultMaxObservationSizeLimit                             Setting[config.Size]
	VaultMaxReportsPlusPrecursorSizeLimit                    Setting[config.Size]
	VaultMaxReportSizeLimit                                  Setting[config.Size]
	VaultMaxReportCount                                      Setting[int]
	VaultMaxKeyValueModifiedKeysPlusValuesSizeLimit          Setting[config.Size]
	VaultMaxKeyValueModifiedKeys                             Setting[int]
	VaultMaxBlobPayloadSizeLimit                             Setting[config.Size]
	VaultMaxPerOracleUnexpiredBlobCumulativePayloadSizeLimit Setting[config.Size]
	VaultMaxPerOracleUnexpiredBlobCount                      Setting[int]

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

	ExecutionConcurrencyLimit  Setting[int] `unit:"{workflow}"`
	ExecutionTimeout           Setting[time.Duration]
	ExecutionResponseLimit     Setting[config.Size]
	ExecutionTimestampsEnabled Setting[bool]

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

	ChainWrite       chainWrite
	ChainRead        chainRead
	Consensus        consensus
	HTTPAction       httpAction
	ConfidentialHTTP confidentialHTTP
	Secrets          secrets

	FeatureMultiTriggerExecutionIDsActiveAt     Setting[config.Timestamp] // Deprecated
	FeatureMultiTriggerExecutionIDsActivePeriod Setting[Range[config.Timestamp]]
	FeatureChainCapabilityHashBasedOCRActivePeriod Setting[Range[config.Timestamp]]
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
	TargetsLimit    Setting[int]         `unit:"{target}"`
	ReportSizeLimit Setting[config.Size] // Deprecated

	EVM    evmChainWrite
	Solana solanaChainWrite
	Aptos  aptosChainWrite
}
type solanaChainWrite struct {
	ReportSizeLimit Setting[config.Size]
	GasLimit        SettingMap[uint32] `unit:"{gas}"`
}
type aptosChainWrite struct {
	ReportSizeLimit Setting[config.Size]
	GasLimit        SettingMap[uint64] `unit:"{gas}"`
}
type evmChainWrite struct {
	TransactionGasLimit Setting[uint64]    `unit:"{gas}"` // Deprecated
	GasLimit            SettingMap[uint64] `unit:"{gas}"`
	ReportSizeLimit     Setting[config.Size]
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
type confidentialHTTP struct {
	CallLimit         Setting[int] `unit:"{call}"`
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
