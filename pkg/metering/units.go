package metering

const (
	payloadName = "payload"
	payloadUnit = "bytes"

	computeName = "compute"
	computeUnit = "seconds"
)

// MeteringUnit stores the canonical values of the units
// that capabilities will meter their spend on
type MeteringUnit interface {
	GetName() string
	GetUnit() string
}

type PayloadMeteringUnit struct {
	name string
	unit string
}

func NewPayloadMeteringUnit() PayloadMeteringUnit {
	return PayloadMeteringUnit{
		name: payloadName,
		unit: payloadUnit,
	}
}

func (p PayloadMeteringUnit) GetName() string {
	return p.name
}

func (p PayloadMeteringUnit) GetUnit() string {
	return p.unit
}

type ComputeMeteringUnit struct {
	name string
	unit string
}

func NewComputeMeteringUnit() ComputeMeteringUnit {
	return ComputeMeteringUnit{
		name: computeName,
		unit: computeUnit,
	}
}

func (c ComputeMeteringUnit) GetName() string {
	return c.name
}

func (c ComputeMeteringUnit) GetUnit() string {
	return c.unit
}
