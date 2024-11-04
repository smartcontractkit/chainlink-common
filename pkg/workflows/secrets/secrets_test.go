package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/nacl/box"
)

// Mock data for testing, see JSON in https://gist.github.com/shileiwill/c077b31193f3f1a124bf4b046a464bf5
var (
	encryptionPublicKeys = map[string][32]byte{
		"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": {17, 65, 221, 30, 70, 121, 124, 237, 155, 15, 186, 212, 145, 21, 241, 133, 7, 246, 246, 230, 227, 204, 134, 231, 229, 186, 22, 158, 88, 100, 90, 220},
		"147d5cc651819b093cd2fdff9760f0f0f77b7ef7798d9e24fc6a350b7300e5d9": {65, 45, 198, 254, 72, 234, 78, 52, 186, 170, 119, 218, 46, 59, 3, 45, 57, 185, 56, 89, 123, 111, 61, 97, 254, 126, 209, 131, 168, 39, 164, 49},
		"2934f31f278e5c60618f85861bd6add54a4525d79a642019bdc87d75d26372c3": {40, 185, 17, 67, 236, 145, 17, 121, 106, 125, 99, 225, 76, 28, 246, 187, 1, 180, 237, 89, 102, 122, 181, 79, 91, 199, 46, 190, 73, 200, 129, 190},
		"298834a041a056df58c839cb53d99b78558693042e54dff238f504f16d18d4b6": {72, 121, 1, 224, 192, 169, 211, 198, 110, 124, 252, 80, 243, 169, 227, 205, 191, 223, 27, 1, 7, 39, 61, 115, 217, 74, 145, 210, 120, 84, 85, 22},
		"5f247f61a6d5bfdd1d5064db0bd25fe443648133c6131975edb23481424e3d9c": {122, 22, 111, 188, 129, 110, 180, 164, 220, 182, 32, 209, 28, 60, 202, 197, 192, 133, 213, 107, 25, 114, 55, 65, 0, 17, 111, 135, 97, 157, 235, 184},
		"77224be9d052343b1d17156a1e463625c0d746468d4f5a44cddd452365b1d4ed": {7, 224, 255, 197, 123, 98, 99, 96, 77, 245, 23, 185, 75, 217, 134, 22, 148, 81, 163, 201, 6, 0, 168, 85, 187, 25, 33, 45, 197, 117, 222, 84},
		"adb6bf005cdb23f21e11b82d66b9f62628c2939640ed93028bf0dad3923c5a8b": {64, 59, 114, 240, 177, 179, 181, 245, 169, 27, 207, 237, 183, 242, 133, 153, 118, 117, 2, 160, 75, 91, 126, 6, 127, 207, 55, 130, 226, 62, 235, 156},
		"b96933429b1a81c811e1195389d7733e936b03e8086e75ea1fa92c61564b6c31": {117, 172, 99, 252, 151, 163, 30, 49, 22, 128, 132, 224, 222, 140, 205, 43, 234, 144, 5, 155, 96, 157, 150, 47, 62, 67, 252, 41, 108, 219, 162, 141},
		"d7e9f2252b09edf0802a65b60bc9956691747894cb3ab9fefd072adf742eb9f1": {180, 115, 9, 31, 225, 212, 219, 188, 38, 173, 113, 198, 123, 68, 50, 248, 244, 40, 14, 6, 186, 181, 226, 18, 42, 146, 244, 171, 139, 111, 242, 245},
		"e38c9f2760db006f070e9cc1bc1c2269ad033751adaa85d022fb760cbc5b5ef6": {69, 66, 244, 253, 46, 209, 80, 200, 201, 118, 179, 152, 2, 254, 61, 153, 74, 236, 58, 201, 79, 209, 30, 120, 23, 246, 147, 177, 201, 161, 218, 187},
	}
	secrets = map[string][]string{
		"SECRET_A": {"one", "two", "three", "four"},
		"SECRET_B": {"all"},
	}
	workflowOwner = "0x9ed925d8206a4f88a2f643b28b3035b315753cd6"
	config        = SecretsConfig{
		SecretsNames: map[string][]string{
			"SECRET_A": {"ENV_VAR_A_FOR_NODE_ONE", "ENV_VAR_A_FOR_NODE_TWO", "ENV_VAR_A_FOR_NODE_THREE", "ENV_VAR_A_FOR_NODE_FOUR"},
			"SECRET_B": {"ENV_VAR_B_FOR_ALL_NODES"},
		},
	}
)

