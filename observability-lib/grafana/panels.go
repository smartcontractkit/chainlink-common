package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/bargauge"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/gauge"
	"github.com/grafana/grafana-foundation-sdk/go/heatmap"
	"github.com/grafana/grafana-foundation-sdk/go/histogram"
	"github.com/grafana/grafana-foundation-sdk/go/logs"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/table"
	"github.com/grafana/grafana-foundation-sdk/go/text"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana/businessvariable"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana/polystat"
)

type QueryType string

const (
	QueryTypeInstant QueryType = "instant"
)

type Query struct {
	Expr      string
	Legend    string
	Instant   bool
	Min       float64
	Format    prometheus.PromQueryFormat
	QueryType QueryType
}

func newQuery(query Query) *prometheus.DataqueryBuilder {
	res := prometheus.NewDataqueryBuilder().
		Expr(query.Expr).
		LegendFormat(query.Legend).
		Format(query.Format)

	if query.Instant {
		res.Instant()
	}
	if query.QueryType != "" {
		res.QueryType(string(query.QueryType))
	}

	return res
}

type LegendOptions struct {
	Placement   common.LegendPlacement
	DisplayMode common.LegendDisplayMode
	Calcs       []string
}

func newLegend(options *LegendOptions) *common.VizLegendOptionsBuilder {
	if options.DisplayMode == "" {
		options.DisplayMode = common.LegendDisplayModeList
	}

	if options.Placement == "" {
		options.Placement = common.LegendPlacementBottom
	}

	builder := common.NewVizLegendOptionsBuilder().
		ShowLegend(true).
		Placement(options.Placement)

	if len(options.Calcs) > 0 {
		options.DisplayMode = common.LegendDisplayModeTable
		builder.Calcs(options.Calcs)
	}

	builder.DisplayMode(options.DisplayMode)

	return builder
}

type ThresholdOptions struct {
	Mode  dashboard.ThresholdsMode
	Steps []dashboard.Threshold
}

func newThresholds(options *ThresholdOptions) *dashboard.ThresholdsConfigBuilder {
	builder := dashboard.NewThresholdsConfigBuilder().
		Mode(options.Mode).
		Steps(options.Steps)

	return builder
}

func newTransform(transform *Transform) dashboard.DataTransformerConfig {
	return dashboard.DataTransformerConfig{
		Id:      transform.ID,
		Options: transform.Options,
	}
}

func newOverride(override *Override) (matcher dashboard.MatcherConfig, properties []dashboard.DynamicConfigValue) {
	matcher = dashboard.MatcherConfig{
		Id:      override.Matcher.ID,
		Options: override.Matcher.Options,
	}

	for _, property := range override.Properties {
		properties = append(
			properties,
			dashboard.DynamicConfigValue{
				Id:    property.ID,
				Value: property.Value,
			},
		)
	}

	return
}

type ToolTipOptions struct {
	Mode      common.TooltipDisplayMode
	Sort      common.SortOrder
	MaxWidth  *float64
	MaxHeight *float64
}

func newToolTip(options *ToolTipOptions) *common.VizTooltipOptionsBuilder {
	if options.Mode == "" {
		options.Mode = common.TooltipDisplayModeSingle
	}

	if options.Sort == "" {
		options.Sort = common.SortOrderNone
	}

	builder := common.NewVizTooltipOptionsBuilder().
		Mode(options.Mode).
		Sort(options.Sort)

	if options.MaxWidth != nil {
		builder.MaxWidth(*options.MaxWidth)
	}

	if options.MaxHeight != nil {
		builder.MaxHeight(*options.MaxHeight)
	}

	return builder
}

type PanelOptions struct {
	Datasource    string
	Title         *string
	Description   string
	Transparent   bool
	Span          uint32
	Height        uint32
	Decimals      *float64
	Unit          string
	NoValue       string
	Min           *float64
	Max           *float64
	MaxDataPoints *float64
	Query         []Query
	Threshold     *ThresholdOptions
	Transforms    []*Transform
	Overrides     []*Override
	ColorScheme   dashboard.FieldColorModeId
	Interval      string
}

