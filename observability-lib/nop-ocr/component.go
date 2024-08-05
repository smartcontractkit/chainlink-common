package nopocr

import (
	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/utils"
)

func BuildDashboard(name string, dataSourceMetric string, platform string, ocrVersion string) (dashboard.Dashboard, error) {
	props := Props{
		MetricsDataSource: dataSourceMetric,
		PlatformOpts:      PlatformPanelOpts(platform, ocrVersion),
		OcrVersion:        ocrVersion,
	}

	builder := dashboard.NewDashboardBuilder(name).
		Tags([]string{"NOP", "Health", ocrVersion}).
		Refresh("30s").
		Time("now-1d", "now")

	utils.AddVars(builder, vars(props))

	builder.WithRow(dashboard.NewRowBuilder("Per Contract"))
	utils.AddPanels(builder, perContract(props))

	builder.WithRow(dashboard.NewRowBuilder("Per NOP"))
	utils.AddPanels(builder, perNOP(props))

	return builder.Build()
}

func vars(p Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	variables = append(variables,
		utils.QueryVariable(p.MetricsDataSource, "env", "Environment",
			`label_values(`+p.OcrVersion+`_contract_config_f{}, env)`, false))

	variables = append(variables, utils.QueryVariable(p.MetricsDataSource, "contract", "Contract",
		`label_values(`+p.OcrVersion+`_contract_oracle_active{env="$env"}, contract)`, false))

	variables = append(variables, utils.QueryVariable(p.MetricsDataSource, "oracle", "NOP",
		`label_values(`+p.OcrVersion+`_contract_oracle_active{env="$env"}, oracle)`, false))

	return variables
}

func perContract(p Props) []cog.Builder[dashboard.Panel] {
	var panelsArray []cog.Builder[dashboard.Panel]

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Rounds Epoch Progression",
		"Rounds have stopped progressing for 90 seconds means NOP is unhealthy",
		10,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_epoch_round{env=~"${env}", contract=~"${contract}"}[90s])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{oracle}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Message Observe",
		"NOP have stopped sending messages for 3mins means NOP is unhealthy",
		10,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_message_observe_total{env=~"${env}", contract=~"${contract}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{oracle}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Observations included in report",
		"NOP observations were not including in report for 3mins means NOP is unhealthy",
		10,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_message_report_req_observation_total{env=~"${env}", contract=~"${contract}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{oracle}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	return panelsArray
}

func perNOP(p Props) []cog.Builder[dashboard.Panel] {
	var panelsArray []cog.Builder[dashboard.Panel]

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Rounds Epoch Progression",
		"Rounds have stopped progressing for 5mins means NOP is unhealthy",
		32,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_epoch_round{env=~"${env}", oracle=~"${oracle}"}[90s])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{contract}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Message Observe",
		"NOP have stopped sending messages for 3mins means NOP is unhealthy",
		32,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_message_observe_total{env=~"${env}", oracle=~"${oracle}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{contract}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"Observations included in report",
		"NOP observations were not including in report for 3mins means NOP is unhealthy",
		32,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_message_report_req_observation_total{env=~"${env}", oracle=~"${oracle}"}[3m])) by (env, contract, feed_id, network_name, oracle) >bool 0)[$__range:])`,
			Legend: `{{contract}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	panelsArray = append(panelsArray, utils.StatPanel(
		p.MetricsDataSource,
		"P2P Connectivity",
		"Connectivity got interrupted for 60 seconds received from other nodes",
		32,
		24,
		2,
		"percentunit",
		common.BigValueColorModeValue,
		common.BigValueGraphModeLine,
		common.BigValueTextModeValueAndName,
		common.VizOrientationAuto,
		utils.PrometheusQuery{
			Query:  `avg_over_time((sum(changes(` + p.OcrVersion + `_telemetry_p2p_received_total{env=~"${env}", receiver=~"${oracle}"}[3m])) by (sender, receiver) >bool 0)[$__range:])`,
			Legend: `{{receiver}} < {{sender}}`,
		},
	).
		Text(common.NewVizTextDisplayOptionsBuilder().TitleSize(10).ValueSize(18)).
		JustifyMode(common.BigValueJustifyModeCenter).
		Thresholds(
			dashboard.NewThresholdsConfigBuilder().
				Mode(dashboard.ThresholdsModeAbsolute).
				Steps([]dashboard.Threshold{
					{Value: utils.Float64Ptr(0), Color: "red"},
					{Value: utils.Float64Ptr(0.80), Color: "orange"},
					{Value: utils.Float64Ptr(0.99), Color: "green"},
				})).
		WithTransformation(dashboard.DataTransformerConfig{
			Id: "renameByRegex",
			Options: map[string]any{
				"regex":         "/^(.*[\\\\\\/])/",
				"renamePattern": "",
			},
		}),
	)

	return panelsArray
}
