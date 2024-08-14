package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

type VariableOption struct {
	Name  string
	Label string
}

type QueryVariableOptions struct {
	*VariableOption
	Datasource string
	Query      string
	Multi      bool
	Regex      string
}

func NewQueryVariable(options *QueryVariableOptions) *dashboard.QueryVariableBuilder {
	variable := dashboard.NewQueryVariableBuilder(options.Name).
		Label(options.Label).
		Query(dashboard.StringOrMap{String: cog.ToPtr[string](options.Query)}).
		Datasource(datasourceRef(options.Datasource)).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{ArrayOfString: []string{"All"}},
			Value:    dashboard.StringOrArrayOfString{ArrayOfString: []string{"$__all"}},
		}).
		Sort(dashboard.VariableSortAlphabeticalAsc).
		Multi(options.Multi).
		IncludeAll(true)

	if options.Regex != "" {
		variable.Regex(options.Regex)
	}

	return variable
}

type IntervalVariableOptions struct {
	*VariableOption
	Interval string
}

func NewIntervalVariable(options *IntervalVariableOptions) *dashboard.IntervalVariableBuilder {
	return dashboard.NewIntervalVariableBuilder(options.Name).
		Label(options.Label).
		Values(dashboard.StringOrMap{String: cog.ToPtr[string](options.Interval)}).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{ArrayOfString: []string{"All"}},
			Value:    dashboard.StringOrArrayOfString{ArrayOfString: []string{"$__all"}},
		})
}
