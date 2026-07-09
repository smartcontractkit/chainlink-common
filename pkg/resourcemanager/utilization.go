package resourcemanager

import (
	"math/big"
	"strconv"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// UtilizationFields identifies one billed utilization dimension.
//
// EventID is intentionally NOT a field here: event_id is the consumer's dedup
// key and is generated fresh per emission by the ResourceManager (a UUIDv4).
// Producers must never populate it, so it cannot be supplied through this
// struct; the manager stamps it on every emitted Utilization.
type UtilizationFields struct {
	ResourceType string
	ResourceID   string
	OrgID        string
}

// NewUtilization builds a Utilization from a pre-formatted numeric string
// value. event_id is left empty; the ResourceManager stamps it at emit time.
func NewUtilization(value string, fields UtilizationFields) *meteringpb.Utilization {
	return &meteringpb.Utilization{
		Value:        value,
		ResourceType: fields.ResourceType,
		ResourceId:   fields.ResourceID,
		OrgId:        fields.OrgID,
	}
}

// NewUtilizationInt builds a Utilization with an int64 quantity encoded as a
// decimal string value. The value may be negative (e.g. an unregister or
// resize-down delta of -N for METER_ACTION_UPDATE).
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
