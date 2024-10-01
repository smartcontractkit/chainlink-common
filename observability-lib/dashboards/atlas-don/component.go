package atlasdon

import (
	"fmt"

	"github.com/grafana/grafana-foundation-sdk/go/cog"
	"github.com/grafana/grafana-foundation-sdk/go/common"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
)

func NewDashboard(props *Props) (*grafana.Dashboard, error) {
	if props.Name == "" {
		return nil, fmt.Errorf("Name is required")
	}

	if props.OCRVersion == "" {
		return nil, fmt.Errorf("OCRVersion is required")
	}

	if props.MetricsDataSource == nil {
		return nil, fmt.Errorf("MetricsDataSource is required")
	} else {
		if props.MetricsDataSource.Name == "" {
			return nil, fmt.Errorf("MetricsDataSource.Name is required")
		}
		if props.MetricsDataSource.UID == "" {
			return nil, fmt.Errorf("MetricsDataSource.UID is required")
		}
	}

	props.platformOpts = platformPanelOpts(props.OCRVersion)

	builder := grafana.NewBuilder(&grafana.BuilderOptions{
		Name:     props.Name,
		Tags:     []string{"DON", props.OCRVersion},
		Refresh:  "30s",
		TimeFrom: "now-30m",
		TimeTo:   "now",
	})

	builder.AddVars(vars(props)...)

	builder.AddRow("Summary")
	builder.AddPanel(summary(props)...)

	builder.AddRow("OCR Contract Oracle")
	builder.AddPanel(ocrContractConfigOracle(props)...)

	builder.AddRow("DON Nodes")
	builder.AddPanel(ocrContractConfigNodes(props)...)

	builder.AddRow("Price Reporting")
	builder.AddPanel(priceReporting(props)...)

	builder.AddRow("Round / Epoch Progression")
	builder.AddPanel(roundEpochProgression(props)...)

	builder.AddRow("OCR Contract Config Delta")
	builder.AddPanel(ocrContractConfigDelta(props)...)

	return builder.Build()
}

func vars(p *Props) []cog.Builder[dashboard.VariableModel] {
	var variables []cog.Builder[dashboard.VariableModel]

	variables = append(variables, grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Job",
			Name:  "job",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(up{namespace` + p.platformOpts.LabelFilters["namespace"] + `}, job)`,
	}))

	variableFeedID := "feed_id"
	if p.OCRVersion == "ocr3" {
		variableFeedID = "feed_id_name"
	}

	variableQueryContract := grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Contract",
			Name:  "contract",
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(` + p.OCRVersion + `_contract_config_f{job="$job"}, contract)`,
	})

	variableQueryFeedID := grafana.NewQueryVariable(&grafana.QueryVariableOptions{
		VariableOption: &grafana.VariableOption{
			Label: "Feed ID",
			Name:  variableFeedID,
		},
		Datasource: p.MetricsDataSource.Name,
		Query:      `label_values(` + p.OCRVersion + `_contract_config_f{job="$job", contract="$contract"}, ` + variableFeedID + `)`,
		Multi:      true,
	})

	variables = append(variables, variableQueryContract)

	switch p.OCRVersion {
	case "ocr2":
		variables = append(variables, variableQueryFeedID)
	case "ocr3":
		variables = append(variables, variableQueryFeedID)
	}

	return variables
}

