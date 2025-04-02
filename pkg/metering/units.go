package metering

var (
	PayloadUnit = unit{Name: "payload", Unit: "bytes"}

	// ComputeUnit is an example. The compute cap will eventually be obsoleted
	// by the CRE No-DAG SDK.
	ComputeUnit = unit{Name: "compute", Unit: "seconds"}
)

// unit provides exported Name and unit fields for
// capability devs to consume when implementing
// metering. Do not export.
type unit struct {
	// Name of the Metering Unit, i.e. payload, compute, storage
	Name string

	// Unit of the Metering Unit, i.e. bytes, seconds
	Unit string
}
