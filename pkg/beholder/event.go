package beholder

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	otelsdklog "go.opentelemetry.io/otel/sdk/log"
)

type Event struct {
	Attrs map[string]any
	Body  []byte
}

type Metadata struct {
	//	REQUIRED FIELDS
	// Schema Registry URI to fetch schema
	BeholderDataSchema string `validate:"required,uri"`

	// OPTIONAL FIELDS
	// The version of the CL node.
	NodeVersion string
	// mTLS public key for the node operator. This is used as an identity key but with the added benefit of being able to provide signatures.
	NodeCsaKey string
	// Signature from CSA private key.
	NodeCsaSignature string
	DonId            string
	// The RDD network name the CL node is operating with.
	NetworkName          []string
	WorkflowId           string
	WorkflowName         string
	WorkflowOwnerAddress string
	// Hash of the workflow spec.
	WorkflowSpecId string
	// The unique execution of a workflow.
	WorkflowExecutionId string
	// The address for the contract.
	CapabilityContractAddress string
	CapabilityId              string
	CapabilityVersion         string
	CapabilityName            string
	NetworkChainId            string
}

func (m Metadata) Attributes() Attributes {
	attrs := make(Attributes, reflect.ValueOf(m).NumField())
	attrs["node_version"] = m.NodeVersion
	attrs["node_csa_key"] = m.NodeCsaKey
	attrs["node_csa_signature"] = m.NodeCsaSignature
	attrs["don_id"] = m.DonId
	attrs["network_name"] = m.NetworkName
	attrs["workflow_id"] = m.WorkflowId
	attrs["workflow_name"] = m.WorkflowName
	attrs["workflow_owner_address"] = m.WorkflowOwnerAddress
	attrs["workflow_spec_id"] = m.WorkflowSpecId
	attrs["workflow_execution_id"] = m.WorkflowExecutionId
	attrs["beholder_data_schema"] = m.BeholderDataSchema
	attrs["capability_contract_address"] = m.CapabilityContractAddress
	attrs["capability_id"] = m.CapabilityId
	attrs["capability_version"] = m.CapabilityVersion
	attrs["capability_name"] = m.CapabilityName
	attrs["network_chain_id"] = m.NetworkChainId
	return attrs
}

type Attributes map[string]any

func NewAttributes(args ...any) Attributes {
	attrs := make(Attributes, len(args)/2)
	attrs.Add(args...)
	return attrs
}

func (a Attributes) Add(args ...any) Attributes {
	for i := 1; i < len(args); i += 2 {
		if key, ok := args[i-1].(string); ok {
			val := args[i]
			a[key] = val
		}
	}
	return a
}

func NewEvent(body []byte, attrs Attributes) Event {
	return Event{
		Body:  body,
		Attrs: attrs,
	}
}

func (e *Event) AddAttributes(attrs Attributes) {
	if e.Attrs == nil {
		e.Attrs = make(map[string]any, len(attrs))
	}
	for k, v := range attrs {
		e.Attrs[k] = v
	}
}

func (e *Event) AddOtelAttributes(attrs ...attribute.KeyValue) {
	if e.Attrs == nil {
		e.Attrs = make(map[string]any, len(attrs))
	}
	for _, v := range attrs {
		e.Attrs[string(v.Key)] = v.Value
	}
}

func (e *Event) OtelRecord() otellog.Record {
	return newRecord(e.Body, e.Attrs)
}

func (e *Event) SdkOtelRecord() otelsdklog.Record {
	return newSdkRecord(e.Body, e.Attrs)
}

func (e *Event) Copy() Event {
	attrs := make(Attributes, len(e.Attrs))
	for k, v := range e.Attrs {
		attrs[k] = v
	}
	c := Event{
		Attrs: attrs,
	}
	if e.Body != nil {
		c.Body = make([]byte, len(e.Body))
		copy(c.Body, e.Body)
	}
	return c
}

// Creates otellog.Record from body and attributes
func newRecord(body []byte, attrs map[string]any) otellog.Record {
	otelRecord := otellog.Record{}
	if body != nil {
		otelRecord.SetBody(otellog.BytesValue(body))
	}
	for k, v := range attrs {
		otelRecord.AddAttributes(OtelAttr(k, v))
	}
	return otelRecord
}

// Creates otelsdklog.Record from body and attributes
// NOTE: internal function otelsdklog.newRecord returns value not pointer
func newSdkRecord(body []byte, attrs map[string]any) otelsdklog.Record {
	sdkRecord := otelsdklog.Record{}
	if body != nil {
		sdkRecord.SetBody(otellog.BytesValue(body))
	}
	for k, v := range attrs {
		sdkRecord.AddAttributes(OtelAttr(k, v))
	}
	return sdkRecord
}

func OtelAttr(key string, value any) otellog.KeyValue {
	switch v := value.(type) {
	case string:
		return otellog.String(key, v)
	case []string:
		vals := make([]otellog.Value, 0, len(v))
		for _, s := range v {
			vals = append(vals, otellog.StringValue(s))
		}
		return otellog.Slice(key, vals...)
	case int64:
		return otellog.Int64(key, v)
	case int:
		return otellog.Int(key, v)
	case float64:
		return otellog.Float64(key, v)
	case bool:
		return otellog.Bool(key, v)
	case []byte:
		return otellog.Bytes(key, v)
	case nil:
		return otellog.Empty(key)
	case otellog.Value:
		return otellog.KeyValue{Key: key, Value: v}
	case attribute.Value:
		return OtelAttr(key, v.AsInterface())
	default:
		return otellog.String(key, fmt.Sprintf("<unhandled beholder attribute value type: %T>", v))
	}
}

func (e Event) String() string {
	return fmt.Sprintf("Event{Attrs: %v, Body: %v}", e.Attrs, e.Body)
}

// Sets metadata fields from  attributes
func (m *Metadata) FromAttributes(attrs Attributes) *Metadata {
	for k, v := range attrs {
		switch k {
		case "node_version":
			m.NodeVersion = v.(string)
		case "node_csa_key":
			m.NodeCsaKey = v.(string)
		case "node_csa_signature":
			m.NodeCsaSignature = v.(string)
		case "don_id":
			m.DonId = v.(string)
		case "network_name":
			m.NetworkName = v.([]string)
		case "workflow_id":
			m.WorkflowId = v.(string)
		case "workflow_name":
			m.WorkflowName = v.(string)
		case "workflow_owner_address":
			m.WorkflowOwnerAddress = v.(string)
		case "workflow_spec_id":
			m.WorkflowSpecId = v.(string)
		case "workflow_execution_id":
			m.WorkflowExecutionId = v.(string)
		case "beholder_data_schema":
			m.BeholderDataSchema = v.(string)
		case "capability_contract_address":
			m.CapabilityContractAddress = v.(string)
		case "capability_id":
			m.CapabilityId = v.(string)
		case "capability_version":
			m.CapabilityVersion = v.(string)
		case "capability_name":
			m.CapabilityName = v.(string)
		case "network_chain_id":
			m.NetworkChainId = v.(string)
		}
	}
	return m
}

func NewMetadata(attrs Attributes) *Metadata {
	m := &Metadata{}
	m.FromAttributes(attrs)
	return m
}

func (m *Metadata) Validate() error {
	validate := validator.New()
	return validate.Struct(m)
}

func (e Event) Validate() error {
	if e.Body == nil {
		return fmt.Errorf("event body is required")
	}
	if len(e.Attrs) == 0 {
		return fmt.Errorf("event attributes are required")
	}
	metadata := NewMetadata(e.Attrs)
	if err := metadata.Validate(); err != nil {
		return err
	}
	return nil
}
