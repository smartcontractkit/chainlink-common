package codec

type ModificationConfig struct {
	Modifications map[string]Modifications
}

type TransformType string

const (
	NoTransform            TransformType = ""
	FirstElementTransform  TransformType = "first element"
	MiddleElementTransform TransformType = "middle element"
	LastElementTransform   TransformType = "last element"
)

type Modifications struct {
	TransformType string
	OutputField   string
}
