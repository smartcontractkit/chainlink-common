// Package cresettings contains configurable settings definitions for the CRE.
package cresettings

import (
	"log"
	"time"

	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	. "github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func init() {
	err := InitKeys(&Config)
	if err != nil {
		log.Fatalf("failed to initialize keys: %v", err)
	}
}

var Config = cfg{
	//TODO prefer Uints?

	OrganizationLimit:              Int(10_000),
	WorkflowLimit:                  Int(-1), //TODO
	WorkflowRegistrationQueueLimit: Int(-1), //TODO
	TriggerEventQueueLimit:         Int(1_000),
	TriggerEventQueueTimeout:       Duration(10 * time.Minute),
	PerOrg: orgs{
		UserLimit:                   Int(100),
		WorkflowLimit:               Int(40),
		WorkflowDeploymentRateLimit: Rate(rate.Every(time.Minute), 1),
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
		CapabilityConcurrencyLimit:    Int(-1), //TODO
		ExecutionConcurrencyLimit:     Int(10),
		ExecutionTimeout:              Duration(5 * time.Minute),
		ExecutionResponseLimit:        Size(100 * config.KByte),
		WASMExecutionTimeout:          Duration(60 * time.Second),
		WASMMemoryLimit:               Size(100 * config.MByte),
		BinarySizeLimit:               Size(30 * config.MByte),
		ArtifactSizeLimit:             Size(-1), //TODO
		ConsensusObservationSizeLimit: Size(100 * config.KByte),
		ConsensusCallsLimit:           Int(2),
		LogTriggerLimit:               Int(-1),  //TODO
		LogSizeLimit:                  Size(-1), //TODO line or total?
		LogEventLimit:                 Int(1_000),

		CRONTrigger: cronTrigger{
			RateLimit: Rate(rate.Every(30*time.Second), 1),
		},
		HTTPTrigger: httpTrigger{
			RateLimit:                Rate(rate.Every(30*time.Second), 3),
			AuthRateLimit:            Rate(-1, -1), //TODO
			IncomingPayloadSizeLimit: Size(10 * config.KByte),
			OutgoingPayloadSizeLimit: Size(-1), //TODO
		},
		LogTrigger: logTrigger{
			Limit:                    Int(-1),
			EventRateLimit:           Rate(-1, -1), //TODO
			RateLimit:                Rate(rate.Every(10*time.Second), -1),
			FilterAddressLimit:       Int(-1), //TODO
			FilterTopicsPerSlotLimit: Int(-1), //TODO
		},
		HTTPAction: httpAction{
			RateLimit:         Rate(rate.Every(30*time.Second), 3),
			ResponseSizeLimit: Size(10 * config.KByte),
			ConnectionTimeout: Duration(10 * time.Second),
			RequestSizeLimit:  Size(100 * config.KByte),
		},
		ChainWrite: chainWrite{
			RateLimit:           Rate(rate.Every(30*time.Second), 3),
			TargetsLimit:        Int(3),
			ReportSizeLimit:     Size(config.KByte),
			TransactionGasLimit: Int(-1), //TODO
		},
		ChainRead: chainRead{
			RateLimit:        Rate(-1, -1), //TODO
			PayloadSizeLimit: Size(-1),     //TODO
		},
	},
}

type cfg struct {
	OrganizationLimit              Setting[int]
	WorkflowLimit                  Setting[int]
	WorkflowRegistrationQueueLimit Setting[int]
	TriggerEventQueueLimit         Setting[int]
	TriggerEventQueueTimeout       Setting[time.Duration]

	PerOrg      orgs      `scope:"org"`
	PerOwner    owners    `scope:"owner"`
	PerWorkflow workflows `scope:"workflow"`
}
type orgs struct {
	UserLimit                   Setting[int]
	WorkflowLimit               Setting[int]
	WorkflowDeploymentRateLimit Setting[config.Rate]
	ZeroBalancePruningTimeout   Setting[time.Duration]
}

type owners struct {
	ExecutionConcurrencyLimit Setting[int]
}

type workflows struct {
	TriggerLimit                  Setting[int]
	TriggerRateLimit              Setting[config.Rate]
	TriggerRegistrationTimeout    Setting[time.Duration]
	TriggerSubscriptionTimeout    Setting[time.Duration]
	CapabilityConcurrencyLimit    Setting[int]
	ExecutionConcurrencyLimit     Setting[int]
	ExecutionTimeout              Setting[time.Duration]
	ExecutionResponseLimit        Setting[config.Size]
	WASMExecutionTimeout          Setting[time.Duration]
	WASMMemoryLimit               Setting[config.Size]
	BinarySizeLimit               Setting[config.Size]
	ArtifactSizeLimit             Setting[config.Size]
	ConsensusObservationSizeLimit Setting[config.Size]
	ConsensusCallsLimit           Setting[int]
	LogTriggerLimit               Setting[int]
	LogSizeLimit                  Setting[config.Size]
	LogEventLimit                 Setting[int]

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
type httpTrigger struct {
	RateLimit                Setting[config.Rate]
	AuthRateLimit            Setting[config.Rate]
	IncomingPayloadSizeLimit Setting[config.Size]
	OutgoingPayloadSizeLimit Setting[config.Size]
}
type logTrigger struct {
	Limit                    Setting[int]
	EventRateLimit           Setting[config.Rate]
	RateLimit                Setting[config.Rate]
	FilterAddressLimit       Setting[int]
	FilterTopicsPerSlotLimit Setting[int]
}
type httpAction struct {
	RateLimit         Setting[config.Rate]
	ResponseSizeLimit Setting[config.Size]
	ConnectionTimeout Setting[time.Duration]
	RequestSizeLimit  Setting[config.Size]
}
type chainWrite struct {
	RateLimit       Setting[config.Rate]
	TargetsLimit    Setting[int]
	ReportSizeLimit Setting[config.Size]
	//TODO EVM.TransactionGasLimit
	//TODO Transaction.EVMGasLimit
	TransactionGasLimit Setting[int]
}
type chainRead struct {
	RateLimit        Setting[config.Rate]
	PayloadSizeLimit Setting[config.Size]
}
