package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_EncodeExecutionID(t *testing.T) {
	var (
		workflowID = "workflowID"
		eventID    = "eventID"
		s          = sha256.New()
	)

	_, err := s.Write([]byte(workflowID))
	assert.NoError(t, err)

	_, err = s.Write([]byte(eventID))
	assert.NoError(t, err)

	expected := hex.EncodeToString(s.Sum(nil))
	actual, err := EncodeExecutionID(workflowID, eventID)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)

	// Test ordering
	s.Reset()
	_, err = s.Write([]byte(eventID))
	assert.NoError(t, err)

	_, err = s.Write([]byte(workflowID))
	assert.NoError(t, err)

	reversed := hex.EncodeToString(s.Sum(nil))
	assert.NotEqual(t, reversed, actual)
}

func Test_GenerateWorkflowIDFromStrings(t *testing.T) {
	// With prefix
	owner := "0x26729408f179371be6433b9585d8427f121bfe82"
	got, err := GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Always starts with the version byte
	assert.Equal(t, got[:2], hex.EncodeToString([]byte{versionByte}))

	// Without prefix
	owner = "26729408f179371be6433b9585d8427f121bfe82"
	got, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// Very short; empty but with a prefix
	owner = "0x"
	got, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	require.NoError(t, err)
	assert.NotNil(t, got)

	owner = "invalid"
	_, err = GenerateWorkflowIDFromStrings(owner, "porporpore", []byte("workflow"), []byte("config"), "http://mysecrets.com")
	assert.ErrorContains(t, err, "encoding/hex")
}

func TestNormalizeWorkflowName(t *testing.T) {
	tt := []struct {
		input    string
		expected string
	}{
		{
			input:    "Hello, world!",
			expected: "315f5bdb76",
		},
		{
			input:    "My Incredible Workflow Name",
			expected: "84002eb9e2",
		},
		{
			input:    "You either die a hero, or live long enough to see yourself become the villain.",
			expected: "6ba1f7a6a0",
		},
		{
			input:    "ï¿½ï¿½ï¿½ï¿½ï¿½ï¿ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿½ï¿",
			expected: "68f19173b3",
		},
	}

	for _, tc := range tt {
		t.Run(tc.input, func(t *testing.T) {
			// Call the function with the test input
			result := HashTruncateName(tc.input)
			var resultBytes = []byte(result)

			// Assert that the result is exactly the expected output
			require.Equal(t, tc.expected, result)

			// Assert that the result is 10 bytes long
			require.Len(t, resultBytes, 10)

			// Assert that the result is UTF8 encoded
			require.True(t, utf8.Valid(resultBytes))
		})
	}
}

func Test_GenerateWorkflowOwnerAddress_DifferentInputsGenerateDifferentAddresses(t *testing.T) {
	tests := []struct {
		name        string
		prefix1     string
		ownerKey1   string
		prefix2     string
		ownerKey2   string
		description string
	}{
		{
			name:        "different_prefix",
			prefix1:     "registry_v1",
			ownerKey1:   "owner123",
			prefix2:     "registry_v2",
			ownerKey2:   "owner123",
			description: "Same ownerKey, different prefix",
		},
		{
			name:        "different_ownerKey",
			prefix1:     "registry_v1",
			ownerKey1:   "owner123",
			prefix2:     "registry_v1",
			ownerKey2:   "owner124",
			description: "Same prefix, different ownerKey",
		},
		{
			name:        "case_sensitive_prefix",
			prefix1:     "Registry",
			ownerKey1:   "owner123",
			prefix2:     "registry",
			ownerKey2:   "owner123",
			description: "Case sensitive prefix difference",
		},
		{
			name:        "case_sensitive_ownerKey",
			prefix1:     "registry",
			ownerKey1:   "Owner123",
			prefix2:     "registry",
			ownerKey2:   "owner123",
			description: "Case sensitive ownerKey difference",
		},
		{
			name:        "single_char_difference_prefix",
			prefix1:     "registrya",
			ownerKey1:   "owner123",
			prefix2:     "registryb",
			ownerKey2:   "owner123",
			description: "Single character difference in prefix",
		},
		{
			name:        "single_char_difference_ownerKey",
			prefix1:     "registry",
			ownerKey1:   "owner123a",
			prefix2:     "registry",
			ownerKey2:   "owner123b",
			description: "Single character difference in ownerKey",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate first address
			addr1, err := GenerateWorkflowOwnerAddress(tt.prefix1, tt.ownerKey1)
			require.NoError(t, err, "Failed to generate first address")

			// Generate second address
			addr2, err := GenerateWorkflowOwnerAddress(tt.prefix2, tt.ownerKey2)
			require.NoError(t, err, "Failed to generate second address")

			// Verify addresses are different
			require.NotEqual(t, hex.EncodeToString(addr1), hex.EncodeToString(addr2), "Addresses should not match")

			// Verify addresses are 20 bytes (Ethereum address length)
			require.Len(t, addr1, 20, "First address should be 20 bytes")
			require.Len(t, addr2, 20, "Second address should be 20 bytes")
		})
	}
}

