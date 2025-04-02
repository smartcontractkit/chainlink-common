package cmd_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/usercode/pkg"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/testdata/fixtures/usercode/pkg2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-user-types -dir ./testdata/fixtures/usercode/pkg -skip_cap time.Time
//go:generate go run github.com/smartcontractkit/chainlink-common/pkg/capabilities/cli/cmd/generate-user-types -dir ./testdata/fixtures/usercode/pkg2 -types OtherPackage

func TestGenerateUserTypes(t *testing.T) {
	t.Parallel()

	t.Run("generated types work as expected", func(t *testing.T) {
		onlyVerifySyntax(func() {
			myVal := pkg.ConstantMyType(pkg.MyType{I: 10})
			// verify both types were generated from different files
			pkg.ConstantMyType2(pkg.MyType2{I: 10})

			var tmp sdk.CapDefinition[pkg.MyType] = myVal // nolint
			_ = tmp

			other := pkg2.ConstantOtherPackage(pkg2.OtherPackage{X: "x", Z: "z"}) //nolint
			other = myVal.O()                                                     // nolint
			_ = other

			var s sdk.CapDefinition[string] = myVal.S() // nolint
			_ = s
		})
	})

	t.Run("specifying types to generate ignores other types", func(t *testing.T) {
		content, err := os.ReadFile("./testdata/fixtures/usercode/pkg2/wrappers_generated.go")
		require.NoError(t, err)

		require.False(t, strings.Contains(string(content), "NotWrappedCap"))
	})

	t.Run("Wrapping wrapped type is no-op", func(t *testing.T) {
		original := pkg.NewMyTypeFromFields(
			sdk.ConstantDefinition(1),
			pkg.ConstantMyNestedType(pkg.MyNestedType{}),
			pkg2.ConstantOtherPackage(pkg2.OtherPackage{}),
			sdk.ConstantDefinition(""),
			sdk.ConstantDefinition(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
		)

		wrapped := pkg.MyTypeWrapper(original)
		require.Same(t, original, wrapped)
	})
}
