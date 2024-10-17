package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/gauge"
	"github.com/grafana/grafana-foundation-sdk/go/heatmap"
	"github.com/grafana/grafana-foundation-sdk/go/logs"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
	"github.com/grafana/grafana-foundation-sdk/go/stat"
	"github.com/grafana/grafana-foundation-sdk/go/table"
	"github.com/grafana/grafana-foundation-sdk/go/timeseries"
)

type Query struct {
	Expr    string
	Legend  string
	Instant bool
	Min     float64
	Format  prometheus.PromQueryFormat
}

func newQuery(query Query) *prometheus.DataqueryBuilder {
	res := prometheus.NewDataqueryBuilder().
		Expr(query.Expr).
		LegendFormat(query.Legend).
		Format(query.Format)

	if query.Instant {
		res.Instant()
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

type TransformOptions struct {
	ID      string
	Options map[string]string
}

func newTransform(options *TransformOptions) dashboard.DataTransformerConfig {
	return dashboard.DataTransformerConfig{
		Id:      options.ID,
		Options: options.Options,
	}
}

type PanelOptions struct {
	Datasource  string
	Title       string
	Description string
	Span        uint32
	Height      uint32
	Decimals    float64
	Unit        string
	NoValue     string
	Min         *float64
	Max         *float64
	Query       []Query
	Threshold   *ThresholdOptions
	Transform   *TransformOptions
	ColorScheme dashboard.FieldColorModeId
}

type Panel struct {
	statPanelBuilder       *stat.PanelBuilder
	timeSeriesPanelBuilder *timeseries.PanelBuilder
	gaugePanelBuilder      *gauge.PanelBuilder
	tablePanelBuilder      *table.PanelBuilder
	logPanelBuilder        *logs.PanelBuilder
	heatmapBuilder         *heatmap.PanelBuilder
	alertBuilder           *alerting.RuleBuilder
}

// panel defaults
func setDefaults(options *PanelOptions) {
	if options.Datasource == "" {
		options.Datasource = "Prometheus"
	}
	if options.Title == "" {
		options.Title = "Panel Title"
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
	TextSize    float64
	ValueSize   float64
	JustifyMode common.BigValueJustifyMode
	ColorMode   common.BigValueColorMode
	GraphMode   common.BigValueGraphMode
	TextMode    common.BigValueTextMode
	Orientation common.VizOrientation
	Mappings    []dashboard.ValueMapping
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

	newPanel := stat.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		Decimals(options.Decimals).
		Unit(options.Unit).
		NoValue(options.NoValue).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		ColorMode(options.ColorMode).
		GraphMode(options.GraphMode).
		TextMode(options.TextMode).
		Orientation(options.Orientation).
		JustifyMode(options.JustifyMode).
		Mappings(options.Mappings).
		ReduceOptions(common.NewReduceDataOptionsBuilder().Calcs([]string{"last"}))

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

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
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
	AlertOptions      *AlertOptions
	FillOpacity       float64
	ScaleDistribution common.ScaleDistribution
	LegendOptions     *LegendOptions
	ThresholdStyle    common.GraphThresholdsStyleMode
}

func NewTimeSeriesPanel(options *TimeSeriesPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	if options.FillOpacity == 0 {
		options.FillOpacity = 2
	}

	if options.ScaleDistribution == "" {
		options.ScaleDistribution = common.ScaleDistributionLinear
	}

	if options.LegendOptions == nil {
		options.LegendOptions = &LegendOptions{}
	}

	newPanel := timeseries.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		Decimals(options.Decimals).
		Unit(options.Unit).
		NoValue(options.NoValue).
		FillOpacity(options.FillOpacity).
		Legend(newLegend(options.LegendOptions)).
		ScaleDistribution(common.NewScaleDistributionConfigBuilder().
			Type(options.ScaleDistribution),
		)

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

		if options.ThresholdStyle != "" {
			newPanel.ThresholdsStyle(common.NewGraphThresholdsStyleConfigBuilder().Mode(options.ThresholdStyle))
		}
	}

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	if options.AlertOptions != nil {
		options.AlertOptions.Name = options.Title

		return &Panel{
			timeSeriesPanelBuilder: newPanel,
			alertBuilder:           NewAlertRule(options.AlertOptions),
		}
	}

	return &Panel{
		timeSeriesPanelBuilder: newPanel,
	}
}

type GaugePanelOptions struct {
	*PanelOptions
}

func NewGaugePanel(options *GaugePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := gauge.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		Decimals(options.Decimals).
		Unit(options.Unit).
		ReduceOptions(
			common.NewReduceDataOptionsBuilder().
				Calcs([]string{"lastNotNull"}).Values(false),
		)

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

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
	}

	return &Panel{
		gaugePanelBuilder: newPanel,
	}
}

type TablePanelOptions struct {
	*PanelOptions
}

func NewTablePanel(options *TablePanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := table.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		Decimals(options.Decimals).
		Unit(options.Unit).
		NoValue(options.NoValue)

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

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	return &Panel{
		tablePanelBuilder: newPanel,
	}
}

type LogPanelOptions struct {
	*PanelOptions
}

func NewLogPanel(options *LogPanelOptions) *Panel {
	setDefaults(options.PanelOptions)

	newPanel := logs.NewPanelBuilder().
		Datasource(datasourceRef(options.Datasource)).
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		NoValue(options.NoValue)

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

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
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
		Title(options.Title).
		Description(options.Description).
		Span(options.Span).
		Height(options.Height).
		Decimals(options.Decimals).
		Unit(options.Unit).
		NoValue(options.NoValue)

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

	if options.Transform != nil {
		newPanel.WithTransformation(newTransform(options.Transform))
	}

	if options.ColorScheme != "" {
		newPanel.ColorScheme(dashboard.NewFieldColorBuilder().Mode(options.ColorScheme))
	}

	return &Panel{
		heatmapBuilder: newPanel,
	}
}