type Panel struct {
	statPanelBuilder             *stat.PanelBuilder
	timeSeriesPanelBuilder       *timeseries.PanelBuilder
	gaugePanelBuilder            *gauge.PanelBuilder
	barGaugePanelBuilder         *bargauge.PanelBuilder
	tablePanelBuilder            *table.PanelBuilder
	logPanelBuilder              *logs.PanelBuilder
	heatmapBuilder               *heatmap.PanelBuilder
	textPanelBuilder             *text.PanelBuilder
	histogramPanelBuilder        *histogram.PanelBuilder
	businessVariablePanelBuilder *businessvariable.PanelBuilder
	polystatPanelBuilder         *polystat.PanelBuilder
	alertBuilders                []*alerting.RuleBuilder
}

// panel defaults
func setDefaults(options *PanelOptions) {
	if options.Datasource == "" {
		options.Datasource = "Prometheus"
	}
	if options.Title == nil {
		options.Title = Pointer("Panel Title")
	}
	if options.Span == 0 {
		options.Span = 24
	}
	if options.Height == 0 {
		options.Height = 6
	}
	if options.NoValue == "" {
		options.NoValue = "No data"
	}
}

type StatPanelOptions struct {
	*PanelOptions
	TextSize      float64
	ValueSize     float64
	JustifyMode   common.BigValueJustifyMode
	ColorMode     common.BigValueColorMode
	GraphMode     common.BigValueGraphMode
	TextMode      common.BigValueTextMode
	Orientation   common.VizOrientation
	Mappings      []dashboard.ValueMapping
	ReduceOptions *common.ReduceDataOptionsBuilder
}

