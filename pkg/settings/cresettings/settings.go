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
	WorkflowExecutionConcurrencyLimit:           Int(200), // as discussed on 10/23/2025 increase to 200
	WorkflowTriggerRateLimit:                    Rate(200, 200), // rps value and burst value globally
	GatewayUnauthenticatedRequestRateLimit:      Rate(rate.Every(time.Second/100), 100), // removed, but enforced in infra, Jin is expert, global
	GatewayUnauthenticatedRequestRateLimitPerIP: Rate(rate.Every(time.Second), 1), // removed, 
	GatewayIncomingPayloadSizeLimit:             Size(1 * config.MByte), // this exists make it 1MB

	PerOrg: Orgs{
		WorkflowDeploymentRateLimit: Rate(rate.Every(time.Minute), 1), // one deploy per minute and no burst
		ZeroBalancePruningTimeout:   Duration(24 * time.Hour), // account has zero balance we garbage collect resources (not in effect)
	},
	PerOwner: Owners{
		WorkflowExecutionConcurrencyLimit: Int(5), //
		WorkflowTriggerRateLimit:          Rate(5, 5), // rps value and burst value for this owner on this DON
	},
	PerWorkflow: Workflows{
		TriggerLimit:                  Int(10), // max number of triggers registered
		TriggerRateLimit:              Rate(rate.Every(30*time.Second), 3), // how often you run
		TriggerRegistrationsTimeout:   Duration(10 * time.Second),
		TriggerEventQueueLimit:        Int(1_000),
		TriggerEventQueueTimeout:      Duration(10 * time.Minute),
		TriggerSubscriptionTimeout:    Duration(15 * time.Second), // top level and includes TriggerRateLimit
		TriggerSubscriptionLimit:      Int(10), // number of subscriptions in this phase ... should thi be here if we have line 63 TriggerLimit?
		CapabilityConcurrencyLimit:    Int(3),  // concurrent number of calls, but they will wait, they will block till they can run. execution helper is paused
		CapabilityCallTimeout:         Duration(3 * time.Minute), // timeout on capability call
		SecretsConcurrencyLimit:       Int(5),  
		ExecutionConcurrencyLimit:     Int(5), // same as per owner , question on http
		ExecutionTimeout:              Duration(5 * time.Minute), //changing to 5 for now
		ExecutionResponseLimit:        Size(100 * config.KByte), // go to logs? in future wf invoke others
		WASMMemoryLimit:               Size(100 * config.MByte), // need load test
		WASMBinarySizeLimit:           Size(100 * config.MByte), 
		WASMCompressedBinarySizeLimit: Size(20 * config.MByte), // limit for storage service (check with AW if another place)
		WASMConfigSizeLimit:           Size(1 * config.MByte), // make 1MB
		WASMSecretsSizeLimit:          Size(1 * config.MByte), //make 1MB
		WASMResponseSizeLimit:         Size(100 * config.KByte), // what is diff between this and ExecutionResponseLimit (possible ) Needs investi
		ConsensusObservationSizeLimit: Size(100 * config.KByte), // investigate this - load test 
		ConsensusCallsLimit:           Int(2000), // plugged into execution helper (consider removing)
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
			TargetsLimit:    Int(10), // making this 10
			ReportSizeLimit: Size(config.KByte),
			EVM: evmChainWrite{
				TransactionGasLimit: Uint64(5_000_000),
			},
		},
		ChainRead: chainRead{
			CallLimit:          Int(10), // making this 10
			LogQueryBlockLimit: Uint64(100),
			PayloadSizeLimit:   Size(5 * config.KByte),
		},
		Consensus: consensus{
			ObservationSizeLimit: Size(100 * config.KByte),
			CallLimit:            Int(2000), // consider removing?
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
