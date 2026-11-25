package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKebab(t *testing.T) {
	t.Run("a lone s after an acronym is its plural", func(t *testing.T) {
		for in, want := range map[string]string{
			"OnchainKeyBundleIDs": "onchain-key-bundle-ids",
			"IDsByChain":          "ids-by-chain",
			"PeerURLs":            "peer-urls",
			"APIs":                "apis",
		} {
			assert.Equal(t, want, kebab(in), in)
		}
	})
	t.Run("digits stay with the word before them", func(t *testing.T) {
		for in, want := range map[string]string{
			"V2Bootstrappers": "v2-bootstrappers",
			"OCR3Config":      "ocr3-config",
			"Http2Server":     "http2-server",
			"MaxRetries3":     "max-retries3",
		} {
			assert.Equal(t, want, kebab(in), in)
		}
	})

	t.Run("basic words", func(t *testing.T) {
		for in, want := range map[string]string{
			"TwoWords":   "two-words",
			"HTTPServer": "http-server",
			"ASet":       "a-set",
			"max_size":   "max-size",
		} {
			assert.Equal(t, want, kebab(in), in)
		}
	})
}