func TestEncryptSecretsForNodes(t *testing.T) {
	encryptedSecrets, secretsEnvVarsByNode, err := EncryptSecretsForNodes(workflowOwner, secrets, encryptionPublicKeys, config)
	// Ensure no error occurred
	assert.NoError(t, err)

	// Ensure all p2pKeys are in encryptedSecrets map
	assert.Equal(t, len(encryptionPublicKeys), len(encryptedSecrets))
	for p2pId := range encryptionPublicKeys {
		_, exists := encryptedSecrets[p2pId]
		assert.True(t, exists, "p2pId %s not found in encryptedSecrets", p2pId)
	}

	// In envVarsAssignedToNodes, ensure SECRET_B has ENV_VAR_B_FOR_ALL_NODES for all nodes
	for _, assignedSecrets := range secretsEnvVarsByNode {
		for _, assignedSecret := range assignedSecrets {
			if assignedSecret.WorkflowSecretName == "SECRET_B" {
				assert.Contains(t, assignedSecret.LocalEnvVarName, "ENV_VAR_B_FOR_ALL_NODES")
			}
		}
	}

	// In envVarsAssignedToNodes, ensure ENV_VAR_A_FOR_NODE_ONE and ENV_VAR_A_FOR_NODE_TWO shows up in 3 nodes and others in 2 nodes
	nodeCount := make(map[string]int)

	for _, assignedSecrets := range secretsEnvVarsByNode {
		for _, assignedSecret := range assignedSecrets {
			nodeCount[assignedSecret.LocalEnvVarName]++
		}
	}

	assert.Equal(t, 3, nodeCount["ENV_VAR_A_FOR_NODE_ONE"], "ENV_VAR_A_FOR_NODE_ONE should be assigned to 3 nodes")
	assert.Equal(t, 3, nodeCount["ENV_VAR_A_FOR_NODE_TWO"], "ENV_VAR_A_FOR_NODE_TWO should be assigned to 3 nodes")
	assert.Equal(t, 2, nodeCount["ENV_VAR_A_FOR_NODE_THREE"], "ENV_VAR_A_FOR_NODE_THREE should be assigned to 2 nodes")
	assert.Equal(t, 2, nodeCount["ENV_VAR_A_FOR_NODE_FOUR"], "ENV_VAR_A_FOR_NODE_FOUR should be assigned to 2 nodes")
}

type key struct {
	publicKey  *[32]byte
	privateKey *[32]byte
}

func (k *key) PublicKey() [32]byte {
	return *k.publicKey
}

func (k *key) PublicKeyString() string {
	return base64.StdEncoding.EncodeToString((*k.publicKey)[:])
}

func (k *key) Decrypt(sealedBox []byte) ([]byte, error) {
	b, ok := box.OpenAnonymous(nil, sealedBox, k.publicKey, k.privateKey)
	if !ok {
		return nil, errors.New("failed to decrypt box")
	}

	return b, nil
}

func newKey() (*key, error) {
	pk, sk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &key{publicKey: pk, privateKey: sk}, nil
}

