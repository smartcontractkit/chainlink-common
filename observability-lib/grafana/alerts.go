package grafana

import (
	"maps"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/expr"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
)

type RuleQueryType string

const (
	RuleQueryTypeInstant RuleQueryType = "instant"
)

type RuleQuery struct {
	Expr         string
	RefID        string
	Datasource   string
	LegendFormat string
	TimeRange    int64
	Instant      bool
	QueryType    RuleQueryType
}

func newRuleQuery(query RuleQuery) *alerting.QueryBuilder {
	if query.LegendFormat == "" {
		query.LegendFormat = "__auto"
	}

	if query.TimeRange == 0 {
		query.TimeRange = 600
	}

	res := alerting.NewQueryBuilder(query.RefID).
		DatasourceUid(query.Datasource).
		RelativeTimeRange(alerting.Duration(query.TimeRange), alerting.Duration(0))

	model := prometheus.NewDataqueryBuilder().
		Expr(query.Expr).
		LegendFormat(query.LegendFormat).
		RefId(query.RefID)

	if query.Instant {
		model.Instant()
	}

	if query.QueryType != "" {
		model.QueryType(string(query.QueryType))
	}

	return res.Model(model)
}

type ConditionQuery struct {
	RefID               string
	IntervalMs          *float64
	MaxDataPoints       *int64
	ReduceExpression    *ReduceExpression
	MathExpression      *MathExpression
	ResampleExpression  *ResampleExpression
	ThresholdExpression *ThresholdExpression
}

type ReduceExpression struct {
	Expression     string
	Reducer        expr.TypeReduceReducer
	ReduceSettings *expr.ExprTypeReduceSettings
}

type MathExpression struct {
	Expression string
}

type ResampleExpression struct {
	Expression  string
	DownSampler expr.TypeResampleDownsampler
	UpSampler   expr.TypeResampleUpsampler
}

type ThresholdExpression struct {
	Expression                 string
	ThresholdConditionsOptions ThresholdConditionsOption
}

type ThresholdConditionsOption struct {
	Params []float64
	Type   expr.ExprTypeThresholdConditionsEvaluatorType
}

func newThresholdConditionsOptions(options ThresholdConditionsOption) []cog.Builder[expr.ExprTypeThresholdConditions] {
	var conditions []cog.Builder[expr.ExprTypeThresholdConditions]

	var params []float64
	params = append(params, options.Params...)

	if len(options.Params) == 1 {
		params = append(params, 0)
	}

	conditions = append(conditions, expr.NewExprTypeThresholdConditionsBuilder().
		Evaluator(
			expr.NewExprTypeThresholdConditionsEvaluatorBuilder().
				Params(params).
				Type(options.Type),
		),
	)

	return conditions
}

func newReduceSettingsOptions(options expr.ExprTypeReduceSettings) cog.Builder[expr.ExprTypeReduceSettings] {
	builder := expr.NewExprTypeReduceSettingsBuilder().
		Mode(options.Mode)

	if options.Mode == expr.ExprTypeReduceSettingsModeReplaceNN && options.ReplaceWithValue != nil {
		builder.ReplaceWithValue(*options.ReplaceWithValue)
	}

	return builder
}

func newConditionQuery(options ConditionQuery) *alerting.QueryBuilder {
	if options.IntervalMs == nil {
		options.IntervalMs = Pointer[float64](1000)
	}

	if options.MaxDataPoints == nil {
		options.MaxDataPoints = Pointer[int64](43200)
	}

	res := alerting.NewQueryBuilder(options.RefID).
		RelativeTimeRange(600, 0).
		DatasourceUid("__expr__")

	if options.ReduceExpression != nil {
		reduceBuider := expr.NewTypeReduceBuilder().
			RefId(options.RefID).
			Expression(options.ReduceExpression.Expression).
			IntervalMs(*options.IntervalMs).
			MaxDataPoints(*options.MaxDataPoints).
			Reducer(options.ReduceExpression.Reducer)

		if options.ReduceExpression.ReduceSettings != nil && options.ReduceExpression.ReduceSettings.Mode != "" {
			reduceBuider.Settings(newReduceSettingsOptions(*options.ReduceExpression.ReduceSettings))
		}
		res.Model(reduceBuider)
	}

	if options.MathExpression != nil {
		res.Model(expr.NewTypeMathBuilder().
			RefId(options.RefID).
			Expression(options.MathExpression.Expression).
			IntervalMs(*options.IntervalMs).
			MaxDataPoints(*options.MaxDataPoints),
		)
	}

	if options.ResampleExpression != nil {
		res.Model(expr.NewTypeResampleBuilder().
			RefId(options.RefID).
			Expression(options.ResampleExpression.Expression).
			IntervalMs(*options.IntervalMs).
			MaxDataPoints(*options.MaxDataPoints).
			Downsampler(options.ResampleExpression.DownSampler).
			Upsampler(options.ResampleExpression.UpSampler),
		)
	}

	if options.ThresholdExpression != nil {
		res.Model(expr.NewTypeThresholdBuilder().
			RefId(options.RefID).
			Expression(options.ThresholdExpression.Expression).
			IntervalMs(*options.IntervalMs).
			MaxDataPoints(*options.MaxDataPoints).
			Conditions(newThresholdConditionsOptions(options.ThresholdExpression.ThresholdConditionsOptions)),
		)
	}

	return res
}

type AlertOptions struct {
	Title             string
	Summary           string
	Description       string
	RunbookURL        string
	For               string
	NoDataState       alerting.RuleNoDataState
	RuleExecErrState  alerting.RuleExecErrState
	Annotations       map[string]string
	Tags              map[string]string
	Query             []RuleQuery
	QueryRefCondition string
	Condition         []ConditionQuery
	PanelTitle        string
	RuleGroupTitle    string
}

func NewAlertRule(options *AlertOptions) *alerting.RuleBuilder {
	if options.For == "" {
		options.For = "5m"
	}

	if options.NoDataState == "" {
		options.NoDataState = alerting.RuleNoDataStateNoData
	}

	if options.RuleExecErrState == "" {
		options.RuleExecErrState = alerting.RuleExecErrStateAlerting
	}

	if options.QueryRefCondition == "" {
		options.QueryRefCondition = "A"
	}

	annotations := map[string]string{
		"summary":     options.Summary,
		"description": options.Description,
		"runbook_url": options.RunbookURL,
	}
	maps.Copy(annotations, options.Annotations)

	if options.PanelTitle != "" {
		annotations["panel_title"] = options.PanelTitle
	}

	rule := alerting.NewRuleBuilder(options.Title).
		For(options.For).
		NoDataState(options.NoDataState).
		ExecErrState(options.RuleExecErrState).
		Condition(options.QueryRefCondition).
		Annotations(annotations).
		Labels(options.Tags)

	if options.RuleGroupTitle != "" {
		rule.RuleGroup(options.RuleGroupTitle)
	} else {
		rule.RuleGroup("Default")
	}

	for _, query := range options.Query {
		rule.WithQuery(newRuleQuery(query))
	}

	for _, condition := range options.Condition {
		rule.WithQuery(newConditionQuery(condition))
	}

	return rule
}
