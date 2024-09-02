package capabilities

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// CapabilityType is an enum for the type of capability.
type CapabilityType string

var ErrStopExecution = &errStopExecution{}

type errStopExecution struct{}

const errStopExecutionMsg = "__workflow_stop_execution"

func (e errStopExecution) Error() string {
	return errStopExecutionMsg
}

func (e errStopExecution) Is(err error) bool {
	return err.Error() == errStopExecutionMsg
}

// CapabilityType enum values.
const (
	CapabilityTypeUnknown   CapabilityType = "unknown"
	CapabilityTypeTrigger   CapabilityType = "trigger"
	CapabilityTypeAction    CapabilityType = "action"
	CapabilityTypeConsensus CapabilityType = "consensus"
	CapabilityTypeTarget    CapabilityType = "target"
)

// IsValid checks if the capability type is valid.
func (c CapabilityType) IsValid() error {
	switch c {
	case CapabilityTypeTrigger,
		CapabilityTypeAction,
		CapabilityTypeConsensus,
		CapabilityTypeTarget:
		return nil
	case CapabilityTypeUnknown:
		return fmt.Errorf("invalid capability type: %s", c)
	}

	return fmt.Errorf("invalid capability type: %s", c)
}

// CapabilityResponse is a struct for the Execute response of a capability.
type CapabilityResponse struct {
	Value *values.Map
}

type RequestMetadata struct {
	WorkflowID               string
	WorkflowOwner            string
	WorkflowExecutionID      string
	WorkflowName             string
	WorkflowDonID            uint32
	WorkflowDonConfigVersion uint32
	ReferenceID              string
}

type RegistrationMetadata struct {
	WorkflowID    string
	WorkflowOwner string
}

// CapabilityRequest is a struct for the Execute request of a capability.
type CapabilityRequest struct {
	Metadata RequestMetadata
	Config   *values.Map
	Inputs   *values.Map
}

type TriggerEvent struct {
	// The ID of the trigger capability
	TriggerType string
	// The ID of the trigger event
	ID string
	// Trigger-specific payload
	Outputs *values.Map
}

type RegisterToWorkflowRequest struct {
	Metadata RegistrationMetadata
	Config   *values.Map
}

type UnregisterFromWorkflowRequest struct {
	Metadata RegistrationMetadata
	Config   *values.Map
}

// CallbackExecutable is an interface for executing a capability.
type CallbackExecutable interface {
	RegisterToWorkflow(ctx context.Context, request RegisterToWorkflowRequest) error
	UnregisterFromWorkflow(ctx context.Context, request UnregisterFromWorkflowRequest) error
	Execute(ctx context.Context, request CapabilityRequest) (CapabilityResponse, error)
}

type Validatable interface {
	// ValidateSchema returns the JSON schema for the capability.
	//
	// This schema includes the configuration, input and output schemas.
	Schema() (string, error)
}

// BaseCapability interface needs to be implemented by all capability types.
// Capability interfaces are intentionally duplicated to allow for an easy change
// or extension in the future.
type BaseCapability interface {
	Info(ctx context.Context) (CapabilityInfo, error)
}

type TriggerRegistrationRequest struct {
	// TriggerID uniquely identifies the trigger by concatenating
	// the workflow ID and the trigger's index in the spec.
	TriggerID string

	Metadata RequestMetadata
	Config   *values.Map
}

type TriggerResponse struct {
	Event TriggerEvent
	Err   error
}

type TriggerExecutable interface {
	RegisterTrigger(ctx context.Context, request TriggerRegistrationRequest) (<-chan TriggerResponse, error)
	UnregisterTrigger(ctx context.Context, request TriggerRegistrationRequest) error
}

// TriggerCapability interface needs to be implemented by all trigger capabilities.
type TriggerCapability interface {
	BaseCapability
	TriggerExecutable
}

// CallbackCapability is the interface implemented by action, consensus and target
// capabilities. This interface is useful when trying to capture capabilities of varying types.
type CallbackCapability interface {
	BaseCapability
	CallbackExecutable
}

// ActionCapability interface needs to be implemented by all action capabilities.
type ActionCapability interface {
	CallbackCapability
}

// ConsensusCapability interface needs to be implemented by all consensus capabilities.
type ConsensusCapability interface {
	CallbackCapability
}

// TargetsCapability interface needs to be implemented by all target capabilities.
type TargetCapability interface {
	CallbackCapability
}

// DON represents a network of connected nodes.
//
// For an example of an empty DON check, see the following link:
// https://github.com/smartcontractkit/chainlink/blob/develop/core/capabilities/transmission/local_target_capability.go#L31
type DON struct {
	ID               uint32
	ConfigVersion    uint32
	Members          []p2ptypes.PeerID
	F                uint8
	IsPublic         bool
	AcceptsWorkflows bool
}

// Node contains the node's peer ID and the DONs it is part of.
//
// Note the following relationships between the workflow and capability DONs and this node.
//
// There is a 1:0..1 relationship between this node and a workflow DON.
// This means that this node can be part at most one workflow DON at a time.
// As a side note, a workflow DON can have multiple nodes.
//
// There is a 1:N relationship between this node and capability DONs, where N is the number of capability DONs.
// This means that this node can be part of multiple capability DONs at a time.
//
// Although WorkflowDON is a value rather than a pointer, a node can be part of no workflow DON but 0 or more capability DONs.
// You can assert this by checking for zero values in the WorkflowDON field.
// See https://github.com/smartcontractkit/chainlink/blob/develop/core/capabilities/transmission/local_target_capability.go#L31 for an example.
type Node struct {
	PeerID         *p2ptypes.PeerID
	WorkflowDON    DON
	CapabilityDONs []DON
}

