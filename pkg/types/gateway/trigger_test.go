package gateway

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPTriggerRequest_MarshalJSON_Deterministic(t *testing.T) {
	tests := []struct {
		name     string
		request  HTTPTriggerRequest
		expected string
	}{
		{
			name: "simple input with sorted keys",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowOwner: "owner1",
					WorkflowName:  "workflow1",
				},
				Input: json.RawMessage(`{
					"zebra": "value", 
					"alpha": "value"
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "pubkey1",
				},
			},
			expected: `{"input":{"alpha":"value","zebra":"value"},"key":{"keyType":"ecdsa","publicKey":"pubkey1"},"workflow":{"workflowName":"workflow1","workflowOwner":"owner1"}}`,
		},
		{
			name: "nested object with sorted keys",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowID: "id1",
				},
				Input: json.RawMessage(`{
					"params": {
						"z_param": 1, 
						"a_param": 2
					}, 
					"config": {
						"y_setting": true, 
						"x_setting": false
					}
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "pubkey2",
				},
			},
			expected: `{"input":{"config":{"x_setting":false,"y_setting":true},"params":{"a_param":2,"z_param":1}},"key":{"keyType":"ecdsa","publicKey":"pubkey2"},"workflow":{"workflowID":"id1"}}`,
		},
		{
			name: "array with nested objects",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowTag: "v1.0.0",
				},
				Input: json.RawMessage(`{
					"items": [
						{
							"z_field": "value1", 
							"a_field": "value2"
						}, 
						{
							"y_field": "value3", 
							"b_field": "value4"
						}
					]
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "pubkey3",
				},
			},
			expected: `{"input":{"items":[{"a_field":"value2","z_field":"value1"},{"b_field":"value4","y_field":"value3"}]},"key":{"keyType":"ecdsa","publicKey":"pubkey3"},"workflow":{"workflowTag":"v1.0.0"}}`,
		},
		{
			name: "empty input",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: nil,
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "pubkey5",
				},
			},
			expected: `{"key":{"keyType":"ecdsa","publicKey":"pubkey5"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "deeply nested structure",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowOwner: "deep",
					WorkflowName:  "nested",
				},
				Input: json.RawMessage(`{
					"level1": {
						"z_level2": {
							"c_level3": {
								"z_final": "value", 
								"a_final": "value"
							}
						}, 
						"a_level2": {
							"b_level3": "simple"
						}
					}
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "pubkey6",
				},
			},
			expected: `{"input":{"level1":{"a_level2":{"b_level3":"simple"},"z_level2":{"c_level3":{"a_final":"value","z_final":"value"}}}},"key":{"keyType":"ecdsa","publicKey":"pubkey6"},"workflow":{"workflowName":"nested","workflowOwner":"deep"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestHTTPTriggerRequest_MarshalJSON_ComplexNesting(t *testing.T) {
	tests := []struct {
		name     string
		request  HTTPTriggerRequest
		expected string
	}{
		{
			name: "nested maps in arrays in maps",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "complex1"},
				Input: json.RawMessage(`{
					"z_section": {
						"items": [
							{
								"z_item": {
									"nested_z": 1, 
									"nested_a": 2
								}, 
								"a_item": "value"
							},
							{
								"y_item": {
									"nested_y": 3, 
									"nested_b": 4
								}, 
								"b_item": "value"
							}
						]
					},
					"a_section": {
						"data": [
							{
								"z_data": {
									"deep_z": "val1", 
									"deep_a": "val2"
								}, 
								"a_data": "simple"
							}
						]
					}
				}`),
				Key: AuthorizedKey{KeyType: "ecdsa", PublicKey: "key1"},
			},
			expected: `{"input":{"a_section":{"data":[{"a_data":"simple","z_data":{"deep_a":"val2","deep_z":"val1"}}]},"z_section":{"items":[{"a_item":"value","z_item":{"nested_a":2,"nested_z":1}},{"b_item":"value","y_item":{"nested_b":4,"nested_y":3}}]}},"key":{"keyType":"ecdsa","publicKey":"key1"},"workflow":{"workflowName":"complex1"}}`,
		},
		{
			name: "nested arrays in maps in arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "complex2"},
				Input: json.RawMessage(`{
					"collections": [
						{
							"z_collection": {
								"items": [
									[
										{
											"z_nested": "val1", 
											"a_nested": "val2"
										}
									],
									[
										{
											"y_nested": "val3", 
											"b_nested": "val4"
										}
									]
								]
							},
							"a_collection": {
								"items": [
									[
										{
											"x_nested": "val5", 
											"c_nested": "val6"
										}
									]
								]
							}
						}
					]
				}`),
				Key: AuthorizedKey{KeyType: "ecdsa", PublicKey: "key2"},
			},
			expected: `{"input":{"collections":[{"a_collection":{"items":[[{"c_nested":"val6","x_nested":"val5"}]]},"z_collection":{"items":[[{"a_nested":"val2","z_nested":"val1"}],[{"b_nested":"val4","y_nested":"val3"}]]}}]},"key":{"keyType":"ecdsa","publicKey":"key2"},"workflow":{"workflowName":"complex2"}}`,
		},
		{
			name: "mixed deep nesting with all combinations",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "ultimate"},
				Input: json.RawMessage(`{
					"z_root": {
						"z_array": [
							{
								"z_map": {
									"z_nested_array": [
										{
											"z_deep_map": {
												"z_final": "value",
												"a_final": "value"
											},
											"a_deep_map": {
												"z_final2": "value2",
												"a_final2": "value2"
											}
										}
									],
									"a_nested_array": [
										{"z_item": "val", "a_item": "val"}
									]
								},
								"a_map": "simple"
							}
						],
						"a_array": [
							{
								"nested": {
									"z_key": "value",
									"a_key": "value"
								}
							}
						]
					},
					"a_root": {
						"simple": "value"
					}
				}`),
				Key: AuthorizedKey{KeyType: "ecdsa", PublicKey: "ultimate_key"},
			},
			expected: `{"input":{"a_root":{"simple":"value"},"z_root":{"a_array":[{"nested":{"a_key":"value","z_key":"value"}}],"z_array":[{"a_map":"simple","z_map":{"a_nested_array":[{"a_item":"val","z_item":"val"}],"z_nested_array":[{"a_deep_map":{"a_final2":"value2","z_final2":"value2"},"z_deep_map":{"a_final":"value","z_final":"value"}}]}}]}},"key":{"keyType":"ecdsa","publicKey":"ultimate_key"},"workflow":{"workflowName":"ultimate"}}`,
		},
		{
			name: "arrays of arrays with nested maps",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "matrix"},
				Input: json.RawMessage(`{
					"matrix": [
						[
							[{"z_cell": "val1", "a_cell": "val2"}],
							[{"y_cell": "val3", "b_cell": "val4"}]
						],
						[
							[{"x_cell": "val5", "c_cell": "val6"}]
						]
					]
				}`),
				Key: AuthorizedKey{KeyType: "ecdsa", PublicKey: "matrix_key"},
			},
			expected: `{"input":{"matrix":[[[{"a_cell":"val2","z_cell":"val1"}],[{"b_cell":"val4","y_cell":"val3"}]],[[{"c_cell":"val6","x_cell":"val5"}]]]},"key":{"keyType":"ecdsa","publicKey":"matrix_key"},"workflow":{"workflowName":"matrix"}}`,
		},
		{
			name: "real-world complex structure",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowOwner: "org",
					WorkflowName:  "payment-processor",
					WorkflowTag:   "v2.1.0",
				},
				Input: json.RawMessage(`{
					"transaction": {
						"z_metadata": {
							"z_tags": ["urgent", "enterprise"],
							"a_tags": ["verified"],
							"z_properties": {
								"z_source": "api",
								"a_source": "web",
								"risk_score": 0.2
							},
							"a_properties": {
								"z_region": "us-east",
								"a_region": "eu-west"
							}
						},
						"a_metadata": {
							"timestamp": "2023-01-01T00:00:00Z"
						},
						"payment_methods": [
							{
								"z_type": "credit_card",
								"z_details": {
									"z_last_four": "1234",
									"a_last_four": "5678",
									"z_network": "visa",
									"a_network": "mastercard"
								},
								"a_details": {
									"encrypted": true
								},
								"a_type": "bank_transfer"
							}
						],
						"amount": {
							"z_currency": "USD",
							"a_currency": "EUR",
							"value": 100.50
						}
					},
					"a_transaction": {
						"simple": "fallback"
					}
				}`),
				Key: AuthorizedKey{KeyType: "ecdsa", PublicKey: "payment_key_12345"},
			},
			expected: `{"input":{"a_transaction":{"simple":"fallback"},"transaction":{"a_metadata":{"timestamp":"2023-01-01T00:00:00Z"},"amount":{"a_currency":"EUR","value":100.5,"z_currency":"USD"},"payment_methods":[{"a_details":{"encrypted":true},"a_type":"bank_transfer","z_details":{"a_last_four":"5678","a_network":"mastercard","z_last_four":"1234","z_network":"visa"},"z_type":"credit_card"}],"z_metadata":{"a_properties":{"a_region":"eu-west","z_region":"us-east"},"a_tags":["verified"],"z_properties":{"a_source":"web","risk_score":0.2,"z_source":"api"},"z_tags":["urgent","enterprise"]}}},"key":{"keyType":"ecdsa","publicKey":"payment_key_12345"},"workflow":{"workflowName":"payment-processor","workflowOwner":"org","workflowTag":"v2.1.0"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestHTTPTriggerRequest_MarshalJSON_Consistency(t *testing.T) {
	// Test that multiple marshalling attempts produce identical results
	request := HTTPTriggerRequest{
		Workflow: WorkflowSelector{
			WorkflowOwner: "owner",
			WorkflowName:  "test",
			WorkflowTag:   "v1.0",
			WorkflowID:    "123",
		},
		Input: json.RawMessage(`{
			"z_param": {
				"nested_z": 1, 
				"nested_a": 2
			}, 
			"a_param": [
				{
					"item_z": 3, 
					"item_a": 4
				}
			]
		}`),
		Key: AuthorizedKey{
			KeyType:   "ecdsa",
			PublicKey: "consistent_key",
		},
	}

	// Marshal multiple times
	results := make([]string, 10)
	for i := range 10 {
		data, err := json.Marshal(request)
		require.NoError(t, err)
		results[i] = string(data)
	}

	// All results should be identical
	for i := 1; i < len(results); i++ {
		require.Equal(t, results[0], results[i], "Marshall attempt %d should match first attempt", i)
	}
}

func TestHTTPTriggerRequest_MarshalJSON_SliceHandling(t *testing.T) {
	tests := []struct {
		name     string
		request  HTTPTriggerRequest
		expected string
	}{
		{
			name: "simple array of primitives",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{
					"numbers": [3, 1, 2], 
					"strings": ["zebra", "alpha", "beta"]
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key1",
				},
			},
			expected: `{"input":{"numbers":[3,1,2],"strings":["zebra","alpha","beta"]},"key":{"keyType":"ecdsa","publicKey":"key1"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "nested arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{"matrix": [[1, 2], [3, 4]], "nested": [["a", "b"], ["c", "d"]]}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key2",
				},
			},
			expected: `{"input":{"matrix":[[1,2],[3,4]],"nested":[["a","b"],["c","d"]]},"key":{"keyType":"ecdsa","publicKey":"key2"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "array with mixed objects and primitives",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{
					"mixed": [
						{
							"z_key": "value", 
							"a_key": "value"
						}, 
						"primitive", 
						42, 
						{
							"nested": {
								"z_nested": 1, 
								"a_nested": 2
							}
						}
					]
				}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key3",
				},
			},
			expected: `{"input":{"mixed":[{"a_key":"value","z_key":"value"},"primitive",42,{"nested":{"a_nested":2,"z_nested":1}}]},"key":{"keyType":"ecdsa","publicKey":"key3"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "deeply nested array with objects",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{"deep": [[[{"z_deep": "value", "a_deep": "value"}]]]}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key4",
				},
			},
			expected: `{"input":{"deep":[[[{"a_deep":"value","z_deep":"value"}]]]},"key":{"keyType":"ecdsa","publicKey":"key4"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "array of objects with nested arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{"items": [{"z_field": [{"z_sub": 1, "a_sub": 2}], "a_field": "value"}]}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key5",
				},
			},
			expected: `{"input":{"items":[{"a_field":"value","z_field":[{"a_sub":2,"z_sub":1}]}]},"key":{"keyType":"ecdsa","publicKey":"key5"},"workflow":{"workflowName":"test"}}`,
		},
		{
			name: "empty arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName: "test",
				},
				Input: json.RawMessage(`{"empty": [], "nested_empty": [[]]}`),
				Key: AuthorizedKey{
					KeyType:   "ecdsa",
					PublicKey: "key6",
				},
			},
			expected: `{"input":{"empty":[],"nested_empty":[[]]},"key":{"keyType":"ecdsa","publicKey":"key6"},"workflow":{"workflowName":"test"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestHTTPTriggerRequest_MarshalJSON_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		request  HTTPTriggerRequest
		expected string
	}{
		{
			name: "empty objects in arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "edge1"},
				Input:    json.RawMessage(`{"items": [{}, {"a": 1}, {}]}`),
				Key:      AuthorizedKey{KeyType: "ecdsa", PublicKey: "edge_key1"},
			},
			expected: `{"input":{"items":[{},{"a":1},{}]},"key":{"keyType":"ecdsa","publicKey":"edge_key1"},"workflow":{"workflowName":"edge1"}}`,
		},
		{
			name: "objects with empty arrays",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "edge3"},
				Input:    json.RawMessage(`{"z_section": {"items": []}, "a_section": {"items": []}}`),
				Key:      AuthorizedKey{KeyType: "ecdsa", PublicKey: "edge_key3"},
			},
			expected: `{"input":{"a_section":{"items":[]},"z_section":{"items":[]}},"key":{"keyType":"ecdsa","publicKey":"edge_key3"},"workflow":{"workflowName":"edge3"}}`,
		},
		{
			name: "numeric keys get sorted as strings",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "edge4"},
				Input:    json.RawMessage(`{"z_nums": {"3": "third", "1": "first", "2": "second", "10": "tenth"}}`),
				Key:      AuthorizedKey{KeyType: "ecdsa", PublicKey: "edge_key4"},
			},
			expected: `{"input":{"z_nums":{"1":"first","10":"tenth","2":"second","3":"third"}},"key":{"keyType":"ecdsa","publicKey":"edge_key4"},"workflow":{"workflowName":"edge4"}}`,
		},
		{
			name: "special characters in keys",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{WorkflowName: "edge5"},
				Input:    json.RawMessage(`{"z-key": "value", "a_key": "value", "z.key": "value", "a key": "value"}`),
				Key:      AuthorizedKey{KeyType: "ecdsa", PublicKey: "edge_key5"},
			},
			expected: `{"input":{"a key":"value","a_key":"value","z-key":"value","z.key":"value"},"key":{"keyType":"ecdsa","publicKey":"edge_key5"},"workflow":{"workflowName":"edge5"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.request)
			require.NoError(t, err)
			require.Equal(t, tt.expected, string(result))
		})
	}
}

func TestHTTPTriggerRequest_MarshalJSON_InvalidInput(t *testing.T) {
	request := HTTPTriggerRequest{
		Workflow: WorkflowSelector{
			WorkflowName: "test",
		},
		Input: json.RawMessage(`{"invalid": json}`), // Invalid JSON
		Key: AuthorizedKey{
			KeyType:   "ecdsa",
			PublicKey: "key",
		},
	}

	_, err := json.Marshal(request)
	require.Error(t, err, "Should error on invalid JSON in Input field")
}

func TestHTTPTriggerRequest_MarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		request HTTPTriggerRequest
	}{
		{
			name:    "empty request",
			request: HTTPTriggerRequest{},
		},
		{
			name: "minimal workflow with ID only",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowID: "test-workflow-123",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0x1234567890abcdef",
				},
			},
		},
		{
			name: "workflow with all fields",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowID:    "test-workflow-456",
					WorkflowName:  "test-workflow",
					WorkflowOwner: "test-owner",
					WorkflowTag:   "v1.0.0",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0xabcdef1234567890",
				},
			},
		},
		{
			name: "workflow with name and owner only",
			request: HTTPTriggerRequest{
				Workflow: WorkflowSelector{
					WorkflowName:  "my-workflow",
					WorkflowOwner: "alice",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0x9876543210fedcba",
				},
			},
		},
		{
			name: "simple input object",
			request: HTTPTriggerRequest{
				Input: json.RawMessage(`{"param1": "value1", "param2": 42}`),
				Workflow: WorkflowSelector{
					WorkflowID: "workflow-with-input",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0xdeadbeefcafebabe",
				},
			},
		},
		{
			name: "complex nested input",
			request: HTTPTriggerRequest{
				Input: json.RawMessage(`{
					"config": {
						"timeout": 300,
						"retries": 3,
						"endpoints": ["http://api1.com", "http://api2.com"]
					},
					"data": {
						"user": {
							"id": 123,
							"name": "John Doe",
							"permissions": ["read", "write"]
						},
						"metadata": {
							"version": "2.1.0",
							"timestamp": "2024-01-01T00:00:00Z"
						}
					}
				}`),
				Workflow: WorkflowSelector{
					WorkflowName:  "complex-workflow",
					WorkflowOwner: "system",
					WorkflowTag:   "production",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0x1111222233334444",
				},
			},
		},
		{
			name: "input with array",
			request: HTTPTriggerRequest{
				Input: json.RawMessage(`{
					"items": [
						{"id": 1, "name": "item1"},
						{"id": 2, "name": "item2"},
						{"id": 3, "name": "item3"}
					],
					"count": 3
				}`),
				Workflow: WorkflowSelector{
					WorkflowID:   "array-processor",
					WorkflowName: "process-items",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0x5555666677778888",
				},
			},
		},
		{
			name: "empty input object",
			request: HTTPTriggerRequest{
				Input: json.RawMessage(`{}`),
				Workflow: WorkflowSelector{
					WorkflowOwner: "test-user",
					WorkflowTag:   "latest",
				},
				Key: AuthorizedKey{
					KeyType:   KeyTypeECDSAEVM,
					PublicKey: "0x9999aaaabbbbcccc",
				},
			},
		},
		{
			name: "different key type",
			request: HTTPTriggerRequest{
				Input: json.RawMessage(`{"test": true}`),
				Workflow: WorkflowSelector{
					WorkflowID: "different-key-workflow",
				},
				Key: AuthorizedKey{
					KeyType:   "rsa", // Different key type
					PublicKey: "rsa-public-key-data",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			require.NoError(t, err, "marshalling should not fail")
			var unmarshaled HTTPTriggerRequest
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err, "unmarshalling should not fail")
			require.Equal(t, tt.request.Workflow, unmarshaled.Workflow, "workflow selector should be preserved")
			require.Equal(t, tt.request.Key, unmarshaled.Key, "authorized key should be preserved")
			if len(tt.request.Input) == 0 {
				require.Empty(t, unmarshaled.Input, "empty input should remain empty")
			} else {
				// Remarshal unmarshaled.Input and compare to original string
				remarshaled, err := json.Marshal(unmarshaled.Input)
				require.NoError(t, err, "remarshalling unmarshaled input should not fail")
				require.JSONEq(t, string(tt.request.Input), string(remarshaled), "input JSON string should be preserved after marshal/unmarshal")
			}
		})
	}
}