func summary(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Telemetry Down",
			Description: "Which jobs are not receiving any telemetry?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_telemetry_down{` + p.platformOpts.LabelQuery + `} == 1`,
					Legend: "{{job}} | {{report_type}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
				},
			},
		},
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Oracle Down",
			Description: "Which NOPs are not providing any telemetry?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_oracle_telemetry_down_except_telemetry_down{job=~"${job}", oracle!="csa_unknown"} == 1`,
					Legend: "{{oracle}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
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
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Feeds reporting failure",
			Description: "Which feeds are failing to report?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_feed_reporting_failure_except_feed_telemetry_down{job=~"${job}", oracle!="csa_unknown"} == 1`,
					Legend: "{{feed_id_name}} on {{job}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
				},
			},
		},
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Feed telemetry Down",
			Description: "Which feeds are not receiving any telemetry?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_feed_telemetry_down_except_telemetry_down{job=~"${job}"} == 1`,
					Legend: "{{feed_id_name}} on {{job}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
				},
			},
		},
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Oracles no observations",
			Description: "Which NOPs are not providing observations?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_oracle_blind_except_telemetry_down{job=~"${job}"} == 1`,
					Legend: "{{oracle}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
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
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Oracles not contributing observations to feeds",
			Description: "Which oracles are failing to make observations on feeds they should be participating in?",
			Span:        8,
			Height:      4,
			Query: []grafana.Query{
				{
					Expr:   `bool:` + p.OCRVersion + `_oracle_feed_no_observations_except_oracle_blind_except_feed_reporting_failure_except_feed_telemetry_down{job=~"${job}"} == 1`,
					Legend: "{{oracle}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "green"},
					{Value: grafana.Pointer[float64](0.99), Color: "red"},
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
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}

func ocrContractConfigOracle(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "OCR Contract Oracle Active",
			Description: "set to one as long as an oracle is on a feed",
			Span:        24,
			Height:      8,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum(` + p.OCRVersion + `_contract_oracle_active{` + p.platformOpts.LabelQuery + `}) by (contract, oracle)`,
					Legend: "{{oracle}}",
				},
			},
			Threshold: &grafana.ThresholdOptions{
				Mode: dashboard.ThresholdsModeAbsolute,
				Steps: []dashboard.Threshold{
					{Value: nil, Color: "default"},
					{Value: grafana.Pointer[float64](0), Color: "red"},
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
		TextMode:    common.BigValueTextModeName,
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}

func ocrContractConfigNodes(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	var variableFeedID string
	switch p.OCRVersion {
	case "ocr":
		variableFeedID = "contract"
	case "ocr2":
		variableFeedID = "feed_id"
	case "ocr3":
		variableFeedID = "feed_id_name"
	}

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Number of NOPs",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_n{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + variableFeedID + `}}`,
				},
				{
					Expr:   `` + p.OCRVersion + `_contract_config_r_max{` + p.platformOpts.LabelQuery + `}`,
					Legend: `Max nodes`,
				},
				{
					Expr:   `avg(2 * ` + p.OCRVersion + `_contract_config_f{` + p.platformOpts.LabelQuery + `} + 1)`,
					Legend: `Min nodes`,
				},
			},
			Min: grafana.Pointer[float64](0),
		},
	}))

	return panels
}

func priceReporting(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	telemetryP2PReceivedTotal := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "P2P messages received",
			Description: "From an individual node's perspective, how many messages are they receiving from other nodes? Uses ocr_telemetry_p2p_received_total",
			Span:        24,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum by (sender, receiver) (increase(` + p.OCRVersion + `_telemetry_p2p_received_total{job=~"${job}"}[5m]))`,
					Legend: `{{sender}} > {{receiver}}`,
				},
			},
		},
	})

	telemetryP2PReceivedTotalRate := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "P2P messages received Rate",
			Description: "From an individual node's perspective, how many messages are they receiving from other nodes? Uses ocr_telemetry_p2p_received_total",
			Span:        24,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `sum by (sender, receiver) (rate(` + p.OCRVersion + `_telemetry_p2p_received_total{job=~"${job}"}[5m]))`,
					Legend: `{{sender}} > {{receiver}}`,
				},
			},
		},
	})

	telemetryObservationAsk := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Ask observation in MessageObserve sent",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_observation_ask{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryObservation := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Price observation in MessageObserve sent",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_observation{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryObservationBid := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Bid observation in MessageObserve sent",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_observation_bid{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryMessageProposeObservationAsk := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Ask MessagePropose observations",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_message_propose_observation_ask{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryMessageProposeObservation := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Price MessagePropose observations",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_message_propose_observation{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryMessageProposeObservationBid := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Bid MessagePropose observations",
			Span:       24,
			Height:     6,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_message_propose_observation_bid{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryMessageProposeObservationTotal := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Total number of observations included in MessagePropose",
			Description: "How often is a node's observation included in the report?",
			Span:        24,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_message_propose_observation_total{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	telemetryMessageObserveTotal := grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Total MessageObserve sent",
			Description: "From an individual node's perspective, how often are they sending an observation?",
			Span:        24,
			Height:      6,
			Decimals:    1,
			Query: []grafana.Query{
				{
					Expr:   `rate(` + p.OCRVersion + `_telemetry_message_observe_total{` + p.platformOpts.LabelQuery + `}[5m])`,
					Legend: `{{oracle}}`,
				},
			},
		},
	})

	switch p.OCRVersion {
	case "ocr":
		panels = append(panels, telemetryP2PReceivedTotal)
		panels = append(panels, telemetryP2PReceivedTotalRate)
		panels = append(panels, telemetryObservation)
		panels = append(panels, telemetryMessageObserveTotal)
	case "ocr2":
		panels = append(panels, telemetryP2PReceivedTotal)
		panels = append(panels, telemetryP2PReceivedTotalRate)
		panels = append(panels, telemetryObservation)
		panels = append(panels, telemetryMessageObserveTotal)
	case "ocr3":
		panels = append(panels, telemetryP2PReceivedTotal)
		panels = append(panels, telemetryP2PReceivedTotalRate)
		panels = append(panels, telemetryObservationAsk)
		panels = append(panels, telemetryObservation)
		panels = append(panels, telemetryObservationBid)
		panels = append(panels, telemetryMessageProposeObservationAsk)
		panels = append(panels, telemetryMessageProposeObservation)
		panels = append(panels, telemetryMessageProposeObservationBid)
		panels = append(panels, telemetryMessageProposeObservationTotal)
		panels = append(panels, telemetryMessageObserveTotal)
	}

	return panels
}

