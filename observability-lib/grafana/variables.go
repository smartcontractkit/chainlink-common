package grafana

import (
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
)

type VariableOptionValues struct {
}

type VariableOption struct {
	Name         string
	Label        string
	Description  string
	Hide         dashboard.VariableHide
	CurrentText  string
	CurrentValue string
}

type CustomVariableOptions struct {
	*VariableOption
	Values     map[string]any
	Multi      bool
	IncludeAll bool
}

func NewCustomVariable(options *CustomVariableOptions) *dashboard.CustomVariableBuilder {
	if options.CurrentText == "" && options.CurrentValue == "" {
		options.CurrentText = "All"
		options.CurrentValue = "$__all"
	}

	variable := dashboard.NewCustomVariableBuilder(options.Name).
		Label(options.Label).
		Hide(options.Hide).
		Description(options.Description).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{String: cog.ToPtr(options.CurrentText)},
			Value:    dashboard.StringOrArrayOfString{String: cog.ToPtr(options.CurrentValue)},
		}).
		Multi(options.Multi).
		IncludeAll(options.IncludeAll)

	optionsList := []dashboard.VariableOption{
		{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{String: cog.ToPtr(options.CurrentText)},
			Value:    dashboard.StringOrArrayOfString{String: cog.ToPtr(options.CurrentValue)},
		},
	}
	for key, value := range options.Values {
		if key != options.CurrentText {
			option := dashboard.VariableOption{
				Text:  dashboard.StringOrArrayOfString{String: cog.ToPtr(key)},
				Value: dashboard.StringOrArrayOfString{String: cog.ToPtr(value.(string))},
			}
			optionsList = append(optionsList, option)
		}
	}
	variable.Options(optionsList)

	valuesString := ""
	for key, value := range options.Values {
		// Escape commas and colons in the value which are reserved characters for values string
		cleanValue := strings.ReplaceAll(value.(string), ",", "\\,")
		cleanValue = strings.ReplaceAll(cleanValue, ":", "\\:")
		valuesString += key
		if key != cleanValue {
			valuesString += " : " + cleanValue
		}
		valuesString += ", "
	}
	variable.Values(dashboard.StringOrMap{String: cog.ToPtr(strings.TrimSuffix(valuesString, ", "))})

	return variable
}

type QueryVariableOptions struct {
	*VariableOption
	Datasource    string
	Query         string
	Multi         bool
	Regex         string
	IncludeAll    bool
	QueryWithType map[string]any
}

func NewQueryVariable(options *QueryVariableOptions) *dashboard.QueryVariableBuilder {
	if options.CurrentText == "" && options.CurrentValue == "" {
		options.CurrentText = "All"
		options.CurrentValue = "$__all"
	}

	variable := dashboard.NewQueryVariableBuilder(options.Name).
		Label(options.Label).
		Description(options.Description).
		Hide(options.Hide).
		Datasource(datasourceRef(options.Datasource)).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentText}},
			Value:    dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentValue}},
		}).
		Sort(dashboard.VariableSortAlphabeticalAsc).
		Multi(options.Multi).
		IncludeAll(options.IncludeAll)

	if options.Query != "" {
		variable.Query(dashboard.StringOrMap{String: cog.ToPtr[string](options.Query)})
	} else if options.QueryWithType != nil {
		variable.Query(dashboard.StringOrMap{Map: options.QueryWithType})
	}

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
	if options.CurrentText == "" && options.CurrentValue == "" {
		options.CurrentText = "All"
		options.CurrentValue = "$__all"
	}

	return dashboard.NewIntervalVariableBuilder(options.Name).
		Label(options.Label).
		Description(options.Description).
		Values(dashboard.StringOrMap{String: cog.ToPtr[string](options.Interval)}).
		Current(dashboard.VariableOption{
			Selected: cog.ToPtr[bool](true),
			Text:     dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentText}},
			Value:    dashboard.StringOrArrayOfString{ArrayOfString: []string{options.CurrentValue}},
		})
}

type DataSourceVariableOptions struct {
	*VariableOption
	Type       string
	Regex      string
	Multi      bool
	IncludeAll bool
}

func NewDataSourceVariable(options *DataSourceVariableOptions) *dashboard.DatasourceVariableBuilder {
	return dashboard.NewDatasourceVariableBuilder(options.Name).
		Label(options.Label).
		Description(options.Description).
		Hide(options.Hide).
		Type(options.Type).
		Regex(options.Regex).
		Multi(options.Multi).
		IncludeAll(options.IncludeAll)
}