func NewStatPanel(options *StatPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.JustifyMode == "" {
		options.JustifyMode = common.BigValueJustifyModeAuto
	}
	if options.ColorMode == "" {
		options.ColorMode = common.BigValueColorModeValue
	}
	if options.GraphMode == "" {
		options.GraphMode = common.BigValueGraphModeNone
	}
	if options.TextMode == "" {
		options.TextMode = common.BigValueTextModeValue
	}
	if options.Orientation == "" {
		options.Orientation = common.VizOrientationAuto
	}

	if options.ReduceOptions == nil {
		options.ReduceOptions = common.NewReduceDataOptionsBuilder().Calcs([]string{"last"})
	}

	newPanel := stat.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Unit(options.Unit).
		NoValue(options.NoValue).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		ColorMode(options.ColorMode).
		GraphMode(options.GraphMode).
		TextMode(options.TextMode).
		Orientation(options.Orientation).
		JustifyMode(options.JustifyMode).
		Mappings(options.Mappings).
		ReduceOptions(options.ReduceOptions)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	if options.TextSize != 0 || options.ValueSize != 0 {
		vizTextDisplayOptions := common.NewVizTextDisplayOptionsBuilder()
		if options.TextSize != 0 {
			vizTextDisplayOptions.TitleSize(options.TextSize)
		}
		if options.ValueSize != 0 {
			vizTextDisplayOptions.ValueSize(options.ValueSize)
		}

		newPanel.Text(vizTextDisplayOptions)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	return &Panel{
		statPanelBuilder: newPanel,
	}
}

type TimeSeriesPanelOptions struct {
	*PanelOptions
	AlertsOptions     []AlertOptions
	LineWidth         *float64
	FillOpacity       float64
	ScaleDistribution common.ScaleDistribution
	LegendOptions     *LegendOptions
	ToolTipOptions    *ToolTipOptions
	ThresholdStyle    common.GraphThresholdsStyleMode
	DrawStyle         common.GraphDrawStyle
	StackingMode      common.StackingMode
	AxisSoftMin       *float64
	AxisSoftMax       *float64
}

func NewTimeSeriesPanel(options *TimeSeriesPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.ScaleDistribution == "" {
		options.ScaleDistribution = common.ScaleDistributionLinear
	}

	if options.LineWidth == nil {
		options.LineWidth = Pointer[float64](1)
	}

	if options.LegendOptions == nil {
		options.LegendOptions = &LegendOptions{}
	}

	if options.ToolTipOptions == nil {
		options.ToolTipOptions = &ToolTipOptions{}
	}

	if options.StackingMode == "" {
		options.StackingMode = common.StackingModeNone
	}

	newPanel := timeseries.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Unit(options.Unit).
		NoValue(options.NoValue).
		LineWidth(*options.LineWidth).
		FillOpacity(options.FillOpacity).
		Legend(newLegend(options.LegendOptions)).
		ScaleDistribution(common.NewScaleDistributionConfigBuilder().
			Type(options.ScaleDistribution),
		).
		Tooltip(newToolTip(options.ToolTipOptions)).
		// Time Series Panel Options
		Stacking(common.NewStackingConfigBuilder().
			Mode(options.StackingMode),
		)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	if options.AxisSoftMin != nil {
		newPanel.AxisSoftMin(*options.AxisSoftMin)
	}

	if options.AxisSoftMax != nil {
		newPanel.AxisSoftMax(*options.AxisSoftMax)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))

		if options.ThresholdStyle != "" {
			newPanel.ThresholdsStyle(common.NewGraphThresholdsStyleConfigBuilder().Mode(options.ThresholdStyle))
		}
	}

	if options.DrawStyle != "" {
		newPanel.DrawStyle(options.DrawStyle)
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	var alertBuilders []*alerting.RuleBuilder
	if len(options.AlertsOptions) > 0 {
		for _, alert := range options.AlertsOptions {
			// this is used as an internal mechanism to set the panel title in the alert to associate panelId with alert
			alert.PanelTitle = *options.Title
			// if name is provided use it, otherwise use panel title
			if alert.Title == "" {
				alert.Title = *options.Title
			}
			if alert.RuleGroupTitle == "" {
				alert.RuleGroupTitle = *options.Title
			}
			alertBuilders = append(alertBuilders, NewAlertRule(&alert))
		}
	}

	return &Panel{
		timeSeriesPanelBuilder: newPanel,
		alertBuilders:          alertBuilders,
	}
}

type BarGaugePanelOptions struct {
	*PanelOptions
	ShowUnfilled  bool
	Orientation   common.VizOrientation
	ShowAllValues bool
}

func NewBarGaugePanel(options *BarGaugePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := bargauge.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		NoValue(options.NoValue).
		Height(options.Height).
		Unit(options.Unit).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}).Values(options.ShowAllValues),
		)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.ShowUnfilled {
		newPanel.ShowUnfilled(options.ShowUnfilled)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.Orientation != "" {
		newPanel.Orientation(options.Orientation)
	}

	return &Panel{
		barGaugePanelBuilder: newPanel,
	}
}

type GaugePanelOptions struct {
	*PanelOptions
	ShowAllValues bool
}

func NewGaugePanel(options *GaugePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := gauge.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		NoValue(options.NoValue).
		Height(options.Height).
		Unit(options.Unit).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}).Values(options.ShowAllValues),
		)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	return &Panel{
		gaugePanelBuilder: newPanel,
	}
}

type SortByOptions struct {
	DisplayName string
	Desc        *bool
}

type FooterReducer string

const (
	FooterReducerSum FooterReducer = "sum"
)

type FooterOptions struct {
	Show             bool
	Reducer          FooterReducer
	Fields           []string
	EnablePagination bool
}

type TablePanelOptions struct {
	*PanelOptions
	Filterable bool
	Footer     *FooterOptions
	SortBy     []*SortByOptions
}

