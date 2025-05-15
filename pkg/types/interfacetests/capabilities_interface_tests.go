package interfacetests

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

var (
	// It takes the form of `{name}:{label1_key}_{labe1_value}:{label2_key}_{label2_value}@{version}`
	validCapabilityID = regexp.MustCompile(`(?P<name>.+):(?P<label1_key>.+)_(?P<label1_value>.+):(?P<label2_key>.+)_(?P<label2_value>.+)@(?P<version>.+)`)
)

type CapabilityTestID string

const (
	CapabilityResponseContainsMeteringData = "capability response contains exactly one value for metering data"
	CapabilityTypeIsValid                  = "capability type is valid"
	CapabilityInfoIDCorrectPattern         = "capability info has correct pattern"
)

type CapabilitiesInterfaceTester interface {
	Name() string
	Setup(*testing.T, CapabilityTestID)
	IsDisabled(CapabilityTestID) bool
}

type TriggerCapabilitiesInterfaceTester interface {
	CapabilitiesInterfaceTester
	GetTriggerCapability(*testing.T, CapabilityTestID) capabilities.TriggerCapability
}

type ExecutableCapabilitiesInterfaceTester interface {
	CapabilitiesInterfaceTester
	GetExecutableCapability(*testing.T, CapabilityTestID) capabilities.ExecutableCapability
}

func RunTriggerCapabilityInterfaceTests(t *testing.T, tester TriggerCapabilitiesInterfaceTester) {
	t.Run(tester.Name(), func(t *testing.T) {
		t.Parallel()

		runBaseCapabilityInterfaceTests(t, tester, func(t *testing.T, id CapabilityTestID) capabilities.BaseCapability {
			return tester.GetTriggerCapability(t, id)
		})
	})
}

func RunExecutableCapabilityInterfaceTests(t *testing.T, tester ExecutableCapabilitiesInterfaceTester) {
	t.Run(tester.Name(), func(t *testing.T) {
		t.Parallel()

		runBaseCapabilityInterfaceTests(t, tester, func(t *testing.T, id CapabilityTestID) capabilities.BaseCapability {
			return tester.GetExecutableCapability(t, id)
		})

		t.Run("Executable", func(t *testing.T) {
			t.Parallel()

			tests := []Testcase[*testing.T]{
				{
					Name: CapabilityResponseContainsMeteringData,
					Test: func(t *testing.T) {
						capability := tester.GetExecutableCapability(t, CapabilityResponseContainsMeteringData)
						response, err := capability.Execute(t.Context(), capabilities.CapabilityRequest{})

						require.NoError(t, err)

						assert.Len(t, response.Metadata.Metering, 1)
					},
				},
			}

			for _, test := range tests {
				testID := CapabilityTestID(test.Name)

				if !tester.IsDisabled(testID) {
					t.Run(test.Name, func(t *testing.T) {
						t.Parallel()

						tester.Setup(t, testID)
						test.Test(t)
					})
				}
			}
		})
	})
}

type baseCapabilityFunc func(*testing.T, CapabilityTestID) capabilities.BaseCapability

func runBaseCapabilityInterfaceTests(t *testing.T, tester CapabilitiesInterfaceTester, base baseCapabilityFunc) {
	t.Run("BaseCapability", func(t *testing.T) {
		t.Parallel()
		tests := []Testcase[*testing.T]{
			{
				Name: CapabilityInfoIDCorrectPattern,
				Test: func(t *testing.T) {
					capability := base(t, CapabilityInfoIDCorrectPattern)
					info, err := capability.Info(t.Context())

					require.NoError(t, err)
					assert.True(t, validCapabilityID.MatchString(info.ID))
				},
			},
			{
				Name: CapabilityTypeIsValid,
				Test: func(t *testing.T) {
					capability := base(t, CapabilityTypeIsValid)

					var info capabilities.CapabilityInfo

					info, err := capability.Info(t.Context())
					require.NoError(t, err)

					assert.NoError(t, info.CapabilityType.IsValid())
				},
			},
		}

		for _, test := range tests {
			testID := CapabilityTestID(test.Name)

			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()

				tester.Setup(t, testID)
				test.Test(t)
			})
		}
	})
}
