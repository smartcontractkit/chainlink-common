package sdk

// Credit types that can be used for max spend limits
const (
	CreditTypeCRE = "CRE_CREDITS" // Universal CRE Capability Credits
	CreditTypeGas = "CRE_GAS"     // Universal CRE Gas Credits
)

// SpendLimit represents a single spending limit for a specific credit type
type SpendLimit struct {
	Credit string
	Value  int64
}

// SpendLimits represents a collection of spending limits that can be applied
// to a workflow or individual steps
type SpendLimits struct {
	Limits []SpendLimit
}

// hasCreditType checks if a credit type already exists in the limits
func (m *SpendLimits) hasCreditType(credit string) bool {
	for _, limit := range m.Limits {
		if limit.Credit == credit {
			return true
		}
	}
	return false
}

// updateExistingLimit updates the value for an existing credit type
func (m *SpendLimits) updateExistingLimit(credit string, value int64) {
	for i, limit := range m.Limits {
		if limit.Credit == credit {
			m.Limits[i].Value = value
			return
		}
	}
}

// WithMaxSpend adds a spending limit to the collection. If the credit type already exists,
// it will update the existing limit with the new value.
func (m *SpendLimits) WithMaxSpend(credit string, value int64) *SpendLimits {
	if m.hasCreditType(credit) {
		m.updateExistingLimit(credit, value)
	} else {
		m.Limits = append(m.Limits, SpendLimit{
			Credit: credit,
			Value:  value,
		})
	}
	return m
}

// WithMaxSpendCRE adds a CRE credit limit to the collection
func (m *SpendLimits) WithMaxSpendCRE(value int64) *SpendLimits {
	return m.WithMaxSpend(CreditTypeCRE, value)
}

// WithMaxSpendGas adds a gas credit limit to the collection
func (m *SpendLimits) WithMaxSpendGas(value int64) *SpendLimits {
	return m.WithMaxSpend(CreditTypeGas, value)
}

// GetLimit returns the limit for a specific credit type, if it exists
func (m *SpendLimits) GetLimit(credit string) (*SpendLimit, bool) {
	for _, limit := range m.Limits {
		if limit.Credit == credit {
			return &limit, true
		}
	}
	return nil, false
}

// NewSpendLimits creates a new SpendLimits instance
func NewSpendLimits() *SpendLimits {
	return &SpendLimits{
		Limits: make([]SpendLimit, 0),
	}
}
