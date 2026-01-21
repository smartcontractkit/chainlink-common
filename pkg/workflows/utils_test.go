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

func Test_GenerateWorkflowOwnerAddress_OtherTestVectors(t *testing.T) {
	testVectors := []struct {
		prefix      string
		ownerKey    string
		keccakInput string
		keccakHash  string
		ethAddress  string
		description string
	}{
		{
			prefix:      "",
			ownerKey:    "",
			ethAddress:  "0x41305062a5e522A01B7D9460E6744C879113C5dB",
			description: "Empty prefix and ownerKey",
		},
		{
			prefix:      "",
			ownerKey:    "non_empty_owner_key",
			ethAddress:  "0xa6b6C2C0D58bD45c83b89F60cae395F7f3c9A0D0",
			description: "Empty prefix, non-empty ownerKey",
		},
		{
			prefix:      "non_empty_prefix",
			ownerKey:    "",
			ethAddress:  "0x8Cb3243107cAD0D5584cD8b37393FE7f5200B920",
			description: "Non-empty prefix, empty ownerKey",
		},
		{
			prefix:      "x",
			ownerKey:    "yz",
			ethAddress:  "0xb2a39e39664A469bc1d1b5dB8592deda4E9410af",
			description: "Collision test case 1: x + yz",
		},
		{
			prefix:      "xy",
			ownerKey:    "z",
			ethAddress:  "0x561960c471b8B288284073457EB77175C49DA9cd",
			description: "Collision test case 2: xy + z (should be different from x + yz)",
		},
		{
			prefix:      "org_123456789",
			ownerKey:    "cre-storage-service",
			ethAddress:  "0x95b028290D5E2aC912f0bA8e9E35931B90740608",
			description: "Realistic org ID and service name",
		},
		{
			prefix:      "org_0x1234567890abcdef",
			ownerKey:    "cre-storage-service",
			ethAddress:  "0x102d1d8570F155D4C9AF463C7098B978fEaEDf1C",
			description: "Hex-prefixed org ID",
		},
		{
			prefix:      "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ownerKey:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ethAddress:  "0x31B1AfA30824ab78fF1c3f7b26ef6ce50D9c3221",
			description: "Long repeating characters",
		},
		{
			prefix:      "org_public_key",
			ownerKey:    "0x02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ethAddress:  "0x9Ed0D44aB4B2FCC1B4C280105D641F668b81A6e9",
			description: "Public key as ownerKey",
		},
		{
			prefix:      "org@special#chars",
			ownerKey:    "key$with%symbols",
			ethAddress:  "0xAbF538FFffFAE8C139c0BF80e7cEC5c1596D0944",
			description: "Special characters in both prefix and ownerKey",
		},
	}

	for _, tv := range testVectors {
		t.Run(tv.description, func(t *testing.T) {
			addr, err := GenerateWorkflowOwnerAddress(tv.prefix, tv.ownerKey)
			require.NoError(t, err, "Failed to generate address")

			expectedAddr := tv.ethAddress
			if expectedAddr[:2] == "0x" {
				expectedAddr = expectedAddr[2:]
			}
			expectedAddr = strings.ToLower(expectedAddr)
			actualAddr := hex.EncodeToString(addr)

			require.Equal(t, expectedAddr, actualAddr, "Address mismatch")

			require.Len(t, addr, 20, "Address should be 20 bytes")
		})
	}
}

func Test_GenerateWorkflowOwnerAddress_CollisionResistanceWithLengthPrefixing(t *testing.T) {
	// Test that the length-prefixing prevents collisions
	testCases := []struct {
		prefix1  string
		owner1   string
		prefix2  string
		owner2   string
		caseName string
	}{
		{
			prefix1:  "x",
			owner1:   "yz",
			prefix2:  "xy",
			owner2:   "z",
			caseName: "x+yz vs xy+z collision resistance",
		},
		{
			prefix1:  "a",
			owner1:   "bc",
			prefix2:  "ab",
			owner2:   "c",
			caseName: "a+bc vs ab+c collision resistance",
		},
		{
			prefix1:  "",
			owner1:   "abc",
			prefix2:  "a",
			owner2:   "bc",
			caseName: "empty+abc vs a+bc collision resistance",
		},
		{
			prefix1:  "test",
			owner1:   "",
			prefix2:  "tes",
			owner2:   "t",
			caseName: "test+empty vs tes+t collision resistance",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			addr1, err := GenerateWorkflowOwnerAddress(tc.prefix1, tc.owner1)
			require.NoError(t, err)

			addr2, err := GenerateWorkflowOwnerAddress(tc.prefix2, tc.owner2)
			require.NoError(t, err)

			require.NotEqual(t, hex.EncodeToString(addr1), hex.EncodeToString(addr2),
				"Addresses should be different")
		})
	}
}
