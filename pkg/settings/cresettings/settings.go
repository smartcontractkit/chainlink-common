// Package cresettings contains configurable settings definitions for nodes in the CRE.
// Environment Variables:
//  - CL_CRE_SETTINGS_DEFAULT: defaults like in ./defaults.json - initializes Default
// 	- CL_CRE_SETTINGS: scoped settings like in ../settings/testdata/config.json - initializes DefaultGetter
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
	WorkflowLimit:                               Int(200),
	WorkflowRegistrationQueueLimit:              Int(20),
	WorkflowExecutionConcurrencyLimit:           Int(50),
	WorkflowTriggerRateLimit:                    Rate(200, 200),
	GatewayUnauthenticatedRequestRateLimit:      Rate(rate.Every(time.Second/100), -1),
	GatewayUnauthenticatedRequestRateLimitPerIP: Rate(rate.Every(time.Second), -1),
	GatewayIncomingPayloadSizeLimit:             Size(10 * config.KByte),

	PerOrg: Orgs{
		WorkflowDeploymentRateLimit: Rate(rate.Every(time.Minute), 1),
		ZeroBalancePruningTimeout:   Duration(24 * time.Hour),
	},
	PerOwner: Owners{
		WorkflowExecutionConcurrencyLimit: Int(50),
		WorkflowTriggerRateLimit:          Rate(200, 200),
	},
	PerWorkflow: Workflows{
		TriggerLimit:                  Int(10),
		TriggerRateLimit:              Rate(rate.Every(30*time.Second), 3),
		TriggerRegistrationsTimeout:   Duration(10 * time.Second),
		TriggerEventQueueLimit:        Int(1_000),
		TriggerEventQueueTimeout:      Duration(10 * time.Minute),
		TriggerSubscriptionTimeout:    Duration(5 * time.Second),
		TriggerSubscriptionLimit:      Int(10),
		CapabilityConcurrencyLimit:    Int(3),
		CapabilityCallTimeout:         Duration(8 * time.Minute),
		SecretsConcurrencyLimit:       Int(3),
		ExecutionConcurrencyLimit:     Int(10),
		ExecutionTimeout:              Duration(10 * time.Minute),
		ExecutionResponseLimit:        Size(100 * config.KByte),
		WASMExecutionTimeout:          Duration(60 * time.Second),
		WASMMemoryLimit:               Size(100 * config.MByte),
		WASMBinarySizeLimit:           Size(30 * config.MByte),
		WASMCompressedBinarySizeLimit: Size(20 * config.MByte),
		WASMConfigSizeLimit:           Size(30 * config.MByte),
		WASMSecretsSizeLimit:          Size(30 * config.MByte),
		WASMResponseSizeLimit:         Size(5 * config.MByte),
		ConsensusObservationSizeLimit: Size(10 * config.KByte),
		ConsensusCallsLimit:           Int(2),
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
			TargetsLimit:    Int(3),
			ReportSizeLimit: Size(config.KByte),
			EVM: evmChainWrite{
				TransactionGasLimit: Uint64(5_000_000),
			},
		},
		ChainRead: chainRead{
			CallLimit:          Int(3),
			LogQueryBlockLimit: Uint64(100),
			PayloadSizeLimit:   Size(5 * config.KByte),
		},
		Consensus: consensus{
			ObservationSizeLimit: Size(10 * config.KByte),
			CallLimit:            Int(2),
		},
		HTTPAction: httpAction{
			CallLimit:         Int(5),
			ResponseSizeLimit: Size(10 * config.KByte),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(100 * config.KByte),
			CacheAgeLimit:     Duration(10 * time.Minute),
		},
	},
}

type Schema struct {
	WorkflowLimit                               Setting[int] `unit:"{workflow}"`
	WorkflowRegistrationQueueLimit              Setting[int] `unit:"{workflow}"`
	WorkflowExecutionConcurrencyLimit           Setting[int] `unit:"{workflow}"`
	WorkflowTriggerRateLimit                    Setting[config.Rate]
	GatewayUnauthenticatedRequestRateLimit      Setting[config.Rate]
	GatewayUnauthenticatedRequestRateLimitPerIP Setting[config.Rate]
	GatewayIncomingPayloadSizeLimit             Setting[config.Size]

	PerOrg      Orgs      `scope:"org"`
	PerOwner    Owners    `scope:"owner"`
	PerWorkflow Workflows `scope:"workflow"`
}
type Orgs struct {
	WorkflowDeploymentRateLimit Setting[config.Rate]
	ZeroBalancePruningTimeout   Setting[time.Duration]
}

type Owners struct {
	WorkflowExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
	WorkflowTriggerRateLimit          Setting[config.Rate]
}

type Workflows struct {
	TriggerLimit                Setting[int] `unit:"{trigger}"`
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

	WASMExecutionTimeout          Setting[time.Duration]
	WASMMemoryLimit               Setting[config.Size]
	WASMBinarySizeLimit           Setting[config.Size]
	WASMCompressedBinarySizeLimit Setting[config.Size]
	WASMConfigSizeLimit           Setting[config.Size]
	WASMSecretsSizeLimit          Setting[config.Size]
	WASMResponseSizeLimit         Setting[config.Size]

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
	RateLimit Setting[config.Rate]
}
type httpTrigger struct {
	RateLimit Setting[config.Rate]
}
type logTrigger struct {
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
	ResponseSizeLimit Setting[config.Size]
	ConnectionTimeout Setting[time.Duration]
	RequestSizeLimit  Setting[config.Size]
	CacheAgeLimit     Setting[time.Duration]
}
type consensus struct {
	ObservationSizeLimit Setting[config.Size]
	CallLimit            Setting[int] `unit:"{call}"`
}
