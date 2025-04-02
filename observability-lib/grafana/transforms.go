package grafana

type Transform struct {
	ID      string
	Options map[string]any
}

type JoinByFieldMode string

const (
	JoinByFieldModeOuter        JoinByFieldMode = "outer"
	JoinByFieldModeOuterTabular JoinByFieldMode = "outerTabular"
	JoinByFieldModeOuterInner   JoinByFieldMode = "inner"
)

type JoinByFieldTransformOptions struct {
	ByField string
	Mode    JoinByFieldMode
}

func NewJoinByFieldTransform(options *JoinByFieldTransformOptions) *Transform {
	return &Transform{
		ID: "joinByField",
		Options: map[string]any{
			"byField": options.ByField,
			"mode":    options.Mode,
		},
	}
}

type OrganizeTransformOptions struct {
	ExcludeByName map[string]bool
	IndexByName   map[string]int
	RenameByName  map[string]string
	IncludeByName map[string]bool
}

func NewOrganizeTransform(options *OrganizeTransformOptions) *Transform {
	return &Transform{
		ID: "organize",
		Options: map[string]any{
			"excludeByName": options.ExcludeByName,
			"indexByName":   options.IndexByName,
			"renameByName":  options.RenameByName,
			"includeByName": options.IncludeByName,
		},
	}
}

type ConversionOptions struct {
	TargetField string
}

type Conversion map[string]any

type TimeConversionOptions struct {
	*ConversionOptions
}

func NewTimeConversion(options *TimeConversionOptions) *Conversion {
	return &Conversion{
		"targetField":     options.TargetField,
		"destinationType": "time",
	}
}

type ConvertFieldTypeTransformOptions struct {
	Conversions []*Conversion
}

func NewConvertFieldTypeTransform(options *ConvertFieldTypeTransformOptions) *Transform {
	return &Transform{
		ID: "convertFieldType",
		Options: map[string]any{
			"conversions": options.Conversions,
		},
	}
}

type RenameByRegexTransformOptions struct {
	Regex         string
	RenamePattern string
}

func NewRenameByRegexTransform(options *RenameByRegexTransformOptions) *Transform {
	return &Transform{
		ID: "renameByRegex",
		Options: map[string]any{
			"regex":         options.Regex,
			"renamePattern": options.RenamePattern,
		},
	}
}
