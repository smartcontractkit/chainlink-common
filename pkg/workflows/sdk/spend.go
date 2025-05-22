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
func (m *MaxSpendLimits) GetLimit(credit string) (*v2.SpendLimit, bool) {
	return m.limits.GetLimit(credit)
}
