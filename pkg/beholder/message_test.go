package beholder_test

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func ExampleMessage() {
	// Create message with body and attributes
	m1 := beholder.NewMessage([]byte{1}, beholder.Attributes{"key_string": "value"})
	fmt.Println("#1", m1)
	// Create attributes
	additionalAttributes := beholder.Attributes{
		"key_string": "new value",
		"key_int32":  int32(1),
	}
	// Add more attributes
	additionalAttributes.Add(
		"key_string", "updated value", // this will overrider previous value
		"key_int32", int32(2),
		"key3", true,
	)
	// Add attributes to message
	m1.AddAttributes(additionalAttributes)
	fmt.Println("#2", m1)
	// Create mmpty message struct
	m2 := beholder.Message{}
	fmt.Println("#3", m2)
	// Add attributes to message
	m2.AddAttributes(beholder.Attributes{"key_int": 1})
	fmt.Println("#4", m2)
	// Update attribute key_int
	m2.AddAttributes(beholder.Attributes{"key_int": 2})
	fmt.Println("#5", m2)
	// Set message body
	m2.Body = []byte("0123")
	fmt.Println("#6", m2)
	// Reset attributes
	m2.Attrs = beholder.Attributes{}
	fmt.Println("#7", m2)
	// Reset body
	m2.Body = nil
	fmt.Println("#8", m2)
	// Shalow copy of message
	m3 := beholder.NewMessage(m1.Body, m1.Attrs)
	fmt.Println("#9", m3)
	m1.Body[0] = byte(2) // Wil mutate m3
	fmt.Println("#10", m3)
	// Deep copy
	m4 := m1.Copy()
	fmt.Println("#11", m4)
	m1.Body[0] = byte(3) // Should not mutate m4
	fmt.Println("#12", m4)
	// Create message with mixed attributes: kv pairs and maps
	m5 := beholder.NewMessage([]byte{1},
		// Add attributes from the map
		map[string]any{
			"key1": "value1",
		},
		// Add attributes from KV pair
		"key2", "value2",
		// Add attributes from Attributes map
		beholder.Attributes{"key3": "value3"},
		// Add attributes from KV pair
		"key4", "value4",
		// Modify key1
		"key1", "value5",
		// Modify key2
		map[string]any{
			"key2": "value6",
		},
	)
	fmt.Println("#13", m5)
	// Output:
	// #1 Message{Attrs: map[key_string:value], Body: [1]}
	// #2 Message{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [1]}
	// #3 Message{Attrs: map[], Body: []}
	// #4 Message{Attrs: map[key_int:1], Body: []}
	// #5 Message{Attrs: map[key_int:2], Body: []}
	// #6 Message{Attrs: map[key_int:2], Body: [48 49 50 51]}
	// #7 Message{Attrs: map[], Body: [48 49 50 51]}
	// #8 Message{Attrs: map[], Body: []}
	// #9 Message{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [1]}
	// #10 Message{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
	// #11 Message{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
	// #12 Message{Attrs: map[key3:true key_int32:2 key_string:updated value], Body: [2]}
	// #13 Message{Attrs: map[key1:value5 key2:value6 key3:value3 key4:value4], Body: [1]}
}

func testMetadata() beholder.Metadata {
	return beholder.Metadata{
		NodeVersion:               "v1.0.0",
		NodeCsaKey:                "test_key",
		NodeCsaSignature:          "test_signature",
		DonID:                     "test_don_id",
		NetworkName:               []string{"test_network"},
		WorkflowID:                "test_workflow_id",
		WorkflowName:              "test_workflow_name",
		WorkflowOwnerAddress:      "test_owner_address",
		WorkflowSpecID:            "test_spec_id",
		WorkflowExecutionID:       "test_execution_id",
		BeholderDataSchema:        "/schemas/ids/test_schema", // required field, URI
		CapabilityContractAddress: "test_contract_address",
		CapabilityID:              "test_capability_id",
		CapabilityVersion:         "test_capability_version",
		CapabilityName:            "test_capability_name",
		NetworkChainID:            "test_chain_id",
	}
}
func ExampleMetadata() {
	m := testMetadata()
	fmt.Printf("%#v\n", m)
	fmt.Println(m.Attributes())
	// Output:
	// beholder.Metadata{BeholderDataSchema:"/schemas/ids/test_schema", NodeVersion:"v1.0.0", NodeCsaKey:"test_key", NodeCsaSignature:"test_signature", DonID:"test_don_id", NetworkName:[]string{"test_network"}, WorkflowID:"test_workflow_id", WorkflowName:"test_workflow_name", WorkflowOwnerAddress:"test_owner_address", WorkflowSpecID:"test_spec_id", WorkflowExecutionID:"test_execution_id", CapabilityContractAddress:"test_contract_address", CapabilityID:"test_capability_id", CapabilityVersion:"test_capability_version", CapabilityName:"test_capability_name", NetworkChainID:"test_chain_id"}
	// map[beholder_data_schema:/schemas/ids/test_schema capability_contract_address:test_contract_address capability_id:test_capability_id capability_name:test_capability_name capability_version:test_capability_version don_id:test_don_id network_chain_id:test_chain_id network_name:[test_network] node_csa_key:test_key node_csa_signature:test_signature node_version:v1.0.0 workflow_execution_id:test_execution_id workflow_id:test_workflow_id workflow_name:test_workflow_name workflow_owner_address:test_owner_address workflow_spec_id:test_spec_id]
}

func ExampleValidate() {
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
