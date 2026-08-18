package grafana

type Override struct {
	Matcher    *Matcher
	Properties []*Property
}

type Matcher struct {
	ID      string
	Options any
}

func NewByNameMatcher(name string) *Matcher {
	return &Matcher{
		ID:      "byName",
		Options: name,
	}
}

type MatcherReducer string

const (
	MatcherReducerLastNotNull MatcherReducer = "lastNotNull"
)

type MatcherOp string

const (
	MatcherOpGTE MatcherOp = "gte"
)

type ByValueMatcherOptions struct {
	Reducer MatcherReducer
	Op      MatcherOp
	Value   float64
}

func NewByValueMatcher(options *ByValueMatcherOptions) *Matcher {
	return &Matcher{
		ID: "byValue",
		Options: map[string]any{
			"reducer": options.Reducer,
			"op":      options.Op,
			"value":   options.Value,
		},
	}
}

type MatcherType string

const (
	MatcherTypeTime MatcherType = "time"
)

func NewByTypeMatcher(t MatcherType) *Matcher {
	return &Matcher{
		ID:      "byType",
		Options: t,
	}
}

func NewByRegexpMatcher(regex string) *Matcher {
	return &Matcher{
		ID:      "byRegexp",
		Options: regex,
	}
}

func NewByQueryMatcher(refID string) *Matcher {
	return &Matcher{
		ID:      "byFrameRefID",
		Options: refID,
	}
}

type Property struct {
	ID    string
	Value any
}

type ColorMode string

const (
	ColorModeFixed          ColorMode = "fixed"
	ColorModeContinuousYlRd ColorMode = "continuous-YlRd"
)

func NewColorModeFixedProperty(fixedColor string) *Property {
	return &Property{
		ID: "color",
		Value: map[string]any{
			"mode":       ColorModeFixed,
			"fixedColor": fixedColor,
		},
	}
}

func NewColorModeContinuousYlRdProperty(seriesBy string) *Property {
	return &Property{
		ID: "color",
		Value: map[string]any{
			"mode":     ColorModeContinuousYlRd,
			"seriesBy": seriesBy,
		},
	}
}

type UnitValue string

const (
	UnitValueBlock UnitValue = "block"
)

func NewUnitProperty(value UnitValue) *Property {
	return &Property{
		ID:    "unit",
		Value: value,
	}
}

type LinksPropertyOptions struct {
	TargetBlank bool
	Title       string
	URL         string
}

func NewLinksProperty(options *LinksPropertyOptions) *Property {
	return &Property{
		ID: "links",
		Value: map[string]any{
			"targetBlank": options.TargetBlank,
			"title":       options.Title,
			"url":         options.URL,
		},
	}
}

func NewFilterableProperty(value bool) *Property {
	return &Property{
		ID:    "filterable",
		Value: value,
	}
}

func NewHiddenProperty(value bool) *Property {
	return &Property{
		ID:    "custom.hidden",
		Value: value,
	}
}

func NewWidthProperty(value float64) *Property {
	return &Property{
		ID:    "custom.width",
		Value: value,
	}
}

func NewMinWidthProperty(value float64) *Property {
	return &Property{
		ID:    "custom.minWidth",
		Value: value,
	}
}

func NewLineWidthProperty(value float64) *Property {
	return &Property{
		ID:    "custom.lineWidth",
		Value: value,
	}
}

type CellOptionsMode string

const (
	CellOptionsModeBasic CellOptionsMode = "basic"
)

type CellOptionsType string

const (
	CellOptionsTypeColorBackground CellOptionsType = "color-background"
)

type CellOptionsOptions struct {
	Mode CellOptionsMode
	Type CellOptionsType
}

func NewCellOptions(options *CellOptionsOptions) *Property {
	return &Property{
		ID: "custom.cellOptions",
		Value: map[string]any{
			"mode": options.Mode,
			"type": options.Type,
		},
	}
}

type LineStyleMode string

const (
	LineStyleSolid  LineStyleMode = "solid"
	LineStyleDashed LineStyleMode = "dashed"
	LineStyleDotted LineStyleMode = "dotted"
)

func NewLineStyleSolidProperty() *Property {
	return &Property{
		ID: "custom.lineStyle",
		Value: map[string]any{
			"fill": LineStyleSolid,
		},
	}
}

func NewLineStyleDashedProperty(gaps ...int) *Property {
	return &Property{
		ID: "custom.lineStyle",
		Value: map[string]any{
			"fill": LineStyleDashed,
			"dash": gaps,
		},
	}
}

func NewLineStyleDottedProperty(gaps ...int) *Property {
	return &Property{
		ID: "custom.lineStyle",
		Value: map[string]any{
			"fill": LineStyleDotted,
			"dash": gaps,
		},
	}
}
