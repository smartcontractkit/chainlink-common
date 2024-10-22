package beholder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
)

type Message struct {
	Attrs Attributes
	Body  []byte
}

type Metadata struct {
	//	REQUIRED FIELDS
	// Schema Registry URI to fetch schema
	BeholderDomain     string `validate:"required,domain_entity"`
	BeholderEntity     string `validate:"required,domain_entity"`
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
	return Attributes{
		"node_version":                m.NodeVersion,
		"node_csa_key":                m.NodeCsaKey,
		"node_csa_signature":          m.NodeCsaSignature,
		"don_id":                      m.DonID,
		"network_name":                m.NetworkName,
		"workflow_id":                 m.WorkflowID,
		"workflow_name":               m.WorkflowName,
		"workflow_owner_address":      m.WorkflowOwnerAddress,
		"workflow_spec_id":            m.WorkflowSpecID,
		"workflow_execution_id":       m.WorkflowExecutionID,
		"beholder_domain":             m.BeholderDomain,
		"beholder_entity":             m.BeholderEntity,
		"beholder_data_schema":        m.BeholderDataSchema,
		"capability_contract_address": m.CapabilityContractAddress,
		"capability_id":               m.CapabilityID,
		"capability_version":          m.CapabilityVersion,
		"capability_name":             m.CapabilityName,
		"network_chain_id":            m.NetworkChainID,
	}
}

type Attributes = map[string]any

func newAttributes(attrKVs ...any) Attributes {
	a := make(Attributes, len(attrKVs)/2)

	l := len(attrKVs)
	for i := 0; i < l; {
		switch t := attrKVs[i].(type) {
		case Attributes:
			for k, v := range t {
				a[k] = v
			}
			i++
		case string:
			if i+1 >= l {
				break
			}
			val := attrKVs[i+1]
			a[t] = val
			i += 2
		default:
			// Unexpected type
			return a
		}
	}
	return a
}

func NewMessage(body []byte, attrKVs ...any) Message {
	return Message{
		Body:  body,
		Attrs: newAttributes(attrKVs...),
	}
}

func (e *Message) AddAttributes(attrKVs ...any) {
	attrs := newAttributes(attrKVs...)
	if e.Attrs == nil {
		e.Attrs = make(map[string]any, len(attrs)/2)
	}
	for k, v := range attrs {
		e.Attrs[k] = v
	}
}

func (e *Message) OtelRecord() otellog.Record {
	return newRecord(e.Body, e.Attrs)
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
		return otellog.String(key, fmt.Sprintf("<unhandled beholder attribute value type: %T, value:%v>", v, v))
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
		case "beholder_domain":
			m.BeholderDomain = v.(string)
		case "beholder_entity":
			m.BeholderEntity = v.(string)
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

// validDomainAndEntityRegex allows for alphanumeric characters and ._-
var validDomainAndEntityRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func NewMetadataValidator() (*validator.Validate, error) {
	validate := validator.New()
	err := validate.RegisterValidation("domain_entity", func(fl validator.FieldLevel) bool {
		str, isStr := fl.Field().Interface().(string)
		if !isStr {
			return false
		}
		if strings.Contains(str, "__") {
			return false
		}
		if !validDomainAndEntityRegex.MatchString(str) {
			return false
		}
		return true
	})
	if err != nil {
		return nil, err
	}
	return validate, nil
}

func (m *Metadata) Validate() error {
	validate, err := NewMetadataValidator()
	if err != nil {
		return err
	}
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
	return metadata.Validate()
}
