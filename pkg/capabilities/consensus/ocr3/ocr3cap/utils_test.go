package ocr3cap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/ocr3cap"
)

func TestEncoders(t *testing.T) {
	t.Parallel()
	t.Run("Returns a copy of the underlying encoders", func(t *testing.T) {
		orig := ocr3cap.Encoders()
		orig[0] = "foo"
		assert.NotEqual(t, orig, ocr3cap.Encoders())
		assert.Equal(t, ocr3cap.Encoder("foo"), orig[0])
	})
}