// CapabilityInfo is a struct for the info of a capability.
type CapabilityInfo struct {
	// The capability ID is a fully qualified identifier for the capability.
	//
	// It takes the form of `{name}:{label1_key}_{labe1_value}:{label2_key}_{label2_value}@{version}`
	//
	// The labels within the ID are ordered alphanumerically.
	ID             string
	CapabilityType CapabilityType
	Description    string
	DON            *DON
	IsLocal        bool
}

// Parse out the version from the ID.
func (c CapabilityInfo) Version() string {
	return c.ID[strings.Index(c.ID, "@")+1:]
}

// Info returns the info of the capability.
func (c CapabilityInfo) Info(ctx context.Context) (CapabilityInfo, error) {
	return c, nil
}

// This regex allows for the following format:
//
// {name}:{label1_key}_{labe1_value}:{label2_key}_{label2_value}@{version}
//
// The version regex is taken from https://semver.org/, but has been modified to support only major versions.
//
// It is also validated when a workflow is being ingested. See the following link for more details:
// https://github.com/smartcontractkit/chainlink/blob/a0d1ce5e9cddc540bba8eb193865646cf0ebc0a8/core/services/workflows/models_yaml.go#L309
//
// The difference between the regex within the link above and this one is that we do not use double backslashes, since
// we only needed those for JSON schema regex validation.
var idRegex = regexp.MustCompile(`^[a-z0-9_\-:]+@(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

const (
	// TODO: this length was largely picked arbitrarily.
	// Consider what a realistic/desirable value should be.
	// See: https://smartcontract-it.atlassian.net/jira/software/c/projects/KS/boards/182
	idMaxLength = 128
)

func newCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
	don *DON,
	isLocal bool,
) (CapabilityInfo, error) {
	if len(id) > idMaxLength {
		return CapabilityInfo{}, fmt.Errorf("invalid id: %s exceeds max length %d", id, idMaxLength)
	}
	if !idRegex.MatchString(id) {
		return CapabilityInfo{}, fmt.Errorf("invalid id: %s. Allowed: %s", id, idRegex)
	}

	if err := capabilityType.IsValid(); err != nil {
		return CapabilityInfo{}, err
	}

	return CapabilityInfo{
		ID:             id,
		CapabilityType: capabilityType,
		Description:    description,
		DON:            don,
		IsLocal:        isLocal,
	}, nil
}

// NewCapabilityInfo returns a new CapabilityInfo.
func NewCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
) (CapabilityInfo, error) {
	return newCapabilityInfo(id, capabilityType, description, nil, true)
}

// NewRemoteCapabilityInfo returns a new CapabilityInfo for remote capabilities.
// This is largely intended for internal use by the registry syncer.
// Capability developers should use `NewCapabilityInfo` instead as this
// omits the requirement to pass in the DON Info.
func NewRemoteCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
	don *DON,
) (CapabilityInfo, error) {
	return newCapabilityInfo(id, capabilityType, description, don, false)
}

// MustNewCapabilityInfo returns a new CapabilityInfo,
// `panic`ing if we could not instantiate a CapabilityInfo.
func MustNewCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
) CapabilityInfo {
	c, err := NewCapabilityInfo(id, capabilityType, description)
	if err != nil {
		panic(err)
	}

	return c
}

// MustNewRemoteCapabilityInfo returns a new CapabilityInfo,
// `panic`ing if we could not instantiate a CapabilityInfo.
func MustNewRemoteCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
	don *DON,
) CapabilityInfo {
	c, err := NewRemoteCapabilityInfo(id, capabilityType, description, don)
	if err != nil {
		panic(err)
	}

	return c
}

const (
	DefaultRegistrationRefresh = 30 * time.Second
	DefaultRegistrationExpiry  = 2 * time.Minute
	DefaultMessageExpiry       = 2 * time.Minute
)

type RemoteTriggerConfig struct {
	RegistrationRefresh     time.Duration
	RegistrationExpiry      time.Duration
	MinResponsesToAggregate uint32
	MessageExpiry           time.Duration
}

type RemoteTargetConfig struct {
	RequestHashExcludedAttributes []string
}

// NOTE: consider splitting this config into values stored in Registry (KS-118)
// and values defined locally by Capability owners.
func (c *RemoteTriggerConfig) ApplyDefaults() {
	if c == nil {
		return
	}
	if c.RegistrationRefresh == 0 {
		c.RegistrationRefresh = DefaultRegistrationRefresh
	}
	if c.RegistrationExpiry == 0 {
		c.RegistrationExpiry = DefaultRegistrationExpiry
	}
	if c.MessageExpiry == 0 {
		c.MessageExpiry = DefaultMessageExpiry
	}
}

type CapabilityConfiguration struct {
	DefaultConfig       *values.Map
	RemoteTriggerConfig *RemoteTriggerConfig
	RemoteTargetConfig  *RemoteTargetConfig
}
