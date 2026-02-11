package capabilities

import (
	"context"
	"fmt"
	"iter"
	"regexp"
	"strconv"
	"strings"
	"time"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
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
	return strings.Contains(err.Error(), errStopExecutionMsg)
}

// CapabilityType enum values.
const (
	CapabilityTypeUnknown   CapabilityType = "unknown"
	CapabilityTypeTrigger   CapabilityType = "trigger"
	CapabilityTypeAction    CapabilityType = "action"
	CapabilityTypeConsensus CapabilityType = "consensus"
	CapabilityTypeTarget    CapabilityType = "target"

	// CapabilityTypeCombined allows capabilities to offer both trigger and executable types.
	CapabilityTypeCombined CapabilityType = "combined"
)

// IsValid checks if the capability type is valid.
func (c CapabilityType) IsValid() error {
	switch c {
	case CapabilityTypeTrigger,
		CapabilityTypeAction,
		CapabilityTypeConsensus,
		CapabilityTypeTarget,
		CapabilityTypeCombined:
		return nil
	case CapabilityTypeUnknown:
		return fmt.Errorf("invalid capability type: %s", c)
	}

	return fmt.Errorf("invalid capability type: %s", c)
}

type CapabilitySpendType string

// CapabilityResponse is a struct for the Execute response of a capability.
type CapabilityResponse struct {
	// Value is used for DAG workflows
	Value    *values.Map
	Metadata ResponseMetadata

	// Payload is used for no DAG workflows
	Payload *anypb.Any
}

type ResponseMetadata struct {
	Metering []MeteringNodeDetail
	CapDON_N uint32
}

type MeteringNodeDetail struct {
	Peer2PeerID string
	SpendUnit   string
	SpendValue  string
}

// ResponseAndMetadata is the action's output structure that includes both the response and its metadata (billing).
type ResponseAndMetadata[T proto.Message] struct {
	Response         T
	ResponseMetadata ResponseMetadata
}

type SpendLimit struct {
	SpendType CapabilitySpendType
	Limit     string
}

type RequestMetadata struct {
	WorkflowID               string
	WorkflowOwner            string
	WorkflowExecutionID      string
	WorkflowName             string
	WorkflowDonID            uint32
	WorkflowDonConfigVersion uint32
	// The step reference ID of the workflow
	ReferenceID string
	// Use DecodedWorkflowName if the human readable name needs to be exposed, such as for logging purposes.
	DecodedWorkflowName string
	// SpendLimits is expected to be an array of tuples of spend type and limit. i.e. CONSENSUS -> 100_000
	SpendLimits                   []SpendLimit
	WorkflowTag                   string
	WorkflowRegistryChainSelector string
	WorkflowRegistryAddress       string
	EngineVersion                 string
}

func (m *RequestMetadata) ContextWithCRE(ctx context.Context) context.Context {
	val := contexts.CREValue(ctx)
	// preserve org, if set
	val.Owner = m.WorkflowOwner
	val.Workflow = m.WorkflowID
	return contexts.WithCRE(ctx, val)
}

type RegistrationMetadata struct {
	WorkflowID    string
	WorkflowOwner string
	// The step reference ID of the workflow
	ReferenceID string
}

func (m *RegistrationMetadata) ContextWithCRE(ctx context.Context) context.Context {
	val := contexts.CREValue(ctx)
	// preserve org, if set
	val.Owner = m.WorkflowOwner
	val.Workflow = m.WorkflowID
	return contexts.WithCRE(ctx, val)
}

// CapabilityRequest is a struct for the Execute request of a capability.
type CapabilityRequest struct {
	Metadata RequestMetadata

	// Config is used for DAG workflows
	Config *values.Map

	// Inputs is used for DAG workflows
	Inputs *values.Map

	// Payload is used for no DAG workflows
	Payload *anypb.Any

	// ConfigPayload is used for no DAG workflows
	ConfigPayload *anypb.Any

	// The method to call for no DAG workflows
	Method       string
	CapabilityId string
}

