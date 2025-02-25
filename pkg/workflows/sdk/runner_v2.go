package sdk

import "github.com/smartcontractkit/chainlink-common/pkg/values"

type DonRunner interface {
	// SubscribeToTrigger is meant to be called by generated code, prefer to use the generated code
	SubscribeToTrigger(id string, triggerCfg *values.Map, handler func(runtime DonRuntime, triggerOutputs *values.Map) (*values.Map, error)) error
	Run()
}