func NewTablePanel(options *TablePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := table.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Unit(options.Unit).
		NoValue(options.NoValue).
		Filterable(options.Filterable)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	if options.Footer != nil {
		footer := common.NewTableFooterOptionsBuilder().
			Show(options.Footer.Show).
			EnablePagination(options.Footer.EnablePagination)

		if options.Footer.Reducer != "" && options.Footer.Fields != nil {
			footer.
				Reducer([]string{string(options.Footer.Reducer)}).
				Fields(options.Footer.Fields)
		}

		newPanel.Footer(footer)
	}

	if options.SortBy != nil {
		var sortBy []cog.Builder[common.TableSortByFieldState]
		for _, sb := range options.SortBy {
			tableSortBy := common.NewTableSortByFieldStateBuilder().
				DisplayName(sb.DisplayName)

			if sb.Desc != nil {
				tableSortBy.Desc(*sb.Desc)
			}

			sortBy = append(sortBy, tableSortBy)
		}
		newPanel.SortBy(sortBy)
	}

	return &Panel{
		tablePanelBuilder: newPanel,
	}
}

type LogPanelOptions struct {
	*PanelOptions
	ShowTime         bool
	PrettifyJSON     bool
	EnableLogDetails *bool
	DedupStrategy    common.LogsDedupStrategy
	SortOrder        common.LogsSortOrder
}