// ParseID parses a capability ID in form of: `{name}:{label1_key}_{labe1_value}:{label2_key}_{label2_value}@{version}`
func ParseID(id string) (name string, labels iter.Seq2[string, string], version string) {
	if i := strings.LastIndex(id, "@"); i != -1 {
		version = id[i+1:]
		id = id[:i]
	}
	if parts := strings.Split(id, ":"); len(parts) >= 1 {
		name = parts[0]
		labels = func(yield func(string, string) bool) {
			for _, label := range parts[1:] {
				kv := strings.SplitN(label, "_", 2)
				var v string
				if len(kv) == 2 {
					v = kv[1]
				}
				if !yield(kv[0], v) {
					return
				}
			}
		}
	}
	return
}

// ChainSelectorLabel returns a chain selector value from the labels if one is present.
// It supports both a normal key/value pair, and sequential keys for historical reasons.
func ChainSelectorLabel(labels iter.Seq2[string, string]) (*uint64, error) {
	const key = "ChainSelector"
	var next bool
	for k, v := range labels {
		if next {
			cs, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid chain selector: %s", v)
			}
			return &cs, nil
		}
		if k == key {
			if v != "" {
				cs, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid chain selector: %s", v)
				}
				return &cs, nil
			} else {
				// empty value means it will be in the next key
				next = true
			}
		}
	}
	return nil, nil
}

type RegisterToWorkflowRequest struct {
	Metadata RegistrationMetadata
	Config   *values.Map
}

type UnregisterFromWorkflowRequest struct {
	Metadata RegistrationMetadata
	Config   *values.Map
}

// Executable is an interface for executing a capability.
type Executable interface {
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

	// Config for DAG workflows
	Config *values.Map

	// Request body for no DAG workflows
	Payload *anypb.Any
	// The method to call for no DAG workflows
	Method string
}

type TriggerResponse struct {
	Event TriggerEvent
	Err   error
}

type TriggerEvent struct {
	// The ID of the trigger capability
	TriggerType string
	// The ID of the trigger event
	ID string
	// Trigger-specific payload for DAG workflows
	Outputs *values.Map

	// Trigger-specific payload for no DAG workflows
	Payload *anypb.Any

	// Deprecated: use Outputs instead
	// TODO: remove after core services are updated (pending https://github.com/smartcontractkit/chainlink/pull/16950)
	OCREvent *OCRTriggerEvent
}

type OCRTriggerEvent struct {
	ConfigDigest []byte
	SeqNr        uint64
	Report       []byte // marshaled pb.OCRTriggerReport
	Sigs         []OCRAttributedOnchainSignature
}

// DO NOT change this. it is in the encoding of [TriggerEvent].Outputs
//
// TODO: a more sophisticated way to handle this would be to have add this const
// in the protobuf definition of the TriggerEvent struct.
const ocrTriggerEventOutputKey = "OCRTriggerEvent"

func (e *OCRTriggerEvent) topLevelKey() string {
	return ocrTriggerEventOutputKey
}

// ToMap converts the OCRTriggerEvent to a map.
// This is useful serialization purposes with the [TriggerEvent] struct.
func (e *OCRTriggerEvent) ToMap() (*values.Map, error) {
	x, err := values.Wrap(e)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap OCRTriggerEvent: %w", err)
	}
	return values.NewMap(map[string]any{
		e.topLevelKey(): x,
	})
}

// FromMap converts a map to an OCRTriggerEvent.
// This is useful deserialization purposes with the [TriggerEvent] struct.
func (e *OCRTriggerEvent) FromMap(m *values.Map) error {
	if m == nil {
		return fmt.Errorf("nil map")
	}
	if m.Underlying == nil {
		return fmt.Errorf("nil underlying map")
	}
	val, ok := m.Underlying[e.topLevelKey()]
	if !ok {
		return fmt.Errorf("missing key: %s", e.topLevelKey())
	}
	var unwrapped OCRTriggerEvent
	err := val.UnwrapTo(&unwrapped)
	if err != nil {
		return fmt.Errorf("failed to unwrap OCRTriggerEvent: %w", err)
	}
	*e = unwrapped
	return nil
}

