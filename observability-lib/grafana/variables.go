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
	Datasource   string
	Query        string
	Multi        bool
	Regex        string
	CurrentText  string
	CurrentValue string
	IncludeAll   bool
}

func NewQueryVariable(options *QueryVariableOptions) *dashboard.QueryVariableBuilder {
	if options.CurrentText == "" {
		options.CurrentText = "All"
	}

	if options.CurrentValue == "" {
		options.CurrentValue = "$__all"
	}

	variable := dashboard.NewQueryVariableBuilder(options.Name).
		Label(options.Label).
		Query(dashboard.StringOrMap{String: cog.ToPtr[string](options.Query)}).
		Datasource(datasourceRef(options.Datasource)).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentText}},
			Value:    dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentValue}},
		}).
		Sort(dashboard.VariableSortAlphabeticalAsc).
		Multi(options.Multi)

	if options.Regex != "" {
		variable.Regex(options.Regex)
	}

	if options.IncludeAll {
		variable.IncludeAll(options.IncludeAll)
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
