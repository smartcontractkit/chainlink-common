package resourcemanager

import (
	"math/big"
	"strconv"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// UtilizationFields identifies one billed utilization dimension.
type UtilizationFields struct {
	ResourceType string
	ResourceID   string
	EventID      string
	OrgID        string
}

// NewUtilization builds a Utilization with int64 quantity encoded as a decimal
// string value.
func NewUtilization(value int64, fields UtilizationFields) *meteringpb.Utilization {
	return NewUtilizationString(strconv.FormatInt(value, 10), fields)
}

// NewUtilizationString builds a Utilization from a pre-formatted numeric string
// value.
func NewUtilizationString(value string, fields UtilizationFields) *meteringpb.Utilization {
	return &meteringpb.Utilization{
		Value:        value,
		ResourceType: fields.ResourceType,
		ResourceId:   fields.ResourceID,
		EventId:      fields.EventID,
		OrgId:        fields.OrgID,
	}
}

// NewUtilizationBig builds a Utilization from an arbitrary-precision integer.
func NewUtilizationBig(value *big.Int, fields UtilizationFields) *meteringpb.Utilization {
	if value == nil {
		return NewUtilizationString("0", fields)
	}
	return NewUtilizationString(value.String(), fields)
}

// NewUtilizationFloat builds a Utilization from a floating-point value.
func NewUtilizationFloat(value float64, fields UtilizationFields) *meteringpb.Utilization {
	return NewUtilizationString(strconv.FormatFloat(value, 'g', -1, 64), fields)
}