func Test_GenerateWorkflowOwnerAddress_SameInputsGenerateSameAddress(t *testing.T) {
	prefix := "test_registry"
	ownerKey := "test_owner_123"

	// Generate address multiple times
	addr1, err := GenerateWorkflowOwnerAddress(prefix, ownerKey)
	require.NoError(t, err, "Failed to generate first address")

	addr2, err := GenerateWorkflowOwnerAddress(prefix, ownerKey)
	require.NoError(t, err, "Failed to generate second address")

	// Verify all addresses are identical
	addr1Hex := hex.EncodeToString(addr1)
	addr2Hex := hex.EncodeToString(addr2)

	require.Equal(t, addr1Hex, addr2Hex, "Same inputs should generate same address")
}

func Test_GenerateWorkflowOwnerAddress_SolidityCompatibility(t *testing.T) {
	/*
		// SPDX-License-Identifier: MIT
		pragma solidity ^0.8.0;

		contract WorkflowOwnerAddressGenerator {

		    function generateWorkflowOwnerAddress(
		        string memory prefix,
		        string memory ownerKey
		    ) public pure returns (address) {
		        // Step 1: Create nested hash of prefix + ownerKey
		        bytes32 nestedHash = keccak256(abi.encodePacked(prefix, ownerKey));

		        // Step 2: Create the full preimage for outer hash
		        // 0xff + 84 zero bytes + nested hash
		        bytes memory preimage = new bytes(117); // 1 + 84 + 32 = 117 bytes

		        // Set first byte to 0xff
		        preimage[0] = 0xff;

		        // Bytes 1-84 are already zero (default in Solidity)

		        // Copy nested hash to bytes 85-116
		        for (uint256 i = 0; i < 32; i++) {
		            preimage[85 + i] = nestedHash[i];
		        }

		        // Step 3: Hash the full preimage and return last 20 bytes as address
		        bytes32 outerHash = keccak256(preimage);
		        return address(uint160(uint256(outerHash)));
		    }
		}
	*/
	// These expected addresses were generated using the Solidity contract above
	// You can verify these by deploying the contract and calling the function
	testCases := []struct {
		prefix      string
		ownerKey    string
		expectedHex string // This should be generated by running the Solidity contract
	}{
		{
			prefix:      "registry1",
			ownerKey:    "owner123",
			expectedHex: "0x58c0e4aaf5fb13fcaea5790f8a19014ad9646da3", // convert to lowercase, not checksum
		},
		{
			prefix:      "registry2",
			ownerKey:    "owner123",
			expectedHex: "0xf094995741cffc6c173fa9edb2e8d766d1524039", // convert to lowercase, not checksum
		},
		{
			prefix:      "registry2",
			ownerKey:    "ownerSomethingElse",
			expectedHex: "0x4be6a8e38aa493cac0aa4c6dd13bad41f8219f0c", // convert to lowercase, not checksum
		},
	}

	for _, tc := range testCases {
		t.Run(tc.prefix+"_"+tc.ownerKey, func(t *testing.T) {
			goAddr, err := GenerateWorkflowOwnerAddress(tc.prefix, tc.ownerKey)
			require.NoError(t, err)

			goAddrHex := hex.EncodeToString(goAddr)

			// Remove 0x prefix if present in expected
			expected := strings.TrimPrefix(tc.expectedHex, "0x")

			require.Equal(t, expected, goAddrHex,
				"Go implementation should match Solidity for prefix='%s', ownerKey='%s'",
				tc.prefix, tc.ownerKey)
		})
	}
}
