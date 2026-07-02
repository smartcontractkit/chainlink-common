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

// NewUtilization builds a Utilization from a pre-formatted numeric string
// value.
func NewUtilization(value string, fields UtilizationFields) *meteringpb.Utilization {
	return &meteringpb.Utilization{
		Value:        value,
		ResourceType: fields.ResourceType,
		ResourceId:   fields.ResourceID,
		EventId:      fields.EventID,
		OrgId:        fields.OrgID,
	}
}

// NewUtilizationInt builds a Utilization with int64 quantity encoded as a
// decimal string value.
func NewUtilizationInt(value int64, fields UtilizationFields) *meteringpb.Utilization {
	return NewUtilization(strconv.FormatInt(value, 10), fields)
}

// NewUtilizationBig builds a Utilization from an arbitrary-precision integer.
func NewUtilizationBig(value *big.Int, fields UtilizationFields) *meteringpb.Utilization {
	if value == nil {
		return NewUtilization("0", fields)
	}
	return NewUtilization(value.String(), fields)
}

// NewUtilizationFloat builds a Utilization from a floating-point value.
func NewUtilizationFloat(value float64, fields UtilizationFields) *meteringpb.Utilization {
	return NewUtilization(strconv.FormatFloat(value, 'f', -1, 64), fields)
}
