package sdk

import v2 "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"

// MaxSpendLimits represents a collection of spending limits that can be applied
// to a workflow or individual steps
type MaxSpendLimits struct {
	limits *v2.SpendLimits
}

// NewMaxSpendLimits creates a new MaxSpendLimits instance
func NewMaxSpendLimits() *MaxSpendLimits {
	return &MaxSpendLimits{
		limits: v2.NewSpendLimits(),
	}
}

// WithMaxSpend adds a spending limit to the collection
func (m *MaxSpendLimits) WithMaxSpend(credit string, value int64) *MaxSpendLimits {
	m.limits.WithMaxSpend(credit, value)
	return m
}

// WithMaxSpendCRE adds a CRE credit limit to the collection
func (m *MaxSpendLimits) WithMaxSpendCRE(value int64) *MaxSpendLimits {
	m.limits.WithMaxSpendCRE(value)
	return m
}

// WithMaxSpendGas adds a gas credit limit to the collection
func (m *MaxSpendLimits) WithMaxSpendGas(value int64) *MaxSpendLimits {
	m.limits.WithMaxSpendGas(value)
	return m
}

// GetLimit returns the limit for a specific credit type, if it exists
func (m *MaxSpendLimits) GetLimit(credit string) (MaxSpendDefinition, bool) {
	if limit, exists := m.limits.GetLimit(credit); exists {
		return MaxSpendDefinition{
			Credit: limit.Credit,
			Value:  limit.Value,
		}, true
	}
	return MaxSpendDefinition{}, false
}

// toV2 converts the v1 MaxSpendLimits to v2 SpendLimits
func (m *MaxSpendLimits) toV2() *v2.SpendLimits {
	return m.limits
}
