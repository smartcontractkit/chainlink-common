package metering

const (
	payloadName = "payload"
	payloadUnit = "bytes"

	computeName = "compute"
	computeUnit = "seconds"
)

type MeteringUnit struct {
	Name string
	Unit string
}

func NewPayloadMeteringUnit() *MeteringUnit {
	return &MeteringUnit{payloadName, payloadUnit}
}

// NewComputeMeteringUnit is an example of how MeteringUnit
// is used for different capabilities. The capability will
// eventually be consumed by the power of the No-DAG SDK.
func NewComputeMeteringUnit() *MeteringUnit {
	return &MeteringUnit{computeName, computeUnit}
}
