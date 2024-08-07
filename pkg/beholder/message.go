package beholder

import (
	"fmt"
	"reflect"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	otelsdklog "go.opentelemetry.io/otel/sdk/log"
)

type Message struct {
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
	DonID            string
	// The RDD network name the CL node is operating with.
	NetworkName          []string
	WorkflowID           string
	WorkflowName         string
	WorkflowOwnerAddress string
	// Hash of the workflow spec.
	WorkflowSpecID string
	// The unique execution of a workflow.
	WorkflowExecutionID string
	// The address for the contract.
	CapabilityContractAddress string
	CapabilityID              string
	CapabilityVersion         string
	CapabilityName            string
	NetworkChainID            string
}

func (m Metadata) Attributes() Attributes {
	attrs := make(Attributes, reflect.ValueOf(m).NumField())
	attrs["node_version"] = m.NodeVersion
	attrs["node_csa_key"] = m.NodeCsaKey
	attrs["node_csa_signature"] = m.NodeCsaSignature
	attrs["don_id"] = m.DonID
	attrs["network_name"] = m.NetworkName
	attrs["workflow_id"] = m.WorkflowID
	attrs["workflow_name"] = m.WorkflowName
	attrs["workflow_owner_address"] = m.WorkflowOwnerAddress
	attrs["workflow_spec_id"] = m.WorkflowSpecID
	attrs["workflow_execution_id"] = m.WorkflowExecutionID
	attrs["beholder_data_schema"] = m.BeholderDataSchema
	attrs["capability_contract_address"] = m.CapabilityContractAddress
	attrs["capability_id"] = m.CapabilityID
	attrs["capability_version"] = m.CapabilityVersion
	attrs["capability_name"] = m.CapabilityName
	attrs["network_chain_id"] = m.NetworkChainID
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

func NewMessage(body []byte, attrs Attributes) Message {
	return Message{
		Body:  body,
		Attrs: attrs,
	}
}

func (e *Message) AddAttributes(attrs Attributes) {
	if e.Attrs == nil {
		e.Attrs = make(map[string]any, len(attrs))
	}
	for k, v := range attrs {
		e.Attrs[k] = v
	}
}

func (e *Message) AddOtelAttributes(attrs ...attribute.KeyValue) {
	if e.Attrs == nil {
		e.Attrs = make(map[string]any, len(attrs))
	}
	for _, v := range attrs {
		e.Attrs[string(v.Key)] = v.Value
	}
}

func (e *Message) OtelRecord() otellog.Record {
	return newRecord(e.Body, e.Attrs)
}

func (e *Message) SdkOtelRecord() otelsdklog.Record {
	return newSdkRecord(e.Body, e.Attrs)
}

func (e *Message) Copy() Message {
	attrs := make(Attributes, len(e.Attrs))
	for k, v := range e.Attrs {
		attrs[k] = v
	}
	c := Message{
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

func (e Message) String() string {
	return fmt.Sprintf("Message{Attrs: %v, Body: %v}", e.Attrs, e.Body)
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
			m.DonID = v.(string)
		case "network_name":
			m.NetworkName = v.([]string)
		case "workflow_id":
			m.WorkflowID = v.(string)
		case "workflow_name":
			m.WorkflowName = v.(string)
		case "workflow_owner_address":
			m.WorkflowOwnerAddress = v.(string)
		case "workflow_spec_id":
			m.WorkflowSpecID = v.(string)
		case "workflow_execution_id":
			m.WorkflowExecutionID = v.(string)
		case "beholder_data_schema":
			m.BeholderDataSchema = v.(string)
		case "capability_contract_address":
			m.CapabilityContractAddress = v.(string)
		case "capability_id":
			m.CapabilityID = v.(string)
		case "capability_version":
			m.CapabilityVersion = v.(string)
		case "capability_name":
			m.CapabilityName = v.(string)
		case "network_chain_id":
			m.NetworkChainID = v.(string)
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

func (e Message) Validate() error {
	if e.Body == nil {
		return fmt.Errorf("message body is required")
	}
	if len(e.Attrs) == 0 {
		return fmt.Errorf("message attributes are required")
	}
	metadata := NewMetadata(e.Attrs)
	if err := metadata.Validate(); err != nil {
		return err
	}
	return nil
}
