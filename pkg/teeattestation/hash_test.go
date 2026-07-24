package teeattestation

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomainHash(t *testing.T) {
	tag := "TestTag"
	data := []byte(`{"key":"value"}`)

	got, err := DomainHash(tag, data)
	require.NoError(t, err)

	h := sha256.New()
	h.Write([]byte(DomainSeparator))
	writeWithLength(h, []byte(tag))
	writeWithLength(h, data)
	want := h.Sum(nil)

	if len(got) != sha256.Size {
		t.Fatalf("expected %d bytes, got %d", sha256.Size, len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("hash mismatch at byte %d: want %x, got %x", i, want, got)
		}
	}
}

func TestDomainHash_DifferentTags(t *testing.T) {
	data := []byte("same-data")
	h1, err := DomainHash("Tag1", data)
	require.NoError(t, err)
	h2, err := DomainHash("Tag2", data)
	require.NoError(t, err)

	for i := range h1 {
		if h1[i] != h2[i] {
			return
		}
	}
	t.Fatal("different tags should produce different hashes")
}

func TestDomainHash_DifferentData(t *testing.T) {
	tag := "SameTag"
	h1, err := DomainHash(tag, []byte("data-a"))
	require.NoError(t, err)
	h2, err := DomainHash(tag, []byte("data-b"))
	require.NoError(t, err)

	for i := range h1 {
		if h1[i] != h2[i] {
			return
		}
	}
	t.Fatal("different data should produce different hashes")
}

func TestDomainHash_InvalidTag(t *testing.T) {
	for _, tag := range []string{"", "A\nB", "tag-1"} {
		_, err := DomainHash(tag, []byte("C"))
		require.EqualError(t, err, "invalid tag: must be non-empty and contain only alphanumeric characters")
	}
}

// CL112-14 regression: without length-prefixing, ("AB","C") and ("A","BC") could collide.
func TestDomainHash_NoBoundaryCollision(t *testing.T) {
	h1, err := DomainHash("AB", []byte("C"))
	require.NoError(t, err)
	h2, err := DomainHash("A", []byte("BC"))
	require.NoError(t, err)
	require.NotEqual(t, h1, h2)
}
