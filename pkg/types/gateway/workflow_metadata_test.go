package gateway

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorkflowMetadata_Digest(t *testing.T) {
	t.Run("deterministic - same input produces same digest", func(t *testing.T) {
		metadata := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{
				WorkflowID:    "workflow-456",
				WorkflowName:  "deterministic-test",
				WorkflowOwner: "0xTestOwner",
			},
			AuthorizedKeys: []AuthorizedKey{
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xkey1"},
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xkey2"},
			},
		}

		digest1, err := metadata.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest1)
		require.Len(t, digest1, 64)

		digest2, err := metadata.Digest()
		require.NoError(t, err)

		require.Equal(t, digest1, digest2, "Multiple calls should produce identical digests")
	})

	t.Run("array ordering does not affect digest - fixes duplicate workflow ID bug", func(t *testing.T) {
		metadata1 := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{WorkflowID: "workflow-789"},
			AuthorizedKeys: []AuthorizedKey{
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xAAAA"},
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xBBBB"},
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xCCCC"},
			},
		}

		metadata2 := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{WorkflowID: "workflow-789"},
			AuthorizedKeys: []AuthorizedKey{
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xBBBB"},
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xCCCC"},
				{KeyType: KeyTypeECDSAEVM, PublicKey: "0xAAAA"},
			},
		}

		digest1, err := metadata1.Digest()
		require.NoError(t, err)

		digest2, err := metadata2.Digest()
		require.NoError(t, err)

		require.Equal(t, digest1, digest2,
			"Different key order should produce identical digests")
	})

	t.Run("different metadata produces different digests", func(t *testing.T) {
		metadata1 := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{WorkflowID: "workflow-abc"},
			AuthorizedKeys:   []AuthorizedKey{{KeyType: KeyTypeECDSAEVM, PublicKey: "0xkey1"}},
		}

		metadata2 := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{WorkflowID: "workflow-xyz"},
			AuthorizedKeys:   []AuthorizedKey{{KeyType: KeyTypeECDSAEVM, PublicKey: "0xkey1"}},
		}

		digest1, err := metadata1.Digest()
		require.NoError(t, err)

		digest2, err := metadata2.Digest()
		require.NoError(t, err)

		require.NotEqual(t, digest1, digest2,
			"Different metadata should produce different digests")
	})

	t.Run("empty authorized keys", func(t *testing.T) {
		metadata := WorkflowMetadata{
			WorkflowSelector: WorkflowSelector{WorkflowID: "workflow-empty"},
			AuthorizedKeys:   []AuthorizedKey{},
		}

		digest, err := metadata.Digest()
		require.NoError(t, err)
		require.NotEmpty(t, digest)
		require.Len(t, digest, 64)
	})
}
