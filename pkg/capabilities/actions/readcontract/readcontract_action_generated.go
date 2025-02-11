// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli, DO NOT EDIT.

package readcontract

import (
	"encoding/json"
	"fmt"
)

type Action struct {
	// Config corresponds to the JSON schema field "Config".
	Config *Config `json:"Config,omitempty" yaml:"Config,omitempty" mapstructure:"Config,omitempty"`

	// Inputs corresponds to the JSON schema field "Inputs".
	Inputs *Input `json:"Inputs,omitempty" yaml:"Inputs,omitempty" mapstructure:"Inputs,omitempty"`

	// Outputs corresponds to the JSON schema field "Outputs".
	Outputs *Output `json:"Outputs,omitempty" yaml:"Outputs,omitempty" mapstructure:"Outputs,omitempty"`
}

type Config struct {
	// ContractAddress corresponds to the JSON schema field "ContractAddress".
	ContractAddress string `json:"ContractAddress" yaml:"ContractAddress" mapstructure:"ContractAddress"`

	// ContractName corresponds to the JSON schema field "ContractName".
	ContractName string `json:"ContractName" yaml:"ContractName" mapstructure:"ContractName"`

	// ContractReaderConfig corresponds to the JSON schema field
	// "ContractReaderConfig".
	ContractReaderConfig string `json:"ContractReaderConfig" yaml:"ContractReaderConfig" mapstructure:"ContractReaderConfig"`

	// ReadIdentifier corresponds to the JSON schema field "ReadIdentifier".
	ReadIdentifier string `json:"ReadIdentifier" yaml:"ReadIdentifier" mapstructure:"ReadIdentifier"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Config) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["ContractAddress"]; raw != nil && !ok {
		return fmt.Errorf("field ContractAddress in Config: required")
	}
	if _, ok := raw["ContractName"]; raw != nil && !ok {
		return fmt.Errorf("field ContractName in Config: required")
	}
	if _, ok := raw["ContractReaderConfig"]; raw != nil && !ok {
		return fmt.Errorf("field ContractReaderConfig in Config: required")
	}
	if _, ok := raw["ReadIdentifier"]; raw != nil && !ok {
		return fmt.Errorf("field ReadIdentifier in Config: required")
	}
	type Plain Config
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Config(plain)
	return nil
}

type Input struct {
	// ConfidenceLevel corresponds to the JSON schema field "ConfidenceLevel".
	ConfidenceLevel string `json:"ConfidenceLevel" yaml:"ConfidenceLevel" mapstructure:"ConfidenceLevel"`

	// Params corresponds to the JSON schema field "Params".
	Params InputParams `json:"Params" yaml:"Params" mapstructure:"Params"`

	// an optional step reference that is a non-data dependency for the current step
	StepDependency interface{} `json:"StepDependency,omitempty" yaml:"StepDependency,omitempty" mapstructure:"StepDependency,omitempty"`
}

type InputParams map[string]interface{}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Input) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["ConfidenceLevel"]; raw != nil && !ok {
		return fmt.Errorf("field ConfidenceLevel in Input: required")
	}
	if _, ok := raw["Params"]; raw != nil && !ok {
		return fmt.Errorf("field Params in Input: required")
	}
	type Plain Input
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Input(plain)
	return nil
}

type Output struct {
	// LatestValue corresponds to the JSON schema field "LatestValue".
	LatestValue interface{} `json:"LatestValue" yaml:"LatestValue" mapstructure:"LatestValue"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *Output) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["LatestValue"]; raw != nil && !ok {
		return fmt.Errorf("field LatestValue in Output: required")
	}
	type Plain Output
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = Output(plain)
	return nil
}