type OCRAttributedOnchainSignature struct {
	Signature []byte
	Signer    uint32 // oracle ID (0,1,...,N-1)
}

type TriggerExecutable interface {
	RegisterTrigger(ctx context.Context, request TriggerRegistrationRequest) (<-chan TriggerResponse, error)
	UnregisterTrigger(ctx context.Context, request TriggerRegistrationRequest) error
	AckEvent(ctx context.Context, triggerId string, eventId string, method string) error
}

// TriggerCapability interface needs to be implemented by all trigger capabilities.
type TriggerCapability interface {
	BaseCapability
	TriggerExecutable
}

// ExecutableCapability is the interface implemented by action, consensus and target
// capabilities. This interface is useful when trying to capture capabilities of varying types.
type ExecutableCapability interface {
	BaseCapability
	Executable
}

// Deprecated: use ExecutableCapability instead.
type ActionCapability = ExecutableCapability

// Deprecated: use ExecutableCapability instead.
type ConsensusCapability = ExecutableCapability

// Deprecated: use ExecutableCapability instead.
type TargetCapability = ExecutableCapability

type ExecutableAndTriggerCapability interface {
	TriggerCapability
	ExecutableCapability
}

type DONWithNodes struct {
	DON   DON
	Nodes []Node
}

// DON represents a network of connected nodes.
//
// For an example of an empty DON check, see the following link:
// https://github.com/smartcontractkit/chainlink/blob/develop/core/capabilities/transmission/local_target_capability.go#L31
type DON struct {
	Name             string
	ID               uint32
	Families         []string
	ConfigVersion    uint32
	Members          []p2ptypes.PeerID
	F                uint8
	IsPublic         bool
	AcceptsWorkflows bool
	Config           []byte
}

// Node contains the node's peer ID and the DONs it is part of.
// The signer is the Node's onchain public key used for OCR signing,
// and the encryption public key is the Node's workflow public key.
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
	PeerID              *p2ptypes.PeerID
	NodeOperatorID      uint32
	Signer              [32]byte
	EncryptionPublicKey [32]byte
	WorkflowDON         DON
	CapabilityDONs      []DON
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
	// SpendTypes denotes the spend types a capability expects to use during an invocation.
	SpendTypes []CapabilitySpendType
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
var idRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-:]+@(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

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
	spendTypes ...CapabilitySpendType,
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

	if spendTypes == nil {
		spendTypes = make([]CapabilitySpendType, 0)
	}

	return CapabilityInfo{
		ID:             id,
		CapabilityType: capabilityType,
		Description:    description,
		DON:            don,
		IsLocal:        isLocal,
		SpendTypes:     spendTypes,
	}, nil
}

// NewCapabilityInfo returns a new CapabilityInfo.
func NewCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
	spendTypes ...CapabilitySpendType,
) (CapabilityInfo, error) {
	return newCapabilityInfo(id, capabilityType, description, nil, true, spendTypes...)
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
	spendTypes ...CapabilitySpendType,
) (CapabilityInfo, error) {
	return newCapabilityInfo(id, capabilityType, description, don, false, spendTypes...)
}

// MustNewCapabilityInfo returns a new CapabilityInfo,
// `panic`ing if we could not instantiate a CapabilityInfo.
func MustNewCapabilityInfo(
	id string,
	capabilityType CapabilityType,
	description string,
	spendTypes ...CapabilitySpendType,
) CapabilityInfo {
	c, err := NewCapabilityInfo(id, capabilityType, description, spendTypes...)
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
	spendTypes ...CapabilitySpendType,
) CapabilityInfo {
	c, err := NewRemoteCapabilityInfo(id, capabilityType, description, don, spendTypes...)
	if err != nil {
		panic(err)
	}

	return c
}