func NewLogPanel(options *LogPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.EnableLogDetails == nil {
		options.EnableLogDetails = Pointer[bool](true)
	}

	if options.DedupStrategy == "" {
		options.DedupStrategy = common.LogsDedupStrategyNone
	}

	if options.SortOrder == "" {
		options.SortOrder = common.LogsSortOrderDescending // Newest First
	}

	newPanel := logs.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		NoValue(options.NoValue).
		ShowTime(options.ShowTime).
		PrettifyLogMessage(options.PrettifyJSON).
		EnableLogDetails(*options.EnableLogDetails).
		DedupStrategy(options.DedupStrategy).
		SortOrder(options.SortOrder)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.MaxDataPoints != nil {
		newPanel.MaxDataPoints(*options.MaxDataPoints)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	return &Panel{
		logPanelBuilder: newPanel,
	}
}

type HeatmapPanelOptions struct {
	*PanelOptions
}

func NewHeatmapPanel(options *HeatmapPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := heatmap.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Unit(options.Unit).
		NoValue(options.NoValue)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		q.Format = prometheus.PromQueryFormatHeatmap
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	return &Panel{
		heatmapBuilder: newPanel,
	}
}

type TextPanelOptions struct {
	*PanelOptions
	Mode    text.TextMode
	Content string
}

func NewTextPanel(options *TextPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.Mode == "" {
		options.Mode = text.TextModeMarkdown
	}

	newPanel := text.NewPanelBuilder().
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Mode(options.Mode).
		Content(options.Content)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	return &Panel{
		textPanelBuilder: newPanel,
	}
}

type HistogramPanelOptions struct {
	*PanelOptions
	LineWidth         *float64
	ScaleDistribution common.ScaleDistribution
	LegendOptions     *LegendOptions
	ToolTipOptions    *ToolTipOptions
	ThresholdStyle    common.GraphThresholdsStyleMode
	DrawStyle         common.GraphDrawStyle
	StackingMode      common.StackingMode
	Combine           *bool
	FillOpacity       *uint32
	BucketOffset      *float32
	BucketCount       *int32
	BucketSize        *int32
	AxisBorderShow    *bool
	AxisLabel         string
	AxisSoftMax       *float64
	AxisSoftMin       *float64
	AxisColorMode     common.AxisColorMode
	AxisCenteredZero  *bool
	AxisGridShow      *bool
	AxisPlacement     common.AxisPlacement
	AxisWidth         *float64
}

func NewHistogramPanel(options *HistogramPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.ScaleDistribution == "" {
		options.ScaleDistribution = common.ScaleDistributionLinear
	}

	if options.LineWidth == nil {
		options.LineWidth = Pointer[float64](1)
	}

	if options.LegendOptions == nil {
		options.LegendOptions = &LegendOptions{}
	}

	if options.ToolTipOptions == nil {
		options.ToolTipOptions = &ToolTipOptions{}
	}

	if options.StackingMode == "" {
		options.StackingMode = common.StackingModeNone
	}

	newPanel := histogram.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		Unit(options.Unit).
		NoValue(options.NoValue).
		ScaleDistribution(common.NewScaleDistributionConfigBuilder().
			Type(options.ScaleDistribution),
		).
		Tooltip(newToolTip(options.ToolTipOptions)).
		Stacking(common.NewStackingConfigBuilder().
			Mode(options.StackingMode),
		)

	if options.Interval != "" {
		newPanel.Interval(options.Interval)
	}

	if options.Decimals != nil {
		newPanel.Decimals(*options.Decimals)
	}

	if options.Min != nil {
		newPanel.Min(*options.Min)
	}

	if options.Max != nil {
		newPanel.Max(*options.Max)
	}

	for _, q := range options.Query {
		newPanel.WithTarget(newQuery(q))
	}

	if options.Threshold != nil {
		newPanel.Thresholds(newThresholds(options.Threshold))
	}

	if options.Transforms != nil {
		for _, transform := range options.Transforms {
			newPanel.WithTransformation(newTransform(transform))
		}
	}

	if options.Overrides != nil {
		for _, override := range options.Overrides {
			newPanel.WithOverride(newOverride(override))
		}
	}

	if options.Combine != nil {
		newPanel.Combine(*options.Combine)
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	if options.FillOpacity != nil {
		newPanel.FillOpacity(*options.FillOpacity)
	}

	if options.BucketOffset != nil {
		newPanel.BucketOffset(*options.BucketOffset)
	}

	if options.BucketCount != nil {
		newPanel.BucketCount(*options.BucketCount)
	}

	if options.BucketSize != nil {
		newPanel.BucketSize(*options.BucketSize)
	}

	if options.AxisBorderShow != nil {
		newPanel.AxisBorderShow(*options.AxisBorderShow)
	}

	if options.AxisLabel != "" {
		newPanel.AxisLabel(options.AxisLabel)
	}

	if options.AxisSoftMax != nil {
		newPanel.AxisSoftMax(*options.AxisSoftMax)
	}

	if options.AxisSoftMin != nil {
		newPanel.AxisSoftMin(*options.AxisSoftMin)
	}

	if options.AxisColorMode != "" {
		newPanel.AxisColorMode(options.AxisColorMode)
	}

	if options.AxisCenteredZero != nil {
		newPanel.AxisCenteredZero(*options.AxisCenteredZero)
	}

	if options.AxisGridShow != nil {
		newPanel.AxisGridShow(*options.AxisGridShow)
	}

	if options.AxisPlacement != "" {
		newPanel.AxisPlacement(options.AxisPlacement)
	}

	if options.AxisWidth != nil {
		newPanel.AxisWidth(*options.AxisWidth)
	}

	return &Panel{
		histogramPanelBuilder: newPanel,
	}
}

type BusinessVariablePanelOptions struct {
	*PanelOptions
	DisplayMode businessvariable.DisplayMode
	Padding     *int
	ShowLabel   bool
	Variable    string
}

func NewBusinessVariablePanel(options *BusinessVariablePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := businessvariable.NewPanelBuilder().
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		DisplayMode(options.DisplayMode).
		ShowLabel(options.ShowLabel).
		Variable(options.Variable)

	if options.Padding != nil {
		newPanel.Padding(*options.Padding)
	}

	return &Panel{
		businessVariablePanelBuilder: newPanel,
	}
}

type PolystatPanelOptions struct {
	*PanelOptions
	Queries                         []Query
	OperatorName                    polystat.OperatorName
	PolygonGlobalFillColor          string
	PolygonSize                     polystat.PolygonSize
	Columns                         *int
	Rows                            *int
	DisplayLimit                    *int
	DefaultClickThrough             string
	DefaultClickThroughNewTab       bool
	DefaultClickThroughSanitize     bool
	AnimationSpeed                  *int
	Radius                          string
	TooltipDisplayMode              string
	TooltipPrimarySortBy            polystat.SortByField
	TooltipPrimarySortDirection     polystat.SortByDirection
	TooltipSecondarySortBy          polystat.SortByField
	TooltipSecondarySortDirection   polystat.SortByDirection
	GlobalUnitFormat                string
	GlobalDecimals                  *int
	GlobalDisplayMode               string
	GlobalDisplayTextTriggeredEmpty string
	GlobalThresholds                []polystat.Threshold
}

func NewPolystatPanel(options *PolystatPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := polystat.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(*options.Title).
		Description(options.Description).
		Transparent(options.Transparent).
		Span(options.Span).
		Height(options.Height).
		OperatorName(options.OperatorName)

	if options.PolygonGlobalFillColor != "" {
		newPanel.PolygonGlobalFillColor(options.PolygonGlobalFillColor)
	}

	if options.PolygonSize != 0 {
		newPanel.PolygonSize(options.PolygonSize)
	}

	if options.Columns != nil {
		newPanel.Columns(*options.Columns)
	}

	if options.Rows != nil {
		newPanel.Rows(*options.Rows)
	}

	if options.DisplayLimit != nil {
		newPanel.DisplayLimit(*options.DisplayLimit)
	}

	if options.DefaultClickThrough != "" {
		newPanel.DefaultClickThrough(options.DefaultClickThrough)
	}

	newPanel.DefaultClickThroughNewTab(options.DefaultClickThroughNewTab)

	newPanel.DefaultClickThroughSanitize(options.DefaultClickThroughSanitize)

	if options.AnimationSpeed != nil {
		newPanel.AnimationSpeed(*options.AnimationSpeed)
	}

	if options.Radius != "" {
		newPanel.Radius(options.Radius)
	}

	if options.TooltipDisplayMode != "" {
		newPanel.TooltipDisplayMode(options.TooltipDisplayMode)
	}

	if options.TooltipPrimarySortBy != "" {
		newPanel.TooltipPrimarySortBy(options.TooltipPrimarySortBy)
	}

	if options.TooltipPrimarySortDirection != "" {
		newPanel.TooltipPrimarySortDirection(options.TooltipPrimarySortDirection)
	}

	if options.TooltipSecondarySortBy != "" {
		newPanel.TooltipSecondarySortBy(options.TooltipSecondarySortBy)
	}

	if options.TooltipSecondarySortDirection != "" {
		newPanel.TooltipSecondarySortDirection(options.TooltipSecondarySortDirection)
	}

	if options.GlobalUnitFormat != "" {
		newPanel.GlobalUnitFormat(options.GlobalUnitFormat)
	}

	if options.GlobalDecimals != nil {
		newPanel.GlobalDecimals(*options.GlobalDecimals)
	}

	if options.GlobalDisplayMode != "" {
		newPanel.GlobalDisplayMode(options.GlobalDisplayMode)
	}

	if options.GlobalDisplayTextTriggeredEmpty != "" {
		newPanel.GlobalDisplayTextTriggeredEmpty(options.GlobalDisplayTextTriggeredEmpty)
	}

	if len(options.GlobalThresholds) > 0 {
		newPanel.GlobalThresholds(options.GlobalThresholds)
	}

	for _, query := range options.Queries {
		newPanel.WithTarget(newQuery(query))
	}

	return &Panel{
		polystatPanelBuilder: newPanel,
	}
}