func TestEncryptDecrypt(t *testing.T) {
	k, err := newKey()
	require.NoError(t, err)

	k2, err := newKey()
	require.NoError(t, err)

	expectedSecrets := map[string]string{
		"foo": "fooToken",
		"bar": "barToken",
	}
	secrets := map[string][]string{
		"foo": []string{expectedSecrets["foo"]},
		"bar": []string{expectedSecrets["bar"]},
	}
	encryptionKeys := map[string][32]byte{
		"nodeAPeerID": k.PublicKey(),
		"nodeBPeerID": k2.PublicKey(),
	}
	config := SecretsConfig{
		SecretsNames: map[string][]string{
			"foo": []string{"ENV_FOO"},
			"bar": []string{"ENV_BAR"},
		},
	}

	encryptedSecrets, _, err := EncryptSecretsForNodes(workflowOwner, secrets, encryptionKeys, config)
	require.NoError(t, err)

	result := EncryptedSecretsResult{
		EncryptedSecrets: encryptedSecrets,
		Metadata: Metadata{
			NodePublicEncryptionKeys: map[string]string{
				"nodeAPeerID": k.PublicKeyString(),
				"nodeBPeerID": k2.PublicKeyString(),
			},
		},
	}
	t.Run("success", func(st *testing.T) {
		gotSecrets, err := DecryptSecretsForNode(result, k, workflowOwner)
		require.NoError(st, err)

		assert.Equal(st, expectedSecrets, gotSecrets)

		gotSecrets, err = DecryptSecretsForNode(result, k2, workflowOwner)
		require.NoError(st, err)

		assert.Equal(st, expectedSecrets, gotSecrets)
	})

	t.Run("incorrect owner", func(st *testing.T) {
		_, err = DecryptSecretsForNode(result, k, "wrong owner")
		assert.ErrorContains(t, err, "invalid secrets bundle: got owner")
	})

	t.Run("key not in metadata", func(st *testing.T) {
		overriddenResult := EncryptedSecretsResult{
			EncryptedSecrets: encryptedSecrets,
			Metadata: Metadata{
				NodePublicEncryptionKeys: map[string]string{
					"nodeBPeerID": k2.PublicKeyString(),
				},
			},
		}
		_, err = DecryptSecretsForNode(overriddenResult, k, workflowOwner)
		assert.ErrorContains(t, err, "cannot find public key")
	})

	t.Run("missing secrets blob", func(st *testing.T) {
		overriddenSecrets := map[string]string{
			"nodeAPeerID": encryptedSecrets["nodeAPeerID"],
		}
		overriddenResult := EncryptedSecretsResult{
			EncryptedSecrets: overriddenSecrets,
			Metadata: Metadata{
				NodePublicEncryptionKeys: map[string]string{
					"nodeAPeerID": k.PublicKeyString(),
					"nodeBPeerID": k2.PublicKeyString(),
				},
			},
		}
		_, err = DecryptSecretsForNode(overriddenResult, k2, workflowOwner)
		assert.ErrorContains(t, err, "cannot find secrets blob")
	})

}

func TestValidateEncryptedSecrets(t *testing.T) {
	// Helper function to generate a valid base64 encoded string
	validBase64 := func(input string) string {
		return base64.StdEncoding.EncodeToString([]byte(input))
	}

	// Define a key for testing
	keyFromMetadata := [32]byte{1, 2, 3}

	// Valid JSON input with matching workflow owner
	validInput := map[string]interface{}{
		"encryptedSecrets": map[string]string{
			"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": validBase64("secret1"),
		},
		"metadata": map[string]interface{}{
			"workflowOwner": "correctOwner",
			"nodePublicEncryptionKeys": map[string]string{
				"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": hex.EncodeToString(keyFromMetadata[:]),
			},
		},
	}

	// Serialize the valid input
	validData, _ := json.Marshal(validInput)

	// Define test cases
	tests := []struct {
		name                 string
		inputData            []byte
		encryptionPublicKeys map[string][32]byte
		workflowOwner        string
		shouldError          bool
	}{
		{
			name:          "Valid input",
			inputData:     validData,
			workflowOwner: "correctOwner",
			encryptionPublicKeys: map[string][32]byte{
				"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": {1, 2, 3},
			},
			shouldError: false,
		},
		{
			name:          "Invalid base64 encoded secret",
			inputData:     []byte(`{"encryptedSecrets": {"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": "invalid-base64!"}}`),
			workflowOwner: "correctOwner",
			encryptionPublicKeys: map[string][32]byte{
				"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": {1, 2, 3},
			},
			shouldError: true,
		},
		{
			name:          "Missing public key",
			inputData:     validData,
			workflowOwner: "correctOwner",
			encryptionPublicKeys: map[string][32]byte{
				"some-other-id": {1, 2, 3},
			},
			shouldError: true,
		},
		{
			name:          "Mismatched workflow owner",
			inputData:     validData,
			workflowOwner: "incorrectOwner",
			encryptionPublicKeys: map[string][32]byte{
				"09ca39cd924653c72fbb0e458b629c3efebdad3e29e7cd0b5760754d919ed829": {1, 2, 3},
			},
			shouldError: true,
		},
	}

	// Run the test cases
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateEncryptedSecrets(test.inputData, test.encryptionPublicKeys, test.workflowOwner)
			if (err != nil) != test.shouldError {
				t.Errorf("Expected error: %v, got: %v", test.shouldError, err != nil)
			}
		})
	}
}
