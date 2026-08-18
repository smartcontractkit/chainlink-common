package polystat

import (
	"encoding/json"

	"github.com/grafana/grafana-foundation-sdk/go/cog/variants"
)

type OperatorName string

const (
	OperatorNameAverage OperatorName = "avg"
	OperatorNameCount   OperatorName = "count"
	OperatorNameCurrent OperatorName = "current"
	OperatorNameDelta   OperatorName = "delta"
	OperatorNameDiff    OperatorName = "diff"
	OperatorNameLast    OperatorName = "last"
	OperatorNameMax     OperatorName = "max"
	OperatorNameMean    OperatorName = "mean"
	OperatorNameMin     OperatorName = "min"
	OperatorNameName    OperatorName = "name"
	OperatorNameSum     OperatorName = "sum"
)

type ClickThroughType string

const (
	ClickThroughTypeAbsolute  ClickThroughType = "absolute"
	ClickThroughTypeDashboard ClickThroughType = "dashboard"
)

type SortByField string

const (
	SortByFieldName  SortByField = "name"
	SortByFieldValue SortByField = "value"
)

type SortByDirection string

const (
	SortByDirectionAscending  SortByDirection = "asc"
	SortByDirectionDescending SortByDirection = "desc"
)

type PolygonSize int

const (
	PolygonSizeSmall  PolygonSize = 15
	PolygonSizeMedium PolygonSize = 25
	PolygonSizeLarge  PolygonSize = 40
)

type Options struct {
	OperatorName                    OperatorName    `json:"operatorName"`
	PolygonGlobalFillColor          string          `json:"polygonGlobalFillColor,omitempty"`
	PolygonSize                     PolygonSize     `json:"polystat.polygonSize,omitempty"`
	Columns                         int             `json:"polystat.columns,omitempty"`
	Rows                            int             `json:"polystat.rows,omitempty"`
	DisplayLimit                    int             `json:"polystat.displayLimit,omitempty"`
	DefaultClickThrough             string          `json:"polystat.defaultClickThrough,omitempty"`
	DefaultClickThroughNewTab       bool            `json:"polystat.defaultClickThroughOpenNewTab,omitempty"`
	DefaultClickThroughSanitize     bool            `json:"polystat.defaultClickThroughSanitize,omitempty"`
	AnimationSpeed                  int             `json:"polystat.animationSpeed,omitempty"`
	Radius                          string          `json:"polystat.radius,omitempty"`
	TooltipDisplayMode              string          `json:"polystat.tooltipDisplayMode,omitempty"`
	TooltipPrimarySortBy            SortByField     `json:"polystat.tooltipPrimarySortBy,omitempty"`
	TooltipPrimarySortDir           SortByDirection `json:"polystat.tooltipPrimarySortDirection,omitempty"`
	TooltipSecondarySortBy          SortByField     `json:"polystat.tooltipSecondarySortBy,omitempty"`
	TooltipSecondarySortDir         SortByDirection `json:"polystat.tooltipSecondarySortDirection,omitempty"`
	GlobalUnitFormat                string          `json:"polystat.globalUnitFormat,omitempty"`
	GlobalDecimals                  int             `json:"polystat.globalDecimals,omitempty"`
	GlobalDisplayMode               string          `json:"polystat.globalDisplayMode,omitempty"`
	GlobalDisplayTextTriggeredEmpty string          `json:"polystat.globalDisplayTextTriggeredEmpty,omitempty"`
	GlobalThresholdsConfig          []Threshold     `json:"globalThresholdsConfig,omitempty"`
}

type Threshold struct {
	Value float64 `json:"value"`
	Color string  `json:"color"`
	State int     `json:"state"`
}

func VariantConfig() variants.PanelcfgConfig {
	return variants.PanelcfgConfig{
		Identifier: "grafana-polystat-panel",
		OptionsUnmarshaler: func(raw []byte) (any, error) {
			options := Options{}

			if err := json.Unmarshal(raw, &options); err != nil {
				return nil, err
			}

			return options, nil
		},
	}
}
