// Package cresettings contains configurable settings definitions for nodes in the CRE.
package cresettings

import (
	"log"
	"time"

	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	. "github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func init() {
	err := InitConfig(&Config)
	if err != nil {
		log.Fatalf("failed to initialize keys: %v", err)
	}
}

var Config = cfg{
	WorkflowLimit:                  Int(200),
	WorkflowRegistrationQueueLimit: Int(20),

	HTTPTrigger: httpTriggerGlobal{
		AuthRateLimit: Rate(100, -1), //TODO
	},

	PerOrg: orgs{
		WorkflowDeploymentRateLimit: Rate(rate.Every(time.Minute), 1),
		TriggerEventQueueLimit:      Int(1_000),
		TriggerEventQueueTimeout:    Duration(10 * time.Minute),
		ZeroBalancePruningTimeout:   Duration(24 * time.Hour),
	},
	PerOwner: owners{
		ExecutionConcurrencyLimit: Int(50),
	},
	PerWorkflow: workflows{
		TriggerLimit:                  Int(10),
		TriggerRateLimit:              Rate(rate.Every(30*time.Second), 3),
		TriggerRegistrationTimeout:    Duration(10 * time.Second),
		TriggerSubscriptionTimeout:    Duration(5 * time.Second),
		CapabilityConcurrencyLimit:    Int(3),
		ExecutionConcurrencyLimit:     Int(10),
		ExecutionTimeout:              Duration(5 * time.Minute),
		ExecutionResponseLimit:        Size(100 * config.KByte),
		WASMExecutionTimeout:          Duration(60 * time.Second),
		WASMMemoryLimit:               Size(100 * config.MByte),
		BinarySizeLimit:               Size(30 * config.MByte),
		ConsensusObservationSizeLimit: Size(10 * config.KByte),
		ConsensusCallsLimit:           Int(2),
		LogLineLimit:                  Size(config.KByte),
		LogEventLimit:                 Int(1_000),

		CRONTrigger: cronTrigger{
			RateLimit: Rate(rate.Every(30*time.Second), 1),
		},
		HTTPTrigger: httpTrigger{
			RateLimit:                Rate(rate.Every(30*time.Second), 3),
			AuthRateLimit:            Rate(1, -1), //TODO
			IncomingPayloadSizeLimit: Size(10 * config.KByte),
			OutgoingPayloadSizeLimit: Size(-1), //TODO
		},
		LogTrigger: logTrigger{
			RateLimit:                Rate(rate.Every(10*time.Second), -1), //TODO
			Limit:                    Int(5),
			EventRateLimit:           Rate(-1, -1), //TODO
			FilterAddressLimit:       Int(5),
			FilterTopicsPerSlotLimit: Int(10),
		},
		HTTPAction: httpAction{
			RateLimit:         Rate(rate.Every(30*time.Second), 3),
			ResponseSizeLimit: Size(10 * config.KByte),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(100 * config.KByte),
		},
		ChainWrite: chainWrite{
			RateLimit:       Rate(rate.Every(30*time.Second), 3),
			TargetsLimit:    Int(3),
			ReportSizeLimit: Size(config.KByte),
			EVM: evmChainWrite{
				TransactionGasLimit: Int(-1), //TODO
			},
		},
		ChainRead: chainRead{
			CallLimit:          Int(3),
			LogQueryBlockLimit: Int(100),
			PayloadSizeLimit:   Size(5 * config.KByte),
		},
	},
}

type cfg struct {
	WorkflowLimit                  Setting[int] `unit:"{workflow}"`
	WorkflowRegistrationQueueLimit Setting[int] `unit:"{workflow}"`

	HTTPTrigger httpTriggerGlobal

	PerOrg      orgs      `scope:"org"`
	PerOwner    owners    `scope:"owner"`
	PerWorkflow workflows `scope:"workflow"`
}
type orgs struct {
	WorkflowDeploymentRateLimit Setting[config.Rate]
	TriggerEventQueueLimit      Setting[int] `unit:"{trigger}"`
	TriggerEventQueueTimeout    Setting[time.Duration]
	ZeroBalancePruningTimeout   Setting[time.Duration]
}

type owners struct {
	ExecutionConcurrencyLimit Setting[int] `unit:"{workflow}"`
}

type workflows struct {
	TriggerLimit                  Setting[int] `unit:"{trigger}"`
	TriggerRateLimit              Setting[config.Rate]
	TriggerRegistrationTimeout    Setting[time.Duration]
	TriggerSubscriptionTimeout    Setting[time.Duration]
	CapabilityConcurrencyLimit    Setting[int] `unit:"{capability}"`
	ExecutionConcurrencyLimit     Setting[int] `unit:"{workflow}"`
	ExecutionTimeout              Setting[time.Duration]
	ExecutionResponseLimit        Setting[config.Size]
	WASMExecutionTimeout          Setting[time.Duration]
	WASMMemoryLimit               Setting[config.Size]
	BinarySizeLimit               Setting[config.Size]
	ConsensusObservationSizeLimit Setting[config.Size]
	ConsensusCallsLimit           Setting[int] `unit:"{call}"`
	LogLineLimit                  Setting[config.Size]
	LogEventLimit                 Setting[int] `unit:"{log}"`

	CRONTrigger cronTrigger
	HTTPTrigger httpTrigger
	LogTrigger  logTrigger
	HTTPAction  httpAction
	ChainWrite  chainWrite
	ChainRead   chainRead
}

type cronTrigger struct {
	RateLimit Setting[config.Rate]
}
type httpTriggerGlobal struct {
	AuthRateLimit Setting[config.Rate]
}
type httpTrigger struct {
	RateLimit                Setting[config.Rate]
	AuthRateLimit            Setting[config.Rate]
	IncomingPayloadSizeLimit Setting[config.Size]
	OutgoingPayloadSizeLimit Setting[config.Size]
}
type logTrigger struct {
	RateLimit                Setting[config.Rate]
	Limit                    Setting[int] `unit:"{trigger}"`
	EventRateLimit           Setting[config.Rate]
	FilterAddressLimit       Setting[int] `unit:"{address}"`
	FilterTopicsPerSlotLimit Setting[int] `unit:"{topic}"`
}
type httpAction struct {
	RateLimit         Setting[config.Rate]
	ResponseSizeLimit Setting[config.Size]
	ConnectionTimeout Setting[time.Duration]
	RequestSizeLimit  Setting[config.Size]
}
type chainWrite struct {
	RateLimit       Setting[config.Rate]
	TargetsLimit    Setting[int] `unit:"{target}"`
	ReportSizeLimit Setting[config.Size]

	EVM evmChainWrite
}
type evmChainWrite struct {
	TransactionGasLimit Setting[int] `unit:"{gas}"`
}

type chainRead struct {
	CallLimit          Setting[int] `unit:"{call}"`
	LogQueryBlockLimit Setting[int] `unit:"{block}"`
	PayloadSizeLimit   Setting[config.Size]
}
