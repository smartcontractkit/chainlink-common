package interfacetests

import (
	"regexp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

var (
	validCapabilityID = regexp.MustCompile(`{name}:{label1_key}_{labe1_value}:{label2_key}_{label2_value}@{version}`)
)

type CapabilityTestID string

const (
	CapabilityResponseContainsMeteringData = "capability response contains exactly one value for metering data"
	CapabilityTypeIsValid                  = "capability type is valid"
	CapabilityInfoIDCorrectPattern         = "capability info has correct pattern"
)

type CapabilitiesInterfaceTester[T TestingT[T]] interface {
	Name() string
	Setup(T, CapabilityTestID)
	IsDisabled(CapabilityTestID) bool
}

type TriggerCapabilitiesInterfaceTester[T TestingT[T]] interface {
	CapabilitiesInterfaceTester[T]
	GetTriggerCapability(T, CapabilityTestID) capabilities.TriggerCapability
}

type ExecutableCapabilitiesInterfaceTester[T TestingT[T]] interface {
	CapabilitiesInterfaceTester[T]
	GetExecutableCapability(T, CapabilityTestID) capabilities.ExecutableCapability
}

func RunTriggerCapabilityInterfaceTests[T TestingT[T]](t T, tester TriggerCapabilitiesInterfaceTester[T]) {
	t.Run(tester.Name(), func(t T) {
		t.Parallel()

		runBaseCapabilityInterfaceTests(t, tester, func(t T, id CapabilityTestID) capabilities.BaseCapability { return tester.GetTriggerCapability(t, id) })
	})
}

func RunExecutableCapabilityInterfaceTests[T TestingT[T]](t T, tester ExecutableCapabilitiesInterfaceTester[T]) {
	t.Run(tester.Name(), func(t T) {
		t.Parallel()

		runBaseCapabilityInterfaceTests(t, tester, func(t T, id CapabilityTestID) capabilities.BaseCapability {
			return tester.GetExecutableCapability(t, id)
		})

		t.Run("Executable", func(t T) {
			t.Parallel()

			tests := []Testcase[T]{
				{
					Name: CapabilityResponseContainsMeteringData,
					Test: func(t T) {
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
					t.Run(test.Name, func(t T) {
						t.Parallel()

						tester.Setup(t, testID)
						test.Test(t)
					})
				}
			}
		})
	})
}

type baseCapabilityFunc[T TestingT[T]] func(T, CapabilityTestID) capabilities.BaseCapability

func runBaseCapabilityInterfaceTests[T TestingT[T]](t T, tester CapabilitiesInterfaceTester[T], base baseCapabilityFunc[T]) {
	t.Run("BaseCapability", func(t T) {
		t.Parallel()
		tests := []Testcase[T]{
			{
				Name: CapabilityInfoIDCorrectPattern,
				Test: func(t T) {
					capability := base(t, CapabilityInfoIDCorrectPattern)
					info, err := capability.Info(t.Context())

					require.NoError(t, err)
					assert.True(t, validCapabilityID.MatchString(info.ID))
				},
			},
			{
				Name: CapabilityTypeIsValid,
				Test: func(t T) {
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

			t.Run(test.Name, func(t T) {
				t.Parallel()

				tester.Setup(t, testID)
				test.Test(t)
			})
		}
	})
}