const (
	DefaultRegistrationRefresh       = 30 * time.Second
	DefaultRegistrationExpiry        = 2 * time.Minute
	DefaultEventTimeout              = 2 * time.Minute // TODO: determine best value
	DefaultMessageExpiry             = 2 * time.Minute
	DefaultBatchSize                 = 100
	DefaultBatchCollectionPeriod     = 100 * time.Millisecond
	DefaultExecutableRequestTimeout  = 8 * time.Minute
	DefaultServerMaxParallelRequests = uint32(1000)
)

type RemoteTriggerConfig struct {
	RegistrationRefresh     time.Duration
	RegistrationExpiry      time.Duration
	EventTimeout            time.Duration
	MinResponsesToAggregate uint32
	MessageExpiry           time.Duration
	MaxBatchSize            uint32
	BatchCollectionPeriod   time.Duration
}

type RemoteTargetConfig struct { // deprecated - v1 only
	RequestHashExcludedAttributes []string
}

type RemoteExecutableConfig struct {
	RequestHashExcludedAttributes []string // deprecated - v1 only

	// Fields below are used only by v2 capabilities
	TransmissionSchedule      TransmissionSchedule
	DeltaStage                time.Duration
	RequestTimeout            time.Duration
	ServerMaxParallelRequests uint32
	RequestHasherType         RequestHasherType
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
	if c.EventTimeout == 0 {
		c.EventTimeout = DefaultEventTimeout
	}
	if c.MessageExpiry == 0 {
		c.MessageExpiry = DefaultMessageExpiry
	}
	if c.MaxBatchSize == 0 {
		c.MaxBatchSize = DefaultBatchSize
	}
	if c.BatchCollectionPeriod == 0 {
		c.BatchCollectionPeriod = DefaultBatchCollectionPeriod
	}
}

func (c *RemoteExecutableConfig) ApplyDefaults() {
	if c == nil {
		return
	}
	// Default schedule is 0 ("all at once"), default delta stage is 0.
	if c.RequestTimeout == 0 {
		c.RequestTimeout = DefaultExecutableRequestTimeout
	}
	if c.ServerMaxParallelRequests == 0 {
		c.ServerMaxParallelRequests = DefaultServerMaxParallelRequests
	}
	// Hasher type 0 is the default type.
}

type AggregatorType int
type TransmissionSchedule int
type RequestHasherType int

const (
	AggregatorType_Mode         AggregatorType = 0
	AggregatorType_SignedReport AggregatorType = 1

	Schedule_AllAtOnce  TransmissionSchedule = 0
	Schedule_OneAtATime TransmissionSchedule = 1

	RequestHasherType_Simple                       RequestHasherType = 0
	RequestHasherType_WriteReportExcludeSignatures RequestHasherType = 1
)

type AggregatorConfig struct {
	AggregatorType AggregatorType
}

type CapabilityMethodConfig struct {
	RemoteTriggerConfig    *RemoteTriggerConfig
	RemoteExecutableConfig *RemoteExecutableConfig
	AggregatorConfig       *AggregatorConfig
}

type CapabilityConfiguration struct {
	DefaultConfig *values.Map
	// RestrictedKeys is a list of keys that can't be provided by users in their
	// configuration; we'll remove these fields before passing them to the capability.
	RestrictedKeys []string
	// RestrictedConfig is configuration that can only be set by us; this
	// takes precedence over any user-provided config.
	RestrictedConfig       *values.Map
	RemoteTriggerConfig    *RemoteTriggerConfig
	RemoteTargetConfig     *RemoteTargetConfig
	RemoteExecutableConfig *RemoteExecutableConfig

	// v2 / "NoDAG" capabilities
	CapabilityMethodConfig map[string]CapabilityMethodConfig
	// if true, the capability won't be callable via don2don
	LocalOnly bool
}
