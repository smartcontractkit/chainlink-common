package beholder_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func ExampleEvent() {
	// Create event with body and attributes
	e1 := beholder.NewEvent([]byte{1}, beholder.Attributes{"key_string": "value"})
	fmt.Println("#1", e1)
	// Create attributes
	additional_attributes := beholder.Attributes{
		"key_string": "new value",
		"key_int32":  int32(1),
	}
	// Add more attributes
	additional_attributes.Add(
		"key_string", "updated value", // this will overrider previous value
		"key_int32", int32(2),
		"key3", true,
	)
	// Add attributes to event
	e1.AddAttributes(additional_attributes)
	fmt.Println("#2", e1)
	// Create empty event struct
	e2 := beholder.Event{}
	fmt.Println("#3", e2)
	// Add attributes to event
	e2.AddAttributes(beholder.Attributes{"key_int": 1})
	fmt.Println("#4", e2)
	// Update attribute key_int
	e2.AddAttributes(beholder.Attributes{"key_int": 2})
	fmt.Println("#5", e2)
	// Set event body
	e2.Body = []byte("0123")
	fmt.Println("#6", e2)
	// Reset attributes
	e2.Attrs = beholder.Attributes{}
	fmt.Println("#7", e2)
	// Reset body
	e2.Body = nil
	fmt.Println("#8", e2)
	// Shalow copy of event
	e3 := beholder.NewEvent(e1.Body, e1.Attrs)
	fmt.Println("#9", e3)
	e1.Body[0] = byte(2) // Wil mutate e3
	fmt.Println("#10", e3)
	// Deep copy
	e4 := e1.Copy()
	fmt.Println("#11", e4)
	e1.Body[0] = byte(3) // Should not mutate e4
	fmt.Println("#12", e4)
	// Output:
	// #1 Event{Attrs: map[key_string:value], Body: [1]}
	// #2 Event{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [1]}
	// #3 Event{Attrs: map[], Body: []}
	// #4 Event{Attrs: map[key_int:1], Body: []}
	// #5 Event{Attrs: map[key_int:2], Body: []}
	// #6 Event{Attrs: map[key_int:2], Body: [48 49 50 51]}
	// #7 Event{Attrs: map[], Body: [48 49 50 51]}
	// #8 Event{Attrs: map[], Body: []}
	// #9 Event{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [1]}
	// #10 Event{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
	// #11 Event{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
	// #12 Event{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
}

func testMetadata() beholder.Metadata {
	return beholder.Metadata{
		NodeVersion:               "v1.0.0",
		NodeCsaKey:                "test_key",
		NodeCsaSignature:          "test_signature",
		DonId:                     "test_don_id",
		NetworkName:               []string{"test_network"},
		WorkflowId:                "test_workflow_id",
		WorkflowName:              "test_workflow_name",
		WorkflowOwnerAddress:      "test_owner_address",
		WorkflowSpecId:            "test_spec_id",
		WorkflowExecutionId:       "test_execution_id",
		BeholderDataSchema:        "/schemas/ids/test_schema", // required field, URI
		CapabilityContractAddress: "test_contract_address",
		CapabilityId:              "test_capability_id",
		CapabilityVersion:         "test_capability_version",
		CapabilityName:            "test_capability_name",
		NetworkChainId:            "test_chain_id",
	}
}
func ExampleMetadata() {
	m := testMetadata()
	fmt.Println(m)
	fmt.Println(m.Attributes())
	// Output:
	// {/schemas/ids/test_schema v1.0.0 test_key test_signature test_don_id [test_network] test_workflow_id test_workflow_name test_owner_address test_spec_id test_execution_id test_contract_address test_capability_id test_capability_version test_capability_name test_chain_id}
	// map[beholder_data_schema:/schemas/ids/test_schema capability_contract_address:test_contract_address capability_id:test_capability_id capability_name:test_capability_name capability_version:test_capability_version don_id:test_don_id network_chain_id:test_chain_id network_name:[test_network] node_csa_key:test_key node_csa_signature:test_signature node_version:v1.0.0 workflow_execution_id:test_execution_id workflow_id:test_workflow_id workflow_name:test_workflow_name workflow_owner_address:test_owner_address workflow_spec_id:test_spec_id]
}

func ExampleMetadataValidate() {
	validate := validator.New()

	metadata := beholder.Metadata{}
	if err := validate.Struct(metadata); err != nil {
		fmt.Println(err)
	}
	metadata.BeholderDataSchema = "example.proto"
	if err := validate.Struct(metadata); err != nil {
		fmt.Println(err)
	}
	metadata.BeholderDataSchema = "/schemas/ids/test_schema"
	if err := validate.Struct(metadata); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Metadata is valid")
	}
	// Output:
	// Key: 'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'required' tag
	// Key: 'Metadata.BeholderDataSchema' Error:Field validation for 'BeholderDataSchema' failed on the 'uri' tag
	// Metadata is valid
}

func TestAttributesConversion(t *testing.T) {
	expected := testMetadata()
	attrs := expected.Attributes()
	actual := beholder.NewMetadata(attrs)
	assert.Equal(t, expected, *actual)
}