func roundEpochProgression(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	var variableFeedID string
	switch p.OCRVersion {
	case "ocr":
		variableFeedID = "contract"
	case "ocr2":
		variableFeedID = "feed_id"
	case "ocr3":
		variableFeedID = "feed_id_name"
	}

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Agreed Epoch Progression",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "short",
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_feed_agreed_epoch{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{` + variableFeedID + `}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Round Epoch Progression",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "short",
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_telemetry_epoch_round{` + p.platformOpts.LabelQuery + `}`,
					Legend: `{{oracle}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource:  p.MetricsDataSource.Name,
			Title:       "Rounds Started",
			Description: `Tracks individual nodes firing "new round" message via telemetry (not part of P2P messages)`,
			Span:        12,
			Height:      6,
			Decimals:    1,
			Unit:        "short",
			Query: []grafana.Query{
				{
					Expr:   `rate(` + p.OCRVersion + `_telemetry_round_started_total{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{oracle}}`,
				},
			},
		},
	}))

	panels = append(panels, grafana.NewTimeSeriesPanel(&grafana.TimeSeriesPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Telemetry Ingested",
			Span:       12,
			Height:     6,
			Decimals:   1,
			Unit:       "short",
			Query: []grafana.Query{
				{
					Expr:   `rate(` + p.OCRVersion + `_telemetry_ingested_total{` + p.platformOpts.LabelQuery + `}[1m])`,
					Legend: `{{oracle}}`,
				},
			},
		},
	}))

	return panels
}

func ocrContractConfigDelta(p *Props) []*grafana.Panel {
	var panels []*grafana.Panel

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Relative Deviation Threshold",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_alpha{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Max Contract Value Age Seconds",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_c_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Observation Grace Period Seconds",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_grace_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Bad Epoch Timeout Seconds",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_progress_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Resend Interval Seconds",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_resend_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Round Interval Seconds",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_round_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	panels = append(panels, grafana.NewStatPanel(&grafana.StatPanelOptions{
		PanelOptions: &grafana.PanelOptions{
			Datasource: p.MetricsDataSource.Name,
			Title:      "Transmission Stage Timeout Second",
			Span:       8,
			Height:     4,
			Decimals:   1,
			Query: []grafana.Query{
				{
					Expr:   `` + p.OCRVersion + `_contract_config_delta_stage_seconds{` + p.platformOpts.LabelQuery + `}`,
					Legend: "{{contract}}",
				},
			},
		},
		TextMode:    common.BigValueTextModeValueAndName,
		Orientation: common.VizOrientationHorizontal,
	}))

	return panels
}
