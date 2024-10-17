package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/expr"
	"github.com/grafana/grafana-foundation-sdk/go/prometheus"
)

type RuleQuery struct {
	Expr         string
	RefID        string
	Datasource   string
	LegendFormat string
	Instant      bool
}

func newRuleQuery(query RuleQuery) *alerting.QueryBuilder {
	if query.LegendFormat == "" {
		query.LegendFormat = "__auto"
	}

	res := alerting.NewQueryBuilder(query.RefID).
		DatasourceUid(query.Datasource).
		RelativeTimeRange(600, 0) // TODO

	model := prometheus.NewDataqueryBuilder().
		Expr(query.Expr).
		LegendFormat(query.LegendFormat).
		RefId(query.RefID)

	if query.Instant {
		model.Instant()
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
	Expression string
	Reducer    expr.TypeReduceReducer
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

type TypeThresholdType string

const (
	TypeThresholdTypeGt           TypeThresholdType = "gt"
	TypeThresholdTypeLt           TypeThresholdType = "lt"
	TypeThresholdTypeWithinRange  TypeThresholdType = "within_range"
	TypeThresholdTypeOutsideRange TypeThresholdType = "outside_range"
)

type ThresholdConditionsOption struct {
	Params []float64
	Type   TypeThresholdType
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
				Type(expr.TypeThresholdType(options.Type)),
		),
	)

	return conditions
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
		res.Model(expr.NewTypeReduceBuilder().
			RefId(options.RefID).
			Expression(options.ReduceExpression.Expression).
			IntervalMs(*options.IntervalMs).
			MaxDataPoints(*options.MaxDataPoints).
			Reducer(options.ReduceExpression.Reducer),
		)
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
	Name              string
	Summary           string
	Description       string
	RunbookURL        string
	For               string
	NoDataState       alerting.RuleNoDataState
	RuleExecErrState  alerting.RuleExecErrState
	Tags              map[string]string
	Query             []RuleQuery
	QueryRefCondition string
	Condition         []ConditionQuery
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

	rule := alerting.NewRuleBuilder(options.Name).
		For(options.For).
		NoDataState(options.NoDataState).
		ExecErrState(options.RuleExecErrState).
		Condition(options.QueryRefCondition).
		Annotations(map[string]string{
			"summary":     options.Summary,
			"description": options.Description,
			"runbook_url": options.RunbookURL,
		}).
		Labels(options.Tags)

	for _, query := range options.Query {
		rule.WithQuery(newRuleQuery(query))
	}

	for _, condition := range options.Condition {
		rule.WithQuery(newConditionQuery(condition))
	}

	return rule
}
