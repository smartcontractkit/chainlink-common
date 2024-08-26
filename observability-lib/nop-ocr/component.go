package nopocr

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

type Props struct {
	Name              string
	FolderUID         string
	MetricsDataSource *grafana.DataSource
	OCRVersion        string
}

func NewDashboard(options *grafana.DashboardOptions) (*grafana.Dashboard, error) {
	props := &Props{
		Name:              options.Name,
		MetricsDataSource: options.MetricsDataSource,
		FolderUID:         options.FolderUID,
		OCRVersion:        options.OCRVersion,
	}

	builder := grafana.NewBuilder(options, &grafana.BuilderOptions{
		Tags:     []string{"NOP", "Health", options.OCRVersion},
		Refresh:  "30s",
		TimeFrom: "now-1d",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Per Contract")
	builder.AddPanel(perContract(props)...)

	builder.AddRow("Per NOP")
	builder.AddPanel(perNOP(props)...)

	return builder.Build()
}

func vars(p *Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Environment",
			Name:  "env",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(` + p.OCRVersion + `_contract_config_f{}, env)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Contract",
			Name:  "contract",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(` + p.OCRVersion + `_contract_oracle_active{env="$env"}, contract)`,
	}))

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "NOP",
			Name:  "oracle",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(` + p.OCRVersion + `_contract_oracle_active{env="$env"}, oracle)`,
	}))

	return variables
}

func perContract(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Rounds Epoch Progression",
			Description: "Rounds have stopped progressing for 90 seconds means NOP is unhealthy",
			Span:        24,
			Height:      10,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_epoch_round{env=~"${env}", contract=~"${contract}"}[90s])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{oracle}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Message Observe",
			Description: "NOP have stopped sending messages for 3mins means NOP is unhealthy",
			Span:        24,
			Height:      10,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_message_observe_total{env=~"${env}", contract=~"${contract}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{oracle}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Observations included in report",
			Description: "NOP observations were not including in report for 3mins means NOP is unhealthy",
			Span:        24,
			Height:      10,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_message_report_req_observation_total{env=~"${env}", contract=~"${contract}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{oracle}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	return panels
}

func perNOP(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Rounds Epoch Progression",
			Description: "Rounds have stopped progressing for 5mins means NOP is unhealthy",
			Span:        24,
			Height:      32,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_epoch_round{env=~"${env}", oracle=~"${oracle}"}[90s])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{contract}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Message Observe",
			Description: "NOP have stopped sending messages for 3mins means NOP is unhealthy",
			Span:        24,
			Height:      32,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_message_observe_total{env=~"${env}", oracle=~"${oracle}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{contract}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "Observations included in report",
			Description: "NOP observations were not including in report for 3mins means NOP is unhealthy",
			Span:        24,
			Height:      32,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_message_report_req_observation_total{env=~"${env}", oracle=~"${oracle}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
					Legend: `{{contract}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Title:       "P2P Connectivity",
			Description: "Connectivity got interrupted for 60 seconds received from other nodes",
			Span:        24,
			Height:      32,
			Decimals:    2,
			Unit:        "percentunit",
			Query: []grafana.Query{
				{
					Expr:   `avg_over_time((sum(changes(` + p.OCRVersion + `_telemetry_p2p_received_total{env=~"${env}", receiver=~"${oracle}"}[3m])) by (sender, receiver) >bool 0)[$__range:])`,
					Legend: `{{receiver}} < {{sender}}`,
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
					{Value: grafana.Pointer[float64](0.80), Color: "orange"},
					{Value: grafana.Pointer[float64](0.99), Color: "green"},
				},
			},
			Transform: &grafana.TransformOptions{
				ID: "renameByRegex",
				Options: map[string]string{
					"regex":         "/^(.*[\\\\\\/])/",
					"renamePattern": "",
				},
			},
		},
		TextSize:  10,
		ValueSize: 18,
		GraphMode: common.BigValueGraphModeLine,
		TextMode:  common.BigValueTextModeValueAndName,
	}))

	return panels
}
